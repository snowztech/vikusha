package vikusha

import (
	"os"
	"path/filepath"
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
