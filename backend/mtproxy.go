package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

type MTProxySecret struct {
	UserID     string
	ConfigUUID string
	Username   string
}

type MTProxyServer struct {
	mu        sync.Mutex
	inboundID string
	port      int
	secrets   map[string]MTProxySecret
	cmd       *exec.Cmd
	logFile   *os.File
	running   bool
}

var (
	tgProxyMu       sync.Mutex
	tgProxyRegistry = map[string]*MTProxyServer{}
)

var mtproxySecretRegex = regexp.MustCompile(`^[0-9a-fA-F]{32}$`)

func mtproxyDir() string {
	d := filepath.Join(State.DataDir, "mtproxy")
	_ = os.MkdirAll(d, 0755)
	return d
}

func mtproxyBinPath() string {
	if p := os.Getenv("MTPROXY_BIN"); p != "" {
		return p
	}
	local := "./mtproto-proxy"
	if st, err := os.Stat(local); err == nil && !st.IsDir() {
		return local
	}
	return "/usr/local/bin/mtproto-proxy"
}

func downloadMTProxyFiles() (string, string, error) {
	dir := mtproxyDir()
	secretFile := filepath.Join(dir, "proxy-secret")
	configFile := filepath.Join(dir, "proxy-multi.conf")
	client := &http.Client{Timeout: 60 * time.Second}

	if st, err := os.Stat(secretFile); err != nil || st.Size() == 0 {
		resp, err := client.Get("https://core.telegram.org/getProxySecret")
		if err != nil {
			return "", "", fmt.Errorf("download proxy-secret: %w", err)
		}
		defer resp.Body.Close()
		f, err := os.Create(secretFile)
		if err != nil {
			return "", "", err
		}
		_, _ = io.Copy(f, resp.Body)
		f.Close()
	}
	needRefresh := true
	if st, err := os.Stat(configFile); err == nil && st.Size() > 0 {
		if time.Since(st.ModTime()) < 24*time.Hour {
			needRefresh = false
		}
	}
	if needRefresh {
		resp, err := client.Get("https://core.telegram.org/getProxyConfig")
		if err != nil {
			return "", "", fmt.Errorf("download proxy-multi.conf: %w", err)
		}
		defer resp.Body.Close()
		f, err := os.Create(configFile)
		if err != nil {
			return "", "", err
		}
		_, _ = io.Copy(f, resp.Body)
		f.Close()
	}
	return secretFile, configFile, nil
}

func mtproxyInternalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return ""
	}
	defer conn.Close()
	if addr, ok := conn.LocalAddr().(*net.UDPAddr); ok {
		return addr.IP.String()
	}
	return ""
}

func mtproxyPublicIP() string {
	client := &http.Client{Timeout: 5 * time.Second}
	for _, u := range []string{"https://api.ipify.org", "https://ifconfig.me/ip"} {
		resp, err := client.Get(u)
		if err != nil || resp.StatusCode != 200 {
			if resp != nil {
				resp.Body.Close()
			}
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		ip := strings.TrimSpace(string(body))
		if ip != "" {
			return ip
		}
	}
	return ""
}

func (s *MTProxyServer) statsPort() int {
	h := sha256.Sum256([]byte(s.inboundID))
	return 18080 + int(h[0])%1000
}

func (s *MTProxyServer) UpdateSecrets(secrets map[string]MTProxySecret) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cleaned := make(map[string]MTProxySecret, len(secrets))
	for secret, info := range secrets {
		secret = strings.ToLower(strings.TrimSpace(secret))
		if mtproxySecretRegex.MatchString(secret) {
			cleaned[secret] = info
		}
	}
	s.secrets = cleaned
}

func (s *MTProxyServer) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return nil
	}
	if len(s.secrets) == 0 {
		log.Printf("[MTProxy %s] no secrets — idle until a user is assigned", s.inboundID)
		return nil
	}

	bin := mtproxyBinPath()
	if st, err := os.Stat(bin); err != nil || st.IsDir() {
		return fmt.Errorf("mtproto-proxy binary missing at %s (set MTPROXY_BIN or place it in CWD)", bin)
	}

	secretFile, configFile, err := downloadMTProxyFiles()
	if err != nil {
		return err
	}

	internalIP := mtproxyInternalIP()
	publicIP := mtproxyPublicIP()

	args := []string{
		"-p", fmt.Sprintf("%d", s.statsPort()),
		"-H", fmt.Sprintf("%d", s.port),
		"-M", "1",
		"--aes-pwd", secretFile,
		"-u", "nobody",
		configFile,
		"--allow-skip-dh",
	}
	if internalIP != "" && publicIP != "" {
		args = append(args, "--nat-info", internalIP+":"+publicIP)
	}
	for secret := range s.secrets {
		args = append(args, "-S", secret)
	}

	logPath := filepath.Join(mtproxyDir(), "mtproxy-"+s.inboundID+".log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	if err := cmd.Start(); err != nil {
		logFile.Close()
		return fmt.Errorf("start mtproto-proxy: %w", err)
	}

	s.cmd = cmd
	s.logFile = logFile
	s.running = true

	go func() {
		err := cmd.Wait()
		s.mu.Lock()
		if s.cmd == cmd {
			s.running = false
			s.cmd = nil
		}
		lf := s.logFile
		s.logFile = nil
		s.mu.Unlock()
		if lf != nil {
			lf.Close()
		}
		log.Printf("[MTProxy %s] exited: %v", s.inboundID, err)
	}()

	log.Printf("[MTProxy %s] started on port %d with %d users", s.inboundID, s.port, len(s.secrets))
	return nil
}

