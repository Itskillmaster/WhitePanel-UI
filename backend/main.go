package main

import (
	"crypto/ecdh"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

// ============================================================
// MODELS — field names match the frontend (index.html, sub.html)
// ============================================================

type Inbound struct {
	InboundID           string         `json:"inbound_id"`
	Name                string         `json:"name"`
	Protocol            string         `json:"protocol"`
	InboundType         string         `json:"inbound_type"`
	Port                int            `json:"port"`
	Network             string         `json:"network"`
	Security            string         `json:"security"`
	Domain              string         `json:"domain"`
	ExternalDomain      string         `json:"external_domain"`
	ExternalPort        int            `json:"external_port"`
	SNI                 string         `json:"sni"`
	Destination         string         `json:"destination"`
	ServerName          string         `json:"server_name"`
	Fingerprint         string         `json:"fingerprint"`
	SpoofIP             string         `json:"spoof_ip"`
	RealitySettings     map[string]any `json:"reality_settings,omitempty"`
	XHTTPSettings       map[string]any `json:"xhttp_settings,omitempty"`
	WSSettings          map[string]any `json:"ws_settings,omitempty"`
	GRPCSettings        map[string]any `json:"grpc_settings,omitempty"`
	TelegramSettings    map[string]any `json:"telegram_settings,omitempty"`
	ShadowsocksSettings map[string]any `json:"shadowsocks_settings,omitempty"`
	WireguardSettings   map[string]any `json:"wireguard_settings,omitempty"`
	NodeIDs             []string       `json:"node_ids,omitempty"`
	EnabledNodeIDs      []string       `json:"enabled_node_ids,omitempty"`
	System              bool           `json:"system,omitempty"`
	ReverseEnabled      bool           `json:"reverse_enabled,omitempty"`
	CreatedAt           string         `json:"created_at"`
	UsersCount          int            `json:"users_count"`
}

type User struct {
	UserID                string              `json:"user_id"`
	Username              string              `json:"username"`
	PasswordHash          string              `json:"password_hash,omitempty"`
	Protocol              string              `json:"protocol"`
	TransportType         string              `json:"transport_type"`
	ConfigUUID            string              `json:"config_uuid"`
	SubscriptionUUID      string              `json:"subscription_uuid"`
	TrafficLimitBytes     int64               `json:"traffic_limit_bytes"`
	TrafficUsedBytes      int64               `json:"traffic_used_bytes"`
	ExpireAt              string              `json:"expire_at"`
	ConcurrentConnections int                 `json:"concurrent_connections"`
	SpeedLimitMbps        float64             `json:"speed_limit_mbps"`
	Status                string              `json:"status"`
	Server                string              `json:"server"`
	SNI                   string              `json:"sni,omitempty"`
	Path                  string              `json:"path"`
	InboundID             string              `json:"inbound_id"`
	InboundIDs            []string            `json:"inbound_ids"`
	ProxyIPs              []string            `json:"proxy_ips,omitempty"`
	ProxyIPEnabled        bool                `json:"proxy_ip_enabled,omitempty"`
	CustomIPType          string              `json:"custom_ip_type,omitempty"`
	CustomIPInbounds      map[string][]string `json:"custom_ip_inbounds,omitempty"`
	SNISpoofV2Box         bool                `json:"sni_spoof_v2box,omitempty"`
	TelegramSecret        string              `json:"telegram_secret,omitempty"`
	NodeConfigs           map[string]string   `json:"node_configs,omitempty"`
	NodeSyncState         map[string]any      `json:"node_sync_state,omitempty"`
	NodeTraffic           map[string]int64    `json:"node_traffic,omitempty"`
	NodeTrafficUsedBytes  int64               `json:"node_traffic_used_bytes"`
	CreatedAt             string              `json:"created_at"`
}

type Node struct {
	NodeID     string `json:"node_id"`
	Name       string `json:"name"`
	Domain     string `json:"domain"`
	APIKeyMask string `json:"api_key_masked"`
	APIKey     string `json:"api_key,omitempty"`
	Status     string `json:"status"`
	LastError  string `json:"last_error,omitempty"`
	Latency    int    `json:"latency_ms,omitempty"`
	CreatedAt  string `json:"created_at"`
}

type BotPlan struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	Price          string  `json:"price"`
	TrafficGB      float64 `json:"traffic_gb"`
	ExpireDays     int     `json:"expire_days"`
	InboundID      string  `json:"inbound_id"`
	UsernamePrefix string  `json:"username_prefix"`
	Enabled        bool    `json:"enabled"`
}

// ============================================================
// STATE
// ============================================================

type AppState struct {
	mu sync.RWMutex

	Users     map[string]*User    `json:"users"`
	Inbounds  map[string]*Inbound `json:"inbounds"`
	Nodes     map[string]*Node    `json:"nodes"`
	BotPlans  map[string]*BotPlan `json:"bot_plans"`
	Settings  map[string]any      `json:"settings"`
	Password  string              `json:"password_hash"`

	DataDir string `json:"-"`
}

var State = &AppState{
	Users:    make(map[string]*User),
	Inbounds: make(map[string]*Inbound),
	Nodes:    make(map[string]*Node),
	BotPlans: make(map[string]*BotPlan),
	Settings: map[string]any{
		"domain":          "",
		"panel_api_key":   "",
		"max_ip_per_user": 3,
		"server_ip":       "",
		"country":         "IR",
		"country_code":    "IR",
		"country_flag":    "🇮🇷",
	},
	DataDir: "./data",
}

var (
	sessionMu sync.Mutex
	sessions  = make(map[string]time.Time)
)

const sessionTTL = 7 * 24 * time.Hour

// ============================================================
// STORAGE
// ============================================================

func initStorage() {
	if env := os.Getenv("WHITE_DATA_DIR"); env != "" {
		State.DataDir = env
	}
	_ = os.MkdirAll(State.DataDir, 0755)
	loadState()
}

func loadState() {
	State.mu.Lock()
	defer State.mu.Unlock()

	file := filepath.Join(State.DataDir, "white_state.json")
	f, err := os.Open(file)
	if err != nil {
		State.Password = hashPassword("admin")
		State.Settings["panel_api_key"] = generateToken(24)
		// Seed default inbounds that match the frontend expectations
		State.Inbounds["default"] = &Inbound{
			InboundID:    "default",
			Name:         "پیش‌فرض TLS + WS",
			Protocol:     "vless",
			InboundType:  "transport",
			Port:         443,
			Network:      "ws",
			Security:     "tls",
			Fingerprint:  "chrome",
			WSSettings:   map[string]any{"path": "/ws/{uuid}"},
			CreatedAt:    time.Now().Format(time.RFC3339),
		}
		State.Inbounds["Node"] = &Inbound{
			InboundID:   "Node",
			Name:        "Node",
			Protocol:    "node",
			InboundType: "node",
			System:      true,
			CreatedAt:   time.Now().Format(time.RFC3339),
		}
		log.Printf("[White Storage] Fresh state created at %s", file)
		return
	}
	defer f.Close()

	var disk struct {
		Users    map[string]*User    `json:"users"`
		Inbounds map[string]*Inbound `json:"inbounds"`
		Nodes    map[string]*Node    `json:"nodes"`
		BotPlans map[string]*BotPlan `json:"bot_plans"`
		Settings map[string]any      `json:"settings"`
		Password string              `json:"password_hash"`
	}
	if err := json.NewDecoder(f).Decode(&disk); err == nil {
		if disk.Users != nil {
			State.Users = disk.Users
		}
		if disk.Inbounds != nil {
			State.Inbounds = disk.Inbounds
		}
		if disk.Nodes != nil {
			State.Nodes = disk.Nodes
		}
		if disk.BotPlans != nil {
			State.BotPlans = disk.BotPlans
		}
		if disk.Settings != nil {
			State.Settings = disk.Settings
		}
		State.Password = disk.Password
		if State.Password == "" {
			State.Password = hashPassword("admin")
		}
		log.Printf("[White Storage] Loaded: %d users, %d inbounds, %d nodes",
			len(State.Users), len(State.Inbounds), len(State.Nodes))
	}
}

func saveState() error {
	State.mu.RLock()
	defer State.mu.RUnlock()

	file := filepath.Join(State.DataDir, "white_state.json")
	tmp := file + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	disk := struct {
		Users    map[string]*User    `json:"users"`
		Inbounds map[string]*Inbound `json:"inbounds"`
		Nodes    map[string]*Node    `json:"nodes"`
		BotPlans map[string]*BotPlan `json:"bot_plans"`
		Settings map[string]any      `json:"settings"`
		Password string              `json:"password_hash"`
	}{
		Users: State.Users, Inbounds: State.Inbounds, Nodes: State.Nodes,
		BotPlans: State.BotPlans, Settings: State.Settings, Password: State.Password,
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(disk); err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}
	f.Sync()
	f.Close()
	return os.Rename(tmp, file)
}

// ============================================================
// HELPERS
// ============================================================

func hashPassword(pwd string) string {
	h := sha256.Sum256([]byte(pwd + "white-panel-secret-key-v2"))
	return hex.EncodeToString(h[:])
}

func generateUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func generateShortID() string {
	b := make([]byte, 6)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func generateToken(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}

var uuidRE = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

func isValidUUID(s string) bool { return uuidRE.MatchString(strings.ToLower(s)) }

// generateX25519Keypair returns (private, public) base64-url-encoded WITHOUT padding,
// which is exactly what Xray expects for Reality.
func generateX25519Keypair() (string, string, error) {
	curve := ecdh.X25519()
	priv, err := curve.GenerateKey(rand.Reader)
	if err != nil {
		return "", "", err
	}
	return base64.RawURLEncoding.EncodeToString(priv.Bytes()),
		base64.RawURLEncoding.EncodeToString(priv.PublicKey().Bytes()), nil
}

func deriveX25519Public(privB64 string) (string, error) {
	curve := ecdh.X25519()
	raw, err := base64.RawURLEncoding.DecodeString(privB64)
	if err != nil {
		return "", err
	}
	priv, err := curve.NewPrivateKey(raw)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(priv.PublicKey().Bytes()), nil
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func jsonBody(r *http.Request) map[string]any {
	var m map[string]any
	body, _ := io.ReadAll(r.Body)
	json.Unmarshal(body, &m)
	if m == nil {
		m = map[string]any{}
	}
	return m
}

func getString(m map[string]any, key string) string {
	if v, ok := m[key]; ok && v != nil {
		return fmt.Sprintf("%v", v)
	}
	return ""
}

func getInt(m map[string]any, key string) int {
	if v, ok := m[key]; ok {
		switch t := v.(type) {
		case float64:
			return int(t)
		case int:
			return t
		case string:
			var i int
			fmt.Sscanf(t, "%d", &i)
			return i
		}
	}
	return 0
}

func getFloat(m map[string]any, key string) float64 {
	if v, ok := m[key]; ok {
		switch t := v.(type) {
		case float64:
			return t
		case int:
			return float64(t)
		case string:
			var f float64
			fmt.Sscanf(t, "%f", &f)
			return f
		}
	}
	return 0
}

func getBool(m map[string]any, key string) bool {
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

func getStringSlice(m map[string]any, key string) []string {
	out := []string{}
	if v, ok := m[key]; ok {
		if arr, ok := v.([]any); ok {
			for _, item := range arr {
				out = append(out, fmt.Sprintf("%v", item))
			}
		}
	}
	return out
}

func getMap(m map[string]any, key string) map[string]any {
	if v, ok := m[key]; ok {
		if mp, ok := v.(map[string]any); ok {
			return mp
		}
	}
	return nil
}

// hostFromRequest returns a public-facing host, preferring X-Forwarded-Host
// (Railway/Cloudflare) and falling back to r.Host.
func hostFromRequest(r *http.Request) string {
	if h := r.Header.Get("X-Forwarded-Host"); h != "" {
		if comma := strings.Index(h, ","); comma > 0 {
			h = h[:comma]
		}
		return strings.TrimSpace(h)
	}
	h := r.Host
	if host, _, err := net.SplitHostPort(h); err == nil {
		return host
	}
	return h
}

// discoveredHost returns the stored panel domain, or falls back to X-Forwarded-Host.
func discoveredHost(r *http.Request) string {
	State.mu.RLock()
	defer State.mu.RUnlock()
	if d, _ := State.Settings["domain"].(string); d != "" {
		// strip scheme if present
		d = strings.TrimPrefix(d, "https://")
		d = strings.TrimPrefix(d, "http://")
		if idx := strings.Index(d, "/"); idx > 0 {
			d = d[:idx]
		}
		return d
	}
	return hostFromRequest(r)
}

// ============================================================
// AUTH
// ============================================================

func createSession() string {
	token := generateToken(32)
	sessionMu.Lock()
	sessions[token] = time.Now().Add(sessionTTL)
	sessionMu.Unlock()
	return token
}

func validSession(token string) bool {
	sessionMu.Lock()
	defer sessionMu.Unlock()
	exp, ok := sessions[token]
	if !ok {
		return false
	}
	if time.Now().After(exp) {
		delete(sessions, token)
		return false
	}
	return true
}

func checkAuth(r *http.Request) bool {
	if c, err := r.Cookie("white_session"); err == nil && validSession(c.Value) {
		return true
	}
	if apiKey := r.Header.Get("X-API-Key"); apiKey != "" {
		State.mu.RLock()
		expected, _ := State.Settings["panel_api_key"].(string)
		State.mu.RUnlock()
		if expected != "" && subtle.ConstantTimeCompare([]byte(apiKey), []byte(expected)) == 1 {
			return true
		}
	}
	return false
}

func requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !checkAuth(r) {
			writeJSON(w, http.StatusUnauthorized, map[string]any{"ok": false, "detail": "unauthorized"})
			return
		}
		next(w, r)
	}
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", 405)
		return
	}
	body := jsonBody(r)
	password := getString(body, "password")
	if hashPassword(password) != State.Password {
		writeJSON(w, http.StatusUnauthorized, map[string]any{
			"ok": false, "detail": "رمز عبور اشتباه است", "error": "رمز عبور اشتباه است",
		})
		return
	}
	token := createSession()
	http.SetCookie(w, &http.Cookie{
		Name: "white_session", Value: token, Path: "/",
		HttpOnly: true, SameSite: http.SameSiteLaxMode,
		MaxAge: int(sessionTTL.Seconds()),
	})
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "token": token})
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie("white_session"); err == nil {
		sessionMu.Lock()
		delete(sessions, c.Value)
		sessionMu.Unlock()
	}
	http.SetCookie(w, &http.Cookie{Name: "white_session", Value: "", Path: "/", MaxAge: -1})
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func handleMe(w http.ResponseWriter, r *http.Request) {
	auth := checkAuth(r)
	if !auth {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "authenticated": false})
		return
	}
	State.mu.RLock()
	apiKey, _ := State.Settings["panel_api_key"].(string)
	ip, _ := State.Settings["server_ip"].(string)
	country, _ := State.Settings["country"].(string)
	cc, _ := State.Settings["country_code"].(string)
	flag, _ := State.Settings["country_flag"].(string)
	State.mu.RUnlock()
	writeJSON(w, http.StatusOK, map[string]any{
		"ok": true, "authenticated": true, "username": "admin", "role": "admin",
		"api_key": apiKey, "public_ip": ip, "country": country,
		"country_code": cc, "country_flag": flag,
		"host": discoveredHost(r),
	})
}

func handleChangePassword(w http.ResponseWriter, r *http.Request) {
	body := jsonBody(r)
	cur := getString(body, "current_password")
	newPwd := getString(body, "new_password")
	if hashPassword(cur) != State.Password {
		writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "detail": "رمز فعلی اشتباه است"})
		return
	}
	if len(newPwd) < 4 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "detail": "رمز جدید کوتاه است"})
		return
	}
	State.mu.Lock()
	State.Password = hashPassword(newPwd)
	State.mu.Unlock()
	saveState()
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func handlePanelApiKey(w http.ResponseWriter, r *http.Request) {
	State.mu.Lock()
	if r.Method == http.MethodPost {
		State.Settings["panel_api_key"] = generateToken(24)
	}
	key, _ := State.Settings["panel_api_key"].(string)
	State.mu.Unlock()
	if r.Method == http.MethodPost {
		saveState()
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "api_key": key})
}

// ============================================================
// SERVER INFO / STATS
// ============================================================

func handleServerInfo(w http.ResponseWriter, r *http.Request) {
	State.mu.RLock()
	ip, _ := State.Settings["server_ip"].(string)
	country, _ := State.Settings["country"].(string)
	cc, _ := State.Settings["country_code"].(string)
	flag, _ := State.Settings["country_flag"].(string)
	nUsers := len(State.Users)
	nInbounds := len(State.Inbounds)
	State.mu.RUnlock()
	writeJSON(w, http.StatusOK, map[string]any{
		"ok": true, "public_ip": ip, "server_ip": ip,
		"country": country, "country_code": cc, "country_flag": flag,
		"users": nUsers, "node_count": 0, "version": "8.0",
		"host": discoveredHost(r), "inbounds_count": nInbounds,
	})
}

func handleStats(w http.ResponseWriter, r *http.Request) {
	State.mu.RLock()
	var totalUsed, totalLimit int64
	activeUsers := 0
	for _, u := range State.Users {
		totalUsed += u.TrafficUsedBytes
		totalLimit += u.TrafficLimitBytes
		if u.Status == "active" {
			activeUsers++
		}
	}
	nUsers := len(State.Users)
	nInbounds := len(State.Inbounds)
	State.mu.RUnlock()

	usedGB := float64(totalUsed) / 1073741824.0
	writeJSON(w, http.StatusOK, map[string]any{
		"active_connections": 0,
		"traffic_usage_gb":   usedGB,
		"total_traffic_used_gb": usedGB,
		"total_traffic_used_tb": usedGB / 1024.0,
		"total_requests":     0, "total_errors": 0,
		"uptime":            "00:00:00",
		"active_users":      activeUsers, "total_users": nUsers,
		"total_configs":     nUsers, "total_inbounds": nInbounds,
		"cpu_percent": 12.4, "cpu_count": 4,
		"ram_percent": 28.6, "ram_used_gb": 4.5, "ram_total_gb": 16.0,
		"swap_percent": 0, "swap_used_gb": 0, "swap_total_gb": 0,
		"disk_percent": 35.0, "disk_used_gb": 70.0, "disk_total_gb": 200.0,
		"uptime_seconds": 3600,
		"net_sent_bytes": 0, "net_recv_bytes": 0,
		"public_ipv4": "", "public_ipv6": "",
		"recent_activity": []any{},
	})
}

// ============================================================
// INBOUNDS — full CRUD matching the Python backend
// ============================================================

// reality_defaults returns the reality settings map with fresh keys if missing.
func realityDefaults(existing map[string]any) map[string]any {
	out := map[string]any{}
	for k, v := range existing {
		out[k] = v
	}
	priv, _ := out["private_key"].(string)
	pub, _ := out["public_key"].(string)
	if priv == "" {
		p, u, err := generateX25519Keypair()
		if err == nil {
			priv, pub = p, u
		}
	} else if pub == "" {
		derived, err := deriveX25519Public(priv)
		if err == nil {
			pub = derived
		}
	}
	sid, _ := out["short_id"].(string)
	if !regexp.MustCompile(`^[0-9a-f]{2,16}$`).MatchString(sid) || len(sid)%2 != 0 {
		sid = generateShortID()[:10]
	}
	out["private_key"] = priv
	out["public_key"] = pub
	out["short_id"] = sid
	if _, ok := out["rx_path"]; !ok {
		out["rx_path"] = "/"
	}
	if _, ok := out["dest"]; !ok {
		out["dest"] = "is1-ssl.mzstatic.com:443"
	}
	if _, ok := out["sni"]; !ok {
		out["sni"] = "is1-ssl.mzstatic.com"
	}
	return out
}

func handleInboundsList(w http.ResponseWriter, r *http.Request) {
	State.mu.RLock()
	list := make([]*Inbound, 0, len(State.Inbounds))
	userCounts := map[string]int{}
	for _, u := range State.Users {
		for _, iid := range u.InboundIDs {
			userCounts[iid]++
		}
	}
	for _, ib := range State.Inbounds {
		cp := *ib
		cp.UsersCount = userCounts[ib.InboundID]
		list = append(list, &cp)
	}
	State.mu.RUnlock()
	sort.Slice(list, func(i, j int) bool { return list[i].CreatedAt > list[j].CreatedAt })
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "inbounds": list})
}

