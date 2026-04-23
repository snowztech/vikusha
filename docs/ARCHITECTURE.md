# Architecture

## Design principles

1. **Small core, clear interfaces.** The core package fits in one head. Extensions plug into named interfaces.
2. **Go-native core.** Vika's core is a single Go binary. Extensions are Go packages imported at build time.
3. **No abstraction layers over provider SDKs.** Call OpenAI, Anthropic, etc. directly. Expose the raw request shape when needed.
4. **Flat conceptual model.** Tools, transports, memory backends, and LLM providers are distinct, named concepts.
5. **Character as data.** A YAML file is the entire assistant definition. Same binary, different YAML = different assistant.
6. **Observability over magic.** The user always sees which tool ran, which model was called, which memory was injected.
7. **Two modes of use.** CLI for 90% of users (`vika run character.yaml`), Go library for the 10% who embed Vika in custom binaries or SaaS backends.

## Core concepts

### Character

YAML file defining an assistant's identity. Loaded once at startup. See [CHARACTER.md](CHARACTER.md).

### Agent

The runtime. Holds the character, the LLM provider, the tool registry, the memory backend, and the active transports. Exposes one method: `Chat(ctx, msg) -> Response`.

### Tool

A function the LLM can call. One interface:

```go
type Tool interface {
    Name() string
    Description() string
    Schema() json.RawMessage // JSON schema for params
    Run(ctx context.Context, input json.RawMessage) (string, error)
}
```

Tools are registered into the agent at construction. A `ToolRegistry` holds them.

### Transport

A channel through which users talk to the agent. Interface:

```go
type Transport interface {
    Name() string
    Start(ctx context.Context, agent *Agent) error
    Stop() error
}
```

Built-in transports: Discord, Telegram, Slack, CLI (REPL), HTTP (REST endpoint).

A single agent can run multiple transports simultaneously. Each incoming message is routed to `agent.Chat()`.

### LLM Provider

An interface that calls the model. One method:

```go
type Provider interface {
    Complete(ctx context.Context, req *Request) (*Response, error)
}
```

Built-in providers: OpenAI-compatible (covers OpenAI, Groq, OpenRouter, OpenCode Zen, LM Studio), Anthropic, Ollama. Google Gemini later.

### Memory

Pluggable store for the assistant's long-term state. Interface:

```go
type Memory interface {
    Load(ctx context.Context, userID string) ([]Entry, error)
    Save(ctx context.Context, userID string, entry Entry) error
    Search(ctx context.Context, userID, query string, k int) ([]Entry, error)
}
```

Built-in backends: file (markdown), SQLite, pgvector (for RAG).

### RAG

Optional retrieval pipeline layered on top of `Memory`. Chunks documents, embeds via a provider, retrieves top-K at query time, injects into the prompt.

## Built-in tools

Vika ships with these tools bundled in the binary:

| Tool | What it does |
|------|--------------|
| `bash` | Run any bash command |
| `file_read` | Read a file |
| `file_edit` | Edit a file (replace text) |
| `file_list` | List directory contents |
| `web_search` | Search the web |
| `web_fetch` | Fetch a URL and extract text |
| `github` | GitHub API tools (issues, PRs, repos) |
| `http` | Generic HTTP requests |

Users enable tools per assistant in their character.yaml.

## Directory layout

```
vika/
├── core/                   # vika-core (Go library)
│   ├── agent/              # agent loop, chat, tool handling
│   ├── character/          # YAML loader + validation
│   ├── tool/               # Tool interface + registry
│   ├── transport/          # discord, telegram, slack, cli, http
│   ├── llm/                # anthropic, openai, ollama
│   ├── memory/             # file, sqlite, pgvector
│   └── rag/                # chunking, retrieval
├── tools/                  # bundled tool implementations
├── cmd/vika/               # CLI binary
│   ├── setup.go            # wizard for initial config
│   ├── run.go              # run agents
│   ├── create.go           # scaffold new agent
│   └── chat.go             # terminal UI
├── characters/
│   └── vika.yaml           # default Vika assistant
├── docs/
└── go.mod                  # module: github.com/snowztech/vika
```

## Data directory

Vika stores per-assistant data in `~/.vika/`:

```
~/.vika/
├── config.yaml             # global config (API keys, tokens)
└── agents/
    └── <name>/
        ├── character.yaml  # agent definition
        ├── memory/         # conversation history
        └── workspace/      # files this agent can access
```

Each assistant has its own workspace for file tool isolation.

## The chat loop

Simplified flow for a single turn:

1. Transport receives user message → calls `agent.Chat(ctx, userID, msg)`.
2. Agent loads relevant memory for `userID` (+ RAG retrieval if enabled).
3. Agent builds the prompt: system prompt from character, memory preamble, tool index, conversation history, user message.
4. Agent calls `provider.Complete()`.
5. If the response contains tool calls, agent runs them sequentially, appends results, loops to step 4.
6. When the response has no tool calls, agent returns the final text to the transport.
7. Transport delivers to the user.

No parallel tool calls in v1. Sequential works. Revisit if a real use case needs it.

## Context management

Vika enforces the same discipline:

- Prompt caching on system and tool definitions (Anthropic, OpenAI `prompt_cache_key`).
- Budgeted history. Default 30k token cap. Trim oldest with a summary preamble.
- Tool result cap (default 4k chars per result) to prevent a rogue bash output from blowing the window.
- Structured logs per turn: tokens in/out, cache hit rate, cost, duration.

## Multi-agent

A single vika binary can run multiple assistants from day one:

```bash
vika run vika coach koda
```

Each assistant has its own character, memory, transports, and tool registry.

## Observability

Every turn emits a structured log line:

```
event=turn agent=vika user=lucas tokens_in=1203 tokens_out=412 cache_read=980 tools=bash,file_read cost=0.0021 duration=2.3s
```

Optional Prometheus endpoint (`/metrics`) for long-running deployments.

## What vika deliberately does NOT do

- **No vector DB of its own.** Use SQLite-VSS, pgvector, or Chroma via the `Memory` interface.
- **No agent-builder UI.** YAML + Go is the interface. Building a web UI is a separate project.
- **No LangChain-style chain abstraction.** The chat loop IS the abstraction.
- **No sub-agents, no delegation, no task graphs.** A single agent calls tools. That's it. If you need multi-step planning, put it in a tool.
- **No MCP.** Too much token overhead for the personal-scale use cases vika targets. Revisit if that changes.
- **No evaluators.** Post-response scoring is out of scope for v1.