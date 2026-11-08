# Multi-stage production build for maximum stealth & efficiency
# Stage 1: Build High-Performance Go Core
FROM golang:1.24-alpine AS builder

WORKDIR /build
COPY backend/go.mod ./
COPY backend/*.go ./

# Compile static, stripped binary (removes all debug info & signatures)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o webapp .

# Stage 2: Minimal & Clean Production Image
FROM alpine:3.21

# Install standard web certificates and python runtime for background worker
RUN apk add --no-cache ca-certificates python3 py3-pip curl

WORKDIR /app

# Copy binary from builder with innocent name
COPY --from=builder /build/webapp /app/webapp

# Copy assets and bot service
COPY static/ /app/static/
COPY bot/ /app/bot/

# Setup stealth entrypoint
COPY entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh /app/webapp

# Default HTTP Web Service Port
EXPOSE 8080

CMD ["/app/entrypoint.sh"]