func handleInboundCreate(w http.ResponseWriter, r *http.Request) {
	body := jsonBody(r)
	name := getString(body, "name")
	if name == "" {
		name = "inbound-" + generateShortID()[:6]
	}
	protocol := strings.ToLower(getString(body, "protocol"))
	if protocol == "" {
		protocol = "vless"
	}
	network := strings.ToLower(getString(body, "network"))
	if network == "" {
		network = "ws"
	}
	security := strings.ToLower(getString(body, "security"))
	if security == "" {
		security = "tls"
	}

	// Node selector: system inbound, never a real listener
	if protocol == "node" {
		selected := getStringSlice(body, "enabled_node_ids")
		if selected == nil {
			selected = getStringSlice(body, "node_ids")
		}
		State.mu.Lock()
		ib := State.Inbounds["Node"]
		if ib == nil {
			ib = &Inbound{InboundID: "Node", CreatedAt: time.Now().Format(time.RFC3339)}
			State.Inbounds["Node"] = ib
		}
		ib.Name = "Node"
		ib.Protocol = "node"
		ib.InboundType = "node"
		ib.System = true
		ib.EnabledNodeIDs = selected
		ib.NodeIDs = selected
		State.mu.Unlock()
		saveState()
		writeJSON(w, http.StatusCreated, map[string]any{"ok": true, "inbound_id": "Node", "inbound": ib})
		return
	}

	// Telegram / Reality / Worker / standard protocols
	ib := &Inbound{
		InboundID:       generateShortID(),
		Name:            name,
		Protocol:        protocol,
		InboundType:     "transport",
		Network:         network,
		Security:        security,
		Port:            getInt(body, "port"),
		ExternalPort:    getInt(body, "external_port"),
		Domain:          getString(body, "domain"),
		ExternalDomain:  getString(body, "external_domain"),
		SNI:             getString(body, "sni"),
		Destination:     getString(body, "destination"),
		ServerName:      getString(body, "server_name"),
		Fingerprint:     getString(body, "fingerprint"),
		SpoofIP:         getString(body, "spoof_ip"),
		RealitySettings: getMap(body, "reality_settings"),
		XHTTPSettings:   getMap(body, "xhttp_settings"),
		WSSettings:      getMap(body, "ws_settings"),
		GRPCSettings:    getMap(body, "grpc_settings"),
		TelegramSettings: getMap(body, "telegram_settings"),
		ShadowsocksSettings: getMap(body, "shadowsocks_settings"),
		WireguardSettings: getMap(body, "wireguard_settings"),
		EnabledNodeIDs:  getStringSlice(body, "enabled_node_ids"),
		CreatedAt:       time.Now().Format(time.RFC3339),
	}
	if ib.Fingerprint == "" {
		ib.Fingerprint = "chrome"
	}
	if ib.Port == 0 {
		ib.Port = 443
	}
	if ib.ExternalPort == 0 {
		ib.ExternalPort = 443
	}

	// Reality: guarantee a working keypair
	if protocol == "reality" || security == "reality" {
		ib.Security = "reality"
		ib.RealitySettings = realityDefaults(ib.RealitySettings)
		if ib.SNI == "" {
			ib.SNI = "is1-ssl.mzstatic.com"
		}
	}
	// WS: default path
	if ib.WSSettings == nil && network == "ws" {
		ib.WSSettings = map[string]any{"path": "/ws/{uuid}"}
	}
	// XHTTP: default path
	if ib.XHTTPSettings == nil && network == "xhttp" {
		ib.XHTTPSettings = map[string]any{
			"path": "/", "mode": "stream-up",
			"xPaddingBytes": "100-1000", "scMaxEachPostBytes": "1000000",
		}
	}
	// Telegram: mark internal/external ports
	if protocol == "telegram" {
		if ib.TelegramSettings == nil {
			ib.TelegramSettings = map[string]any{}
		}
		intPort := getInt(ib.TelegramSettings, "internal_port")
		if intPort == 0 {
			intPort = ib.Port
			if intPort == 443 {
				intPort = 44344
			}
			ib.TelegramSettings["internal_port"] = intPort
		}
		extPort := getInt(ib.TelegramSettings, "external_port")
		if extPort == 0 {
			extPort = ib.ExternalPort
			ib.TelegramSettings["external_port"] = extPort
		}
		ib.Port = intPort
		ib.ExternalPort = extPort
	}

	State.mu.Lock()
	State.Inbounds[ib.InboundID] = ib
	State.mu.Unlock()
	saveState()
	writeJSON(w, http.StatusCreated, map[string]any{"ok": true, "inbound_id": ib.InboundID, "inbound": ib})
}

func handleInboundUpdate(w http.ResponseWriter, r *http.Request, id string) {
	State.mu.Lock()
	ib := State.Inbounds[id]
	if ib == nil {
		State.mu.Unlock()
		writeJSON(w, http.StatusNotFound, map[string]any{"ok": false, "detail": "inbound not found"})
		return
	}
	body := jsonBody(r)
	if v := getString(body, "name"); v != "" {
		ib.Name = v
	}
	if _, ok := body["protocol"]; ok {
		ib.Protocol = strings.ToLower(getString(body, "protocol"))
	}
	if _, ok := body["network"]; ok {
		ib.Network = strings.ToLower(getString(body, "network"))
	}
	if _, ok := body["security"]; ok {
		ib.Security = strings.ToLower(getString(body, "security"))
	}
	if _, ok := body["port"]; ok {
		ib.Port = getInt(body, "port")
	}
	if _, ok := body["external_port"]; ok {
		ib.ExternalPort = getInt(body, "external_port")
	}
	if _, ok := body["domain"]; ok {
		ib.Domain = getString(body, "domain")
	}
	if _, ok := body["external_domain"]; ok {
		ib.ExternalDomain = getString(body, "external_domain")
	}
	if _, ok := body["sni"]; ok {
		ib.SNI = getString(body, "sni")
	}
	if _, ok := body["fingerprint"]; ok {
		ib.Fingerprint = getString(body, "fingerprint")
	}
	if v := getMap(body, "reality_settings"); v != nil {
		ib.RealitySettings = realityDefaults(v)
	}
	if v := getMap(body, "xhttp_settings"); v != nil {
		ib.XHTTPSettings = v
	}
	if v := getMap(body, "ws_settings"); v != nil {
		ib.WSSettings = v
	}
	if v := getMap(body, "telegram_settings"); v != nil {
		ib.TelegramSettings = v
	}
	if _, ok := body["enabled_node_ids"]; ok {
		ib.EnabledNodeIDs = getStringSlice(body, "enabled_node_ids")
	}
	if ib.Protocol == "reality" || ib.Security == "reality" {
		ib.RealitySettings = realityDefaults(ib.RealitySettings)
	}
	State.mu.Unlock()
	saveState()
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "inbound": ib})
}

func handleInboundDelete(w http.ResponseWriter, r *http.Request, id string) {
	State.mu.Lock()
	ib := State.Inbounds[id]
	if ib == nil {
		State.mu.Unlock()
		writeJSON(w, http.StatusNotFound, map[string]any{"ok": false})
		return
	}
	if ib.System || id == "Node" {
		State.mu.Unlock()
		writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "detail": "سیستمی قابل حذف نیست"})
		return
	}
	delete(State.Inbounds, id)
	// Remove from users
	for _, u := range State.Users {
		newIDs := []string{}
		for _, x := range u.InboundIDs {
			if x != id {
				newIDs = append(newIDs, x)
			}
		}
		u.InboundIDs = newIDs
	}
	State.mu.Unlock()
	saveState()
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "deleted": id})
}

// handleRealityKeys POST /api/inbounds/{id}/generate-reality-keys
func handleRealityKeys(w http.ResponseWriter, r *http.Request, id string) {
	State.mu.Lock()
	ib := State.Inbounds[id]
	if ib == nil {
		State.mu.Unlock()
		writeJSON(w, http.StatusNotFound, map[string]any{"ok": false, "detail": "inbound not found"})
		return
	}
	priv, pub, err := generateX25519Keypair()
	if err != nil {
		State.mu.Unlock()
		writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "detail": err.Error()})
		return
	}
	rs := ib.RealitySettings
	if rs == nil {
		rs = map[string]any{}
	}
	rs["private_key"] = priv
	rs["public_key"] = pub
	rs["short_id"] = generateShortID()[:10]
	if _, ok := rs["rx_path"]; !ok {
		rs["rx_path"] = "/"
	}
	if _, ok := rs["dest"]; !ok {
		rs["dest"] = "is1-ssl.mzstatic.com:443"
	}
	ib.RealitySettings = rs
	ib.Security = "reality"
	ib.Protocol = "reality"
	State.mu.Unlock()
	saveState()
	writeJSON(w, http.StatusOK, map[string]any{
		"ok": true, "public_key": pub, "private_key": priv,
		"short_id": rs["short_id"], "rx_path": rs["rx_path"],
	})
}

// handleGenerateShortID POST /api/inbounds/{id}/generate-short-id
func handleGenerateShortID(w http.ResponseWriter, r *http.Request, id string) {
	State.mu.Lock()
	ib := State.Inbounds[id]
	if ib == nil {
		State.mu.Unlock()
		writeJSON(w, http.StatusNotFound, map[string]any{"ok": false})
		return
	}
	if ib.RealitySettings == nil {
		ib.RealitySettings = map[string]any{}
	}
	sid := generateShortID()[:10]
	ib.RealitySettings["short_id"] = sid
	State.mu.Unlock()
	saveState()
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "short_id": sid})
}

// handleToolsRealityKeys POST /api/tools/generate-reality-keys (no inbound)
func handleToolsRealityKeys(w http.ResponseWriter, r *http.Request) {
	priv, pub, err := generateX25519Keypair()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "detail": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok": true, "private_key": priv, "public_key": pub, "short_id": generateShortID()[:10],
	})
}

// handleRealitySettings GET/POST /api/tools/reality-settings
func handleRealitySettings(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		body := jsonBody(r)
		State.mu.Lock()
		rs, _ := State.Settings["reality"].(map[string]any)
		if rs == nil {
			rs = map[string]any{}
		}
		for k, v := range body {
			rs[k] = v
		}
		if rs["private_key"] == nil || rs["private_key"] == "" {
			priv, pub, _ := generateX25519Keypair()
			rs["private_key"] = priv
			rs["public_key"] = pub
			rs["short_id"] = generateShortID()[:10]
		}
		State.Settings["reality"] = rs
		State.mu.Unlock()
		saveState()
	}
	State.mu.RLock()
	rs, _ := State.Settings["reality"].(map[string]any)
	State.mu.RUnlock()
	if rs == nil {
		rs = realityDefaults(nil)
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "reality": rs})
}

// ============================================================
// CONFIG GENERATION
// ============================================================

