package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestInboundAdvancedSettings(t *testing.T) {
	ib := &Inbound{
		InboundID:   "test-ib",
		Name:        "Test Inbound",
		Protocol:    "vless",
		Network:     "ws",
		Security:    "tls",
		SNI:         "custom.sni.com",
		Host:        "custom.host.com",
		Path:        "/my-custom-path",
		Fingerprint: "firefox",
		ALPN:        "http/1.1",
		Flow:        "xtls-rprx-vision",
		AdvancedSettings: map[string]any{
			"headerType": "none",
			"alpn":       "h2",
		},
	}

	q := url.Values{}
	q = applyInboundQuery(q, ib)

	if q.Get("sni") != "custom.sni.com" {
		t.Errorf("Expected sni to be custom.sni.com, got %s", q.Get("sni"))
	}
	if q.Get("host") != "custom.host.com" {
		t.Errorf("Expected host to be custom.host.com, got %s", q.Get("host"))
	}
	if q.Get("path") != "/my-custom-path" {
		t.Errorf("Expected path to be /my-custom-path, got %s", q.Get("path"))
	}
	if q.Get("fp") != "firefox" {
		t.Errorf("Expected fp to be firefox, got %s", q.Get("fp"))
	}
	if q.Get("alpn") != "h2" {
		t.Errorf("Expected overridden alpn to be h2, got %s", q.Get("alpn"))
	}
	if q.Get("headerType") != "none" {
		t.Errorf("Expected headerType to be none, got %s", q.Get("headerType"))
	}
}

func TestRealityNoForcedALPN(t *testing.T) {
	u := &User{
		UserID:     "u1",
		Username:   "testuser",
		ConfigUUID: "11111111-2222-3333-4444-555555555555",
		InboundIDs: []string{"reality-ib"},
		Status:     "active",
	}

	priv, pub, err := generateX25519Keypair()
	if err != nil {
		t.Fatalf("Failed keygen: %v", err)
	}

	ib := &Inbound{
		InboundID:      "reality-ib",
		Name:           "Reality TCP",
		Protocol:       "vless",
		Security:       "reality",
		Network:        "tcp",
		ExternalDomain: "myreality.com",
		ExternalPort:   443,
		SNI:            "www.apple.com",
		RealitySettings: map[string]any{
			"private_key": priv,
			"public_key":  pub,
			"short_id":    "1234abcd",
			"sni":         "www.apple.com",
			"dest":        "www.apple.com:443",
		},
	}

	State.mu.Lock()
	State.Users[u.UserID] = u
	State.Inbounds[ib.InboundID] = ib
	State.mu.Unlock()

	cfg := generateUserConfig(nil, u, ib.InboundID)
	if !strings.HasPrefix(cfg, "vless://") {
		t.Fatalf("Expected vless config, got: %s", cfg)
	}

	if strings.Contains(cfg, "alpn=h2%2Chttp%2F1.1") || strings.Contains(cfg, "alpn=h2") {
		t.Errorf("Reality TCP should not force alpn=h2, config: %s", cfg)
	}
	if !strings.Contains(cfg, "sni=www.apple.com") {
		t.Errorf("Reality config missing sni: %s", cfg)
	}
}

