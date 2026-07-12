package agent

import (
	"bytes"
	"context"
	"encoding/json"
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
