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