func TestBotStatusAndConfig(t *testing.T) {
	// Test Bot Status
	req := httptest.NewRequest("GET", "/api/bot/status", nil)
	w := httptest.NewRecorder()
	handleBotStatus(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var res map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if ok, _ := res["ok"].(bool); !ok {
		t.Fatalf("Expected ok: true")
	}

	// Test Bot Sell Plan Creation
	planPayload := map[string]any{
		"name":            "VIP 1 Month",
		"price":           "100000",
		"traffic_gb":      50.0,
		"expire_days":     30,
		"inbound_id":      "default",
		"username_prefix": "vip",
		"enabled":         true,
	}
	body, _ := json.Marshal(planPayload)
	reqPlan := httptest.NewRequest("POST", "/api/bot/sell/plans", bytes.NewReader(body))
	wPlan := httptest.NewRecorder()
	handleBotSellPlans(wPlan, reqPlan)

	if wPlan.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", wPlan.Code)
	}

	var planRes map[string]any
	if err := json.Unmarshal(wPlan.Body.Bytes(), &planRes); err != nil {
		t.Fatalf("Failed to parse plan json: %v", err)
	}
	planObj, _ := planRes["plan"].(map[string]any)
	planID, _ := planObj["id"].(string)
	if planID == "" {
		t.Fatalf("Expected valid plan ID, got empty")
	}

	// Test Bot Sell Plan Update (PATCH)
	updatePayload := map[string]any{
		"name":  "VIP 1 Month Updated",
		"price": "120000",
	}
	upBody, _ := json.Marshal(updatePayload)
	reqUp := httptest.NewRequest("PATCH", "/api/bot/sell/plans/"+planID, bytes.NewReader(upBody))
	wUp := httptest.NewRecorder()
	handleBotSellPlanByID(wUp, reqUp)

	if wUp.Code != http.StatusOK {
		t.Fatalf("Expected status 200 on plan patch, got %d", wUp.Code)
	}

	// Test Bot Sell Plan Delete
	reqDel := httptest.NewRequest("DELETE", "/api/bot/sell/plans/"+planID, nil)
	wDel := httptest.NewRecorder()
	handleBotSellPlanByID(wDel, reqDel)

	if wDel.Code != http.StatusOK {
		t.Fatalf("Expected status 200 on plan delete, got %d", wDel.Code)
	}
}


func TestHealthzNoConfidentialData(t *testing.T) {
	State.mu.Lock()
	State.Settings["bot_token"] = "123456789:AAFakeSecretBotTokenNotToLeak"
	State.Settings["panel_api_key"] = "SuperSecretPanelApiKey999"
	State.Password = "SuperSecretPasswordHash"
	State.mu.Unlock()

	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()

	// Simulate healthz mux handler
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		BotManager.mu.RLock()
		botPolling := BotManager.isPolling
		botScheduling := BotManager.isScheduling
		BotManager.mu.RUnlock()

		State.mu.RLock()
		botToken := getBotTokenLocked()
		botConfigured := botToken != ""
		totalUsers := len(State.Users)
		activeUsers := 0
		for _, u := range State.Users {
			if isActiveUser(u) {
				activeUsers++
			}
		}
		totalInbounds := 0
		activeInbounds := 0
		for _, ib := range State.Inbounds {
			if !ib.System && ib.InboundID != "Node" {
				totalInbounds++
				if ib.Port > 0 {
					activeInbounds++
				}
			}
		}
		State.mu.RUnlock()

		xrayUp := xrayIsRunning()
		mtproxyCount := countActiveMTProxies()

		botStatus := "stopped"
		if botPolling || botScheduling {
			botStatus = "running"
		} else if botConfigured {
			botStatus = "idle"
		}

		xrayStatus := "stopped"
		if xrayUp {
			xrayStatus = "running"
		}

		mtproxyStatus := "stopped"
		if mtproxyCount > 0 {
			mtproxyStatus = "running"
		}

		overallStatus := "healthy"
		if !xrayUp && activeInbounds > 0 {
			overallStatus = "degraded"
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"status":         overallStatus,
			"service":        "WhitePanel",
			"uptime_seconds": int(time.Since(serverStartTime).Seconds()),
			"services": map[string]any{
				"panel": map[string]any{
					"status": "running",
					"active": true,
					"port":   "8080",
				},
				"xray": map[string]any{
					"status": xrayStatus,
					"active": xrayUp,
				},
				"telegram_bot": map[string]any{
					"status":     botStatus,
					"configured": botConfigured,
					"polling":    botPolling,
					"scheduler":  botScheduling,
				},
				"mtproxy": map[string]any{
					"status":           mtproxyStatus,
					"active_instances": mtproxyCount,
				},
			},
			"stats": map[string]any{
				"total_inbounds":  totalInbounds,
				"active_inbounds": activeInbounds,
				"total_users":     totalUsers,
				"active_users":    activeUsers,
			},
		})
	})

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", w.Code)
	}

	bodyStr := w.Body.String()

	// Assert confidential data is NOT present
	confidentialSnippets := []string{
		"123456789",
		"AAFakeSecretBotTokenNotToLeak",
		"SuperSecretPanelApiKey999",
		"SuperSecretPasswordHash",
		"password",
		"api_key",
		"private_key",
		"short_id",
		"config_uuid",
	}

	for _, s := range confidentialSnippets {
		if strings.Contains(bodyStr, s) {
			t.Errorf("Security Leak: healthz contains confidential string %q: %s", s, bodyStr)
		}
	}

	// Assert service statuses are reported
	var res map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}

	if res["status"] == nil || res["services"] == nil || res["stats"] == nil {
		t.Errorf("Missing expected healthz fields in response: %s", bodyStr)
	}

	services, _ := res["services"].(map[string]any)
	if services["panel"] == nil || services["xray"] == nil || services["telegram_bot"] == nil || services["mtproxy"] == nil {
		t.Errorf("Missing expected sub-services in healthz: %s", bodyStr)
	}
}