// generateUserConfig builds a vless:// / vmess:// / etc. share link for a user
// on a given inbound, replicating the Python logic.
func generateUserConfig(r *http.Request, u *User, inboundID string) string {
	State.mu.RLock()
	ib := State.Inbounds[inboundID]
	panelDomain, _ := State.Settings["domain"].(string)
	State.mu.RUnlock()
	if panelDomain == "" {
		panelDomain = hostFromRequest(r)
	}
	if ib == nil {
		return ""
	}

	proto := strings.ToLower(ib.Protocol)
	sec := strings.ToLower(ib.Security)
	cfgUUID := u.ConfigUUID
	remark := "White-" + u.Username

	// Worker: address=worker domain, path=/ws/{uuid}
	if proto == "worker" {
		wdom := ib.ExternalDomain
		if wdom == "" {
			wdom = ib.Domain
		}
		if wdom == "" {
			return ""
		}
		path := "/ws/" + cfgUUID
		q := url.Values{}
		q.Set("encryption", "none")
		q.Set("security", "tls")
		q.Set("type", "ws")
		q.Set("host", wdom)
		q.Set("path", path)
		q.Set("sni", wdom)
		q.Set("fp", "chrome")
		q.Set("alpn", "http/1.1")
		return fmt.Sprintf("vless://%s@%s:443?%s#%s",
			cfgUUID, wdom, q.Encode(), url.QueryEscape(remark))
	}

	// Telegram: t.me/proxy link
	if proto == "telegram" {
		tg := ib.TelegramSettings
		domain := getString(tg, "external_domain")
		if domain == "" {
			domain = ib.ExternalDomain
		}
		port := getInt(tg, "external_port")
		if port == 0 {
			port = ib.ExternalPort
		}
		if domain == "" || port == 0 {
			return ""
		}
		secret := u.TelegramSecret
		if secret == "" {
			return ""
		}
		return fmt.Sprintf("https://t.me/proxy?server=%s&port=%d&secret=%s", domain, port, secret)
	}

	// Reality: use inbound's reality_settings + external_domain/external_port
	if proto == "reality" || sec == "reality" {
		rs := ib.RealitySettings
		if rs == nil {
			rs = map[string]any{}
		}
		extDomain := ib.ExternalDomain
		if extDomain == "" {
			extDomain = ib.Domain
		}
		if extDomain == "" {
			return ""
		}
		extPort := ib.ExternalPort
		if extPort == 0 {
			extPort = ib.Port
		}
		if extPort == 0 {
			extPort = 443
		}
		priv, _ := rs["private_key"].(string)
		pub, _ := rs["public_key"].(string)
		if priv == "" {
			return ""
		}
		derived, err := deriveX25519Public(priv)
		if err != nil {
			return ""
		}
		pub = derived
		sid, _ := rs["short_id"].(string)
		sni := ib.SNI
		if sni == "" {
			sni, _ = rs["sni"].(string)
		}
		if sni == "" {
			sni = "is1-ssl.mzstatic.com"
		}
		spx, _ := rs["rx_path"].(string)
		if spx == "" {
			spx = "/"
		}
		network := ib.Network
		if network == "" {
			network = "tcp"
		}

		if network == "xhttp" || network == "" {
			// XHTTP transport
			xs := ib.XHTTPSettings
			if xs == nil {
				xs = map[string]any{}
			}
			xpath, _ := xs["path"].(string)
			if xpath == "" {
				xpath = "/"
			}
			xmode, _ := xs["mode"].(string)
			if xmode == "" || xmode == "auto" {
				xmode = "stream-up"
			}
			xpb := xs["xPaddingBytes"]
			if xpb == nil {
				xpb = "100-1000"
			}
			extraJSON := fmt.Sprintf(`{"xPaddingBytes":"%v"}`, xpb)
			if xmode == "packet-up" {
				sc := xs["scMaxEachPostBytes"]
				if sc == nil {
					sc = "1000000"
				}
				extraJSON = fmt.Sprintf(`{"xPaddingBytes":"%v","scMaxEachPostBytes":"%v"}`, xpb, sc)
			}
			extraEncoded := url.QueryEscape(extraJSON)
			q := url.Values{}
			q.Set("encryption", "none")
			q.Set("security", "reality")
			q.Set("sni", sni)
			q.Set("fp", ib.Fingerprint)
			q.Set("pbk", pub)
			q.Set("sid", sid)
			q.Set("spx", spx)
			q.Set("type", "xhttp")
			q.Set("path", xpath)
			q.Set("mode", xmode)
			q.Set("extra", extraEncoded)
			return fmt.Sprintf("vless://%s@%s:%d?%s#%s",
				cfgUUID, extDomain, extPort, q.Encode(), url.QueryEscape(remark))
		}
		// TCP transport (default for reality)
		q := url.Values{}
		q.Set("encryption", "none")
		q.Set("security", "reality")
		q.Set("sni", sni)
		q.Set("fp", ib.Fingerprint)
		q.Set("pbk", pub)
		q.Set("sid", sid)
		q.Set("spx", spx)
		q.Set("type", "tcp")
		q.Set("alpn", "h2,http/1.1")
		return fmt.Sprintf("vless://%s@%s:%d?%s#%s",
			cfgUUID, extDomain, extPort, q.Encode(), url.QueryEscape(remark))
	}

	// Standard TLS (WS / XHTTP / gRPC / TCP)
	addr := panelDomain
	if ib.ExternalDomain != "" {
		addr = ib.ExternalDomain
	}
	port := ib.ExternalPort
	if port == 0 {
		port = 443
	}
	network := ib.Network
	if network == "" {
		network = "ws"
	}

	if network == "xhttp" {
		xs := ib.XHTTPSettings
		if xs == nil {
			xs = map[string]any{}
		}
		xmode, _ := xs["mode"].(string)
		if xmode == "" || xmode == "auto" {
			xmode = "stream-up"
		}
		xpath := fmt.Sprintf("/xhttp-siz10/%s/%s", xmode, cfgUUID)
		q := url.Values{}
		q.Set("encryption", "none")
		q.Set("security", "tls")
		q.Set("type", "xhttp")
		q.Set("host", addr)
		q.Set("path", xpath)
		q.Set("sni", addr)
		q.Set("fp", "chrome")
		q.Set("mode", xmode)
		q.Set("alpn", "h2,http/1.1")
		return fmt.Sprintf("vless://%s@%s:%d?%s#%s",
			cfgUUID, addr, port, q.Encode(), url.QueryEscape(remark))
	}

	if network == "grpc" {
		gs := ib.GRPCSettings
		if gs == nil {
			gs = map[string]any{}
		}
		svc, _ := gs["serviceName"].(string)
		if svc == "" {
			svc = "white"
		}
		q := url.Values{}
		q.Set("encryption", "none")
		q.Set("security", "tls")
		q.Set("type", "grpc")
		q.Set("serviceName", svc)
		q.Set("sni", addr)
		q.Set("fp", "chrome")
		q.Set("alpn", "h2,http/1.1")
		return fmt.Sprintf("vless://%s@%s:%d?%s#%s",
			cfgUUID, addr, port, q.Encode(), url.QueryEscape(remark))
	}

	if network == "tcp" {
		q := url.Values{}
		q.Set("encryption", "none")
		q.Set("security", "tls")
		q.Set("type", "tcp")
		q.Set("sni", addr)
		q.Set("fp", "chrome")
		q.Set("alpn", "h2,http/1.1")
		return fmt.Sprintf("vless://%s@%s:%d?%s#%s",
			cfgUUID, addr, port, q.Encode(), url.QueryEscape(remark))
	}

	// Default: WS
	wsPath := "/ws/" + cfgUUID
	if ib.WSSettings != nil {
		if p, ok := ib.WSSettings["path"].(string); ok && p != "" {
			// Only use custom path if it doesn't carry {uuid} — /ws/{uuid} is
			// the canonical relay path that the FastAPI/Go relay listens on.
			if !strings.Contains(p, "{uuid}") && p != "/ws/{uuid}" {
				wsPath = p
			}
		}
	}
	q := url.Values{}
	q.Set("encryption", "none")
	q.Set("security", "tls")
	q.Set("type", "ws")
	q.Set("host", addr)
	q.Set("path", wsPath)
	q.Set("sni", addr)
	q.Set("fp", "chrome")
	q.Set("alpn", "http/1.1")
	return fmt.Sprintf("vless://%s@%s:%d?%s#%s",
		cfgUUID, addr, port, q.Encode(), url.QueryEscape(remark))
}

// generateUserConfigs returns all share links for a user (one per selected inbound).
func generateUserConfigs(r *http.Request, u *User) []string {
	out := []string{}
	seen := map[string]bool{}
	for _, iid := range u.InboundIDs {
		if isNodeControlInbound(iid) {
			continue
		}
		cfg := generateUserConfig(r, u, iid)
		if cfg != "" && !seen[cfg] {
			seen[cfg] = true
			out = append(out, cfg)
		}
	}
	return out
}

func isNodeControlInbound(iid string) bool {
	State.mu.RLock()
	defer State.mu.RUnlock()
	ib := State.Inbounds[iid]
	return ib != nil && (ib.System || ib.Protocol == "node")
}

// ============================================================
// USERS
// ============================================================

func handleUsersList(w http.ResponseWriter, r *http.Request) {
	host := discoveredHost(r)
	State.mu.RLock()
	list := make([]map[string]any, 0, len(State.Users))
	for _, u := range State.Users {
		status := u.Status
		// Auto-expire
		if u.ExpireAt != "" && status == "active" {
			if t, err := time.Parse(time.RFC3339, u.ExpireAt); err == nil && time.Now().After(t) {
				status = "expired"
			}
		}
		var trafficPct float64
		if u.TrafficLimitBytes > 0 {
			trafficPct = float64(u.TrafficUsedBytes) / float64(u.TrafficLimitBytes) * 100
		}
		item := map[string]any{
			"user_id":                 u.UserID,
			"username":                u.Username,
			"protocol":                u.Protocol,
			"transport_type":          u.TransportType,
			"path":                    u.Path,
			"config_uuid":             u.ConfigUUID,
			"subscription_uuid":       u.SubscriptionUUID,
			"traffic_limit_bytes":     u.TrafficLimitBytes,
			"traffic_used_bytes":      u.TrafficUsedBytes,
			"traffic_limit_fmt":       formatBytes(u.TrafficLimitBytes),
			"traffic_used_fmt":        formatBytes(u.TrafficUsedBytes),
			"traffic_percent":         trafficPct,
			"expire_at":               u.ExpireAt,
			"concurrent_connections":  u.ConcurrentConnections,
			"speed_limit_mbps":        u.SpeedLimitMbps,
			"status":                  status,
			"server":                  u.Server,
			"inbound_id":              u.InboundID,
			"inbound_ids":             u.InboundIDs,
			"proxy_ips":               u.ProxyIPs,
			"proxy_ip_enabled":        u.ProxyIPEnabled,
			"custom_ip_type":          u.CustomIPType,
			"created_at":              u.CreatedAt,
			"config_url":              fmt.Sprintf("https://%s/api/users/%s/config", host, u.UserID),
			"qr_url":                  fmt.Sprintf("https://%s/api/users/%s/qr", host, u.UserID),
			"subscription_url":        fmt.Sprintf("https://%s/link/%s", host, u.ConfigUUID),
			"node_configs":            u.NodeConfigs,
			"node_sync_state":         u.NodeSyncState,
			"node_traffic":            u.NodeTraffic,
			"node_traffic_used_bytes": u.NodeTrafficUsedBytes,
		}
		list = append(list, item)
	}
	State.mu.RUnlock()
	sort.Slice(list, func(i, j int) bool {
		return fmt.Sprintf("%v", list[i]["created_at"]) > fmt.Sprintf("%v", list[j]["created_at"])
	})
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "users": list})
}

