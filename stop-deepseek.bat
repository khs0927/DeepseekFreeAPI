@echo off
chcp 65001 > nul
title Stop DeepSeek Free API Bridge

echo ========================================================
echo   Stopping DeepSeek Free API Bridge (Port 8000)
echo ========================================================

taskkill /F /IM deepseek-proxy.exe >nul 2>nul
if %errorlevel% equ 0 (
    echo [OK] Stopped deepseek-proxy.exe
) else (
    echo [INFO] deepseek-proxy.exe was not running.
)

echo ========================================================
timeout /t 2 >nul
