package main

// Runtime enforcement for user expiry and traffic quotas.
//
// The Go-based stream/mux relays already call isActiveUser/checkAndUse on every
// message, so expired users are dropped immediately there. Connections that go
// through the engine (profile, cipher, curve) are not visible to the Go relay
// code, so we periodically:
//
//  1. read per-user traffic counters from the engine's local stats API and add
//     them to each user's TrafficUsedBytes, and
//  2. detect when the set of active users changes (expired / over quota) and
//     re-apply the engine config so those users' credentials are removed and
//     their live connections are terminated.
//
// Engine stats are only enabled when at least one managed inbound exists (see
// buildEngineConfig), and the API listens on loopback only.

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
	"sync"
	"time"
)

const (
	engineStatsPort   = 10085
	engineStatsServer = "127.0.0.1:10085"
)

// engineUserTraffic is the per-user traffic read from engine. Uplink and downlink
// are kept apart so the panel can show real numbers instead of reporting every
// byte as a download.
type engineUserTraffic struct {
	Up   int64
	Down int64
}

func (t engineUserTraffic) Total() int64 { return t.Up + t.Down }

var (
	engineStatMarker = ">>>traffic>>>"
	engineNameRe     = regexp.MustCompile(`name"?\s*[:=]\s*"([^"]+)"`)
	engineValueRe    = regexp.MustCompile(`value"?\s*[:=]\s*"?(\d+)"?`)

	// engineTrafficMu serialises statsquery calls. The enforcement loop and the
	// "apply/restart engine" path both read the counters with -reset; letting two
	// calls overlap can make one of them observe an already-reset (zero) set
	// and silently drop the bytes counted since the previous tick.
	engineTrafficMu sync.Mutex

	// lastengineStatsLog rate-limits diagnostics so a permanently broken stats
	// API does not flood the panel log every five seconds.
	lastEngineStatsLog time.Time
)

// queryengineStats runs the bundled engine binary in API-client mode and returns
// its raw output. It returns an error when engine is not running or the binary
// does not support the api subcommand.
func queryEngineMetrics(reset bool) (string, error) {
	args := []string{
		"api", "statsquery",
		"-server=" + engineStatsServer,
		"-pattern=user>>>",
	}
	if reset {
		args = append(args, "-reset")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, engineBinPath(), args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// addengineStat folds one statsquery record into the result map. Records look
// like user>>>{uuid}>>>traffic>>>{uplink|downlink}.
func addEngineMetric(out map[string]engineUserTraffic, name, value string) {
	val, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil || val <= 0 {
		return
	}
	idx := strings.Index(name, engineStatMarker)
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

// parseengineStatDeltas converts statsquery output into UUID -> traffic.
// Both the JSON shape and the protobuf-text shape are accepted because the
// output format depends on the bundled engine version.
func parseEngineMetricDeltas(raw string) map[string]engineUserTraffic {
	out := map[string]engineUserTraffic{}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return out
	}

	// JSON: {"stat":[{"name":"user>>>...","value":"123"}]}
	// engine's CLI may print a warning/log line before or after the payload, so
	// locate the JSON object instead of requiring the whole output to be valid
	// JSON. Losing this payload would lose traffic: the counters have already
	// been reset on the engine side when -reset is used.
	if start := strings.IndexByte(raw, '{'); start >= 0 {
		if end := strings.LastIndexByte(raw, '}'); end > start {
			var asJSON struct {
				Stat []struct {
					Name  string `json:"name"`
					Value any    `json:"value"`
				} `json:"stat"`
			}
			if err := json.Unmarshal([]byte(raw[start:end+1]), &asJSON); err == nil && len(asJSON.Stat) > 0 {
				for _, s := range asJSON.Stat {
					addEngineMetric(out, s.Name, strings.TrimSpace(strings.Trim(fmtAny(s.Value), `"`)))
				}
				return out
			}
		}
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
		if m := engineNameRe.FindStringSubmatch(line); m != nil {
			pendingName = m[1]
			// Some builds place both fields on the same line.
			if v := engineValueRe.FindStringSubmatch(line); v != nil {
				addEngineMetric(out, pendingName, v[1])
				pendingName = ""
			}
			continue
		}
		if m := engineValueRe.FindStringSubmatch(line); m != nil && pendingName != "" {
			addEngineMetric(out, pendingName, m[1])
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

// applyengineTraffic reads and resets the engine user counters, then adds the
// deltas to the matching panel users. Returns true when state changed.
func applyEngineMetrics() bool {
	engineTrafficMu.Lock()
	defer engineTrafficMu.Unlock()

	if !engineIsRunning() {
		return false
	}
	raw, err := queryEngineMetrics(true)

	// The server resets the counters before the CLI writes its response. If the
	// process exits non-zero after printing a valid payload (for example when a
	// warning is mixed into the output), the payload must still be accounted;
	// throwing it away would under-report usage permanently.
	deltas := parseEngineMetricDeltas(raw)
	if len(deltas) == 0 {
		if err != nil {
			// Rate-limit so a permanently broken stats API does not flood the
			// log every enforcement tick.
			if time.Since(lastEngineStatsLog) > time.Minute {
				lastEngineStatsLog = time.Now()
				log.Printf("[Sync] engine statsquery failed: %v", err)
			}
		}
		return false
	}
	if err != nil {
		log.Printf("[Sync] engine statsquery returned data with an error (%v); accounting the payload anyway", err)
	}

	changed := applyTrafficDeltas(deltas)
	if changed {
		_ = saveState()
	}
	return changed
}

// applyTrafficDeltas adds the per-UUID deltas to the matching users. It is kept
// separate from the engine I/O so it can be unit-tested and so a UUID that is
// shared by more than one record can never be counted twice.
func applyTrafficDeltas(deltas map[string]engineUserTraffic) bool {
	State.mu.Lock()
	defer State.mu.Unlock()

	changed := false
	seen := make(map[string]bool, len(deltas))
	for _, u := range State.Users {
		if u == nil || seen[u.ConfigUUID] {
			continue
		}
		d, ok := deltas[u.ConfigUUID]
		if !ok || d.Total() <= 0 {
			continue
		}
		seen[u.ConfigUUID] = true
		u.TrafficUsedBytes += d.Total()
		u.TrafficUploadBytes += d.Up
		u.TrafficDownloadBytes += d.Down
		changed = true
	}
	return changed
}

// flushengineTraffic persists whatever engine has counted so far. It is called
// before the core is stopped or restarted: engine keeps its counters in memory,
// so a restart without a final read would silently lose everything since the
// previous tick and make the reported usage too low.
func flushEngineMetrics() {
	if !engineIsRunning() {
		return
	}
	applyEngineMetrics()
}

// activeSetHash fingerprints the set of users that are currently allowed to
// connect. Any change (expiry, quota, enable/disable, deletion) causes engine to
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
			applyEngineMetrics()
			if trafficDirty.Swap(false) {
				_ = saveState()
			}
			if coreEnabled() && countEngineInbounds() > 0 && !engineIsRunning() {
				log.Printf("[Sync] engine is offline with active inbounds - attempting auto-recovery")
				engineApply()
			}
			h := activeSetHash()
			if first {
				lastHash = h
				first = false
				continue
			}
			if h != lastHash {
				lastHash = h
				log.Printf("[Sync] active user set changed - re-applying engine config")
				scheduleEngineApply()
			}
		}
	}
}
