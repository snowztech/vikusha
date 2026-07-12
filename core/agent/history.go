package agent

import (
	"encoding/json"

	"github.com/snowztech/vikusha/core/llm"
)

func (a *Agent) messagesForTurn(userID, text string) []llm.Message {
	current := llm.Message{Role: "user", Content: []llm.Block{{Type: llm.BlockText, Text: text}}}
	key := userKey(userID)

	a.historyMu.Lock()
	history := cloneMessages(a.history[key])
	a.historyMu.Unlock()

	msgs := append(history, current)
	return trimMessages(msgs, a.historyBudget)
}

func (a *Agent) saveHistory(userID string, msgs []llm.Message) {
	key := userKey(userID)
	trimmed := trimMessages(msgs, a.historyBudget)

	a.historyMu.Lock()
	if a.history == nil {
		a.history = map[string][]llm.Message{}
	}
	a.history[key] = cloneMessages(trimmed)
	a.historyMu.Unlock()
}

// trimMessages keeps the newest messages that fit the configured conversation
// history budget. The estimate is intentionally provider-agnostic; provider
// usage accounting remains the source of truth for billed tokens.
func trimMessages(msgs []llm.Message, budget int) []llm.Message {
	if budget <= 0 || len(msgs) == 0 {
		return cloneMessages(msgs)
	}

	var tokens int
	start := len(msgs)
	for i := len(msgs) - 1; i >= 0; i-- {
		next := messageTokens(msgs[i])
		if tokens+next > budget && start < len(msgs) {
			break
		}
		tokens += next
		start = i
	}
	return cloneMessages(msgs[start:])
}

func messageTokens(msg llm.Message) int {
	tokens := estimateTokens(msg.Role) + 4
	for _, block := range msg.Content {
		tokens += blockTokens(block)
	}
	return tokens
}

func blockTokens(block llm.Block) int {
	tokens := estimateTokens(string(block.Type)) + estimateTokens(block.Text)
	tokens += estimateTokens(block.ToolUseID) + estimateTokens(block.ToolName)
	tokens += estimateTokens(string(block.ToolInput))
	if block.ToolError {
		tokens++
	}
	return tokens + 2
}

func estimateTokens(s string) int {
	if s == "" {
		return 0
	}
	return (len(s) + 3) / 4
}

func cloneMessages(msgs []llm.Message) []llm.Message {
	out := make([]llm.Message, len(msgs))
	for i, msg := range msgs {
		out[i] = llm.Message{
			Role:    msg.Role,
			Content: cloneBlocks(msg.Content),
		}
	}
	return out
}

func cloneBlocks(blocks []llm.Block) []llm.Block {
	out := make([]llm.Block, len(blocks))
	for i, block := range blocks {
		out[i] = block
		if block.ToolInput != nil {
			out[i].ToolInput = append(json.RawMessage(nil), block.ToolInput...)
		}
	}
	return out
}
