# =============================================================================
# Stage 1: Build the Go core with automatic go.sum generation
# =============================================================================
FROM golang:1.24-alpine AS builder

# Install git (needed for go mod download over HTTPS) and certificates
RUN apk add --no-cache git ca-certificates

WORKDIR /build

# ---- Dependency layer (cacheable) ----
# Wildcard on go.sum so the build works whether the file exists or not.
COPY backend/go.mod backend/go.sum* ./

# Copy Go sources (including mtproxy.go if you added it)
COPY backend/*.go ./

# Generate go.sum + download modules, then build a static stripped binary.
# `go mod tidy` scans the source and writes any missing go.sum entries —
# this is what fixes the "missing go.sum entry for gorilla/websocket" error.
RUN go mod tidy \
 && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /build/webapp .

# =============================================================================
# Stage 2: Minimal runtime image
# =============================================================================
FROM alpine:3.21

# ca-certificates: for HTTPS (Xray download, Telegram API, QR service)
# curl:            for HEALTHCHECK and debugging
# tzdata:          correct timestamps
# xz:              Xray zip is deflate only (zip is enough), but useful for future
RUN apk add --no-cache \
        ca-certificates \
        curl \
        tzdata \
        unzip

WORKDIR /app

# Compiled binary
COPY --from=builder /build/webapp /app/webapp

# Static assets and optional Python bot
COPY static/ /app/static/
COPY bot/    /app/bot/ 2>/dev/null || true

# Entrypoint
COPY entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh /app/webapp

# Data + xray directories (Railway volume will mount at /app/data)
RUN mkdir -p /app/data /app/xray /app/logs \
 && chown -R nobody:nobody /app/data /app/xray /app/logs

# Environment
ENV PANEL_PORT=8080 \
    WHITE_DATA_DIR=/app/data \
    MTPROXY_BIN=/app/mtproto-proxy

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=45s --retries=3 \
    CMD curl -fsS http://localhost:8080/healthz > /dev/null || exit 1

CMD ["/app/entrypoint.sh"]
