package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/snowztech/vikusha"
	"github.com/snowztech/vikusha/core/agent"
)

const version = "dev"

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
	case "chat", "run":
		fs := flag.NewFlagSet(args[0], flag.ContinueOnError)
		fs.SetOutput(stderr)
		userID := fs.String("user", os.Getenv("USER"), "user id for the conversation")
		timeout := fs.Duration("timeout", 2*time.Minute, "timeout per turn")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if fs.NArg() != 1 {
			return fmt.Errorf("usage: vikusha %s <character.yaml>", args[0])
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
	fmt.Fprintln(w, "  vikusha chat <character.yaml>")
	fmt.Fprintln(w, "  vikusha run <character.yaml>")
	fmt.Fprintln(w, "  vikusha version")
}

func buildAgent(path string) (*agent.Agent, error) {
	return vikusha.LoadAgent(path, vikusha.Options{})
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
