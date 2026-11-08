package main

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ==========================================
// Data Structures
// ==========================================

type User struct {
	ID                    string `json:"id"`
	Username              string `json:"username"`
	TrafficLimit          int64  `json:"traffic_limit"`
	TrafficUsed           int64  `json:"traffic_used"`
	ExpireAt              string `json:"expire_at"`
	Status                string `json:"status"`
	ConfigUUID            string `json:"config_uuid"`
	SubscriptionURL       string `json:"subscription_url"`
	ConcurrentConnections int    `json:"concurrent_connections"`
	CreatedAt             string `json:"created_at"`
	Note                  string `json:"note"`
}

type Inbound struct {
	ID         string `json:"id"`
	Tag        string `json:"tag"`
	Protocol   string `json:"protocol"`
	Port       int    `json:"port"`
	Network    string `json:"network"`
	Security   string `json:"security"`
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
	ShortIDs   string `json:"short_ids"`
	SNI        string `json:"sni"`
	WhiteX     string `json:"whitex"`
	Enabled    bool   `json:"enabled"`
}

type Node struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Address   string `json:"address"`
	Port      int    `json:"port"`
	Status    string `json:"status"`
	Heartbeat string `json:"heartbeat"`
}

type ProxyItem struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Server  string `json:"server"`
	Port    int    `json:"port"`
	Enabled bool   `json:"enabled"`
}

type BotPlan struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	TrafficGB   int     `json:"traffic_gb"`
	ExpireDays  int     `json:"expire_days"`
	PriceToman  float64 `json:"price_toman"`
	Description string  `json:"description"`
}

type AppState struct {
	mu           sync.RWMutex
	Users        map[string]*User      `json:"users"`
	Inbounds     map[string]*Inbound   `json:"inbounds"`
	Nodes        map[string]*Node      `json:"nodes"`
	Proxies      map[string]*ProxyItem `json:"proxies"`
	BotPlans     map[string]*BotPlan   `json:"bot_plans"`
	Settings     map[string]any        `json:"settings"`
	ScannerCache map[string]any        `json:"scanner_cache"`
	DataDir      string                `json:"-"`
}

var State = &AppState{
	Users:        make(map[string]*User),
	Inbounds:     make(map[string]*Inbound),
	Nodes:        make(map[string]*Node),
	Proxies:      make(map[string]*ProxyItem),
	BotPlans:     make(map[string]*BotPlan),
	Settings:     make(map[string]any),
	ScannerCache: make(map[string]any),
	DataDir:      "./data",
}

// In-memory session store (restart = everyone logs out — acceptable for small panel)
var (
	sessionMu sync.Mutex
	sessions  = make(map[string]time.Time)
)

const sessionTTL = 7 * 24 * time.Hour

// ==========================================
// Storage
// ==========================================

func initStorage() {
	if envDir := os.Getenv("WHITE_DATA_DIR"); envDir != "" {
		State.DataDir = envDir
	}
	_ = os.MkdirAll(State.DataDir, 0755)
	loadState()
}

