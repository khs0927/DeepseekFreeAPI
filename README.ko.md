# DeepSeek Free API (한국어 가이드)

> DeepSeek 웹 클라이언트를 순수 Go(Golang) 환경에서 역공학하여 제공하는 고성능 OpenAI 호환 로컬 프록시 서버입니다. 유료 API 키 없이 개인 계정 토큰만으로 완전한 도구 호출(Tool Calling)과 추론(Thinking)을 사용할 수 있습니다.

---

## 🌟 주요 특징

- **순수 Go 구현 (Zero Dependency)**: 외부 파이썬이나 CGo 의존성 없이 단일 실행 파일(`.exe`)로 동작합니다.
- **클라우드플레어(Cloudflare) 우회 및 내장 WASM PoW 솔버**: `wazero` WebAssembly 런타임을 통해 DeepSeek의 `DeepSeekHashV1` 작업증명(Challenge)을 브라우저와 동일하게 실시간으로 풀어냅니다.
- **계정 차단 방지 철통 보호 (Anti-Ban Strict Guardrails)**:
  - **파일 업로드 API 전면 배제 (Zero File Upload)**: 계정 즉시 정지/뮤트(`biz_code: 5, user is muted`)의 주원인인 `/api/v0/file/upload`를 일체 호출하지 않으며, 모든 요청에 `"ref_file_ids": []`를 안전하게 전송합니다.
  - **세션 리사이클링 및 쓰레기 수거 (GC)**:
    - 정상 세션은 세션 풀(`pool.SetReuse(true)`)로 재활용하여 초당/분당 수십 번의 세션 생성을 유발하지 않아 429 속도 제한 및 이상 트래픽 탐지를 회피합니다.
    - 각 요청은 `parent_message_id: nil`로 격리되며, 에러 발생 또는 서버 종료 시 `POST /chat_session/delete`를 통해 계정 내 잔여 세션을 완전히 파기합니다.
  - **정확한 토큰 만료 진단**: DeepSeek 서버가 HTTP 200과 함께 `code: 40003` 반환 시 모호한 에러 대신 `DeepSeek Auth Error (40003): Authorization Failed`를 명시적으로 출력하여 즉각적인 토큰 갱신을 유도합니다.
  - **UTF-8 BOM 자동 제거**: Windows PowerShell `Set-Content`로 작성된 `.env`의 UTF-8 BOM(`\xef\xbb\xbf`)을 감지하여 안전하게 제거합니다.
  - **특수 토큰 유출 방지**: `<|begin_of_sentence|>` 등 원시 탈옥 토큰 대신 XML 구조화 프롬프트 딤(`agent_mode`)을 사용하여 업스트림 필터 오작동을 차단합니다.
  - **최신 브라우저 규격 헤더**: Chrome 149 기반 핑거프린트(`x-client-version: 2.4.0`)를 사용하여 정상적인 브라우저 웹 사용자로 인식됩니다.
- **사전 예열 세션 풀 (Async Session Pool)**:
  - 백그라운드에서 5개의 세션을 미리 준비하여 요청 즉시 2~3초 내로 응답합니다.
  - 세션이 소모되면 즉시 자동 보충됩니다.
- **에이전트 모드 지원 (`--agent-mode`)**:
  - OpenAI 표준 함수 호출(Function Calling / Tools) 및 DeepSeek 고유 DSML 태그(`<<<TOOL_CALL>>>` 및 `<tool_call>`)를 표준 OpenAI `tool_calls` 델타로 자동 변환합니다.
  - 모델의 사고 과정(`reasoning_content`) 및 도구 호출 결과를 실시간 SSE 스트리밍으로 전달합니다.
- **환경변수(.env) 자동 로드**: 작업 디렉터리 또는 실행 파일 옆의 `.env` 파일을 자동으로 감지하여 로드합니다.

---

## 📋 지원 모델 목록 (포트 8000)

DeepSeek 공식 웹사이트(`chat.deepseek.com`) 하단의 두 가지 핵심 토글인 **깊은 생각 (DeepThink)**과 **검색 (Web Search)**의 조합에 대응하는 **4가지 핵심 모델**을 제공합니다:

| 모델 ID | 웹 토글 대응 모드 | 웹 검색 | 사고/추론 (Think) | 기본값 | 설명 |
| :--- | :--- | :---: | :---: | :---: | :--- |
| `deepseek-chat` | 기본 모드 (깊은생각 OFF, 검색 OFF) | X | X | **기본 모델** | 고속 일반 대화 및 코딩 지원 (DeepSeek-V3) |
| `deepseek-reasoner` | 깊은생각 모드 (깊은생각 ON, 검색 OFF) | X | **O (자동)** | - | 심층 사고/추론(`reasoning_content`) 모델 (DeepSeek-R1) |
| `deepseek-search` | 검색 모드 (깊은생각 OFF, 검색 ON) | **O** | X | - | 실시간 웹 검색 지원 모델 |
| `deepseek-deep-search` | 깊은생각 + 검색 모드 (깊은생각 ON, 검색 ON) | **O** | **O (자동)** | - | 심층 사고 + 실시간 웹 검색 결합 모델 |

