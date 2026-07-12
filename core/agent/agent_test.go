package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/snowztech/vikusha/core/llm"
	"github.com/snowztech/vikusha/core/memory"
	"github.com/snowztech/vikusha/core/tool"
)

type memoryStub struct {
	entries []memory.Entry
	err     error
}

func (m memoryStub) Load(ctx context.Context, userID string) ([]memory.Entry, error) {
	return m.entries, m.err
}

func (m memoryStub) Save(ctx context.Context, userID string, entry memory.Entry) error {
	return m.err
}

func (m memoryStub) Search(ctx context.Context, userID, query string, k int) ([]memory.Entry, error) {
	if m.err != nil {
		return nil, m.err
	}
	if len(m.entries) > k {
		return m.entries[:k], nil
	}
	return m.entries, nil
}

type staticProvider struct{}

func (staticProvider) Name() string { return "static" }

func (staticProvider) Complete(ctx context.Context, req *llm.Request) (*llm.Response, error) {
	return &llm.Response{Content: []llm.Block{{Type: llm.BlockText, Text: "ok"}}}, nil
}

func TestSystemWithMemoryAppendsEntries(t *testing.T) {
	a := &Agent{
		systemPrompt: "You are helpful.",
		memory: memoryStub{entries: []memory.Entry{
			{Type: memory.EntryPreference, Content: "Lucas likes concise answers.", CreatedAt: time.Now()},
			{Type: memory.EntryFact, Content: "Lucas is building Vikusha.", CreatedAt: time.Now()},
		}},
	}

	system, err := a.systemWithMemory(context.Background(), "lucas")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"You are helpful.",
		"Relevant memory for this user:",
		"- preference: Lucas likes concise answers.",
		"- fact: Lucas is building Vikusha.",
	} {
		if !strings.Contains(system, want) {
			t.Fatalf("system prompt missing %q:\n%s", want, system)
		}
	}
}

func TestSystemWithMemoryNoEntriesUsesBasePrompt(t *testing.T) {
	a := &Agent{systemPrompt: "Base.", memory: memoryStub{}}

	system, err := a.systemWithMemory(context.Background(), "lucas")
	if err != nil {
		t.Fatal(err)
	}
	if system != "Base." {
		t.Fatalf("system = %q, want base prompt", system)
	}
}

