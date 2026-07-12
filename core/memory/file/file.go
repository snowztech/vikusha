package file

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/snowztech/vikusha/core/memory"
)

type Memory struct {
	dir string
}

var _ memory.Memory = (*Memory)(nil)

func New(dir string) *Memory {
	return &Memory{dir: dir}
}

func (m *Memory) Load(ctx context.Context, userID string) ([]memory.Entry, error) {
	path, err := m.path(userID)
	if err != nil {
		return nil, err
	}
	return readEntries(ctx, path)
}

func (m *Memory) Save(ctx context.Context, userID string, entry memory.Entry) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := validateEntry(entry); err != nil {
		return err
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now().UTC()
	}

	path, err := m.path(userID)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()

	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	if _, err := f.Write(append(data, '\n')); err != nil {
		return err
	}
	return nil
}

func (m *Memory) Search(ctx context.Context, userID, query string, k int) ([]memory.Entry, error) {
	entries, err := m.Load(ctx, userID)
	if err != nil {
		return nil, err
	}
	if k <= 0 {
		return nil, nil
	}

	query = strings.ToLower(strings.TrimSpace(query))
	out := make([]memory.Entry, 0, k)
	for i := len(entries) - 1; i >= 0 && len(out) < k; i-- {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if query == "" || strings.Contains(strings.ToLower(entries[i].Content), query) {
			out = append(out, entries[i])
		}
	}
	return out, nil
}

func (m *Memory) path(userID string) (string, error) {
	if strings.TrimSpace(m.dir) == "" {
		return "", fmt.Errorf("memory file: dir is required")
	}
	name := safeUserID(userID)
	if name == "" {
		return "", fmt.Errorf("memory file: userID is required")
	}
	return filepath.Join(m.dir, name+".jsonl"), nil
}

func readEntries(ctx context.Context, path string) ([]memory.Entry, error) {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var entries []memory.Entry
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	line := 0
	for scanner.Scan() {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		line++
		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			continue
		}
		var entry memory.Entry
		if err := json.Unmarshal([]byte(text), &entry); err != nil {
			return nil, fmt.Errorf("%s:%d: %w", path, line, err)
		}
		if err := validateEntry(entry); err != nil {
			return nil, fmt.Errorf("%s:%d: %w", path, line, err)
		}
		entries = append(entries, entry)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}

func validateEntry(entry memory.Entry) error {
	switch entry.Type {
	case memory.EntryPreference, memory.EntryFact, memory.EntryNote:
	default:
		return fmt.Errorf("memory file: unsupported entry type %q", entry.Type)
	}
	if strings.TrimSpace(entry.Content) == "" {
		return fmt.Errorf("memory file: content is required")
	}
	return nil
}

func safeUserID(userID string) string {
	userID = strings.TrimSpace(userID)
	var b strings.Builder
	for _, r := range userID {
		switch {
		case unicode.IsLetter(r), unicode.IsDigit(r):
			b.WriteRune(r)
		case r == '-', r == '_', r == '.':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	return strings.Trim(b.String(), ".")
}
