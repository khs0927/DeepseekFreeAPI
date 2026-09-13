package dsproxy

import (
	"testing"
)

func TestParseAgentToolCallsDSML(t *testing.T) {
	// The exact live DSML emitted by DeepSeek model in the user's report
	liveDSML := `<｜｜DSML｜｜ calls> <｜｜DSML｜｜ invoke name="Bash"> <｜｜DSML｜｜ parameter name="command" string="true">ls -la "/c/Users/khs09/.zcode/cli/artifacts/sess_384e46e3-bf03-4d89-ba56-2a1fee9ae7a5/" 2>/dev/null | head -50; echo "===WORKSPACE==="; ls -la "/c/Users/khs09/.zcode/workspace/default/" 2>/dev/null</｜｜DSML｜｜ parameter> <｜｜DSML｜｜ parameter name="description" string="true">Check artifact files and workspace contents</｜｜DSML｜｜ parameter> </｜｜DSML｜｜ invoke> </｜｜DSML｜｜ calls>`

	calls := ParseAgentToolCalls(liveDSML)
	if len(calls) != 1 {
		t.Fatalf("ParseAgentToolCalls(liveDSML) returned %d calls, want 1", len(calls))
	}

	fn := calls[0]["function"].(map[string]any)
	if fn["name"] != "Bash" {
		t.Errorf("tool name = %v, want Bash", fn["name"])
	}

	stripped := StripAgentToolCalls(liveDSML)
	if stripped != "" {
		t.Errorf("StripAgentToolCalls(liveDSML) = %q, want empty", stripped)
	}
}

func TestParseAgentToolCallsDSMLWithCommentary(t *testing.T) {
	text := `I will list the workspace directory for you.
<｜｜DSML｜｜ calls>
<｜｜DSML｜｜ invoke name="Bash">
<｜｜DSML｜｜ parameter name="command" string="true">ls -la</｜｜DSML｜｜ parameter>
</｜｜DSML｜｜ invoke>
</｜｜DSML｜｜ calls>
Please wait while I execute.`

	calls := ParseAgentToolCalls(text)
	if len(calls) != 1 {
		t.Fatalf("calls count = %d, want 1", len(calls))
	}

	stripped := StripAgentToolCalls(text)
	want := "I will list the workspace directory for you.\n\nPlease wait while I execute."
	if stripped != want {
		t.Errorf("StripAgentToolCalls = %q, want %q", stripped, want)
	}
}

func TestStreamInterceptorDSML(t *testing.T) {
	liveDSML := `Let me check. <｜｜DSML｜｜ calls> <｜｜DSML｜｜ invoke name="Bash"> <｜｜DSML｜｜ parameter name="command" string="true">pwd</｜｜DSML｜｜ parameter> </｜｜DSML｜｜ invoke> </｜｜DSML｜｜ calls>`

	// Chunk it finely (e.g. 5 runes per chunk) to verify chunk split boundary handling
	in := &AgentStreamInterceptor{}
	runes := []rune(liveDSML)
	var contentOut string
	var toolCallsOut []map[string]any

	chunkSize := 5
	for i := 0; i < len(runes); i += chunkSize {
		end := i + chunkSize
		if end > len(runes) {
			end = len(runes)
		}
		chunk := string(runes[i:end])
		parsed := in.Feed(chunk)
		if parsed.Content != "" {
			contentOut += parsed.Content
		}
		toolCallsOut = append(toolCallsOut, parsed.ToolCalls...)
	}

	final := in.Finish()
	if final.Content != "" {
		contentOut += final.Content
	}
	toolCallsOut = append(toolCallsOut, final.ToolCalls...)

	if len(toolCallsOut) != 1 {
		t.Fatalf("StreamInterceptor toolCalls = %d, want 1", len(toolCallsOut))
	}

	fn := toolCallsOut[0]["function"].(map[string]any)
	if fn["name"] != "Bash" {
		t.Errorf("tool name = %v, want Bash", fn["name"])
	}

	if contentOut != "Let me check. " {
		t.Errorf("contentOut = %q, want %q", contentOut, "Let me check. ")
	}
}

func TestParseAgentToolCallsASCIIPipes(t *testing.T) {
	text := `<||DSML|| calls>
<||DSML|| invoke name="Read">
<||DSML|| parameter name="path" string="true">test.txt</||DSML|| parameter>
</||DSML|| invoke>
</||DSML|| calls>`

	calls := ParseAgentToolCalls(text)
	if len(calls) != 1 {
		t.Fatalf("ParseAgentToolCalls(ASCII) returned %d calls, want 1", len(calls))
	}
	fn := calls[0]["function"].(map[string]any)
	if fn["name"] != "Read" {
		t.Errorf("tool name = %v, want Read", fn["name"])
	}
}
