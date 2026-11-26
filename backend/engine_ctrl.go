package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ============================================================================
// ENGINE CONTROLLER & HTTP API
// ============================================================================

var (
	engineStartedAt       time.Time
	cachedEngineVersion   string
	cachedEngineVersionAt time.Time
	cachedEngineVersionMu sync.RWMutex
)

func invalidateEngineCache() {
	cachedEngineVersionMu.Lock()
	cachedEngineVersion = ""
	cachedEngineVersionAt = time.Time{}
	cachedEngineVersionMu.Unlock()
}

func getEngineVersion() string {
	cachedEngineVersionMu.RLock()
	if cachedEngineVersion != "" && time.Since(cachedEngineVersionAt) < 5*time.Minute {
		ver := cachedEngineVersion
		cachedEngineVersionMu.RUnlock()
		return ver
	}
	cachedEngineVersionMu.RUnlock()

	bin := engineBinPath()
	if st, err := os.Stat(bin); err != nil || st.Size() < 10000 {
		return "نصب نشده (روی دکمه آپدیت هسته کلیک کنید)"
	}

	ver := "engine-core"
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, "version")
	cmd.Stdin = nil
	out, err := cmd.Output()
	if err == nil {
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		if len(lines) > 0 && lines[0] != "" {
			ver = strings.TrimSpace(lines[0])
		}
	} else {
		ver = "نصب شده (آماده)"
	}

	cachedEngineVersionMu.Lock()
	cachedEngineVersion = ver
	cachedEngineVersionAt = time.Now()
	cachedEngineVersionMu.Unlock()

	return ver
}

func readEngineRecentLogs(maxLines int) []string {
	logFile := engineLogPath()
	f, err := os.Open(logFile)
	if err != nil {
		return []string{"هنوز لاگی برای هسته ثبت نشده است."}
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil || fi.Size() == 0 {
		return []string{"[Engine] لاگ هسته خالی است."}
	}

	const maxRead = 64 * 1024
	offset := int64(0)
	if fi.Size() > maxRead {
		offset = fi.Size() - maxRead
	}
	buf := make([]byte, fi.Size()-offset)
	_, _ = f.ReadAt(buf, offset)

	lines := strings.Split(string(buf), "\n")
	cleaned := make([]string, 0, len(lines))
	for _, l := range lines {
		l = strings.TrimRight(l, "\r")
		if l != "" {
			cleaned = append(cleaned, l)
		}
	}
	if len(cleaned) == 0 {
		return []string{"[Engine] لاگ هسته خالی است."}
	}
	if len(cleaned) > maxLines {
		cleaned = cleaned[len(cleaned)-maxLines:]
	}
	return cleaned
}

func countEngineInbounds() int {
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

func handleEngineStatus(w http.ResponseWriter, r *http.Request) {
	isRunning := engineIsRunning()
	version := getEngineVersion()
	inboundsCount := countEngineInbounds()

	uptimeSec := 0
	if isRunning && !engineStartedAt.IsZero() {
		uptimeSec = int(time.Since(engineStartedAt).Seconds())
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok": true,
		"status": map[string]any{
			"running":        isRunning,
			"status_text":    map[bool]string{true: "ONLINE", false: "OFFLINE"}[isRunning],
			"version":        version,
			"uptime_seconds": uptimeSec,
			"inbounds_count": inboundsCount,
			"os_arch":        fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
			"binary_path":    engineBinPath(),
			"config_path":    engineCfgPath(),
			"log_path":       engineLogPath(),
		},
	})
}

func handleEngineStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"ok": false})
		return
	}
	if engineIsRunning() {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "msg": "هسته در حال حاضر روشن است"})
		return
	}

	engineApply()
	engineStartedAt = time.Now()

	running := engineIsRunning()
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":      running,
		"running": running,
		"msg":     map[bool]string{true: "هسته با موفقیت روشن شد", false: "روشن کردن هسته ناموفق بود (لاگ را بررسی کنید)"}[running],
	})
}

func handleEngineStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"ok": false})
		return
	}
	engineStop()
	engineStartedAt = time.Time{}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "msg": "هسته متوقف شد"})
}

func handleEngineRestart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"ok": false})
		return
	}
	engineStop()
	time.Sleep(500 * time.Millisecond)
	engineApply()
	engineStartedAt = time.Now()

	running := engineIsRunning()
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":      running,
		"running": running,
		"msg":     map[bool]string{true: "هسته با موفقیت ریستارت شد", false: "شروع مجدد هسته ناموفق بود"}[running],
	})
}

func handleEngineUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"ok": false})
		return
	}

	// Stop engine first
	wasRunning := engineIsRunning()
	engineStop()

	// Remove current binary to force re-download
	bin := engineBinPath()
	if strings.HasPrefix(bin, engineDir()) {
		_ = os.Remove(bin)
	}
	invalidateEngineCache()

	// Trigger download of latest version
	if err := ensureEngine(); err != nil {
		if wasRunning {
			engineApply()
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"ok":  false,
			"msg": fmt.Sprintf("خطا در بروزرسانی هسته: %v", err),
		})
		return
	}

	// Restart if it was running or apply config
	engineApply()
	engineStartedAt = time.Now()

	newVersion := getEngineVersion()
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":      true,
		"version": newVersion,
		"msg":     fmt.Sprintf("هسته با موفقیت بروزرسانی و فعال شد: %s", newVersion),
		"running": engineIsRunning(),
	})
}

func handleEngineLogs(w http.ResponseWriter, r *http.Request) {
	limit := 100
	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if n, err := strconv.Atoi(lStr); err == nil && n > 0 && n <= 500 {
			limit = n
		}
	}
	logs := readEngineRecentLogs(limit)
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":   true,
		"logs": logs,
	})
}

func handleEngineLogsClear(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"ok": false})
		return
	}
	logFile := engineLogPath()
	_ = os.Truncate(logFile, 0)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "msg": "لاگ‌های هسته پاک شدند"})
}
