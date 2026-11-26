package main

// ============================================================================
// Edge Gateway - advanced cache layer
//
// This file contains the building blocks used by the HTTP layer:
//   * TTLCache        : sharded-free, LRU + TTL in-memory cache with stats
//   * GetOrLoad       : cache stampede protection (single-flight)
//   * gzip helpers    : pooled compressors for dynamic + static responses
//   * middleware      : compression, security headers, cache-control policy
//
// Everything here is dependency free (standard library only) so the Docker
// build stays offline friendly.
// ============================================================================

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"container/list"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ---------------------------------------------------------------------------
// Cache registry (used by /api/cache/stats)
// ---------------------------------------------------------------------------

type cacheSnapshot struct {
	Name      string `json:"name"`
	Items     int    `json:"items"`
	Bytes     int64  `json:"bytes"`
	Hits      int64  `json:"hits"`
	Misses    int64  `json:"misses"`
	Loads     int64  `json:"loads"`
	Evictions int64  `json:"evictions"`
	Expired   int64  `json:"expired"`
	MaxItems  int    `json:"max_items"`
	MaxBytes  int64  `json:"max_bytes"`
}

type cacheReporter interface {
	snapshot() cacheSnapshot
	sweep()
}

var (
	cacheRegistryMu sync.RWMutex
	cacheRegistry   []cacheReporter
)

func registerCache(c cacheReporter) {
	cacheRegistryMu.Lock()
	cacheRegistry = append(cacheRegistry, c)
	cacheRegistryMu.Unlock()
}

func allCacheSnapshots() []cacheSnapshot {
	cacheRegistryMu.RLock()
	defer cacheRegistryMu.RUnlock()
	out := make([]cacheSnapshot, 0, len(cacheRegistry))
	for _, c := range cacheRegistry {
		out = append(out, c.snapshot())
	}
	return out
}

func sweepAllCaches() {
	cacheRegistryMu.RLock()
	caches := append([]cacheReporter(nil), cacheRegistry...)
	cacheRegistryMu.RUnlock()
	for _, c := range caches {
		c.sweep()
	}
}

// startCacheJanitor periodically drops expired entries and keeps memory flat.
func startCacheJanitor(ctx context.Context, every time.Duration) {
	if every <= 0 {
		every = 5 * time.Minute
	}
	go func() {
		t := time.NewTicker(every)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				sweepAllCaches()
			}
		}
	}()
}

// ---------------------------------------------------------------------------
// TTLCache
// ---------------------------------------------------------------------------

type cacheItem[V any] struct {
	key     string
	val     V
	size    int64
	expires time.Time
	elem    *list.Element
}

type inflightLoad[V any] struct {
	done chan struct{}
	val  V
	err  error
}

// TTLCache is a bounded, LRU evicted, TTL based cache.
//
// It is safe for concurrent use. Expensive loads are de-duplicated: if N
// requests ask for the same key at the same time only one loader runs and all
// callers receive the same value (cache stampede protection).
type TTLCache[V any] struct {
	mu sync.Mutex

	name       string
	items      map[string]*cacheItem[V]
	lru        *list.List // front = most recently used
	inflight   map[string]*inflightLoad[V]
	maxItems   int
	maxBytes   int64
	bytes      int64
	defaultTTL time.Duration
	sizeOf     func(V) int64

	hits      int64
	misses    int64
	loads     int64
	evictions int64
	expired   int64
}

func NewTTLCache[V any](name string, maxItems int, maxBytes int64, defaultTTL time.Duration, sizeOf func(V) int64) *TTLCache[V] {
	if maxItems <= 0 {
		maxItems = 128
	}
	if defaultTTL <= 0 {
		defaultTTL = 5 * time.Minute
	}
	if sizeOf == nil {
		sizeOf = func(V) int64 { return 1 }
	}
	c := &TTLCache[V]{
		name:       name,
		items:      make(map[string]*cacheItem[V]),
		lru:        list.New(),
		inflight:   make(map[string]*inflightLoad[V]),
		maxItems:   maxItems,
		maxBytes:   maxBytes,
		defaultTTL: defaultTTL,
		sizeOf:     sizeOf,
	}
	registerCache(c)
	return c
}

func (c *TTLCache[V]) Name() string { return c.name }

