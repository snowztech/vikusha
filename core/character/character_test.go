package character

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadValidatesAllRequiredFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "character.yaml")
	if err := os.WriteFile(path, []byte("tools:\n  - file_read\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected validation error")
	}
	msg := err.Error()
	for _, want := range []string{"name is required", "model is required", "system_prompt is required"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("expected %q in %q", want, msg)
		}
	}
}

func TestLoadInfersAnthropicProvider(t *testing.T) {
	path := filepath.Join(t.TempDir(), "character.yaml")
	if err := os.WriteFile(path, []byte(`
name: Coach
model: claude-sonnet-4-6
system_prompt: Be useful.
`), 0o600); err != nil {
		t.Fatal(err)
	}

	c, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := c.ProviderName(); got != "anthropic" {
		t.Fatalf("ProviderName() = %q, want anthropic", got)
	}
	if got := c.APIKeyEnv(); got != "ANTHROPIC_API_KEY" {
		t.Fatalf("APIKeyEnv() = %q, want ANTHROPIC_API_KEY", got)
	}
}

func TestLoadRejectsUnknownFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "character.yaml")
	if err := os.WriteFile(path, []byte(`
name: Helper
model: gpt-4o-mini
system_prompt: Be useful.
prompt: typo
`), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected unknown field error")
	}
	if !strings.Contains(err.Error(), "field prompt not found") {
		t.Fatalf("expected unknown field in error, got %q", err.Error())
	}
}

func TestLoadRejectsUnknownNestedFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "character.yaml")
	if err := os.WriteFile(path, []byte(`
name: Helper
model: gpt-4o-mini
system_prompt: Be useful.
provider:
  name: openai
  token_env: OPENAI_API_KEY
`), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected unknown field error")
	}
	if !strings.Contains(err.Error(), "field token_env not found") {
		t.Fatalf("expected unknown nested field in error, got %q", err.Error())
	}
}

func TestOpenRouterDefaultAPIKeyEnv(t *testing.T) {
	c := Character{
		Name:         "Router",
		Model:        "openai/gpt-4o-mini",
		SystemPrompt: "Be useful.",
		Provider:     ProviderConfig{Name: "openrouter"},
	}
	if got := c.ProviderName(); got != "openai" {
		t.Fatalf("ProviderName() = %q, want openai", got)
	}
	if got := c.APIKeyEnv(); got != "OPENROUTER_API_KEY" {
		t.Fatalf("APIKeyEnv() = %q, want OPENROUTER_API_KEY", got)
	}
}

func TestMemoryBackendValidation(t *testing.T) {
	c := Character{
		Name:         "Memory",
		Model:        "gpt-4o-mini",
		SystemPrompt: "Be useful.",
		Memory:       MemoryConfig{Backend: "sqlite"},
	}
	errs := c.Validate()
	if len(errs) != 1 || !strings.Contains(errs[0], "memory.backend") {
		t.Fatalf("Validate() = %#v, want memory backend error", errs)
	}
}

func TestLoadContextConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "character.yaml")
	if err := os.WriteFile(path, []byte(`
name: Helper
model: gpt-4o-mini
system_prompt: Be useful.
context:
  history_token_budget: 12000
`), 0o600); err != nil {
		t.Fatal(err)
	}

	c, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if c.Context.HistoryTokenBudget != 12000 {
		t.Fatalf("history_token_budget = %d, want 12000", c.Context.HistoryTokenBudget)
	}
}

func TestContextValidation(t *testing.T) {
	c := Character{
		Name:         "Context",
		Model:        "gpt-4o-mini",
		SystemPrompt: "Be useful.",
		Context:      ContextConfig{HistoryTokenBudget: -1},
	}
	errs := c.Validate()
	if len(errs) != 1 || !strings.Contains(errs[0], "context.history_token_budget") {
		t.Fatalf("Validate() = %#v, want context history budget error", errs)
	}
}