func loadState() {
	State.mu.Lock()
	defer State.mu.Unlock()

	dataFile := filepath.Join(State.DataDir, "white_state.json")
	f, err := os.Open(dataFile)
	if err != nil {
		log.Printf("[White Storage] Creating fresh state at %s", dataFile)
		State.Settings["panel_port"] = 8080
		State.Settings["secret_key"] = "white-panel-secret-key-v2"
		State.Settings["username"] = "admin"
		State.Settings["password_hash"] = hashPassword("admin")
		State.Settings["panel_api_key"] = generateToken(32)
		State.Settings["server_ip"] = "127.0.0.1"
		State.Settings["domain"] = ""

		defInbound := &Inbound{
			ID: "inbound-1", Tag: "VLESS-Reality", Protocol: "vless", Port: 443,
			Network: "tcp", Security: "reality", SNI: "www.google.com",
			WhiteX: "/", Enabled: true,
		}
		State.Inbounds[defInbound.ID] = defInbound
		return
	}
	defer f.Close()

	var diskState struct {
		Users        map[string]*User      `json:"users"`
		Inbounds     map[string]*Inbound   `json:"inbounds"`
		Nodes        map[string]*Node      `json:"nodes"`
		Proxies      map[string]*ProxyItem `json:"proxies"`
		BotPlans     map[string]*BotPlan   `json:"bot_plans"`
		Settings     map[string]any        `json:"settings"`
		ScannerCache map[string]any        `json:"scanner_cache"`
	}

	if err := json.NewDecoder(f).Decode(&diskState); err == nil {
		if diskState.Users != nil {
			State.Users = diskState.Users
		}
		if diskState.Inbounds != nil {
			State.Inbounds = diskState.Inbounds
		}
		if diskState.Nodes != nil {
			State.Nodes = diskState.Nodes
		}
		if diskState.Proxies != nil {
			State.Proxies = diskState.Proxies
		}
		if diskState.BotPlans != nil {
			State.BotPlans = diskState.BotPlans
		}
		if diskState.Settings != nil {
			State.Settings = diskState.Settings
		}
		if diskState.ScannerCache != nil {
			State.ScannerCache = diskState.ScannerCache
		}
		log.Printf("[White Storage] Loaded: %d users, %d inbounds, %d nodes",
			len(State.Users), len(State.Inbounds), len(State.Nodes))
	}
}

// Atomic write: temp file + rename, so a crash can't corrupt the state.
func saveState() error {
	State.mu.RLock()
	defer State.mu.RUnlock()

	dataFile := filepath.Join(State.DataDir, "white_state.json")
	tmpFile := dataFile + ".tmp"

	f, err := os.Create(tmpFile)
	if err != nil {
		return err
	}

	diskState := struct {
		Users        map[string]*User      `json:"users"`
		Inbounds     map[string]*Inbound   `json:"inbounds"`
		Nodes        map[string]*Node      `json:"nodes"`
		Proxies      map[string]*ProxyItem `json:"proxies"`
		BotPlans     map[string]*BotPlan   `json:"bot_plans"`
		Settings     map[string]any        `json:"settings"`
		ScannerCache map[string]any        `json:"scanner_cache"`
	}{
		Users: State.Users, Inbounds: State.Inbounds, Nodes: State.Nodes,
		Proxies: State.Proxies, BotPlans: State.BotPlans,
		Settings: State.Settings, ScannerCache: State.ScannerCache,
	}

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(diskState); err != nil {
		f.Close()
		os.Remove(tmpFile)
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		os.Remove(tmpFile)
		return err
	}
	if err := f.Close(); err != nil {
		os.Remove(tmpFile)
		return err
	}
	return os.Rename(tmpFile, dataFile)
}

// ==========================================
// Utilities
// ==========================================

func generateUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func generateToken(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func hashPassword(pwd string) string {
	h := sha256.Sum256([]byte(pwd))
	return hex.EncodeToString(h[:])
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// ==========================================
// Sessions & Auth
// ==========================================

func createSession() string {
	token := generateToken(32)
	sessionMu.Lock()
	sessions[token] = time.Now().Add(sessionTTL)
	sessionMu.Unlock()
	return token
}

func isValidSession(token string) bool {
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

// Real auth check: validates the session cookie or the X-API-Key header.
func checkAuth(r *http.Request) bool {
	if c, err := r.Cookie("white_session"); err == nil && c.Value != "" {
		if isValidSession(c.Value) {
			return true
		}
	}
	if apiKey := r.Header.Get("X-API-Key"); apiKey != "" {
		State.mu.RLock()
		expectedKey, _ := State.Settings["panel_api_key"].(string)
		State.mu.RUnlock()
		if expectedKey != "" && subtle.ConstantTimeCompare([]byte(apiKey), []byte(expectedKey)) == 1 {
			return true
		}
	}
	return false
}

// Middleware for protected routes.
func requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !checkAuth(r) {
			writeJSON(w, http.StatusUnauthorized, map[string]any{
				"ok": false, "error": "unauthorized", "detail": "احراز هویت لازم است",
			})
			return
		}
		next(w, r)
	}
}

// ==========================================
// Handlers: Health / Info / Stats
// ==========================================

func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "ok", "service": "WhitePanel-Go-Core",
		"version": "4.0.1", "time": time.Now().Unix(),
	})
}

