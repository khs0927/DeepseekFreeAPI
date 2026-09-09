#!/usr/bin/env bash
# DeepseekFreeAPI (DeepSeek Bridge) Startup Script for macOS / Linux
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# 1. Automatic Doppler Sync if token available
if [ -f "$SCRIPT_DIR/sync-doppler.sh" ]; then
    bash "$SCRIPT_DIR/sync-doppler.sh" || true
fi

# 2. Load .env file
if [ -f "$SCRIPT_DIR/.env" ]; then
    set -a
    source <(grep -v '^#' "$SCRIPT_DIR/.env" | sed 's/\r$//')
    set +a
fi

export PORT="${PORT:-8000}"
export PROXY_API_KEY="${PROXY_API_KEY:-local-dev-token}"

echo "========================================================"
echo "  Starting DeepSeek Free API Bridge with Agent Mode"
echo "========================================================"
echo "  Port:          $PORT"
echo "  Proxy API Key: $PROXY_API_KEY"
echo "  Agent Mode:    ENABLED"
echo "========================================================"

if [ -f "$SCRIPT_DIR/deepseek-proxy" ]; then
    echo "[INFO] Running deepseek-proxy binary..."
    exec "$SCRIPT_DIR/deepseek-proxy" --agent-mode
elif command -v go >/dev/null 2>&1; then
    echo "[INFO] Binary not found, running via go run..."
    exec go run . --agent-mode
else
    echo "[ERROR] Neither deepseek-proxy binary nor Go was found on PATH!"
    echo "Please build the binary ('go build -o deepseek-proxy .') or install Go."
    exit 1
fi