### 하위 호환 별칭 (Aliases)
기존에 사용하던 레거시 모델명은 4대 모드로 자동 변환되어 에러 없이 정상 작동합니다:
- `deepseek-reasoner-search`, `deepseek-r1-search`, `deepseek-think-search` → `deepseek-deep-search`
- `deepseek-r1`, `deepseek-think`, `deepseek-expert`, `deepseek-v4-pro` → `deepseek-reasoner`
- `deepseek-web-search` → `deepseek-search`
- `deepseek-v4-flash`, `deepseek-v4.1-flash`, `deepseek-v4.1`, `deepseek-v3`, `deepseek` → `deepseek-chat`

---

## 🚀 빠른 시작 (Quick Start)

### 1. 사전 요구사항
- [Go 1.26 이상](https://go.dev/dl/)

### 2. DeepSeek 계정 토큰 발급 방법
1. 웹 브라우저에서 [chat.deepseek.com](https://chat.deepseek.com)에 로그인합니다.
2. `F12` 키를 눌러 개발자 도구를 열고 **Console(콘솔)** 탭으로 이동합니다.
3. 아래 코드를 붙여넣고 엔터를 누릅니다:
   ```javascript
   JSON.parse(localStorage.getItem("userToken")).value
   ```
4. 따옴표 없이 출력된 문자열(토큰)을 복사합니다.

### 3. 환경변수 설정 (`.env`)
저장소 루트에 `.env` 파일을 생성하거나 `.env.example`을 복사하여 사용합니다:
```env
PORT=8000
PROXY_API_KEY=local-dev-token
DEEPSEEK_TOKEN=발급받은_DEEPSEEK_토큰
```

> **주의**: `.env` 파일은 `.gitignore`에 등록되어 깃허브 커밋 시 안전하게 보호됩니다.

### 4. 빌드 및 실행

#### Windows (원클릭 배치 파일)
```cmd
start-deepseek.bat
```

#### 명령줄 직접 빌드 및 실행
```bash
# 바이너리 빌드
go build -o deepseek-proxy.exe .

# 에이전트 모드로 실행
deepseek-proxy.exe --agent-mode
```

서버가 정상 구동되면 아래 주소로 엔드포인트가 열립니다:
- **Base URL**: `http://127.0.0.1:8000/v1`
- **Auth Token**: `local-dev-token`

---

## 💻 ZCode / OpenCode 연동 방법

OpenCode 또는 ZCode 설정 파일(`settings.json` 또는 UI 공급자 등록)에 다음과 같이 추가합니다:

```json
{
  "provider": "openai-compatible",
  "base_url": "http://127.0.0.1:8000/v1",
  "api_key": "local-dev-token",
  "models": [
    "deepseek-chat",
    "deepseek-reasoner",
    "deepseek-search",
    "deepseek-deep-search"
  ]
}
```

---

## 🍎 macOS 및 Linux 지원 & 부팅 시 자동 시작 (Auto-Start)

### macOS / Linux 실행 및 관리
```bash
# 실행 권한 부여
chmod +x *.sh

# 서버 시작
./start-deepseek.sh

# 서버 중지
./stop-deepseek.sh
```

### macOS 부팅/로그인 시 자동 실행 등록 (launchd)
macOS 표준 백그라운드 서비스(`launchd`)로 등록되어 부팅 시 자동 실행됩니다:
```bash
# 자동 시작 등록 및 즉시 시작
./register-startup.sh

# 자동 시작 등록 해제
./unregister-startup.sh
```
- 로그 위치: `~/Library/Logs/Antigravity/deepseek-bridge.log`

### Windows 부팅 시 자동 시작 등록
```powershell
# Windows 시작프로그램 바로가기 등록 (백그라운드 최소화 실행)
powershell -ExecutionPolicy Bypass -File .\register-startup.ps1

# 시작프로그램 등록 해제
powershell -ExecutionPolicy Bypass -File .\unregister-startup.ps1
```

---

## ☁️ Doppler 클라우드 시크릿 자동 동기화

Doppler 클라우드 금고에서 시크릿을 자동으로 내려받아 `.env`를 갱신합니다:
```bash
# 환경변수 DOPPLER_TOKEN 설정 또는 .doppler_token 파일 생성 후 실행
# macOS / Linux
./sync-doppler.sh

# Windows
sync-doppler.bat
```
- `start-deepseek.sh` 및 `start-deepseek.bat` 실행 시 자동으로 Doppler 동기화가 진행됩니다.

---

## 🔒 깃허브 시크릿 (GitHub Secrets) 관리
GitHub Actions 또는 워크플로우를 위해 토큰을 GitHub Secrets로 저장할 수 있습니다:
- `DEEPSEEK_TOKEN`: DeepSeek 사용자 인증 토큰
- `PROXY_API_KEY`: 로컬 인증 키 (`local-dev-token`)
- `PORT`: `8000`