// Get returns a cached value when it exists and has not expired.
func (c *TTLCache[V]) Get(key string) (V, bool) {
	var zero V
	now := time.Now()

	c.mu.Lock()
	defer c.mu.Unlock()

	it, ok := c.items[key]
	if !ok {
		c.misses++
		return zero, false
	}
	if now.After(it.expires) {
		c.removeLocked(it)
		c.expired++
		c.misses++
		return zero, false
	}
	c.lru.MoveToFront(it.elem)
	c.hits++
	return it.val, true
}

// Set stores a value. A ttl <= 0 uses the cache default.
func (c *TTLCache[V]) Set(key string, val V, ttl time.Duration) {
	if ttl <= 0 {
		ttl = c.defaultTTL
	}
	c.mu.Lock()
	c.setLocked(key, val, ttl)
	c.mu.Unlock()
}

// GetOrLoad returns a cached value or runs load once for concurrent callers.
// The second return value reports whether the value came from the cache.
func (c *TTLCache[V]) GetOrLoad(ctx context.Context, key string, ttl time.Duration, load func(context.Context) (V, error)) (V, bool, error) {
	var zero V
	if ttl <= 0 {
		ttl = c.defaultTTL
	}

	now := time.Now()
	c.mu.Lock()
	if it, ok := c.items[key]; ok {
		if now.Before(it.expires) {
			c.lru.MoveToFront(it.elem)
			c.hits++
			val := it.val
			c.mu.Unlock()
			return val, true, nil
		}
		c.removeLocked(it)
		c.expired++
	}
	if call, ok := c.inflight[key]; ok {
		c.mu.Unlock()
		select {
		case <-call.done:
			if call.err != nil {
				return zero, false, call.err
			}
			return call.val, true, nil
		case <-ctx.Done():
			return zero, false, ctx.Err()
		}
	}

	call := &inflightLoad[V]{done: make(chan struct{})}
	c.inflight[key] = call
	c.misses++
	c.mu.Unlock()

	val, err := load(ctx)

	c.mu.Lock()
	delete(c.inflight, key)
	if err == nil {
		c.setLocked(key, val, ttl)
		c.loads++
	}
	call.val = val
	call.err = err
	c.mu.Unlock()
	close(call.done)

	if err != nil {
		return zero, false, err
	}
	return val, false, nil
}

func (c *TTLCache[V]) setLocked(key string, val V, ttl time.Duration) {
	size := c.sizeOf(val)
	// Never store an item that can never fit; it would evict the whole cache.
	if c.maxBytes > 0 && size > c.maxBytes {
		return
	}
	if it, ok := c.items[key]; ok {
		c.bytes -= it.size
		it.val = val
		it.size = size
		it.expires = time.Now().Add(ttl)
		c.bytes += size
		c.lru.MoveToFront(it.elem)
	} else {
		it := &cacheItem[V]{key: key, val: val, size: size, expires: time.Now().Add(ttl)}
		it.elem = c.lru.PushFront(it)
		c.items[key] = it
		c.bytes += size
	}
	c.evictLocked()
}

func (c *TTLCache[V]) evictLocked() {
	for c.maxItems > 0 && len(c.items) > c.maxItems {
		if !c.evictOneLocked() {
			return
		}
	}
	for c.maxBytes > 0 && c.bytes > c.maxBytes && len(c.items) > 0 {
		if !c.evictOneLocked() {
			return
		}
	}
}

func (c *TTLCache[V]) evictOneLocked() bool {
	back := c.lru.Back()
	if back == nil {
		return false
	}
	it, ok := back.Value.(*cacheItem[V])
	if !ok {
		return false
	}
	c.removeLocked(it)
	c.evictions++
	return true
}

func (c *TTLCache[V]) removeLocked(it *cacheItem[V]) {
	delete(c.items, it.key)
	c.lru.Remove(it.elem)
	c.bytes -= it.size
	if c.bytes < 0 {
		c.bytes = 0
	}
}

// Delete removes a single key.
func (c *TTLCache[V]) Delete(key string) {
	c.mu.Lock()
	if it, ok := c.items[key]; ok {
		c.removeLocked(it)
	}
	c.mu.Unlock()
}

// Purge drops every entry (stats are preserved).
func (c *TTLCache[V]) Purge() int {
	c.mu.Lock()
	n := len(c.items)
	c.items = make(map[string]*cacheItem[V])
	c.lru.Init()
	c.bytes = 0
	c.mu.Unlock()
	return n
}

// Len returns the number of live entries.
func (c *TTLCache[V]) Len() int {
	c.mu.Lock()
	n := len(c.items)
	c.mu.Unlock()
	return n
}

