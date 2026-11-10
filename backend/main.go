// Made By KILLMASTER https://github.com/Itskillmaster

package main

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/ecdh"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/skip2/go-qrcode"
)

// ============================================================================
// MODELS
// ============================================================================

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
	NamePrefix          string         `json:"name_prefix,omitempty"`
	Path                string         `json:"path,omitempty"`
	Host                string         `json:"host,omitempty"`
	ALPN                string         `json:"alpn,omitempty"`
	Flow                string         `json:"flow,omitempty"`
	PacketEncoding      string         `json:"packet_encoding,omitempty"`
	AllowInsecure       bool           `json:"allow_insecure,omitempty"`
	AdvancedSettings    map[string]any `json:"advanced_settings,omitempty"`
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
	ProxyCountries        []string            `json:"proxy_countries,omitempty"`
	ProxyCountry          string              `json:"proxy_country,omitempty"`
	CustomIPType          string              `json:"custom_ip_type,omitempty"`
	CustomIPInbounds      map[string][]string `json:"custom_ip_inbounds,omitempty"`
	SNISpoofV2Box         bool                `json:"sni_spoof_v2box,omitempty"`
	TelegramSecret        string              `json:"telegram_secret,omitempty"`
	ShadowsocksPassword   string              `json:"shadowsocks_password,omitempty"`
	WireguardKeys         map[string]string   `json:"wireguard_keys,omitempty"`
	NodeConfigs           map[string]string   `json:"node_configs,omitempty"`
	NodeSyncState         map[string]any      `json:"node_sync_state,omitempty"`
	NodeTraffic           map[string]int64    `json:"node_traffic,omitempty"`
	NodeTrafficUsedBytes  int64               `json:"node_traffic_used_bytes"`
	Note                  string              `json:"note,omitempty"`
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

// ============================================================================
// STATE
// ============================================================================

type AppState struct {
	mu sync.RWMutex

	Users    map[string]*User    `json:"users"`
	Inbounds map[string]*Inbound `json:"inbounds"`
	Nodes    map[string]*Node    `json:"nodes"`
	BotPlans map[string]*BotPlan `json:"bot_plans"`
	Settings map[string]any      `json:"settings"`
	Password string              `json:"password_hash"`

	DataDir string `json:"-"`
}

var serverStartTime = time.Now()

// ✅ NEW: پورت واقعی پنل، برای ساخت کانفیگ‌های درست در حالت بدون reverse proxy
var panelPortNum int = 8080

var State = &AppState{
	Users:    make(map[string]*User),
	Inbounds: make(map[string]*Inbound),
	Nodes:    make(map[string]*Node),
	BotPlans: make(map[string]*BotPlan),
	Settings: map[string]any{
		"domain":             "",
		"panel_api_key":      "",
		"max_ip_per_user":    5,
		"server_ip":          "",
		"country":            "IR",
		"country_code":       "IR",
		"country_flag":       "🇮🇷",
		"config_name_prefix": "White-",
	},
	DataDir: "./data",
}

var (
	sessionMu sync.Mutex
	sessions  = make(map[string]time.Time)

	userIPTrackerMu sync.Mutex
	userIPTracker   = make(map[string]map[string]time.Time)
)

func getMaxIPPerUser() int {
	State.mu.RLock()
	defer State.mu.RUnlock()
	v, ok := State.Settings["max_ip_per_user"]
	if !ok {
		return 5
	}
	switch t := v.(type) {
	case int:
		if t > 0 {
			return t
		}
	case int64:
		if t > 0 {
			return int(t)
		}
	case float64:
		if t > 0 {
			return int(t)
		}
	case json.Number:
		if n, err := t.Int64(); err == nil && n > 0 {
			return int(n)
		}
	case string:
		if n, err := strconv.Atoi(t); err == nil && n > 0 {
			return n
		}
	}
	return 5
}

func recordUserIP(uuid, clientIP string, maxAllowed int) bool {
	if clientIP == "" || clientIP == "127.0.0.1" {
		return true
	}
	userIPTrackerMu.Lock()
	defer userIPTrackerMu.Unlock()

	ips, ok := userIPTracker[uuid]
	if !ok {
		ips = make(map[string]time.Time)
		userIPTracker[uuid] = ips
	}

	now := time.Now()
	for ip, last := range ips {
		if now.Sub(last) > 10*time.Minute {
			delete(ips, ip)
		}
	}

	if maxAllowed <= 0 {
		maxAllowed = 5
	}

	if _, exists := ips[clientIP]; exists {
		ips[clientIP] = now
		return true
	}

	if len(ips) >= maxAllowed {
		var oldestIP string
		var oldestTime time.Time
		first := true
		for ip, t := range ips {
			if first || t.Before(oldestTime) {
				oldestIP = ip
				oldestTime = t
				first = false
			}
		}
		if oldestIP != "" && now.Sub(oldestTime) > 2*time.Minute {
			delete(ips, oldestIP)
		} else {
			return false
		}
	}

	ips[clientIP] = now
	return true
}

func extractRemoteIP(r *http.Request) string {
	if xff := r.Header.Get("CF-Connecting-IP"); xff != "" {
		return strings.TrimSpace(xff)
	}
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}

const sessionTTL = 7 * 24 * time.Hour

// ============================================================================
// STORAGE
// ============================================================================

