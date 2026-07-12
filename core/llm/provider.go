package llm

import (
	"context"
	"encoding/json"
)

type Provider interface {
	Name() string
	Complete(ctx context.Context, req *Request) (*Response, error)
}

type Request struct {
	Model     string
	System    string
	Messages  []Message
	Tools     []ToolDef
	MaxTokens int
}

type Message struct {
	Role    string
	Content []Block
}

type ToolDef struct {
	Name        string
	Description string
	Schema      json.RawMessage
}

type Block struct {
	Type      BlockType
	Text      string
	ToolUseID string
	ToolName  string
	ToolInput json.RawMessage
	ToolError bool
}

type BlockType string

const (
	BlockText       BlockType = "text"
	BlockToolUse    BlockType = "tool_use"
	BlockToolResult BlockType = "tool_result"
)

type Response struct {
	Content    []Block
	StopReason string
	Usage      Usage
}

type Usage struct {
	InputTokens           int
	OutputTokens          int
	CacheReadTokens       int
	CacheWriteTokens      int
	ReasoningOutputTokens int
}
