package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ============================================================================
// TELEGRAM BOT SERVICE & ENGINE (WITH ADMIN APPROVAL & ADVANCED PANEL)
// ============================================================================

type BotOrder struct {
	ID          string  `json:"id"`
	ChatID      int64   `json:"chat_id"`
	SenderName  string  `json:"sender_name"`
	Username    string  `json:"username"`
	PlanID      string  `json:"plan_id"`
	PlanName    string  `json:"plan_name"`
	Price       string  `json:"price"`
	TrafficGB   float64 `json:"traffic_gb"`
	ExpireDays  int     `json:"expire_days"`
	InboundID   string  `json:"inbound_id"`
	ReceiptType string  `json:"receipt_type"` // "photo" or "text"
	PhotoFileID string  `json:"photo_file_id,omitempty"`
	ReceiptText string  `json:"receipt_text,omitempty"`
	Status      string  `json:"status"` // "pending", "approved", "rejected"
	ApprovedBy  string  `json:"approved_by,omitempty"`
	GeneratedUser string `json:"generated_user,omitempty"`
	CreatedAt   string  `json:"created_at"`
}

type TelegramBotManager struct {
	mu           sync.RWMutex
	client       *http.Client
	pollerCancel context.CancelFunc
	schedCancel  context.CancelFunc
	isPolling    bool
	isScheduling bool
	botID        int64
	botUsername  string
}

var BotManager = &TelegramBotManager{
	client: &http.Client{Timeout: 35 * time.Second},
}

// In-memory conversation state
var (
	botUserStateMu sync.Mutex
	botUserStates  = make(map[int64]string) // chatID -> state: "order_<planID>", "adm_wait_user", "adm_wait_admin", "adm_wait_bc"
	botOrdersMu    sync.RWMutex
	botOrdersDB    = make(map[string]*BotOrder)
)

func initTelegramBot() {
	loadStoredOrders()
	State.mu.RLock()
	token := getBotTokenLocked()
	channelCfg := getBotChannelConfigLocked()
	sellCfg := getBotSellConfigLocked()
	State.mu.RUnlock()

	if token != "" && (channelCfg["enabled"] == true || sellCfg["enabled"] == true) {
		go BotManager.Start(token)
	}
}

func loadStoredOrders() {
	State.mu.RLock()
	defer State.mu.RUnlock()
	if raw, ok := State.Settings["bot_orders_db"].(map[string]any); ok {
		botOrdersMu.Lock()
		for id, v := range raw {
			if b, err := json.Marshal(v); err == nil {
				var ord BotOrder
				if err := json.Unmarshal(b, &ord); err == nil {
					botOrdersDB[id] = &ord
				}
			}
		}
		botOrdersMu.Unlock()
	}
}

func saveOrder(ord *BotOrder) {
	botOrdersMu.Lock()
	botOrdersDB[ord.ID] = ord
	botOrdersMu.Unlock()

	State.mu.Lock()
	db, ok := State.Settings["bot_orders_db"].(map[string]any)
	if !ok || db == nil {
		db = map[string]any{}
	}
	db[ord.ID] = ord
	State.Settings["bot_orders_db"] = db

	// Update counter
	orders := getBotOrdersLocked()
	orders[ord.Status] = orders[ord.Status] + 1
	State.Settings["bot_orders"] = orders
	State.mu.Unlock()
	saveState()
}

func recordSubscriber(chatID int64) {
	State.mu.Lock()
	subs, _ := State.Settings["bot_subscribers"].([]any)
	idStr := strconv.FormatInt(chatID, 10)
	for _, s := range subs {
		if fmt.Sprint(s) == idStr {
			State.mu.Unlock()
			return
		}
	}
	State.Settings["bot_subscribers"] = append(subs, idStr)
	State.mu.Unlock()
	saveState()
}

func getAllSubscribers() []int64 {
	State.mu.RLock()
	defer State.mu.RUnlock()
	subs, _ := State.Settings["bot_subscribers"].([]any)
	out := []int64{}
	for _, s := range subs {
		if id, err := strconv.ParseInt(fmt.Sprint(s), 10, 64); err == nil && id != 0 {
			out = append(out, id)
		}
	}
	return out
}

func (bm *TelegramBotManager) Start(token string) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	token = strings.TrimSpace(token)
	if token == "" {
		bm.stopLocked()
		return
	}

	bm.stopLocked()

	// Verify token first
	botUser, err := bm.callAPI(token, "getMe", nil)
	if err != nil {
		log.Printf("[Bot] Failed to verify token: %v", err)
		return
	}
	if res, ok := botUser["result"].(map[string]any); ok {
		if id, ok := res["id"].(float64); ok {
			bm.botID = int64(id)
		}
		if u, ok := res["username"].(string); ok {
			bm.botUsername = u
		}
	}
	log.Printf("[Bot] Authenticated as @%s (id=%d)", bm.botUsername, bm.botID)

	pollerCtx, pollerCancel := context.WithCancel(context.Background())
	schedCtx, schedCancel := context.WithCancel(context.Background())
	bm.pollerCancel = pollerCancel
	bm.schedCancel = schedCancel
	bm.isPolling = true
	bm.isScheduling = true

	go bm.pollerLoop(pollerCtx, token)
	go bm.schedulerLoop(schedCtx, token)
}

func (bm *TelegramBotManager) Stop() {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	bm.stopLocked()
}

func (bm *TelegramBotManager) stopLocked() {
	if bm.pollerCancel != nil {
		bm.pollerCancel()
		bm.pollerCancel = nil
	}
	if bm.schedCancel != nil {
		bm.schedCancel()
		bm.schedCancel = nil
	}
	bm.isPolling = false
	bm.isScheduling = false
}

func (bm *TelegramBotManager) callAPI(token, method string, payload any) (map[string]any, error) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/%s", token, method)
	var bodyReader io.Reader
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(http.MethodPost, url, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := bm.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if ok, _ := result["ok"].(bool); !ok {
		desc, _ := result["description"].(string)
		if desc == "" {
			desc = "Telegram API error"
		}
		return result, fmt.Errorf("%s", desc)
	}

	return result, nil
}

// ----------------- Admin Helpers -----------------

func getBotTokenLocked() string {
	if t, ok := State.Settings["bot_token"].(string); ok {
		return strings.TrimSpace(t)
	}
	return ""
}

func getBotChannelConfigLocked() map[string]any {
	if m, ok := State.Settings["bot_channel"].(map[string]any); ok {
		return m
	}
	return map[string]any{
		"enabled":                 false,
		"channel":                 "",
		"interval_minutes":        60,
		"username_prefix":         "white",
		"traffic_limit_gb":        0.0,
		"expire_days":             30,
		"replace_previous":        true,
		"inbound_id":              "",
		"success_count":           0,
		"error_count":             0,
		"last_run_at":             "",
		"next_run_at":             "",
		"pending_delete_user_ids": []any{},
		"last_delete_error":       "",
	}
}

func getBotSellConfigLocked() map[string]any {
	if m, ok := State.Settings["bot_sell"].(map[string]any); ok {
		return m
	}
	return map[string]any{
		"enabled":           false,
		"admin_chat_id":     "",
		"support_username":  "",
		"support_text":      "برای پشتیبانی با مدیر در ارتباط باشید.",
		"payment_url":       "",
		"payment_details":   "",
		"required_channels": []any{},
		"welcome_text":      "سلام خوش آمدید!",
	}
}

func getBotOrdersLocked() map[string]int {
	if m, ok := State.Settings["bot_orders"].(map[string]any); ok {
		res := map[string]int{}
		for k, v := range m {
			if f, ok := v.(float64); ok {
				res[k] = int(f)
			} else if i, ok := v.(int); ok {
				res[k] = i
			}
		}
		return res
	}
	return map[string]int{
		"pending":           0,
		"payment_submitted": 0,
		"approved":          0,
		"rejected":          0,
		"cancelled":         0,
	}
}

func getBotAdminIDs() []int64 {
	State.mu.RLock()
	defer State.mu.RUnlock()
	res := []int64{}
	sellCfg := getBotSellConfigLocked()
	primaryAdmin := gs(sellCfg, "admin_chat_id")
	if primaryAdmin != "" {
		for _, part := range strings.FieldsFunc(primaryAdmin, func(r rune) bool { return r == ',' || r == ' ' || r == ';' }) {
			if id, err := strconv.ParseInt(strings.TrimSpace(part), 10, 64); err == nil && id != 0 {
				res = append(res, id)
			}
		}
	}
	if admins, ok := State.Settings["bot_admins"].([]any); ok {
		for _, a := range admins {
			s := fmt.Sprint(a)
			if id, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64); err == nil && id != 0 {
				found := false
				for _, existing := range res {
					if existing == id {
						found = true
						break
					}
				}
				if !found {
					res = append(res, id)
				}
			}
		}
	}
	return res
}

