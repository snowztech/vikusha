# Roadmap

Vika's journey from prototype to production-ready framework.

## Phase 1: Core Engine (v0.1)

**Goal:** Working agent harness with essential features.

- [ ] Agent loop with tool execution
- [ ] File-based memory (markdown)
- [ ] Basic tools: bash, file_read, file_edit, file_list
- [ ] CLI transport (REPL)
- [ ] Anthropic and OpenAI providers
- [ ] Character YAML loading and validation
- [ ] Token budgeting (30k history, 4k tool results)
- [ ] Structured logging per turn

**Status:** Derived from existing nevinho code.

---

## Phase 2: Transports (v0.2)

**Goal:** Connect agents to popular chat platforms.

- [ ] Discord transport (bot, DMs, slash commands)
- [ ] Telegram transport (bot, commands)
- [ ] Transport interface (clean abstraction)
- [ ] Per-user history isolation
- [ ] Approval workflow for dangerous operations

---

## Phase 3: Configuration (v0.3)

**Goal:** Easy setup for non-technical users.

- [ ] `vika setup` interactive wizard
- [ ] Global config (~/.vika/config.yaml)
- [ ] Provider configuration (API keys, defaults)
- [ ] Transport token management
- [ ] Connection testing (verify keys work)

---

## Phase 4: Terminal UI (v0.4)

**Goal:** Claude Code-like experience in terminal.

- [ ] `vika chat` command
- [ ] Interactive TUI (bubbletea or tview)
- [ ] Markdown rendering
- [ ] Command history
- [ ] Syntax highlighting for code blocks

---

## Phase 5: Isolation (v0.5)

**Goal:** Safe multi-agent operation.

- [ ] Per-agent workspace (~/.vika/agents/<name>/workspace/)
- [ ] Workspace-scoped file tools
- [ ] Memory per agent
- [ ] Agent creation wizard (`vika create`)

---

## Phase 6: Expansion (v0.6-v0.9)

**Goal:** More providers, tools, and integrations.

- [ ] Slack transport
- [ ] HTTP transport (REST endpoint)
- [ ] Ollama provider (local models)
- [ ] Web tools: web_search, web_fetch
- [ ] GitHub tool
- [ ] SQLite memory backend
- [ ] pgvector memory backend (RAG)

---

## Phase 7: Release (v1.0)

**Goal:** Production-ready, documented, shipped.

- [ ] Comprehensive docs (all vika-docs files)
- [ ] Install script (curl ... | sh)
- [ ] Go module published
- [ ] GitHub releases with binaries
- [ ] First community feedback cycle
- [ ] Performance tuning

---

## Future ideas (post-v1.0)

These are not planned yet but mentioned for direction:

- **Voice** — Whisper transcription (like [nevinho](https://github.com/lucasnevespereira/nevinho))
- **Extensions** — External Go modules (vika-ext-notion, etc.)
- **Multi-agent** — Agents talking to each other
- **Web UI** — Simple dashboard
- **Plugins** — MCP support (if demand materializes)

---

## Release cadence

- Alpha/beta during v0.1-v0.5 (internal testing)
- Release candidate for v1.0
- Semantic versioning after v1.0
- Patch releases for bugs, minor releases for features

---

## Contributing

Early phases: maintainer-driven.
Post-v1.0: contributions welcome. Check issues for starter tasks.