func TestXrayManagerEndpoints(t *testing.T) {
	// Status
	req := httptest.NewRequest("GET", "/api/xray/status", nil)
	w := httptest.NewRecorder()
	handleXrayStatus(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", w.Code)
	}

	var res map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("Failed to decode json: %v", err)
	}

	st, ok := res["status"].(map[string]any)
	if !ok || st["version"] == nil || st["status_text"] == nil {
		t.Errorf("Missing version or status_text in xray status: %v", res)
	}

	// Logs
	reqLogs := httptest.NewRequest("GET", "/api/xray/logs?limit=10", nil)
	wLogs := httptest.NewRecorder()
	handleXrayLogs(wLogs, reqLogs)

	if wLogs.Code != http.StatusOK {
		t.Fatalf("Expected 200 for logs, got %d", wLogs.Code)
	}
}


func TestAdminApprovalWorkflow(t *testing.T) {
	// 1. Add admin
	adminChatID := int64(987654321)
	addBotAdmin(adminChatID)

	if !isBotAdmin(adminChatID) {
		t.Fatalf("Expected chatID %d to be recognized as bot admin", adminChatID)
	}

	// 2. Setup a plan
	plan := &BotPlan{
		ID:         "test-plan-adm",
		Name:       "Test Gold",
		Price:      "200000",
		TrafficGB:  50.0,
		ExpireDays: 30,
		Enabled:    true,
	}
	State.mu.Lock()
	State.BotPlans[plan.ID] = plan
	State.mu.Unlock()

	// 3. Create a pending order
	order := &BotOrder{
		ID:          "ORD-12345",
		ChatID:      11223344,
		SenderName:  "Test Buyer",
		Username:    "testbuyer",
		PlanID:      plan.ID,
		PlanName:    plan.Name,
		Price:       plan.Price,
		TrafficGB:   plan.TrafficGB,
		ExpireDays:  plan.ExpireDays,
		ReceiptType: "text",
		ReceiptText: "Track 123456",
		Status:      "pending",
		CreatedAt:   time.Now().Format("2006-01-02 15:04:05"),
	}
	saveOrder(order)

	// 4. Simulate Admin approving order
	BotManager.handleAdminApprovalAction("fake-token", adminChatID, "admin_user", "approve", order.ID, nil)

	botOrdersMu.RLock()
	updatedOrder := botOrdersDB[order.ID]
	botOrdersMu.RUnlock()

	if updatedOrder.Status != "approved" {
		t.Errorf("Expected order status approved, got %s", updatedOrder.Status)
	}
	if updatedOrder.GeneratedUser == "" {
		t.Errorf("Expected generated user not to be empty")
	}

	// 5. Verify user was created in State
	State.mu.RLock()
	var createdUser *User
	for _, u := range State.Users {
		if u.Username == updatedOrder.GeneratedUser {
			createdUser = u
			break
		}
	}
	State.mu.RUnlock()

	if createdUser == nil {
		t.Fatalf("User %s was not found in State.Users", updatedOrder.GeneratedUser)
	}
	if createdUser.TrafficLimitBytes != int64(50*1024*1024*1024) {
		t.Errorf("Traffic limit mismatch: %d", createdUser.TrafficLimitBytes)
	}
}

