package vikusha

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/snowztech/vikusha/core/character"
	"github.com/snowztech/vikusha/core/tool"
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

	a, err := LoadAgent(path, BuildOptions{
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

func TestBuildAgentFromCharacter(t *testing.T) {
	a, err := newAgent(&character.Character{
		Name:         "Helper",
		Model:        "gpt-4o-mini",
		SystemPrompt: "Be useful.",
		Provider: character.ProviderConfig{
			Name:      "openai",
			APIKeyEnv: "OPENAI_API_KEY",
		},
		Tools: []string{"file_list", "file_read"},
	}, BuildOptions{
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
	if a.Name() != "Helper" {
		t.Fatalf("agent name = %q, want Helper", a.Name())
	}
}

type customTool struct{}

func (customTool) Name() string { return "custom" }

func (customTool) Description() string { return "custom tool" }

func (customTool) Schema() json.RawMessage { return json.RawMessage(`{"type":"object"}`) }

func (customTool) Run(ctx context.Context, input json.RawMessage) (string, error) {
	return "ok", nil
}

func TestBuildAgentUsesAvailableTools(t *testing.T) {
	a, err := newAgent(&character.Character{
		Name:         "Helper",
		Model:        "gpt-4o-mini",
		SystemPrompt: "Be useful.",
		Tools:        []string{"custom"},
	}, BuildOptions{
		Env: func(name string) string {
			if name == "OPENAI_API_KEY" {
				return "test-key"
			}
			return ""
		},
		AvailableTools: map[string]tool.Tool{
			"custom": customTool{},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if a.Name() != "Helper" {
		t.Fatalf("agent name = %q, want Helper", a.Name())
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

	_, err := LoadAgent(path, BuildOptions{
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

func TestAgentToolConfigConvertsCharacterToolConfig(t *testing.T) {
	cfg, err := agentToolConfig(map[string]character.ToolConfig{
		" file_read ": {Timeout: "2s", ResultCap: 8000},
	})
	if err != nil {
		t.Fatal(err)
	}
	got := cfg["file_read"]
	if got.Timeout != 2*time.Second {
		t.Fatalf("timeout = %s, want 2s", got.Timeout)
	}
	if got.ResultCap != 8000 {
		t.Fatalf("result cap = %d, want 8000", got.ResultCap)
	}
}

func TestAgentToolConfigRejectsInvalidTimeout(t *testing.T) {
	_, err := agentToolConfig(map[string]character.ToolConfig{
		"file_read": {Timeout: "soon"},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "tool_config.file_read.timeout") {
		t.Fatalf("error = %q, want tool_config timeout", err)
	}
}

func TestRegistryScopesBuiltInFileToolsToWorkspace(t *testing.T) {
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

	reg, err := registry([]string{"file_list", "file_read", "file_edit"}, BuildOptions{Workspace: workspace})
	if err != nil {
		t.Fatal(err)
	}
	list, ok := reg.Get("file_list")
	if !ok {
		t.Fatal("file_list not registered")
	}
	if got, err := list.Run(context.Background(), mustJSON(t, map[string]string{"path": "."})); err != nil {
		t.Fatal(err)
	} else if !strings.Contains(got, "note.txt") {
		t.Fatalf("file_list = %q, want note.txt", got)
	}
	if _, err := list.Run(context.Background(), mustJSON(t, map[string]string{"path": "../secret.txt"})); err == nil {
		t.Fatal("expected workspace escape error from file_list")
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

	edit, ok := reg.Get("file_edit")
	if !ok {
		t.Fatal("file_edit not registered")
	}
	if _, err := edit.Run(context.Background(), mustJSON(t, map[string]any{"path": "new.txt", "content": "new"})); err != nil {
		t.Fatal(err)
	}
	if got, err := os.ReadFile(filepath.Join(workspace, "new.txt")); err != nil {
		t.Fatal(err)
	} else if string(got) != "new" {
		t.Fatalf("file_edit wrote %q, want new", got)
	}
	if _, err := edit.Run(context.Background(), mustJSON(t, map[string]any{"path": "../secret.txt", "content": "oops"})); err == nil {
		t.Fatal("expected workspace escape error from file_edit")
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
