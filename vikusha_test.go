package vikusha

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadAgentFromCharacter(t *testing.T) {
	path := filepath.Join(t.TempDir(), "character.yaml")
	if err := os.WriteFile(path, []byte(`
name: Reader
model: gpt-4o-mini
system_prompt: Read files when useful.
tools:
  - file_read
`), 0o600); err != nil {
		t.Fatal(err)
	}

	a, err := LoadAgent(path, Options{
		Env: func(name string) string {
			if name == "OPENAI_API_KEY" {
				return "test-key"
			}
			return ""
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if a.Name() != "Reader" {
		t.Fatalf("agent name = %q, want Reader", a.Name())
	}
}

func TestLoadAgentRequiresFileMemoryPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "character.yaml")
	if err := os.WriteFile(path, []byte(`
name: Reader
model: gpt-4o-mini
system_prompt: Read files when useful.
memory:
  backend: file
`), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := LoadAgent(path, Options{
		Env: func(name string) string {
			if name == "OPENAI_API_KEY" {
				return "test-key"
			}
			return ""
		},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "memory.path") {
		t.Fatalf("error = %q, want memory.path", err)
	}
}

func TestRegistryScopesBuiltInFileReadToWorkspace(t *testing.T) {
	base := t.TempDir()
	workspace := filepath.Join(base, "workspace")
	if err := os.Mkdir(workspace, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(workspace, "note.txt"), []byte("inside"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(base, "secret.txt"), []byte("outside"), 0o600); err != nil {
		t.Fatal(err)
	}

	reg, err := registry([]string{"file_read"}, Options{Workspace: workspace})
	if err != nil {
		t.Fatal(err)
	}
	read, ok := reg.Get("file_read")
	if !ok {
		t.Fatal("file_read not registered")
	}

	got, err := read.Run(context.Background(), mustJSON(t, map[string]string{"path": "note.txt"}))
	if err != nil {
		t.Fatal(err)
	}
	if got != "inside" {
		t.Fatalf("file_read = %q, want inside", got)
	}
	if _, err := read.Run(context.Background(), mustJSON(t, map[string]string{"path": "../secret.txt"})); err == nil {
		t.Fatal("expected workspace escape error")
	}
}

func mustJSON(t *testing.T, v any) json.RawMessage {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
