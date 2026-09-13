package tests

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"deepseek/internal/dsproxy"
)

// discardLogger keeps proxy logs out of test output.
func discardLogger() *log.Logger {
	return log.New(io.Discard, "", 0)
}

// ── registry ─────────────────────────────────────────────────────────────────

func mustModel(t *testing.T, id string) dsproxy.Model {
	t.Helper()
	m, err := dsproxy.ResolveModel(id)
	if err != nil {
		t.Fatalf("ResolveModel(%q): %v", id, err)
	}
	return m
}

// TestRegistryPinsModels locks down the served model set and key
// models' capabilities and upstream model_type mapping.
func TestRegistryPinsModels(t *testing.T) {
	// Pin the wire values of the model_type constants.
	if got := string(dsproxy.ModelTypeDefault); got != "default" {
		t.Errorf("ModelTypeDefault = %q, want \"default\"", got)
	}
	if got := string(dsproxy.ModelTypeExpert); got != "expert" {
		t.Errorf("ModelTypeExpert = %q, want \"expert\"", got)
	}

	chat := mustModel(t, "deepseek-chat")
	if !chat.IsDefault {
		t.Error("deepseek-chat must be the default model")
	}
	if chat.Type != dsproxy.ModelTypeDefault {
		t.Errorf("chat model_type = %q, want \"default\"", chat.Type)
	}

	reasoner := mustModel(t, "deepseek-reasoner")
	if reasoner.Type != dsproxy.ModelTypeExpert {
		t.Errorf("reasoner model_type = %q, want \"expert\"", reasoner.Type)
	}
	if !reasoner.SupportsThink {
		t.Error("reasoner must support reasoning")
	}

	pro := mustModel(t, "deepseek-v4-pro")
	if pro.Type != dsproxy.ModelTypeExpert {
		t.Errorf("pro model_type = %q, want \"expert\"", pro.Type)
	}
	if !pro.SupportsThink {
		t.Error("pro must support reasoning")
	}
}

// TestResolveModel covers defaulting, case/whitespace tolerance and the
// unknown-model error.
func TestResolveModel(t *testing.T) {
	if m := mustModel(t, ""); m.ID != "deepseek-chat" {
		t.Errorf("empty model must resolve to the default, got %q", m.ID)
	}
	if m := mustModel(t, "  DeepSeek-V4-Pro  "); m.ID != "deepseek-v4-pro" {
		t.Errorf("matching must be case-insensitive/trimmed, got %q", m.ID)
	}
	if m := mustModel(t, "deepseek-reasoner"); m.ID != "deepseek-reasoner" {
		t.Errorf("deepseek-reasoner should resolve, got %q", m.ID)
	}

	_, err := dsproxy.ResolveModel("nonexistent-model-xyz")
	if err == nil {
		t.Fatal("unknown model must return error")
	}
	for _, want := range []string{"deepseek-chat", "deepseek-reasoner", "deepseek-v4-pro"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q must mention %q", err.Error(), want)
		}
	}
}

// TestValidateCapabilities pins capability validation.
func TestValidateCapabilities(t *testing.T) {
	chat := mustModel(t, "deepseek-chat")
	reasoner := mustModel(t, "deepseek-reasoner")

	if err := chat.ValidateCapabilities(true); err != nil {
		t.Errorf("chat + search must pass: %v", err)
	}
	if err := reasoner.ValidateCapabilities(false); err != nil {
		t.Errorf("reasoner without search must pass: %v", err)
	}
}

// TestProxyModelsAdvertisesRegistry verifies /v1/models data is generated
// from the registry.
func TestProxyModelsAdvertisesRegistry(t *testing.T) {
	want := map[string]bool{"deepseek-chat": false, "deepseek-reasoner": false, "deepseek-v4-pro": false}
	for _, m := range dsproxy.ProxyModels {
		id, _ := m["id"].(string)
		if _, ok := want[id]; ok {
			want[id] = true
		}
		if m["object"] != "model" || m["owned_by"] != "deepseek" {
			t.Errorf("model %q has wrong object/owned_by: %v", id, m)
		}
	}
	for id, seen := range want {
		if !seen {
			t.Errorf("model %q missing from /v1/models payload", id)
		}
	}
}

// ── upstream wire format ─────────────────────────────────────────────────────

// TestBuildChatCompletionBody pins the /chat/completion payload sent to
// DeepSeek, especially the always-present model_type field.
func TestBuildChatCompletionBody(t *testing.T) {
	cases := []struct {
		name      string
		modelType string
		want      string
	}{
		{"flash maps to default", "default", "default"},
		{"pro maps to expert", "expert", "expert"},
		{"empty falls back to default", "", "default"},
	}
	for _, tc := range cases {
		body := dsproxy.BuildChatCompletionBody(dsproxy.ChatParams{
			ChatSessionID:   "sess",
			Prompt:          "hi",
			ModelType:       tc.modelType,
			ThinkingEnabled: true,
			SearchEnabled:   false,
		})
		got, _ := body["model_type"].(string)
		if got != tc.want {
			t.Errorf("%s: model_type = %q, want %q", tc.name, got, tc.want)
		}
	}

	pro := dsproxy.BuildChatCompletionBody(dsproxy.ChatParams{
		ChatSessionID:   "sess",
		Prompt:          "hi",
		ModelType:       "expert",
		ThinkingEnabled: true,
	})
	if pro["thinking_enabled"] != true {
		t.Error("thinking_enabled must be forwarded")
	}
	if v, ok := pro["parent_message_id"]; !ok || v != nil {
		t.Errorf("absent parent must serialize as null, got %v (ok=%v)", v, ok)
	}
	if _, ok := pro["ref_file_ids"]; !ok {
		t.Error("ref_file_ids must stay present")
	}
}

// ── end-to-end rejection paths via ServeHTTP ────────────────────────────────

// postChat fires one chat completion request at the proxy (auth disabled).
func postChat(t *testing.T, body string) (*httptest.ResponseRecorder, map[string]any) {
	t.Helper()
	srv := dsproxy.NewProxyServer(discardLogger(), "")
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response not JSON: %v (%q)", err, rec.Body.String())
	}
	return rec, resp
}

func errorCode(t *testing.T, resp map[string]any) string {
	t.Helper()
	errObj, _ := resp["error"].(map[string]any)
	code, _ := errObj["code"].(string)
	return code
}

// TestHandleChatRejectsUnknownModel: an unregistered model id fails fast with
// model_not_found before any upstream session/PoW work happens.
func TestHandleChatRejectsUnknownModel(t *testing.T) {
	for _, model := range []string{"gpt-4o", "gibberish", "unknown-model-xyz"} {
		rec, resp := postChat(t, `{"model":"`+model+`","messages":[{"role":"user","content":"hi"}]}`)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("model %q: status = %d, want 400 (body: %s)", model, rec.Code, rec.Body.String())
		}
		if got := errorCode(t, resp); got != "model_not_found" {
			t.Errorf("model %q: error code = %q, want model_not_found", model, got)
		}
	}
}


// TestModelsEndpointListsNewRegistry checks GET /v1/models serves the registry models.
func TestModelsEndpointListsNewRegistry(t *testing.T) {
	srv := dsproxy.NewProxyServer(discardLogger(), "")
	req := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	var resp struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("bad JSON: %v", err)
	}
	ids := map[string]bool{}
	for _, m := range resp.Data {
		ids[m.ID] = true
	}
	if !ids["deepseek-chat"] || !ids["deepseek-reasoner"] || !ids["deepseek-v4-pro"] {
		t.Fatalf("/v1/models must list registered models, got %v", ids)
	}
}
