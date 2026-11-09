package main

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// ============================================================================
// XRAY ENGINE CONTROLLER & HTTP API
// ============================================================================

var xrayStartedAt time.Time

func getXrayVersion() string {
	bin := xrayBinPath()
	if st, err := os.Stat(bin); err != nil || st.Size() < 10000 {
		return "نصب نشده (Not Installed)"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, "version")
	out, err := cmd.Output()
	if err != nil {
		return "نصب شده (آماده)"
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) > 0 && lines[0] != "" {
		return strings.TrimSpace(lines[0])
	}
	return "Xray-core"
}

func readXrayRecentLogs(maxLines int) []string {
	logFile := xrayLogPath()
	f, err := os.Open(logFile)
	if err != nil {
		return []string{"هنوز لاگی برای هسته Xray ثبت نشده است."}
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if len(lines) == 0 {
		return []string{"[Xray] لاگ هسته خالی است."}
	}

	if len(lines) > maxLines {
		lines = lines[len(lines)-maxLines:]
	}
	return lines
}

func countXrayInbounds() int {
	State.mu.RLock()
	defer State.mu.RUnlock()
	count := 0
	for _, ib := range State.Inbounds {
		if ib.System || ib.Port <= 0 {
			continue
		}
		proto := strings.ToLower(ib.Protocol)
		sec := strings.ToLower(ib.Security)
		if proto == "reality" || sec == "reality" || proto == "shadowsocks" || proto == "wireguard" {
			count++
		}
	}
	return count
}

func handleXrayStatus(w http.ResponseWriter, r *http.Request) {
	isRunning := xrayIsRunning()
	version := getXrayVersion()
	inboundsCount := countXrayInbounds()

	uptimeSec := 0
	if isRunning && !xrayStartedAt.IsZero() {
		uptimeSec = int(time.Since(xrayStartedAt).Seconds())
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":             true,
		"status":         map[string]any{
			"running":        isRunning,
			"status_text":    map[bool]string{true: "ONLINE", false: "OFFLINE"}[isRunning],
			"version":        version,
			"uptime_seconds": uptimeSec,
			"inbounds_count": inboundsCount,
			"os_arch":        fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
			"binary_path":    xrayBinPath(),
			"config_path":    xrayCfgPath(),
			"log_path":       xrayLogPath(),
		},
	})
}

func handleXrayStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"ok": false})
		return
	}
	if xrayIsRunning() {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "msg": "هسته Xray در حال حاضر روشن است"})
		return
	}

	xrayApply()
	xrayStartedAt = time.Now()

	running := xrayIsRunning()
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":      running,
		"running": running,
		"msg":     map[bool]string{true: "هسته Xray با موفقیت روشن شد", false: "روشن کردن Xray ناموفق بود (لاگ را بررسی کنید)"}[running],
	})
}

func handleXrayStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"ok": false})
		return
	}
	xrayStop()
	xrayStartedAt = time.Time{}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "msg": "هسته Xray متوقف شد"})
}

func handleXrayRestart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"ok": false})
		return
	}
	xrayStop()
	time.Sleep(500 * time.Millisecond)
	xrayApply()
	xrayStartedAt = time.Now()

	running := xrayIsRunning()
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":      running,
		"running": running,
		"msg":     map[bool]string{true: "هسته Xray با موفقیت ریستارت شد", false: "شروع مجدد Xray ناموفق بود"}[running],
	})
}

func handleXrayUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"ok": false})
		return
	}

	// Stop Xray first
	wasRunning := xrayIsRunning()
	xrayStop()

	// Remove current binary to force re-download
	bin := xrayBinPath()
	_ = os.Remove(bin)

	// Trigger download of latest version
	if err := ensureXray(); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"ok":     false,
			"detail": fmt.Sprintf("خطا در بروزرسانی Xray: %v", err),
		})
		return
	}

	// Restart if it was running or apply config
	if wasRunning {
		xrayApply()
		xrayStartedAt = time.Now()
	}

	newVersion := getXrayVersion()
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":      true,
		"version": newVersion,
		"msg":     fmt.Sprintf("هسته Xray با موفقیت بروزرسانی شد: %s", newVersion),
	})
}

func handleXrayLogs(w http.ResponseWriter, r *http.Request) {
	limit := 100
	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if n, err := strconv.Atoi(lStr); err == nil && n > 0 && n <= 500 {
			limit = n
		}
	}
	logs := readXrayRecentLogs(limit)
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":   true,
		"logs": logs,
	})
}

func handleXrayLogsClear(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"ok": false})
		return
	}
	logFile := xrayLogPath()
	_ = os.Truncate(logFile, 0)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "msg": "لاگ‌های Xray پاک شدند"})
}
