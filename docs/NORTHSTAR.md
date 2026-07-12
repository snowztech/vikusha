# North Star

Vikusha is a Go framework and runtime for always-on AI assistants.

The goal is simple: define an assistant once, then run it wherever people need it.

## Core Idea

Vikusha turns a character into a runnable agent. The normal user-facing character format is YAML:

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

Go developers can load the same character from code, or use lower-level APIs when they need custom wiring.

## Always-On Assistants

Always-on means the assistant is a long-running process, not just a one-off prompt.

Examples:

- A personal writing assistant available from the terminal or Telegram.
- A coding assistant with access to a scoped workspace.
- A support assistant running in Slack.
- An ops assistant that can inspect approved systems and remember team preferences.

The core runtime should make these assistants predictable, observable, and safe to deploy as a single Go binary.