func TestBotDomainNot127001(t *testing.T) {
	State.mu.Lock()
	State.Settings["bot_domain"] = "bot-vpn.myserver.com"
	delete(State.Settings, "domain")
	ib := &Inbound{
		InboundID: "ib-dom-test",
		Protocol:  "vless",
		Network:   "ws",
		Security:  "tls",
		Port:      443,
	}
	State.Inbounds[ib.InboundID] = ib
	u := &User{
		UserID:            "user-dom-1",
		Username:          "domuser",
		Status:            "active",
		ConfigUUID:        "uuid-dom-test-1111",
		InboundIDs:        []string{ib.InboundID},
		TrafficLimitBytes: 10 * 1024 * 1024 * 1024,
	}
	State.Users[u.UserID] = u
	State.mu.Unlock()

	cfgs := generateUserConfigs(nil, u)
	if len(cfgs) == 0 {
		t.Fatalf("Expected at least 1 config")
	}
	cfg := cfgs[0]
	if strings.Contains(cfg, "127.0.0.1") {
		t.Fatalf("Config should NOT contain 127.0.0.1 when bot_domain is configured, got: %s", cfg)
	}
	if !strings.Contains(cfg, "bot-vpn.myserver.com") {
		t.Fatalf("Config should contain bot-vpn.myserver.com, got: %s", cfg)
	}
}

func TestSubscriptionBrowserVsClient(t *testing.T) {
	State.mu.Lock()
	ib := &Inbound{
		InboundID: "ib-sub-test",
		Protocol:  "vless",
		Network:   "ws",
		Security:  "tls",
		Port:      443,
	}
	State.Inbounds[ib.InboundID] = ib
	u := &User{
		UserID:            "sub-user-test-99",
		Username:          "subtester",
		Status:            "active",
		ConfigUUID:        "sub-uuid-abc-xyz-777",
		InboundIDs:        []string{ib.InboundID},
		TrafficLimitBytes: 20 * 1024 * 1024 * 1024,
	}
	State.Users[u.UserID] = u
	State.mu.Unlock()

	// 1. Browser request (User-Agent: Chrome/Mozilla, Accept: text/html)
	reqBrowser := httptest.NewRequest("GET", "/sub/"+u.ConfigUUID, nil)
	reqBrowser.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120.0.0.0 Safari/537.36")
	reqBrowser.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	wBrowser := httptest.NewRecorder()
	handleSubPage(wBrowser, reqBrowser, u.ConfigUUID)

	if wBrowser.Code != http.StatusOK {
		t.Fatalf("Browser expected 200, got %d", wBrowser.Code)
	}
	ctBrowser := wBrowser.Header().Get("Content-Type")
	if !strings.Contains(ctBrowser, "text/html") {
		t.Fatalf("Browser expected text/html, got %s", ctBrowser)
	}

	// 2. Client App request (v2rayNG)
	reqApp := httptest.NewRequest("GET", "/sub/"+u.ConfigUUID, nil)
	reqApp.Header.Set("User-Agent", "v2rayNG/1.8.5 (Linux; Android 13)")
	reqApp.Header.Set("Accept", "*/*")
	wApp := httptest.NewRecorder()
	handleSubPage(wApp, reqApp, u.ConfigUUID)

	if wApp.Code != http.StatusOK {
		t.Fatalf("Client App expected 200, got %d", wApp.Code)
	}
	ctApp := wApp.Header().Get("Content-Type")
	if !strings.Contains(ctApp, "text/plain") {
		t.Fatalf("Client App expected text/plain, got %s", ctApp)
	}
	// Verify response is base64 and decodes to vless configs
	bodyStr := wApp.Body.String()
	decoded, err := base64.StdEncoding.DecodeString(bodyStr)
	if err != nil {
		t.Fatalf("Failed to decode base64 configs from subscription: %v", err)
	}
	if !strings.Contains(string(decoded), "vless://") {
		t.Fatalf("Decoded configs should contain vless://, got: %s", string(decoded))
	}

	// 3. Subscription-Userinfo header exists
	subInfo := wApp.Header().Get("Subscription-Userinfo")
	if subInfo == "" || !strings.Contains(subInfo, "total=") {
		t.Fatalf("Expected Subscription-Userinfo header, got: %s", subInfo)
	}
}