func isBotAdmin(chatID int64) bool {
	for _, id := range getBotAdminIDs() {
		if id == chatID {
			return true
		}
	}
	return false
}

func addBotAdmin(chatID int64) bool {
	State.mu.Lock()
	admins, _ := State.Settings["bot_admins"].([]any)
	idStr := strconv.FormatInt(chatID, 10)
	for _, a := range admins {
		if fmt.Sprint(a) == idStr {
			State.mu.Unlock()
			return false
		}
	}
	State.Settings["bot_admins"] = append(admins, idStr)
	State.mu.Unlock()
	saveState()
	return true
}

func removeBotAdmin(chatID int64) bool {
	State.mu.Lock()
	admins, _ := State.Settings["bot_admins"].([]any)
	idStr := strconv.FormatInt(chatID, 10)
	newAdmins := []any{}
	found := false
	for _, a := range admins {
		if fmt.Sprint(a) != idStr {
			newAdmins = append(newAdmins, a)
		} else {
			found = true
		}
	}
	State.Settings["bot_admins"] = newAdmins
	State.mu.Unlock()
	if found {
		saveState()
	}
	return found
}

func getBotOrPanelDomain() string {
	State.mu.RLock()
	d, _ := State.Settings["domain"].(string)
	if d == "" {
		d, _ = State.Settings["bot_domain"].(string)
	}
	if d == "" {
		if sell, ok := State.Settings["bot_sell"].(map[string]any); ok {
			d, _ = sell["domain"].(string)
		}
	}
	if d == "" {
		d, _ = State.Settings["public_ip"].(string)
	}
	State.mu.RUnlock()
	if d == "" || d == "127.0.0.1" {
		if h := discoveredHost(nil); h != "" && h != "127.0.0.1" {
			d = h
		}
	}
	return strings.TrimSpace(d)
}

// ----------------- HTTP Handlers for Panel Web API -----------------

func handleBotStatus(w http.ResponseWriter, r *http.Request) {
	State.mu.RLock()
	token := getBotTokenLocked()
	channelCfg := getBotChannelConfigLocked()
	sellCfg := getBotSellConfigLocked()
	orders := getBotOrdersLocked()
	admins := getBotAdminIDs()

	plans := make([]*BotPlan, 0, len(State.BotPlans))
	for _, p := range State.BotPlans {
		plans = append(plans, p)
	}
	State.mu.RUnlock()

	BotManager.mu.RLock()
	pollerRunning := BotManager.isPolling
	schedRunning := BotManager.isScheduling
	BotManager.mu.RUnlock()

	masked := ""
	if len(token) > 10 {
		masked = token[:6] + "••••••••" + token[len(token)-4:]
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":                true,
		"token_configured":  token != "",
		"token_masked":      masked,
		"domain":            getBotOrPanelDomain(),
		"channel":           channelCfg,
		"sell":              sellCfg,
		"plans":             plans,
		"orders":            orders,
		"admin_ids":         admins,
		"scheduler_running": schedRunning,
		"poller_running":    pollerRunning,
	})
}

func handleBotConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"ok": false})
		return
	}
	body := readBody(r)

	State.mu.Lock()
	if rawToken := gs(body, "token"); rawToken != "" {
		State.Settings["bot_token"] = strings.TrimSpace(rawToken)
	}

	if dom := gs(body, "domain"); dom != "" {
		dom = strings.TrimSpace(dom)
		State.Settings["bot_domain"] = dom
		if State.Settings["domain"] == "" || State.Settings["domain"] == "127.0.0.1" {
			State.Settings["domain"] = dom
		}
	}

	if ch := gm(body, "channel"); ch != nil {
		curr := getBotChannelConfigLocked()
		for k, v := range ch {
			curr[k] = v
		}
		State.Settings["bot_channel"] = curr
	}

	if sell := gm(body, "sell"); sell != nil {
		curr := getBotSellConfigLocked()
		for k, v := range sell {
			curr[k] = v
		}
		State.Settings["bot_sell"] = curr
	}

	if adminIDs, ok := body["admin_ids"].([]any); ok {
		State.Settings["bot_admins"] = adminIDs
	}

	token := getBotTokenLocked()
	State.mu.Unlock()
	saveState()

	if token != "" {
		go BotManager.Start(token)
	} else {
		BotManager.Stop()
	}

	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func handleBotTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"ok": false})
		return
	}
	body := readBody(r)
	token := strings.TrimSpace(gs(body, "token"))
	channel := strings.TrimSpace(gs(body, "channel"))

	if token == "" {
		State.mu.RLock()
		token = getBotTokenLocked()
		State.mu.RUnlock()
	}
	if token == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "detail": "Token is required"})
		return
	}

	res, err := BotManager.callAPI(token, "getMe", nil)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "detail": fmt.Sprintf("Bot test failed: %v", err)})
		return
	}

	botData, _ := res["result"].(map[string]any)
	resp := map[string]any{
		"ok":  true,
		"bot": botData,
	}

	if channel != "" {
		chanIdStr := channel
		if !strings.HasPrefix(chanIdStr, "@") && !strings.HasPrefix(chanIdStr, "-100") && !strings.HasPrefix(chanIdStr, "-") {
			if strings.Contains(chanIdStr, "t.me/") {
				parts := strings.Split(chanIdStr, "t.me/")
				chanIdStr = "@" + strings.Trim(parts[len(parts)-1], "/")
			} else {
				chanIdStr = "@" + chanIdStr
			}
		}

		cres, cerr := BotManager.callAPI(token, "getChat", map[string]any{"chat_id": chanIdStr})
		if cerr == nil {
			chatMap, _ := cres["result"].(map[string]any)
			isAdmin := false
			if botId, ok := botData["id"].(float64); ok {
				mres, merr := BotManager.callAPI(token, "getChatMember", map[string]any{
					"chat_id": chanIdStr,
					"user_id": int64(botId),
				})
				if merr == nil {
					memberMap, _ := mres["result"].(map[string]any)
					status, _ := memberMap["status"].(string)
					if status == "administrator" || status == "creator" {
						isAdmin = true
					}
				}
			}
			chatMap["is_admin"] = isAdmin
			resp["channel"] = chatMap
		} else {
			resp["channel"] = map[string]any{
				"id":       chanIdStr,
				"title":    chanIdStr,
				"is_admin": false,
				"error":    cerr.Error(),
			}
		}
	}

	writeJSON(w, http.StatusOK, resp)
}

func handleBotSellPlans(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		State.mu.RLock()
		plans := make([]*BotPlan, 0, len(State.BotPlans))
		for _, p := range State.BotPlans {
			plans = append(plans, p)
		}
		State.mu.RUnlock()
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "plans": plans})

	case http.MethodPost:
		body := readBody(r)
		p := &BotPlan{
			ID:             generateShortID(),
			Name:           gs(body, "name"),
			Price:          gs(body, "price"),
			TrafficGB:      gf(body, "traffic_gb"),
			ExpireDays:     gi(body, "expire_days"),
			InboundID:      gs(body, "inbound_id"),
			UsernamePrefix: gs(body, "username_prefix"),
			Enabled:        gb(body, "enabled"),
		}
		if p.UsernamePrefix == "" {
			p.UsernamePrefix = "shop"
		}
		State.mu.Lock()
		State.BotPlans[p.ID] = p
		State.mu.Unlock()
		saveState()
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "plan": p})

	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"ok": false})
	}
}

func handleBotSellPlanByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/bot/sell/plans/")
	id = strings.TrimSpace(id)
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "detail": "plan ID is required"})
		return
	}

	switch r.Method {
	case http.MethodPatch:
		body := readBody(r)
		State.mu.Lock()
		p := State.BotPlans[id]
		if p == nil {
			State.mu.Unlock()
			writeJSON(w, http.StatusNotFound, map[string]any{"ok": false, "detail": "plan not found"})
			return
		}
		if v := gs(body, "name"); v != "" {
			p.Name = v
		}
		if v := gs(body, "price"); v != "" {
			p.Price = v
		}
		if _, ok := body["traffic_gb"]; ok {
			p.TrafficGB = gf(body, "traffic_gb")
		}
		if _, ok := body["expire_days"]; ok {
			p.ExpireDays = gi(body, "expire_days")
		}
		if _, ok := body["inbound_id"]; ok {
			p.InboundID = gs(body, "inbound_id")
		}
		if _, ok := body["username_prefix"]; ok {
			p.UsernamePrefix = gs(body, "username_prefix")
		}
		if _, ok := body["enabled"]; ok {
			p.Enabled = gb(body, "enabled")
		}
		State.mu.Unlock()
		saveState()
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "plan": p})

	case http.MethodDelete:
		State.mu.Lock()
		delete(State.BotPlans, id)
		State.mu.Unlock()
		saveState()
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "deleted": id})

	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"ok": false})
	}
}

func handleBotChannelRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"ok": false})
		return
	}

	user, err := executeChannelPost(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "detail": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "user": map[string]any{"username": user.Username}})
}

// ----------------- Core Telegram Automation Logic -----------------

func executeChannelPost(r *http.Request) (*User, error) {
	State.mu.Lock()
	token := getBotTokenLocked()
	channelCfg := getBotChannelConfigLocked()
	State.mu.Unlock()

	if token == "" {
		return nil, fmt.Errorf("Telegram Bot token is not configured")
	}

	chanTarget := gs(channelCfg, "channel")
	if chanTarget == "" {
		return nil, fmt.Errorf("Channel username or ID is empty")
	}
	if !strings.HasPrefix(chanTarget, "@") && !strings.HasPrefix(chanTarget, "-100") && !strings.HasPrefix(chanTarget, "-") {
		if strings.Contains(chanTarget, "t.me/") {
			parts := strings.Split(chanTarget, "t.me/")
			chanTarget = "@" + strings.Trim(parts[len(parts)-1], "/")
		} else {
			chanTarget = "@" + chanTarget
		}
	}

	prefix := gs(channelCfg, "username_prefix")
	if prefix == "" {
		prefix = "white"
	}
	trafficGB := gf(channelCfg, "traffic_limit_gb")
	expireDays := gi(channelCfg, "expire_days")
	if expireDays <= 0 {
		expireDays = 30
	}
	inboundID := gs(channelCfg, "inbound_id")
	replacePrev := gb(channelCfg, "replace_previous")

	if replacePrev {
		State.mu.Lock()
		toDelete := []string{}
		for uid, u := range State.Users {
			if strings.HasPrefix(u.Username, prefix+"_ch_") {
				toDelete = append(toDelete, uid)
			}
		}
		for _, uid := range toDelete {
			delete(State.Users, uid)
		}
		State.mu.Unlock()
		if len(toDelete) > 0 {
			saveState()
			scheduleXrayApply()
		}
	}

	uname := fmt.Sprintf("%s_ch_%s", prefix, generateShortID()[:5])
	trafficBytes := int64(trafficGB * 1024 * 1024 * 1024)
	expireAt := time.Now().Add(time.Duration(expireDays) * 24 * time.Hour).Format(time.RFC3339)

	inbounds := []string{}
	if inboundID != "" {
		inbounds = append(inbounds, inboundID)
	} else {
		State.mu.RLock()
		for iid, ib := range State.Inbounds {
			if !ib.System && ib.InboundID != "Node" {
				inbounds = append(inbounds, iid)
				break
			}
		}
		State.mu.RUnlock()
	}

	u := &User{
		UserID:            generateShortID(),
		Username:          uname,
		Status:            "active",
		TrafficLimitBytes: trafficBytes,
		TrafficUsedBytes:  0,
		ExpireAt:          expireAt,
		ConfigUUID:        generateUUID(),
		InboundIDs:        inbounds,
		CreatedAt:         time.Now().Format(time.RFC3339),
	}

	State.mu.Lock()
	State.Users[u.UserID] = u
	State.mu.Unlock()
	saveState()
	scheduleXrayApply()

	cfgs := generateUserConfigs(r, u)
	cfgText := ""
	if len(cfgs) > 0 {
		cfgText = cfgs[0]
	}

	panelDomain := getBotOrPanelDomain()
	if panelDomain == "" && r != nil {
		panelDomain = hostFromRequest(r)
	}

	subURL := ""
	if panelDomain != "" {
		subURL = fmt.Sprintf("https://%s/sub/%s", panelDomain, u.ConfigUUID)
	}

	trafficLabel := "نامحدود"
	if trafficGB > 0 {
		trafficLabel = fmt.Sprintf("%.1f گیگابایت", trafficGB)
	}

	msg := fmt.Sprintf("🚀 <b>کانفیگ جدید اختصاصی کانال</b>\n\n"+
		"👤 نام سرویس: <code>%s</code>\n"+
		"📊 حجم مجاز: <b>%s</b>\n"+
		"⏳ مهلت استفاده: <b>%d روز</b>\n\n",
		uname, trafficLabel, expireDays)

	if subURL != "" {
		msg += fmt.Sprintf("🔗 <b>لینک اتصال ساب‌اسکریپشن:</b>\n<code>%s</code>\n\n", subURL)
	}
	if cfgText != "" {
		msg += fmt.Sprintf("🔑 <b>کانفیگ مستقیم:</b>\n<code>%s</code>\n\n", cfgText)
	}
	msg += "📌 <i>جهت اتصال، لینک یا کانفیگ را کپی کرده و در برنامه v2rayNG / Streisand / Sing-box وارد کنید.</i>"

	_, err := BotManager.callAPI(token, "sendMessage", map[string]any{
		"chat_id":    chanTarget,
		"text":       msg,
		"parse_mode": "HTML",
	})

	nowStr := time.Now().Format("2006-01-02 15:04:05")
	intervalMin := gi(channelCfg, "interval_minutes")
	if intervalMin <= 0 {
		intervalMin = 60
	}
	nextStr := time.Now().Add(time.Duration(intervalMin) * time.Minute).Format("2006-01-02 15:04:05")

	State.mu.Lock()
	ch := getBotChannelConfigLocked()
	if err != nil {
		ch["error_count"] = gi(ch, "error_count") + 1
		ch["last_delete_error"] = err.Error()
	} else {
		ch["success_count"] = gi(ch, "success_count") + 1
		ch["last_run_at"] = nowStr
		ch["next_run_at"] = nextStr
	}
	State.Settings["bot_channel"] = ch
	State.mu.Unlock()
	saveState()

	if err != nil {
		return nil, fmt.Errorf("Telegram send failed: %v", err)
	}

	return u, nil
}

// ----------------- Scheduler Loop -----------------

func (bm *TelegramBotManager) schedulerLoop(ctx context.Context, token string) {
	ticker := time.NewTicker(45 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			State.mu.RLock()
			ch := getBotChannelConfigLocked()
			enabled := gb(ch, "enabled")
			nextAtStr := gs(ch, "next_run_at")
			State.mu.RUnlock()

			if !enabled {
				continue
			}

			shouldRun := false
			if nextAtStr == "" {
				shouldRun = true
			} else {
				if t, err := time.Parse("2006-01-02 15:04:05", nextAtStr); err == nil {
					if time.Now().After(t) {
						shouldRun = true
					}
				} else {
					shouldRun = true
				}
			}

			if shouldRun {
				if _, err := executeChannelPost(nil); err != nil {
					log.Printf("[Bot Scheduler] execute failed: %v", err)
				}
			}
		}
	}
}

// ----------------- Interactive Poller Loop -----------------

func (bm *TelegramBotManager) pollerLoop(ctx context.Context, token string) {
	offset := 0
	log.Printf("[Bot Poller] Started Telegram polling loop...")

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		res, err := bm.callAPI(token, "getUpdates", map[string]any{
			"offset":  offset,
			"timeout": 20,
		})
		if err != nil {
			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Second):
			}
			continue
		}

		updates, _ := res["result"].([]any)
		for _, rawUpd := range updates {
			upd, ok := rawUpd.(map[string]any)
			if !ok {
				continue
			}
			if uid, ok := upd["update_id"].(float64); ok {
				offset = int(uid) + 1
			}

			// 1. Handle Inline Button clicks (callback_query)
			if cb, ok := upd["callback_query"].(map[string]any); ok {
				bm.handleCallbackQuery(token, cb)
				continue
			}

			// 2. Handle Messages
			if msg, ok := upd["message"].(map[string]any); ok {
				bm.handleMessage(token, msg)
			}
		}
	}
}

