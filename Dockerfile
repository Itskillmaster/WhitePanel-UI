# =============================================================================
# Stage 1: Build the Go binary
# =============================================================================
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /build

# Copy go.mod / go.sum first so dependency layer is cached
COPY backend/go.mod backend/go.sum* ./

# Download deps separately (better layer caching + explicit network step)
ENV GOPROXY=https://proxy.golang.org,direct
ENV GOSUMDB=sum.golang.org
ENV GOFLAGS=-mod=mod

RUN go mod download

# Copy sources
COPY backend/*.go ./

# Build static, stripped binary (no go mod tidy - avoids runtime network)
RUN CGO_ENABLED=0 GOOS=linux go build \
        -trimpath \
        -ldflags="-s -w" \
        -o /build/webapp .

# =============================================================================
# Stage 2: Runtime image
# =============================================================================
FROM alpine:3.21

RUN apk add --no-cache ca-certificates curl tzdata unzip python3

WORKDIR /app

# Pre-install Xray inside the image (no runtime download needed)
RUN mkdir -p /app/xray /app/data /app/logs

ARG TARGETARCH
RUN set -eux; \
    case "${TARGETARCH:-amd64}" in \
        amd64) XRAY_ARCH="64" ;; \
        arm64) XRAY_ARCH="arm64-v8a" ;; \
        arm)   XRAY_ARCH="arm32-v7a" ;; \
        386)   XRAY_ARCH="32" ;; \
        *)     XRAY_ARCH="64" ;; \
    esac; \
    curl -L --retry 3 --retry-delay 2 --fail \
        -o /tmp/xray.zip \
        "https://github.com/XTLS/Xray-core/releases/latest/download/Xray-linux-${XRAY_ARCH}.zip"; \
    unzip /tmp/xray.zip -d /app/xray/; \
    rm /tmp/xray.zip; \
    chmod +x /app/xray/xray; \
    /app/xray/xray version

# Binary
COPY --from=builder /build/webapp /app/webapp

# Static assets
COPY static/ /app/static/

# Bot (delete this line if you don't have a bot/ folder)
COPY bot/ /app/bot/

# Entrypoint
COPY entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh /app/webapp

ENV PANEL_PORT=8080 \
    WHITE_DATA_DIR=/app/data

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=45s --retries=3 \
    CMD curl -fsS http://localhost:8080/healthz > /dev/null || exit 1

CMD ["/app/entrypoint.sh"]
