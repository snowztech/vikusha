package file

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/snowztech/vikusha/core/memory"
)

func TestSaveLoadAndSearch(t *testing.T) {
	ctx := context.Background()
	store := New(t.TempDir())

	first := memory.Entry{
		Type:      memory.EntryPreference,
		Content:   "Lucas prefers concise answers.",
		CreatedAt: time.Date(2026, 7, 12, 10, 0, 0, 0, time.UTC),
	}
	second := memory.Entry{
		Type:    memory.EntryFact,
		Content: "Lucas is building Vikusha.",
	}
	if err := store.Save(ctx, "lucas", first); err != nil {
		t.Fatal(err)
	}
	if err := store.Save(ctx, "lucas", second); err != nil {
		t.Fatal(err)
	}

	entries, err := store.Load(ctx, "lucas")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("Load returned %d entries, want 2", len(entries))
	}
	if entries[0] != first {
		t.Fatalf("first entry = %#v, want %#v", entries[0], first)
	}
	if entries[1].CreatedAt.IsZero() {
		t.Fatal("Save should fill CreatedAt when missing")
	}

	found, err := store.Search(ctx, "lucas", "vikusha", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(found) != 1 || found[0].Content != second.Content {
		t.Fatalf("Search returned %#v, want second entry", found)
	}
}

func TestLoadMissingUserReturnsEmpty(t *testing.T) {
	entries, err := New(t.TempDir()).Load(context.Background(), "missing")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("Load returned %d entries, want 0", len(entries))
	}
}

func TestUserIDIsSafeFilename(t *testing.T) {
	dir := t.TempDir()
	store := New(dir)
	if err := store.Save(context.Background(), "../lucas/slash", memory.Entry{
		Type:    memory.EntryNote,
		Content: "safe",
	}); err != nil {
		t.Fatal(err)
	}

	got, err := filepath.Glob(filepath.Join(dir, "*.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("created files = %v, want one jsonl file", got)
	}
	if filepath.Dir(got[0]) != dir {
		t.Fatalf("memory file escaped dir: %s", got[0])
	}
}

func TestInvalidEntryTypeFails(t *testing.T) {
	err := New(t.TempDir()).Save(context.Background(), "lucas", memory.Entry{
		Type:    "unknown",
		Content: "bad",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestInvalidJSONLineFails(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "lucas.jsonl")
	if err := os.WriteFile(path, []byte("{bad json}\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := New(dir).Load(context.Background(), "lucas")
	if err == nil {
		t.Fatal("expected error")
	}
}
