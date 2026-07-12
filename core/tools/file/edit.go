package file

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Edit struct {
	scope scope
}

func NewEdit(root ...string) *Edit {
	e := &Edit{}
	if len(root) > 0 {
		e.scope = newScope(root[0])
	}
	return e
}

func (e *Edit) Name() string { return "file_edit" }

func (e *Edit) Description() string {
	return "Create or replace a text file at the given path. Paths are scoped to the agent workspace when configured."
}

func (e *Edit) Schema() json.RawMessage {
	return json.RawMessage(`{
  "type": "object",
  "properties": {
    "path": {
      "type": "string",
      "description": "File path to create or replace."
    },
    "content": {
      "type": "string",
      "description": "Full file content to write."
    },
    "create_dirs": {
      "type": "boolean",
      "description": "Create missing parent directories inside the workspace."
    }
  },
  "required": ["path", "content"]
}`)
}

type editInput struct {
	Path       string `json:"path"`
	Content    string `json:"content"`
	CreateDirs bool   `json:"create_dirs"`
}

func (e *Edit) Run(ctx context.Context, input json.RawMessage) (string, error) {
	var in editInput
	if err := json.Unmarshal(input, &in); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}
	if in.Path == "" {
		return "", fmt.Errorf("invalid input: path is required")
	}

	path, err := e.resolvePath(in)
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(path, []byte(in.Content), 0o600); err != nil {
		return "", err
	}
	return fmt.Sprintf("wrote %d bytes to %s", len(in.Content), path), nil
}

func (e *Edit) resolvePath(in editInput) (string, error) {
	if in.CreateDirs {
		parent, err := e.scope.resolveForMkdir(filepath.Dir(in.Path))
		if err != nil {
			return "", err
		}
		if err := os.MkdirAll(parent, 0o700); err != nil {
			return "", err
		}
	}
	return e.scope.resolveForWrite(in.Path)
}
