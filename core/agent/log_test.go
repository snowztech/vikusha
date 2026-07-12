package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestJSONLoggerWritesTurnEvent(t *testing.T) {
	var buf bytes.Buffer
	logger := NewJSONLogger(&buf)

	logger.LogTurn(context.Background(), TurnEvent{
		Agent:           "writer",
		UserID:          "lucas",
		Model:           "test-model",
		Duration:        "10ms",
		DurationMS:      10,
		Iterations:      2,
		InputTokens:     20,
		OutputTokens:    5,
		ReasoningTokens: 2,
		Tools:           []string{"file_list"},
		FinishReason:    "stop",
	})

	var got TurnEvent
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Agent != "writer" || got.UserID != "lucas" || got.Tools[0] != "file_list" {
		t.Fatalf("event = %#v", got)
	}
	if got.InputTokens != 20 || got.OutputTokens != 5 {
		t.Fatalf("tokens = input %d output %d, want 20/5", got.InputTokens, got.OutputTokens)
	}
	if got.ReasoningTokens != 2 {
		t.Fatalf("reasoning tokens = %d, want 2", got.ReasoningTokens)
	}
}

func TestTerminalLoggerWritesReadableTurnEvent(t *testing.T) {
	var buf bytes.Buffer
	logger := NewTerminalLogger(&buf, false)

	logger.LogTurn(context.Background(), TurnEvent{
		Duration:         "12ms",
		Iterations:       2,
		InputTokens:      20,
		OutputTokens:     5,
		CacheReadTokens:  3,
		CacheWriteTokens: 1,
		ReasoningTokens:  2,
		Tools:            []string{"file_list", "file_read"},
		Truncated:        true,
		FinishReason:     "stop",
	})

	got := buf.String()
	for _, want := range []string{
		"turn stop",
		"duration=12ms",
		"iterations=2",
		"tokens=20/5",
		"cache=3/1",
		"reasoning=2",
		"tools=file_list,file_read",
		"truncated=true",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("terminal log = %q, want %q", got, want)
		}
	}
	if strings.Contains(got, "\x1b[") {
		t.Fatalf("terminal log should not contain color escapes: %q", got)
	}
}

func TestTerminalLoggerCanColorize(t *testing.T) {
	var buf bytes.Buffer
	logger := NewTerminalLogger(&buf, true)

	logger.LogTurn(context.Background(), TurnEvent{
		Duration:     "1ms",
		Iterations:   1,
		FinishReason: "error",
		Error:        "failed",
	})

	got := buf.String()
	if !strings.Contains(got, "\x1b[31m") || !strings.Contains(got, "\x1b[0m") {
		t.Fatalf("terminal log = %q, want red color escapes", got)
	}
	if !strings.Contains(got, "error=failed") {
		t.Fatalf("terminal log = %q, want error", got)
	}
}