func (c *TTLCache[V]) snapshot() cacheSnapshot {
	c.mu.Lock()
	defer c.mu.Unlock()
	return cacheSnapshot{
		Name:      c.name,
		Items:     len(c.items),
		Bytes:     c.bytes,
		Hits:      c.hits,
		Misses:    c.misses,
		Loads:     c.loads,
		Evictions: c.evictions,
		Expired:   c.expired,
		MaxItems:  c.maxItems,
		MaxBytes:  c.maxBytes,
	}
}

func (c *TTLCache[V]) sweep() {
	now := time.Now()
	c.mu.Lock()
	for _, it := range c.items {
		if now.After(it.expires) {
			c.removeLocked(it)
			c.expired++
		}
	}
	c.mu.Unlock()
}

// ---------------------------------------------------------------------------
// gzip helpers
// ---------------------------------------------------------------------------

var gzipPool = sync.Pool{
	New: func() any {
		w, _ := gzip.NewWriterLevel(io.Discard, gzip.BestSpeed)
		return w
	},
}

// gzipBytes compresses src. level == gzip.BestSpeed uses the shared pool which
// is the right choice for per-request compression; a higher level is used for
// static assets that are compressed once and cached.
func gzipBytes(src []byte, level int) ([]byte, error) {
	var buf bytes.Buffer
	buf.Grow(len(src) / 3)

	if level <= gzip.BestSpeed {
		w := gzipPool.Get().(*gzip.Writer)
		w.Reset(&buf)
		_, err := w.Write(src)
		if err == nil {
			err = w.Close()
		}
		gzipPool.Put(w)
		if err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	}

	w, err := gzip.NewWriterLevel(&buf, level)
	if err != nil {
		return nil, err
	}
	if _, err := w.Write(src); err != nil {
		_ = w.Close()
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// acceptsEncoding implements a small, correct Accept-Encoding parser
// (q-values included) for gzip/br negotiation.
func acceptsEncoding(r *http.Request, enc string) bool {
	header := r.Header.Get("Accept-Encoding")
	if header == "" {
		return false
	}
	enc = strings.ToLower(enc)
	for _, part := range strings.Split(header, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		fields := strings.Split(part, ";")
		name := strings.ToLower(strings.TrimSpace(fields[0]))
		if name != enc && name != "*" {
			continue
		}
		q := 1.0
		for _, f := range fields[1:] {
			f = strings.TrimSpace(f)
			if strings.HasPrefix(f, "q=") {
				if v, err := strconv.ParseFloat(strings.TrimPrefix(f, "q="), 64); err == nil {
					q = v
				}
			}
		}
		return q > 0
	}
	return false
}

// compressibleType reports whether a Content-Type benefits from gzip.
func compressibleType(ctype string) bool {
	if i := strings.IndexByte(ctype, ';'); i >= 0 {
		ctype = ctype[:i]
	}
	ctype = strings.ToLower(strings.TrimSpace(ctype))
	if ctype == "" {
		return false // let net/http sniff it instead of guessing wrong
	}
	if strings.HasPrefix(ctype, "text/") {
		return true
	}
	switch ctype {
	case "application/json",
		"application/javascript",
		"application/x-javascript",
		"application/xml",
		"application/xhtml+xml",
		"application/manifest+json",
		"application/rss+xml",
		"application/atom+xml",
		"application/ld+json",
		"image/svg+xml":
		return true
	}
	return false
}

// ---------------------------------------------------------------------------
// Compression middleware
// ---------------------------------------------------------------------------

type gzipResponseWriter struct {
	http.ResponseWriter
	req        *http.Request
	gz         *gzip.Writer
	wroteHead  bool
	compress   bool
	statusCode int
}

func (g *gzipResponseWriter) WriteHeader(code int) {
	if code >= 100 && code < 200 {
		// Informational response (e.g. 103 Early Hints): forward and keep the
		// writer open for the real status line.
		g.ResponseWriter.WriteHeader(code)
		return
	}
	if g.wroteHead {
		return
	}
	g.wroteHead = true
	g.statusCode = code

	if code >= 200 && code != http.StatusNoContent && code != http.StatusNotModified &&
		g.Header().Get("Content-Encoding") == "" &&
		g.req.Header.Get("Range") == "" &&
		compressibleType(g.Header().Get("Content-Type")) {
		// Tiny payloads (small JSON errors etc.) are not worth the CPU.
		small := false
		if cl := g.Header().Get("Content-Length"); cl != "" {
			if n, err := strconv.ParseInt(cl, 10, 64); err == nil && n < 1024 {
				small = true
			}
		}
		if !small {
			g.compress = true
			g.Header().Del("Content-Length")
			g.Header().Set("Content-Encoding", "gzip")
			addVary(g.Header(), "Accept-Encoding")
		}
	}
	g.ResponseWriter.WriteHeader(code)
}

func (g *gzipResponseWriter) Write(p []byte) (int, error) {
	if !g.wroteHead {
		g.WriteHeader(http.StatusOK)
	}
	if !g.compress {
		return g.ResponseWriter.Write(p)
	}
	if g.gz == nil {
		gz, err := gzip.NewWriterLevel(g.ResponseWriter, gzip.BestSpeed)
		if err != nil {
			return g.ResponseWriter.Write(p)
		}
		g.gz = gz
	}
	return g.gz.Write(p)
}

func (g *gzipResponseWriter) Flush() {
	if g.gz != nil {
		_ = g.gz.Flush()
	}
	if f, ok := g.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (g *gzipResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hj, ok := g.ResponseWriter.(http.Hijacker); ok {
		return hj.Hijack()
	}
	return nil, nil, fmt.Errorf("hijacking not supported")
}

func (g *gzipResponseWriter) Unwrap() http.ResponseWriter { return g.ResponseWriter }

func (g *gzipResponseWriter) close() {
	if g.gz != nil {
		_ = g.gz.Close()
	}
}

func addVary(h http.Header, value string) {
	for _, existing := range h.Values("Vary") {
		for _, part := range strings.Split(existing, ",") {
			if strings.EqualFold(strings.TrimSpace(part), value) {
				return
			}
		}
	}
	h.Add("Vary", value)
}

func withCompression(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Never wrap protocol upgrades (stream / mux relays) or HEAD.
		if r.Method == http.MethodHead ||
			r.Header.Get("Upgrade") != "" ||
			strings.EqualFold(r.Header.Get("Connection"), "upgrade") ||
			strings.HasPrefix(r.URL.Path, "/ws/") ||
			strings.HasPrefix(r.URL.Path, streamPathPrefix()+"/") ||
			strings.HasPrefix(r.URL.Path, "/xhttp-") ||
			strings.HasPrefix(r.URL.Path, muxPathPrefix()+"/") ||
			!acceptsEncoding(r, "gzip") {
			next.ServeHTTP(w, r)
			return
		}
		gw := &gzipResponseWriter{ResponseWriter: w, req: r}
		defer gw.close()
		next.ServeHTTP(gw, r)
	})
}

// ---------------------------------------------------------------------------
// Security / default headers middleware
// ---------------------------------------------------------------------------

func withSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("X-Frame-Options", "SAMEORIGIN")
		h.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")

		proto := strings.ToLower(strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-Proto"), ",")[0]))
		if r.TLS != nil || proto == "https" {
			h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		p := r.URL.Path
		if strings.HasPrefix(p, "/api/") {
			// Authenticated, per-user data must never be cached by proxies.
			h.Set("Cache-Control", "no-store")
			h.Set("X-Robots-Tag", "noindex, nofollow")
		}
		if strings.HasPrefix(p, "/sub/") || strings.HasPrefix(p, subPathPrefix()+"/") ||
			strings.HasPrefix(p, "/link/") || strings.HasPrefix(p, "/ws/") ||
			strings.HasPrefix(p, streamPathPrefix()+"/") {
			h.Set("X-Robots-Tag", "noindex, nofollow")
		}
		next.ServeHTTP(w, r)
	})
}

// ---------------------------------------------------------------------------
// Small HTTP helpers shared by the static / SEO layer
// ---------------------------------------------------------------------------

func etagMatches(r *http.Request, etag string) bool {
	if etag == "" {
		return false
	}
	header := r.Header.Get("If-None-Match")
	if header == "" {
		return false
	}
	if strings.TrimSpace(header) == "*" {
		return true
	}
	normalize := func(s string) string {
		s = strings.TrimSpace(s)
		s = strings.TrimPrefix(s, "W/")
		s = strings.Trim(s, `"`)
		return s
	}
	want := normalize(etag)
	for _, part := range strings.Split(header, ",") {
		if normalize(part) == want {
			return true
		}
	}
	return false
}

func writeNotModified(w http.ResponseWriter, etag string) {
	if etag != "" {
		w.Header().Set("ETag", etag)
	}
	w.WriteHeader(http.StatusNotModified)
}
