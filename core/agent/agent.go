package agent

import (
	"fmt"

	"github.com/snowztech/vikusha/core/llm"
	"github.com/snowztech/vikusha/core/memory"
	"github.com/snowztech/vikusha/core/tool"
)

const maxIterations = 10

type Agent struct {
	name         string
	model        string
	systemPrompt string
	provider     llm.Provider
	tools        *tool.Registry
	memory       memory.Memory
}

type Options struct {
	Name         string
	Model        string
	SystemPrompt string
	Provider     llm.Provider
	Tools        *tool.Registry
	Memory       memory.Memory
}

func New(opts Options) (*Agent, error) {
	if opts.Name == "" {
		return nil, fmt.Errorf("agent: Name is required")
	}
	if opts.Model == "" {
		return nil, fmt.Errorf("agent: Model is required")
	}
	if opts.Provider == nil {
		return nil, fmt.Errorf("agent: Provider is required")
	}
	if opts.Tools == nil {
		opts.Tools = tool.NewRegistry()
	}
	return &Agent{
		name:         opts.Name,
		model:        opts.Model,
		systemPrompt: opts.SystemPrompt,
		provider:     opts.Provider,
		tools:        opts.Tools,
		memory:       opts.Memory,
	}, nil
}

func (a *Agent) Name() string { return a.name }
