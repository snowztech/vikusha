package character

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Character struct {
	Name         string                `yaml:"name"`
	Model        string                `yaml:"model"`
	SystemPrompt string                `yaml:"system_prompt"`
	Provider     ProviderConfig        `yaml:"provider"`
	Memory       MemoryConfig          `yaml:"memory"`
	Context      ContextConfig         `yaml:"context"`
	Tools        []string              `yaml:"tools"`
	ToolConfig   map[string]ToolConfig `yaml:"tool_config"`
	Logging      LoggingConfig         `yaml:"logging"`
}

type ProviderConfig struct {
	Name      string `yaml:"name"`
	APIKeyEnv string `yaml:"api_key_env"`
	BaseURL   string `yaml:"base_url"`
}

type MemoryConfig struct {
	Backend string `yaml:"backend"`
	Path    string `yaml:"path"`
}

type ContextConfig struct {
	HistoryTokenBudget int `yaml:"history_token_budget"`
}

type ToolConfig struct {
	Timeout   string `yaml:"timeout"`
	ResultCap int    `yaml:"result_cap"`
}

type LoggingConfig struct {
	JSON     bool `yaml:"json"`
	Terminal bool `yaml:"terminal"`
	Color    bool `yaml:"color"`
}

func Load(path string) (*Character, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var c Character
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(&c); err != nil {
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
	if c.Memory.Backend != "" && memoryBackend(c.Memory.Backend) == "" {
		errs = append(errs, fmt.Sprintf("memory.backend %q is not supported", c.Memory.Backend))
	}
	if c.Context.HistoryTokenBudget < 0 {
		errs = append(errs, "context.history_token_budget cannot be negative")
	}
	if c.Logging.JSON && c.Logging.Terminal {
		errs = append(errs, "logging.json and logging.terminal cannot both be true")
	}
	for name, cfg := range c.ToolConfig {
		toolName := strings.TrimSpace(name)
		if toolName == "" {
			errs = append(errs, "tool_config cannot contain empty tool names")
		}
		if strings.TrimSpace(cfg.Timeout) != "" {
			if _, err := time.ParseDuration(cfg.Timeout); err != nil {
				errs = append(errs, fmt.Sprintf("tool_config.%s.timeout is invalid: %v", toolName, err))
			}
		}
		if cfg.ResultCap < 0 {
			errs = append(errs, fmt.Sprintf("tool_config.%s.result_cap cannot be negative", toolName))
		}
	}
	for _, t := range c.Tools {
		if strings.TrimSpace(t) == "" {
			errs = append(errs, "tools cannot contain empty names")
		}
	}
	return errs
}

func (c Character) MemoryBackend() string {
	return memoryBackend(c.Memory.Backend)
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

func memoryBackend(name string) string {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "":
		return ""
	case "file":
		return "file"
	default:
		return ""
	}
}

func (c Character) APIKeyEnv() string {
	if c.Provider.APIKeyEnv != "" {
		return c.Provider.APIKeyEnv
	}
	switch strings.ToLower(strings.TrimSpace(c.Provider.Name)) {
	case "openrouter":
		return "OPENROUTER_API_KEY"
	case "groq":
		return "GROQ_API_KEY"
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
