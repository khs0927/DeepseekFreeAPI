#!/usr/bin/env bash
# DeepseekFreeAPI Doppler Secret Sync for macOS / Linux
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

DOPPLER_TOKEN="${DOPPLER_TOKEN:-}"
if [ -z "$DOPPLER_TOKEN" ] && [ -f "$SCRIPT_DIR/.doppler_token" ]; then
    DOPPLER_TOKEN="$(cat "$SCRIPT_DIR/.doppler_token" | tr -d '\r\n[:space:]')"
fi
if [ -z "$DOPPLER_TOKEN" ] && [ -f "$SCRIPT_DIR/../.doppler_token" ]; then
    DOPPLER_TOKEN="$(cat "$SCRIPT_DIR/../.doppler_token" | tr -d '\r\n[:space:]')"
fi

PROJECT="${DOPPLER_PROJECT:-free-api-bridges}"
CONFIG="${DOPPLER_CONFIG:-prd}"

RAW_ENV=""
if [ -n "$DOPPLER_TOKEN" ]; then
    echo "[INFO] Fetching secrets from Doppler API ($PROJECT/$CONFIG)..."
    RAW_ENV=$(curl -s -f -H "Authorization: Bearer $DOPPLER_TOKEN" "https://api.doppler.com/v3/configs/config/secrets/download?project=${PROJECT}&config=${CONFIG}&format=env" 2>/dev/null || true)
elif command -v doppler >/dev/null 2>&1; then
    echo "[INFO] Fetching secrets using Doppler CLI ($PROJECT/$CONFIG)..."
    RAW_ENV=$(doppler secrets download --project "$PROJECT" --config "$CONFIG" --no-file --format env 2>/dev/null || true)
fi

if [ -n "$RAW_ENV" ] && echo "$RAW_ENV" | grep -q "DEEPSEEK_TOKEN"; then
    DEEPSEEK_TOKEN_VAL=$(echo "$RAW_ENV" | grep -E '^DEEPSEEK_TOKEN=' | head -n 1 | sed -e 's/^DEEPSEEK_TOKEN=//' -e 's/^"//' -e 's/"$//')
    PORT_VAL=$(echo "$RAW_ENV" | grep -E '^(DEEPSEEK_PORT|PORT)=' | head -n 1 | sed -e 's/^[A-Z_]*=//' -e 's/^"//' -e 's/"$//')
    PROXY_KEY_VAL=$(echo "$RAW_ENV" | grep -E '^PROXY_API_KEY=' | head -n 1 | sed -e 's/^PROXY_API_KEY=//' -e 's/^"//' -e 's/"$//')

    PORT_VAL="${PORT_VAL:-8000}"
    PROXY_KEY_VAL="${PROXY_KEY_VAL:-local-dev-token}"

    cat <<EOF > "$SCRIPT_DIR/.env"
PORT=${PORT_VAL}
PROXY_API_KEY=${PROXY_KEY_VAL}
DEEPSEEK_TOKEN=${DEEPSEEK_TOKEN_VAL}
EOF
    echo "[SUCCESS] Updated .env from Doppler ($PROJECT/$CONFIG)"
else
    echo "[INFO] Doppler sync skipped or token unavailable. Preserving current .env."
fi