func (s *MTProxyServer) Stop() {
	s.mu.Lock()
	cmd := s.cmd
	s.cmd = nil
	s.running = false
	logFile := s.logFile
	s.logFile = nil
	s.mu.Unlock()

	if cmd != nil && cmd.Process != nil {
		_ = cmd.Process.Kill()
		time.Sleep(200 * time.Millisecond)
	}
	if logFile != nil {
		logFile.Close()
	}
	log.Printf("[MTProxy %s] stopped", s.inboundID)
}

func (s *MTProxyServer) Restart(ctx context.Context) error {
	s.Stop()
	return s.Start(ctx)
}

// ---------- Registry ----------

func startMTProxyForInbound(ctx context.Context, inboundID string, ib *Inbound) error {
	tg := ib.TelegramSettings
	if tg == nil {
		tg = map[string]any{}
	}
	intPort := gi(tg, "internal_port")
	if intPort == 0 {
		intPort = 44344
	}

	secrets := map[string]MTProxySecret{}
	State.mu.RLock()
	for _, u := range State.Users {
		if !isActiveUser(u) {
			continue
		}
		for _, iid := range u.InboundIDs {
			if iid == inboundID {
				secret := strings.ToLower(u.TelegramSecret)
				if !mtproxySecretRegex.MatchString(secret) {
					secret = sha256Hex("white-tg-proxy:" + u.ConfigUUID)[:32]
					State.mu.RUnlock()
					State.mu.Lock()
					u.TelegramSecret = secret
					State.mu.Unlock()
					State.mu.RLock()
				}
				secrets[secret] = MTProxySecret{
					UserID:     u.UserID,
					ConfigUUID: u.ConfigUUID,
					Username:   u.Username,
				}
				break
			}
		}
	}
	State.mu.RUnlock()

	tgProxyMu.Lock()
	server := tgProxyRegistry[inboundID]
	if server == nil {
		server = &MTProxyServer{
			inboundID: inboundID,
			port:      intPort,
			secrets:   map[string]MTProxySecret{},
		}
		tgProxyRegistry[inboundID] = server
	} else {
		server.port = intPort
	}
	tgProxyMu.Unlock()

	server.UpdateSecrets(secrets)
	return server.Start(ctx)
}

func stopMTProxyForInbound(inboundID string) {
	tgProxyMu.Lock()
	server := tgProxyRegistry[inboundID]
	delete(tgProxyRegistry, inboundID)
	tgProxyMu.Unlock()
	if server != nil {
		server.Stop()
	}
}

func restartMTProxyForInbound(ctx context.Context, inboundID string) {
	State.mu.RLock()
	ib := State.Inbounds[inboundID]
	State.mu.RUnlock()
	if ib == nil || ib.Protocol != "telegram" {
		stopMTProxyForInbound(inboundID)
		return
	}
	if err := startMTProxyForInbound(ctx, inboundID, ib); err != nil {
		log.Printf("[MTProxy %s] restart failed: %v", inboundID, err)
	}
}

func startAllMTProxies(ctx context.Context) {
	State.mu.RLock()
	list := []*Inbound{}
	for _, ib := range State.Inbounds {
		if ib.Protocol == "telegram" {
			list = append(list, ib)
		}
	}
	State.mu.RUnlock()
	for _, ib := range list {
		if err := startMTProxyForInbound(ctx, ib.InboundID, ib); err != nil {
			log.Printf("[MTProxy %s] startup failed: %v", ib.InboundID, err)
		}
	}
}

func stopAllMTProxies() {
	tgProxyMu.Lock()
	servers := make([]*MTProxyServer, 0, len(tgProxyRegistry))
	for _, s := range tgProxyRegistry {
		servers = append(servers, s)
	}
	tgProxyRegistry = map[string]*MTProxyServer{}
	tgProxyMu.Unlock()
	for _, s := range servers {
		s.Stop()
	}
}


func countActiveMTProxies() int {
	tgProxyMu.Lock()
	defer tgProxyMu.Unlock()
	count := 0
	for _, s := range tgProxyRegistry {
		if s.running {
			count++
		}
	}
	return count
}