func TestBackupAndRestore(t *testing.T) {
	// 1. Populate state with specific user, inbound, plan, node
	State.mu.Lock()
	testUser := &User{
		UserID:            "bkp-user-1",
		Username:          "backup_tester",
		Status:            "active",
		ConfigUUID:        "bkp-uuid-9999",
		TrafficLimitBytes: 50 * 1024 * 1024 * 1024,
		TrafficUsedBytes:  2 * 1024 * 1024 * 1024,
	}
	State.Users[testUser.UserID] = testUser

	testIb := &Inbound{
		InboundID: "bkp-ib-1",
		Name:      "Backup Inbound",
		Protocol:  "vless",
		Network:   "ws",
		Security:  "tls",
		Port:      8443,
	}
	State.Inbounds[testIb.InboundID] = testIb

	testPlan := &BotPlan{
		ID:         "bkp-plan-1",
		Name:       "Super Backup Plan",
		Price:      "500000",
		TrafficGB:  100,
		ExpireDays: 60,
		Enabled:    true,
	}
	State.BotPlans[testPlan.ID] = testPlan

	testNode := &Node{
		NodeID: "bkp-node-1",
		Name:   "German Node",
		Domain: "de.example.com",
	}
	State.Nodes[testNode.NodeID] = testNode

	State.Settings["test_custom_setting"] = "verified_123"
	State.mu.Unlock()

	// 2. Download Backup
	reqBkp := httptest.NewRequest("GET", "/api/settings/backup", nil)
	wBkp := httptest.NewRecorder()
	handleSettingsBackup(wBkp, reqBkp)

	if wBkp.Code != http.StatusOK {
		t.Fatalf("Backup download expected 200, got %d", wBkp.Code)
	}
	cd := wBkp.Header().Get("Content-Disposition")
	if !strings.Contains(cd, "whitepanel-backup-") {
		t.Errorf("Expected backup filename in Content-Disposition, got: %s", cd)
	}

	backupJSON := wBkp.Body.Bytes()
	var bkpMap map[string]any
	if err := json.Unmarshal(backupJSON, &bkpMap); err != nil {
		t.Fatalf("Invalid JSON produced by backup: %v", err)
	}
	if bkpMap["backup_format"] != "WhitePanel" {
		t.Errorf("Expected backup_format WhitePanel, got %v", bkpMap["backup_format"])
	}

	// 3. Clear State to simulate empty panel / fresh install
	State.mu.Lock()
	State.Users = make(map[string]*User)
	State.Inbounds = make(map[string]*Inbound)
	State.Nodes = make(map[string]*Node)
	State.BotPlans = make(map[string]*BotPlan)
	delete(State.Settings, "test_custom_setting")
	State.mu.Unlock()

	// 4. Restore via multipart form
	var b bytes.Buffer
	mw := multipart.NewWriter(&b)
	part, err := mw.CreateFormFile("file", "my-backup.json")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}
	part.Write(backupJSON)
	mw.Close()

	reqRestore := httptest.NewRequest("POST", "/api/settings/restore", &b)
	reqRestore.Header.Set("Content-Type", mw.FormDataContentType())
	wRestore := httptest.NewRecorder()
	handleSettingsRestore(wRestore, reqRestore)

	if wRestore.Code != http.StatusOK {
		t.Fatalf("Restore expected 200, got %d: %s", wRestore.Code, wRestore.Body.String())
	}

	var restoreResp map[string]any
	json.Unmarshal(wRestore.Body.Bytes(), &restoreResp)
	if restoreResp["ok"] != true {
		t.Fatalf("Restore response ok!=true: %v", restoreResp)
	}

	// 5. Verify all data was completely restored in State
	State.mu.RLock()
	restoredUser := State.Users[testUser.UserID]
	restoredIb := State.Inbounds[testIb.InboundID]
	restoredPlan := State.BotPlans[testPlan.ID]
	restoredNode := State.Nodes[testNode.NodeID]
	customSetting := State.Settings["test_custom_setting"]
	State.mu.RUnlock()

	if restoredUser == nil || restoredUser.Username != "backup_tester" {
		t.Fatalf("User was not properly restored: %v", restoredUser)
	}
	if restoredUser.ConfigUUID != "bkp-uuid-9999" {
		t.Errorf("User ConfigUUID mismatch: %s", restoredUser.ConfigUUID)
	}
	if restoredIb == nil || restoredIb.Port != 8443 {
		t.Fatalf("Inbound was not properly restored: %v", restoredIb)
	}
	if restoredPlan == nil || restoredPlan.TrafficGB != 100 {
		t.Fatalf("Bot Plan was not properly restored: %v", restoredPlan)
	}
	if restoredNode == nil || restoredNode.Domain != "de.example.com" {
		t.Fatalf("Node was not properly restored: %v", restoredNode)
	}
	if customSetting != "verified_123" {
		t.Fatalf("Setting was not properly restored: %v", customSetting)
	}
}

