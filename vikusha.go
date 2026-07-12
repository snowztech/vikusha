package vikusha

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/snowztech/vikusha/core/agent"
	"github.com/snowztech/vikusha/core/character"
	"github.com/snowztech/vikusha/core/llm"
	"github.com/snowztech/vikusha/core/memory"
	filememory "github.com/snowztech/vikusha/core/memory/file"
	"github.com/snowztech/vikusha/core/tool"
	"github.com/snowztech/vikusha/core/tools/file"
)

// BuildOptions controls how Vikusha wires character YAML into a runtime agent.
//
// The character file describes the assistant. BuildOptions describes the host
// process: env lookup, custom tool implementations, workspace, logging, and
// runtime limits.
type BuildOptions struct {
	Env            func(string) string
	AvailableTools map[string]tool.Tool
	Workspace      string
	ToolResultCap  int
	Logger         agent.TurnLogger
}

func LoadAgent(path string, opts BuildOptions) (*agent.Agent, error) {
	c, err := character.Load(path)
	if err != nil {
		return nil, err
	}
	return newAgent(c, opts)
}

func newAgent(c *character.Character, opts BuildOptions) (*agent.Agent, error) {
	p, err := provider(c, opts)
	if err != nil {
		return nil, err
	}
	reg, err := registry(c.Tools, opts)
	if err != nil {
		return nil, err
	}
	mem, err := buildMemory(c)
	if err != nil {
		return nil, err
	}
	toolConfig, err := agentToolConfig(c.ToolConfig)
	if err != nil {
		return nil, err
	}
	return agent.New(agent.Options{
		Name:               c.Name,
		Model:              c.Model,
		SystemPrompt:       c.SystemPrompt,
		Provider:           p,
		Tools:              reg,
		ToolConfig:         toolConfig,
		Memory:             mem,
		ToolResultCap:      opts.ToolResultCap,
		HistoryTokenBudget: c.Context.HistoryTokenBudget,
		Logger:             opts.Logger,
	})
}

func agentToolConfig(config map[string]character.ToolConfig) (map[string]agent.ToolConfig, error) {
	if len(config) == 0 {
		return nil, nil
	}
	out := make(map[string]agent.ToolConfig, len(config))
	for name, cfg := range config {
		var timeout time.Duration
		if strings.TrimSpace(cfg.Timeout) != "" {
			var err error
			timeout, err = time.ParseDuration(cfg.Timeout)
			if err != nil {
				return nil, fmt.Errorf("tool_config.%s.timeout is invalid: %w", strings.TrimSpace(name), err)
			}
		}
		out[strings.TrimSpace(name)] = agent.ToolConfig{
			Timeout:   timeout,
			ResultCap: cfg.ResultCap,
		}
	}
	return out, nil
}

func provider(c *character.Character, opts BuildOptions) (llm.Provider, error) {
	lookup := opts.Env
	if lookup == nil {
		lookup = os.Getenv
	}
	apiKey := lookup(c.APIKeyEnv())
	if apiKey == "" {
		return nil, fmt.Errorf("%s is not set", c.APIKeyEnv())
	}

	switch c.ProviderName() {
	case "anthropic":
		return llm.NewAnthropic(apiKey), nil
	case "openai":
		if c.Provider.BaseURL != "" {
			return llm.NewOpenAICompat(apiKey, c.Provider.BaseURL), nil
		}
		if strings.EqualFold(c.Provider.Name, "openrouter") {
			return llm.NewOpenRouter(apiKey), nil
		}
		return llm.NewOpenAI(apiKey), nil
	default:
		return nil, fmt.Errorf("unsupported provider %q", c.ProviderName())
	}
}

func buildMemory(c *character.Character) (memory.Memory, error) {
	switch c.MemoryBackend() {
	case "":
		return nil, nil
	case "file":
		if strings.TrimSpace(c.Memory.Path) == "" {
			return nil, fmt.Errorf("memory.path is required for file backend")
		}
		return filememory.New(c.Memory.Path), nil
	default:
		return nil, fmt.Errorf("unsupported memory backend %q", c.Memory.Backend)
	}
}

func registry(names []string, opts BuildOptions) (*tool.Registry, error) {
	reg := tool.NewRegistry()
	available := map[string]tool.Tool{
		"file_list": file.NewList(opts.Workspace),
		"file_read": file.NewRead(opts.Workspace),
		"file_edit": file.NewEdit(opts.Workspace),
	}
	for name, t := range opts.AvailableTools {
		available[name] = t
	}

	for _, raw := range names {
		name := strings.TrimSpace(raw)
		if name == "" {
			continue
		}
		t, ok := available[name]
		if !ok {
			return nil, fmt.Errorf("tool %q is not registered in this build", name)
		}
		reg.Register(t)
	}
	return reg, nil
}