func (bm *TelegramBotManager) handleMessage(token string, msg map[string]any) {
	chat, _ := msg["chat"].(map[string]any)
	if chat == nil {
		return
	}
	chatID := int64(chat["id"].(float64))
	senderName, _ := chat["first_name"].(string)
	if senderName == "" {
		senderName = "کاربر"
	}
	userName, _ := chat["username"].(string)

	recordSubscriber(chatID)

	text, _ := msg["text"].(string)
	text = strings.TrimSpace(text)
	isAdmin := isBotAdmin(chatID)

	botUserStateMu.Lock()
	step := botUserStates[chatID]
	botUserStateMu.Unlock()

	// ─── ADMIN CONVERSATION STEPS ──────────────────────────────
	if isAdmin && step != "" {
		if text == "🔙 بازگشت به پنل" || text == "/cancel" {
			botUserStateMu.Lock()
			delete(botUserStates, chatID)
			botUserStateMu.Unlock()
			bm.sendAdminPanelMenu(token, chatID)
			return
		}

		if step == "adm_wait_user" && text != "" {
			botUserStateMu.Lock()
			delete(botUserStates, chatID)
			botUserStateMu.Unlock()
			bm.handleAdminCreateUser(token, chatID, text)
			return
		}

		if step == "adm_wait_admin" && text != "" {
			botUserStateMu.Lock()
			delete(botUserStates, chatID)
			botUserStateMu.Unlock()
			newID, err := strconv.ParseInt(text, 10, 64)
			if err != nil || newID == 0 {
				_, _ = bm.callAPI(token, "sendMessage", map[string]any{
					"chat_id": chatID,
					"text":    "❌ شناسه عددی نامعتبر است. لطفاً فقط عدد Chat ID را وارد کنید.",
				})
			} else {
				if addBotAdmin(newID) {
					_, _ = bm.callAPI(token, "sendMessage", map[string]any{
						"chat_id": chatID,
						"text":    fmt.Sprintf("✅ ادمین جدید با شناسه <code>%d</code> با موفقیت افزوده شد.", newID),
						"parse_mode": "HTML",
					})
				} else {
					_, _ = bm.callAPI(token, "sendMessage", map[string]any{
						"chat_id": chatID,
						"text":    "این شناسه قبلاً در لیست ادمین‌ها ثبت شده است.",
					})
				}
			}
			bm.sendAdminPanelMenu(token, chatID)
			return
		}

		if step == "adm_wait_bc" && text != "" {
			botUserStateMu.Lock()
			delete(botUserStates, chatID)
			botUserStateMu.Unlock()
			bm.handleBroadcast(token, chatID, text)
			return
		}
	}

	// ─── CUSTOMER PAYMENT SUBMISSION (PHOTO OR TEXT) ─────────────
	if strings.HasPrefix(step, "order_") {
		planID := strings.TrimPrefix(step, "order_")
		receiptType := "text"
		photoFileID := ""
		receiptText := text

		if photos, ok := msg["photo"].([]any); ok && len(photos) > 0 {
			receiptType = "photo"
			if lastPhoto, ok := photos[len(photos)-1].(map[string]any); ok {
				photoFileID, _ = lastPhoto["file_id"].(string)
			}
			caption, _ := msg["caption"].(string)
			if caption != "" {
				receiptText = caption
			}
		}

		if photoFileID != "" || receiptText != "" {
			botUserStateMu.Lock()
			delete(botUserStates, chatID)
			botUserStateMu.Unlock()

			bm.submitOrderForApproval(token, chatID, senderName, userName, planID, receiptType, photoFileID, receiptText)
			return
		}
	}

	// ─── ADMIN COMMANDS ──────────────────────────────────────────
	if isAdmin {
		if text == "/panel" || text == "👑 پنل ادمین" || text == "پنل ادمین" {
			bm.sendAdminPanelMenu(token, chatID)
			return
		}
		if strings.HasPrefix(text, "/addadmin ") {
			target := strings.TrimSpace(strings.TrimPrefix(text, "/addadmin "))
			if newID, err := strconv.ParseInt(target, 10, 64); err == nil && newID != 0 {
				addBotAdmin(newID)
				_, _ = bm.callAPI(token, "sendMessage", map[string]any{
					"chat_id":    chatID,
					"text":       fmt.Sprintf("✅ ادمین جدید با شناسه <code>%d</code> افزوده شد.", newID),
					"parse_mode": "HTML",
				})
			} else {
				_, _ = bm.callAPI(token, "sendMessage", map[string]any{"chat_id": chatID, "text": "فرمت دستور: /addadmin 123456789"})
			}
			return
		}
		if strings.HasPrefix(text, "/deladmin ") {
			target := strings.TrimSpace(strings.TrimPrefix(text, "/deladmin "))
			if delID, err := strconv.ParseInt(target, 10, 64); err == nil && delID != 0 {
				removeBotAdmin(delID)
				_, _ = bm.callAPI(token, "sendMessage", map[string]any{
					"chat_id":    chatID,
					"text":       fmt.Sprintf("🗑 ادمین با شناسه <code>%d</code> حذف شد.", delID),
					"parse_mode": "HTML",
				})
			}
			return
		}
	}

	// ─── GENERAL USER COMMANDS ───────────────────────────────────
	switch text {
	case "/start":
		bm.sendWelcomeMenu(token, chatID, senderName, isAdmin)

	case "🛍 خرید اشتراک":
		bm.sendPlansMenu(token, chatID)

	case "👤 حساب کاربری من":
		bm.sendUserAccountInfo(token, chatID)

	case "📊 وضعیت سرورها":
		bm.sendServerStats(token, chatID)

	case "📞 پشتیبانی":
		bm.sendSupportInfo(token, chatID)

	default:
		if text != "" {
			bm.sendWelcomeMenu(token, chatID, senderName, isAdmin)
		}
	}
}

func (bm *TelegramBotManager) handleCallbackQuery(token string, cb map[string]any) {
	cbID, _ := cb["id"].(string)
	data, _ := cb["data"].(string)
	from, _ := cb["from"].(map[string]any)
	if from == nil {
		return
	}
	chatID := int64(from["id"].(float64))
	adminUser, _ := from["username"].(string)
	if adminUser == "" {
		adminUser, _ = from["first_name"].(string)
	}

	// Answer callback to dismiss loading indicator
	_, _ = bm.callAPI(token, "answerCallbackQuery", map[string]any{"callback_query_id": cbID})

	isAdmin := isBotAdmin(chatID)

	// 1. Order Approval & Rejection by Admin
	if strings.HasPrefix(data, "approve_") || strings.HasPrefix(data, "reject_") {
		if !isAdmin {
			_, _ = bm.callAPI(token, "sendMessage", map[string]any{
				"chat_id": chatID,
				"text":    "⛔ شما دسترسی ادمین برای تایید این سفارش را ندارید.",
			})
			return
		}

		parts := strings.SplitN(data, "_", 2)
		action := parts[0]
		orderID := parts[1]
		msgObj, _ := cb["message"].(map[string]any)
		bm.handleAdminApprovalAction(token, chatID, adminUser, action, orderID, msgObj)
		return
	}

	// 2. Admin Panel Actions
	if strings.HasPrefix(data, "adm_") {
		if !isAdmin {
			return
		}
		switch data {
		case "adm_create_user":
			botUserStateMu.Lock()
			botUserStates[chatID] = "adm_wait_user"
			botUserStateMu.Unlock()
			_, _ = bm.callAPI(token, "sendMessage", map[string]any{
				"chat_id": chatID,
				"text": "➕ <b>ساخت اکانت دستی:</b>\n\nلطفاً مشخصات کاربر را با فاصله بفرستید:\n<code>نام_کاربری حجم(گیگ) اعتبار(روز)</code>\n\nمثال:\n<code>ali 50 30</code>\n(حجم صفر یعنی نامحدود)\nبرای انصراف /cancel بزنید.",
				"parse_mode": "HTML",
			})

		case "adm_users_list":
			bm.sendAdminUsersList(token, chatID)

		case "adm_pending_orders":
			bm.sendAdminPendingOrders(token, chatID)

		case "adm_stats":
			bm.sendAdminDetailedStats(token, chatID)

		case "adm_add_admin":
			botUserStateMu.Lock()
			botUserStates[chatID] = "adm_wait_admin"
			botUserStateMu.Unlock()
			_, _ = bm.callAPI(token, "sendMessage", map[string]any{
				"chat_id": chatID,
				"text": "👑 <b>افزودن ادمین جدید:</b>\n\nلطفاً شناسه عددی (Chat ID) ادمین جدید را ارسال نمایید (مثال: <code>123456789</code>).\nجهت انصراف /cancel را ارسال کنید.",
				"parse_mode": "HTML",
			})

		case "adm_broadcast":
			botUserStateMu.Lock()
			botUserStates[chatID] = "adm_wait_bc"
			botUserStateMu.Unlock()
			_, _ = bm.callAPI(token, "sendMessage", map[string]any{
				"chat_id": chatID,
				"text": "📢 <b>ارسال پیام همگانی به تمام کاربران ربات:</b>\n\nمتن مورد نظر را ارسال کنید تا برای همه کاربران فرستاده شود.\nبرای انصراف /cancel بزنید.",
				"parse_mode": "HTML",
			})
		}
		return
	}

	// 3. User Actions
	if strings.HasPrefix(data, "buy_") {
		planID := strings.TrimPrefix(data, "buy_")
		bm.initiatePurchase(token, chatID, planID)
	} else if data == "check_join" || data == "menu_plans" {
		bm.sendPlansMenu(token, chatID)
	}
}