func initStorage() {
	if env := os.Getenv("WHITE_DATA_DIR"); env != "" {
		State.DataDir = env
	}
	_ = os.MkdirAll(State.DataDir, 0755)
	seedScannerFiles()
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
		State.Inbounds["default"] = &Inbound{
			InboundID:   "default",
			Name:        "پیش‌فرض TLS + WS",
			Protocol:    "vless",
			InboundType: "transport",
			Port:        443,
			Network:     "ws",
			Security:    "tls",
			Fingerprint: "chrome",
			WSSettings:  map[string]any{"path": "/ws/{uuid}"},
			CreatedAt:   time.Now().Format(time.RFC3339),
		}
		State.Inbounds["Node"] = &Inbound{
			InboundID: "Node", Name: "Node", Protocol: "node",
			InboundType: "node", System: true, CreatedAt: time.Now().Format(time.RFC3339),
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

// ============================================================================
// BASIC HELPERS
// ============================================================================

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

func generateWireGuardKeypair() (string, string, error) {
	curve := ecdh.X25519()
	priv, err := curve.GenerateKey(rand.Reader)
	if err != nil {
		return "", "", err
	}
	return base64.StdEncoding.EncodeToString(priv.Bytes()),
		base64.StdEncoding.EncodeToString(priv.PublicKey().Bytes()), nil
}

func deriveWireGuardPublic(privB64 string) (string, error) {
	curve := ecdh.X25519()
	raw, err := base64.StdEncoding.DecodeString(privB64)
	if err != nil {
		return "", err
	}
	priv, err := curve.NewPrivateKey(raw)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(priv.PublicKey().Bytes()), nil
}

func generateSS2022Key(method string) string {
	n := 32
	if strings.Contains(strings.ToLower(method), "aes-128") {
		n = 16
	}
	b := make([]byte, n)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

func isSS2022Method(method string) bool {
	return strings.HasPrefix(strings.ToLower(method), "2022-")
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func readBody(r *http.Request) map[string]any {
	var m map[string]any
	body, _ := io.ReadAll(r.Body)
	json.Unmarshal(body, &m)
	if m == nil {
		m = map[string]any{}
	}
	return m
}

func gs(m map[string]any, key string) string {
	if v, ok := m[key]; ok && v != nil {
		return fmt.Sprintf("%v", v)
	}
	return ""
}
func gi(m map[string]any, key string) int {
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
func gf(m map[string]any, key string) float64 {
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
func gb(m map[string]any, key string) bool {
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}
func gss(m map[string]any, key string) []string {
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
func gm(m map[string]any, key string) map[string]any {
	if v, ok := m[key]; ok {
		if mp, ok := v.(map[string]any); ok {
			return mp
		}
	}
	return nil
}

func hostFromRequest(r *http.Request) string {
	if r == nil {
		return "127.0.0.1"
	}
	if h := r.Header.Get("X-Forwarded-Host"); h != "" {
		if i := strings.Index(h, ","); i > 0 {
			h = h[:i]
		}
		return strings.TrimSpace(h)
	}
	h := r.Host
	if host, _, err := net.SplitHostPort(h); err == nil {
		return host
	}
	return h
}

// ✅ NEW: استخراج هاست تمیز از هر ورودی. اگر معتبر نبود، "" برمی‌گرداند.
func sanitizeHost(raw string) string {
	s := strings.TrimSpace(raw)
	if s == "" {
		return ""
	}
	// حذف تکراری اسکیم‌ها (ممکن است "https://https://x" باشد)
	for {
		lower := strings.ToLower(s)
		trimmed := false
		for _, p := range []string{"https://", "http://", "wss://", "ws://", "grpc://", "tcp://", "vless://", "ss://"} {
			if strings.HasPrefix(lower, p) {
				s = s[len(p):]
				trimmed = true
				break
			}
		}
		if !trimmed {
			break
		}
	}
	// حذف مسیر/کوئری/fragment
	if i := strings.IndexAny(s, "/?#"); i >= 0 {
		s = s[:i]
	}
	// حذف پورت
	if h, _, err := net.SplitHostPort(s); err == nil {
		s = h
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	// رد کردن مقادیر نامعتبر رایج
	switch strings.ToLower(s) {
	case "localhost", "127.0.0.1", "::1", "0.0.0.0",
		"example.com", "your-domain.com",
		"https", "http", "ws", "wss", "grpc", "tcp", "vless", "ss":
		return ""
	}
	// باید یا IP معتبر باشد یا حداقل یک نقطه داشته باشد
	if net.ParseIP(s) == nil && !strings.Contains(s, ".") {
		return ""
	}
	return s
}

// ✅ NEW: پورت واقعی پنل را از env می‌خواند.
func initPanelPort() {
	p := os.Getenv("PANEL_PORT")
	if p == "" {
		p = os.Getenv("PORT")
	}
	if p == "" {
		p = "8080"
	}
	if n, err := strconv.Atoi(p); err == nil && n > 0 {
		panelPortNum = n
	} else {
		panelPortNum = 8080
	}
}

// ✅ NEW: بهترین هاست معتبر را از همه‌ی منابع پیدا می‌کند.
func resolvePanelHost(r *http.Request) string {
	State.mu.RLock()
	cands := []string{}
	for _, k := range []string{"domain", "bot_domain", "server_ip", "public_ip"} {
		if v, _ := State.Settings[k].(string); v != "" {
			cands = append(cands, v)
		}
	}
	if sell, ok := State.Settings["bot_sell"].(map[string]any); ok {
		if v, _ := sell["domain"].(string); v != "" {
			cands = append(cands, v)
		}
	}
	State.mu.RUnlock()

	for _, c := range cands {
		if h := sanitizeHost(c); h != "" {
			return h
		}
	}
	if r != nil {
		if h := sanitizeHost(hostFromRequest(r)); h != "" {
			return h
		}
	}
	return ""
}

// ✅ NEW: آیا کاربر domain معتبر تنظیم کرده؟ (یعنی احتمالاً reverse proxy روی 443 دارد)
func hasReverseProxyDomain() bool {
	State.mu.RLock()
	defer State.mu.RUnlock()
	d, _ := State.Settings["domain"].(string)
	return sanitizeHost(d) != ""
}

func discoveredHost(r *http.Request) string {
	return resolvePanelHost(r)
}

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

// ============================================================================
// AUTH
// ============================================================================

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
	body := readBody(r)
	password := gs(body, "password")
	State.mu.RLock()
	currentHash := State.Password
	State.mu.RUnlock()
	if subtle.ConstantTimeCompare([]byte(hashPassword(password)), []byte(currentHash)) != 1 {
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
		"host": discoveredHost(r), "xray_running": xrayIsRunning(),
	})
}

func handleChangePassword(w http.ResponseWriter, r *http.Request) {
	body := readBody(r)
	cur := gs(body, "current_password")
	newPwd := gs(body, "new_password")
	State.mu.RLock()
	currentHash := State.Password
	State.mu.RUnlock()
	if subtle.ConstantTimeCompare([]byte(hashPassword(cur)), []byte(currentHash)) != 1 {
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

// ============================================================================
// XRAY PROCESS MANAGER (unchanged)
// ============================================================================

var (
	xrayMu      sync.Mutex
	xrayProc    *exec.Cmd
	xrayRunning atomic.Bool
	xrayApplyMu sync.Mutex
	xrayTimer   *time.Timer
	xrayTimerMu sync.Mutex
)

func xrayDir() string {
	d, _ := filepath.Abs("./xray")
	_ = os.MkdirAll(d, 0755)
	return d
}

func xrayBinPath() string {
	local := filepath.Join(xrayDir(), "xray")
	if runtime.GOOS == "windows" {
		local = filepath.Join(xrayDir(), "xray.exe")
	}
	if st, err := os.Stat(local); err == nil && st.Size() > 10000 {
		return local
	}
	dataBin := filepath.Join(State.DataDir, "xray", "xray")
	if runtime.GOOS == "windows" {
		dataBin = filepath.Join(State.DataDir, "xray", "xray.exe")
	}
	if st, err := os.Stat(dataBin); err == nil && st.Size() > 10000 {
		return dataBin
	}
	if p, err := exec.LookPath("xray"); err == nil {
		return p
	}
	if runtime.GOOS == "windows" {
		if p, err := exec.LookPath("xray.exe"); err == nil {
			return p
		}
	}
	for _, p := range []string{"/usr/local/bin/xray", "/usr/bin/xray", "/bin/xray", "/opt/xray/xray"} {
		if st, err := os.Stat(p); err == nil && st.Size() > 10000 {
			return p
		}
	}
	return local
}
func xrayCfgPath() string { return filepath.Join(xrayDir(), "config.json") }
func xrayLogPath() string { return filepath.Join(xrayDir(), "xray-runtime.log") }

func xrayIsRunning() bool { return xrayRunning.Load() }

func ensureXray() error {
	bin := xrayBinPath()
	if st, err := os.Stat(bin); err == nil && st.Size() > 100000 {
		return nil
	}
	var arch string
	switch runtime.GOARCH {
	case "amd64":
		arch = "64"
	case "arm64":
		arch = "arm64-v8a"
	case "arm":
		arch = "arm32-v7a"
	case "386":
		arch = "32"
	default:
		return fmt.Errorf("unsupported architecture: %s", runtime.GOARCH)
	}
	osName := "linux"
	if runtime.GOOS == "windows" {
		osName = "windows"
	} else if runtime.GOOS == "darwin" {
		osName = "macos"
	}
	urls := []string{
		fmt.Sprintf("https://github.com/XTLS/Xray-core/releases/latest/download/Xray-%s-%s.zip", osName, arch),
		fmt.Sprintf("https://mirror.ghproxy.com/https://github.com/XTLS/Xray-core/releases/latest/download/Xray-%s-%s.zip", osName, arch),
		fmt.Sprintf("https://ghproxy.net/https://github.com/XTLS/Xray-core/releases/latest/download/Xray-%s-%s.zip", osName, arch),
		fmt.Sprintf("https://github.com/XTLS/Xray-core/releases/download/v25.12.8/Xray-%s-%s.zip", osName, arch),
	}
	client := &http.Client{Timeout: 35 * time.Second}
	var lastErr error
	for _, u := range urls {
		log.Printf("[Xray] trying download: %s", u)
		resp, err := client.Get(u)
		if err != nil {
			lastErr = err
			continue
		}
		if resp.StatusCode != 200 {
			resp.Body.Close()
			lastErr = fmt.Errorf("HTTP %d", resp.StatusCode)
			continue
		}
		tmpZip := filepath.Join(xrayDir(), "xray.zip")
		f, err := os.Create(tmpZip)
		if err != nil {
			resp.Body.Close()
			lastErr = err
			continue
		}
		_, err = io.Copy(f, resp.Body)
		resp.Body.Close()
		f.Close()
		if err != nil {
			lastErr = err
			continue
		}
		if err := extractXrayBinary(tmpZip, bin); err != nil {
			lastErr = err
			continue
		}
		os.Remove(tmpZip)
		_ = os.Chmod(bin, 0755)
		log.Printf("[Xray] installed: %s (%d bytes)", bin, mustStatSize(bin))
		return nil
	}
	return fmt.Errorf("all download attempts failed: %v", lastErr)
}

func extractXrayBinary(zipPath, destBin string) error {
	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer zr.Close()
	targetName := "xray"
	if runtime.GOOS == "windows" {
		targetName = "xray.exe"
	}
	for _, zf := range zr.File {
		base := filepath.Base(zf.Name)
		if strings.EqualFold(base, targetName) || strings.EqualFold(base, "xray") || strings.EqualFold(base, "xray.exe") {
			rc, err := zf.Open()
			if err != nil {
				return err
			}
			out, err := os.OpenFile(destBin, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
			if err != nil {
				rc.Close()
				return err
			}
			_, copyErr := io.Copy(out, rc)
			out.Close()
			rc.Close()
			if copyErr != nil {
				return copyErr
			}
			return nil
		}
	}
	return errors.New("xray binary not found in zip")
}

func mustStatSize(p string) int64 {
	if st, err := os.Stat(p); err == nil {
		return st.Size()
	}
	return 0
}

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
	} else {
		if derived, err := deriveX25519Public(priv); err == nil {
			pub = derived
		}
	}
	sid, _ := out["short_id"].(string)
	if sid == "" {
		if arr, ok := out["short_ids"].([]any); ok && len(arr) > 0 {
			sid = fmt.Sprint(arr[0])
		} else if arr, ok := out["short_ids"].([]string); ok && len(arr) > 0 {
			sid = arr[0]
		}
	}
	if !regexp.MustCompile(`^[0-9a-f]{2,16}$`).MatchString(sid) || len(sid)%2 != 0 {
		sid = generateShortID()[:10]
	}
	out["private_key"] = priv
	out["public_key"] = pub
	out["short_id"] = sid
	if _, ok := out["rx_path"]; !ok {
		out["rx_path"] = "/"
	}
	sni, _ := out["sni"].(string)
	if sni == "" {
		sni = "is1-ssl.mzstatic.com"
		out["sni"] = sni
	}
	dest, _ := out["dest"].(string)
	if dest == "" {
		out["dest"] = sni + ":443"
	}
	return out
}

func isActiveUser(u *User) bool {
	if u == nil || u.Status != "active" {
		return false
	}
	if u.ExpireAt != "" {
		if t, err := time.Parse(time.RFC3339, u.ExpireAt); err == nil && time.Now().After(t) {
			return false
		}
	}
	if u.TrafficLimitBytes > 0 && u.TrafficUsedBytes >= u.TrafficLimitBytes {
		return false
	}
	return true
}

func generateXrayConfig() ([]byte, error) {
	type xInbound struct {
		Tag            string         `json:"tag"`
		Listen         string         `json:"listen"`
		Port           int            `json:"port"`
		Protocol       string         `json:"protocol"`
		Settings       map[string]any `json:"settings"`
		StreamSettings map[string]any `json:"streamSettings"`
		Sniffing       map[string]any `json:"sniffing"`
	}
	config := map[string]any{
		"log":      map[string]any{"loglevel": "warning"},
		"inbounds": []xInbound{},
		"outbounds": []map[string]any{
			{"protocol": "freedom", "tag": "direct"},
			{"protocol": "blackhole", "tag": "blocked"},
		},
		"api": map[string]any{
			"tag":      "api",
			"services": []string{"StatsService"},
		},
		"stats": map[string]any{},
		"policy": map[string]any{
			"levels": map[string]any{
				"0": map[string]any{
					"statsUserUplink":   true,
					"statsUserDownlink": true,
				},
			},
			"system": map[string]any{
				"statsInboundUplink":   true,
				"statsInboundDownlink": true,
			},
		},
		"routing": map[string]any{
			"domainStrategy": "IPIfNonMatch",
			"rules": []any{
				map[string]any{
					"type":        "field",
					"inboundTag":  []string{"api"},
					"outboundTag": "api",
				},
				map[string]any{
					"type":        "field",
					"outboundTag": "blocked",
					"ip":          []string{"geoip:private"},
				},
			},
		},
	}

	panelPort := os.Getenv("PANEL_PORT")
	if panelPort == "" {
		panelPort = os.Getenv("PORT")
	}
	if panelPort == "" {
		panelPort = "8080"
	}
	panelPortN, _ := strconv.Atoi(panelPort)

	inbounds := []xInbound{}
	usedPorts := map[int]bool{xrayStatsPort: true}
	if panelPortN > 0 {
		usedPorts[panelPortN] = true
	}

	State.mu.RLock()
	usersByInbound := map[string][]*User{}
	for _, u := range State.Users {
		if !isActiveUser(u) {
			continue
		}
		for _, iid := range u.InboundIDs {
			usersByInbound[iid] = append(usersByInbound[iid], u)
		}
	}
	stateInbounds := make([]*Inbound, 0, len(State.Inbounds))
	for _, ib := range State.Inbounds {
		stateInbounds = append(stateInbounds, ib)
	}
	State.mu.RUnlock()

	for _, ib := range stateInbounds {
		if ib.System || ib.Port <= 0 {
			continue
		}
		proto := strings.ToLower(ib.Protocol)
		sec := strings.ToLower(ib.Security)
		isReality := proto == "reality" || sec == "reality"
		isSS := proto == "shadowsocks"
		isWG := proto == "wireguard"
		if !isReality && !isSS && !isWG {
			continue
		}
		if usedPorts[ib.Port] {
			log.Printf("[Xray] port %d already in use (panel or another inbound), skipping %s", ib.Port, ib.Name)
			continue
		}
		usedPorts[ib.Port] = true

		xi := xInbound{
			Tag:      "inbound-" + ib.InboundID,
			Listen:   "0.0.0.0",
			Port:     ib.Port,
			Protocol: "vless",
			Sniffing: map[string]any{"enabled": true, "destOverride": []string{"http", "tls", "quic"}},
		}

		if isSS {
			xi.Protocol = "shadowsocks"
			ss := ib.ShadowsocksSettings
			if ss == nil {
				ss = map[string]any{}
			}
			method := strings.ToLower(fmt.Sprint(ss["method"]))
			if method == "" || method == "<nil>" {
				method = "aes-256-gcm"
			}
			network := fmt.Sprint(ss["network"])
			if network == "" || network == "<nil>" {
				network = "tcp,udp"
			}
			serverPwd := fmt.Sprint(ss["password"])
			if serverPwd == "" || serverPwd == "<nil>" {
				if isSS2022Method(method) {
					serverPwd = generateSS2022Key(method)
				} else {
					serverPwd = generateToken(16)
				}
				State.mu.Lock()
				ss["password"] = serverPwd
				ib.ShadowsocksSettings = ss
				State.mu.Unlock()
			}
			users := []map[string]any{}
			for _, u := range usersByInbound[ib.InboundID] {
				upwd := u.ShadowsocksPassword
				if upwd == "" {
					upwd = generateToken(16)
					State.mu.Lock()
					u.ShadowsocksPassword = upwd
					State.mu.Unlock()
				}
				if isSS2022Method(method) {
					users = append(users, map[string]any{"password": serverPwd + ":" + upwd, "email": u.ConfigUUID})
				} else {
					users = append(users, map[string]any{"password": upwd, "method": method, "email": u.ConfigUUID})
				}
			}
			if len(users) == 0 {
				if isSS2022Method(method) {
					users = append(users, map[string]any{"password": serverPwd + ":" + generateToken(16)})
				} else {
					users = append(users, map[string]any{"password": generateToken(16), "method": method})
				}
			}
			settings := map[string]any{
				"network": network, "method": method, "users": users,
			}
			if !isSS2022Method(method) {
				settings["password"] = serverPwd
			}
			xi.Settings = settings
			xi.StreamSettings = map[string]any{"network": "tcp"}
			inbounds = append(inbounds, xi)
			continue
		}

		if isWG {
			xi.Protocol = "wireguard"
			wg := ib.WireguardSettings
			if wg == nil {
				wg = map[string]any{}
			}
			serverPriv := fmt.Sprint(wg["server_private_key"])
			if serverPriv == "" || serverPriv == "<nil>" {
				priv, pub, err := generateWireGuardKeypair()
				if err == nil {
					serverPriv = priv
					State.mu.Lock()
					wg["server_private_key"] = priv
					wg["server_public_key"] = pub
					ib.WireguardSettings = wg
					State.mu.Unlock()
				}
			}
			mtu := gi(wg, "mtu")
			if mtu == 0 {
				mtu = 1420
			}
			subnet := fmt.Sprint(wg["address_subnet"])
			if subnet == "" || subnet == "<nil>" {
				subnet = "10.44.0"
			}
			peers := []map[string]any{}
			for _, u := range usersByInbound[ib.InboundID] {
				keys := u.WireguardKeys
				if keys == nil || keys["public_key"] == "" {
					priv, pub, err := generateWireGuardKeypair()
					if err != nil {
						continue
					}
					keys = map[string]string{"private_key": priv, "public_key": pub}
					State.mu.Lock()
					u.WireguardKeys = keys
					State.mu.Unlock()
				}
				hsh := sha256.Sum256([]byte(u.ConfigUUID))
				h := 2 + int(hsh[0])%253
				peers = append(peers, map[string]any{
					"publicKey":  keys["public_key"],
					"allowedIPs": []string{fmt.Sprintf("%s.%d/32", subnet, h)},
				})
			}
			xi.Settings = map[string]any{
				"secretKey": serverPriv,
				"address":   []string{fmt.Sprintf("%s.1/24", subnet)},
				"mtu":       mtu,
				"peers":     peers,
			}
			xi.StreamSettings = map[string]any{"network": "tcp"}
			inbounds = append(inbounds, xi)
			continue
		}

		// Reality (VLESS)
		rs := ib.RealitySettings
		if rs == nil {
			rs = realityDefaults(nil)
			State.mu.Lock()
			ib.RealitySettings = rs
			State.mu.Unlock()
		}
		priv, _ := rs["private_key"].(string)
		if priv == "" {
			rs = realityDefaults(rs)
			State.mu.Lock()
			ib.RealitySettings = rs
			State.mu.Unlock()
			priv, _ = rs["private_key"].(string)
		}
		network := ib.Network
		if network == "" {
			network = "tcp"
		}
		ibFlow := strings.TrimSpace(ib.Flow)
		if network != "tcp" {
			ibFlow = ""
		}
		clients := []map[string]any{}
		for _, u := range usersByInbound[ib.InboundID] {
			clients = append(clients, map[string]any{
				"id":    u.ConfigUUID,
				"flow":  ibFlow,
				"email": u.ConfigUUID,
			})
		}
		if len(clients) == 0 {
			clients = append(clients, map[string]any{"id": generateUUID(), "flow": ibFlow})
		}
		xi.Settings = map[string]any{"clients": clients, "decryption": "none"}
		sni := ib.SNI
		if sni == "" {
			sni, _ = rs["sni"].(string)
		}
		if sni == "" {
			sni = "is1-ssl.mzstatic.com"
		}
		dest, _ := rs["dest"].(string)
		if dest == "" {
			dest = sni + ":443"
		} else if !strings.Contains(dest, ":") {
			dest = dest + ":443"
		}
		serverNames := []string{sni}
		if names, ok := rs["server_names"].([]string); ok && len(names) > 0 {
			serverNames = names
		}
		sid, _ := rs["short_id"].(string)
		realitySettings := map[string]any{
			"show":        false,
			"dest":        dest,
			"xver":        0,
			"serverNames": serverNames,
			"privateKey":  priv,
			"shortIds":    []string{sid},
		}
		stream := map[string]any{
			"network":         network,
			"security":        "reality",
			"realitySettings": realitySettings,
		}
		if network == "xhttp" {
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
			xhttp := map[string]any{
				"path":               xpath,
				"mode":               xmode,
				"xPaddingBytes":      xs["xPaddingBytes"],
				"scMaxEachPostBytes": xs["scMaxEachPostBytes"],
			}
			if xhttp["xPaddingBytes"] == nil {
				xhttp["xPaddingBytes"] = "100-1000"
			}
			if xhttp["scMaxEachPostBytes"] == nil {
				xhttp["scMaxEachPostBytes"] = "1000000"
			}
			stream["xhttpSettings"] = xhttp
		}
		xi.StreamSettings = stream
		inbounds = append(inbounds, xi)
	}

	inbounds = append(inbounds, xInbound{
		Tag:      "api",
		Listen:   "127.0.0.1",
		Port:     xrayStatsPort,
		Protocol: "dokodemo-door",
		Settings: map[string]any{"address": "127.0.0.1"},
	})

	config["inbounds"] = inbounds
	return json.MarshalIndent(config, "", "  ")
}

func xrayApply() {
	xrayApplyMu.Lock()
	defer xrayApplyMu.Unlock()

	if err := ensureXray(); err != nil {
		log.Printf("[Xray] ensureXray failed: %v", err)
		return
	}
	cfgBytes, err := generateXrayConfig()
	if err != nil {
		log.Printf("[Xray] config generation failed: %v", err)
		return
	}
	if cfgBytes == nil {
		xrayStopLocked()
		log.Printf("[Xray] no real inbounds — Xray stopped")
		return
	}
	if err := os.WriteFile(xrayCfgPath(), cfgBytes, 0644); err != nil {
		log.Printf("[Xray] write config failed: %v", err)
		return
	}
	xrayStopLocked()

	logFile, err := os.OpenFile(xrayLogPath(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("[Xray] open log failed: %v", err)
		return
	}
	cmd := exec.Command(xrayBinPath(), "-c", xrayCfgPath())
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	if err := cmd.Start(); err != nil {
		logFile.Close()
		log.Printf("[Xray] start failed: %v", err)
		return
	}
	xrayMu.Lock()
	xrayProc = cmd
	xrayRunning.Store(true)
	xrayMu.Unlock()

	go func() {
		time.Sleep(1500 * time.Millisecond)
		xrayMu.Lock()
		p := xrayProc
		xrayMu.Unlock()
		if p != nil {
			log.Printf("[Xray] started (pid=%d)", p.Process.Pid)
		}
	}()
	go func() {
		err := cmd.Wait()
		xrayMu.Lock()
		if xrayProc == cmd {
			xrayRunning.Store(false)
			xrayProc = nil
		}
		xrayMu.Unlock()
		log.Printf("[Xray] exited: %v", err)
	}()
}

func xrayStopLocked() {
	xrayMu.Lock()
	p := xrayProc
	xrayProc = nil
	xrayRunning.Store(false)
	xrayMu.Unlock()
	if p != nil && p.Process != nil {
		_ = p.Process.Kill()
		time.Sleep(1 * time.Second)
	}
}

func xrayStop() {
	xrayApplyMu.Lock()
	defer xrayApplyMu.Unlock()
	xrayStopLocked()
}

func scheduleXrayApply() {
	xrayTimerMu.Lock()
	defer xrayTimerMu.Unlock()
	if xrayTimer != nil {
		xrayTimer.Stop()
	}
	xrayTimer = time.AfterFunc(700*time.Millisecond, func() {
		xrayApply()
		saveState()
	})
}

// ============================================================================
// VLESS HEADER PARSER
// ============================================================================

type vlessHeader struct {
	UUID    string
	Command byte
	Address string
	Port    int
	Payload []byte
}

func parseVlessHeader(buf []byte) (*vlessHeader, error) {
	if len(buf) < 24 {
		return nil, errors.New("header too short")
	}
	if buf[0] != 0x00 {
		return nil, errors.New("bad version")
	}
	u := buf[1:17]
	uuid := fmt.Sprintf("%x-%x-%x-%x-%x", u[0:4], u[4:6], u[6:8], u[8:10], u[10:16])
	pos := 17
	optLen := int(buf[pos])
	pos++
	pos += optLen
	if pos+3 > len(buf) {
		return nil, errors.New("truncated header")
	}
	command := buf[pos]
	pos++
	port := int(binary.BigEndian.Uint16(buf[pos : pos+2]))
	pos += 2
	addrType := buf[pos]
	pos++
	var addr string
	switch addrType {
	case 1:
		if pos+4 > len(buf) {
			return nil, errors.New("truncated IPv4")
		}
		addr = net.IP(buf[pos : pos+4]).String()
		pos += 4
	case 2:
		if pos+1 > len(buf) {
			return nil, errors.New("truncated domain len")
		}
		dlen := int(buf[pos])
		pos++
		if pos+dlen > len(buf) {
			return nil, errors.New("truncated domain")
		}
		addr = string(buf[pos : pos+dlen])
		pos += dlen
	case 3:
		if pos+16 > len(buf) {
			return nil, errors.New("truncated IPv6")
		}
		addr = net.IP(buf[pos : pos+16]).String()
		pos += 16
	default:
		return nil, fmt.Errorf("unknown addr type %d", addrType)
	}
	return &vlessHeader{
		UUID: uuid, Command: command, Address: addr, Port: port, Payload: buf[pos:],
	}, nil
}

// ============================================================================
// QUOTA ENFORCEMENT
// ============================================================================

var trafficDirty atomic.Bool

func findUserByConfigUUID(uuid string) *User {
	State.mu.RLock()
	defer State.mu.RUnlock()
	for _, u := range State.Users {
		if strings.EqualFold(u.ConfigUUID, uuid) {
			return u
		}
	}
	return nil
}

func checkAndUse(configUUID string, n int64) bool {
	State.mu.Lock()
	defer State.mu.Unlock()
	var u *User
	for _, candidate := range State.Users {
		if strings.EqualFold(candidate.ConfigUUID, configUUID) {
			u = candidate
			break
		}
	}
	if u == nil {
		return false
	}
	if !isActiveUser(u) {
		return false
	}
	if u.TrafficLimitBytes > 0 && u.TrafficUsedBytes+n > u.TrafficLimitBytes {
		return false
	}
	u.TrafficUsedBytes += n
	trafficDirty.Store(true)
	return true
}

// ============================================================================
// WEBSOCKET RELAY  /ws/{uuid}
// ============================================================================

var wsUpgrader = websocket.Upgrader{
	CheckOrigin:     func(r *http.Request) bool { return true },
	ReadBufferSize:  32 * 1024,
	WriteBufferSize: 32 * 1024,
}

func hasWSRelayAccess(u *User) bool {
	if u == nil {
		return false
	}
	State.mu.RLock()
	defer State.mu.RUnlock()
	if len(u.InboundIDs) == 0 {
		return true
	}
	for _, iid := range u.InboundIDs {
		if iid == "default" {
			return true
		}
		ib := State.Inbounds[iid]
		if ib == nil {
			continue
		}
		p := strings.ToLower(ib.Protocol)
		net := strings.ToLower(ib.Network)
		if net == "" {
			net = "ws"
		}
		if p == "worker" || (p == "vless" && net == "ws") || ib.InboundID == "default" {
			return true
		}
	}
	return false
}

func handleWSRelay(w http.ResponseWriter, r *http.Request, uuid string) {
	u := findUserByConfigUUID(uuid)
	if u == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if !isActiveUser(u) {
		http.Error(w, "user not active", http.StatusForbidden)
		return
	}
	if !hasWSRelayAccess(u) {
		log.Printf("[WS] uuid=%s has no WS inbound access", uuid)
		http.Error(w, "no ws inbound", http.StatusForbidden)
		return
	}

	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[WS] upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(15 * time.Second))
	msgType, firstMsg, err := conn.ReadMessage()
	if err != nil {
		return
	}
	if msgType != websocket.BinaryMessage {
		return
	}
	header, err := parseVlessHeader(firstMsg)
	if err != nil {
		log.Printf("[WS] bad VLESS header: %v", err)
		return
	}
	if !strings.EqualFold(header.UUID, uuid) {
		log.Printf("[WS] uuid mismatch: header=%s path=%s", header.UUID, uuid)
		return
	}
	if header.Command != 0x01 {
		return
	}
	if !checkAndUse(uuid, int64(len(firstMsg))) {
		return
	}

	clientIP := extractRemoteIP(r)
	maxIP := u.ConcurrentConnections
	if maxIP <= 0 {
		maxIP = getMaxIPPerUser()
	}
	if !recordUserIP(uuid, clientIP, maxIP) {
		log.Printf("[Security] User %s exceeded max concurrent IPs (%d). Connection dropped from %s", u.Username, maxIP, clientIP)
		return
	}

	target, err := net.DialTimeout("tcp", net.JoinHostPort(header.Address, strconv.Itoa(header.Port)), 10*time.Second)
	if err != nil {
		log.Printf("[WS] dial %s:%d failed: %v", header.Address, header.Port, err)
		return
	}
	defer target.Close()

	if tcpConn, ok := target.(*net.TCPConn); ok {
		_ = tcpConn.SetKeepAlive(true)
		_ = tcpConn.SetKeepAlivePeriod(30 * time.Second)
	}

	if len(header.Payload) > 0 {
		target.Write(header.Payload)
	}

	const pongWait = 60 * time.Second
	const pingPeriod = (pongWait * 9) / 10

	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPingHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return conn.WriteControl(websocket.PongMessage, []byte{}, time.Now().Add(5*time.Second))
	})
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	var wsWriteMu sync.Mutex
	writeWS := func(messageType int, data []byte) error {
		wsWriteMu.Lock()
		defer wsWriteMu.Unlock()
		conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
		return conn.WriteMessage(messageType, data)
	}

	done := make(chan struct{}, 2)

	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()
	go func() {
		for {
			select {
			case <-ticker.C:
				wsWriteMu.Lock()
				err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(5*time.Second))
				wsWriteMu.Unlock()
				if err != nil {
					target.Close()
					conn.Close()
					return
				}
			case <-done:
				return
			}
		}
	}()

	go func() {
		defer func() {
			target.Close()
			select {
			case done <- struct{}{}:
			default:
			}
		}()
		for {
			mt, data, err := conn.ReadMessage()
			if err != nil {
				return
			}
			conn.SetReadDeadline(time.Now().Add(pongWait))
			if mt != websocket.BinaryMessage || len(data) == 0 {
				continue
			}
			if !checkAndUse(uuid, int64(len(data))) {
				return
			}
			if _, err := target.Write(data); err != nil {
				return
			}
		}
	}()

	go func() {
		defer func() {
			conn.Close()
			select {
			case done <- struct{}{}:
			default:
			}
		}()
		buf := make([]byte, 32*1024)
		firstChunk := true
		for {
			n, err := target.Read(buf)
			if n > 0 {
				if !checkAndUse(uuid, int64(n)) {
					return
				}
				chunk := buf[:n]
				if firstChunk {
					out := make([]byte, 2+n)
					copy(out[2:], chunk)
					chunk = out
					firstChunk = false
				}
				if err := writeWS(websocket.BinaryMessage, chunk); err != nil {
					return
				}
			}
			if err != nil {
				return
			}
		}
	}()

	<-done
}

// ============================================================================
// XHTTP RELAY
// ============================================================================

type xhttpSession struct {
	mu        sync.Mutex
	uuid      string
	tcpConn   net.Conn
	downQueue chan []byte
	closed    atomic.Bool
	lastSeen  time.Time
	seqBuf    map[int][]byte
	nextSeq   int
}

var (
	xhttpSessions      sync.Map
	xhttpReaperStarted sync.Once
)

func ensureXHTTPReaper() {
	xhttpReaperStarted.Do(func() {
		go func() {
			for {
				time.Sleep(30 * time.Second)
				now := time.Now()
				xhttpSessions.Range(func(k, v any) bool {
					s := v.(*xhttpSession)
					if now.Sub(s.lastSeen) > 60*time.Second && !s.closed.Load() {
						s.close()
					}
					return true
				})
			}
		}()
	})
}

func (s *xhttpSession) close() {
	if s.closed.Swap(true) {
		return
	}
	if s.tcpConn != nil {
		s.tcpConn.Close()
	}
	select {
	case s.downQueue <- nil:
	default:
	}
}

func (s *xhttpSession) openTCP(h *vlessHeader) error {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(h.Address, strconv.Itoa(h.Port)), 10*time.Second)
	if err != nil {
		return err
	}
	s.tcpConn = conn
	if len(h.Payload) > 0 {
		conn.Write(h.Payload)
	}
	go s.pumpDownstream()
	return nil
}

func (s *xhttpSession) pumpDownstream() {
	buf := make([]byte, 32*1024)
	first := true
	for {
		n, err := s.tcpConn.Read(buf)
		if n > 0 {
			if !checkAndUse(s.uuid, int64(n)) {
				s.close()
				return
			}
			chunk := make([]byte, n)
			copy(chunk, buf[:n])
			if first {
				out := make([]byte, 2+n)
				copy(out[2:], chunk)
				chunk = out
				first = false
			}
			select {
			case s.downQueue <- chunk:
			default:
			}
		}
		if err != nil {
			s.close()
			return
		}
	}
}

func getOrCreateXHTTPSession(sessionID, uuid string) *xhttpSession {
	if v, ok := xhttpSessions.Load(sessionID); ok {
		s := v.(*xhttpSession)
		s.lastSeen = time.Now()
		return s
	}
	s := &xhttpSession{
		uuid: uuid, downQueue: make(chan []byte, 256),
		lastSeen: time.Now(), seqBuf: make(map[int][]byte),
	}
	xhttpSessions.Store(sessionID, s)
	ensureXHTTPReaper()
	return s
}

func handleXHTTPDownlink(w http.ResponseWriter, r *http.Request, mode, uuid, sessionID string) {
	if mode != "packet-up" && mode != "stream-up" {
		http.Error(w, "unknown mode", 404)
		return
	}
	sess := getOrCreateXHTTPSession(sessionID, uuid)
	w.Header().Set("Content-Type", "application/grpc")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Accel-Buffering", "no")
	flusher, _ := w.(http.Flusher)
	w.WriteHeader(200)
	if flusher != nil {
		flusher.Flush()
	}
	ctx := r.Context()
	for {
		select {
		case chunk := <-sess.downQueue:
			if chunk == nil {
				return
			}
			sess.lastSeen = time.Now()
			if _, err := w.Write(chunk); err != nil {
				return
			}
			if flusher != nil {
				flusher.Flush()
			}
		case <-ctx.Done():
			return
		}
	}
}

func handleXHTTPStreamUp(w http.ResponseWriter, r *http.Request, uuid, sessionID string) {
	sess := getOrCreateXHTTPSession(sessionID, uuid)
	defer r.Body.Close()
	buf := make([]byte, 32*1024)
	var headerBuf []byte
	for {
		n, err := r.Body.Read(buf)
		if n > 0 {
			chunk := buf[:n]
			sess.lastSeen = time.Now()

			sess.mu.Lock()
			haveConn := sess.tcpConn != nil
			sess.mu.Unlock()

			if !haveConn {
				headerBuf = append(headerBuf, chunk...)
				if len(headerBuf) > 8192 {
					log.Printf("[XHTTP] header too large")
					return
				}
				header, perr := parseVlessHeader(headerBuf)
				if perr != nil {
					if err == nil {
						continue
					}
					return
				}
				if !strings.EqualFold(header.UUID, uuid) {
					return
				}
				if !checkAndUse(uuid, int64(len(headerBuf))) {
					sess.close()
					return
				}
				if err := sess.openTCP(header); err != nil {
					log.Printf("[XHTTP] dial failed: %v", err)
					return
				}
				headerBuf = nil
			} else {
				if !checkAndUse(uuid, int64(n)) {
					sess.close()
					return
				}
				sess.mu.Lock()
				if sess.tcpConn != nil {
					sess.tcpConn.Write(chunk)
				}
				sess.mu.Unlock()
			}
		}
		if err != nil {
			return
		}
	}
}

func handleXHTTPPacketUp(w http.ResponseWriter, r *http.Request, uuid, sessionID, seqStr string) {
	seq, err := strconv.Atoi(seqStr)
	if err != nil || seq < 0 {
		http.Error(w, "bad seq", 400)
		return
	}
	sess := getOrCreateXHTTPSession(sessionID, uuid)
	body, _ := io.ReadAll(r.Body)
	if len(body) == 0 {
		writeJSON(w, 200, map[string]any{"ok": true})
		return
	}
	if !checkAndUse(uuid, int64(len(body))) {
		sess.close()
		writeJSON(w, 403, map[string]any{"ok": false})
		return
	}
	sess.mu.Lock()
	defer sess.mu.Unlock()

	if len(sess.seqBuf) > 1024 {
		writeJSON(w, http.StatusTooManyRequests, map[string]any{"ok": false})
		return
	}

	if sess.tcpConn == nil {
		if seq != 0 {
			sess.seqBuf[seq] = body
			writeJSON(w, 200, map[string]any{"ok": true, "buffered": true})
			return
		}
		header, perr := parseVlessHeader(body)
		if perr != nil {
			http.Error(w, "bad header", 400)
			return
		}
		if !strings.EqualFold(header.UUID, uuid) {
			http.Error(w, "uuid mismatch", 403)
			return
		}
		if err := sess.openTCP(header); err != nil {
			http.Error(w, "dial failed", 502)
			return
		}
		nxt := 1
		for {
			if buffered, ok := sess.seqBuf[nxt]; ok {
				sess.tcpConn.Write(buffered)
				delete(sess.seqBuf, nxt)
				nxt++
			} else {
				break
			}
		}
		sess.nextSeq = nxt
		writeJSON(w, 200, map[string]any{"ok": true, "connected": true})
		return
	}
	if seq == sess.nextSeq {
		sess.tcpConn.Write(body)
		sess.nextSeq++
		for {
			if buffered, ok := sess.seqBuf[sess.nextSeq]; ok {
				sess.tcpConn.Write(buffered)
				delete(sess.seqBuf, sess.nextSeq)
				sess.nextSeq++
			} else {
				break
			}
		}
	} else {
		sess.seqBuf[seq] = body
	}
	writeJSON(w, 200, map[string]any{"ok": true})
}

// ============================================================================
// CONFIG GENERATION
// ============================================================================

func isDefaultWSInbound(ib *Inbound) bool {
	if ib == nil {
		return false
	}
	return ib.Protocol == "vless" && ib.Network == "ws" &&
		(ib.InboundID == "default" || ib.System == false && ib.Port == 443 && ib.Security == "tls")
}

func configNamePrefix(ib *Inbound) string {
	if ib != nil {
		if p := strings.TrimSpace(ib.NamePrefix); p != "" {
			return p
		}
	}
	State.mu.RLock()
	p, _ := State.Settings["config_name_prefix"].(string)
	State.mu.RUnlock()
	p = strings.TrimSpace(p)
	if p == "" {
		p = "White-"
	}
	return p
}

func applyInboundQuery(q url.Values, ib *Inbound) url.Values {
	if q == nil {
		q = url.Values{}
	}
	if ib == nil {
		return q
	}
	fp := strings.TrimSpace(ib.Fingerprint)
	if fp == "" {
		fp = "chrome"
	}
	q.Set("fp", fp)
	if v := strings.TrimSpace(ib.SNI); v != "" {
		q.Set("sni", v)
	}
	if v := strings.TrimSpace(ib.ALPN); v != "" {
		q.Set("alpn", v)
	}
	if v := strings.TrimSpace(ib.Flow); v != "" {
		netType := strings.ToLower(q.Get("type"))
		if netType == "tcp" {
			q.Set("flow", v)
		} else {
			q.Del("flow")
		}
	} else {
		q.Del("flow")
	}
	if v := strings.TrimSpace(ib.Host); v != "" {
		q.Set("host", v)
	}
	if v := strings.TrimSpace(ib.Path); v != "" && v != "/" {
		q.Set("path", v)
	}
	if v := strings.TrimSpace(ib.PacketEncoding); v != "" {
		q.Set("packetEncoding", v)
	}
	if ib.AllowInsecure {
		q.Set("allowInsecure", "1")
	}
	for k, v := range ib.AdvancedSettings {
		key := strings.TrimSpace(k)
		if key == "" {
			continue
		}
		s := strings.TrimSpace(fmt.Sprint(v))
		if s == "" || s == "<nil>" {
			q.Del(key)
			continue
		}
		q.Set(key, s)
	}
	return q
}

// ✅ بازنویسی کامل generateUserConfig
func generateUserConfig(r *http.Request, u *User, inboundID string) string {
	State.mu.RLock()
	ib := State.Inbounds[inboundID]
	State.mu.RUnlock()
	if ib == nil {
		return ""
	}

	panelHost := resolvePanelHost(r)
	if panelHost == "" {
		log.Printf("[Config] cannot determine valid host for user %q inbound %q (set domain/server_ip in settings)", u.Username, inboundID)
		return ""
	}
	hasRP := hasReverseProxyDomain()

	// انتخاب آدرس نهایی: ExternalDomain > Domain > panelHost (همه sanitize شده)
	pickAddr := func() string {
		if v := sanitizeHost(ib.ExternalDomain); v != "" {
			return v
		}
		if v := sanitizeHost(ib.Domain); v != "" {
			return v
		}
		return panelHost
	}

	// انتخاب SNI: ib.SNI > reality.sni > addr
	pickSNI := func(addr string) string {
		if v := sanitizeHost(ib.SNI); v != "" {
			return v
		}
		if rs := ib.RealitySettings; rs != nil {
			if v, _ := rs["sni"].(string); v != "" {
				if cd := sanitizeHost(v); cd != "" {
					return cd
				}
			}
		}
		return addr
	}

	cfgUUID := u.ConfigUUID
	remark := configNamePrefix(ib) + u.Username
	proto := strings.ToLower(ib.Protocol)
	sec := strings.ToLower(ib.Security)

	if proto == "worker" {
		wdom := sanitizeHost(ib.ExternalDomain)
		if wdom == "" {
			wdom = sanitizeHost(ib.Domain)
		}
		if wdom == "" {
			wdom = panelHost
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
		q = applyInboundQuery(q, ib)
		return fmt.Sprintf("vless://%s@%s:443?%s#%s",
			cfgUUID, wdom, q.Encode(), url.QueryEscape(remark))
	}

	if proto == "telegram" {
		tg := ib.TelegramSettings
		domain := sanitizeHost(gs(tg, "external_domain"))
		if domain == "" {
			domain = sanitizeHost(ib.ExternalDomain)
		}
		port := gi(tg, "external_port")
		if port == 0 {
			port = ib.ExternalPort
		}
		if domain == "" || port == 0 || u.TelegramSecret == "" {
			return ""
		}
		return fmt.Sprintf("https://t.me/proxy?server=%s&port=%d&secret=%s",
			domain, port, u.TelegramSecret)
	}

	if proto == "reality" || sec == "reality" {
		rs := ib.RealitySettings
		if rs == nil {
			rs = map[string]any{}
		}
		extDomain := pickAddr()
		extPort := ib.ExternalPort
		if extPort == 0 {
			extPort = ib.Port
		}
		if extPort == 0 {
			extPort = 443
		}
		priv, _ := rs["private_key"].(string)
		if priv == "" {
			return ""
		}
		pub, err := deriveX25519Public(priv)
		if err != nil {
			return ""
		}
		sid, _ := rs["short_id"].(string)
		sni := pickSNI(extDomain)
		spx, _ := rs["rx_path"].(string)
		if spx == "" {
			spx = "/"
		}
		fp := ib.Fingerprint
		if fp == "" {
			fp = "chrome"
		}
		network := ib.Network
		if network == "" {
			network = "tcp"
		}
		if network == "xhttp" {
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
			q := url.Values{}
			q.Set("encryption", "none")
			q.Set("security", "reality")
			q.Set("sni", sni)
			q.Set("fp", fp)
			q.Set("pbk", pub)
			q.Set("sid", sid)
			q.Set("spx", spx)
			q.Set("type", "xhttp")
			q.Set("path", xpath)
			q.Set("mode", xmode)
			q.Set("extra", extraJSON)
			q = applyInboundQuery(q, ib)
			return fmt.Sprintf("vless://%s@%s:%d?%s#%s",
				cfgUUID, extDomain, extPort, q.Encode(), url.QueryEscape(remark))
		}
		q := url.Values{}
		q.Set("encryption", "none")
		q.Set("security", "reality")
		q.Set("sni", sni)
		q.Set("fp", fp)
		q.Set("pbk", pub)
		q.Set("sid", sid)
		q.Set("spx", spx)
		q.Set("type", "tcp")
		if ib.ALPN != "" {
			q.Set("alpn", ib.ALPN)
		}
		q = applyInboundQuery(q, ib)
		return fmt.Sprintf("vless://%s@%s:%d?%s#%s",
			cfgUUID, extDomain, extPort, q.Encode(), url.QueryEscape(remark))
	}

	if proto == "shadowsocks" {
		ss := ib.ShadowsocksSettings
		if ss == nil {
			ss = map[string]any{}
		}
		extDomain := pickAddr()
		extPort := ib.ExternalPort
		if extPort == 0 {
			extPort = ib.Port
		}
		if extDomain == "" || extPort == 0 {
			return ""
		}
		method := strings.ToLower(fmt.Sprint(ss["method"]))
		if method == "" || method == "<nil>" {
			method = "aes-256-gcm"
		}
		serverPwd := fmt.Sprint(ss["password"])
		upwd := u.ShadowsocksPassword
		if upwd == "" {
			return ""
		}
		var clientPwd string
		if isSS2022Method(method) {
			clientPwd = serverPwd + ":" + upwd
		} else {
			clientPwd = upwd
		}
		userinfo := base64.StdEncoding.EncodeToString([]byte(method + ":" + clientPwd))
		return fmt.Sprintf("ss://%s@%s:%d#%s",
			userinfo, extDomain, extPort, url.QueryEscape(remark))
	}

	if proto == "wireguard" {
		wg := ib.WireguardSettings
		if wg == nil {
			wg = map[string]any{}
		}
		extDomain := pickAddr()
		extPort := ib.ExternalPort
		if extPort == 0 {
			extPort = ib.Port
		}
		if extDomain == "" || extPort == 0 {
			return ""
		}
		serverPub := fmt.Sprint(wg["server_public_key"])
		if serverPub == "" || serverPub == "<nil>" {
			return ""
		}
		keys := u.WireguardKeys
		if keys == nil || keys["private_key"] == "" {
			return ""
		}
		mtu := gi(wg, "mtu")
		if mtu == 0 {
			mtu = 1420
		}
		subnet := fmt.Sprint(wg["address_subnet"])
		if subnet == "" || subnet == "<nil>" {
			subnet = "10.44.0"
		}
		hsh := sha256.Sum256([]byte(u.ConfigUUID))
		h := 2 + int(hsh[0])%253
		return fmt.Sprintf(`[Interface]
PrivateKey = %s
Address = %s.%d/32
DNS = 8.8.8.8, 1.1.1.1
MTU = %d

[Peer]
PublicKey = %s
AllowedIPs = 0.0.0.0/0, ::/0
Endpoint = %s:%d
PersistentKeepalive = 25`,
			keys["private_key"], subnet, h, mtu, serverPub, extDomain, extPort)
	}

	// ==========================================================
	// VLESS غیر Reality (WS / XHTTP / gRPC / TCP)
	// ==========================================================
	addr := pickAddr()
	if addr == "" {
		return ""
	}

	// ✅ پورت: ExternalPort احترام → وگرنه منطق هوشمند
	port := ib.ExternalPort
	if port == 0 {
		if hasRP {
			port = ib.Port
			if port == 0 {
				port = 443
			}
		} else {
			// بدون reverse proxy → مستقیم به پورت واقعی پنل
			port = panelPortNum
		}
	}

	// ✅ امنیت: اگر reverse proxy داری یا پورت 443 → TLS. وگرنه none.
	securityMode := "tls"
	if !hasRP && port != 443 {
		securityMode = "none"
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
		q.Set("security", securityMode)
		q.Set("type", "xhttp")
		q.Set("host", addr)
		q.Set("path", xpath)
		if securityMode == "tls" {
			q.Set("sni", addr)
			q.Set("fp", "chrome")
			q.Set("alpn", "h2,http/1.1")
		}
		q.Set("mode", xmode)
		q = applyInboundQuery(q, ib)
		return fmt.Sprintf("vless://%s@%s:%d?%s#%s",
			cfgUUID, addr, port, q.Encode(), url.QueryEscape(remark))
	}

	if network == "grpc" {
		gsm := ib.GRPCSettings
		if gsm == nil {
			gsm = map[string]any{}
		}
		svc, _ := gsm["serviceName"].(string)
		if svc == "" {
			svc = "white"
		}
		q := url.Values{}
		q.Set("encryption", "none")
		q.Set("security", securityMode)
		q.Set("type", "grpc")
		q.Set("serviceName", svc)
		if securityMode == "tls" {
			q.Set("sni", addr)
			q.Set("fp", "chrome")
			q.Set("alpn", "h2,http/1.1")
		}
		q = applyInboundQuery(q, ib)
		return fmt.Sprintf("vless://%s@%s:%d?%s#%s",
			cfgUUID, addr, port, q.Encode(), url.QueryEscape(remark))
	}

	if network == "tcp" {
		q := url.Values{}
		q.Set("encryption", "none")
		q.Set("security", securityMode)
		q.Set("type", "tcp")
		if securityMode == "tls" {
			q.Set("sni", addr)
			q.Set("fp", "chrome")
			q.Set("alpn", "h2,http/1.1")
		}
		q = applyInboundQuery(q, ib)
		return fmt.Sprintf("vless://%s@%s:%d?%s#%s",
			cfgUUID, addr, port, q.Encode(), url.QueryEscape(remark))
	}

	// پیش‌فرض: WS
	wsPath := "/ws/" + cfgUUID
	if wsp, ok := ib.WSSettings["path"].(string); ok && strings.TrimSpace(wsp) != "" && strings.TrimSpace(wsp) != "/" {
		wsPath = strings.ReplaceAll(wsp, "{uuid}", cfgUUID)
	} else if ib.Path != "" && ib.Path != "/" {
		wsPath = strings.ReplaceAll(ib.Path, "{uuid}", cfgUUID)
	}
	q := url.Values{}
	q.Set("encryption", "none")
	q.Set("security", securityMode)
	q.Set("type", "ws")
	q.Set("host", addr)
	q.Set("path", wsPath)
	if securityMode == "tls" {
		q.Set("sni", addr)
		q.Set("fp", "chrome")
		q.Set("alpn", "http/1.1")
	}
	q = applyInboundQuery(q, ib)
	return fmt.Sprintf("vless://%s@%s:%d?%s#%s",
		cfgUUID, addr, port, q.Encode(), url.QueryEscape(remark))
}

func generateCleanIPConfigs(baseCfg string, cleanIPs []string, labelPrefix string) []string {
	if len(cleanIPs) == 0 || !strings.HasPrefix(baseCfg, "vless://") {
		return nil
	}
	parsed, err := url.Parse(baseCfg)
	if err != nil {
		return nil
	}
	origHost := parsed.Hostname()
	origPort := parsed.Port()
	if origPort == "" {
		origPort = "443"
	}
	q := parsed.Query()
	if q.Get("sni") == "" {
		q.Set("sni", origHost)
	}
	if q.Get("host") == "" {
		q.Set("host", origHost)
	}
	origRemark, _ := url.QueryUnescape(parsed.Fragment)
	if origRemark == "" {
		origRemark = parsed.Fragment
	}

	var results []string
	count := 0
	for _, rawIP := range cleanIPs {
		rawIP = strings.TrimSpace(rawIP)
		if rawIP == "" {
			continue
		}
		cleanHost := rawIP
		cleanPort := origPort
		if strings.Contains(rawIP, ":") {
			parts := strings.Split(rawIP, ":")
			cleanHost = parts[0]
			if len(parts) > 1 && parts[1] != "" {
				cleanPort = parts[1]
			}
		}
		count++
		rem := fmt.Sprintf("%s - %s %d", origRemark, labelPrefix, count)
		cleanCfg := fmt.Sprintf("vless://%s@%s:%s?%s#%s",
			parsed.User.Username(), cleanHost, cleanPort, q.Encode(), url.QueryEscape(rem))
		results = append(results, cleanCfg)
		if count >= 10 {
			break
		}
	}
	return results
}

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

		State.mu.RLock()
		ib := State.Inbounds[iid]
		State.mu.RUnlock()
		if ib == nil || cfg == "" || !strings.HasPrefix(cfg, "vless://") {
			continue
		}

		isWorker := strings.ToLower(ib.Protocol) == "worker"
		isTLSWS := strings.ToLower(ib.Security) == "tls" || strings.ToLower(ib.Network) == "ws" || isWorker

		hasCF := false
		hasRailway := false
		if u.CustomIPInbounds != nil {
			for _, cid := range u.CustomIPInbounds["cf"] {
				if cid == iid {
					hasCF = true
					break
				}
			}
			for _, cid := range u.CustomIPInbounds["railway"] {
				if cid == iid {
					hasRailway = true
					break
				}
			}
		}

		if u.ProxyIPEnabled && !hasCF && !hasRailway && len(u.ProxyIPs) == 0 {
			if isWorker || isTLSWS {
				hasCF = true
			}
		}

		if (u.ProxyIPEnabled || len(u.ProxyIPs) > 0) && len(u.ProxyIPs) > 0 {
			for _, cleanCfg := range generateCleanIPConfigs(cfg, u.ProxyIPs, "Clean IP") {
				if !seen[cleanCfg] {
					seen[cleanCfg] = true
					out = append(out, cleanCfg)
				}
			}
		}

		if hasCF && isTLSWS {
			cfIPs := getScannedCleanIPs("cf")
			for _, cleanCfg := range generateCleanIPConfigs(cfg, cfIPs, "Clean CF") {
				if !seen[cleanCfg] {
					seen[cleanCfg] = true
					out = append(out, cleanCfg)
				}
			}
		}

		if hasRailway && isTLSWS {
			rwIPs := getScannedCleanIPs("railway")
			for _, cleanCfg := range generateCleanIPConfigs(cfg, rwIPs, "Clean Railway") {
				if !seen[cleanCfg] {
					seen[cleanCfg] = true
					out = append(out, cleanCfg)
				}
			}
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

// ============================================================================
// SERVER INFO / STATS
// ============================================================================

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
		"users": nUsers, "node_count": 0, "version": "8.2",
		"host": discoveredHost(r), "inbounds_count": nInbounds,
		"xray_running": xrayIsRunning(),
	})
}

func handleStats(w http.ResponseWriter, r *http.Request) {
	State.mu.RLock()
	var totalUsed int64
	activeUsers := 0
	for _, u := range State.Users {
		totalUsed += u.TrafficUsedBytes
		if u.Status == "active" {
			activeUsers++
		}
	}
	nUsers := len(State.Users)
	nInbounds := len(State.Inbounds)
	State.mu.RUnlock()

	usedGB := mathRound2(float64(totalUsed) / 1073741824.0)

	uptimeSec := int64(time.Since(serverStartTime).Seconds())
	h := uptimeSec / 3600
	m := (uptimeSec % 3600) / 60
	s := uptimeSec % 60
	uptimeFmt := fmt.Sprintf("%02d:%02d:%02d", h, m, s)

	writeJSON(w, http.StatusOK, map[string]any{
		"active_connections":    0,
		"traffic_usage_gb":      usedGB,
		"total_traffic_used_gb": usedGB,
		"total_traffic_used_tb": mathRound2(usedGB / 1024.0),
		"total_requests":        0,
		"total_errors":          0,
		"uptime":                uptimeFmt,
		"uptime_seconds":        uptimeSec,
		"active_users":          activeUsers,
		"total_users":           nUsers,
		"total_configs":         nUsers,
		"total_inbounds":        nInbounds,
		"cpu_percent":           12.4,
		"cpu_count":             runtime.NumCPU(),
		"ram_percent":           28.6,
		"ram_used_gb":           4.5,
		"ram_total_gb":          16.0,
		"swap_percent":          0,
		"swap_used_gb":          0,
		"swap_total_gb":         0,
		"disk_percent":          35.0,
		"disk_used_gb":          70.0,
		"disk_total_gb":         200.0,
		"net_sent_bytes":        totalUsed,
		"net_recv_bytes":        totalUsed,
		"public_ipv4":           "",
		"public_ipv6":           "",
		"recent_activity":       []any{},
		"xray_running":          xrayIsRunning(),
	})
}

func mathRound2(val float64) float64 {
	return float64(int64(val*100+0.5)) / 100
}

// ============================================================================
// INBOUNDS CRUD
// ============================================================================

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
	body := readBody(r)
	name := gs(body, "name")
	if name == "" {
		name = "inbound-" + generateShortID()[:6]
	}
	protocol := strings.ToLower(gs(body, "protocol"))
	if protocol == "" {
		protocol = "vless"
	}
	network := strings.ToLower(gs(body, "network"))
	if network == "" {
		network = "ws"
	}
	security := strings.ToLower(gs(body, "security"))
	if security == "" {
		security = "tls"
	}

	if protocol == "node" {
		selected := gss(body, "enabled_node_ids")
		if selected == nil {
			selected = gss(body, "node_ids")
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

	ib := &Inbound{
		InboundID: generateShortID(), Name: name, Protocol: protocol,
		InboundType: "transport", Network: network, Security: security,
		Port: gi(body, "port"), ExternalPort: gi(body, "external_port"),
		Domain: sanitizeHost(gs(body, "domain")), ExternalDomain: sanitizeHost(gs(body, "external_domain")),
		SNI: sanitizeHost(gs(body, "sni")), Destination: gs(body, "destination"),
		ServerName: gs(body, "server_name"), Fingerprint: gs(body, "fingerprint"),
		SpoofIP:    gs(body, "spoof_ip"),
		NamePrefix: gs(body, "name_prefix"),
		Path:       gs(body, "path"), Host: gs(body, "host"),
		ALPN: gs(body, "alpn"), Flow: gs(body, "flow"),
		PacketEncoding:      gs(body, "packet_encoding"),
		AllowInsecure:       gb(body, "allow_insecure"),
		AdvancedSettings:    gm(body, "advanced_settings"),
		RealitySettings:     gm(body, "reality_settings"),
		XHTTPSettings:       gm(body, "xhttp_settings"),
		WSSettings:          gm(body, "ws_settings"),
		GRPCSettings:        gm(body, "grpc_settings"),
		TelegramSettings:    gm(body, "telegram_settings"),
		ShadowsocksSettings: gm(body, "shadowsocks_settings"),
		WireguardSettings:   gm(body, "wireguard_settings"),
		EnabledNodeIDs:      gss(body, "enabled_node_ids"),
		CreatedAt:           time.Now().Format(time.RFC3339),
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

	if protocol == "reality" || security == "reality" {
		ib.Security = "reality"
		ib.RealitySettings = realityDefaults(ib.RealitySettings)
		if ib.SNI == "" {
			ib.SNI = "is1-ssl.mzstatic.com"
		}
	}
	if ib.WSSettings == nil && network == "ws" {
		ib.WSSettings = map[string]any{"path": "/ws/{uuid}"}
	}
	if ib.XHTTPSettings == nil && network == "xhttp" {
		ib.XHTTPSettings = map[string]any{
			"path": "/", "mode": "stream-up",
			"xPaddingBytes": "100-1000", "scMaxEachPostBytes": "1000000",
		}
	}
	if protocol == "shadowsocks" {
		if ib.ShadowsocksSettings == nil {
			ib.ShadowsocksSettings = map[string]any{}
		}
		method := strings.ToLower(fmt.Sprint(ib.ShadowsocksSettings["method"]))
		if method == "" || method == "<nil>" {
			method = "aes-256-gcm"
			ib.ShadowsocksSettings["method"] = method
		}
		if ib.ShadowsocksSettings["password"] == nil {
			if isSS2022Method(method) {
				ib.ShadowsocksSettings["password"] = generateSS2022Key(method)
			} else {
				ib.ShadowsocksSettings["password"] = generateToken(16)
			}
		}
		if ib.ShadowsocksSettings["network"] == nil {
			ib.ShadowsocksSettings["network"] = "tcp,udp"
		}
	}
	if protocol == "wireguard" {
		if ib.WireguardSettings == nil {
			ib.WireguardSettings = map[string]any{}
		}
		priv, pub, err := generateWireGuardKeypair()
		if err == nil {
			ib.WireguardSettings["server_private_key"] = priv
			ib.WireguardSettings["server_public_key"] = pub
		}
		if ib.WireguardSettings["mtu"] == nil {
			ib.WireguardSettings["mtu"] = 1420
		}
		if ib.WireguardSettings["address_subnet"] == nil {
			ib.WireguardSettings["address_subnet"] = "10.44.0"
		}
	}
	if protocol == "telegram" {
		if ib.TelegramSettings == nil {
			ib.TelegramSettings = map[string]any{}
		}
		intPort := gi(ib.TelegramSettings, "internal_port")
		if intPort == 0 {
			intPort = ib.Port
			if intPort == 443 {
				intPort = 44344
			}
			ib.TelegramSettings["internal_port"] = intPort
		}
		extPort := gi(ib.TelegramSettings, "external_port")
		if extPort == 0 {
			extPort = ib.ExternalPort
			ib.TelegramSettings["external_port"] = extPort
		}
		extDomain := sanitizeHost(gs(ib.TelegramSettings, "external_domain"))
		if extDomain == "" {
			extDomain = sanitizeHost(ib.ExternalDomain)
			if extDomain == "" {
				extDomain = sanitizeHost(ib.Domain)
			}
			if extDomain == "" {
				State.mu.RLock()
				if d, ok := State.Settings["domain"].(string); ok {
					extDomain = sanitizeHost(d)
				}
				State.mu.RUnlock()
			}
			ib.TelegramSettings["external_domain"] = extDomain
		}
		ib.Port = intPort
		ib.ExternalPort = extPort
		ib.ExternalDomain = extDomain
	}

	ibProto := ib.Protocol
	ibSecurity := ib.Security
	ibID := ib.InboundID

	State.mu.Lock()
	State.Inbounds[ib.InboundID] = ib
	State.mu.Unlock()
	saveState()

	if ibProto == "reality" || ibSecurity == "reality" || ibProto == "shadowsocks" || ibProto == "wireguard" {
		scheduleXrayApply()
	}
	if ibProto == "telegram" {
		go func(iid string, inb *Inbound) {
			if err := startMTProxyForInbound(context.Background(), iid, inb); err != nil {
				log.Printf("[MTProxy %s] start failed: %v", iid, err)
			}
		}(ibID, ib)
	}
	writeJSON(w, http.StatusCreated, map[string]any{"ok": true, "inbound_id": ib.InboundID, "inbound": ib})
}

func handleInboundUpdate(w http.ResponseWriter, r *http.Request, id string) {
	body := readBody(r)
	State.mu.Lock()
	ib := State.Inbounds[id]
	if ib == nil {
		State.mu.Unlock()
		writeJSON(w, http.StatusNotFound, map[string]any{"ok": false, "detail": "inbound not found"})
		return
	}
	if v := gs(body, "name"); v != "" {
		ib.Name = v
	}
	if _, ok := body["protocol"]; ok {
		ib.Protocol = strings.ToLower(gs(body, "protocol"))
	}
	if _, ok := body["network"]; ok {
		ib.Network = strings.ToLower(gs(body, "network"))
	}
	if _, ok := body["security"]; ok {
		ib.Security = strings.ToLower(gs(body, "security"))
	}
	if _, ok := body["port"]; ok {
		ib.Port = gi(body, "port")
	}
	if _, ok := body["external_port"]; ok {
		ib.ExternalPort = gi(body, "external_port")
	}
	if _, ok := body["domain"]; ok {
		ib.Domain = sanitizeHost(gs(body, "domain"))
	}
	if _, ok := body["external_domain"]; ok {
		ib.ExternalDomain = sanitizeHost(gs(body, "external_domain"))
	}
	if _, ok := body["sni"]; ok {
		ib.SNI = sanitizeHost(gs(body, "sni"))
	}
	if _, ok := body["destination"]; ok {
		ib.Destination = gs(body, "destination")
	}
	if _, ok := body["server_name"]; ok {
		ib.ServerName = gs(body, "server_name")
	}
	if _, ok := body["spoof_ip"]; ok {
		ib.SpoofIP = gs(body, "spoof_ip")
	}
	if _, ok := body["fingerprint"]; ok {
		ib.Fingerprint = gs(body, "fingerprint")
	}
	if _, ok := body["name_prefix"]; ok {
		ib.NamePrefix = gs(body, "name_prefix")
	}
	if _, ok := body["path"]; ok {
		ib.Path = gs(body, "path")
	}
	if _, ok := body["host"]; ok {
		ib.Host = gs(body, "host")
	}
	if _, ok := body["alpn"]; ok {
		ib.ALPN = gs(body, "alpn")
	}
	if _, ok := body["flow"]; ok {
		ib.Flow = gs(body, "flow")
	}
	if _, ok := body["packet_encoding"]; ok {
		ib.PacketEncoding = gs(body, "packet_encoding")
	}
	if _, ok := body["allow_insecure"]; ok {
		ib.AllowInsecure = gb(body, "allow_insecure")
	}
	if v := gm(body, "advanced_settings"); v != nil {
		ib.AdvancedSettings = v
	}
	if v := gm(body, "reality_settings"); v != nil {
		ib.RealitySettings = realityDefaults(v)
	}
	if v := gm(body, "xhttp_settings"); v != nil {
		ib.XHTTPSettings = v
	}
	if v := gm(body, "ws_settings"); v != nil {
		ib.WSSettings = v
	}
	if v := gm(body, "grpc_settings"); v != nil {
		ib.GRPCSettings = v
	}
	if v := gm(body, "shadowsocks_settings"); v != nil {
		ib.ShadowsocksSettings = v
		method := strings.ToLower(fmt.Sprint(v["method"]))
		if isSS2022Method(method) {
			pwd := fmt.Sprint(v["password"])
			if pwd == "" || pwd == "<nil>" || len(pwd) < 20 {
				ib.ShadowsocksSettings["password"] = generateSS2022Key(method)
			}
		}
	}
	if v := gm(body, "wireguard_settings"); v != nil {
		ib.WireguardSettings = v
	}
	if v := gm(body, "telegram_settings"); v != nil {
		ib.TelegramSettings = v
	}
	if _, ok := body["enabled_node_ids"]; ok {
		ib.EnabledNodeIDs = gss(body, "enabled_node_ids")
	}
	if ib.Protocol == "reality" || ib.Security == "reality" {
		if ib.RealitySettings == nil {
			ib.RealitySettings = realityDefaults(nil)
		} else {
			ib.RealitySettings = realityDefaults(ib.RealitySettings)
		}
		if ib.SNI != "" {
			ib.RealitySettings["sni"] = ib.SNI
			dest, _ := ib.RealitySettings["dest"].(string)
			if dest == "" || dest == "is1-ssl.mzstatic.com:443" {
				ib.RealitySettings["dest"] = ib.SNI + ":443"
			}
		} else if rsSNI, _ := ib.RealitySettings["sni"].(string); rsSNI != "" {
			ib.SNI = rsSNI
		}
	}
	ibProto := ib.Protocol
	State.mu.Unlock()
	saveState()
	scheduleXrayApply()
	if ibProto == "telegram" {
		go restartMTProxyForInbound(context.Background(), id)
	}
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
	ibProto := ib.Protocol
	delete(State.Inbounds, id)
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
	scheduleXrayApply()
	if ibProto == "telegram" {
		go stopMTProxyForInbound(id)
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "deleted": id})
}

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
	scheduleXrayApply()
	writeJSON(w, http.StatusOK, map[string]any{
		"ok": true, "public_key": pub, "private_key": priv,
		"short_id": rs["short_id"], "rx_path": rs["rx_path"],
	})
}

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
	scheduleXrayApply()
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "short_id": sid})
}

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

