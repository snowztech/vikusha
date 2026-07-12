package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/snowztech/vikusha"
	"github.com/snowztech/vikusha/core/agent"
)

var version = "dev"

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	_ = godotenv.Load()

	if len(args) == 0 {
		usage(stderr)
		return errors.New("missing command")
	}

	switch args[0] {
	case "version":
		fmt.Fprintln(stdout, version)
		return nil
	case "chat", "run", "start":
		fs := flag.NewFlagSet(args[0], flag.ContinueOnError)
		fs.SetOutput(stderr)
		userID := fs.String("user", os.Getenv("USER"), "user id for the conversation")
		timeout := fs.Duration("timeout", 2*time.Minute, "timeout per turn")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if fs.NArg() != 1 {
			return fmt.Errorf("usage: vikusha %s <character.yaml|agent>", args[0])
		}
		a, err := buildAgent(fs.Arg(0))
		if err != nil {
			return err
		}
		return repl(stdin, stdout, a, *userID, *timeout)
	default:
		usage(stderr)
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func usage(w io.Writer) {
	fmt.Fprintln(w, "usage:")
	fmt.Fprintln(w, "  vikusha chat <character.yaml|agent>")
	fmt.Fprintln(w, "  vikusha run <character.yaml|agent>")
	fmt.Fprintln(w, "  vikusha start <character.yaml|agent>")
	fmt.Fprintln(w, "  vikusha version")
}

func buildAgent(input string) (*agent.Agent, error) {
	path, err := resolveCharacterPath(input)
	if err != nil {
		return nil, err
	}
	return vikusha.LoadAgent(path, vikusha.Options{})
}

func resolveCharacterPath(input string) (string, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return "", fmt.Errorf("character path or agent name is required")
	}
	if fileExists(trimmed) || looksLikePath(trimmed) {
		return trimmed, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("find home directory: %w", err)
	}
	path := namedAgentCharacterPath(home, trimmed)
	if !fileExists(path) {
		return "", fmt.Errorf("agent %q not found at %s", trimmed, path)
	}
	return path, nil
}

func namedAgentCharacterPath(home, name string) string {
	return filepath.Join(home, ".vikusha", "agents", name, "character.yaml")
}

func looksLikePath(input string) bool {
	return filepath.IsAbs(input) ||
		strings.ContainsRune(input, os.PathSeparator) ||
		filepath.Ext(input) == ".yaml" ||
		filepath.Ext(input) == ".yml"
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func repl(stdin io.Reader, stdout io.Writer, a *agent.Agent, userID string, timeout time.Duration) error {
	if userID == "" {
		userID = "local"
	}
	scanner := bufio.NewScanner(stdin)
	fmt.Fprintf(stdout, "%s ready. Type /exit to quit.\n", a.Name())
	for {
		fmt.Fprint(stdout, "> ")
		if !scanner.Scan() {
			return scanner.Err()
		}
		msg := strings.TrimSpace(scanner.Text())
		if msg == "" {
			continue
		}
		if msg == "/exit" || msg == "/quit" {
			return nil
		}

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		reply, err := a.Chat(ctx, userID, msg)
		cancel()
		if err != nil {
			fmt.Fprintf(stdout, "error: %v\n", err)
			continue
		}
		fmt.Fprintln(stdout, reply)
	}
}