func handlePublicEndpoint(w http.ResponseWriter, r *http.Request) {
	host := r.Host
	proto := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		proto = "https"
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok": true, "endpoint": fmt.Sprintf("%s://%s", proto, host), "host": host,
	})
}

func handleServerInfo(w http.ResponseWriter, r *http.Request) {
	State.mu.RLock()
	ip, _ := State.Settings["server_ip"].(string)
	if ip == "" {
		ip = "127.0.0.1"
	}
	nodeCount := len(State.Nodes)
	State.mu.RUnlock()

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":           true,
		"server_ip":    ip,
		"public_ip":    ip,
		"country":      "IR",
		"country_code": "IR",
		"country_flag": "🇮🇷",
		"version":      "4.0.1",
		"core":         "Golang Engine",
		"os":           "Linux/Windows",
		"node_count":   nodeCount,
		"uptime":       time.Now().Unix(),
	})
}

func handleServerStats(w http.ResponseWriter, r *http.Request) {
	State.mu.RLock()
	totalUsers := len(State.Users)
	activeUsers := 0
	var usedTraffic int64
	for _, u := range State.Users {
		if u.Status == "active" {
			activeUsers++
		}
		usedTraffic += u.TrafficUsed
	}
	State.mu.RUnlock()

	usedGB := float64(usedTraffic) / 1073741824.0
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":                    true,
		"users_count":           totalUsers,
		"active_users":          activeUsers,
		"total_users":           totalUsers,
		"total_traffic_used_gb": usedGB,
		"traffic_usage_gb":      usedGB,
		"total_traffic_used_tb": usedGB / 1024.0,
		"total_traffic":         usedTraffic,
		"used_traffic":          usedTraffic,
		"active_connections":    0,
		"tcp_connections":       0,
		"udp_connections":       0,
		"cpu":                   12.4,
		"cpu_percent":           12.4,
		"cpu_count":             4,
		"memory":                28.6,
		"ram_percent":           28.6,
		"ram_used_gb":           4.5,
		"ram_total_gb":          16.0,
		"swap_percent":          0.0,
		"swap_used_gb":          0.0,
		"swap_total_gb":         0.0,
		"disk_percent":          35.0,
		"disk_used_gb":          70.0,
		"disk_total_gb":         200.0,
		"uptime_seconds":        3600,
		"net_sent_bytes":        0,
		"net_recv_bytes":        0,
		"public_ipv4":           "",
		"public_ipv6":           "",
	})
}