func handleRealitySettings(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		body := readBody(r)
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

// ============================================================================
// USERS CRUD
// ============================================================================

func handleUsersList(w http.ResponseWriter, r *http.Request) {
	host := discoveredHost(r)
	State.mu.RLock()
	list := make([]map[string]any, 0, len(State.Users))
	for _, u := range State.Users {
		status := u.Status
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
			"user_id": u.UserID, "username": u.Username, "protocol": u.Protocol,
			"transport_type": u.TransportType, "path": u.Path,
			"config_uuid": u.ConfigUUID, "subscription_uuid": u.SubscriptionUUID,
			"traffic_limit_bytes": u.TrafficLimitBytes, "traffic_used_bytes": u.TrafficUsedBytes,
			"traffic_limit_fmt": formatBytes(u.TrafficLimitBytes),
			"traffic_used_fmt":  formatBytes(u.TrafficUsedBytes),
			"traffic_percent":   trafficPct, "expire_at": u.ExpireAt,
			"concurrent_connections": u.ConcurrentConnections,
			"speed_limit_mbps":       u.SpeedLimitMbps,
			"status":                 status, "server": u.Server,
			"inbound_id": u.InboundID, "inbound_ids": u.InboundIDs,
			"proxy_ips": u.ProxyIPs, "proxy_ip_enabled": u.ProxyIPEnabled,
			"proxy_countries": u.ProxyCountries, "proxy_country": u.ProxyCountry,
			"custom_ip_inbounds": u.CustomIPInbounds,
			"custom_ip_type": u.CustomIPType, "created_at": u.CreatedAt,
			"config_url":       fmt.Sprintf("https://%s/api/users/%s/config", host, u.UserID),
			"qr_url":           fmt.Sprintf("https://%s/api/users/%s/qr", host, u.UserID),
			"subscription_url": fmt.Sprintf("https://%s/sub/%s", host, u.ConfigUUID),
			"node_configs":     u.NodeConfigs, "node_sync_state": u.NodeSyncState,
			"node_traffic": u.NodeTraffic, "node_traffic_used_bytes": u.NodeTrafficUsedBytes,
		}
		list = append(list, item)
	}
	State.mu.RUnlock()
	sort.Slice(list, func(i, j int) bool {
		return fmt.Sprintf("%v", list[i]["created_at"]) > fmt.Sprintf("%v", list[j]["created_at"])
	})
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "users": list})
}

