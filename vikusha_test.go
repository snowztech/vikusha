package vikusha

import (
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
