# DeepSeek Free API

> A pure Go proxy for DeepSeek with Cloudflare bypass and OpenAI-compatible interface — no paid API key required.

![Go](https://img.shields.io/badge/Go-1.26+-blue?style=flat-square)
![OpenAI Compatible](https://img.shields.io/badge/OpenAI-Compatible-green?style=flat-square)
![Cloudflare Bypass](https://img.shields.io/badge/Cloudflare-Bypass-purple?style=flat-square)
![Free API](https://img.shields.io/badge/API-Free-orange?style=flat-square)

[한국어 문서 (Korean README)](README.ko.md)

---

## Features

- Full DeepSeek API implementation (pure Go, zero runtime dependencies)
- Cloudflare protection detection with automatic retry
- **Proof of Work (PoW) challenge solving** — WASM run in-process via `wazero` (no external browser or binary needed)
- OpenAI-compatible proxy server on port `8000`
- **Real model registry** — includes:
  - `deepseek-chat` (default, fast conversation & coding)
  - `deepseek-reasoner` (thinking/reasoning enabled, returns `reasoning_content`)
  - `deepseek-expert` (expert thinking enabled)
  - `deepseek-v4-flash`, `deepseek-v4-pro`
- **Anti-Ban Strict Architecture**:
  - **Zero file uploads**: Never touches `/api/v0/file/upload`, which triggers bot detection and account mutes (`biz_code: 5, user is muted`). All payloads send `"ref_file_ids": []`.
  - **Session recycling & garbage collector**: Recycles healthy sessions (`pool.SetReuse(true)`) to avoid rate limit spikes, and purges throwaway sessions upstream (`POST /chat_session/delete`) so dead sessions never pile up on the account.
  - **Explicit auth error reporting**: Accurately intercepts upstream `code: 40003` to report `DeepSeek Auth Error (40003): Authorization Failed (Please update DEEPSEEK_TOKEN in .env)`.
  - **UTF-8 BOM resilience**: Automatically detects and trims UTF-8 BOM (`\xef\xbb\xbf` / `\ufeff`) from `.env` files written on Windows PowerShell.
  - **No raw special tokens**: Replaces raw special tokens with clean XML-sectioned prompt shims (`<system>`, `<tools>`, `<current_task>`, `<output_rules>`).
  - **Chrome 149 / Web client 2.4.0 fingerprinting**: Full compliance with the official web client protocol.
- **Async session-pool mode (default)**: Pre-warms 5 sessions so completion requests never wait on per-request session creation.
- **Agent mode** (`--agent-mode` / `AGENT_MODE=true`): OpenAI tools and DeepSeek DSML tool tags (`<<<TOOL_CALL>>>` and `<tool_call>`) translated into native OpenAI streaming tool_calls.
- **Automatic .env loading**: Automatically loads `.env` from the project or binary directory.

---

## Supported Models (Port 8000)

| Model ID | Backend Type | Search | Thinking | Default | Notes |
| :--- | :--- | :---: | :---: | :---: | :--- |
| `deepseek-chat` | `default` | Yes | Yes | **Default** | Fast general chat and coding |
| `deepseek-reasoner` | `default` | No | Yes (auto) | - | Returns `reasoning_content` |
| `deepseek-expert` | `default` | No | Yes (auto) | - | Expert reasoning model |
| `deepseek-v4-flash` | `default` | Yes | Yes | - | V4 Flash model |
| `deepseek-v4-pro` | `default` | No | Yes (auto) | - | V4 Pro reasoning model |

---

## Quick Start

### 1. Requirements
- [Go 1.26+](https://go.dev/dl/)

### 2. Obtain your DeepSeek token
1. Sign in to [chat.deepseek.com](https://chat.deepseek.com) in your browser.
2. Open DevTools (F12) -> Console.
3. Run:
   ```js
   JSON.parse(localStorage.getItem("userToken")).value
   ```
4. Copy the output string.

### 3. Setup .env
Create `.env` or copy `.env.example`:
```env
PORT=8000
PROXY_API_KEY=local-dev-token
DEEPSEEK_TOKEN=your_token_here
```

### 4. Build and Run

#### Windows One-Click
```cmd
start-deepseek.bat
```

#### Command Line
```bash
go build -o deepseek-proxy.exe .
deepseek-proxy.exe --agent-mode
```

Endpoint: `http://127.0.0.1:8000/v1`  
API Key: `local-dev-token`

---

## ZCode & OpenCode Configuration

```json
{
  "provider": "openai-compatible",
  "base_url": "http://127.0.0.1:8000/v1",
  "api_key": "local-dev-token",
  "models": [
    "deepseek-chat",
    "deepseek-reasoner",
    "deepseek-expert"
  ]
}
```

---

## Cross-Platform Scripts & Auto-Start

### macOS & Linux
- **Start / Stop**:
  ```bash
  chmod +x *.sh
  ./start-deepseek.sh
  ./stop-deepseek.sh
  ```
- **macOS Auto-Start via launchd**:
  ```bash
  ./register-startup.sh    # Installs ~/Library/LaunchAgents/com.antigravity.deepseek-bridge.plist
  ./unregister-startup.sh  # Removes LaunchAgent
  ```
  Logs: `~/Library/Logs/Antigravity/deepseek-bridge.log`

### Windows
- **Start / Stop**:
  ```cmd
  start-deepseek.bat
  stop-deepseek.bat
  ```
- **Auto-Start at Logon**:
  ```powershell
  powershell -ExecutionPolicy Bypass -File .\register-startup.ps1
  powershell -ExecutionPolicy Bypass -File .\unregister-startup.ps1
  ```

### Doppler Cloud Secret Auto-Sync
If `DOPPLER_TOKEN` is exported or `.doppler_token` is present, secrets are automatically synchronized from Doppler vault (`free-api-bridges/prd`):
- macOS / Linux: `./sync-doppler.sh`
- Windows: `sync-doppler.bat`
- `start-deepseek.sh` and `start-deepseek.bat` auto-trigger this before booting the bridge.

---

## License

MIT License

