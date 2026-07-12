package agent

import (
	"context"
	"encoding/json"
	"io"
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
