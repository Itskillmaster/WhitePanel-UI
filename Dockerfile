# =============================================================================
# Stage 1: Build the Go binary
# =============================================================================
FROM golang:1.24-alpine AS builder

# git is required by `go mod download` for HTTPS fetches
RUN apk add --no-cache git ca-certificates

WORKDIR /build

# Copy only go.mod. We intentionally do NOT copy go.sum:
#   - if it exists, `go mod tidy` will reuse it
#   - if it's missing, `go mod tidy` generates it (this is what fixed the
#     "missing go.sum entry for github.com/gorilla/websocket" error)
COPY backend/go.mod ./

# Copy Go sources (main.go, mtproxy.go, etc.)
COPY backend/*.go ./

# Generate go.sum and build a static, stripped binary
RUN go mod tidy \
 && CGO_ENABLED=0 GOOS=linux go build \
        -trimpath \
        -ldflags="-s -w" \
        -o /build/webapp .

# =============================================================================
# Stage 2: Runtime image
# =============================================================================
FROM alpine:3.21

# ca-certificates: HTTPS to Xray releases, Telegram API, etc.
# curl:            healthcheck
# unzip:           Xray zip extraction (used by the Go process at runtime)
# tzdata:          correct logs
RUN apk add --no-cache ca-certificates curl tzdata unzip

WORKDIR /app

# Binary
COPY --from=builder /build/webapp /app/webapp

# Static assets
COPY static/ /app/static/

# ---- bot/ is OPTIONAL ----
# If your repo contains bot/, keep this line.
# If you do NOT have a bot/ folder, DELETE the next line entirely.
COPY bot/ /app/bot/

# Entrypoint
COPY entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh /app/webapp

# Runtime directories (persist with a Railway volume at /app/data)
RUN mkdir -p /app/data /app/xray /app/logs

ENV PANEL_PORT=8080 \
    WHITE_DATA_DIR=/app/data

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=45s --retries=3 \
    CMD curl -fsS http://localhost:8080/healthz > /dev/null || exit 1

CMD ["/app/entrypoint.sh"]