// ----------------- Menus and Workflows -----------------

func (bm *TelegramBotManager) sendWelcomeMenu(token string, chatID int64, name string, isAdmin bool) {
	State.mu.RLock()
	sellCfg := getBotSellConfigLocked()
	welcome := gs(sellCfg, "welcome_text")
	State.mu.RUnlock()

	if welcome == "" {
		welcome = fmt.Sprintf("سلام <b>%s</b> عزیز! 🌟\nبه ربات رسمی White Panel خوش آمدید.\nجهت شروع یکی از گزینه‌های زیر را انتخاب فرمایید:", name)
	}

	rows := [][]map[string]any{
		{{"text": "🛍 خرید اشتراک"}, {"text": "👤 حساب کاربری من"}},
		{{"text": "📊 وضعیت سرورها"}, {"text": "📞 پشتیبانی"}},
	}
	if isAdmin {
		rows = append(rows, []map[string]any{{"text": "👑 پنل ادمین"}})
	}

	menu := map[string]any{
		"keyboard":        rows,
		"resize_keyboard": true,
	}

	_, _ = bm.callAPI(token, "sendMessage", map[string]any{
		"chat_id":      chatID,
		"text":         welcome,
		"parse_mode":   "HTML",
		"reply_markup": menu,
	})
}

func (bm *TelegramBotManager) sendPlansMenu(token string, chatID int64) {
	State.mu.RLock()
	sellCfg := getBotSellConfigLocked()
	reqChannels, _ := sellCfg["required_channels"].([]any)
	plansList := make([]*BotPlan, 0, len(State.BotPlans))
	for _, p := range State.BotPlans {
		if p.Enabled {
			plansList = append(plansList, p)
		}
	}
	State.mu.RUnlock()

	// Check required channel membership if any
	if len(reqChannels) > 0 {
		notJoined := []map[string]string{}
		for _, rc := range reqChannels {
			if rcm, ok := rc.(map[string]any); ok {
				chatTarget := gs(rcm, "chat_id")
				joinUrl := gs(rcm, "join_url")
				if chatTarget != "" {
					mres, err := bm.callAPI(token, "getChatMember", map[string]any{
						"chat_id": chatTarget,
						"user_id": chatID,
					})
					if err == nil {
						resMap, _ := mres["result"].(map[string]any)
						st, _ := resMap["status"].(string)
						if st != "member" && st != "administrator" && st != "creator" {
							notJoined = append(notJoined, map[string]string{"chat": chatTarget, "url": joinUrl})
						}
					}
				}
			}
		}

		if len(notJoined) > 0 {
			buttons := [][]map[string]any{}
			for i, nj := range notJoined {
				url := nj["url"]
				if url == "" {
					url = fmt.Sprintf("https://t.me/%s", strings.TrimPrefix(nj["chat"], "@"))
				}
				buttons = append(buttons, []map[string]any{
					{"text": fmt.Sprintf("📢 عضویت در کانال %d", i+1), "url": url},
				})
			}
			buttons = append(buttons, []map[string]any{
				{"text": "✅ عضو شدم (بررسی مجدد)", "callback_data": "check_join"},
			})

			_, _ = bm.callAPI(token, "sendMessage", map[string]any{
				"chat_id":    chatID,
				"text":       "⚠️ برای دسترسی به خرید، ابتدا باید در کانال‌های زیر عضو شوید:",
				"parse_mode": "HTML",
				"reply_markup": map[string]any{
					"inline_keyboard": buttons,
				},
			})
			return
		}
	}

	if len(plansList) == 0 {
		_, _ = bm.callAPI(token, "sendMessage", map[string]any{
			"chat_id": chatID,
			"text":    "در حال حاضر هیچ پلن فعالی تعریف نشده است. لطفاً بعداً بررسی فرمایید.",
		})
		return
	}

	buttons := [][]map[string]any{}
	for _, p := range plansList {
		trafficStr := "نامحدود"
		if p.TrafficGB > 0 {
			trafficStr = fmt.Sprintf("%.0f گیگ", p.TrafficGB)
		}
		btnLabel := fmt.Sprintf("⚡ %s (%s / %d روز) - %s تومان", p.Name, trafficStr, p.ExpireDays, p.Price)
		buttons = append(buttons, []map[string]any{
			{"text": btnLabel, "callback_data": "buy_" + p.ID},
		})
	}

	_, _ = bm.callAPI(token, "sendMessage", map[string]any{
		"chat_id": chatID,
		"text":    "✨ <b>لطفاً یکی از پلن‌های زیر را انتخاب کنید:</b>",
		"parse_mode": "HTML",
		"reply_markup": map[string]any{
			"inline_keyboard": buttons,
		},
	})
}

func (bm *TelegramBotManager) initiatePurchase(token string, chatID int64, planID string) {
	State.mu.RLock()
	plan := State.BotPlans[planID]
	sellCfg := getBotSellConfigLocked()
	payDetails := gs(sellCfg, "payment_details")
	payUrl := gs(sellCfg, "payment_url")
	State.mu.RUnlock()

	if plan == nil {
		_, _ = bm.callAPI(token, "sendMessage", map[string]any{
			"chat_id": chatID,
			"text":    "❌ پلن مورد نظر یافت نشد.",
		})
		return
	}

	botUserStateMu.Lock()
	botUserStates[chatID] = "order_" + planID
	botUserStateMu.Unlock()

	msg := fmt.Sprintf("🛒 <b>سفارش انتخابی شما:</b>\n\n"+
		"🔹 پلن: <b>%s</b>\n"+
		"📊 حجم: <b>%.0f گیگابایت</b>\n"+
		"⏳ مدت اعتبار: <b>%d روز</b>\n"+
		"💰 مبلغ قابل پرداخت: <b>%s تومان</b>\n\n",
		plan.Name, plan.TrafficGB, plan.ExpireDays, plan.Price)

	if payDetails != "" {
		msg += fmt.Sprintf("💳 <b>اطلاعات پرداخت / شماره کارت:</b>\n<code>%s</code>\n\n", payDetails)
	}
	if payUrl != "" {
		msg += fmt.Sprintf("🌐 <a href=\"%s\">جهت پرداخت آنلاین اینجا کلیک کنید</a>\n\n", payUrl)
	}

	msg += "📸 <i>لطفاً پس از واریز، تصویر فیش پرداخت یا شماره پیگیری تراکنش را در همین چت ارسال نمایید تا پس از تایید مدیریت، سرویس شما فعال گردد.</i>"

	_, _ = bm.callAPI(token, "sendMessage", map[string]any{
		"chat_id":    chatID,
		"text":       msg,
		"parse_mode": "HTML",
	})
}

// ----------------- Order Submission with Admin Approval -----------------