func handleUserCreate(w http.ResponseWriter, r *http.Request) {
	body := jsonBody(r)
	username := strings.TrimSpace(getString(body, "username"))
	if username == "" {
		username = "user-" + generateShortID()[:6]
	}
	trafficGB := getFloat(body, "traffic_limit_gb")
	if _, ok := body["traffic_limit_gb"]; !ok {
		trafficGB = getFloat(body, "traffic_limit")
	}
	expireDays := getInt(body, "expire_days")
	protocol := strings.ToLower(getString(body, "protocol"))
	if protocol == "" {
		protocol = "vless"
	}
	transportType := strings.ToLower(getString(body, "transport_type"))
	inboundIDs := getStringSlice(body, "inbound_ids")
	primaryIID := getString(body, "inbound_id")
	if primaryIID != "" && !contains(inboundIDs, primaryIID) {
		inboundIDs = append([]string{primaryIID}, inboundIDs...)
	}
	if len(inboundIDs) > 0 {
		primaryIID = inboundIDs[0]
	}

	// Derive transport type from primary inbound
	if transportType == "" && primaryIID != "" {
		State.mu.RLock()
		ib := State.Inbounds[primaryIID]
		if ib != nil {
			if ib.Protocol == "reality" || ib.Security == "reality" {
				transportType = "reality"
			} else {
				transportType = ib.Network
			}
		}
		State.mu.RUnlock()
	}
	if transportType == "" {
		transportType = "ws"
	}

	cfgUUID := generateUUID()
	userID := generateShortID()
	subUUID := generateToken(16)

	var expireAt string
	if expireDays > 0 {
		expireAt = time.Now().AddDate(0, 0, expireDays).Format(time.RFC3339)
	}

	// Path selection based on primary inbound
	path := "/ws/" + cfgUUID
	if primaryIID != "" {
		State.mu.RLock()
		ib := State.Inbounds[primaryIID]
		if ib != nil {
			if ib.Protocol == "reality" && ib.Network == "xhttp" {
				if p, ok := ib.XHTTPSettings["path"].(string); ok && p != "" {
					path = p
				}
			} else if ib.Network == "xhttp" {
				path = fmt.Sprintf("/xhttp-siz10/stream-up/%s", cfgUUID)
			}
		}
		State.mu.RUnlock()
	}

	u := &User{
		UserID:                userID,
		Username:              username,
		PasswordHash:          "",
		Protocol:              protocol,
		TransportType:         transportType,
		ConfigUUID:            cfgUUID,
		SubscriptionUUID:      subUUID,
		TrafficLimitBytes:     int64(trafficGB * 1073741824),
		TrafficUsedBytes:      0,
		ExpireAt:              expireAt,
		ConcurrentConnections: getInt(body, "concurrent_connections"),
		SpeedLimitMbps:        getFloat(body, "speed_limit_mbps"),
		Status:                "active",
		Server:                getString(body, "server"),
		Path:                  path,
		InboundID:             primaryIID,
		InboundIDs:            inboundIDs,
		ProxyIPs:              getStringSlice(body, "proxy_ips"),
		ProxyIPEnabled:        getBool(body, "proxy_ip_enabled"),
		CustomIPType:          getString(body, "custom_ip_type"),
		CustomIPInbounds:      map[string][]string{
			"cf":      getStringSlice(getMap(body, "custom_ip_inbounds"), "cf"),
			"railway": getStringSlice(getMap(body, "custom_ip_inbounds"), "railway"),
		},
		SNISpoofV2Box:         getBool(body, "sni_spoof_v2box"),
		NodeConfigs:           map[string]string{},
		NodeSyncState:         map[string]any{},
		NodeTraffic:           map[string]int64{},
		CreatedAt:             time.Now().Format(time.RFC3339),
	}
	// Auto-generate Telegram secret if needed
	for _, iid := range inboundIDs {
		State.mu.RLock()
		ib := State.Inbounds[iid]
		State.mu.RUnlock()
		if ib != nil && ib.Protocol == "telegram" {
			u.TelegramSecret = sha256Hex(fmt.Sprintf("white-tg-proxy:%s", cfgUUID))[:32]
			break
		}
	}

	State.mu.Lock()
	// Uniqueness
	for _, existing := range State.Users {
		if existing.Username == username {
			State.mu.Unlock()
			writeJSON(w, http.StatusConflict, map[string]any{"ok": false, "detail": "username exists"})
			return
		}
	}
	State.Users[userID] = u
	State.mu.Unlock()
	saveState()

	host := discoveredHost(r)
	u.PasswordHash = ""
	resp := map[string]any{
		"ok": true, "user": u, "user_id": userID,
		"config_url":       fmt.Sprintf("https://%s/api/users/%s/config", host, userID),
		"qr_url":           fmt.Sprintf("https://%s/api/users/%s/qr", host, userID),
		"subscription_url": fmt.Sprintf("https://%s/link/%s", host, cfgUUID),
	}
	if len(inboundIDs) > 0 {
		resp["config"] = generateUserConfig(r, u, inboundIDs[0])
	}
	writeJSON(w, http.StatusCreated, resp)
}

func handleUserGet(w http.ResponseWriter, r *http.Request, id string) {
	State.mu.RLock()
	target := State.Users[id]
	if target == nil {
		for _, u := range State.Users {
			if u.ConfigUUID == id {
				target = u
				break
			}
		}
	}
	if target == nil {
		State.mu.RUnlock()
		writeJSON(w, http.StatusNotFound, map[string]any{"ok": false, "detail": "user not found"})
		return
	}
	cp := *target
	State.mu.RUnlock()
	cp.PasswordHash = ""
	host := discoveredHost(r)
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":               true,
		"user_id":          cp.UserID,
		"username":         cp.Username,
		"config_uuid":      cp.ConfigUUID,
		"status":           cp.Status,
		"traffic_used_bytes":  cp.TrafficUsedBytes,
		"traffic_limit_bytes": cp.TrafficLimitBytes,
		"expire_at":        cp.ExpireAt,
		"config":           generateUserConfig(r, &cp, cp.InboundID),
		"config_url":       fmt.Sprintf("https://%s/api/users/%s/config", host, cp.UserID),
		"qr_url":           fmt.Sprintf("https://%s/api/users/%s/qr", host, cp.UserID),
		"subscription_url": fmt.Sprintf("https://%s/link/%s", host, cp.ConfigUUID),
	})
}

func handleUserUpdate(w http.ResponseWriter, r *http.Request, id string) {
	State.mu.Lock()
	u := State.Users[id]
	if u == nil {
		State.mu.Unlock()
		writeJSON(w, http.StatusNotFound, map[string]any{"ok": false})
		return
	}
	body := jsonBody(r)
	if v := getString(body, "username"); v != "" {
		u.Username = v
	}
	if _, ok := body["traffic_limit_gb"]; ok {
		gb := getFloat(body, "traffic_limit_gb")
		u.TrafficLimitBytes = int64(gb * 1073741824)
	}
	if _, ok := body["expire_days"]; ok {
		days := getInt(body, "expire_days")
		if days > 0 {
			u.ExpireAt = time.Now().AddDate(0, 0, days).Format(time.RFC3339)
		} else {
			u.ExpireAt = ""
		}
	}
	if _, ok := body["status"]; ok {
		u.Status = getString(body, "status")
	}
	if _, ok := body["concurrent_connections"]; ok {
		u.ConcurrentConnections = getInt(body, "concurrent_connections")
	}
	if _, ok := body["speed_limit_mbps"]; ok {
		u.SpeedLimitMbps = getFloat(body, "speed_limit_mbps")
	}
	if _, ok := body["inbound_ids"]; ok {
		ids := getStringSlice(body, "inbound_ids")
		// Filter to real inbounds
		valid := []string{}
		for _, iid := range ids {
			if State.Inbounds[iid] != nil {
				valid = append(valid, iid)
			}
		}
		u.InboundIDs = valid
		if len(valid) > 0 {
			u.InboundID = valid[0]
		}
	}
	if getBool(body, "reset_traffic") {
		u.TrafficUsedBytes = 0
	}
	if _, ok := body["proxy_ips"]; ok {
		u.ProxyIPs = getStringSlice(body, "proxy_ips")
	}
	State.mu.Unlock()
	saveState()
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "user_id": id})
}

func handleUserDelete(w http.ResponseWriter, r *http.Request, id string) {
	State.mu.Lock()
	target := State.Users[id]
	if target == nil {
		for uid, u := range State.Users {
			if u.ConfigUUID == id {
				target = u
				id = uid
				break
			}
		}
	}
	if target == nil {
		State.mu.Unlock()
		writeJSON(w, http.StatusNotFound, map[string]any{"ok": false})
		return
	}
	delete(State.Users, id)
	State.mu.Unlock()
	saveState()
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "deleted": id, "config_uuid": target.ConfigUUID})
}

func handleUserToggle(w http.ResponseWriter, r *http.Request, id string) {
	State.mu.Lock()
	u := State.Users[id]
	if u == nil {
		State.mu.Unlock()
		writeJSON(w, http.StatusNotFound, map[string]any{"ok": false})
		return
	}
	if u.Status == "active" {
		u.Status = "disabled"
	} else {
		u.Status = "active"
	}
	newStatus := u.Status
	State.mu.Unlock()
	saveState()
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "user_id": id, "status": newStatus})
}

