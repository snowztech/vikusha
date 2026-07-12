package file

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEditWritesFileInsideWorkspace(t *testing.T) {
	workspace := t.TempDir()
	edit := NewEdit(workspace)

	out, err := edit.Run(context.Background(), editJSON(t, map[string]any{
		"path":    "note.txt",
		"content": "hello",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "wrote 5 bytes") {
		t.Fatalf("output = %q, want bytes written", out)
	}
	data, err := os.ReadFile(filepath.Join(workspace, "note.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "hello" {
		t.Fatalf("file = %q, want hello", data)
	}
}

func TestEditCreatesNestedDirectoriesWhenRequested(t *testing.T) {
	workspace := t.TempDir()
	edit := NewEdit(workspace)

	_, err := edit.Run(context.Background(), editJSON(t, map[string]any{
		"path":        "notes/today/note.txt",
		"content":     "hello",
		"create_dirs": true,
	}))
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(workspace, "notes", "today", "note.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "hello" {
		t.Fatalf("file = %q, want hello", data)
	}
}

func TestEditRejectsMissingParentWithoutCreateDirs(t *testing.T) {
	workspace := t.TempDir()
	edit := NewEdit(workspace)

	_, err := edit.Run(context.Background(), editJSON(t, map[string]any{
		"path":    "missing/note.txt",
		"content": "hello",
	}))
	if err == nil {
		t.Fatal("expected missing parent error")
	}
}

func TestEditRejectsWorkspaceEscape(t *testing.T) {
	workspace := t.TempDir()
	outside := filepath.Join(t.TempDir(), "secret.txt")
	edit := NewEdit(workspace)

	_, err := edit.Run(context.Background(), editJSON(t, map[string]any{
		"path":    outside,
		"content": "secret",
	}))
	if err == nil {
		t.Fatal("expected workspace escape error")
	}
	if _, statErr := os.Stat(outside); !os.IsNotExist(statErr) {
		t.Fatalf("outside file exists or stat failed unexpectedly: %v", statErr)
	}
}

func TestEditRejectsSymlinkEscapeParent(t *testing.T) {
	workspace := t.TempDir()
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(workspace, "link")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	edit := NewEdit(workspace)

	_, err := edit.Run(context.Background(), editJSON(t, map[string]any{
		"path":    "link/secret.txt",
		"content": "secret",
	}))
	if err == nil {
		t.Fatal("expected symlink escape error")
	}
}

func editJSON(t *testing.T, v any) json.RawMessage {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
