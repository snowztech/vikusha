package llm

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func testClient(fn roundTripFunc) *http.Client {
	return &http.Client{Transport: fn}
}

func testResp(status int, headers map[string]string, body string) *http.Response {
	h := http.Header{}
	for name, value := range headers {
		h.Set(name, value)
	}
	return &http.Response{
		StatusCode: status,
		Header:     h,
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestPostJSONWithRetryRetries429And5xx(t *testing.T) {
	var sleeps []time.Duration
	oldWait := retryWait
	retryWait = func(ctx context.Context, d time.Duration) error {
		sleeps = append(sleeps, d)
		return nil
	}
	defer func() { retryWait = oldWait }()

	var calls int32
	client := testClient(func(r *http.Request) (*http.Response, error) {
		switch atomic.AddInt32(&calls, 1) {
		case 1:
			return testResp(http.StatusTooManyRequests, map[string]string{"retry-after": "0"}, "rate limited"), nil
		case 2:
			return testResp(http.StatusBadGateway, nil, "temporary"), nil
		default:
			return testResp(http.StatusOK, nil, `{"ok":true}`), nil
		}
	})

	raw, err := postJSONWithRetry(context.Background(), client, "http://example.test", nil, []byte(`{}`), "test")
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != `{"ok":true}` {
		t.Fatalf("body = %q, want success body", raw)
	}
	if calls != 3 {
		t.Fatalf("calls = %d, want 3", calls)
	}
	if len(sleeps) != 2 {
		t.Fatalf("sleeps = %#v, want two sleeps", sleeps)
	}
	if sleeps[0] != 0 || sleeps[1] != 200*time.Millisecond {
		t.Fatalf("sleeps = %#v, want retry-after 0 then 200ms", sleeps)
	}
}

func TestPostJSONWithRetryDoesNotRetry400(t *testing.T) {
	var calls int32
	client := testClient(func(r *http.Request) (*http.Response, error) {
		atomic.AddInt32(&calls, 1)
		return testResp(http.StatusBadRequest, nil, "bad request"), nil
	})

	_, err := postJSONWithRetry(context.Background(), client, "http://example.test", nil, []byte(`{}`), "test")
	if err == nil {
		t.Fatal("expected error")
	}
	if calls != 1 {
		t.Fatalf("calls = %d, want 1", calls)
	}
	if !strings.Contains(err.Error(), "test 400") {
		t.Fatalf("error = %q, want status", err)
	}
}

func TestPostJSONWithRetryStopsWhenContextCanceledDuringBackoff(t *testing.T) {
	oldWait := retryWait
	retryWait = waitRetry
	defer func() { retryWait = oldWait }()

	client := testClient(func(r *http.Request) (*http.Response, error) {
		return testResp(http.StatusServiceUnavailable, nil, "temporary"), nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := postJSONWithRetry(ctx, client, "http://example.test", nil, []byte(`{}`), "test")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context canceled", err)
	}
}

func TestParseRetryAfterHTTPDate(t *testing.T) {
	delay, ok := parseRetryAfter(time.Now().Add(time.Second).UTC().Format(http.TimeFormat))
	if !ok {
		t.Fatal("parseRetryAfter did not parse HTTP date")
	}
	if delay <= 0 {
		t.Fatalf("delay = %s, want positive", delay)
	}
}
