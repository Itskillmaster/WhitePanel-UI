package main

// Persistent storage for the scanner lists (scanned IPs and SNI results).
// Everything is stored under the panel data directory so it survives restarts
// and Docker redeploys with a mounted volume.

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
)

var (
	scannerMu  sync.Mutex
	scannerSeq = map[string]int{}
)

var scannerKindRe = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,32}$`)

func scannerKind(raw string) string {
	k := strings.ToLower(strings.TrimSpace(raw))
	if k == "rw" {
		k = "railway"
	}
	if !scannerKindRe.MatchString(k) {
		return ""
	}
	return k
}

func scannerIPFile(kind string) string {
	return filepath.Join(State.DataDir, "scanned_"+kind+".txt")
}

func scannerSNIFile() string {
	return filepath.Join(State.DataDir, "sni_reality.txt")
}

func scannerSNIInputFile() string {
	return filepath.Join(State.DataDir, "sni_reality_for_scan.txt")
}

// seedScannerFiles copies the bundled default lists into the data directory the
// first time the panel starts. Existing files are never overwritten, so user
// scan results survive restarts and redeploys.
func seedScannerFiles() {
	seedDir := "./seed"
	if st, err := os.Stat(seedDir); err != nil || !st.IsDir() {
		return
	}
	files := []string{
		"cf_subnets.txt",
		"sni-list.txt",
		"sni_reality_for_scan.txt",
		"sni_reality.txt",
		"endpoint.txt",
	}
	for _, name := range files {
		dst := filepath.Join(State.DataDir, name)
		if _, err := os.Stat(dst); err == nil {
			continue
		}
		b, err := os.ReadFile(filepath.Join(seedDir, name))
		if err != nil {
			continue
		}
		_ = os.MkdirAll(State.DataDir, 0755)
		_ = os.WriteFile(dst, b, 0644)
	}
}

func readLineFile(path string) []string {
	b, err := os.ReadFile(path)
	if err != nil {
		return []string{}
	}
	out := []string{}
	seen := map[string]bool{}
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(strings.TrimSuffix(line, "\r"))
		if line == "" || strings.HasPrefix(line, "#") || seen[line] {
			continue
		}
		seen[line] = true
		out = append(out, line)
	}
	return out
}

func writeLineFile(path string, lines []string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	body := strings.Join(lines, "\n")
	if body != "" {
		body += "\n"
	}
	if err := os.WriteFile(tmp, []byte(body), 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func cleanScannerIPs(entries []string) []string {
	out := []string{}
	seen := map[string]bool{}
	for _, e := range entries {
		e = strings.TrimSpace(e)
		if e == "" || strings.ContainsAny(e, "\r\n\t ;") {
			continue
		}
		if seen[e] {
			continue
		}
		seen[e] = true
		out = append(out, e)
	}
	return out
}

func handleScannerSaveImpl(body map[string]any) map[string]any {
	kind := scannerKind(gs(body, "type"))
	if kind == "" {
		return map[string]any{"ok": false, "detail": "invalid scanner type", "ips": []string{}, "seq": 0}
	}
	incoming := cleanScannerIPs(gss(body, "ips"))
	replace := gb(body, "replace")
	seq := gi(body, "seq")

	scannerMu.Lock()
	defer scannerMu.Unlock()

	cur := scannerSeq[kind]
	if _, ok := body["seq"]; ok && seq < cur {
		// A write from an older scan generation: ignore it.
		return map[string]any{
			"ok": true, "stale": true, "type": kind,
			"ips": readLineFile(scannerIPFile(kind)), "seq": cur,
		}
	}

	lines := []string{}
	if !replace {
		lines = readLineFile(scannerIPFile(kind))
	}
	seen := map[string]bool{}
	for _, x := range lines {
		seen[x] = true
	}
	for _, x := range incoming {
		if !seen[x] {
			seen[x] = true
			lines = append(lines, x)
		}
	}
	// Latest scan results are latency-sorted on the client; keep the file in
	// the same order and cap it to a sane size.
	if len(lines) > 500 {
		lines = lines[:500]
	}
	if err := writeLineFile(scannerIPFile(kind), lines); err != nil {
		return map[string]any{"ok": false, "detail": err.Error(), "ips": lines, "seq": cur}
	}
	cur++
	scannerSeq[kind] = cur
	return map[string]any{"ok": true, "type": kind, "ips": lines, "seq": cur}
}

func handleScannerIPSImpl(kind string) map[string]any {
	kind = scannerKind(kind)
	if kind == "" {
		return map[string]any{"ok": false, "detail": "invalid scanner type", "ips": []string{}, "seq": 0}
	}
	scannerMu.Lock()
	seq := scannerSeq[kind]
	scannerMu.Unlock()
	return map[string]any{"ok": true, "type": kind, "ips": readLineFile(scannerIPFile(kind)), "seq": seq}
}

type sniResult struct {
	SNI       string `json:"sni"`
	LatencyMs int64  `json:"latency_ms"`
}

// sniResultsFromAny converts a decoded JSON array (from the scanner UI) into
// sniResult values. Both {"sni":...} and {"host":...} shapes are accepted.
func sniResultsFromAny(v any) []sniResult {
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]sniResult, 0, len(arr))
	for _, item := range arr {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		sni := strings.TrimSpace(gs(m, "sni"))
		if sni == "" {
			sni = strings.TrimSpace(gs(m, "host"))
		}
		if sni == "" {
			continue
		}
		out = append(out, sniResult{SNI: sni, LatencyMs: int64(gf(m, "latency_ms"))})
	}
	return out
}

func readSNIResults() []sniResult {
	lines := readLineFile(scannerSNIFile())
	out := []sniResult{}
	for _, line := range lines {
		parts := strings.SplitN(line, "|", 2)
		sni := strings.TrimSpace(parts[0])
		if sni == "" || strings.HasPrefix(sni, "#") {
			continue
		}
		var ms int64
		if len(parts) == 2 {
			ms, _ = strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
		}
		out = append(out, sniResult{SNI: sni, LatencyMs: ms})
	}
	sort.SliceStable(out, func(i, j int) bool {
		ai, aj := out[i].LatencyMs, out[j].LatencyMs
		if ai == 0 {
			ai = 1 << 62
		}
		if aj == 0 {
			aj = 1 << 62
		}
		return ai < aj
	})
	return out
}

func writeSNIResults(results []sniResult) error {
	lines := []string{"# Fastest SNIs"}
	for _, r := range results {
		lines = append(lines, fmt.Sprintf("%s|%d", r.SNI, r.LatencyMs))
	}
	return writeLineFile(scannerSNIFile(), lines)
}

func mergeSNIResults(results []sniResult, top int) []sniResult {
	if top <= 0 {
		top = 20
	}
	byName := map[string]sniResult{}
	for _, r := range readSNIResults() {
		byName[r.SNI] = r
	}
	for _, r := range results {
		r.SNI = strings.TrimSpace(strings.ToLower(r.SNI))
		if r.SNI == "" {
			continue
		}
		if old, ok := byName[r.SNI]; !ok || (r.LatencyMs > 0 && (old.LatencyMs == 0 || r.LatencyMs < old.LatencyMs)) {
			byName[r.SNI] = r
		}
	}
	out := make([]sniResult, 0, len(byName))
	for _, r := range byName {
		out = append(out, r)
	}
	sort.SliceStable(out, func(i, j int) bool {
		ai, aj := out[i].LatencyMs, out[j].LatencyMs
		if ai == 0 {
			ai = 1 << 62
		}
		if aj == 0 {
			aj = 1 << 62
		}
		return ai < aj
	})
	if len(out) > top {
		out = out[:top]
	}
	return out
}
