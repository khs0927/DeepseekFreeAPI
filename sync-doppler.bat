@echo off
chcp 65001 > nul
setlocal enabledelayedexpansion

echo ========================================================
echo   Syncing Secrets from Doppler (DeepSeek Bridge)
echo ========================================================

set SCRIPT_DIR=%~dp0
cd /d "%SCRIPT_DIR%"

set PROJECT=free-api-bridges
set CONFIG=prd

if not defined DOPPLER_TOKEN (
    if exist "%SCRIPT_DIR%.doppler_token" (
        set /p DOPPLER_TOKEN=<"%SCRIPT_DIR%.doppler_token"
    ) else if exist "%SCRIPT_DIR%..\.doppler_token" (
        set /p DOPPLER_TOKEN=<"%SCRIPT_DIR%..\.doppler_token"
    )
)

powershell -ExecutionPolicy Bypass -Command ^
    "$token = $env:DOPPLER_TOKEN;" ^
    "if (-not $token -and (Test-Path '.doppler_token')) { $token = (Get-Content '.doppler_token' -Raw).Trim() }" ^
    "if (-not $token -and (Test-Path '..\.doppler_token')) { $token = (Get-Content '..\.doppler_token' -Raw).Trim() }" ^
    "if ($token) {" ^
    "    try {" ^
    "        $headers = @{ Authorization = 'Bearer ' + $token };" ^
    "        $url = 'https://api.doppler.com/v3/configs/config/secrets/download?project=%PROJECT%&config=%CONFIG%&format=env';" ^
    "        $raw = (Invoke-WebRequest -Uri $url -Headers $headers -UseBasicParsing).Content;" ^
    "        if ($raw -match 'DEEPSEEK_TOKEN') {" ^
    "            $lines = $raw -split \"`r?`n\";" ^
    "            $ds = ($lines | Where-Object { $_ -match '^DEEPSEEK_TOKEN=' } | Select-Object -First 1) -replace '^DEEPSEEK_TOKEN=[\"''\s]*','' -replace '[\"''\s]*$','';" ^
    "            $proxy = ($lines | Where-Object { $_ -match '^PROXY_API_KEY=' } | Select-Object -First 1) -replace '^PROXY_API_KEY=[\"''\s]*','' -replace '[\"''\s]*$','';" ^
    "            $port = ($lines | Where-Object { $_ -match '^(DEEPSEEK_PORT|PORT)=' } | Select-Object -First 1) -replace '^[A-Z_]+=[\"''\s]*','' -replace '[\"''\s]*$','';" ^
    "            if (-not $proxy) { $proxy = 'local-dev-token' };" ^
    "            if (-not $port) { $port = '8000' };" ^
    "            $utf8NoBom = New-Object System.Text.UTF8Encoding($false);" ^
    "            [System.IO.File]::WriteAllText('.env', \"PORT=$port`r`nPROXY_API_KEY=$proxy`r`nDEEPSEEK_TOKEN=$ds`r`n\", $utf8NoBom);" ^
    "            Write-Output '[SUCCESS] Updated .env from Doppler (%PROJECT%/%CONFIG%)';" ^
    "            exit 0;" ^
    "        }" ^
    "    } catch {" ^
    "        Write-Output ('[WARN] Doppler sync failed: ' + $_.Exception.Message);" ^
    "    }" ^
    "} else {" ^
    "    Write-Output '[INFO] No DOPPLER_TOKEN provided. Preserving existing .env.';" ^
    "}"

exit /b 0
