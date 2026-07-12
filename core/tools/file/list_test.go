package file

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestListScopedListsWorkspaceDirectory(t *testing.T) {
	workspace := t.TempDir()
	if err := os.Mkdir(filepath.Join(workspace, "docs"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(workspace, "note.txt"), []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := NewList(workspace).Run(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}

	var entries []listEntry
	if err := json.Unmarshal([]byte(got), &entries); err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("entries = %#v, want 2 entries", entries)
	}
	if entries[0] != (listEntry{Name: "docs", IsDir: true}) {
		t.Fatalf("entries[0] = %#v, want docs directory", entries[0])
	}
	if entries[1] != (listEntry{Name: "note.txt", IsDir: false}) {
		t.Fatalf("entries[1] = %#v, want note file", entries[1])
	}
}

func TestListScopedRejectsParentEscape(t *testing.T) {
	base := t.TempDir()
	workspace := filepath.Join(base, "workspace")
	if err := os.Mkdir(workspace, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(base, "outside"), 0o700); err != nil {
		t.Fatal(err)
	}

	_, err := NewList(workspace).Run(context.Background(), readJSON("../outside"))
	if err == nil {
		t.Fatal("expected workspace escape error")
	}
	if !strings.Contains(err.Error(), "outside workspace") {
		t.Fatalf("error = %q, want outside workspace", err)
	}
}

func TestListScopedRejectsSymlinkEscape(t *testing.T) {
	base := t.TempDir()
	workspace := filepath.Join(base, "workspace")
	if err := os.Mkdir(workspace, 0o700); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(base, "outside")
	if err := os.Mkdir(outside, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(workspace, "outside-link")); err != nil {
		t.Fatal(err)
	}

	_, err := NewList(workspace).Run(context.Background(), readJSON("outside-link"))
	if err == nil {
		t.Fatal("expected workspace escape error")
	}
}
