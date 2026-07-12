package file

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

type List struct {
	scope scope
}

func NewList(root ...string) *List {
	l := &List{}
	if len(root) > 0 {
		l.scope = newScope(root[0])
	}
	return l
}

func (l *List) Name() string { return "file_list" }

func (l *List) Description() string {
	return "List files and directories at the given path. Returns a JSON array of entries."
}

func (l *List) Schema() json.RawMessage {
	return json.RawMessage(`{
  "type": "object",
  "properties": {
    "path": {
      "type": "string",
      "description": "Directory path to list. Defaults to the workspace root when omitted."
    }
  }
}`)
}

type listInput struct {
	Path string `json:"path"`
}

type listEntry struct {
	Name  string `json:"name"`
	IsDir bool   `json:"is_dir"`
}

func (l *List) Run(ctx context.Context, input json.RawMessage) (string, error) {
	var in listInput
	if len(input) > 0 {
		if err := json.Unmarshal(input, &in); err != nil {
			return "", fmt.Errorf("invalid input: %w", err)
		}
	}
	path := in.Path
	if path == "" {
		path = "."
	}
	resolved, err := l.scope.resolve(path)
	if err != nil {
		return "", err
	}

	entries, err := os.ReadDir(resolved)
	if err != nil {
		return "", err
	}
	out := make([]listEntry, 0, len(entries))
	for _, entry := range entries {
		out = append(out, listEntry{Name: entry.Name(), IsDir: entry.IsDir()})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].IsDir != out[j].IsDir {
			return out[i].IsDir
		}
		return out[i].Name < out[j].Name
	})

	data, err := json.Marshal(out)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