func (bm *TelegramBotManager) submitOrderForApproval(token string, chatID int64, senderName, username, planID, receiptType, photoFileID, receiptText string) {
	State.mu.RLock()
	plan := State.BotPlans[planID]
	State.mu.RUnlock()

	if plan == nil {
		_, _ = bm.callAPI(token, "sendMessage", map[string]any{
			"chat_id": chatID,
			"text":    "❌ پلن انتخابی معتبر نیست.",
		})
		return
	}

	orderID := fmt.Sprintf("W%d%s", time.Now().Unix()%100000, generateShortID()[:3])
	order := &BotOrder{
		ID:          orderID,
		ChatID:      chatID,
		SenderName:  senderName,
		Username:    username,
		PlanID:      plan.ID,
		PlanName:    plan.Name,
		Price:       plan.Price,
		TrafficGB:   plan.TrafficGB,
		ExpireDays:  plan.ExpireDays,
		InboundID:   plan.InboundID,
		ReceiptType: receiptType,
		PhotoFileID: photoFileID,
		ReceiptText: receiptText,
		Status:      "pending",
		CreatedAt:   time.Now().Format("2006-01-02 15:04:05"),
	}

	saveOrder(order)

	// Notify customer
	customerMsg := fmt.Sprintf("✅ <b>رسید پرداخت شما با موفقیت دریافت شد!</b>\n\n"+
		"🧾 کد پیگیری سفارش: <code>%s</code>\n"+
		"📦 پلن انتخابی: <b>%s</b>\n"+
		"💰 مبلغ: <b>%s تومان</b>\n\n"+
		"⏳ <i>سفارش شما در صف بررسی و تایید مدیریت قرار گرفت. به محض تایید توسط مدیر، کانفیگ اختصاصی و لینک ساب‌اسکریپشن به صورت خودکار برای شما ارسال خواهد شد.</i>",
		order.ID, order.PlanName, order.Price)

	_, _ = bm.callAPI(token, "sendMessage", map[string]any{
		"chat_id":    chatID,
		"text":       customerMsg,
		"parse_mode": "HTML",
	})

	// Send to all Admins with Approve / Reject inline buttons
	userHandle := order.Username
	if userHandle != "" {
		userHandle = "@" + userHandle
	} else {
		userHandle = "ندارد"
	}

	adminCaption := fmt.Sprintf("🔔 <b>رسید پرداخت جدید نیازمند تایید!</b>\n\n"+
		"🧾 شناسه سفارش: <code>%s</code>\n"+
		"👤 خریدار: <b>%s</b> (%s)\n"+
		"🆔 شناسه عددی (Chat ID): <code>%d</code>\n"+
		"📦 پلن: <b>%s</b> (%.0f گیگ / %d روز)\n"+
		"💰 مبلغ واریزی: <b>%s تومان</b>\n"+
		"🕒 زمان: <code>%s</code>\n",
		order.ID, order.SenderName, userHandle, order.ChatID,
		order.PlanName, order.TrafficGB, order.ExpireDays,
		order.Price, order.CreatedAt)

	if receiptText != "" {
		adminCaption += fmt.Sprintf("📝 توضیحات / کد فیش: <code>%s</code>\n", receiptText)
	}

	markup := map[string]any{
		"inline_keyboard": [][]map[string]any{
			{
				{"text": "✅ تایید و صدور اشتراک", "callback_data": "approve_" + order.ID},
				{"text": "❌ رد سفارش", "callback_data": "reject_" + order.ID},
			},
		},
	}

	admins := getBotAdminIDs()
	if len(admins) == 0 {
		log.Printf("[Bot] Warning: No admins configured to receive order %s", order.ID)
	}

	for _, adminID := range admins {
		if photoFileID != "" {
			_, _ = bm.callAPI(token, "sendPhoto", map[string]any{
				"chat_id":      adminID,
				"photo":        photoFileID,
				"caption":      adminCaption,
				"parse_mode":   "HTML",
				"reply_markup": markup,
			})
		} else {
			_, _ = bm.callAPI(token, "sendMessage", map[string]any{
				"chat_id":      adminID,
				"text":         adminCaption,
				"parse_mode":   "HTML",
				"reply_markup": markup,
			})
		}
	}
}

func (bm *TelegramBotManager) handleAdminApprovalAction(token string, adminChatID int64, adminUser, action, orderID string, cbMsg map[string]any) {
	botOrdersMu.Lock()
	order := botOrdersDB[orderID]
	botOrdersMu.Unlock()

	if order == nil {
		_, _ = bm.callAPI(token, "sendMessage", map[string]any{
			"chat_id": adminChatID,
			"text":    fmt.Sprintf("❌ سفارش <code>%s</code> یافت نشد.", orderID),
			"parse_mode": "HTML",
		})
		return
	}

	if order.Status != "pending" {
		statusLabel := map[string]string{"approved": "قبلاً تایید شده", "rejected": "قبلاً رد شده"}[order.Status]
		_, _ = bm.callAPI(token, "sendMessage", map[string]any{
			"chat_id": adminChatID,
			"text":    fmt.Sprintf("⚠️ این سفارش %s است (توسط %s).", statusLabel, order.ApprovedBy),
		})
		return
	}

	msgID := 0
	if cbMsg != nil {
		if mid, ok := cbMsg["message_id"].(float64); ok {
			msgID = int(mid)
		}
	}

	if action == "reject" {
		order.Status = "rejected"
		order.ApprovedBy = adminUser
		saveOrder(order)

		// Notify customer
		rejectMsg := fmt.Sprintf("❌ <b>رسید پرداخت شما تایید نشد!</b>\n\n"+
			"سفارش <code>%s</code> برای پلن <b>%s</b> مورد تایید مدیریت قرار نگرفت.\n"+
			"لطفاً در صورت واریز اشتباه یا نیاز به پیگیری با پشتیبانی در ارتباط باشید.",
			order.ID, order.PlanName)

		_, _ = bm.callAPI(token, "sendMessage", map[string]any{
			"chat_id":    order.ChatID,
			"text":       rejectMsg,
			"parse_mode": "HTML",
		})

		// Edit admin message
		if msgID > 0 {
			resText := fmt.Sprintf("❌ <b>سفارش %s رد شد.</b>\nتوسط: @%s", order.ID, adminUser)
			if cbMsg["photo"] != nil {
				_, _ = bm.callAPI(token, "editMessageCaption", map[string]any{
					"chat_id":    adminChatID,
					"message_id": msgID,
					"caption":    resText,
					"parse_mode": "HTML",
				})
			} else {
				_, _ = bm.callAPI(token, "editMessageText", map[string]any{
					"chat_id":    adminChatID,
					"message_id": msgID,
					"text":       resText,
					"parse_mode": "HTML",
				})
			}
		}
		return
	}

	// ACTION: APPROVE -> Generate user and send configs
	order.Status = "approved"
	order.ApprovedBy = adminUser

	// Create user in White Panel
	uname := fmt.Sprintf("wp_%d_%s", order.ChatID%10000, generateShortID()[:4])
	trafficBytes := int64(order.TrafficGB * 1024 * 1024 * 1024)
	expireDays := order.ExpireDays
	if expireDays <= 0 {
		expireDays = 30
	}
	expireAt := time.Now().Add(time.Duration(expireDays) * 24 * time.Hour).Format(time.RFC3339)

	inbounds := []string{}
	if order.InboundID != "" {
		inbounds = append(inbounds, order.InboundID)
	} else {
		State.mu.RLock()
		for iid, ib := range State.Inbounds {
			if !ib.System && ib.InboundID != "Node" {
				inbounds = append(inbounds, iid)
				break
			}
		}
		State.mu.RUnlock()
	}

	u := &User{
		UserID:            generateShortID(),
		Username:          uname,
		Status:            "active",
		TrafficLimitBytes: trafficBytes,
		TrafficUsedBytes:  0,
		ExpireAt:          expireAt,
		ConfigUUID:        generateUUID(),
		InboundIDs:        inbounds,
		Note:              fmt.Sprintf("Telegram Order: %s (Chat %d - %s)", order.ID, order.ChatID, order.SenderName),
		CreatedAt:         time.Now().Format(time.RFC3339),
	}

	State.mu.Lock()
	State.Users[u.UserID] = u
	State.mu.Unlock()
	saveState()
	scheduleXrayApply()

	order.GeneratedUser = uname
	saveOrder(order)

	// Generate configs
	cfgs := generateUserConfigs(nil, u)
	cfgDirect := ""
	if len(cfgs) > 0 {
		cfgDirect = cfgs[0]
	}

	panelDomain := getBotOrPanelDomain()

	subURL := ""
	if panelDomain != "" {
		subURL = fmt.Sprintf("https://%s/sub/%s", panelDomain, u.ConfigUUID)
	}

	// Send to customer
	deliveryMsg := fmt.Sprintf("🎉 <b>پرداخت شما تایید شد و سرویس اختصاصی شما فعال گردید!</b>\n\n"+
		"👤 نام کاربری: <code>%s</code>\n"+
		"📊 حجم پلن: <b>%.0f گیگابایت</b>\n"+
		"⏳ مدت اعتبار: <b>%d روز</b>\n\n",
		uname, order.TrafficGB, expireDays)

	if subURL != "" {
		deliveryMsg += fmt.Sprintf("🔗 <b>لینک ساب‌اسکریپشن خودکار (Subscription):</b>\n<code>%s</code>\n\n", subURL)
	}
	if cfgDirect != "" {
		deliveryMsg += fmt.Sprintf("🔑 <b>کانفیگ مستقیم (VLESS):</b>\n<code>%s</code>\n\n", cfgDirect)
	}
	deliveryMsg += "📌 <i>کانفیگ یا لینک بالا را کپی و در نرم‌افزار v2rayNG / Streisand / Sing-box ایمپورت نمایید.</i>"

	_, _ = bm.callAPI(token, "sendMessage", map[string]any{
		"chat_id":    order.ChatID,
		"text":       deliveryMsg,
		"parse_mode": "HTML",
	})

	// Edit admin message
	if msgID > 0 {
		confText := fmt.Sprintf("✅ <b>سفارش %s با موفقیت تایید شد!</b>\nتوسط: @%s\nاکانت ساخته‌شده: <code>%s</code>", order.ID, adminUser, uname)
		if cbMsg["photo"] != nil {
			_, _ = bm.callAPI(token, "editMessageCaption", map[string]any{
				"chat_id":    adminChatID,
				"message_id": msgID,
				"caption":    confText,
				"parse_mode": "HTML",
			})
		} else {
			_, _ = bm.callAPI(token, "editMessageText", map[string]any{
				"chat_id":    adminChatID,
				"message_id": msgID,
				"text":       confText,
				"parse_mode": "HTML",
			})
		}
	}
}