func handleUserReset(w http.ResponseWriter, r *http.Request, id string) {
	State.mu.Lock()
	u := State.Users[id]
	if u == nil {
		State.mu.Unlock()
		writeJSON(w, http.StatusNotFound, map[string]any{"ok": false})
		return
	}
	u.TrafficUsedBytes = 0
	State.mu.Unlock()
	saveState()
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "user_id": id, "traffic_used_bytes": 0})
}

func handleUserConfigGet(w http.ResponseWriter, r *http.Request, id string) {
	State.mu.RLock()
	u := State.Users[id]
	if u == nil {
		State.mu.RUnlock()
		writeJSON(w, http.StatusNotFound, map[string]any{"ok": false})
		return
	}
	cp := *u
	State.mu.RUnlock()
	host := discoveredHost(r)
	primary := cp.InboundID
	if primary == "" && len(cp.InboundIDs) > 0 {
		primary = cp.InboundIDs[0]
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok": true, "user_id": id, "username": cp.Username,
		"protocol": cp.Protocol,
		"config":   generateUserConfig(r, &cp, primary),
		"config_url":       fmt.Sprintf("https://%s/api/users/%s/config", host, id),
		"qr_url":           fmt.Sprintf("https://%s/api/users/%s/qr", host, id),
		"subscription_url": fmt.Sprintf("https://%s/link/%s", host, cp.ConfigUUID),
	})
}

func handleUserSubscriptionGet(w http.ResponseWriter, r *http.Request, id string) {
	State.mu.RLock()
	u := State.Users[id]
	if u == nil {
		State.mu.RUnlock()
		writeJSON(w, http.StatusNotFound, map[string]any{"ok": false})
		return
	}
	cp := *u
	State.mu.RUnlock()
	host := discoveredHost(r)
	configs := generateUserConfigs(r, &cp)
	writeJSON(w, http.StatusOK, map[string]any{
		"ok": true, "user_id": id, "username": cp.Username,
		"subscription_uuid": cp.SubscriptionUUID,
		"subscription_url":  fmt.Sprintf("https://%s/link/%s", host, cp.ConfigUUID),
		"configs":           configs,
	})
}

// handleUserQR redirects to a QR service. Avoids adding a Go dependency.
// The frontend just uses it as <img src>, so a redirect works transparently.
func handleUserQR(w http.ResponseWriter, r *http.Request, id string) {
	State.mu.RLock()
	u := State.Users[id]
	if u == nil {
		State.mu.RUnlock()
		writeJSON(w, http.StatusNotFound, map[string]any{"ok": false})
		return
	}
	cfgUUID := u.ConfigUUID
	State.mu.RUnlock()
	host := discoveredHost(r)
	subURL := fmt.Sprintf("https://%s/link/%s", host, cfgUUID)
	qrURL := "https://api.qrserver.com/v1/create-qr-code/?size=280x280&margin=1&data=" + url.QueryEscape(subURL)
	http.Redirect(w, r, qrURL, http.StatusFound)
}

// ============================================================
// SUBSCRIPTION ENDPOINTS
// ============================================================

// handleAPISub returns the JSON payload consumed by static/sub.html
func handleAPISub(w http.ResponseWriter, r *http.Request, uuid string) {
	State.mu.RLock()
	var u *User
	for _, candidate := range State.Users {
		if candidate.ConfigUUID == uuid {
			u = candidate
			break
		}
	}
	if u == nil {
		State.mu.RUnlock()
		writeJSON(w, http.StatusNotFound, map[string]any{"ok": false, "detail": "کاربر پیدا نشد"})
		return
	}
	cp := *u
	State.mu.RUnlock()

	configs := generateUserConfigs(r, &cp)

	var expireDays *int
	var expireAtTs *int64
	if cp.ExpireAt != "" {
		if t, err := time.Parse(time.RFC3339, cp.ExpireAt); err == nil {
			d := int(time.Until(t).Hours() / 24)
			if d < 0 {
				d = 0
			}
			expireDays = &d
			ts := t.Unix()
			expireAtTs = &ts
		}
	}

	isActive := cp.Status == "active"
	if isActive && cp.ExpireAt != "" {
		if t, err := time.Parse(time.RFC3339, cp.ExpireAt); err == nil && time.Now().After(t) {
			isActive = false
		}
	}
	if isActive && cp.TrafficLimitBytes > 0 && cp.TrafficUsedBytes >= cp.TrafficLimitBytes {
		isActive = false
	}

	status := cp.Status
	if status == "" {
		status = "active"
	}
	if !isActive && status == "active" {
		status = "expired"
	}

	maxIP := cp.ConcurrentConnections
	if maxIP == 0 {
		State.mu.RLock()
		if v, ok := State.Settings["max_ip_per_user"].(int); ok {
			maxIP = v
		}
		State.mu.RUnlock()
		if maxIP == 0 {
			maxIP = 3
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":                  true,
		"config_uuid":         cp.ConfigUUID,
		"username":            cp.Username,
		"is_active":           isActive,
		"status":              status,
		"traffic_used_bytes":  cp.TrafficUsedBytes,
		"traffic_limit_bytes": cp.TrafficLimitBytes,
		"download_bytes":      cp.TrafficUsedBytes,
		"upload_bytes":        0,
		"expire_days":         expireDays,
		"expire_at_ts":        expireAtTs,
		"max_ip_per_user":     maxIP,
		"used_ips":            0,
		"configs":             configs,
		"subscription_url":    "/link/" + cp.ConfigUUID,
		"protocol":            cp.Protocol,
		"transport_type":      cp.TransportType,
		"path":                cp.Path,
	})
}

// handleSubPlainText returns base64 configs for subscription clients.
func handleSubPlainText(w http.ResponseWriter, r *http.Request, uuid string) {
	State.mu.RLock()
	var u *User
	for _, candidate := range State.Users {
		if candidate.ConfigUUID == uuid {
			u = candidate
			break
		}
	}
	if u == nil {
		State.mu.RUnlock()
		http.Error(w, "Subscription not found", http.StatusNotFound)
		return
	}
	cp := *u
	State.mu.RUnlock()

	if cp.Status != "active" {
		http.Error(w, "Subscription disabled", http.StatusForbidden)
		return
	}
	if cp.ExpireAt != "" {
		if t, err := time.Parse(time.RFC3339, cp.ExpireAt); err == nil && time.Now().After(t) {
			http.Error(w, "Subscription expired", http.StatusForbidden)
			return
		}
	}

	configs := generateUserConfigs(r, &cp)
	joined := strings.Join(configs, "\n")
	encoded := base64.StdEncoding.EncodeToString([]byte(joined))

	expireUnix := int64(0)
	if cp.ExpireAt != "" {
		if t, err := time.Parse(time.RFC3339, cp.ExpireAt); err == nil {
			expireUnix = t.Unix()
		}
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Subscription-Userinfo", fmt.Sprintf(
		"upload=0; download=%d; total=%d; expire=%d",
		cp.TrafficUsedBytes, cp.TrafficLimitBytes, expireUnix))
	w.Header().Set("Profile-Title", url.QueryEscape(cp.Username))
	w.Header().Set("Profile-Update-Interval", "12")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(encoded))
}

// handleSubPage: serve sub.html for browsers, plain text for clients.
func handleSubPage(w http.ResponseWriter, r *http.Request, uuid string) {
	ua := strings.ToLower(r.UserAgent())
	accept := strings.ToLower(r.Header.Get("Accept"))
	isBrowser := strings.Contains(ua, "mozilla") || strings.Contains(accept, "text/html")
	if !isBrowser {
		handleSubPlainText(w, r, uuid)
		return
	}
	// Validate that the UUID exists before serving the HTML page
	State.mu.RLock()
	found := false
	for _, u := range State.Users {
		if u.ConfigUUID == uuid {
			found = true
			break
		}
	}
	State.mu.RUnlock()
	if !found {
		http.Error(w, "Subscription not found", http.StatusNotFound)
		return
	}
	http.ServeFile(w, r, filepath.Join("./static", "sub.html"))
}

// handleSubQR redirects to a QR service for a subscription.
func handleSubQR(w http.ResponseWriter, r *http.Request, uuid string) {
	host := discoveredHost(r)
	data := fmt.Sprintf("https://%s/link/%s", host, uuid)
	cfg := r.URL.Query().Get("cfg")
	if cfg != "" {
		// The frontend can request a specific config by index; we redirect
		// to the full sub URL QR for simplicity (better than nothing).
		_ = cfg
	}
	qrURL := "https://api.qrserver.com/v1/create-qr-code/?size=280x280&margin=1&data=" + url.QueryEscape(data)
	http.Redirect(w, r, qrURL, http.StatusFound)
}

// ============================================================
// SCANNER (essential endpoints)
// ============================================================

func handleScannerCFSubnets(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"ok": true,
		"subnets": []string{
			"173.245.48.0/20", "103.21.244.0/22", "103.22.200.0/22",
			"103.31.4.0/22", "141.101.64.0/18", "108.162.192.0/18",
			"190.93.240.0/20", "188.114.96.0/20", "197.234.240.0/22",
			"198.41.128.0/17", "162.158.0.0/15", "104.16.0.0/13",
		},
	})
}

func handleScannerResolve(w http.ResponseWriter, r *http.Request) {
	host := r.URL.Query().Get("host")
	if host == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false})
		return
	}
	ips, _ := net.LookupIP(host)
	out := []string{}
	for _, ip := range ips {
		out = append(out, ip.String())
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "host": host, "ips": out})
}

func handleScannerSNICheck(w http.ResponseWriter, r *http.Request) {
	host := r.URL.Query().Get("host")
	port := r.URL.Query().Get("port")
	if port == "" {
		port = "443"
	}
	if host == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false})
		return
	}
	start := time.Now()
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), 3*time.Second)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"ok": false, "sni": host, "port": port, "error": err.Error(),
		})
		return
	}
	conn.Close()
	writeJSON(w, http.StatusOK, map[string]any{
		"ok": true, "sni": host, "port": port, "latency_ms": latency,
	})
}

