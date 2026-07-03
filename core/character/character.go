package character

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Character struct {
	Name         string         `yaml:"name"`
	Model        string         `yaml:"model"`
	SystemPrompt string         `yaml:"system_prompt"`
	Provider     ProviderConfig `yaml:"provider"`
	Tools        []string       `yaml:"tools"`
}

type ProviderConfig struct {
	Name      string `yaml:"name"`
	APIKeyEnv string `yaml:"api_key_env"`
	BaseURL   string `yaml:"base_url"`
}

func Load(path string) (*Character, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var c Character
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("parse character: %w", err)
	}
	if errs := c.Validate(); len(errs) > 0 {
		return nil, fmt.Errorf("invalid character:\n- %s", strings.Join(errs, "\n- "))
	}
	return &c, nil
}

func (c Character) Validate() []string {
	var errs []string
	if strings.TrimSpace(c.Name) == "" {
		errs = append(errs, "name is required")
	}
	if strings.TrimSpace(c.Model) == "" {
		errs = append(errs, "model is required")
	}
	if strings.TrimSpace(c.SystemPrompt) == "" {
		errs = append(errs, "system_prompt is required")
	}
	if c.Provider.Name != "" && providerName(c.Provider.Name) == "" {
		errs = append(errs, fmt.Sprintf("provider.name %q is not supported", c.Provider.Name))
	}
	for _, t := range c.Tools {
		if strings.TrimSpace(t) == "" {
			errs = append(errs, "tools cannot contain empty names")
		}
	}
	return errs
}

func (c Character) ProviderName() string {
	if name := providerName(c.Provider.Name); name != "" {
		return name
	}
	model := strings.ToLower(c.Model)
	switch {
	case strings.HasPrefix(model, "claude"):
		return "anthropic"
	default:
		return "openai"
	}
}

func (c Character) APIKeyEnv() string {
	if c.Provider.APIKeyEnv != "" {
		return c.Provider.APIKeyEnv
	}
	switch c.ProviderName() {
	case "anthropic":
		return "ANTHROPIC_API_KEY"
	default:
		return "OPENAI_API_KEY"
	}
}

func providerName(name string) string {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "":
		return ""
	case "anthropic":
		return "anthropic"
	case "openai", "openai-compatible", "openai_compatible", "openrouter", "groq", "lmstudio", "lm-studio":
		return "openai"
	default:
		return ""
	}
}
