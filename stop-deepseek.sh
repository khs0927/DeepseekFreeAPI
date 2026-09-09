#!/usr/bin/env bash
# DeepseekFreeAPI (DeepSeek Bridge) Stop Script for macOS / Linux
echo "========================================================"
echo "  Stopping DeepSeek Bridge (deepseek-proxy)"
echo "========================================================"

pkill -f "deepseek-proxy" 2>/dev/null && echo "[OK] Stopped deepseek-proxy processes" || true

PORT="${PORT:-8000}"
if command -v lsof >/dev/null 2>&1; then
    PID=$(lsof -ti tcp:"$PORT" 2>/dev/null || true)
    if [ -n "$PID" ]; then
        kill -9 $PID 2>/dev/null && echo "[OK] Freed port $PORT (PID: $PID)" || true
    fi
fi
echo "[INFO] DeepSeek Bridge stopped."
