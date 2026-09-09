#!/usr/bin/env bash
# DeepseekFreeAPI (DeepSeek Bridge) - macOS Auto-Start Uninstaller (launchd)
PLIST_FILE="$HOME/Library/LaunchAgents/com.antigravity.deepseek-bridge.plist"

if [ -f "$PLIST_FILE" ]; then
    echo "[INFO] Unloading LaunchAgent..."
    launchctl unload -w "$PLIST_FILE" 2>/dev/null || true
    rm -f "$PLIST_FILE"
    echo "[SUCCESS] Removed $PLIST_FILE. Auto-start disabled."
else
    echo "[INFO] LaunchAgent was not registered."
fi
