package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

// ============================================================================
// TELEGRAM BOT SERVICE & ENGINE
// ============================================================================

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

func initTelegramBot() {
	State.mu.RLock()
	token := getBotTokenLocked()
	channelCfg := getBotChannelConfigLocked()
	sellCfg := getBotSellConfigLocked()
	State.mu.RUnlock()

	if token != "" && (channelCfg["enabled"] == true || sellCfg["enabled"] == true) {
		go BotManager.Start(token)
	}
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

// ----------------- Helpers to read/write state safely -----------------

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

// ----------------- HTTP Handlers for Panel Web API -----------------

func handleBotStatus(w http.ResponseWriter, r *http.Request) {
	State.mu.RLock()
	token := getBotTokenLocked()
	channelCfg := getBotChannelConfigLocked()
	sellCfg := getBotSellConfigLocked()
	orders := getBotOrdersLocked()

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
		"channel":           channelCfg,
		"sell":              sellCfg,
		"plans":             plans,
		"orders":            orders,
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

	// If replace_previous is requested, delete previously generated channel users
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

	// Create new user
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

	// Generate configs
	cfgs := generateUserConfigs(r, u)
	cfgText := ""
	if len(cfgs) > 0 {
		cfgText = cfgs[0]
	}

	panelDomain := ""
	State.mu.RLock()
	if d, ok := State.Settings["domain"].(string); ok && d != "" {
		panelDomain = d
	}
	State.mu.RUnlock()
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

// In-memory conversation state for pending payments
var (
	botUserStateMu sync.Mutex
	botUserStates  = make(map[int64]string) // chatID -> "order_<planID>"
)

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
	text, _ := msg["text"].(string)
	text = strings.TrimSpace(text)
	senderName, _ := chat["first_name"].(string)
	if senderName == "" {
		senderName = "کاربر گرامی"
	}

	botUserStateMu.Lock()
	step := botUserStates[chatID]
	botUserStateMu.Unlock()

	// If waiting for payment receipt (photo or payment tracking code)
	if strings.HasPrefix(step, "order_") && (msg["photo"] != nil || text != "") {
		planID := strings.TrimPrefix(step, "order_")
		botUserStateMu.Lock()
		delete(botUserStates, chatID)
		botUserStateMu.Unlock()

		bm.completeOrderAndDeliver(token, chatID, senderName, planID, text)
		return
	}

	switch text {
	case "/start":
		bm.sendWelcomeMenu(token, chatID, senderName)

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
			bm.sendWelcomeMenu(token, chatID, senderName)
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

	// Answer callback to dismiss loading indicator
	_, _ = bm.callAPI(token, "answerCallbackQuery", map[string]any{"callback_query_id": cbID})

	if strings.HasPrefix(data, "buy_") {
		planID := strings.TrimPrefix(data, "buy_")
		bm.initiatePurchase(token, chatID, planID)
	} else if data == "check_join" {
		bm.sendPlansMenu(token, chatID)
	} else if data == "menu_plans" {
		bm.sendPlansMenu(token, chatID)
	}
}

func (bm *TelegramBotManager) sendWelcomeMenu(token string, chatID int64, name string) {
	State.mu.RLock()
	sellCfg := getBotSellConfigLocked()
	welcome := gs(sellCfg, "welcome_text")
	State.mu.RUnlock()

	if welcome == "" {
		welcome = fmt.Sprintf("سلام <b>%s</b> عزیز! 🌟\nبه پنل هوشمند خرید اشتراک خوش آمدید.\nجهت شروع یکی از گزینه‌های زیر را انتخاب فرمایید:", name)
	}

	menu := map[string]any{
		"keyboard": [][]map[string]any{
			{{"text": "🛍 خرید اشتراک"}, {"text": "👤 حساب کاربری من"}},
			{{"text": "📊 وضعیت سرورها"}, {"text": "📞 پشتیبانی"}},
		},
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
		msg += fmt.Sprintf("💳 <b>اطلاعات پرداخت:</b>\n%s\n\n", payDetails)
	}
	if payUrl != "" {
		msg += fmt.Sprintf("🌐 <a href=\"%s\">جهت پرداخت آنلاین اینجا کلیک کنید</a>\n\n", payUrl)
	}

	msg += "📸 <i>لطفاً پس از واریز، تصویر فیش پرداخت یا شماره پیگیری تراکنش را در همین چت ارسال نمایید تا کانفیگ شما به صورت آنی تحویل داده شود.</i>"

	_, _ = bm.callAPI(token, "sendMessage", map[string]any{
		"chat_id":    chatID,
		"text":       msg,
		"parse_mode": "HTML",
	})
}

func (bm *TelegramBotManager) completeOrderAndDeliver(token string, chatID int64, senderName, planID, receiptNote string) {
	State.mu.Lock()
	plan := State.BotPlans[planID]
	sellCfg := getBotSellConfigLocked()
	adminChatID := gs(sellCfg, "admin_chat_id")
	State.mu.Unlock()

	if plan == nil {
		_, _ = bm.callAPI(token, "sendMessage", map[string]any{
			"chat_id": chatID,
			"text":    "❌ پلن معتبر نیست.",
		})
		return
	}

	prefix := plan.UsernamePrefix
	if prefix == "" {
		prefix = "shop"
	}
	uname := fmt.Sprintf("%s_%d_%s", prefix, chatID%10000, generateShortID()[:4])
	trafficBytes := int64(plan.TrafficGB * 1024 * 1024 * 1024)
	expireDays := plan.ExpireDays
	if expireDays <= 0 {
		expireDays = 30
	}
	expireAt := time.Now().Add(time.Duration(expireDays) * 24 * time.Hour).Format(time.RFC3339)

	inbounds := []string{}
	if plan.InboundID != "" {
		inbounds = append(inbounds, plan.InboundID)
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
		Note:              fmt.Sprintf("Telegram Order: %d (%s) Receipt: %s", chatID, senderName, receiptNote),
		CreatedAt:         time.Now().Format(time.RFC3339),
	}

	State.mu.Lock()
	State.Users[u.UserID] = u
	orders := getBotOrdersLocked()
	orders["approved"] = orders["approved"] + 1
	State.Settings["bot_orders"] = orders
	State.mu.Unlock()
	saveState()
	scheduleXrayApply()

	// Generate configs
	cfgs := generateUserConfigs(nil, u)
	cfgDirect := ""
	if len(cfgs) > 0 {
		cfgDirect = cfgs[0]
	}

	panelDomain := ""
	State.mu.RLock()
	if d, ok := State.Settings["domain"].(string); ok && d != "" {
		panelDomain = d
	}
	State.mu.RUnlock()

	subURL := ""
	if panelDomain != "" {
		subURL = fmt.Sprintf("https://%s/sub/%s", panelDomain, u.ConfigUUID)
	}

	deliveryMsg := fmt.Sprintf("🎉 <b>پرداخت شما تایید شد و سرویس اختصاصی شما فعال گردید!</b>\n\n"+
		"👤 نام کاربری: <code>%s</code>\n"+
		"📊 حجم پلن: <b>%.0f گیگابایت</b>\n"+
		"⏳ مدت اعتبار: <b>%d روز</b>\n\n",
		uname, plan.TrafficGB, expireDays)

	if subURL != "" {
		deliveryMsg += fmt.Sprintf("🔗 <b>لینک اشتراک خودکار (Subscription):</b>\n<code>%s</code>\n\n", subURL)
	}
	if cfgDirect != "" {
		deliveryMsg += fmt.Sprintf("🔑 <b>کانفیگ مستقیم (VLESS):</b>\n<code>%s</code>\n\n", cfgDirect)
	}
	deliveryMsg += "📌 <i>کانفیگ یا لینک بالا را کپی و در نرم‌افزار v2rayNG / Streisand / Sing-box ایمپورت نمایید.</i>"

	_, _ = bm.callAPI(token, "sendMessage", map[string]any{
		"chat_id":    chatID,
		"text":       deliveryMsg,
		"parse_mode": "HTML",
	})

	// Notify Admin if configured
	if adminChatID != "" {
		adminAlert := fmt.Sprintf("🔔 <b>سفارش جدید تلگرام با موفقیت ثبت و تحویل شد!</b>\n\n"+
			"👤 مشتری: <b>%s</b> (ID: <code>%d</code>)\n"+
			"📦 پلن: <b>%s</b> (مبلغ: %s تومان)\n"+
			"🏷 نام کاربری ساخته‌شده: <code>%s</code>\n"+
			"📝 اطلاعات رسید: <code>%s</code>",
			senderName, chatID, plan.Name, plan.Price, uname, receiptNote)

		_, _ = bm.callAPI(token, "sendMessage", map[string]any{
			"chat_id":    adminChatID,
			"text":       adminAlert,
			"parse_mode": "HTML",
		})
	}
}

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
