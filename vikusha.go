package vikusha

import (
	"fmt"
	"os"
	"strings"

	"github.com/snowztech/vikusha/core/agent"
	"github.com/snowztech/vikusha/core/character"
	"github.com/snowztech/vikusha/core/llm"
	"github.com/snowztech/vikusha/core/tool"
	"github.com/snowztech/vikusha/core/tools/file"
)

type Options struct {
	Env   func(string) string
	Tools map[string]tool.Tool
}

func LoadAgent(path string, opts Options) (*agent.Agent, error) {
	c, err := character.Load(path)
	if err != nil {
		return nil, err
	}
	return NewAgent(c, opts)
}

func NewAgent(c *character.Character, opts Options) (*agent.Agent, error) {
	p, err := provider(c, opts)
	if err != nil {
		return nil, err
	}
	reg, err := registry(c.Tools, opts)
	if err != nil {
		return nil, err
	}
	return agent.New(agent.Options{
		Name:         c.Name,
		Model:        c.Model,
		SystemPrompt: c.SystemPrompt,
		Provider:     p,
		Tools:        reg,
	})
}

func provider(c *character.Character, opts Options) (llm.Provider, error) {
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
		return llm.NewOpenAI(apiKey), nil
	default:
		return nil, fmt.Errorf("unsupported provider %q", c.ProviderName())
	}
}

func registry(names []string, opts Options) (*tool.Registry, error) {
	reg := tool.NewRegistry()
	available := map[string]tool.Tool{
		"file_read": file.NewRead(),
	}
	for name, t := range opts.Tools {
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
