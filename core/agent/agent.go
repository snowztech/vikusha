package agent

import (
	"context"
	"fmt"
	"strings"

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

func (a *Agent) Chat(ctx context.Context, userID, msg string) (string, error) {
	system, err := a.systemWithMemory(ctx, userID)
	if err != nil {
		return "", err
	}
	msgs := []llm.Message{
		{Role: "user", Content: []llm.Block{{Type: llm.BlockText, Text: msg}}},
	}
	tools := a.toolDefs()

	for range maxIterations {
		resp, err := a.provider.Complete(ctx, &llm.Request{
			Model:    a.model,
			System:   system,
			Messages: msgs,
			Tools:    tools,
		})
		if err != nil {
			return "", fmt.Errorf("provider: %w", err)
		}

		text, toolCalls := splitBlocks(resp.Content)
		if len(toolCalls) == 0 {
			return text, nil
		}

		msgs = append(msgs, llm.Message{Role: "assistant", Content: resp.Content})

		results := make([]llm.Block, 0, len(toolCalls))
		for _, call := range toolCalls {
			results = append(results, a.runTool(ctx, call))
		}
		msgs = append(msgs, llm.Message{Role: "user", Content: results})
	}

	return "", fmt.Errorf("agent: hit max iterations (%d)", maxIterations)
}

func (a *Agent) systemWithMemory(ctx context.Context, userID string) (string, error) {
	if a.memory == nil {
		return a.systemPrompt, nil
	}
	entries, err := a.memory.Search(ctx, userID, "", 20)
	if err != nil {
		return "", fmt.Errorf("memory: %w", err)
	}
	if len(entries) == 0 {
		return a.systemPrompt, nil
	}

	var b strings.Builder
	b.WriteString(a.systemPrompt)
	b.WriteString("\n\nRelevant memory for this user:\n")
	for i := len(entries) - 1; i >= 0; i-- {
		b.WriteString("- ")
		b.WriteString(string(entries[i].Type))
		b.WriteString(": ")
		b.WriteString(entries[i].Content)
		b.WriteByte('\n')
	}
	return strings.TrimRight(b.String(), "\n"), nil
}

func (a *Agent) toolDefs() []llm.ToolDef {
	if a.tools == nil {
		return nil
	}
	all := a.tools.All()
	defs := make([]llm.ToolDef, len(all))
	for i, t := range all {
		defs[i] = llm.ToolDef{Name: t.Name(), Description: t.Description(), Schema: t.Schema()}
	}
	return defs
}

func (a *Agent) runTool(ctx context.Context, call llm.Block) (out llm.Block) {
	defer func() {
		if r := recover(); r != nil {
			out = errResult(call.ToolUseID, fmt.Sprintf("tool panic: %v", r))
		}
	}()

	t, ok := a.tools.Get(call.ToolName)
	if !ok {
		return errResult(call.ToolUseID, fmt.Sprintf("tool not found: %s", call.ToolName))
	}
	output, err := t.Run(ctx, call.ToolInput)
	if err != nil {
		return errResult(call.ToolUseID, err.Error())
	}
	return llm.Block{Type: llm.BlockToolResult, ToolUseID: call.ToolUseID, Text: output}
}

func splitBlocks(blocks []llm.Block) (string, []llm.Block) {
	var text strings.Builder
	var calls []llm.Block
	for _, b := range blocks {
		switch b.Type {
		case llm.BlockText:
			text.WriteString(b.Text)
		case llm.BlockToolUse:
			calls = append(calls, b)
		}
	}
	return text.String(), calls
}

func errResult(toolUseID, msg string) llm.Block {
	return llm.Block{
		Type:      llm.BlockToolResult,
		ToolUseID: toolUseID,
		Text:      msg,
		ToolError: true,
	}
}
