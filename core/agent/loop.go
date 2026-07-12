package agent

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/snowztech/vikusha/core/llm"
)

// Chat runs one user turn through the agent loop.
//
// The loop is intentionally small: call the model, run requested tools,
// feed tool results back, and repeat until the model returns final text.
func (a *Agent) Chat(ctx context.Context, userID, msg string) (string, error) {
	release, err := a.acquireUserTurn(ctx, userID)
	if err != nil {
		return "", err
	}
	defer release()

	ctx, releaseCancel := a.acquireUserCancel(ctx, userID)
	defer releaseCancel()

	start := time.Now()
	event := turnEvent(start, a.name, userID, a.model)
	defer func() {
		duration := time.Since(start)
		event.Duration = duration.String()
		event.DurationMS = duration.Milliseconds()
		a.logTurn(ctx, event)
	}()

	system, err := a.systemWithMemory(ctx, userID)
	if err != nil {
		event.FinishReason = "error"
		event.Error = err.Error()
		return "", err
	}
	msgs := a.messagesForTurn(userID, msg)
	tools := a.toolDefs()

	for i := range maxIterations {
		event.Iterations = i + 1
		resp, err := a.provider.Complete(ctx, &llm.Request{
			Model:    a.model,
			System:   system,
			Messages: msgs,
			Tools:    tools,
		})
		if err != nil {
			event.FinishReason = "error"
			event.Error = fmt.Sprintf("provider: %v", err)
			return "", fmt.Errorf("provider: %w", err)
		}
		event.addUsage(resp.Usage)

		text, toolCalls := splitBlocks(resp.Content)
		if len(toolCalls) == 0 {
			event.FinishReason = "stop"
			a.saveHistory(userID, append(msgs, llm.Message{Role: "assistant", Content: resp.Content}))
			return text, nil
		}

		msgs = append(msgs, llm.Message{Role: "assistant", Content: resp.Content})

		results := make([]llm.Block, 0, len(toolCalls))
		for _, call := range toolCalls {
			result, truncated := a.runTool(ctx, call)
			results = append(results, result)
			event.Tools = append(event.Tools, call.ToolName)
			event.Truncated = event.Truncated || truncated
		}
		msgs = append(msgs, llm.Message{Role: "user", Content: results})
	}

	event.FinishReason = "max_iterations"
	event.Error = fmt.Sprintf("agent: hit max iterations (%d)", maxIterations)
	return "", fmt.Errorf("agent: hit max iterations (%d)", maxIterations)
}

func (a *Agent) logTurn(ctx context.Context, event TurnEvent) {
	if a.logger == nil {
		return
	}
	if event.FinishReason == "" {
		event.FinishReason = "unknown"
	}
	a.logger.LogTurn(ctx, event)
}

func (a *Agent) acquireUserTurn(ctx context.Context, userID string) (func(), error) {
	key := userKey(userID)

	// Each user gets a one-slot gate so concurrent transports cannot overlap
	// turns for the same conversation context.
	a.turnsMu.Lock()
	if a.userTurns == nil {
		a.userTurns = map[string]chan struct{}{}
	}
	turn := a.userTurns[key]
	if turn == nil {
		turn = make(chan struct{}, 1)
		a.userTurns[key] = turn
	}
	a.turnsMu.Unlock()

	select {
	case turn <- struct{}{}:
		return func() { <-turn }, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (a *Agent) acquireUserCancel(ctx context.Context, userID string) (context.Context, func()) {
	key := userKey(userID)
	turnCtx, cancel := context.WithCancel(ctx)

	a.cancelMu.Lock()
	if a.userCancels == nil {
		a.userCancels = map[string]turnCancel{}
	}
	a.nextCancelID++
	active := turnCancel{id: a.nextCancelID, cancel: cancel}
	a.userCancels[key] = active
	a.cancelMu.Unlock()

	return turnCtx, func() {
		cancel()
		a.cancelMu.Lock()
		if a.userCancels[key].id == active.id {
			delete(a.userCancels, key)
		}
		a.cancelMu.Unlock()
	}
}

func userKey(userID string) string {
	key := strings.TrimSpace(userID)
	if key == "" {
		return "default"
	}
	return key
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

func (a *Agent) runTool(ctx context.Context, call llm.Block) (out llm.Block, truncated bool) {
	defer func() {
		if r := recover(); r != nil {
			out = errResult(call.ToolUseID, fmt.Sprintf("tool panic: %v", r))
		}
	}()

	t, ok := a.tools.Get(call.ToolName)
	if !ok {
		return errResult(call.ToolUseID, fmt.Sprintf("tool not found: %s", call.ToolName)), false
	}
	output, err := t.Run(ctx, call.ToolInput)
	if err != nil {
		return errResult(call.ToolUseID, err.Error()), false
	}
	text, truncated := a.capToolResult(output)
	return llm.Block{Type: llm.BlockToolResult, ToolUseID: call.ToolUseID, Text: text}, truncated
}

func (a *Agent) capToolResult(output string) (string, bool) {
	if a.toolResultCap <= 0 || len(output) <= a.toolResultCap {
		return output, false
	}
	return output[:a.toolResultCap] + fmt.Sprintf("\n\n[tool result truncated: %d bytes omitted]", len(output)-a.toolResultCap), true
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