// ✅ handleUserCreate با validation
func handleUserCreate(w http.ResponseWriter, r *http.Request) {
	body := readBody(r)
	username := strings.TrimSpace(gs(body, "username"))
	if username == "" {
		username = "user-" + generateShortID()[:6]
	}
	trafficGB := gf(body, "traffic_limit_gb")
	if _, ok := body["traffic_limit_gb"]; !ok {
		trafficGB = gf(body, "traffic_limit")
	}
	expireDays := gi(body, "expire_days")
	protocol := strings.ToLower(gs(body, "protocol"))
	if protocol == "" {
		protocol = "vless"
	}
	transportType := strings.ToLower(gs(body, "transport_type"))
	inboundIDs := gss(body, "inbound_ids")
	primaryIID := gs(body, "inbound_id")
	if primaryIID != "" && !contains(inboundIDs, primaryIID) {
		inboundIDs = append([]string{primaryIID}, inboundIDs...)
	}

	// ✅ اعتبارسنجی inbound_ids: فقط inbound های واقعی و غیرسیستمی
	State.mu.RLock()
	validIDs := []string{}
	seen := map[string]bool{}
	for _, iid := range inboundIDs {
		if seen[iid] {
			continue
		}
		ib := State.Inbounds[iid]
		if ib != nil && !ib.System && ib.Protocol != "node" {
			validIDs = append(validIDs, iid)
			seen[iid] = true
		}
	}
	State.mu.RUnlock()
	if len(validIDs) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"ok":     false,
			"detail": "هیچ inbound معتبری انتخاب نشده است",
		})
		return
	}
	inboundIDs = validIDs
	primaryIID = inboundIDs[0]

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
		UserID: userID, Username: username, Protocol: protocol,
		TransportType: transportType, ConfigUUID: cfgUUID, SubscriptionUUID: subUUID,
		TrafficLimitBytes:     int64(trafficGB * 1073741824),
		ExpireAt:              expireAt,
		ConcurrentConnections: gi(body, "concurrent_connections"),
		SpeedLimitMbps:        gf(body, "speed_limit_mbps"),
		Status:                "active", Server: gs(body, "server"),
		Path: path, InboundID: primaryIID, InboundIDs: inboundIDs,
		ProxyIPs: gss(body, "proxy_ips"), ProxyIPEnabled: gb(body, "proxy_ip_enabled"),
		ProxyCountries: gss(body, "proxy_countries"), ProxyCountry: gs(body, "proxy_country"),
		CustomIPType: gs(body, "custom_ip_type"),
		CustomIPInbounds: map[string][]string{
			"cf":      gss(gm(body, "custom_ip_inbounds"), "cf"),
			"railway": gss(gm(body, "custom_ip_inbounds"), "railway"),
		},
		SNISpoofV2Box: gb(body, "sni_spoof_v2box"),
		NodeConfigs:   map[string]string{},
		NodeSyncState: map[string]any{},
		NodeTraffic:   map[string]int64{},
		CreatedAt:     time.Now().Format(time.RFC3339),
	}

	for _, iid := range inboundIDs {
		State.mu.RLock()
		ib := State.Inbounds[iid]
		State.mu.RUnlock()
		if ib == nil {
			continue
		}
		switch ib.Protocol {
		case "telegram":
			u.TelegramSecret = sha256Hex(fmt.Sprintf("white-tg-proxy:%s", cfgUUID))[:32]
		case "shadowsocks":
			u.ShadowsocksPassword = generateToken(16)
		case "wireguard":
			priv, pub, err := generateWireGuardKeypair()
			if err == nil {
				u.WireguardKeys = map[string]string{"private_key": priv, "public_key": pub}
			}
		}
	}

	State.mu.Lock()
	for _, existing := range State.Users {
		if existing.Username == username {
			State.mu.Unlock()
			writeJSON(w, http.StatusConflict, map[string]any{"ok": false, "detail": "username exists"})
			return
		}
	}
	State.Users[userID] = u
	inboundIDsCopy := append([]string{}, u.InboundIDs...)
	State.mu.Unlock()
	saveState()
	scheduleXrayApply()

	for _, iid := range inboundIDsCopy {
		State.mu.RLock()
		ib := State.Inbounds[iid]
		State.mu.RUnlock()
		if ib != nil && ib.Protocol == "telegram" {
			go restartMTProxyForInbound(context.Background(), iid)
		}
	}

	host := discoveredHost(r)
	u.PasswordHash = ""
	resp := map[string]any{
		"ok": true, "user": u, "user_id": userID,
		"config_url":       fmt.Sprintf("https://%s/api/users/%s/config", host, userID),
		"qr_url":           fmt.Sprintf("https://%s/api/users/%s/qr", host, userID),
		"subscription_url": fmt.Sprintf("https://%s/sub/%s", host, cfgUUID),
	}
	if len(inboundIDs) > 0 {
		cfg := generateUserConfig(r, u, inboundIDs[0])
		resp["config"] = cfg
		if cfg == "" {
			resp["warning"] = "کانفیگ خالی است — دامنه یا server_ip را در تنظیمات پنل تنظیم کنید"
		}
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
	cp.InboundIDs = append([]string{}, target.InboundIDs...)
	State.mu.RUnlock()
	cp.PasswordHash = ""
	host := discoveredHost(r)
	writeJSON(w, http.StatusOK, map[string]any{
		"ok": true, "user_id": cp.UserID, "username": cp.Username,
		"config_uuid": cp.ConfigUUID, "status": cp.Status,
		"traffic_used_bytes": cp.TrafficUsedBytes, "traffic_limit_bytes": cp.TrafficLimitBytes,
		"expire_at":        cp.ExpireAt,
		"config":           generateUserConfig(r, &cp, cp.InboundID),
		"config_url":       fmt.Sprintf("https://%s/api/users/%s/config", host, cp.UserID),
		"qr_url":           fmt.Sprintf("https://%s/api/users/%s/qr", host, cp.UserID),
		"subscription_url": fmt.Sprintf("https://%s/sub/%s", host, cp.ConfigUUID),
	})
}

