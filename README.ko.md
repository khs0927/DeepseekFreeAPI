# DeepSeek Free API (한국어 가이드)

> DeepSeek 웹 클라이언트를 순수 Go(Golang) 환경에서 역공학하여 제공하는 고성능 OpenAI 호환 로컬 프록시 서버입니다. 유료 API 키 없이 개인 계정 토큰만으로 완전한 도구 호출(Tool Calling)과 추론(Thinking)을 사용할 수 있습니다.

---

## 🌟 주요 특징

- **순수 Go 구현 (Zero Dependency)**: 외부 파이썬이나 CGo 의존성 없이 단일 실행 파일(`.exe`)로 동작합니다.
- **클라우드플레어(Cloudflare) 우회 및 내장 WASM PoW 솔버**: `wazero` WebAssembly 런타임을 통해 DeepSeek의 `DeepSeekHashV1` 작업증명(Challenge)을 브라우저와 동일하게 실시간으로 풀어냅니다.
- **계정 차단 방지 (Anti-Ban Guardrails)**:
  - **파일 업로드 API 전면 배제**: 계정 차단의 주원인인 비정상 파일 업로드(`/api/v0/file/upload`)를 일체 호출하지 않고 `ref_file_ids: []`로 안전하게 인라인 처리합니다.
  - **Throwaway 세션 자동 파기 (GC)**: 요청이 완료되는 즉시 DeepSeek 서버에서 해당 세션을 백그라운드로 자동 삭제(`POST /chat_session/delete`)하여 계정에 비정상적인 세션이 누적되지 않습니다.
  - **최신 브라우저 규격 헤더**: Chrome 149 기반 핑거프린트(`x-client-version: 2.4.0`)를 사용하여 정상적인 브라우저 웹 사용자로 인식됩니다.
- **사전 예열 세션 풀 (Async Session Pool)**:
  - 백그라운드에서 5개의 세션을 미리 준비하여 요청 즉시 2~3초 내로 응답합니다.
  - 세션이 소모되면 즉시 자동 보충됩니다.
- **에이전트 모드 지원 (`--agent-mode`)**:
  - OpenAI 표준 함수 호출(Function Calling / Tools)을 자연어 에이전트 규약으로 완벽하게 번역합니다.
  - 모델의 사고 과정(`reasoning_content`) 및 도구 호출 결과를 자동 파싱합니다.
- **환경변수(.env) 자동 로드**: 작업 디렉터리 또는 실행 파일 옆의 `.env` 파일을 자동으로 감지하여 로드합니다.

---

## 📋 지원 모델 목록 (포트 8000)

| 모델 ID | 유형 (Type) | 웹 검색 | 사고/추론 (Think) | 기본값 | 설명 |
| :--- | :--- | :---: | :---: | :---: | :--- |
| `deepseek-chat` | `default` | O | O | **기본 모델** | 고속 일반 대화 및 코딩 지원 |
| `deepseek-reasoner` | `default` | X | O (자동) | - | 심층 사고(`reasoning_content`) 모델 |
| `deepseek-expert` | `default` | X | O (자동) | - | 고급 전문가 추론 모델 |
| `deepseek-v4-flash` | `default` | O | O | - | v4 플래시 모델 |
| `deepseek-v4-pro` | `default` | X | O | - | v4 프로 추론 모델 |

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
    "deepseek-expert"
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