func TestNewDefaultsToolResultCap(t *testing.T) {
	a, err := New(Options{
		Name:     "test",
		Model:    "test-model",
		Provider: staticProvider{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if a.toolResultCap != defaultToolResultCap {
		t.Fatalf("toolResultCap = %d, want %d", a.toolResultCap, defaultToolResultCap)
	}
}

func TestCapToolResultLeavesShortOutput(t *testing.T) {
	a := &Agent{toolResultCap: 5}
	got, truncated := a.capToolResult("hey")
	if got != "hey" {
		t.Fatalf("capToolResult() = %q, want hey", got)
	}
	if truncated {
		t.Fatal("capToolResult() truncated short output")
	}
}

func TestCapToolResultTruncatesLongOutput(t *testing.T) {
	a := &Agent{toolResultCap: 5}
	got, truncated := a.capToolResult("hello world")
	if !truncated {
		t.Fatal("capToolResult() did not report truncation")
	}
	wantSuffix := fmt.Sprintf("[tool result truncated: %d bytes omitted]", 6)
	if !strings.HasPrefix(got, "hello\n\n") {
		t.Fatalf("capToolResult() = %q, want hello prefix", got)
	}
	if !strings.Contains(got, wantSuffix) {
		t.Fatalf("capToolResult() = %q, want suffix %q", got, wantSuffix)
	}
}

type recordingLogger struct {
	events []TurnEvent
}

func (l *recordingLogger) LogTurn(ctx context.Context, event TurnEvent) {
	l.events = append(l.events, event)
}

type toolCallingProvider struct {
	calls int
}

func (p *toolCallingProvider) Name() string { return "tool-calling" }

func (p *toolCallingProvider) Complete(ctx context.Context, req *llm.Request) (*llm.Response, error) {
	p.calls++
	if p.calls == 1 {
		return &llm.Response{Content: []llm.Block{{
			Type:      llm.BlockToolUse,
			ToolUseID: "tool-1",
			ToolName:  "echo",
		}}}, nil
	}
	return &llm.Response{Content: []llm.Block{{Type: llm.BlockText, Text: "done"}}}, nil
}

type echoTool struct{}

func (echoTool) Name() string { return "echo" }

func (echoTool) Description() string { return "echo" }

func (echoTool) Schema() json.RawMessage { return json.RawMessage(`{"type":"object"}`) }

func (echoTool) Run(ctx context.Context, input json.RawMessage) (string, error) {
	return "hello world", nil
}

func TestChatLogsTurnEvent(t *testing.T) {
	logger := &recordingLogger{}
	reg := tool.NewRegistry()
	reg.Register(echoTool{})
	a, err := New(Options{
		Name:          "test",
		Model:         "test-model",
		SystemPrompt:  "test",
		Provider:      &toolCallingProvider{},
		Tools:         reg,
		ToolResultCap: 5,
		Logger:        logger,
	})
	if err != nil {
		t.Fatal(err)
	}

	reply, err := a.Chat(context.Background(), "lucas", "hello")
	if err != nil {
		t.Fatal(err)
	}
	if reply != "done" {
		t.Fatalf("reply = %q, want done", reply)
	}
	if len(logger.events) != 1 {
		t.Fatalf("events = %#v, want one event", logger.events)
	}
	event := logger.events[0]
	if event.Agent != "test" || event.UserID != "lucas" || event.Model != "test-model" {
		t.Fatalf("event identity = %#v", event)
	}
	if event.FinishReason != "stop" {
		t.Fatalf("finish reason = %q, want stop", event.FinishReason)
	}
	if event.Iterations != 2 {
		t.Fatalf("iterations = %d, want 2", event.Iterations)
	}
	if len(event.Tools) != 1 || event.Tools[0] != "echo" {
		t.Fatalf("tools = %#v, want echo", event.Tools)
	}
	if !event.Truncated {
		t.Fatal("event did not report truncated tool result")
	}
}

type concurrentProvider struct {
	mu        sync.Mutex
	active    int
	maxActive int
}

func (p *concurrentProvider) Name() string { return "concurrent" }

func (p *concurrentProvider) Complete(ctx context.Context, req *llm.Request) (*llm.Response, error) {
	p.mu.Lock()
	p.active++
	if p.active > p.maxActive {
		p.maxActive = p.active
	}
	p.mu.Unlock()

	time.Sleep(10 * time.Millisecond)

	p.mu.Lock()
	p.active--
	p.mu.Unlock()

	return &llm.Response{
		Content: []llm.Block{{Type: llm.BlockText, Text: "ok"}},
	}, nil
}

func TestChatSerializesTurnsPerUser(t *testing.T) {
	provider := &concurrentProvider{}
	a, err := New(Options{
		Name:         "test",
		Model:        "test-model",
		SystemPrompt: "test",
		Provider:     provider,
	})
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := a.Chat(context.Background(), "lucas", "hello"); err != nil {
				t.Errorf("Chat returned error: %v", err)
			}
		}()
	}
	wg.Wait()

	if provider.maxActive != 1 {
		t.Fatalf("max concurrent provider calls = %d, want 1", provider.maxActive)
	}
}

type blockingProvider struct {
	entered chan struct{}
	release chan struct{}
}

func (p *blockingProvider) Name() string { return "blocking" }

func (p *blockingProvider) Complete(ctx context.Context, req *llm.Request) (*llm.Response, error) {
	close(p.entered)
	select {
	case <-p.release:
		return &llm.Response{Content: []llm.Block{{Type: llm.BlockText, Text: "ok"}}}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func TestChatWaitForUserTurnRespectsContext(t *testing.T) {
	provider := &blockingProvider{
		entered: make(chan struct{}),
		release: make(chan struct{}),
	}
	a, err := New(Options{
		Name:         "test",
		Model:        "test-model",
		SystemPrompt: "test",
		Provider:     provider,
	})
	if err != nil {
		t.Fatal(err)
	}

	done := make(chan error, 1)
	go func() {
		_, err := a.Chat(context.Background(), "lucas", "first")
		done <- err
	}()

	<-provider.entered

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	if _, err := a.Chat(ctx, "lucas", "second"); err == nil {
		t.Fatal("expected context error while waiting for user turn")
	}

	close(provider.release)
	if err := <-done; err != nil {
		t.Fatalf("first Chat returned error: %v", err)
	}
}
