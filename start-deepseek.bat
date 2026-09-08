@echo off
chcp 65001 > nul
setlocal

echo ========================================================
echo   Starting DeepSeek Free API Bridge with Agent Mode
echo ========================================================
echo   Port:          8000
echo   Proxy API Key: local-dev-token
echo   Agent Mode:    ENABLED
echo ========================================================

set PORT=8000
set PROXY_API_KEY=local-dev-token

if exist "%~dp0deepseek-proxy.exe" (
    echo [INFO] Starting deepseek-proxy.exe...
    "%~dp0deepseek-proxy.exe" --agent-mode
) else (
    echo [INFO] deepseek-proxy.exe not found, running via go run...
    go run . --agent-mode
)
