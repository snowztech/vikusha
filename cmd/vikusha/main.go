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
	case "create":
		fs := flag.NewFlagSet(args[0], flag.ContinueOnError)
		fs.SetOutput(stderr)
		model := fs.String("model", "gpt-4o-mini", "model id for the agent")
		provider := fs.String("provider", "openai", "provider name")
		apiKeyEnv := fs.String("api-key-env", "", "environment variable containing the provider API key")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if fs.NArg() != 1 {
			return fmt.Errorf("usage: vikusha create <agent>")
		}
		path, err := createAgent(fs.Arg(0), createOptions{
			Model:     *model,
			Provider:  *provider,
			APIKeyEnv: *apiKeyEnv,
		})
		if err != nil {
			return err
		}
		fmt.Fprintf(stdout, "created %s\n", path)
		return nil
	case "chat", "run", "start":
		fs := flag.NewFlagSet(args[0], flag.ContinueOnError)
		fs.SetOutput(stderr)
		userID := fs.String("user", os.Getenv("USER"), "user id for the conversation")
		timeout := fs.Duration("timeout", 2*time.Minute, "timeout per turn")
		logJSON := fs.Bool("log-json", false, "write structured turn logs to stderr")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if fs.NArg() != 1 {
			return fmt.Errorf("usage: vikusha %s <character.yaml|agent>", args[0])
		}
		input := fs.Arg(0)
		path, err := resolveCharacterPath(input)
		if err != nil {
			return err
		}
		var logger agent.TurnLogger
		var logFile *os.File
		if *logJSON {
			logWriter := stderr
			if namedAgentInput(input) {
				logFile, err = openTurnLog(path)
				if err != nil {
					return err
				}
				defer logFile.Close()
				logWriter = logFile
			}
			logger = agent.NewJSONLogger(logWriter)
		}
		a, err := buildAgent(path, logger)
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
	fmt.Fprintln(w, "  vikusha create <agent>")
	fmt.Fprintln(w, "  vikusha run <character.yaml|agent>")
	fmt.Fprintln(w, "  vikusha start <character.yaml|agent>")
	fmt.Fprintln(w, "  vikusha version")
}

type createOptions struct {
	Model     string
	Provider  string
	APIKeyEnv string
}

func createAgent(name string, opts createOptions) (string, error) {
	agentName := strings.TrimSpace(name)
	if !validAgentName(agentName) {
		return "", fmt.Errorf("agent name must be a simple name")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("find home directory: %w", err)
	}

	characterPath := namedAgentCharacterPath(home, agentName)
	if fileExists(characterPath) {
		return "", fmt.Errorf("agent %q already exists at %s", agentName, characterPath)
	}
	agentDir := filepath.Dir(characterPath)
	for _, dir := range []string{"memory", "workspace", "logs"} {
		if err := os.MkdirAll(filepath.Join(agentDir, dir), 0o700); err != nil {
			return "", err
		}
	}
	memoryPath := filepath.Join(agentDir, "memory")
	if err := os.WriteFile(characterPath, []byte(characterYAML(agentName, opts, memoryPath)), 0o600); err != nil {
		return "", err
	}
	return characterPath, nil
}

func validAgentName(name string) bool {
	if name == "" {
		return false
	}
	for _, r := range name {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_' {
			continue
		}
		return false
	}
	return true
}

func characterYAML(name string, opts createOptions, memoryPath string) string {
	model := strings.TrimSpace(opts.Model)
	if model == "" {
		model = "gpt-4o-mini"
	}
	provider := strings.TrimSpace(opts.Provider)
	if provider == "" {
		provider = "openai"
	}
	apiKeyEnv := strings.TrimSpace(opts.APIKeyEnv)
	if apiKeyEnv == "" {
		apiKeyEnv = defaultAPIKeyEnv(provider)
	}

	return fmt.Sprintf(`name: %s
model: %s
system_prompt: You are %s, a concise and helpful assistant.
provider:
  name: %s
  api_key_env: %s
memory:
  backend: file
  path: %s
tools:
  - file_list
  - file_read
`, name, model, name, provider, apiKeyEnv, filepath.ToSlash(memoryPath))
}

func defaultAPIKeyEnv(provider string) string {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "anthropic":
		return "ANTHROPIC_API_KEY"
	case "openrouter":
		return "OPENROUTER_API_KEY"
	case "groq":
		return "GROQ_API_KEY"
	default:
		return "OPENAI_API_KEY"
	}
}

func buildAgent(path string, logger agent.TurnLogger) (*agent.Agent, error) {
	return vikusha.LoadAgent(path, vikusha.BuildOptions{
		Workspace: workspaceForCharacter(path),
		Logger:    logger,
	})
}

func namedAgentInput(input string) bool {
	trimmed := strings.TrimSpace(input)
	return trimmed != "" && !fileExists(trimmed) && !looksLikePath(trimmed)
}

func openTurnLog(characterPath string) (*os.File, error) {
	logDir := filepath.Join(filepath.Dir(characterPath), "logs")
	if err := os.MkdirAll(logDir, 0o700); err != nil {
		return nil, err
	}
	return os.OpenFile(filepath.Join(logDir, "turns.jsonl"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
}

func workspaceForCharacter(path string) string {
	workspace := filepath.Join(filepath.Dir(path), "workspace")
	if fileExists(workspace) {
		return workspace
	}
	return ""
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