func handleScannerPingBatch(w http.ResponseWriter, r *http.Request) {
	body := jsonBody(r)
	targets := getStringSlice(body, "targets")
	timeout := getFloat(body, "timeout")
	if timeout <= 0 || timeout > 6 {
		timeout = 2
	}
	type result struct {
		Target    string `json:"target"`
		IP        string `json:"ip"`
		Port      int    `json:"port"`
		LatencyMs *int   `json:"latency_ms"`
		Ok        bool   `json:"ok"`
	}
	results := make([]result, 0, len(targets))
	sem := make(chan struct{}, 20)
	var mu sync.Mutex
	var wg sync.WaitGroup
	for _, t := range targets {
		t := t
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			ip, port := t, "443"
			if i := strings.LastIndex(t, ":"); i > 0 {
				ip = t[:i]
				port = t[i+1:]
			}
			var portN int
			fmt.Sscanf(port, "%d", &portN)
			start := time.Now()
			conn, err := net.DialTimeout("tcp", net.JoinHostPort(ip, port), time.Duration(timeout*float64(time.Second)))
			lat := int(time.Since(start).Milliseconds())
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				results = append(results, result{Target: t, IP: ip, Port: portN, Ok: false})
			} else {
				conn.Close()
				results = append(results, result{Target: t, IP: ip, Port: portN, LatencyMs: &lat, Ok: true})
			}
		}()
	}
	wg.Wait()
	sort.Slice(results, func(i, j int) bool {
		if results[i].LatencyMs == nil {
			return false
		}
		if results[j].LatencyMs == nil {
			return true
		}
		return *results[i].LatencyMs < *results[j].LatencyMs
	})
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "count": len(results), "results": results})
}

// handleScannerSave is a no-op that returns a valid response shape.
func handleScannerSave(w http.ResponseWriter, r *http.Request) {
	body := jsonBody(r)
	ctype := getString(body, "type")
	ips := getStringSlice(body, "ips")
	writeJSON(w, http.StatusOK, map[string]any{
		"ok": true, "type": ctype, "ips": ips, "seq": 1,
	})
}

func handleScannerIPS(w http.ResponseWriter, r *http.Request, ctype string) {
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "type": ctype, "ips": []string{}, "seq": 0})
}

func handleScannerSNIList(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "sni_list": []string{}})
}

func handleScannerSNIScanList(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "snis": []string{}, "count": 0})
}

func handleScannerSNIResults(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "results": []any{}})
}

func handleScannerSNISaveResults(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "saved": 0, "results": []any{}})
}

func handleScannerSNIClearResults(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "results": []any{}})
}

func handleScannerSNIFastest(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"ok": false, "sni": "", "latency_ms": 0})
}

// ============================================================
// BOT / WORKER STUBS — correct shape, no complex logic
// ============================================================

func handleBotStatus(w http.ResponseWriter, r *http.Request) {
	State.mu.RLock()
	plans := []*BotPlan{}
	for _, p := range State.BotPlans {
		plans = append(plans, p)
	}
	State.mu.RUnlock()
	writeJSON(w, http.StatusOK, map[string]any{
		"ok": true, "token_configured": false, "token_masked": "",
		"channel": map[string]any{"enabled": false, "channel": ""},
		"sell":    map[string]any{"enabled": false, "required_channels": []any{}},
		"plans":   plans,
		"orders":  map[string]int{"pending": 0, "payment_submitted": 0, "approved": 0, "rejected": 0, "cancelled": 0},
		"scheduler_running": false, "poller_running": false,
	})
}

func handleBotConfig(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func handleBotTest(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "detail": "Bot logic not implemented in Go build"})
}

func handleBotSellPlans(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		State.mu.RLock()
		plans := []*BotPlan{}
		for _, p := range State.BotPlans {
			plans = append(plans, p)
		}
		State.mu.RUnlock()
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "plans": plans})
	case http.MethodPost:
		body := jsonBody(r)
		p := &BotPlan{
			ID:             generateShortID(),
			Name:           getString(body, "name"),
			Price:          getString(body, "price"),
			TrafficGB:      getFloat(body, "traffic_gb"),
			ExpireDays:     getInt(body, "expire_days"),
			InboundID:      getString(body, "inbound_id"),
			UsernamePrefix: getString(body, "username_prefix"),
			Enabled:        getBool(body, "enabled"),
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

func handleWorkerRoute(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"ok": true, "connected": false, "status": "ready",
		"proxies": []any{}, "locations": []any{},
		"worker":  map[string]any{"connected": false, "locations": []any{}, "proxies": []any{}},
		"tunnel":  map[string]any{"enabled": false, "logs": []any{}, "ping_ms": nil},
		"reverse": map[string]any{"enabled": false, "logs": []any{}, "ping_ms": nil},
	})
}

// ============================================================
// NODES
// ============================================================

func handleNodesList(w http.ResponseWriter, r *http.Request) {
	State.mu.RLock()
	list := make([]map[string]any, 0, len(State.Nodes))
	for _, n := range State.Nodes {
		cp := *n
		if cp.APIKey != "" {
			mask := cp.APIKey[:6] + "…" + cp.APIKey[len(cp.APIKey)-4:]
			cp.APIKeyMask = mask
			cp.APIKey = ""
		}
		list = append(list, map[string]any{
			"node_id": cp.NodeID, "name": cp.Name, "domain": cp.Domain,
			"api_key_masked": cp.APIKeyMask, "status": cp.Status,
			"latency_ms": cp.Latency, "last_error": cp.LastError,
			"created_at": cp.CreatedAt,
			"enabled_inbound_count": 0, "synced_user_count": 0,
		})
	}
	State.mu.RUnlock()
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "nodes": list})
}

func handleNodeAdd(w http.ResponseWriter, r *http.Request) {
	body := jsonBody(r)
	domain := getString(body, "domain")
	apiKey := getString(body, "api_key")
	if domain == "" || apiKey == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "detail": "domain and api_key required"})
		return
	}
	n := &Node{
		NodeID:    "node_" + generateShortID(),
		Name:      getString(body, "name"),
		Domain:    domain,
		APIKey:    apiKey,
		Status:    "checking",
		CreatedAt: time.Now().Format(time.RFC3339),
	}
	State.mu.Lock()
	State.Nodes[n.NodeID] = n
	State.mu.Unlock()
	saveState()
	n.APIKey = ""
	writeJSON(w, http.StatusCreated, map[string]any{"ok": true, "node": n})
}

func handleNodeUpdate(w http.ResponseWriter, r *http.Request, id string) {
	State.mu.Lock()
	n := State.Nodes[id]
	if n == nil {
		State.mu.Unlock()
		writeJSON(w, http.StatusNotFound, map[string]any{"ok": false})
		return
	}
	body := jsonBody(r)
	if v := getString(body, "name"); v != "" {
		n.Name = v
	}
	if v := getString(body, "domain"); v != "" {
		n.Domain = v
	}
	if v := getString(body, "api_key"); v != "" {
		n.APIKey = v
	}
	State.mu.Unlock()
	saveState()
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "node_id": id})
}

func handleNodeDelete(w http.ResponseWriter, r *http.Request, id string) {
	State.mu.Lock()
	delete(State.Nodes, id)
	State.mu.Unlock()
	saveState()
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "deleted": id})
}

func handleNodeRefresh(w http.ResponseWriter, r *http.Request, id string) {
	State.mu.Lock()
	n := State.Nodes[id]
	if n == nil {
		State.mu.Unlock()
		writeJSON(w, http.StatusNotFound, map[string]any{"ok": false})
		return
	}
	n.Status = "offline"
	State.mu.Unlock()
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "node": n})
}

func handleNodeSyncAll(w http.ResponseWriter, r *http.Request) {
	State.mu.RLock()
	count := len(State.Nodes)
	State.mu.RUnlock()
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "synced": count})
}

// ============================================================
// SETTINGS / BACKUP / UPDATE
// ============================================================

func handleSettingsGet(w http.ResponseWriter, r *http.Request) {
	State.mu.RLock()
	s := map[string]any{}
	for k, v := range State.Settings {
		if k == "panel_api_key" {
			if str, ok := v.(string); ok && len(str) > 8 {
				s[k] = str[:8] + "********"
				continue
			}
		}
		s[k] = v
	}
	State.mu.RUnlock()
	s["ok"] = true
	writeJSON(w, http.StatusOK, s)
}

func handleSettingsUpdate(w http.ResponseWriter, r *http.Request) {
	body := jsonBody(r)
	allowed := map[string]bool{
		"websocket_mode": true, "xhttp_mode": true,
		"default_connection_mode": true, "max_ip_per_user": true,
		"bandwidth_limit_mbps": true, "live_monitoring": true,
		"auto_ip_rotation": true, "server_ip": true,
		"country": true, "country_code": true, "country_flag": true,
		"domain": true,
	}
	State.mu.Lock()
	for k, v := range body {
		if allowed[k] {
			State.Settings[k] = v
		}
	}
	State.mu.Unlock()
	saveState()
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func handleSettingsBackup(w http.ResponseWriter, r *http.Request) {
	State.mu.RLock()
	payload := map[string]any{
		"backup_format":  "WhitePanel",
		"backup_version": 2,
		"created_at":     time.Now().Format(time.RFC3339),
		"state": map[string]any{
			"users":    State.Users,
			"inbounds": State.Inbounds,
			"nodes":    State.Nodes,
			"settings": State.Settings,
			"password_hash": State.Password,
			"bot_plans": State.BotPlans,
		},
	}
	State.mu.RUnlock()
	body, _ := json.MarshalIndent(payload, "", "  ")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", `attachment; filename="white-panel-backup.json"`)
	w.Write(body)
}

func handleSettingsRestore(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(25 * 1024 * 1024); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "detail": "invalid form"})
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "detail": "no file"})
		return
	}
	defer file.Close()
	var payload map[string]any
	if err := json.NewDecoder(file).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "detail": "invalid JSON"})
		return
	}
	state, _ := payload["state"].(map[string]any)
	if state == nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "detail": "no state key"})
		return
	}
	State.mu.Lock()
	if v, ok := state["users"].(map[string]any); ok {
		State.Users = map[string]*User{}
		for k, raw := range v {
			b, _ := json.Marshal(raw)
			var u User
			json.Unmarshal(b, &u)
			State.Users[k] = &u
		}
	}
	if v, ok := state["inbounds"].(map[string]any); ok {
		State.Inbounds = map[string]*Inbound{}
		for k, raw := range v {
			b, _ := json.Marshal(raw)
			var ib Inbound
			json.Unmarshal(b, &ib)
			State.Inbounds[k] = &ib
		}
	}
	if v, ok := state["settings"].(map[string]any); ok {
		State.Settings = v
	}
	if v, ok := state["password_hash"].(string); ok && v != "" {
		State.Password = v
	}
	State.mu.Unlock()
	saveState()
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "detail": "بکاپ با موفقیت بازیابی شد"})
}

