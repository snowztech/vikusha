# Architecture

This document describes Vikusha's target architecture. When a concept is not implemented yet, the text calls that out explicitly.

## Design principles

1. **Small core, clear interfaces.** The core package fits in one head. Extensions plug into named interfaces.
2. **Go-native core.** Vikusha's core is a single Go binary. Extensions are Go packages imported at build time.
3. **No abstraction layers over provider SDKs.** Call OpenAI, Anthropic, etc. directly. Expose the raw request shape when needed.
4. **Flat conceptual model.** Tools, transports, memory backends, and LLM providers are distinct, named concepts.
5. **Character as data.** A YAML file is the entire assistant definition. Same binary, different YAML = different assistant.
6. **Observability over magic.** The user should be able to see which tool ran, which model was called, and which memory was injected.
7. **Two modes of use.** Go framework for developers who embed Vikusha, YAML runtime for configured assistants.

## Target System Flow

Vikusha has two entry paths that meet at the same agent loop:

```mermaid
flowchart LR
    yaml[character.yaml] --> loader[character loader]
    cli[vikusha CLI] --> loader
    app[Go app] --> manual[agent.New options]
    loader --> build[provider + tools + memory]
    manual --> build
    build --> agent[Agent]
    transport[CLI / future transports] --> chat[agent.Chat(ctx, userID, msg)]
    agent --> chat
    chat --> memory[Memory]
    chat --> provider[LLM provider]
    provider --> tools{tool calls?}
    tools -- yes --> registry[Tool registry]
    registry --> provider
    tools -- no --> reply[reply]
```

The framework path is explicit Go construction: the caller passes providers, tools, and memory. The runtime path starts from YAML and lets Vikusha wire those pieces. Both paths produce the same `Agent` runtime.

## Core concepts

### Character

YAML file defining an assistant's identity and runtime config. Loaded once at startup. See [CHARACTER.md](CHARACTER.md).

### Agent

The runtime core. Holds the assistant name, model, system prompt, LLM provider, tool registry, and optional memory backend. Exposes one method: `Chat(ctx, userID, msg) -> (string, error)`.

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
    Start(ctx context.Context, chat ChatFunc) error
    Stop() error
}
```

Target transports: CLI REPL, Discord, Telegram, Slack, and HTTP. Current implementation has the CLI REPL path.

The target runtime can run multiple transports for the same agent. Each incoming message is routed to `agent.Chat()`.

### LLM Provider

An interface that calls the model. One method:

```go
type Provider interface {
    Complete(ctx context.Context, req *Request) (*Response, error)
}
```

Target providers: OpenAI-compatible, Anthropic, Ollama, and eventually Gemini. Current implementation has OpenAI-compatible providers and Anthropic.

### Memory

Pluggable store for the assistant's long-term state. Interface:

```go
type Memory interface {
    Load(ctx context.Context, userID string) ([]Entry, error)
    Save(ctx context.Context, userID string, entry Entry) error
    Search(ctx context.Context, userID, query string, k int) ([]Entry, error)
}
```

Target backends: file JSONL, SQLite, and pgvector. Current implementation has file JSONL.

### RAG

Optional retrieval pipeline layered on top of `Memory`. Chunks documents, embeds via a provider, retrieves top-K at query time, injects into the prompt.

## Built-in tools

Target built-in tools:

| Tool         | What it does                          |
| ------------ | ------------------------------------- |
| `file_read`  | Read a file                           |
| `bash`       | Run approved shell commands           |
| `file_edit`  | Edit a file                           |
| `file_list`  | List directory contents               |
| `web_search` | Search the web                        |
| `web_fetch`  | Fetch a URL and extract text          |

Users enable tools per assistant in their character.yaml.

Current implementation has `file_read`.

## Directory layout

```
vikusha/
├── core/                   # vikusha-core (Go library)
│   ├── agent/              # agent loop, chat, tool handling
│   ├── character/          # YAML loader + validation
│   ├── tool/               # Tool interface + registry
│   ├── transport/          # transport interface
│   ├── llm/                # anthropic, openai-compatible
│   ├── memory/             # memory interface + file backend
│   └── tools/              # bundled tool implementations
├── cmd/vikusha/            # CLI binary
├── examples/               # runnable examples
├── docs/
└── go.mod                  # module: github.com/snowztech/vikusha
```

## Data directory

Vikusha stores per-assistant data in `~/.vikusha/`:

```
~/.vikusha/
├── config.yaml             # global config (API keys, tokens)
└── agents/
    └── <name>/
        ├── character.yaml  # agent definition
        ├── memory/         # conversation history
        └── workspace/      # files this agent can access
```

Each assistant has its own workspace for file tool isolation.

This named-agent layout is the target runtime layout. Current implementation can load a character YAML directly from a path.

## The chat loop

Simplified flow for a single turn:

1. Transport receives user message and calls `agent.Chat(ctx, userID, msg)`.
2. Agent builds the prompt from the system prompt, user message, tool definitions, and memory when enabled.
3. Agent calls `provider.Complete()`.
4. If the response contains tool calls, agent runs them sequentially, appends results, and loops to step 3.
5. When the response has no tool calls, agent returns the final text to the transport.
6. Transport delivers to the user.

No parallel tool calls in v1. Sequential works. Revisit if a real use case needs it.

## Context management

Vikusha should enforce the same discipline:

- Prompt caching on system and tool definitions (Anthropic, OpenAI `prompt_cache_key`).
- Budgeted history. Default 30k token cap. Trim oldest with a summary preamble.
- Tool result cap (default 4k chars per result) to prevent a rogue bash output from blowing the window.
- Structured logs per turn: tokens in/out, cache hit rate, cost, duration.

## Multi-agent

The target runtime can run multiple named assistants from one binary:

```bash
vikusha start writer
vikusha start coach
```

Each assistant has its own character, memory, transports, and tool registry.

## Observability

Every turn should emit a structured log line:

```
event=turn agent=vikusha user=lucas tokens_in=1203 tokens_out=412 cache_read=980 tools=bash,file cost=0.0021 duration=2.3s
```

## What vikusha deliberately does NOT do

- **No vector DB of its own.** Use SQLite-VSS, pgvector, or Chroma via the `Memory` interface.
- **No agent-builder UI.** YAML + Go is the interface. Building a web UI is a separate project.
- **No LangChain-style chain abstraction.** The chat loop IS the abstraction.
- **No sub-agents, no delegation, no task graphs.** A single agent calls tools. That's it. If you need multi-step planning, put it in a tool.
- **No MCP.** Too much token overhead for the personal-scale use cases vikusha targets. Revisit if that changes.
- **No evaluators.** Post-response scoring is out of scope for v1.