// ==========================================
// Login / Logout / Me
// ==========================================

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"ok": false, "detail": "پاسخ نامعتبر",
		})
		return
	}

	State.mu.RLock()
	expectedUser, _ := State.Settings["username"].(string)
	expectedHash, _ := State.Settings["password_hash"].(string)
	State.mu.RUnlock()
	if expectedUser == "" {
		expectedUser = "admin"
	}

	// Login page sends only the password; treat empty username as the configured admin user.
	if req.Username == "" {
		req.Username = expectedUser
	}

	inputHash := hashPassword(req.Password)
	if req.Username == expectedUser && expectedHash != "" && inputHash == expectedHash {
		token := createSession()
		http.SetCookie(w, &http.Cookie{
			Name:     "white_session",
			Value:    token,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   int(sessionTTL.Seconds()),
		})
		writeJSON(w, http.StatusOK, map[string]any{
			"ok": true, "token": token,
		})
		return
	}

	writeJSON(w, http.StatusUnauthorized, map[string]any{
		"ok":     false,
		"error":  "نام کاربری یا رمز عبور اشتباه است",
		"detail": "نام کاربری یا رمز عبور اشتباه است",
	})
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie("white_session"); err == nil {
		sessionMu.Lock()
		delete(sessions, c.Value)
		sessionMu.Unlock()
	}
	http.SetCookie(w, &http.Cookie{
		Name: "white_session", Value: "", Path: "/", MaxAge: -1, HttpOnly: true,
	})
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func handleMe(w http.ResponseWriter, r *http.Request) {
	if !checkAuth(r) {
		writeJSON(w, http.StatusOK, map[string]any{
			"ok": true, "authenticated": false,
		})
		return
	}
	State.mu.RLock()
	username, _ := State.Settings["username"].(string)
	apiKey, _ := State.Settings["panel_api_key"].(string)
	State.mu.RUnlock()

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":            true,
		"authenticated": true,
		"username":      username,
		"api_key":       apiKey,
		"role":          "admin",
	})
}

func handlePanelApiKey(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		newKey := generateToken(32)
		State.mu.Lock()
		State.Settings["panel_api_key"] = newKey
		State.mu.Unlock()
		_ = saveState()
	}
	State.mu.RLock()
	apiKey, _ := State.Settings["panel_api_key"].(string)
	State.mu.RUnlock()
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "api_key": apiKey})
}

func handleNotifications(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"ok": true, "notifications": []any{}, "items": []any{}, "source": "",
	})
}

// ==========================================
// Users
// ==========================================