// ----------------- Dedicated Admin Panel in Telegram -----------------

func (bm *TelegramBotManager) sendAdminPanelMenu(token string, chatID int64) {
	text := "👑 <b>پنل مدیریت هوشمند White Panel</b>\n\n" +
		"شما به عنوان ادمین مجاز وارد شدید. لطفاً یکی از بخش‌های زیر را انتخاب فرمایید:"

	markup := map[string]any{
		"inline_keyboard": [][]map[string]any{
			{
				{"text": "➕ ساخت اکانت دستی", "callback_data": "adm_create_user"},
				{"text": "👥 لیست کاربران", "callback_data": "adm_users_list"},
			},
			{
				{"text": "📦 سفارش‌های معلق", "callback_data": "adm_pending_orders"},
				{"text": "📊 آمار پنل و سرور", "callback_data": "adm_stats"},
			},
			{
				{"text": "👑 افزودن ادمین", "callback_data": "adm_add_admin"},
				{"text": "📢 پیام همگانی", "callback_data": "adm_broadcast"},
			},
		},
	}

	_, _ = bm.callAPI(token, "sendMessage", map[string]any{
		"chat_id":      chatID,
		"text":         text,
		"parse_mode":   "HTML",
		"reply_markup": markup,
	})
}

func (bm *TelegramBotManager) handleAdminCreateUser(token string, chatID int64, input string) {
	parts := strings.Fields(input)
	if len(parts) < 3 {
		_, _ = bm.callAPI(token, "sendMessage", map[string]any{
			"chat_id": chatID,
			"text":    "❌ فرمت اشتباه است. لطفاً به این شکل بفرستید:\n<code>نام حجم(GB) روز</code>\nمثال:\n<code>ali 50 30</code>",
			"parse_mode": "HTML",
		})
		return
	}

	uname := parts[0]
	trafficGB, _ := strconv.ParseFloat(parts[1], 64)
	expireDays, _ := strconv.Atoi(parts[2])
	if expireDays <= 0 {
		expireDays = 30
	}

	trafficBytes := int64(trafficGB * 1024 * 1024 * 1024)
	expireAt := time.Now().Add(time.Duration(expireDays) * 24 * time.Hour).Format(time.RFC3339)

	inbounds := []string{}
	State.mu.RLock()
	for iid, ib := range State.Inbounds {
		if !ib.System && ib.InboundID != "Node" {
			inbounds = append(inbounds, iid)
			break
		}
	}
	State.mu.RUnlock()

	u := &User{
		UserID:            generateShortID(),
		Username:          uname,
		Status:            "active",
		TrafficLimitBytes: trafficBytes,
		TrafficUsedBytes:  0,
		ExpireAt:          expireAt,
		ConfigUUID:        generateUUID(),
		InboundIDs:        inbounds,
		Note:              fmt.Sprintf("Created via Telegram Admin Panel by %d", chatID),
		CreatedAt:         time.Now().Format(time.RFC3339),
	}

	State.mu.Lock()
	State.Users[u.UserID] = u
	State.mu.Unlock()
	saveState()
	scheduleXrayApply()

	cfgs := generateUserConfigs(nil, u)
	cfgDirect := ""
	if len(cfgs) > 0 {
		cfgDirect = cfgs[0]
	}

	panelDomain := getBotOrPanelDomain()

	subURL := ""
	if panelDomain != "" {
		subURL = fmt.Sprintf("https://%s/sub/%s", panelDomain, u.ConfigUUID)
	}

	msg := fmt.Sprintf("✅ <b>اکانت جدید با موفقیت ایجاد شد!</b>\n\n"+
		"👤 نام کاربری: <code>%s</code>\n"+
		"📊 حجم مجاز: <b>%.1f GB</b>\n"+
		"⏳ مهلت اعتبار: <b>%d روز</b>\n\n",
		uname, trafficGB, expireDays)

	if subURL != "" {
		msg += fmt.Sprintf("🔗 <b>لینک ساب‌اسکریپشن:</b>\n<code>%s</code>\n\n", subURL)
	}
	if cfgDirect != "" {
		msg += fmt.Sprintf("🔑 <b>کانفیگ مستقیم:</b>\n<code>%s</code>\n\n", cfgDirect)
	}
	msg += "📌 <i>کانفیگ را کپی کرده و به خریدار تحویل دهید.</i>"

	_, _ = bm.callAPI(token, "sendMessage", map[string]any{
		"chat_id":    chatID,
		"text":       msg,
		"parse_mode": "HTML",
	})
	bm.sendAdminPanelMenu(token, chatID)
}

func (bm *TelegramBotManager) sendAdminUsersList(token string, chatID int64) {
	State.mu.RLock()
	users := make([]*User, 0, len(State.Users))
	for _, u := range State.Users {
		users = append(users, u)
	}
	State.mu.RUnlock()

	if len(users) == 0 {
		_, _ = bm.callAPI(token, "sendMessage", map[string]any{
			"chat_id": chatID,
			"text":    "هیچ کاربری در پنل ثبت نشده است.",
		})
		return
	}

	maxShow := 15
	if len(users) > maxShow {
		users = users[len(users)-maxShow:]
	}

	msg := fmt.Sprintf("👥 <b>لیست %d کاربر اخیر پنل:</b>\n\n", len(users))
	for i, u := range users {
		usedGB := float64(u.TrafficUsedBytes) / (1024 * 1024 * 1024)
		totalGB := float64(u.TrafficLimitBytes) / (1024 * 1024 * 1024)
		totalStr := "نامحدود"
		if u.TrafficLimitBytes > 0 {
			totalStr = fmt.Sprintf("%.1f GB", totalGB)
		}
		statusBadge := "🟢"
		if !isActiveUser(u) {
			statusBadge = "🔴"
		}
		msg += fmt.Sprintf("%s <b>%d.</b> <code>%s</code> (%.1f / %s)\n", statusBadge, i+1, u.Username, usedGB, totalStr)
	}

	_, _ = bm.callAPI(token, "sendMessage", map[string]any{
		"chat_id":    chatID,
		"text":       msg,
		"parse_mode": "HTML",
	})
	bm.sendAdminPanelMenu(token, chatID)
}

func (bm *TelegramBotManager) sendAdminPendingOrders(token string, chatID int64) {
	botOrdersMu.RLock()
	pending := []*BotOrder{}
	for _, o := range botOrdersDB {
		if o.Status == "pending" {
			pending = append(pending, o)
		}
	}
	botOrdersMu.RUnlock()

	if len(pending) == 0 {
		_, _ = bm.callAPI(token, "sendMessage", map[string]any{
			"chat_id": chatID,
			"text":    "✨ در حال حاضر هیچ سفارشی در صف انتظار وجود ندارد.",
		})
		bm.sendAdminPanelMenu(token, chatID)
		return
	}

	msg := fmt.Sprintf("📦 <b>تعداد %d سفارش معلق در انتظار تایید:</b>\n\n", len(pending))
	buttons := [][]map[string]any{}
	for _, o := range pending {
		msg += fmt.Sprintf("🔹 کد: <code>%s</code> | خریدار: <b>%s</b> | پلن: <b>%s</b> (%s تومان)\n",
			o.ID, o.SenderName, o.PlanName, o.Price)

		buttons = append(buttons, []map[string]any{
			{"text": "✅ تایید " + o.ID, "callback_data": "approve_" + o.ID},
			{"text": "❌ رد " + o.ID, "callback_data": "reject_" + orderCleanID(o.ID)},
		})
	}

	_, _ = bm.callAPI(token, "sendMessage", map[string]any{
		"chat_id":      chatID,
		"text":         msg,
		"parse_mode":   "HTML",
		"reply_markup": map[string]any{"inline_keyboard": buttons},
	})
}

