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
