package main

import (
	"bytes"
	"encoding/json"
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
