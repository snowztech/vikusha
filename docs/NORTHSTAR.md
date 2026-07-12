# North Star

Vikusha is a Go framework and runtime for always-on AI assistants.

The goal is simple: define an assistant once, then run it wherever people need it.

## Two Paths

### Go framework

Go developers can import Vikusha, wire providers and tools, and call:

```go
reply, err := agent.Chat(ctx, userID, msg)
```

This path is for custom binaries, SaaS backends, internal tools, and tests.

### YAML runtime

Users can define an assistant in a character YAML and run it with the `vikusha` CLI.

Today that starts with:

```bash
vikusha chat character.yaml
```

The target experience is named agents:

```bash
vikusha create writer
vikusha start writer
vikusha chat writer
```

In that model, `writer` is backed by a character YAML and has its own tools, memory, workspace, logs, secrets, and transports.

## Always-On Assistants

Always-on means the assistant is a long-running process, not just a one-off prompt.

Examples:

- A personal writing assistant available from the terminal or Telegram.
- A coding assistant with access to a scoped workspace.
- A support assistant running in Slack.
- An ops assistant that can inspect approved systems and remember team preferences.

The core runtime should make these assistants predictable, observable, and safe to deploy as a single Go binary.
