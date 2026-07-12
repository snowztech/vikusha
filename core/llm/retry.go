package llm

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

const (
	defaultRetryAttempts = 3
	defaultRetryBase     = 100 * time.Millisecond
)

var retryWait = waitRetry

func postJSONWithRetry(ctx context.Context, client *http.Client, url string, headers map[string]string, body []byte, provider string) ([]byte, error) {
	if client == nil {
		client = http.DefaultClient
	}

	var lastErr error
	for attempt := 1; attempt <= defaultRetryAttempts; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		for name, value := range headers {
			req.Header.Set(name, value)
		}

		resp, err := client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("http call: %w", err)
			if attempt == defaultRetryAttempts {
				return nil, lastErr
			}
			if err := retryWait(ctx, exponentialBackoff(attempt, "")); err != nil {
				return nil, err
			}
			continue
		}

		raw, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return nil, fmt.Errorf("read body: %w", readErr)
		}
		if resp.StatusCode == http.StatusOK {
			return raw, nil
		}

		lastErr = fmt.Errorf("%s %d: %s", provider, resp.StatusCode, raw)
		if !retryableStatus(resp.StatusCode) || attempt == defaultRetryAttempts {
			return nil, lastErr
		}
		if err := retryWait(ctx, exponentialBackoff(attempt, resp.Header.Get("retry-after"))); err != nil {
			return nil, err
		}
	}
	return nil, lastErr
}

func retryableStatus(status int) bool {
	return status == http.StatusTooManyRequests || status >= 500
}

func exponentialBackoff(attempt int, retryAfter string) time.Duration {
	if delay, ok := parseRetryAfter(retryAfter); ok {
		return delay
	}
	delay := defaultRetryBase
	for i := 1; i < attempt; i++ {
		delay *= 2
	}
	return delay
}

func parseRetryAfter(value string) (time.Duration, bool) {
	if value == "" {
		return 0, false
	}
	if seconds, err := strconv.Atoi(value); err == nil {
		if seconds < 0 {
			return 0, true
		}
		return time.Duration(seconds) * time.Second, true
	}
	t, err := http.ParseTime(value)
	if err != nil {
		return 0, false
	}
	delay := time.Until(t)
	if delay < 0 {
		return 0, true
	}
	return delay, true
}

func waitRetry(ctx context.Context, delay time.Duration) error {
	if delay <= 0 {
		return nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
