package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVersion(t *testing.T) {
	oldVersion := version
	version = "test-version"
	t.Cleanup(func() { version = oldVersion })

	var out, errOut bytes.Buffer
	if err := run([]string{"version"}, strings.NewReader(""), &out, &errOut); err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(out.String()); got != "test-version" {
		t.Fatalf("version output = %q, want test-version", got)
	}
}

func TestMissingCommand(t *testing.T) {
	var out, errOut bytes.Buffer
	err := run(nil, strings.NewReader(""), &out, &errOut)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(errOut.String(), "vikusha chat") {
		t.Fatalf("expected usage in stderr, got %q", errOut.String())
	}
}

func TestCreateAgent(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	var out, errOut bytes.Buffer
	err := run([]string{"create", "writer"}, strings.NewReader(""), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}

	characterPath := namedAgentCharacterPath(home, "writer")
	if !strings.Contains(out.String(), characterPath) {
		t.Fatalf("expected created path in output, got %q", out.String())
	}
	data, err := os.ReadFile(characterPath)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	for _, want := range []string{
		"name: writer",
		"model: gpt-4o-mini",
		"api_key_env: OPENAI_API_KEY",
		"tools:\n  - file_list\n  - file_read",
		filepath.ToSlash(filepath.Join(home, ".vikusha", "agents", "writer", "memory")),
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("expected %q in character.yaml:\n%s", want, content)
		}
	}
	for _, dir := range []string{"memory", "workspace", "logs"} {
		if _, err := os.Stat(filepath.Join(home, ".vikusha", "agents", "writer", dir)); err != nil {
			t.Fatalf("expected %s directory: %v", dir, err)
		}
	}
}

func TestCreateAgentWithProviderOptions(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	var out, errOut bytes.Buffer
	err := run([]string{"create", "-model", "claude-sonnet-4-6", "-provider", "anthropic", "coach"}, strings.NewReader(""), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(namedAgentCharacterPath(home, "coach"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	for _, want := range []string{
		"model: claude-sonnet-4-6",
		"name: anthropic",
		"api_key_env: ANTHROPIC_API_KEY",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("expected %q in character.yaml:\n%s", want, content)
		}
	}
}

func TestCreateAgentRefusesExistingAgent(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	if _, err := createAgent("writer", createOptions{}); err != nil {
		t.Fatal(err)
	}
	if _, err := createAgent("writer", createOptions{}); err == nil {
		t.Fatal("expected existing agent error")
	}
}

func TestCreateAgentRejectsInvalidName(t *testing.T) {
	for _, name := range []string{"", "bad/name", "bad name", "bad.yaml"} {
		if _, err := createAgent(name, createOptions{}); err == nil {
			t.Fatalf("expected invalid name error for %q", name)
		}
	}
}

func TestResolveCharacterPathKeepsExplicitPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "character.yaml")

	got, err := resolveCharacterPath(path)
	if err != nil {
		t.Fatal(err)
	}
	if got != path {
		t.Fatalf("resolveCharacterPath() = %q, want %q", got, path)
	}
}

func TestResolveCharacterPathFindsNamedAgent(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	path := namedAgentCharacterPath(home, "writer")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("name: Writer\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := resolveCharacterPath("writer")
	if err != nil {
		t.Fatal(err)
	}
	if got != path {
		t.Fatalf("resolveCharacterPath() = %q, want %q", got, path)
	}
}

func TestResolveCharacterPathReportsMissingNamedAgent(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	_, err := resolveCharacterPath("writer")
	if err == nil {
		t.Fatal("expected missing named agent error")
	}
	if !strings.Contains(err.Error(), namedAgentCharacterPath(home, "writer")) {
		t.Fatalf("expected agent path in error, got %q", err.Error())
	}
}

func TestWorkspaceForCharacterUsesSiblingWorkspace(t *testing.T) {
	dir := t.TempDir()
	workspace := filepath.Join(dir, "workspace")
	if err := os.Mkdir(workspace, 0o700); err != nil {
		t.Fatal(err)
	}

	got := workspaceForCharacter(filepath.Join(dir, "character.yaml"))
	if got != workspace {
		t.Fatalf("workspaceForCharacter() = %q, want %q", got, workspace)
	}
}

func TestWorkspaceForCharacterReturnsEmptyWhenMissing(t *testing.T) {
	got := workspaceForCharacter(filepath.Join(t.TempDir(), "character.yaml"))
	if got != "" {
		t.Fatalf("workspaceForCharacter() = %q, want empty", got)
	}
}

func TestNamedAgentInput(t *testing.T) {
	if !namedAgentInput("writer") {
		t.Fatal("expected bare agent name to be named input")
	}
	for _, input := range []string{"./writer.yaml", "/tmp/writer.yaml", "writer.yaml"} {
		if namedAgentInput(input) {
			t.Fatalf("expected %q to be path input", input)
		}
	}
}

func TestOpenTurnLogCreatesLogFile(t *testing.T) {
	characterPath := filepath.Join(t.TempDir(), "writer", "character.yaml")
	f, err := openTurnLog(characterPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString("{}\n"); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	logPath := filepath.Join(filepath.Dir(characterPath), "logs", "turns.jsonl")
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "{}\n" {
		t.Fatalf("log file = %q, want json line", data)
	}
}
