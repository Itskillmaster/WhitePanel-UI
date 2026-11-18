package main

// Runtime enforcement for user expiry and traffic quotas.
//
// The Go-based WS/XHTTP relays already call isActiveUser/checkAndUse on every
// message, so expired users are dropped immediately there. Connections that go
// through Xray (reality, shadowsocks, wireguard) are not visible to the Go
// relay code, so we periodically:
//
//  1. read per-user traffic counters from Xray's local stats API and add them
//     to each user's TrafficUsedBytes, and
//  2. detect when the set of active users changes (expired / over quota) and
//     re-apply the Xray config so those users' credentials are removed and
//     their live connections are terminated.
//
// Xray stats are only enabled when at least one managed inbound exists (see
// generateXrayConfig), and the API listens on loopback only.

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	xrayStatsPort   = 10085
	xrayStatsServer = "127.0.0.1:10085"
)

// xrayUserTraffic is the per-user traffic read from Xray. Uplink and downlink
// are kept apart so the panel can show real numbers instead of reporting every
// byte as a download.
type xrayUserTraffic struct {
	Up   int64
	Down int64
}

func (t xrayUserTraffic) Total() int64 { return t.Up + t.Down }

var (
	xrayStatMarker = ">>>traffic>>>"
	xrayNameRe     = regexp.MustCompile(`name"?\s*[:=]\s*"([^"]+)"`)
	xrayValueRe    = regexp.MustCompile(`value"?\s*[:=]\s*"?(\d+)"?`)
)

// queryXrayStats runs the bundled xray binary in API-client mode and returns
// its raw output. It returns an error when Xray is not running or the binary
// does not support the api subcommand.
func queryXrayStats(reset bool) (string, error) {
	args := []string{
		"api", "statsquery",
		"-server=" + xrayStatsServer,
		"-pattern=user>>>",
	}
	if reset {
		args = append(args, "-reset")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, xrayBinPath(), args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// addXrayStat folds one statsquery record into the result map. Records look
// like user>>>{uuid}>>>traffic>>>{uplink|downlink}.
func addXrayStat(out map[string]xrayUserTraffic, name, value string) {
	val, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil || val <= 0 {
		return
	}
	idx := strings.Index(name, xrayStatMarker)
	if idx <= 0 || !strings.HasPrefix(name, "user>>>") {
		return
	}
	uuid := strings.TrimPrefix(name[:idx], "user>>>")
	if uuid == "" {
		return
	}
	rec := out[uuid]
	switch {
	case strings.HasSuffix(name, "uplink"):
		rec.Up += val
	case strings.HasSuffix(name, "downlink"):
		rec.Down += val
	default:
		return
	}
	out[uuid] = rec
}

// parseXrayStatDeltas converts statsquery output into UUID -> traffic.
// Both the JSON shape and the protobuf-text shape are accepted because the
// output format depends on the bundled Xray version.
func parseXrayStatDeltas(raw string) map[string]xrayUserTraffic {
	out := map[string]xrayUserTraffic{}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return out
	}

	// JSON: {"stat":[{"name":"user>>>...","value":"123"}]}
	var asJSON struct {
		Stat []struct {
			Name  string `json:"name"`
			Value any    `json:"value"`
		} `json:"stat"`
	}
	if err := json.Unmarshal([]byte(raw), &asJSON); err == nil && len(asJSON.Stat) > 0 {
		for _, s := range asJSON.Stat {
			addXrayStat(out, s.Name, strings.TrimSpace(strings.Trim(fmtAny(s.Value), `"`)))
		}
		return out
	}

	// Protobuf text:
	//   stat: <
	//     name: "user>>>uuid>>>traffic>>>uplink"
	//     value: 1234
	//   >
	pendingName := ""
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if m := xrayNameRe.FindStringSubmatch(line); m != nil {
			pendingName = m[1]
			// Some builds place both fields on the same line.
			if v := xrayValueRe.FindStringSubmatch(line); v != nil {
				addXrayStat(out, pendingName, v[1])
				pendingName = ""
			}
			continue
		}
		if m := xrayValueRe.FindStringSubmatch(line); m != nil && pendingName != "" {
			addXrayStat(out, pendingName, m[1])
			pendingName = ""
		}
	}
	return out
}

// fmtAny renders a JSON value (string or number) as plain text.
func fmtAny(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case float64:
		return strconv.FormatInt(int64(t), 10)
	case json.Number:
		return t.String()
	default:
		return ""
	}
}

// applyXrayTraffic reads and resets the Xray user counters, then adds the
// deltas to the matching panel users. Returns true when state changed.
func applyXrayTraffic() bool {
	if !xrayIsRunning() {
		return false
	}
	raw, err := queryXrayStats(true)
	if err != nil {
		return false
	}
	deltas := parseXrayStatDeltas(raw)
	if len(deltas) == 0 {
		return false
	}
	State.mu.Lock()
	changed := false
	for _, u := range State.Users {
		d, ok := deltas[u.ConfigUUID]
		if !ok || d.Total() <= 0 {
			continue
		}
		u.TrafficUsedBytes += d.Total()
		u.TrafficUploadBytes += d.Up
		u.TrafficDownloadBytes += d.Down
		changed = true
	}
	State.mu.Unlock()
	if changed {
		_ = saveState()
	}
	return changed
}

// flushXrayTraffic persists whatever Xray has counted so far. It is called
// before the core is stopped or restarted: Xray keeps its counters in memory,
// so a restart without a final read would silently lose everything since the
// previous tick and make the reported usage too low.
func flushXrayTraffic() {
	if !xrayIsRunning() {
		return
	}
	applyXrayTraffic()
}

// activeSetHash fingerprints the set of users that are currently allowed to
// connect. Any change (expiry, quota, enable/disable, deletion) causes Xray to
// be reloaded.
func activeSetHash() string {
	State.mu.RLock()
	ids := make([]string, 0, len(State.Users))
	for _, u := range State.Users {
		if isActiveUser(u) {
			ids = append(ids, u.ConfigUUID)
		}
	}
	State.mu.RUnlock()
	sort.Strings(ids)
	sum := sha256.Sum256([]byte(strings.Join(ids, "\n")))
	return hex.EncodeToString(sum[:8])
}

func enforcementLoop(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	lastHash := ""
	first := true
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			applyXrayTraffic()
			if trafficDirty.Swap(false) {
				_ = saveState()
			}
			if countXrayInbounds() > 0 && !xrayIsRunning() {
				log.Printf("[Enforce] Xray is offline with active inbounds — attempting auto-recovery")
				xrayApply()
			}
			h := activeSetHash()
			if first {
				lastHash = h
				first = false
				continue
			}
			if h != lastHash {
				lastHash = h
				log.Printf("[Enforce] active user set changed — re-applying Xray config")
				scheduleXrayApply()
			}
		}
	}
}
