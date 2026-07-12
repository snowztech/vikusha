package llm

import (
	"context"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestOpenAICompleteRetriesTransientStatus(t *testing.T) {
	oldWait := retryWait
	retryWait = func(ctx context.Context, d time.Duration) error { return nil }
	defer func() { retryWait = oldWait }()

	var calls int32
	client := testClient(func(r *http.Request) (*http.Response, error) {
		if atomic.AddInt32(&calls, 1) == 1 {
			return testResp(http.StatusInternalServerError, nil, "temporary"), nil
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body: io.NopCloser(strings.NewReader(`{
			"choices": [
				{
					"message": {"role": "assistant", "content": "ok"},
					"finish_reason": "stop"
				}
			],
			"usage": {
				"prompt_tokens": 3,
				"completion_tokens": 2,
				"prompt_tokens_details": {"cached_tokens": 1},
				"completion_tokens_details": {"reasoning_tokens": 4}
			}
		}`)),
		}, nil
	})

	provider := NewOpenAICompat("test-key", "http://example.test")
	provider.http = client
	resp, err := provider.Complete(context.Background(), &Request{
		Model: "test-model",
		Messages: []Message{{
			Role:    "user",
			Content: []Block{{Type: BlockText, Text: "hello"}},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if calls != 2 {
		t.Fatalf("calls = %d, want 2", calls)
	}
	if len(resp.Content) != 1 || resp.Content[0].Text != "ok" {
		t.Fatalf("content = %#v, want ok", resp.Content)
	}
	if resp.Usage.InputTokens != 3 || resp.Usage.OutputTokens != 2 {
		t.Fatalf("usage = %#v, want input/output 3/2", resp.Usage)
	}
	if resp.Usage.CacheReadTokens != 1 || resp.Usage.ReasoningOutputTokens != 4 {
		t.Fatalf("usage = %#v, want cache read 1 and reasoning 4", resp.Usage)
	}
}
