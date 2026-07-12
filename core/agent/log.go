package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
)

type JSONLogger struct {
	mu sync.Mutex
	w  io.Writer
}

func NewJSONLogger(w io.Writer) *JSONLogger {
	return &JSONLogger{w: w}
}

func (l *JSONLogger) LogTurn(ctx context.Context, event TurnEvent) {
	if l == nil || l.w == nil {
		return
	}
	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	_, _ = l.w.Write(append(data, '\n'))
}

type TerminalLogger struct {
	mu    sync.Mutex
	w     io.Writer
	color bool
}

func NewTerminalLogger(w io.Writer, color bool) *TerminalLogger {
	return &TerminalLogger{w: w, color: color}
}

func (l *TerminalLogger) LogTurn(ctx context.Context, event TurnEvent) {
	if l == nil || l.w == nil {
		return
	}

	line := terminalTurnLine(event)
	if l.color {
		line = colorizeTurnLine(event, line)
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	_, _ = fmt.Fprintln(l.w, line)
}

func terminalTurnLine(event TurnEvent) string {
	parts := []string{
		fmt.Sprintf("turn %s", event.FinishReason),
		fmt.Sprintf("duration=%s", event.Duration),
		fmt.Sprintf("iterations=%d", event.Iterations),
	}
	if event.InputTokens > 0 || event.OutputTokens > 0 {
		parts = append(parts, fmt.Sprintf("tokens=%d/%d", event.InputTokens, event.OutputTokens))
	}
	if event.CacheReadTokens > 0 || event.CacheWriteTokens > 0 {
		parts = append(parts, fmt.Sprintf("cache=%d/%d", event.CacheReadTokens, event.CacheWriteTokens))
	}
	if event.ReasoningTokens > 0 {
		parts = append(parts, fmt.Sprintf("reasoning=%d", event.ReasoningTokens))
	}
	if len(event.Tools) > 0 {
		parts = append(parts, "tools="+strings.Join(event.Tools, ","))
	}
	if event.Truncated {
		parts = append(parts, "truncated=true")
	}
	if event.Error != "" {
		parts = append(parts, "error="+event.Error)
	}
	return strings.Join(parts, " ")
}

func colorizeTurnLine(event TurnEvent, line string) string {
	switch event.FinishReason {
	case "stop":
		return "\x1b[32m" + line + "\x1b[0m"
	case "error", "max_iterations":
		return "\x1b[31m" + line + "\x1b[0m"
	default:
		return "\x1b[36m" + line + "\x1b[0m"
	}
}
