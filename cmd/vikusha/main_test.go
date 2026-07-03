package main

import (
	"bytes"
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
