#!/bin/sh
set -e

# Masking process signatures
export APP_NAME="cloud-web-service"
export PANEL_PORT="${PORT:-8080}"

echo "Starting Application Service on port ${PANEL_PORT}..."

# Launch Python Bot Worker in background if token is provided
if [ -n "$WHITE_BOT_TOKEN" ]; then
    echo "Initializing background worker..."
    python3 /app/bot/bot_service.py &
fi

# Launch Primary Go Core Web Application
exec /app/webapp
