# Roadmap

## North star

Vikusha is a Go framework and runtime for always-on AI assistants.

The framework path is for Go developers: import Vikusha, wire providers and tools, and call `agent.Chat(ctx, userID, msg)`.

The runtime path is for configured assistants: define a named agent in YAML, give it memory, tools, workspace, and transports, then start it as a long-running process.

The target experience is:

```bash
vikusha create writer
vikusha start writer
vikusha chat writer
```

In that model, `writer` is backed by a character YAML under `~/.vikusha/agents/writer/`, with its own memory, workspace, logs, secrets, and transports.

Priorities for the framework, grouped by what ships first.

- **Now**: what blocks a usable v0.1.
- **Next**: what turns the harness into a real framework.
- **Later**: ideas to explore once the foundation is solid.

## Now (v0.1): the core harness

A single agent you can talk to from the terminal, backed by a core harness that is fast, predictable, and easy to build on. Everything in Next depends on this being right.

### Agent loop

- [x] `agent.Chat(ctx, userID, msg)` as the single entry point.
- [x] Tool-call loop with a hard iteration cap.
- [x] Per-user serialization so one user's turn cannot interleave with itself.
- [ ] Cancellation via context, surfaced through a per-user cancel handle.
- [x] Panic recovery around tool execution.

### Providers

- [x] Anthropic (raw HTTP, no SDK).
- [x] OpenAI-compatible backend (covers OpenAI, Groq, OpenRouter, LM Studio).
- [ ] Streaming responses for terminal and future TUI use.
- [ ] Retry with exponential backoff on 429 and 5xx, respecting `retry-after`.
- [ ] Usage accounting split across input, output, cache read, cache write.

### Context engineering

- [ ] Prompt caching on system prompt and tool definitions.
- [ ] Token-budgeted history (default 30k tokens), not a message count.
- [ ] Tool-result cap (default 4k) before results enter history.
- [ ] Summarize-on-trim for evicted messages.
- [ ] Rolling summary compaction every N turns to keep the cache prefix stable.

### Memory

- [x] File backend (jsonl, per agent).
- [x] Typed entries: preference, fact, note.
- [x] Interface: `Load`, `Save`, `Search`. Substring search is fine for v0.1.
- [x] Memory entries injected into the agent prompt.
- [ ] Automatic detection of user corrections and preferences from chat.

### Per-agent workspace

- [x] Each agent owns `~/.vikusha/agents/<name>/` with its own character, memory directory, and workspace directory.
- [ ] Per-agent logs under `~/.vikusha/agents/<name>/logs`.
- [x] File tools default-scoped to the agent's workspace.
- [ ] Paths outside the workspace require explicit approval, persisted per agent.
- [x] Path resolution blocks `..` escapes and symlinks pointing outside the workspace.

### Tools

- [x] Tool interface with a stable JSON schema so definitions cache cleanly.
- [ ] Built-in: `bash`, `file_read`, `file_edit`, `file_list`, `web_search`, `web_fetch`.
- [ ] Per-tool timeout and result-cap overrides via character YAML.
- [ ] Danger detection on bash and file writes, with an approval flow.

### Character

- [x] YAML loader with strict validation.
- [x] Required fields: `name`, `model`, `system_prompt`.
- [ ] Optional fields: `tools`, `tool_config`, `memory`, `context`, `logging`.
- [x] Validation reports every problem at once, not just the first.

### Transports

- [x] CLI REPL (`vikusha chat <char.yaml>`).
- [x] Named-agent CLI REPL (`vikusha chat <agent>`).

### Observability

- [ ] Structured JSON log line per turn: tokens, cache hits, tools used, cost, duration, loop iterations.
- [ ] Colored terminal logger for interactive sessions.
- [ ] Cost estimation per provider and model.

### CLI

- [x] `vikusha run <char.yaml>`: start an agent.
- [x] `vikusha start <char.yaml>`: preferred alias for starting an agent.
- [x] `vikusha start <agent>`: start a named agent in the terminal.
- [x] `vikusha chat <char.yaml>`: interactive terminal session.
- [x] `vikusha version`.

## Next (v0.2 to v0.7)

### v0.2: chat transports

- [ ] Discord transport (bot, DMs, slash commands).
- [ ] Slack transport (bot, DMs, slash commands).
- [ ] Telegram transport (bot, commands).
- [ ] Approval flow shared across transports.
- [ ] Per-user conversation isolation inside a single agent.

### v0.3: scaffolding new agents

- [x] `vikusha create <name>` scaffolds a new agent from the default template.
- [x] Created agents are stored under `~/.vikusha/agents/<name>/`.
- [ ] Built-in templates: `personal`, `support`, `dev`.
- [ ] Generated output: `main.go`, `character.yaml`, `.env.example`, `Makefile`.
- [ ] `vikusha build <dir>` wraps `go build` so non-Go users get one command.

### v0.4: config and setup

- [ ] `vikusha setup` wizard for API keys, transport tokens, default provider.
- [ ] Encrypted global config at `~/.vikusha/config`.
- [ ] Per-agent secret store, separate from the global config.
- [ ] Connection test to verify keys and tokens before first run.

### v0.5: RAG and richer memory

- [ ] SQLite memory backend.
- [ ] pgvector memory backend for larger deployments.
- [ ] RAG pipeline: chunking, embedding, retrieval, injection.
- [ ] `vikusha ingest <agent> <path>` loads documents into an agent's memory.
- [ ] Retrieval config in YAML: top-k, min score, sources.

### v0.6: terminal polish

- [ ] Bubbletea TUI for `vikusha chat`.
- [ ] Streaming token rendering.
- [ ] Markdown rendering with syntax highlighting.
- [ ] Slash commands: `/new`, `/model`, `/tools`, `/status`.
- [ ] Ctrl+C cancels the current turn, not the session.

### v0.7: deploy and operate

- [ ] Install script (`curl ... | sh`).
- [ ] systemd service template for Linux VPS deploys.
- [ ] `vikusha logs <agent>`, `vikusha status <agent>`, `vikusha stop <agent>`.
- [ ] GitHub releases with prebuilt binaries for common platforms.

## Later

- **Extensions repo.** A separate `vikusha-extensions` repo for integrations like Notion, Google Calendar, GitHub issues, Linear.
- **Plugin loading beyond Go modules.** Subprocess tools over JSON stdio and WASM plugins are both worth exploring once the core is stable.
- **Voice input.** Local Whisper transcription for voice messages on Discord or Telegram.
- **Web dashboard.** Read-only view of agents, memory, and turn logs.
- **Multi-agent coordination.** Handoff to another agent as a tool call, pursued only when a concrete use case shows up.
- **Evaluators.** Regression tests against a fixed prompt set.
- **MCP support.** Reconsidered once there is demand from real users.

## v1.0 criteria

- Core harness stable and documented.
- Discord and Slack transports running in production somewhere.
- At least one non-default agent built with Vikusha and deployed.
- Test coverage on the agent loop, context trimming, and tool registry.
- Install script and prebuilt binaries published.
- Clear upgrade path from v0.x.
