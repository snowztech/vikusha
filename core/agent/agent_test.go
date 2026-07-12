package agent

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/snowztech/vikusha/core/llm"
	"github.com/snowztech/vikusha/core/memory"
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