func TestCleanIPConfigGeneration(t *testing.T) {
	State.mu.Lock()
	ib := &Inbound{
		InboundID:      "ib-clean-test",
		Name:           "TLS WS Inbound",
		Protocol:       "vless",
		Network:        "ws",
		Security:       "tls",
		Port:           443,
		ExternalPort:   443,
		Domain:         "main.server.com",
		ExternalDomain: "main.server.com",
	}
	State.Inbounds[ib.InboundID] = ib

	u := &User{
		UserID:           "u-clean-1",
		Username:         "clean_user",
		ConfigUUID:       "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		InboundIDs:       []string{ib.InboundID},
		Status:           "active",
		ProxyIPEnabled:   true,
		CustomIPInbounds: map[string][]string{"cf": {ib.InboundID}},
	}
	State.Users[u.UserID] = u
	State.mu.Unlock()

	req := httptest.NewRequest("GET", "/sub/"+u.ConfigUUID, nil)
	cfgs := generateUserConfigs(req, u)

	if len(cfgs) <= 1 {
		t.Fatalf("Expected multiple configs including clean IPs, got %d configs", len(cfgs))
	}

	foundClean := false
	for _, c := range cfgs {
		if strings.Contains(c, "Clean CF") || strings.Contains(c, "CF") {
			foundClean = true
			if !strings.Contains(c, "sni=main.server.com") {
				t.Errorf("Clean IP config lost SNI parameter: %s", c)
			}
			if !strings.Contains(c, "host=main.server.com") {
				t.Errorf("Clean IP config lost host parameter: %s", c)
			}
		}
	}

	if !foundClean {
		t.Errorf("Clean CF config was not found in generated configs: %v", cfgs)
	}
}

func TestTrafficAccountingAccurate(t *testing.T) {
	State.mu.Lock()
	u := &User{
		UserID:            "u-traffic-1",
		Username:          "traffic_user",
		ConfigUUID:        "7ba7b810-9dad-11d1-80b4-00c04fd430c9",
		Status:            "active",
		TrafficLimitBytes: 10 * 1024 * 1024 * 1024,
		TrafficUsedBytes:  0,
	}
	State.Users[u.UserID] = u
	trafficDirty.Store(false)
	State.mu.Unlock()

	// 1. Uplink test
	uplinkBytes := int64(102400)
	if !checkAndUse(u.ConfigUUID, uplinkBytes) {
		t.Fatalf("checkAndUse failed for uplink")
	}

	// 2. Downlink test
	downlinkBytes := int64(5242880)
	if !checkAndUse(u.ConfigUUID, downlinkBytes) {
		t.Fatalf("checkAndUse failed for downlink")
	}

	State.mu.RLock()
	totalUsed := u.TrafficUsedBytes
	State.mu.RUnlock()

	if totalUsed != uplinkBytes+downlinkBytes {
		t.Fatalf("Expected TrafficUsedBytes %d, got %d", uplinkBytes+downlinkBytes, totalUsed)
	}

	if !trafficDirty.Load() {
		t.Fatalf("Expected trafficDirty flag to be set to true")
	}
}

func TestXrayFlowValidationTCPOnly(t *testing.T) {
	// TCP inbound with Flow
	tcpIb := &Inbound{
		InboundID: "ib-tcp-test",
		Protocol:  "vless",
		Network:   "tcp",
		Security:  "reality",
		Flow:      "xtls-rprx-vision",
	}

	// WS inbound with Flow
	wsIb := &Inbound{
		InboundID: "ib-ws-test",
		Protocol:  "vless",
		Network:   "ws",
		Security:  "tls",
		Flow:      "xtls-rprx-vision",
	}

	qTCP := applyInboundQuery(url.Values{"type": []string{"tcp"}}, tcpIb)
	if qTCP.Get("flow") != "xtls-rprx-vision" {
		t.Fatalf("Expected flow to remain on TCP, got %s", qTCP.Get("flow"))
	}

	qWS := applyInboundQuery(url.Values{"type": []string{"ws"}}, wsIb)
	if qWS.Get("flow") != "" {
		t.Fatalf("Expected flow to be removed on WS, got %s", qWS.Get("flow"))
	}
}
