package dsproxy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"time"
)

// AutoRefreshTokens runs the python auto-refresh script to extract/refresh tokens.
func (s *ProxyServer) AutoRefreshTokens() bool {
	s.log.Printf("[AutoRefresh] Running token refresh daemon oneshot...")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "python", "C:\\CODE\\auto-refresh-tokens.py", "--oneshot")
	out, err := cmd.CombinedOutput()
	if err != nil {
		s.log.Printf("[AutoRefresh] Token refresh script error: %v, out: %s", err, string(out))
		return false
	}
	s.log.Printf("[AutoRefresh] Token refresh result:\n%s", string(out))

	ReloadDotEnv()
	s.mu.Lock()
	s.apiSet = false
	s.mu.Unlock()
	return true
}

// ForwardToGLM transparently proxies a chat completion request to the GLM bridge (port 3101).
func (s *ProxyServer) ForwardToGLM(w http.ResponseWriter, r *http.Request, rawBody []byte, targetModel string) error {
	s.log.Printf("[Failover] -> Forwarding request to GLM bridge on port 3101 (target model: %s)", targetModel)

	var payload map[string]any
	if err := json.Unmarshal(rawBody, &payload); err == nil {
		if targetModel != "" {
			payload["model"] = targetModel
		}
		rawBody, _ = json.Marshal(payload)
	}

	glmURL := envOr("GLM_URL", "http://127.0.0.1:3101/v1/chat/completions")
	req, err := http.NewRequestWithContext(r.Context(), "POST", glmURL, bytes.NewReader(rawBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+envOr("PROXY_API_KEY", "local-dev-token"))

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		s.log.Printf("[Failover] GLM bridge connection failed: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBytes, _ := io.ReadAll(resp.Body)
		s.log.Printf("[Failover] GLM bridge returned HTTP %d: %s", resp.StatusCode, string(errBytes))
		return fmt.Errorf("GLM returned status %d: %s", resp.StatusCode, string(errBytes))
	}

	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)

	if flusher, ok := w.(http.Flusher); ok {
		buf := make([]byte, 4096)
		for {
			n, rerr := resp.Body.Read(buf)
			if n > 0 {
				w.Write(buf[:n])
				flusher.Flush()
			}
			if rerr != nil {
				break
			}
		}
		return nil
	}

	_, copyErr := io.Copy(w, resp.Body)
	return copyErr
}
