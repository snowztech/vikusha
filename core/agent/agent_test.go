package agent

import (
	"context"
	"strings"
	"testing"
	"time"

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