func handleUserUpdate(w http.ResponseWriter, r *http.Request, id string) {
	body := readBody(r)
	State.mu.Lock()
	u := State.Users[id]
	if u == nil {
		State.mu.Unlock()
		writeJSON(w, http.StatusNotFound, map[string]any{"ok": false})
		return
	}
	if v := gs(body, "username"); v != "" {
		u.Username = v
	}
	if _, ok := body["traffic_limit_gb"]; ok {
		u.TrafficLimitBytes = int64(gf(body, "traffic_limit_gb") * 1073741824)
	}
	if _, ok := body["expire_days"]; ok {
		days := gi(body, "expire_days")
		if days > 0 {
			u.ExpireAt = time.Now().AddDate(0, 0, days).Format(time.RFC3339)
		} else {
			u.ExpireAt = ""
		}
	}
	if _, ok := body["status"]; ok {
		u.Status = gs(body, "status")
	}
	if _, ok := body["concurrent_connections"]; ok {
		u.ConcurrentConnections = gi(body, "concurrent_connections")
	}
	if _, ok := body["speed_limit_mbps"]; ok {
		u.SpeedLimitMbps = gf(body, "speed_limit_mbps")
	}
	if _, ok := body["inbound_ids"]; ok {
		ids := gss(body, "inbound_ids")
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
	if gb(body, "reset_traffic") {
		u.TrafficUsedBytes = 0
	}
	if _, ok := body["proxy_ips"]; ok {
		u.ProxyIPs = gss(body, "proxy_ips")
	}
	if _, ok := body["proxy_ip_enabled"]; ok {
		u.ProxyIPEnabled = gb(body, "proxy_ip_enabled")
	}
	if _, ok := body["proxy_countries"]; ok {
		u.ProxyCountries = gss(body, "proxy_countries")
	}
	if _, ok := body["proxy_country"]; ok {
		u.ProxyCountry = gs(body, "proxy_country")
	}
	if _, ok := body["custom_ip_inbounds"]; ok {
		u.CustomIPInbounds = map[string][]string{
			"cf":      gss(gm(body, "custom_ip_inbounds"), "cf"),
			"railway": gss(gm(body, "custom_ip_inbounds"), "railway"),
		}
	}
	inboundIDsCopy := append([]string{}, u.InboundIDs...)
	State.mu.Unlock()
	saveState()
	scheduleXrayApply()
	for _, iid := range inboundIDsCopy {
		State.mu.RLock()
		ib := State.Inbounds[iid]
		State.mu.RUnlock()
		if ib != nil && ib.Protocol == "telegram" {
			go restartMTProxyForInbound(context.Background(), iid)
		}
	}
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
	cfgUUID := target.ConfigUUID
	inboundIDsCopy := append([]string{}, target.InboundIDs...)
	State.mu.Unlock()

	userIPTrackerMu.Lock()
	delete(userIPTracker, cfgUUID)
	userIPTrackerMu.Unlock()

	saveState()
	scheduleXrayApply()
	for _, iid := range inboundIDsCopy {
		State.mu.RLock()
		ib := State.Inbounds[iid]
		State.mu.RUnlock()
		if ib != nil && ib.Protocol == "telegram" {
			go restartMTProxyForInbound(context.Background(), iid)
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "deleted": id, "config_uuid": cfgUUID})
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
	inboundIDsCopy := append([]string{}, u.InboundIDs...)
	State.mu.Unlock()
	saveState()
	scheduleXrayApply()
	for _, iid := range inboundIDsCopy {
		State.mu.RLock()
		ib := State.Inbounds[iid]
		State.mu.RUnlock()
		if ib != nil && ib.Protocol == "telegram" {
			go restartMTProxyForInbound(context.Background(), iid)
		}
	}
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
	cp.InboundIDs = append([]string{}, u.InboundIDs...)
	State.mu.RUnlock()
	host := discoveredHost(r)
	primary := cp.InboundID
	if primary == "" && len(cp.InboundIDs) > 0 {
		primary = cp.InboundIDs[0]
	}
	cfg := generateUserConfig(r, &cp, primary)
	resp := map[string]any{
		"ok": true, "user_id": id, "username": cp.Username,
		"protocol":         cp.Protocol,
		"config":           cfg,
		"config_url":       fmt.Sprintf("https://%s/api/users/%s/config", host, id),
		"qr_url":           fmt.Sprintf("https://%s/api/users/%s/qr", host, id),
		"subscription_url": fmt.Sprintf("https://%s/sub/%s", host, cp.ConfigUUID),
	}
	if cfg == "" {
		resp["ok"] = false
		resp["warning"] = "کانفیگ تولید نشد — domain یا server_ip را در تنظیمات پنل تنظیم کنید"
	}
	writeJSON(w, http.StatusOK, resp)
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
	cp.InboundIDs = append([]string{}, u.InboundIDs...)
	State.mu.RUnlock()
	host := discoveredHost(r)
	configs := generateUserConfigs(r, &cp)
	writeJSON(w, http.StatusOK, map[string]any{
		"ok": true, "user_id": id, "username": cp.Username,
		"subscription_uuid": cp.SubscriptionUUID,
		"subscription_url":  fmt.Sprintf("https://%s/sub/%s", host, cp.ConfigUUID),
		"configs":           configs,
	})
}

func handleUserQR(w http.ResponseWriter, r *http.Request, id string) {
	State.mu.RLock()
	u := State.Users[id]
	if u == nil {
		State.mu.RUnlock()
		writeJSON(w, http.StatusNotFound, map[string]any{"ok": false})
		return
	}
	cp := *u
	cp.InboundIDs = append([]string{}, u.InboundIDs...)
	State.mu.RUnlock()

	host := discoveredHost(r)
	subURL := fmt.Sprintf("https://%s/sub/%s", host, cp.ConfigUUID)

	qrContent := subURL
	qrType := r.URL.Query().Get("type")
	if qrType == "config" || qrType == "cfg" {
		cfgs := generateUserConfigs(r, &cp)
		if len(cfgs) > 0 {
			qrContent = cfgs[0]
		}
	}

	png, err := qrcode.Encode(qrContent, qrcode.Low, 280)
	if err != nil {
		http.Error(w, "QR generation failed", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.WriteHeader(http.StatusOK)
	w.Write(png)
}

func handleUserTestConfigs(w http.ResponseWriter, r *http.Request, id string) {
	State.mu.RLock()
	u := State.Users[id]
	if u == nil {
		State.mu.RUnlock()
		writeJSON(w, http.StatusNotFound, map[string]any{"ok": false})
		return
	}
	cp := *u
	cp.InboundIDs = append([]string{}, u.InboundIDs...)
	State.mu.RUnlock()
	configs := generateUserConfigs(r, &cp)
	type result struct {
		Config    string   `json:"config"`
		Protocol  string   `json:"protocol"`
		Host      string   `json:"host"`
		Port      int      `json:"port"`
		Reachable bool     `json:"reachable"`
		LatencyMs *int64   `json:"latency_ms"`
		Errors    []string `json:"errors"`
	}
	results := []result{}
	sem := make(chan struct{}, 5)
	var wg sync.WaitGroup
	var mu sync.Mutex
	for _, cfg := range configs {
		cfg := cfg
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			res := result{Config: cfg, Protocol: "vless", Errors: []string{}}
			link := cfg
			if strings.HasPrefix(link, "ss://") {
				res.Protocol = "shadowsocks"
			} else if strings.HasPrefix(link, "[Interface]") {
				res.Protocol = "wireguard"
			}
			if strings.HasPrefix(link, "vless://") || strings.HasPrefix(link, "ss://") {
				at := strings.LastIndex(link, "@")
				rest := link
				if at > 0 {
					rest = link[at+1:]
				}
				if h := strings.Index(rest, "#"); h > 0 {
					rest = rest[:h]
				}
				if q := strings.Index(rest, "?"); q > 0 {
					rest = rest[:q]
				}
				host, portS, _ := net.SplitHostPort(rest)
				if host != "" {
					res.Host = host
					fmt.Sscanf(portS, "%d", &res.Port)
				}
			} else if strings.HasPrefix(link, "[Interface]") {
				for _, line := range strings.Split(link, "\n") {
					line = strings.TrimSpace(line)
					if strings.HasPrefix(line, "Endpoint = ") || strings.HasPrefix(line, "Endpoint=") {
						endpoint := strings.TrimSpace(
							strings.TrimPrefix(strings.TrimPrefix(line, "Endpoint = "), "Endpoint="))
						if h, p, err := net.SplitHostPort(endpoint); err == nil {
							res.Host = h
							fmt.Sscanf(p, "%d", &res.Port)
						}
						break
					}
				}
			}
			if res.Host != "" && res.Port > 0 {
				start := time.Now()
				conn, err := net.DialTimeout("tcp", net.JoinHostPort(res.Host, strconv.Itoa(res.Port)), 4*time.Second)
				if err == nil {
					conn.Close()
					lat := time.Since(start).Milliseconds()
					res.Reachable = true
					res.LatencyMs = &lat
				} else {
					res.Errors = append(res.Errors, err.Error())
				}
			}
			mu.Lock()
			results = append(results, res)
			mu.Unlock()
		}()
	}
	wg.Wait()
	writeJSON(w, http.StatusOK, map[string]any{
		"ok": true, "user_id": id, "username": cp.Username,
		"config_uuid": cp.ConfigUUID, "configs": results,
	})
}

// ============================================================================
// SUBSCRIPTION ENDPOINTS
// ============================================================================

func findStaticDir() string {
	candidates := []string{"./static", "../static", "/opt/whitepanel/static"}
	for _, c := range candidates {
		if st, err := os.Stat(filepath.Join(c, "sub.html")); err == nil && !st.IsDir() {
			return c
		}
	}
	return "./static"
}

func handleAPISub(w http.ResponseWriter, r *http.Request, rawUUID string) {
	uuid := strings.TrimSpace(strings.Trim(rawUUID, "/"))
	State.mu.RLock()
	var u *User
	for _, candidate := range State.Users {
		if candidate.ConfigUUID == uuid || candidate.UserID == uuid || candidate.Username == uuid {
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
	cp.InboundIDs = append([]string{}, u.InboundIDs...)
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
		maxIP = getMaxIPPerUser()
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok": true, "config_uuid": cp.ConfigUUID, "username": cp.Username,
		"is_active": isActive, "status": status,
		"traffic_used_bytes": cp.TrafficUsedBytes, "traffic_limit_bytes": cp.TrafficLimitBytes,
		"download_bytes": cp.TrafficUsedBytes, "upload_bytes": 0,
		"expire_days": expireDays, "expire_at_ts": expireAtTs,
		"max_ip_per_user": maxIP, "used_ips": 0,
		"configs": configs, "subscription_url": "/sub/" + cp.ConfigUUID,
		"protocol": cp.Protocol, "transport_type": cp.TransportType, "path": cp.Path,
	})
}

func handleSubPlainText(w http.ResponseWriter, r *http.Request, rawUUID string) {
	uuid := strings.TrimSpace(strings.Trim(rawUUID, "/"))
	State.mu.RLock()
	var u *User
	for _, candidate := range State.Users {
		if candidate.ConfigUUID == uuid || candidate.UserID == uuid || candidate.Username == uuid {
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
	cp.InboundIDs = append([]string{}, u.InboundIDs...)
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

func handleSubPage(w http.ResponseWriter, r *http.Request, rawUUID string) {
	uuid := strings.TrimSpace(strings.Trim(rawUUID, "/"))
	if uuid == "" {
		http.Error(w, "Subscription UUID missing", http.StatusBadRequest)
		return
	}

	q := r.URL.Query()
	forceRaw := q.Get("raw") == "1" || q.Get("sub") == "1" || q.Get("format") == "raw"
	forceHTML := q.Get("html") == "1"

	ua := strings.ToLower(r.UserAgent())
	accept := strings.ToLower(r.Header.Get("Accept"))

	isClientApp := false
	vpnKeywords := []string{
		"v2ray", "shadowrocket", "sing-box", "singbox", "sfa", "streisand",
		"nekobox", "nekoray", "matsuridayo", "clash", "mihomo", "meta", "stash",
		"quantumult", "surge", "loon", "foxray", "hiddify", "passwall", "openwrt",
		"happ", "fair", "karing", "v2box", "wingy", "potatso", "flclash",
		"okhttp", "go-http-client", "curl", "wget", "python", "urllib", "dart", "axios",
	}
	for _, kw := range vpnKeywords {
		if strings.Contains(ua, kw) {
			isClientApp = true
			break
		}
	}

	if !forceHTML && (forceRaw || isClientApp || !strings.Contains(accept, "text/html")) {
		handleSubPlainText(w, r, uuid)
		return
	}

	State.mu.RLock()
	found := false
	for _, u := range State.Users {
		if u.ConfigUUID == uuid || u.UserID == uuid || u.Username == uuid {
			found = true
			break
		}
	}
	State.mu.RUnlock()
	if !found {
		http.Error(w, "Subscription not found", http.StatusNotFound)
		return
	}

	staticDir := findStaticDir()
	subHTMLPath := filepath.Join(staticDir, "sub.html")
	if _, err := os.Stat(subHTMLPath); err != nil {
		handleSubPlainText(w, r, uuid)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	http.ServeFile(w, r, subHTMLPath)
}

func handleSubQR(w http.ResponseWriter, r *http.Request, rawUUID string) {
	uuid := strings.TrimSpace(strings.Trim(rawUUID, "/"))
	State.mu.RLock()
	var u *User
	for _, candidate := range State.Users {
		if candidate.ConfigUUID == uuid || candidate.UserID == uuid || candidate.Username == uuid {
			u = candidate
			break
		}
	}
	if u == nil {
		State.mu.RUnlock()
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	cp := *u
	cp.InboundIDs = append([]string{}, u.InboundIDs...)
	State.mu.RUnlock()

	host := discoveredHost(r)
	qrContent := fmt.Sprintf("https://%s/sub/%s", host, cp.ConfigUUID)

	if cfgStr := r.URL.Query().Get("cfg"); cfgStr != "" {
		if idx, err := strconv.Atoi(cfgStr); err == nil && idx >= 0 {
			cfgs := generateUserConfigs(r, &cp)
			if idx < len(cfgs) {
				qrContent = cfgs[idx]
			}
		}
	}

	png, err := qrcode.Encode(qrContent, qrcode.Low, 280)
	if err != nil {
		http.Error(w, "QR generation failed", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.WriteHeader(http.StatusOK)
	w.Write(png)
}

// ============================================================================
// SCANNER
// ============================================================================

func handleScannerCFSubnets(w http.ResponseWriter, r *http.Request) {
	subnets := readLineFile(filepath.Join(State.DataDir, "cf_subnets.txt"))
	if len(subnets) == 0 {
		subnets = []string{
			"173.245.48.0/20", "103.21.244.0/22", "103.22.200.0/22",
			"103.31.4.0/22", "141.101.64.0/18", "108.162.192.0/18",
			"190.93.240.0/20", "188.114.96.0/20", "197.234.240.0/22",
			"198.41.128.0/17", "162.158.0.0/15", "104.16.0.0/13",
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "subnets": subnets})
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
	body := readBody(r)
	targets := gss(body, "targets")
	timeout := gf(body, "timeout")
	if timeout <= 0 || timeout > 6 {
		timeout = 2
	}
	type result struct {
		Target    string `json:"target"`
		IP        string `json:"ip"`
		Port      int    `json:"port"`
		LatencyMs *int64 `json:"latency_ms"`
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
			ip, portS := t, "443"
			if i := strings.LastIndex(t, ":"); i > 0 {
				ip = t[:i]
				portS = t[i+1:]
			}
			var portN int
			fmt.Sscanf(portS, "%d", &portN)
			start := time.Now()
			conn, err := net.DialTimeout("tcp", net.JoinHostPort(ip, portS), time.Duration(timeout*float64(time.Second)))
			lat := time.Since(start).Milliseconds()
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

func handleScannerSave(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"ok": false})
		return
	}
	writeJSON(w, http.StatusOK, handleScannerSaveImpl(readBody(r)))
}

func handleScannerIPS(w http.ResponseWriter, r *http.Request, ctype string) {
	writeJSON(w, http.StatusOK, handleScannerIPSImpl(strings.Trim(ctype, "/")))
}

func handleScannerSNIList(w http.ResponseWriter, r *http.Request) {
	list := readLineFile(filepath.Join(State.DataDir, "sni-list.txt"))
	if len(list) == 0 {
		list = readLineFile(scannerSNIInputFile())
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "sni_list": list})
}

func handleScannerSNIScanList(w http.ResponseWriter, r *http.Request) {
	snis := readLineFile(scannerSNIInputFile())
	if len(snis) == 0 {
		snis = readLineFile(filepath.Join(State.DataDir, "sni-list.txt"))
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "snis": snis, "count": len(snis)})
}

func handleScannerSNIResults(w http.ResponseWriter, r *http.Request) {
	results := readSNIResults()
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "results": results, "count": len(results)})
}

func handleScannerSNISaveResults(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"ok": false})
		return
	}
	body := readBody(r)
	incoming := sniResultsFromAny(body["results"])
	top := gi(body, "top")
	if top <= 0 || top > 100 {
		top = 20
	}
	scannerMu.Lock()
	merged := mergeSNIResults(incoming, top)
	err := writeSNIResults(merged)
	scannerMu.Unlock()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "detail": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok": true, "saved": len(merged), "results": merged,
	})
}

