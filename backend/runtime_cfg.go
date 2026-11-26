package main

// Runtime deployment helpers.
//
// The panel is frequently run on public cloud providers where automated
// probes fetch the root path, probe health endpoints and look at process
// names and log output. Everything in this file keeps that surface neutral:
// silent logging, configurable URL prefixes, optional core services and a
// harmless default page.

import (
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

// initRuntime silences the standard logger unless DEBUG=1 is set. Debug output
// is useful locally but would otherwise expose the service identity in the
// platform log stream.
func initRuntime() {
	if os.Getenv("DEBUG") != "1" {
		log.SetOutput(io.Discard)
	}
}

// envPath reads a URL prefix from the environment. The returned value always
// starts with "/" and never ends with one, so it can be concatenated directly
// with "/segment".
func envPath(name, fallback string) string {
	v := strings.TrimSpace(os.Getenv(name))
	if v == "" {
		v = fallback
	}
	if !strings.HasPrefix(v, "/") {
		v = "/" + v
	}
	v = strings.TrimRight(v, "/")
	if v == "" {
		return fallback
	}
	return v
}

func adminPathPrefix() string  { return envPath("ADMIN_PATH", "/portal-8x4k2") }
func streamPathPrefix() string { return envPath("WS_PATH", "/events") }
func muxPathPrefix() string    { return envPath("XHTTP_PATH", "/cdn-v1") }
func subPathPrefix() string    { return envPath("SUB_PATH", "/feed") }

// coreEnabled reports whether the relay engine may be started. It is off by
// default so a plain panel deployment does not download or run a core binary.
func coreEnabled() bool { return os.Getenv("ENABLE_CORE") == "1" }

// bridgeEnabled reports whether Telegram bridge instances may be started.
func bridgeEnabled() bool { return os.Getenv("ENABLE_MTPROXY") == "1" }

// subscriptionPath returns the public subscription path for a user.
func subscriptionPath(id string) string {
	return subPathPrefix() + "/" + id
}

// clientStreamPath resolves the public stream path for a user. Values that were
// stored by older versions (for example "/stream/{uuid}") are rewritten to the
// configured stream prefix so existing inbounds follow WS_PATH as well.
func clientStreamPath(configured, uuid string) string {
	p := strings.TrimSpace(configured)
	p = strings.ReplaceAll(p, "{uuid}", uuid)
	switch {
	case p == "" || p == "/":
		return streamPathPrefix() + "/" + uuid
	case strings.HasPrefix(p, "/ws/"):
		return streamPathPrefix() + "/" + strings.TrimPrefix(p, "/ws/")
	case !strings.HasPrefix(p, "/"):
		return "/" + p
	default:
		return p
	}
}

// clientmuxPath returns the public mux path for a user and mode.
func clientMuxPath(mode, uuid string) string {
	if mode == "" || mode == "auto" {
		mode = "stream-up"
	}
	return muxPathPrefix() + "/" + mode + "/" + uuid
}

// serveStatusPage serves a neutral status page for the public root path. It must
// not reference the panel, relay software or any of the panel's services.
func serveStatusPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, statusPageHTML)
}

const statusPageHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<meta name="robots" content="noindex, nofollow">
<title>Edge Telemetry Gateway — All Systems Operational</title>
<style>
  :root { color-scheme: light dark; }
  * { box-sizing: border-box; }
  body {
    margin: 0; min-height: 100vh; display: flex; align-items: center; justify-content: center;
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
    background: #0f172a; color: #e2e8f0; padding: 24px;
  }
  .card {
    width: 100%; max-width: 680px; background: #111c33; border: 1px solid #1e293b;
    border-radius: 14px; padding: 32px 32px 24px; box-shadow: 0 20px 50px rgba(0,0,0,.35);
  }
  .brand { display: flex; align-items: center; gap: 12px; margin-bottom: 6px; }
  .dot { width: 10px; height: 10px; border-radius: 50%; background: #22c55e; box-shadow: 0 0 0 4px rgba(34,197,94,.15); }
  h1 { font-size: 20px; font-weight: 600; margin: 0; letter-spacing: .2px; }
  .sub { margin: 6px 0 26px; color: #94a3b8; font-size: 14px; }
  .banner {
    display: flex; align-items: center; justify-content: space-between; gap: 16px;
    background: rgba(34,197,94,.08); border: 1px solid rgba(34,197,94,.25);
    border-radius: 10px; padding: 14px 16px; margin-bottom: 24px;
  }
  .banner strong { color: #4ade80; font-size: 14px; font-weight: 600; }
  .banner span { color: #86efac; font-size: 12px; }
  ul { list-style: none; margin: 0; padding: 0; border-top: 1px solid #1e293b; }
  li {
    display: flex; align-items: center; justify-content: space-between;
    padding: 14px 2px; border-bottom: 1px solid #1e293b; font-size: 14px;
  }
  .name { color: #cbd5e1; }
  .state { display: inline-flex; align-items: center; gap: 7px; color: #4ade80; font-size: 12px; }
  .state i { width: 7px; height: 7px; border-radius: 50%; background: #22c55e; }
  footer { margin-top: 22px; display: flex; justify-content: space-between; color: #64748b; font-size: 12px; }
  @media (max-width: 480px) {
    .card { padding: 24px 20px 18px; }
    .banner { flex-direction: column; align-items: flex-start; gap: 4px; }
  }
</style>
</head>
<body>
  <main class="card">
    <div class="brand"><span class="dot"></span><h1>Edge Telemetry Gateway</h1></div>
    <p class="sub">Distributed edge platform &middot; status overview</p>

    <div class="banner">
      <strong>All Systems Operational</strong>
      <span id="checked">Last checked: just now</span>
    </div>

    <ul>
      <li><span class="name">API Gateway</span><span class="state"><i></i>Operational</span></li>
      <li><span class="name">Edge Network</span><span class="state"><i></i>Operational</span></li>
      <li><span class="name">Data Pipeline</span><span class="state"><i></i>Operational</span></li>
      <li><span class="name">Control Plane</span><span class="state"><i></i>Operational</span></li>
    </ul>

    <footer><span>Edge Telemetry Gateway</span><span>v3</span></footer>
  </main>
  <script>
    (function () {
      var el = document.getElementById('checked');
      if (el) el.textContent = 'Last checked: ' + new Date().toISOString().replace('T', ' ').slice(0, 19) + ' UTC';
    })();
  </script>
</body>
</html>
`
