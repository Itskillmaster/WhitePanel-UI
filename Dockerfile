# =============================================================================
# Stage 1: Obfuscated build
# =============================================================================
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /src

# Cache module downloads separately from the sources.
COPY backend/go.mod backend/go.sum ./
ENV GOPROXY=https://proxy.golang.org,direct
ENV GOSUMDB=sum.golang.org
RUN go mod download

# Install garble: renames symbols and encrypts string literals in the binary.
RUN GOBIN=/usr/local/bin go install mvdan.cc/garble@latest

COPY backend/*.go ./

ENV CGO_ENABLED=0 GOOS=linux
RUN garble -literals -tiny build -ldflags="-s -w" -o /out/edge-gateway .

# =============================================================================
# Stage 2: Minimal runtime
# =============================================================================
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata \
    && mkdir -p /app/data /app/.engine /app/static /app/seed

WORKDIR /app

COPY --from=builder /out/edge-gateway /app/edge-gateway
COPY static/ /app/static/
COPY seed/ /app/seed/

RUN chmod +x /app/edge-gateway

ENV PORT=8080 \
    ADMIN_PATH=/portal-8x4k2 \
    WS_PATH=/events \
    XHTTP_PATH=/cdn-v1 \
    SUB_PATH=/feed \
    ENABLE_CORE=0 \
    ENABLE_MTPROXY=0

EXPOSE 8080

CMD ["/app/edge-gateway"]