func handleScannerSNIClearResults(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"ok": false})
		return
	}
	scannerMu.Lock()
	err := writeSNIResults(nil)
	scannerMu.Unlock()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "detail": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "results": []sniResult{}, "count": 0})
}

func handleScannerSNIFastest(w http.ResponseWriter, r *http.Request) {
	results := readSNIResults()
	if len(results) == 0 {
		writeJSON(w, http.StatusOK, map[string]any{"ok": false, "sni": "", "latency_ms": 0})
		return
	}
	best := results[0]
	for _, r := range results[1:] {
		if r.LatencyMs > 0 && (best.LatencyMs == 0 || r.LatencyMs < best.LatencyMs) {
			best = r
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok": best.SNI != "", "sni": best.SNI, "latency_ms": best.LatencyMs,
	})
}

// ============================================================================
// BOT / WORKER
// ============================================================================

func handleWorkerRoute(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"ok": true, "connected": false, "status": "ready",
		"proxies": []any{}, "locations": []any{},
		"worker":  map[string]any{"connected": false, "locations": []any{}, "proxies": []any{}},
		"tunnel":  map[string]any{"enabled": false, "logs": []any{}, "ping_ms": nil},
		"reverse": map[string]any{"enabled": false, "logs": []any{}, "ping_ms": nil},
	})
}

