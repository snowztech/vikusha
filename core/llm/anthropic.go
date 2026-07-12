package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const anthropicEndpoint = "https://api.anthropic.com/v1/messages"
const anthropicVersion = "2023-06-01"

type Anthropic struct {
	apiKey string
	http   *http.Client
}

func NewAnthropic(apiKey string) *Anthropic {
	return &Anthropic{apiKey: apiKey, http: http.DefaultClient}
}

func (a *Anthropic) Name() string { return "anthropic" }

func (a *Anthropic) Complete(ctx context.Context, req *Request) (*Response, error) {
	body, err := json.Marshal(toAnthropicReq(req))
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	raw, err := postJSONWithRetry(ctx, a.http, anthropicEndpoint, map[string]string{
		"content-type":      "application/json",
		"x-api-key":         a.apiKey,
		"anthropic-version": anthropicVersion,
	}, body, "anthropic")
	if err != nil {
		return nil, err
	}

	var wire anthropicResp
	if err := json.Unmarshal(raw, &wire); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return fromAnthropicResp(&wire), nil
}

// wire types

type anthropicReq struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	System    string             `json:"system,omitempty"`
	Messages  []anthropicMessage `json:"messages"`
	Tools     []anthropicTool    `json:"tools,omitempty"`
}

type anthropicMessage struct {
	Role    string           `json:"role"`
	Content []anthropicBlock `json:"content"`
}

type anthropicBlock struct {
	Type      string          `json:"type"`
	Text      string          `json:"text,omitempty"`
	ID        string          `json:"id,omitempty"`
	Name      string          `json:"name,omitempty"`
	Input     json.RawMessage `json:"input,omitempty"`
	ToolUseID string          `json:"tool_use_id,omitempty"`
	Content   string          `json:"content,omitempty"`
	IsError   bool            `json:"is_error,omitempty"`
}

type anthropicTool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"input_schema"`
}

type anthropicResp struct {
	Content    []anthropicBlock `json:"content"`
	StopReason string           `json:"stop_reason"`
	Usage      struct {
		InputTokens              int `json:"input_tokens"`
		OutputTokens             int `json:"output_tokens"`
		CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
		CacheReadInputTokens     int `json:"cache_read_input_tokens"`
	} `json:"usage"`
}

func toAnthropicReq(r *Request) anthropicReq {
	out := anthropicReq{
		Model:     r.Model,
		MaxTokens: r.MaxTokens,
		System:    r.System,
		Messages:  make([]anthropicMessage, len(r.Messages)),
		Tools:     make([]anthropicTool, len(r.Tools)),
	}
	if out.MaxTokens == 0 {
		out.MaxTokens = 1024
	}
	for i, m := range r.Messages {
		out.Messages[i] = anthropicMessage{Role: m.Role, Content: encodeBlocks(m.Content)}
	}
	for i, t := range r.Tools {
		out.Tools[i] = anthropicTool{Name: t.Name, Description: t.Description, InputSchema: t.Schema}
	}
	return out
}

func encodeBlocks(blocks []Block) []anthropicBlock {
	out := make([]anthropicBlock, len(blocks))
	for i, b := range blocks {
		switch b.Type {
		case BlockText:
			out[i] = anthropicBlock{Type: "text", Text: b.Text}
		case BlockToolUse:
			out[i] = anthropicBlock{Type: "tool_use", ID: b.ToolUseID, Name: b.ToolName, Input: b.ToolInput}
		case BlockToolResult:
			out[i] = anthropicBlock{Type: "tool_result", ToolUseID: b.ToolUseID, Content: b.Text, IsError: b.ToolError}
		}
	}
	return out
}

func fromAnthropicResp(w *anthropicResp) *Response {
	out := &Response{
		StopReason: w.StopReason,
		Content:    make([]Block, len(w.Content)),
		Usage: Usage{
			InputTokens:      w.Usage.InputTokens,
			OutputTokens:     w.Usage.OutputTokens,
			CacheReadTokens:  w.Usage.CacheReadInputTokens,
			CacheWriteTokens: w.Usage.CacheCreationInputTokens,
		},
	}
	for i, b := range w.Content {
		switch b.Type {
		case "text":
			out.Content[i] = Block{Type: BlockText, Text: b.Text}
		case "tool_use":
			out.Content[i] = Block{Type: BlockToolUse, ToolUseID: b.ID, ToolName: b.Name, ToolInput: b.Input}
		}
	}
	return out
}
