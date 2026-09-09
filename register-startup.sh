#!/usr/bin/env bash
# DeepseekFreeAPI (DeepSeek Bridge) - macOS Auto-Start Installer (launchd)
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PLIST_DIR="$HOME/Library/LaunchAgents"
PLIST_FILE="$PLIST_DIR/com.antigravity.deepseek-bridge.plist"
LOG_DIR="$HOME/Library/Logs/Antigravity"

mkdir -p "$PLIST_DIR" "$LOG_DIR"

chmod +x "$SCRIPT_DIR/start-deepseek.sh"
if [ -f "$SCRIPT_DIR/sync-doppler.sh" ]; then
    chmod +x "$SCRIPT_DIR/sync-doppler.sh"
fi

cat <<EOF > "$PLIST_FILE"
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.antigravity.deepseek-bridge</string>
    <key>ProgramArguments</key>
    <array>
        <string>/bin/bash</string>
        <string>${SCRIPT_DIR}/start-deepseek.sh</string>
    </array>
    <key>WorkingDirectory</key>
    <string>${SCRIPT_DIR}</string>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>${LOG_DIR}/deepseek-bridge.log</string>
    <key>StandardErrorPath</key>
    <string>${LOG_DIR}/deepseek-bridge.err</string>
</dict>
</plist>
EOF

echo "[INFO] Generated LaunchAgent: $PLIST_FILE"

# Unload previous instance if running
launchctl unload -w "$PLIST_FILE" 2>/dev/null || true

# Load and start daemon
launchctl load -w "$PLIST_FILE"

echo "========================================================"
echo "  [SUCCESS] DeepSeek Bridge registered for macOS auto-start!"
echo "  Port:      8000"
echo "  Logs:      $LOG_DIR/deepseek-bridge.log"
echo "  Launchd:   launchctl status: loaded"
echo "========================================================"