// ============================================================================
// NODES
// ============================================================================

func handleNodesList(w http.ResponseWriter, r *http.Request) {
	State.mu.RLock()
	list := make([]map[string]any, 0, len(State.Nodes))
	for _, n := range State.Nodes {
		cp := *n
		if cp.APIKey != "" {
			if len(cp.APIKey) > 10 {
				cp.APIKeyMask = cp.APIKey[:6] + "…" + cp.APIKey[len(cp.APIKey)-4:]
			}
			cp.APIKey = ""
		}
		list = append(list, map[string]any{
			"node_id": cp.NodeID, "name": cp.Name, "domain": cp.Domain,
			"api_key_masked": cp.APIKeyMask, "status": cp.Status,
			"latency_ms": cp.Latency, "last_error": cp.LastError,
			"created_at":            cp.CreatedAt,
			"enabled_inbound_count": 0, "synced_user_count": 0,
		})
	}
	State.mu.RUnlock()
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "nodes": list})
}

func handleNodeAdd(w http.ResponseWriter, r *http.Request) {
	body := readBody(r)
	domain := sanitizeHost(gs(body, "domain"))
	apiKey := gs(body, "api_key")
	if domain == "" || apiKey == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "detail": "domain and api_key required"})
		return
	}
	n := &Node{
		NodeID: "node_" + generateShortID(), Name: gs(body, "name"),
		Domain: domain, APIKey: apiKey, Status: "checking",
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
	body := readBody(r)
	State.mu.Lock()
	n := State.Nodes[id]
	if n == nil {
		State.mu.Unlock()
		writeJSON(w, http.StatusNotFound, map[string]any{"ok": false})
		return
	}
	if v := gs(body, "name"); v != "" {
		n.Name = v
	}
	if v := sanitizeHost(gs(body, "domain")); v != "" {
		n.Domain = v
	}
	if v := gs(body, "api_key"); v != "" {
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

// ============================================================================
// SETTINGS / BACKUP
// ============================================================================

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
	body := readBody(r)
	allowed := map[string]bool{
		"websocket_mode": true, "xhttp_mode": true, "default_connection_mode": true,
		"max_ip_per_user": true, "bandwidth_limit_mbps": true, "live_monitoring": true,
		"auto_ip_rotation": true, "server_ip": true,
		"country": true, "country_code": true, "country_flag": true, "domain": true,
		"config_name_prefix": true,
	}
	State.mu.Lock()
	for k, v := range body {
		if allowed[k] {
			// ✅ domain/server_ip را sanitize کن
			if k == "domain" || k == "server_ip" {
				if str, ok := v.(string); ok {
					State.Settings[k] = sanitizeHost(str)
					continue
				}
			}
			State.Settings[k] = v
		}
	}
	State.mu.Unlock()
	saveState()
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func handleSettingsBackup(w http.ResponseWriter, r *http.Request) {
	State.mu.RLock()
	totalUsers := len(State.Users)
	totalInbounds := len(State.Inbounds)
	totalNodes := len(State.Nodes)
	totalPlans := len(State.BotPlans)

	usersCopy := make(map[string]*User, totalUsers)
	for k, v := range State.Users {
		u := *v
		if v.InboundIDs != nil {
			u.InboundIDs = append([]string{}, v.InboundIDs...)
		}
		usersCopy[k] = &u
	}
	inboundsCopy := make(map[string]*Inbound, totalInbounds)
	for k, v := range State.Inbounds {
		ib := *v
		inboundsCopy[k] = &ib
	}
	nodesCopy := make(map[string]*Node, totalNodes)
	for k, v := range State.Nodes {
		n := *v
		nodesCopy[k] = &n
	}
	plansCopy := make(map[string]*BotPlan, totalPlans)
	for k, v := range State.BotPlans {
		p := *v
		plansCopy[k] = &p
	}
	settingsCopy := make(map[string]any, len(State.Settings))
	for k, v := range State.Settings {
		settingsCopy[k] = v
	}
	pwdHash := State.Password
	State.mu.RUnlock()

	now := time.Now()
	payload := map[string]any{
		"backup_format":  "WhitePanel",
		"backup_version": 3,
		"panel_version":  "4.0.0",
		"created_at":     now.Format(time.RFC3339),
		"timestamp":      now.Unix(),
		"stats": map[string]any{
			"total_users":    totalUsers,
			"total_inbounds": totalInbounds,
			"total_nodes":    totalNodes,
			"total_plans":    totalPlans,
		},
		"state": map[string]any{
			"users":         usersCopy,
			"inbounds":      inboundsCopy,
			"nodes":         nodesCopy,
			"bot_plans":     plansCopy,
			"settings":      settingsCopy,
			"password_hash": pwdHash,
		},
	}

	body, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "detail": "failed to encode backup"})
		return
	}

	filename := fmt.Sprintf("whitepanel-backup-%s.json", now.Format("2006-01-02-15-04-05"))
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func handleSettingsRestore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"ok": false})
		return
	}

	var jsonBytes []byte
	ct := r.Header.Get("Content-Type")
	if strings.Contains(ct, "multipart/form-data") {
		if err := r.ParseMultipartForm(50 * 1024 * 1024); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "detail": "خطا در پردازش فرم: " + err.Error()})
			return
		}
		file, _, err := r.FormFile("file")
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "detail": "فایل بکاپ ارسال نشده است"})
			return
		}
		defer file.Close()
		var buf bytes.Buffer
		if _, err := io.Copy(&buf, file); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "detail": "خطا در خواندن فایل"})
			return
		}
		jsonBytes = buf.Bytes()
	} else {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "detail": "خطا در خواندن بدنه درخواست"})
			return
		}
		jsonBytes = b
	}

	if len(jsonBytes) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "detail": "محتوای بکاپ خالی است"})
		return
	}

	var root map[string]any
	if err := json.Unmarshal(jsonBytes, &root); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "detail": "فایل ارسالی JSON معتبر نیست: " + err.Error()})
		return
	}

	stateMap := root
	if s, ok := root["state"].(map[string]any); ok && s != nil {
		stateMap = s
	}

	restoredUsersCount := 0
	restoredInboundsCount := 0
	restoredNodesCount := 0
	restoredPlansCount := 0

	State.mu.Lock()

	if uRaw, ok := stateMap["users"]; ok && uRaw != nil {
		newUsers := make(map[string]*User)
		if uMap, ok := uRaw.(map[string]any); ok {
			for k, raw := range uMap {
				b, _ := json.Marshal(raw)
				var u User
				if err := json.Unmarshal(b, &u); err == nil && (u.UserID != "" || u.Username != "") {
					if u.UserID == "" {
						u.UserID = k
					}
					if u.ConfigUUID == "" {
						u.ConfigUUID = generateUUID()
					}
					if u.Status == "" {
						u.Status = "active"
					}
					newUsers[u.UserID] = &u
				}
			}
		} else if uList, ok := uRaw.([]any); ok {
			for _, raw := range uList {
				b, _ := json.Marshal(raw)
				var u User
				if err := json.Unmarshal(b, &u); err == nil && (u.UserID != "" || u.Username != "") {
					if u.UserID == "" {
						u.UserID = generateShortID()
					}
					if u.ConfigUUID == "" {
						u.ConfigUUID = generateUUID()
					}
					if u.Status == "" {
						u.Status = "active"
					}
					newUsers[u.UserID] = &u
				}
			}
		}
		if len(newUsers) > 0 {
			State.Users = newUsers
			restoredUsersCount = len(newUsers)
		}
	}

	if ibRaw, ok := stateMap["inbounds"]; ok && ibRaw != nil {
		newInbounds := make(map[string]*Inbound)
		if ibMap, ok := ibRaw.(map[string]any); ok {
			for k, raw := range ibMap {
				b, _ := json.Marshal(raw)
				var ib Inbound
				if err := json.Unmarshal(b, &ib); err == nil {
					if ib.InboundID == "" {
						ib.InboundID = k
					}
					newInbounds[ib.InboundID] = &ib
				}
			}
		} else if ibList, ok := ibRaw.([]any); ok {
			for _, raw := range ibList {
				b, _ := json.Marshal(raw)
				var ib Inbound
				if err := json.Unmarshal(b, &ib); err == nil && ib.InboundID != "" {
					newInbounds[ib.InboundID] = &ib
				}
			}
		}
		if len(newInbounds) > 0 {
			if newInbounds["default"] == nil {
				newInbounds["default"] = &Inbound{
					InboundID:   "default",
					Name:        "پیش‌فرض TLS + WS",
					Protocol:    "vless",
					InboundType: "transport",
					Port:        443,
					Network:     "ws",
					Security:    "tls",
					Fingerprint: "chrome",
					WSSettings:  map[string]any{"path": "/ws/{uuid}"},
					CreatedAt:   time.Now().Format(time.RFC3339),
				}
			}
			if newInbounds["Node"] == nil {
				newInbounds["Node"] = &Inbound{
					InboundID:   "Node",
					Name:        "Node",
					Protocol:    "node",
					InboundType: "node",
					System:      true,
					CreatedAt:   time.Now().Format(time.RFC3339),
				}
			}
			State.Inbounds = newInbounds
			restoredInboundsCount = len(newInbounds)
		}
	}

	if nRaw, ok := stateMap["nodes"]; ok && nRaw != nil {
		newNodes := make(map[string]*Node)
		if nMap, ok := nRaw.(map[string]any); ok {
			for k, raw := range nMap {
				b, _ := json.Marshal(raw)
				var n Node
				if err := json.Unmarshal(b, &n); err == nil {
					if n.NodeID == "" {
						n.NodeID = k
					}
					newNodes[n.NodeID] = &n
				}
			}
		}
		State.Nodes = newNodes
		restoredNodesCount = len(newNodes)
	}

	if pRaw, ok := stateMap["bot_plans"]; ok && pRaw != nil {
		newPlans := make(map[string]*BotPlan)
		if pMap, ok := pRaw.(map[string]any); ok {
			for k, raw := range pMap {
				b, _ := json.Marshal(raw)
				var p BotPlan
				if err := json.Unmarshal(b, &p); err == nil {
					if p.ID == "" {
						p.ID = k
					}
					newPlans[p.ID] = &p
				}
			}
		}
		State.BotPlans = newPlans
		restoredPlansCount = len(newPlans)
	}

	if sMap, ok := stateMap["settings"].(map[string]any); ok && sMap != nil {
		for k, v := range sMap {
			State.Settings[k] = v
		}
	}

	if pwd, ok := stateMap["password_hash"].(string); ok && pwd != "" {
		State.Password = pwd
	}

	State.mu.Unlock()

	_ = saveState()
	loadStoredOrders()
	xrayApply()

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":  true,
		"msg": "بکاپ با موفقیت و به طور کامل بازیابی شد",
		"restored": map[string]any{
			"users":     restoredUsersCount,
			"inbounds":  restoredInboundsCount,
			"nodes":     restoredNodesCount,
			"bot_plans": restoredPlansCount,
		},
	})
}

