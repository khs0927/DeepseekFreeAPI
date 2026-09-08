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
- **Anti-Ban Architecture**:
  - **Zero file uploads**: Never touches `/api/v0/file/upload`, which triggers bot detection. Uses clean inline payloads.
  - **Session garbage collector**: Deletes throwaway chat sessions on DeepSeek (`POST /chat_session/delete`) immediately after response completion so session IDs never accumulate on the account.
  - **Chrome 149 / Web client 2.4.0 fingerprinting**: Full compliance with the official web client protocol.
- **Async session-pool mode (default)**: Pre-warms 5 sessions so completion requests never wait on per-request session creation.
- **Agent mode** (`--agent-mode` / `AGENT_MODE=true`): OpenAI function/tool calling translated into clean role-tagged prompt shims.
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

## License

MIT License