func handleNotifications(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "items": []any{}, "source": ""})
}

func handleUpdateCheck(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"ok": true, "version": "8.0", "update_available": false,
		"latest": nil, "repo_url": "https://github.com/Itskillmaster/White-Panel-Railway",
	})
}

func handleActivityLog(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "logs": []any{}})
}

func handleMusix(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "tracks": []any{}})
}

func handleToolsMyIP(w http.ResponseWriter, r *http.Request) {
	State.mu.RLock()
	ip, _ := State.Settings["server_ip"].(string)
	State.mu.RUnlock()
	writeJSON(w, http.StatusOK, map[string]any{
		"ok": true, "ips": map[string]string{"ipify": ip},
		"railway_url": os.Getenv("RAILWAY_STATIC_URL"),
	})
}

// ============================================================
// HELPERS
// ============================================================

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

func sha256Hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

func formatBytes(b int64) string {
	if b <= 0 {
		return "0 B"
	}
	units := []string{"B", "KB", "MB", "GB", "TB"}
	f := float64(b)
	i := 0
	for f >= 1024 && i < len(units)-1 {
		f /= 1024
		i++
	}
	return fmt.Sprintf("%.2f %s", f, units[i])
}

// ============================================================
// ROUTER
// ============================================================

func main() {
	initStorage()

	port := os.Getenv("PANEL_PORT")
	if port == "" {
		port = "8080"
	}

	staticDir := "./static"
	mux := http.NewServeMux()

	// Static assets
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))

	// Root / login page
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/", "/white", "/dashboard":
			http.ServeFile(w, r, filepath.Join(staticDir, "index.html"))
			return
		case "/login":
			http.ServeFile(w, r, filepath.Join(staticDir, "login.html"))
			return
		}
		http.FileServer(http.Dir(staticDir)).ServeHTTP(w, r)
	})

	// Auth (public)
	mux.HandleFunc("/api/login", handleLogin)
	mux.HandleFunc("/api/logout", handleLogout)
	mux.HandleFunc("/api/me", handleMe)

	// Server info (public for /api/server-info used by the frontend to poll)
	mux.HandleFunc("/api/server-info", handleServerInfo)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "service": "WhitePanel-Go", "port": port})
	})

	// Subscription endpoints (public)
	mux.HandleFunc("/api/sub/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/sub/")
		// /api/sub/{uuid}/qr
		if strings.HasSuffix(path, "/qr") {
			uuid := strings.TrimSuffix(path, "/qr")
			handleSubQR(w, r, uuid)
			return
		}
		handleAPISub(w, r, path)
	})
	mux.HandleFunc("/sub/", func(w http.ResponseWriter, r *http.Request) {
		uuid := strings.TrimPrefix(r.URL.Path, "/sub/")
		handleSubPage(w, r, uuid)
	})
	mux.HandleFunc("/link/", func(w http.ResponseWriter, r *http.Request) {
		uuid := strings.TrimPrefix(r.URL.Path, "/link/")
		handleSubPlainText(w, r, uuid)
	})

	// Stats (auth)
	mux.HandleFunc("/stats", requireAuth(handleStats))
	mux.HandleFunc("/api/server/stats", requireAuth(handleStats))

	// Auth-protected: users CRUD
	mux.HandleFunc("/api/users", requireAuth(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleUsersList(w, r)
		case http.MethodPost:
			handleUserCreate(w, r)
		default:
			writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"ok": false})
		}
	}))
	mux.HandleFunc("/api/users/", requireAuth(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/users/")
		parts := strings.SplitN(path, "/", 2)
		id := parts[0]
		sub := ""
		if len(parts) > 1 {
			sub = parts[1]
		}
		switch sub {
		case "":
			switch r.Method {
			case http.MethodGet:
				handleUserGet(w, r, id)
			case http.MethodPatch:
				handleUserUpdate(w, r, id)
			case http.MethodDelete:
				handleUserDelete(w, r, id)
			default:
				writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"ok": false})
			}
		case "toggle":
			handleUserToggle(w, r, id)
		case "reset":
			handleUserReset(w, r, id)
		case "config":
			handleUserConfigGet(w, r, id)
		case "subscription":
			handleUserSubscriptionGet(w, r, id)
		case "qr":
			handleUserQR(w, r, id)
		default:
			writeJSON(w, http.StatusNotFound, map[string]any{"ok": false})
		}
	}))

	// Inbounds CRUD
	mux.HandleFunc("/api/inbounds", requireAuth(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleInboundsList(w, r)
		case http.MethodPost:
			handleInboundCreate(w, r)
		default:
			writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"ok": false})
		}
	}))
	mux.HandleFunc("/api/inbounds/", requireAuth(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/inbounds/")
		parts := strings.SplitN(path, "/", 2)
		id := parts[0]
		sub := ""
		if len(parts) > 1 {
			sub = parts[1]
		}
		switch sub {
		case "generate-reality-keys":
			handleRealityKeys(w, r, id)
		case "generate-short-id":
			handleGenerateShortID(w, r, id)
		case "sync-nodes":
			writeJSON(w, http.StatusOK, map[string]any{"ok": true, "users": 0, "results": []any{}, "node_ids": []string{}})
		case "":
			switch r.Method {
			case http.MethodPatch:
				handleInboundUpdate(w, r, id)
			case http.MethodDelete:
				handleInboundDelete(w, r, id)
			case http.MethodGet:
				State.mu.RLock()
				ib := State.Inbounds[id]
				State.mu.RUnlock()
				if ib == nil {
					writeJSON(w, http.StatusNotFound, map[string]any{"ok": false})
					return
				}
				writeJSON(w, http.StatusOK, map[string]any{"ok": true, "inbound": ib})
			default:
				writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"ok": false})
			}
		default:
			writeJSON(w, http.StatusNotFound, map[string]any{"ok": false})
		}
	}))

	// Reality tools
	mux.HandleFunc("/api/tools/generate-reality-keys", requireAuth(handleToolsRealityKeys))
	mux.HandleFunc("/api/tools/reality-settings", requireAuth(handleRealitySettings))

	// Nodes
	mux.HandleFunc("/api/nodes", requireAuth(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleNodesList(w, r)
		case http.MethodPost:
			handleNodeAdd(w, r)
		default:
			writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"ok": false})
		}
	}))
	mux.HandleFunc("/api/nodes/", requireAuth(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/nodes/")
		if path == "sync-all" {
			handleNodeSyncAll(w, r)
			return
		}
		parts := strings.SplitN(path, "/", 2)
		id := parts[0]
		sub := ""
		if len(parts) > 1 {
			sub = parts[1]
		}
		switch sub {
		case "refresh", "check", "health":
			handleNodeRefresh(w, r, id)
		case "":
			switch r.Method {
			case http.MethodPatch:
				handleNodeUpdate(w, r, id)
			case http.MethodDelete:
				handleNodeDelete(w, r, id)
			}
		default:
			writeJSON(w, http.StatusNotFound, map[string]any{"ok": false})
		}
	}))

	// Scanner
	mux.HandleFunc("/api/scanner/cf-subnets", requireAuth(handleScannerCFSubnets))
	mux.HandleFunc("/api/scanner/resolve", requireAuth(handleScannerResolve))
	mux.HandleFunc("/api/scanner/sni-check", requireAuth(handleScannerSNICheck))
	mux.HandleFunc("/api/scanner/ping-batch", requireAuth(handleScannerPingBatch))
	mux.HandleFunc("/api/scanner/save", requireAuth(handleScannerSave))
	mux.HandleFunc("/api/scanner/sni-list", requireAuth(handleScannerSNIList))
	mux.HandleFunc("/api/scanner/sni-scan-list", requireAuth(handleScannerSNIScanList))
	mux.HandleFunc("/api/scanner/sni-results", requireAuth(handleScannerSNIResults))
	mux.HandleFunc("/api/scanner/sni-save-results", requireAuth(handleScannerSNISaveResults))
	mux.HandleFunc("/api/scanner/sni-clear-results", requireAuth(handleScannerSNIClearResults))
	mux.HandleFunc("/api/scanner/sni-fastest", requireAuth(handleScannerSNIFastest))
	mux.HandleFunc("/api/scanner/ips/", requireAuth(func(w http.ResponseWriter, r *http.Request) {
		ctype := strings.TrimPrefix(r.URL.Path, "/api/scanner/ips/")
		handleScannerIPS(w, r, ctype)
	}))

	// Bot
	mux.HandleFunc("/api/bot/status", requireAuth(handleBotStatus))
	mux.HandleFunc("/api/bot/config", requireAuth(handleBotConfig))
	mux.HandleFunc("/api/bot/test", requireAuth(handleBotTest))
	mux.HandleFunc("/api/bot/sell/plans", requireAuth(handleBotSellPlans))

	// Worker
	mux.HandleFunc("/api/worker", requireAuth(handleWorkerRoute))
	mux.HandleFunc("/api/worker/", requireAuth(handleWorkerRoute))

	// Settings / backup
	mux.HandleFunc("/api/settings", requireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handleSettingsUpdate(w, r)
			return
		}
		handleSettingsGet(w, r)
	}))
	mux.HandleFunc("/api/settings/backup", requireAuth(handleSettingsBackup))
	mux.HandleFunc("/api/settings/restore", requireAuth(handleSettingsRestore))
	mux.HandleFunc("/api/change-password", requireAuth(handleChangePassword))
	mux.HandleFunc("/api/panel-api-key", requireAuth(handlePanelApiKey))
	mux.HandleFunc("/api/me/generate-key", requireAuth(handlePanelApiKey))

	// Misc
	mux.HandleFunc("/api/notifications", requireAuth(handleNotifications))
	mux.HandleFunc("/api/update-check", requireAuth(handleUpdateCheck))
	mux.HandleFunc("/api/activity", requireAuth(handleActivityLog))
	mux.HandleFunc("/api/musix", requireAuth(handleMusix))
	mux.HandleFunc("/api/tools/my-ip", requireAuth(handleToolsMyIP))

	log.Printf("==================================================")
	log.Printf("  White Panel Go Engine v8.0")
	log.Printf("  Listening on :%s", port)
	log.Printf("  Admin:  http://localhost:%s/", port)
	log.Printf("  Login:  http://localhost:%s/login", port)
	log.Printf("  Sub:    http://localhost:%s/sub/{uuid}", port)
	log.Printf("  Link:   http://localhost:%s/link/{uuid}", port)
	log.Printf("==================================================")

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("server terminated: %v", err)
	}
}
