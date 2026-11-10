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

// Xray's `api statsquery` prints protobuf text format (or JSON when -json is
// supported). These tolerantly match both shapes.
var (
	xrayStatBlockRe = regexp.MustCompile(`name"?\s*[:=]\s*"([^"]+)"[\s\S]*?value"?\s*[:=]\s*"?(\d+)"?`)
	xrayStatMarker  = ">>>traffic>>>"
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

// parseXrayStatDeltas converts statsquery output into UUID -> total bytes.
// Uplink and downlink are summed because quota limits are bidirectional.
func parseXrayStatDeltas(raw string) map[string]int64 {
	deltas := map[string]int64{}
	for _, m := range xrayStatBlockRe.FindAllStringSubmatch(raw, -1) {
		if len(m) < 3 {
			continue
		}
		name := m[1]
		val, err := strconv.ParseInt(m[2], 10, 64)
		if err != nil || val <= 0 {
			continue
		}
		idx := strings.Index(name, xrayStatMarker)
		if idx <= 0 || !strings.HasPrefix(name, "user>>>") {
			continue
		}
		uuid := strings.TrimPrefix(name[:idx], "user>>>")
		if uuid == "" {
			continue
		}
		deltas[uuid] += val
	}
	return deltas
}

// applyXrayTraffic reads and resets the Xray user counters, then adds the
// deltas to the matching panel users. Returns true when state changed.
func applyXrayTraffic() bool {
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
		for k, add := range deltas {
			if strings.EqualFold(u.ConfigUUID, k) && add > 0 {
				u.TrafficUsedBytes += add
				changed = true
				break
			}
		}
	}
	State.mu.Unlock()
	if changed {
		_ = saveState()
	}
	return changed
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
	ticker := time.NewTicker(30 * time.Second)
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
