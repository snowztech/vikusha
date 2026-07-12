package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type OpenAICompat struct {
	apiKey  string
	baseURL string
	http    *http.Client
}

func NewOpenAI(apiKey string) *OpenAICompat {
	return &OpenAICompat{
		apiKey:  apiKey,
		baseURL: "https://api.openai.com/v1",
		http:    http.DefaultClient,
	}
}

func NewOpenRouter(apiKey string) *OpenAICompat {
	return &OpenAICompat{
		apiKey:  apiKey,
		baseURL: "https://openrouter.ai/api/v1",
		http:    http.DefaultClient,
	}
}

func NewOpenAICompat(apiKey, baseURL string) *OpenAICompat {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	return &OpenAICompat{
		apiKey:  apiKey,
		baseURL: baseURL,
		http:    http.DefaultClient,
	}
}

func (o *OpenAICompat) Name() string { return "openai" }

func (o *OpenAICompat) Complete(ctx context.Context, req *Request) (*Response, error) {
	body, err := json.Marshal(toOpenAIReq(req))
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	raw, err := postJSONWithRetry(ctx, o.http, o.baseURL+"/chat/completions", map[string]string{
		"content-type":  "application/json",
		"authorization": "Bearer " + o.apiKey,
	}, body, "openai")
	if err != nil {
		return nil, err
	}

	var wire openaiResp
	if err := json.Unmarshal(raw, &wire); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return fromOpenAIResp(&wire), nil
}

// wire types

type openaiReq struct {
	Model     string          `json:"model"`
	Messages  []openaiMessage `json:"messages"`
	Tools     []openaiTool    `json:"tools,omitempty"`
	MaxTokens int             `json:"max_tokens,omitempty"`
}

type openaiMessage struct {
	Role       string           `json:"role"`
	Content    string           `json:"content,omitempty"`
	ToolCalls  []openaiToolCall `json:"tool_calls,omitempty"`
	ToolCallID string           `json:"tool_call_id,omitempty"`
}

type openaiToolCall struct {
	ID       string     `json:"id"`
	Type     string     `json:"type"`
	Function openaiFunc `json:"function"`
}

type openaiFunc struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type openaiTool struct {
	Type     string            `json:"type"`
	Function openaiFunctionDef `json:"function"`
}

type openaiFunctionDef struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}

type openaiResp struct {
	Choices []struct {
		Message      openaiMessage `json:"message"`
		FinishReason string        `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens         int                           `json:"prompt_tokens"`
		CompletionTokens     int                           `json:"completion_tokens"`
		PromptTokensDetails  openaiPromptTokensDetails     `json:"prompt_tokens_details"`
		CompletionTokensInfo openaiCompletionTokensDetails `json:"completion_tokens_details"`
	} `json:"usage"`
}

type openaiPromptTokensDetails struct {
	CachedTokens int `json:"cached_tokens"`
}

type openaiCompletionTokensDetails struct {
	ReasoningTokens int `json:"reasoning_tokens"`
}

func toOpenAIReq(r *Request) openaiReq {
	out := openaiReq{Model: r.Model, MaxTokens: r.MaxTokens}
	if r.System != "" {
		out.Messages = append(out.Messages, openaiMessage{Role: "system", Content: r.System})
	}
	for _, m := range r.Messages {
		out.Messages = append(out.Messages, encodeOpenAIMessage(m)...)
	}
	for _, t := range r.Tools {
		out.Tools = append(out.Tools, openaiTool{
			Type: "function",
			Function: openaiFunctionDef{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  t.Schema,
			},
		})
	}
	return out
}

// One vikusha Message can map to multiple OpenAI messages: tool_result blocks
// each become their own `role:"tool"` message, separate from user text.
func encodeOpenAIMessage(m Message) []openaiMessage {
	switch m.Role {
	case "user":
		var text string
		var toolMsgs []openaiMessage
		for _, b := range m.Content {
			switch b.Type {
			case BlockText:
				text += b.Text
			case BlockToolResult:
				toolMsgs = append(toolMsgs, openaiMessage{
					Role:       "tool",
					ToolCallID: b.ToolUseID,
					Content:    b.Text,
				})
			}
		}
		var out []openaiMessage
		if text != "" {
			out = append(out, openaiMessage{Role: "user", Content: text})
		}
		out = append(out, toolMsgs...)
		return out
	case "assistant":
		msg := openaiMessage{Role: "assistant"}
		for _, b := range m.Content {
			switch b.Type {
			case BlockText:
				msg.Content += b.Text
			case BlockToolUse:
				args := string(b.ToolInput)
				if args == "" {
					args = "{}"
				}
				msg.ToolCalls = append(msg.ToolCalls, openaiToolCall{
					ID:   b.ToolUseID,
					Type: "function",
					Function: openaiFunc{
						Name:      b.ToolName,
						Arguments: args,
					},
				})
			}
		}
		return []openaiMessage{msg}
	}
	return nil
}

func fromOpenAIResp(w *openaiResp) *Response {
	out := &Response{
		Usage: Usage{
			InputTokens:           w.Usage.PromptTokens,
			OutputTokens:          w.Usage.CompletionTokens,
			CacheReadTokens:       w.Usage.PromptTokensDetails.CachedTokens,
			ReasoningOutputTokens: w.Usage.CompletionTokensInfo.ReasoningTokens,
		},
	}
	if len(w.Choices) == 0 {
		return out
	}
	choice := w.Choices[0]
	out.StopReason = choice.FinishReason

	if choice.Message.Content != "" {
		out.Content = append(out.Content, Block{Type: BlockText, Text: choice.Message.Content})
	}
	for _, tc := range choice.Message.ToolCalls {
		args := tc.Function.Arguments
		if args == "" {
			args = "{}"
		}
		out.Content = append(out.Content, Block{
			Type:      BlockToolUse,
			ToolUseID: tc.ID,
			ToolName:  tc.Function.Name,
			ToolInput: json.RawMessage(args),
		})
	}
	return out
}
