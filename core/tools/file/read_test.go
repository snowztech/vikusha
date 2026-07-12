package file

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadUnscopedReadsPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "note.txt")
	if err := os.WriteFile(path, []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := NewRead().Run(context.Background(), readJSON(path))
	if err != nil {
		t.Fatal(err)
	}
	if got != "hello" {
		t.Fatalf("Run() = %q, want hello", got)
	}
}

func TestReadScopedReadsRelativePathFromWorkspace(t *testing.T) {
	workspace := t.TempDir()
	if err := os.WriteFile(filepath.Join(workspace, "note.txt"), []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := NewRead(workspace).Run(context.Background(), readJSON("note.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if got != "hello" {
		t.Fatalf("Run() = %q, want hello", got)
	}
}

func TestReadScopedRejectsParentEscape(t *testing.T) {
	base := t.TempDir()
	workspace := filepath.Join(base, "workspace")
	if err := os.Mkdir(workspace, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(base, "secret.txt"), []byte("secret"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := NewRead(workspace).Run(context.Background(), readJSON("../secret.txt"))
	if err == nil {
		t.Fatal("expected workspace escape error")
	}
	if !strings.Contains(err.Error(), "outside workspace") {
		t.Fatalf("error = %q, want outside workspace", err)
	}
}

func TestReadScopedRejectsAbsolutePathOutsideWorkspace(t *testing.T) {
	workspace := t.TempDir()
	outside := filepath.Join(t.TempDir(), "secret.txt")
	if err := os.WriteFile(outside, []byte("secret"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := NewRead(workspace).Run(context.Background(), readJSON(outside))
	if err == nil {
		t.Fatal("expected workspace escape error")
	}
}

func TestReadScopedRejectsSymlinkEscape(t *testing.T) {
	base := t.TempDir()
	workspace := filepath.Join(base, "workspace")
	if err := os.Mkdir(workspace, 0o700); err != nil {
		t.Fatal(err)
	}
	secret := filepath.Join(base, "secret.txt")
	if err := os.WriteFile(secret, []byte("secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(secret, filepath.Join(workspace, "secret-link")); err != nil {
		t.Fatal(err)
	}

	_, err := NewRead(workspace).Run(context.Background(), readJSON("secret-link"))
	if err == nil {
		t.Fatal("expected workspace escape error")
	}
}

func readJSON(path string) json.RawMessage {
	data, _ := json.Marshal(readInput{Path: path})
	return data
}