func handleUsersRoute(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/users")
	path = strings.Trim(path, "/")

	if path == "" {
		switch r.Method {
		case http.MethodGet:
			State.mu.RLock()
			list := make([]*User, 0, len(State.Users))
			for _, u := range State.Users {
				list = append(list, u)
			}
			State.mu.RUnlock()
			writeJSON(w, http.StatusOK, map[string]any{
				"ok": true, "users": list, "list": list,
			})
		case http.MethodPost:
			var req struct {
				Username     string `json:"username"`
				TrafficLimit int64  `json:"traffic_limit"`
				ExpireDays   int    `json:"expire_days"`
				Note         string `json:"note"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			uid := generateUUID()
			cfgUUID := generateUUID()
			now := time.Now()
			exp := now.AddDate(0, 0, req.ExpireDays).Format(time.RFC3339)
			user := &User{
				ID: uid, Username: req.Username, TrafficLimit: req.TrafficLimit,
				ExpireAt: exp, Status: "active", ConfigUUID: cfgUUID,
				SubscriptionURL: fmt.Sprintf("/sub/%s", cfgUUID),
				CreatedAt:       now.Format(time.RFC3339), Note: req.Note,
			}
			State.mu.Lock()
			State.Users[uid] = user
			State.mu.Unlock()
			_ = saveState()
			writeJSON(w, http.StatusCreated, map[string]any{"ok": true, "user": user})
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	parts := strings.Split(path, "/")
	uid := parts[0]

	State.mu.Lock()
	user, exists := State.Users[uid]
	if !exists {
		for _, u := range State.Users {
			if u.Username == uid {
				user = u
				exists = true
				break
			}
		}
	}
	if !exists {
		State.mu.Unlock()
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if len(parts) > 1 {
		switch parts[1] {
		case "toggle":
			if user.Status == "active" {
				user.Status = "disabled"
			} else {
				user.Status = "active"
			}
			State.mu.Unlock()
			_ = saveState()
			writeJSON(w, http.StatusOK, map[string]any{"ok": true, "status": user.Status})
			return
		case "reset":
			user.TrafficUsed = 0
			State.mu.Unlock()
			_ = saveState()
			writeJSON(w, http.StatusOK, map[string]any{"ok": true, "message": "Traffic reset"})
			return
		}
	}

	if r.Method == http.MethodDelete {
		delete(State.Users, user.ID)
		State.mu.Unlock()
		_ = saveState()
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
		return
	}

	State.mu.Unlock()
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "user": user})
}

// ==========================================
// Inbounds
// ==========================================

func handleInboundsRoute(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/inbounds")
	path = strings.Trim(path, "/")

	if path == "" {
		switch r.Method {
		case http.MethodGet:
			State.mu.RLock()
			list := make([]*Inbound, 0, len(State.Inbounds))
			for _, ib := range State.Inbounds {
				list = append(list, ib)
			}
			State.mu.RUnlock()
			writeJSON(w, http.StatusOK, map[string]any{"ok": true, "inbounds": list})
		case http.MethodPost:
			var ib Inbound
			if err := json.NewDecoder(r.Body).Decode(&ib); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if ib.ID == "" {
				ib.ID = generateUUID()
			}
			if ib.WhiteX == "" {
				ib.WhiteX = "/"
			}
			ib.Enabled = true
			State.mu.Lock()
			State.Inbounds[ib.ID] = &ib
			State.mu.Unlock()
			_ = saveState()
			writeJSON(w, http.StatusCreated, map[string]any{"ok": true, "inbound": ib})
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	parts := strings.Split(path, "/")
	iid := parts[0]

	if len(parts) > 1 && parts[1] == "generate-short-id" {
		writeJSON(w, http.StatusOK, map[string]any{
			"ok": true, "short_id": generateToken(4),
		})
		return
	}

	State.mu.Lock()
	defer State.mu.Unlock()

	if r.Method == http.MethodDelete {
		delete(State.Inbounds, iid)
		_ = saveState()
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
		return
	}

	ib, exists := State.Inbounds[iid]
	if !exists {
		http.Error(w, "Inbound not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "inbound": ib})
}

func handleRealityKeys(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":          true,
		"public_key":  generateToken(22),
		"private_key": generateToken(22),
		"short_id":    generateToken(4),
	})
}

// ==========================================
// Nodes
// ==========================================

func handleNodesRoute(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/nodes")
	path = strings.Trim(path, "/")

	if path == "" {
		switch r.Method {
		case http.MethodGet:
			State.mu.RLock()
			list := make([]*Node, 0, len(State.Nodes))
			for _, n := range State.Nodes {
				list = append(list, n)
			}
			State.mu.RUnlock()
			writeJSON(w, http.StatusOK, map[string]any{"ok": true, "nodes": list})
		case http.MethodPost:
			var n Node
			if err := json.NewDecoder(r.Body).Decode(&n); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if n.ID == "" {
				n.ID = generateUUID()
			}
			n.Status = "online"
			n.Heartbeat = time.Now().Format(time.RFC3339)
			State.mu.Lock()
			State.Nodes[n.ID] = &n
			State.mu.Unlock()
			_ = saveState()
			writeJSON(w, http.StatusCreated, map[string]any{"ok": true, "node": n})
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	if path == "sync-all" {
		State.mu.RLock()
		count := len(State.Nodes)
		State.mu.RUnlock()
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "synced": count})
		return
	}

	nid := path
	State.mu.Lock()
	defer State.mu.Unlock()

	if r.Method == http.MethodDelete {
		delete(State.Nodes, nid)
		_ = saveState()
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
		return
	}

	n, exists := State.Nodes[nid]
	if !exists {
		http.Error(w, "Node not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "node": n})
}

// ==========================================
// Scanner
// ==========================================

func handleScanner(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/scanner")
	path = strings.Trim(path, "/")

	if strings.HasPrefix(path, "resolve") {
		host := r.URL.Query().Get("host")
		ips, err := net.LookupIP(host)
		res := []string{}
		if err == nil {
			for _, ip := range ips {
				res = append(res, ip.String())
			}
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "ips": res})
		return
	}

	if strings.HasPrefix(path, "sni-check") {
		host := r.URL.Query().Get("host")
		start := time.Now()
		conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, "443"), 2*time.Second)
		if err == nil {
			conn.Close()
			writeJSON(w, http.StatusOK, map[string]any{
				"ok": true, "host": host, "status": "reachable",
				"latency_ms": time.Since(start).Milliseconds(),
			})
		} else {
			writeJSON(w, http.StatusOK, map[string]any{
				"ok": false, "host": host, "status": "unreachable",
			})
		}
		return
	}

	if strings.HasPrefix(path, "cf-subnets") {
		writeJSON(w, http.StatusOK, map[string]any{
			"ok": true,
			"subnets": []string{
				"173.245.48.0/20", "103.21.244.0/22", "103.22.200.0/22",
				"103.31.4.0/22", "141.101.64.0/18", "108.162.192.0/18",
				"190.93.240.0/20", "188.114.96.0/20", "197.234.240.0/22",
				"198.41.128.0/17", "162.158.0.0/15", "104.16.0.0/13",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok": true, "results": []any{}, "ips": []any{},
		"seq": 0, "message": "Scanner ready",
	})
}

// ==========================================
// Bot / Worker stubs
// ==========================================

func handleBotRoute(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/bot")
	path = strings.Trim(path, "/")

	if strings.HasPrefix(path, "sell/plans") {
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
			var p BotPlan
			if err := json.NewDecoder(r.Body).Decode(&p); err == nil {
				if p.ID == "" {
					p.ID = generateUUID()
				}
				State.mu.Lock()
				State.BotPlans[p.ID] = &p
				State.mu.Unlock()
				_ = saveState()
			}
			writeJSON(w, http.StatusOK, map[string]any{"ok": true, "plan": p})
		default:
			writeJSON(w, http.StatusOK, map[string]any{"ok": true})
		}
		return
	}

	if path == "status" {
		writeJSON(w, http.StatusOK, map[string]any{
			"ok": true, "running": true, "token_valid": true,
			"token_configured": false, "token_masked": "",
			"channel": map[string]any{"enabled": false},
			"sell":    map[string]any{"enabled": false, "required_channels": []any{}, "plans": []any{}},
			"orders":  map[string]any{},
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func handleWorkerRoute(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"ok": true, "status": "ready", "proxies": []any{},
		"connected": false,
		"worker":    map[string]any{"connected": false, "locations": []any{}},
		"tunnel":    map[string]any{"enabled": false, "logs": []any{}},
		"reverse":   map[string]any{"enabled": false, "logs": []any{}},
	})
}

// ==========================================
// Subscription page + JSON API
// ==========================================

// isBrowserUA decides between serving HTML (browsers) and plain text (clients).
func isBrowserUA(ua string) bool {
	ua = strings.ToLower(ua)
	clientHints := []string{
		"v2ray", "xray", "clash", "sing-box", "singbox", "shadowrocket",
		"hiddify", "nekobox", "v2box", "streisand", "wireguard",
		"quantumult", "surge", "stash", "loon", "passwall",
	}
	for _, c := range clientHints {
		if strings.Contains(ua, c) {
			return false
		}
	}
	return true
}

// findUserBySubUUID looks up a user by config_uuid or id.
func findUserBySubUUID(subUUID string) *User {
	State.mu.RLock()
	defer State.mu.RUnlock()
	for _, u := range State.Users {
		if u.ConfigUUID == subUUID || u.ID == subUUID {
			return u
		}
	}
	return nil
}

// buildConfigs returns the list of vless:// links for a user.
func buildConfigs(r *http.Request, u *User) []string {
	host := r.Host
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	State.mu.RLock()
	inbounds := make([]*Inbound, 0, len(State.Inbounds))
	for _, ib := range State.Inbounds {
		if ib.Enabled {
			inbounds = append(inbounds, ib)
		}
	}
	State.mu.RUnlock()

	configs := []string{}
	for _, ib := range inbounds {
		port := ib.Port
		if port == 0 {
			port = 443
		}
		sni := ib.SNI
		if sni == "" {
			sni = "www.google.com"
		}
		security := ib.Security
		if security == "" {
			security = "tls"
		}
		network := ib.Network
		if network == "" {
			network = "tcp"
		}
		cfg := fmt.Sprintf("vless://%s@%s:%d?security=%s&type=%s&sni=%s#White-%s-%s",
			u.ConfigUUID, host, port, security, network, sni, u.Username, ib.Tag)
		configs = append(configs, cfg)
	}
	if len(configs) == 0 {
		configs = append(configs, fmt.Sprintf(
			"vless://%s@%s:443?security=reality&type=tcp&sni=www.google.com#White-%s",
			u.ConfigUUID, host, u.Username))
	}
	return configs
}

// /api/sub/{uuid} — JSON payload for sub.html.
func handleApiSub(w http.ResponseWriter, r *http.Request) {
	trimmed := strings.TrimPrefix(r.URL.Path, "/api/sub/")
	subUUID := strings.Trim(trimmed, "/")
	if subUUID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"ok": false, "detail": "invalid subscription",
		})
		return
	}

	u := findUserBySubUUID(subUUID)
	if u == nil {
		writeJSON(w, http.StatusNotFound, map[string]any{
			"ok": false, "detail": "کاربر پیدا نشد",
		})
		return
	}

	var expireDays *int
	var expireAtTs *int64
	if u.ExpireAt != "" {
		if t, err := time.Parse(time.RFC3339, u.ExpireAt); err == nil {
			days := int(time.Until(t).Hours() / 24)
			if days < 0 {
				days = 0
			}
			expireDays = &days
			ts := t.Unix()
			expireAtTs = &ts
		}
	}

	configs := buildConfigs(r, u)

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":                  true,
		"config_uuid":         u.ConfigUUID,
		"username":            u.Username,
		"is_active":           u.Status == "active",
		"status":              u.Status,
		"traffic_used_bytes":  u.TrafficUsed,
		"traffic_limit_bytes": u.TrafficLimit,
		"download_bytes":      u.TrafficUsed,
		"upload_bytes":        0,
		"expire_days":         expireDays,
		"expire_at_ts":        expireAtTs,
		"max_ip_per_user":     u.ConcurrentConnections,
		"used_ips":            0,
		"configs":             configs,
		"subscription_url":    "/link/" + u.ConfigUUID,
	})
}

// /sub/{uuid} — HTML page for browsers, plain text for clients.
func handleSubPage(w http.ResponseWriter, r *http.Request) {
	trimmed := strings.TrimPrefix(r.URL.Path, "/sub/")
	subUUID := strings.Trim(trimmed, "/")
	if subUUID == "" {
		http.Error(w, "Invalid subscription URL", http.StatusBadRequest)
		return
	}
	if u := findUserBySubUUID(subUUID); u == nil {
		http.Error(w, "Subscription not found", http.StatusNotFound)
		return
	}
	if !isBrowserUA(r.UserAgent()) {
		writeSubPlainText(w, r, subUUID)
		return
	}
	http.ServeFile(w, r, filepath.Join("./static", "sub.html"))
}

// /link/{uuid} — canonical plain-text subscription URL for clients.
func handleSubPlainText(w http.ResponseWriter, r *http.Request) {
	trimmed := strings.TrimPrefix(r.URL.Path, "/link/")
	subUUID := strings.Trim(trimmed, "/")
	if subUUID == "" {
		http.Error(w, "Invalid subscription URL", http.StatusBadRequest)
		return
	}
	writeSubPlainText(w, r, subUUID)
}

func writeSubPlainText(w http.ResponseWriter, r *http.Request, subUUID string) {
	u := findUserBySubUUID(subUUID)
	if u == nil {
		http.Error(w, "Subscription not found", http.StatusNotFound)
		return
	}
	if u.Status != "active" {
		http.Error(w, "Subscription disabled", http.StatusForbidden)
		return
	}
	if u.ExpireAt != "" {
		if t, err := time.Parse(time.RFC3339, u.ExpireAt); err == nil && time.Now().After(t) {
			http.Error(w, "Subscription expired", http.StatusForbidden)
			return
		}
	}

	configs := buildConfigs(r, u)
	expireUnix := int64(0)
	if u.ExpireAt != "" {
		if t, err := time.Parse(time.RFC3339, u.ExpireAt); err == nil {
			expireUnix = t.Unix()
		}
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Subscription-Userinfo", fmt.Sprintf(
		"upload=0; download=%d; total=%d; expire=%d",
		u.TrafficUsed, u.TrafficLimit, expireUnix))
	w.Header().Set("Profile-Update-Interval", "24")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(strings.Join(configs, "\n")))
}

// ==========================================
// main
// ==========================================

func main() {
	initStorage()

	port := os.Getenv("PANEL_PORT")
	if port == "" {
		port = "8080"
	}

	staticDir := "./static"
	mux := http.NewServeMux()

	// Static assets
	fs := http.FileServer(http.Dir(staticDir))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Root handler: / and /white → admin panel; /login → login page.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/", "/white", "/dashboard":
			http.ServeFile(w, r, filepath.Join(staticDir, "index.html"))
			return
		case "/login":
			http.ServeFile(w, r, filepath.Join(staticDir, "login.html"))
			return
		}
		fs.ServeHTTP(w, r)
	})

	// Public endpoints
	mux.HandleFunc("/healthz", handleHealth)
	mux.HandleFunc("/api/login", handleLogin)
	mux.HandleFunc("/api/logout", handleLogout)
	mux.HandleFunc("/api/me", handleMe)
	mux.HandleFunc("/api/notifications", handleNotifications)
	mux.HandleFunc("/api/runtime/public-endpoint", handlePublicEndpoint)
	mux.HandleFunc("/api/server-info", handleServerInfo)
	mux.HandleFunc("/api/server/stats", handleServerStats)
	mux.HandleFunc("/stats", handleServerStats)

	// Subscription endpoints
	mux.HandleFunc("/sub/", handleSubPage)
	mux.HandleFunc("/link/", handleSubPlainText)
	mux.HandleFunc("/api/sub/", handleApiSub)

	// Protected admin endpoints
	mux.HandleFunc("/api/panel-api-key", requireAuth(handlePanelApiKey))
	mux.HandleFunc("/api/users/", requireAuth(handleUsersRoute))
	mux.HandleFunc("/api/users", requireAuth(handleUsersRoute))
	mux.HandleFunc("/api/inbounds/", requireAuth(handleInboundsRoute))
	mux.HandleFunc("/api/inbounds", requireAuth(handleInboundsRoute))
	mux.HandleFunc("/api/nodes/", requireAuth(handleNodesRoute))
	mux.HandleFunc("/api/nodes", requireAuth(handleNodesRoute))
	mux.HandleFunc("/api/tools/generate-reality-keys", requireAuth(handleRealityKeys))

	mux.HandleFunc("/api/scanner/", requireAuth(handleScanner))
	mux.HandleFunc("/api/scanner", requireAuth(handleScanner))
	mux.HandleFunc("/api/bot/", requireAuth(handleBotRoute))
	mux.HandleFunc("/api/worker/", requireAuth(handleWorkerRoute))

	log.Printf("==================================================")
	log.Printf("  White Panel Core Engine (Go) — v4.0.1")
	log.Printf("  Serving on port: %s", port)
	log.Printf("  Admin panel:  http://localhost:%s/", port)
	log.Printf("  Login page:   http://localhost:%s/login", port)
	log.Printf("  Sub page:     http://localhost:%s/sub/{uuid}", port)
	log.Printf("  Sub API:      http://localhost:%s/api/sub/{uuid}", port)
	log.Printf("==================================================")

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("WhitePanel server terminated: %v", err)
	}
}