func handleNotifications(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "items": []any{}, "source": ""})
}

func handleUpdateCheck(w http.ResponseWriter, r *http.Request) {
	currentVer := "4.0.0"
	repoURL := "https://github.com/Itskillmaster/WhitePanel-UI"
	rawJSONURL := "https://raw.githubusercontent.com/Itskillmaster/WhitePanel-UI/main/version.json"

	client := &http.Client{Timeout: 6 * time.Second}
	resp, err := client.Get(rawJSONURL)

	var latestVer string
	var changelog string
	updateAvailable := false

	if err == nil {
		defer resp.Body.Close()
		if resp.StatusCode == 200 {
			var verData struct {
				Version   string `json:"version"`
				Changelog string `json:"changelog"`
			}
			if derr := json.NewDecoder(resp.Body).Decode(&verData); derr == nil && verData.Version != "" {
				latestVer = strings.TrimSpace(verData.Version)
				changelog = verData.Changelog
				if latestVer != currentVer {
					updateAvailable = true
				}
			}
		}
	}

	if latestVer == "" {
		latestVer = currentVer
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":               true,
		"version":          currentVer,
		"latest":           latestVer,
		"update_available": updateAvailable,
		"changelog":        changelog,
		"repo_url":         repoURL,
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

// ============================================================================
// ROUTER
// ============================================================================

func main() {
	initPanelPort() // ✅
	initStorage()

	port := os.Getenv("PANEL_PORT")
	if port == "" {
		port = os.Getenv("PORT")
	}
	if port == "" {
		port = "8080"
	}
	staticDir := "./static"

	go func() {
		if err := ensureXray(); err != nil {
			log.Printf("[Xray] initial install failed: %v", err)
			return
		}
		xrayApply()
	}()

	mux := http.NewServeMux()
	fileServer := http.FileServer(http.Dir(staticDir))
	mux.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, must-revalidate")
		http.StripPrefix("/static/", fileServer).ServeHTTP(w, r)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, must-revalidate")
		switch r.URL.Path {
		case "/", "/white", "/dashboard":
			http.ServeFile(w, r, filepath.Join(staticDir, "index.html"))
			return
		case "/login":
			http.ServeFile(w, r, filepath.Join(staticDir, "login.html"))
			return
		}
		fileServer.ServeHTTP(w, r)
	})

	mux.HandleFunc("/api/login", handleLogin)
	mux.HandleFunc("/api/logout", handleLogout)
	mux.HandleFunc("/api/me", handleMe)
	mux.HandleFunc("/api/server-info", handleServerInfo)
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
					"port":   port,
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

	mux.HandleFunc("/api/sub/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/sub/")
		if strings.HasSuffix(path, "/qr") {
			handleSubQR(w, r, strings.TrimSuffix(path, "/qr"))
			return
		}
		handleAPISub(w, r, path)
	})
	mux.HandleFunc("/sub/", func(w http.ResponseWriter, r *http.Request) {
		handleSubPage(w, r, strings.TrimPrefix(r.URL.Path, "/sub/"))
	})
	mux.HandleFunc("/link/", func(w http.ResponseWriter, r *http.Request) {
		handleSubPlainText(w, r, strings.TrimPrefix(r.URL.Path, "/link/"))
	})

	mux.HandleFunc("/ws/", func(w http.ResponseWriter, r *http.Request) {
		uuid := strings.TrimPrefix(r.URL.Path, "/ws/")
		if uuid == "live" {
			writeJSON(w, http.StatusOK, map[string]any{"ok": true, "message": "live stats not implemented"})
			return
		}
		handleWSRelay(w, r, uuid)
	})
	mux.HandleFunc("/xhttp-siz10/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/xhttp-siz10/")
		parts := strings.Split(path, "/")
		if len(parts) < 3 {
			http.Error(w, "bad xhttp path", 400)
			return
		}
		mode, uuid, sessionID := parts[0], parts[1], parts[2]
		if r.Method == http.MethodGet {
			handleXHTTPDownlink(w, r, mode, uuid, sessionID)
			return
		}
		if mode == "stream-up" {
			handleXHTTPStreamUp(w, r, uuid, sessionID)
			return
		}
		if mode == "packet-up" {
			if len(parts) < 4 {
				http.Error(w, "missing seq", 400)
				return
			}
			handleXHTTPPacketUp(w, r, uuid, sessionID, parts[3])
			return
		}
		http.Error(w, "unknown mode", 404)
	})

	mux.HandleFunc("/stats", requireAuth(handleStats))
	mux.HandleFunc("/api/server/stats", requireAuth(handleStats))

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
		case "test-configs":
			handleUserTestConfigs(w, r, id)
		default:
			writeJSON(w, http.StatusNotFound, map[string]any{"ok": false})
		}
	}))

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

	mux.HandleFunc("/api/tools/generate-reality-keys", requireAuth(handleToolsRealityKeys))
	mux.HandleFunc("/api/tools/reality-settings", requireAuth(handleRealitySettings))

	mux.HandleFunc("/api/xray/status", requireAuth(handleXrayStatus))
	mux.HandleFunc("/api/xray/start", requireAuth(handleXrayStart))
	mux.HandleFunc("/api/xray/stop", requireAuth(handleXrayStop))
	mux.HandleFunc("/api/xray/restart", requireAuth(handleXrayRestart))
	mux.HandleFunc("/api/xray/update", requireAuth(handleXrayUpdate))
	mux.HandleFunc("/api/xray/logs", requireAuth(handleXrayLogs))
	mux.HandleFunc("/api/xray/logs/clear", requireAuth(handleXrayLogsClear))

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
		handleScannerIPS(w, r, strings.TrimPrefix(r.URL.Path, "/api/scanner/ips/"))
	}))

	mux.HandleFunc("/api/bot/status", requireAuth(handleBotStatus))
	mux.HandleFunc("/api/bot/config", requireAuth(handleBotConfig))
	mux.HandleFunc("/api/bot/test", requireAuth(handleBotTest))
	mux.HandleFunc("/api/bot/sell/plans", requireAuth(handleBotSellPlans))
	mux.HandleFunc("/api/bot/sell/plans/", requireAuth(handleBotSellPlanByID))
	mux.HandleFunc("/api/bot/channel/run", requireAuth(handleBotChannelRun))
	mux.HandleFunc("/api/worker", requireAuth(handleWorkerRoute))
	mux.HandleFunc("/api/worker/", requireAuth(handleWorkerRoute))

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

	mux.HandleFunc("/api/notifications", requireAuth(handleNotifications))
	mux.HandleFunc("/api/update-check", requireAuth(handleUpdateCheck))
	mux.HandleFunc("/api/activity", requireAuth(handleActivityLog))
	mux.HandleFunc("/api/musix", requireAuth(handleMusix))
	mux.HandleFunc("/api/tools/my-ip", requireAuth(handleToolsMyIP))

	log.Printf("==================================================")
	log.Printf("  White Panel Go Engine v8.3 + Xray-core")
	log.Printf("  Listening on :%s  (panelPortNum=%d)", port, panelPortNum)
	log.Printf("  Admin:  http://localhost:%s/", port)
	log.Printf("  Login:  http://localhost:%s/login", port)
	log.Printf("  Sub:    http://localhost:%s/sub/{uuid}", port)
	log.Printf("  Link:   http://localhost:%s/link/{uuid}", port)
	log.Printf("  WS:     ws://.../ws/{uuid}")
	log.Printf("  XHTTP:  .../xhttp-siz10/{mode}/{uuid}/{session_id}")
	log.Printf("==================================================")

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 0,
		IdleTimeout:  60 * time.Second,
	}

	go enforcementLoop(context.Background())
	go startAllMTProxies(context.Background())
	go initTelegramBot()

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Printf("[Main] shutting down...")
		stopAllMTProxies()
		xrayStop()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server terminated: %v", err)
	}
}
