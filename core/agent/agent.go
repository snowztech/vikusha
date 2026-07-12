package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/snowztech/vikusha/core/llm"
	"github.com/snowztech/vikusha/core/memory"
	"github.com/snowztech/vikusha/core/tool"
)

const (
	maxIterations        = 10
	defaultToolResultCap = 4000
)

type Agent struct {
	name          string
	model         string
	systemPrompt  string
	provider      llm.Provider
	tools         *tool.Registry
	memory        memory.Memory
	toolResultCap int
	logger        TurnLogger
	turnsMu       sync.Mutex
	userTurns     map[string]chan struct{}
	cancelMu      sync.Mutex
	nextCancelID  uint64
	userCancels   map[string]turnCancel
}

type turnCancel struct {
	id     uint64
	cancel context.CancelFunc
}

// Options contains the runtime dependencies required to construct an Agent
// directly. Most users should prefer vikusha.LoadAgent or vikusha.NewAgent.
type Options struct {
	Name          string
	Model         string
	SystemPrompt  string
	Provider      llm.Provider
	Tools         *tool.Registry
	Memory        memory.Memory
	ToolResultCap int
	Logger        TurnLogger
}

type TurnLogger interface {
	LogTurn(ctx context.Context, event TurnEvent)
}

type TurnEvent struct {
	Agent            string   `json:"agent"`
	UserID           string   `json:"user_id"`
	Model            string   `json:"model"`
	Duration         string   `json:"duration"`
	DurationMS       int64    `json:"duration_ms"`
	Iterations       int      `json:"iterations"`
	InputTokens      int      `json:"input_tokens,omitempty"`
	OutputTokens     int      `json:"output_tokens,omitempty"`
	CacheReadTokens  int      `json:"cache_read_tokens,omitempty"`
	CacheWriteTokens int      `json:"cache_write_tokens,omitempty"`
	Tools            []string `json:"tools,omitempty"`
	Error            string   `json:"error,omitempty"`
	Truncated        bool     `json:"truncated,omitempty"`
	FinishReason     string   `json:"finish_reason"`
}

func New(opts Options) (*Agent, error) {
	if opts.Name == "" {
		return nil, fmt.Errorf("agent: Name is required")
	}
	if opts.Model == "" {
		return nil, fmt.Errorf("agent: Model is required")
	}
	if opts.Provider == nil {
		return nil, fmt.Errorf("agent: Provider is required")
	}
	if opts.Tools == nil {
		opts.Tools = tool.NewRegistry()
	}
	if opts.ToolResultCap <= 0 {
		opts.ToolResultCap = defaultToolResultCap
	}
	return &Agent{
		name:          opts.Name,
		model:         opts.Model,
		systemPrompt:  opts.SystemPrompt,
		provider:      opts.Provider,
		tools:         opts.Tools,
		memory:        opts.Memory,
		toolResultCap: opts.ToolResultCap,
		logger:        opts.Logger,
		userTurns:     map[string]chan struct{}{},
		userCancels:   map[string]turnCancel{},
	}, nil
}

func (a *Agent) Name() string { return a.name }

func (a *Agent) Cancel(userID string) bool {
	key := userKey(userID)

	a.cancelMu.Lock()
	active := a.userCancels[key]
	a.cancelMu.Unlock()
	if active.cancel == nil {
		return false
	}
	active.cancel()
	return true
}

func turnEvent(start time.Time, agent, userID, model string) TurnEvent {
	duration := time.Since(start)
	return TurnEvent{
		Agent:      agent,
		UserID:     userID,
		Model:      model,
		Duration:   duration.String(),
		DurationMS: duration.Milliseconds(),
	}
}

func (e *TurnEvent) addUsage(usage llm.Usage) {
	e.InputTokens += usage.InputTokens
	e.OutputTokens += usage.OutputTokens
	e.CacheReadTokens += usage.CacheReadTokens
	e.CacheWriteTokens += usage.CacheWriteTokens
}