func orderCleanID(id string) string {
	return id
}

func (bm *TelegramBotManager) sendAdminDetailedStats(token string, chatID int64) {
	State.mu.RLock()
	userCount := len(State.Users)
	activeUserCount := 0
	for _, u := range State.Users {
		if isActiveUser(u) {
			activeUserCount++
		}
	}
	inboundCount := len(State.Inbounds)
	domain, _ := State.Settings["domain"].(string)
	admins := getBotAdminIDs()
	subsCount := len(getAllSubscribers())
	orders := getBotOrdersLocked()
	State.mu.RUnlock()

	xrayStatus := "خاموش 🔴"
	if xrayIsRunning() {
		xrayStatus = "روشن 🟢"
	}

	msg := fmt.Sprintf("📊 <b>آمار جامع سرور و پنل:</b>\n\n"+
		"🌐 دامنه پنل: <code>%s</code>\n"+
		"⚡ هسته Xray: <b>%s</b>\n"+
		"👥 تعداد کل کاربران: <b>%d</b> (فعال: <b>%d</b>)\n"+
		"🛣 اینباندهای فعال: <b>%d</b>\n"+
		"🤖 اعضای تعاملی ربات: <b>%d</b> نفر\n"+
		"👑 تعداد ادمین‌های ربات: <b>%d</b> نفر\n\n"+
		"📦 <b>وضعیت سفارش‌ها:</b>\n"+
		"• تایید شده: <b>%d</b>\n"+
		"• در انتظار: <b>%d</b>\n"+
		"• رد شده: <b>%d</b>",
		domain, xrayStatus, userCount, activeUserCount, inboundCount,
		subsCount, len(admins),
		orders["approved"], orders["pending"], orders["rejected"])

	_, _ = bm.callAPI(token, "sendMessage", map[string]any{
		"chat_id":    chatID,
		"text":       msg,
		"parse_mode": "HTML",
	})
	bm.sendAdminPanelMenu(token, chatID)
}

func (bm *TelegramBotManager) handleBroadcast(token string, adminChatID int64, text string) {
	subs := getAllSubscribers()
	if len(subs) == 0 {
		_, _ = bm.callAPI(token, "sendMessage", map[string]any{
			"chat_id": adminChatID,
			"text":    "هیچ کاربری در لیست تعامل با ربات ثبت نشده است.",
		})
		bm.sendAdminPanelMenu(token, adminChatID)
		return
	}

	_, _ = bm.callAPI(token, "sendMessage", map[string]any{
		"chat_id": adminChatID,
		"text":    fmt.Sprintf("📢 ارسال پیام همگانی به %d کاربر آغاز شد...", len(subs)),
	})

	go func() {
		sent := 0
		failed := 0
		for _, subID := range subs {
			_, err := bm.callAPI(token, "sendMessage", map[string]any{
				"chat_id":    subID,
				"text":       text,
				"parse_mode": "HTML",
			})
			if err == nil {
				sent++
			} else {
				failed++
			}
			time.Sleep(40 * time.Millisecond) // Avoid rate limits
		}

		report := fmt.Sprintf("📢 <b>گزارش ارسال همگانی:</b>\n\n✅ ارسال موفق: <b>%d</b>\n❌ خطا یا بلاک: <b>%d</b>", sent, failed)
		_, _ = bm.callAPI(token, "sendMessage", map[string]any{
			"chat_id":    adminChatID,
			"text":       report,
			"parse_mode": "HTML",
		})
	}()
}

// ----------------- Customer Account / Stats / Support -----------------

func (bm *TelegramBotManager) sendUserAccountInfo(token string, chatID int64) {
	State.mu.RLock()
	panelDomain, _ := State.Settings["domain"].(string)
	idStr := fmt.Sprintf("%d", chatID)
	var foundUsers []*User
	for _, u := range State.Users {
		if strings.Contains(u.Note, idStr) || strings.Contains(u.Username, fmt.Sprintf("_%d_", chatID%10000)) {
			foundUsers = append(foundUsers, u)
		}
	}
	State.mu.RUnlock()

	if len(foundUsers) == 0 {
		_, _ = bm.callAPI(token, "sendMessage", map[string]any{
			"chat_id": chatID,
			"text":    "🔍 شما در حال حاضر اشتراک فعالی در سیستم ندارید.\nبرای تهیه اشتراک گزینه «🛍 خرید اشتراک» را لمس کنید.",
		})
		return
	}

	msg := "👤 <b>اطلاعات اشتراک‌های شما:</b>\n\n"
	for i, u := range foundUsers {
		usedGB := float64(u.TrafficUsedBytes) / (1024 * 1024 * 1024)
		totalGB := float64(u.TrafficLimitBytes) / (1024 * 1024 * 1024)
		totalStr := "نامحدود"
		if u.TrafficLimitBytes > 0 {
			totalStr = fmt.Sprintf("%.1f GB", totalGB)
		}

		subURL := ""
		if panelDomain != "" {
			subURL = fmt.Sprintf("https://%s/sub/%s", panelDomain, u.ConfigUUID)
		}

		msg += fmt.Sprintf("🔹 <b>اشتراک %d:</b> <code>%s</code>\n"+
			"📊 مصرف: <b>%.2f / %s</b>\n"+
			"⏳ تاریخ انقضا: <code>%s</code>\n",
			i+1, u.Username, usedGB, totalStr, u.ExpireAt)

		if subURL != "" {
			msg += fmt.Sprintf("🔗 لینک ساب: <code>%s</code>\n\n", subURL)
		} else {
			msg += "\n"
		}
	}

	_, _ = bm.callAPI(token, "sendMessage", map[string]any{
		"chat_id":    chatID,
		"text":       msg,
		"parse_mode": "HTML",
	})
}

func (bm *TelegramBotManager) sendServerStats(token string, chatID int64) {
	State.mu.RLock()
	userCount := len(State.Users)
	inboundCount := len(State.Inbounds)
	State.mu.RUnlock()

	msg := fmt.Sprintf("🟢 <b>وضعیت سرورها: پایدار و آنلاین</b>\n\n"+
		"⚡ پروتکل‌های فعال: VLESS Reality, TLS WS, XHTTP, gRPC, Shadowsocks\n"+
		"👥 کاربران فعال سیستم: <b>%d</b> نفر\n"+
		"🌐 مسیرهای اتصال سرور: <b>%d</b> درگاه مستقیم و بهینه‌شده\n\n"+
		"🛡 تمامی اتصالات مجهز به سیستم ضد تحریم و کمترین پینگ هستند.",
		userCount, inboundCount)

	_, _ = bm.callAPI(token, "sendMessage", map[string]any{
		"chat_id":    chatID,
		"text":       msg,
		"parse_mode": "HTML",
	})
}

func (bm *TelegramBotManager) sendSupportInfo(token string, chatID int64) {
	State.mu.RLock()
	sellCfg := getBotSellConfigLocked()
	supportUser := gs(sellCfg, "support_username")
	supportText := gs(sellCfg, "support_text")
	State.mu.RUnlock()

	if supportText == "" {
		supportText = "💬 جهت هرگونه سوال، مشاوره یا راهنمایی در اتصال با پشتیبانی تماس بگیرید:"
	}

	msg := fmt.Sprintf("📞 <b>پشتیبانی White Panel:</b>\n\n%s\n", supportText)
	if supportUser != "" {
		if !strings.HasPrefix(supportUser, "@") {
			supportUser = "@" + supportUser
		}
		msg += fmt.Sprintf("\nآیدی تلگرام: <b>%s</b>", supportUser)
	}

	replyMarkup := map[string]any{}
	if supportUser != "" {
		cleanUser := strings.TrimPrefix(supportUser, "@")
		replyMarkup["inline_keyboard"] = [][]map[string]any{
			{{"text": "💬 ارتباط مستقیم با پشتیبان", "url": "https://t.me/" + cleanUser}},
		}
	}

	payload := map[string]any{
		"chat_id":    chatID,
		"text":       msg,
		"parse_mode": "HTML",
	}
	if len(replyMarkup) > 0 {
		payload["reply_markup"] = replyMarkup
	}

	_, _ = bm.callAPI(token, "sendMessage", payload)
}
