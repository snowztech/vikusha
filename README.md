<div align="center">
    <img src="assets/vikusha.png" alt="Vikusha" width="180" style="border-radius: 10px; border-width: 1px;">
    <h1>Vikusha</h1>
    <h3><em>Go framework for AI assistants. Run where you need them.</em></h3>
    <p>
    <a href="https://github.com/snowztech/vikusha/stargazers"><img src="https://img.shields.io/github/stars/snowztech/vikusha?style=flat&logo=github" alt="Stars"></a>
    <a href="https://github.com/snowztech/vikusha/network/members"><img src="https://img.shields.io/github/forks/snowztech/vikusha?style=flat&logo=github" alt="Forks"></a>
    <a href="https://github.com/snowztech/vikusha/issues"><img src="https://img.shields.io/github/issues/snowztech/vikusha?style=flat&logo=github" alt="Issues"></a>
    <a href="https://github.com/snowztech/vikusha/graphs/contributors"><img src="https://img.shields.io/github/contributors/snowztech/vikusha?style=flat&logo=github" alt="Contributors"></a>
    <a href="LICENSE"><img src="https://img.shields.io/badge/License-MIT-green?style=flat" alt="License"></a>
    </p>
</div>

An assistant has personality, tools, memory, and a transport. You define one in YAML, run it, and use it on Discord or Slack. Same binary, different YAMLs: a coding assistant, a support bot, a Discord bot. Vikusha ships with a default assistant ready to use.

## What is Vikusha?

Vikusha is the harness your assistants run on. It handles the agent loop, context engineering, prompt caching, memory, tool execution, and transport wiring. You write the character YAML, the harness does the rest.

- **Providers**: Anthropic, OpenAI-compatible, Ollama.
- **Tools**: bash, file, web search, grep, glob.
- **Memory**: file, SQLite, pgvector.
- **Transports**: terminal, Discord, Slack, Telegram.
- **Isolation**: separate workspace and secrets per assistant.
- **Scaffolder**: `vikusha create name` spawns a new assistant.
- **Observability**: structured logs, tokens, cost, duration.

---

## Quickstart

Create `character.yaml`:

```yaml
name: Helper
model: gpt-4o-mini
system_prompt: You are a concise assistant. Answer in one short paragraph.
provider:
  name: openai
  api_key_env: OPENAI_API_KEY
tools:
  - file_read
```

```bash
go install github.com/snowztech/vikusha/cmd/vikusha@latest
export OPENAI_API_KEY=...
vikusha chat character.yaml
```

## Creating Agents

Vikusha can create agents from YAML or from Go code. All paths return an `*agent.Agent`, which you use through `Chat(ctx, userID, msg)`.

### 1. YAML-first

Use this when you want the same assistant definition to work from the CLI and from an embedded Go app.

```go
a, err := vikusha.LoadAgent("character.yaml", vikusha.Options{})
reply, err := a.Chat(ctx, "lucas", "hello")
```

Currently implemented character fields:

```yaml
name: Helper
model: gpt-4o-mini
system_prompt: You are helpful.
provider:
  name: openai # openai, openrouter, anthropic
  api_key_env: OPENAI_API_KEY
  # base_url: http://localhost:1234/v1
tools:
  - file_read
```

If `provider` is omitted, Vikusha infers Anthropic for models beginning with `claude`; otherwise it uses OpenAI-compatible chat completions. The default env vars are `ANTHROPIC_API_KEY`, `OPENAI_API_KEY`, `OPENROUTER_API_KEY`, and `GROQ_API_KEY`, depending on the provider.

### 2. Struct-first

Use this when your app already has config in memory, but you still want Vikusha to wire providers and built-in tools.

```go
c := &character.Character{
	Name:         "Helper",
	Model:        "gpt-4o-mini",
	SystemPrompt: "You are helpful.",
	Tools:        []string{"file_read"},
}

a, err := vikusha.NewAgent(c, vikusha.Options{})
```

### 3. Manual wiring

Use `agent.New` directly when you want full control over provider instances, tool registries, memory, or tests. This is the lower-level API that the YAML and struct-first helpers build on.

```go
reg := tool.NewRegistry()
reg.Register(file.NewRead())

a, err := agent.New(agent.Options{
	Name:         "Helper",
	Model:        "gpt-4o-mini",
	SystemPrompt: "You are helpful.",
	Provider:     llm.NewOpenAI(os.Getenv("OPENAI_API_KEY")),
	Tools:        reg,
})
```

See:

- [examples/from_yaml](examples/from_yaml): load a character file through the high-level API.
- [examples/hello](examples/hello): create the smallest agent with `agent.New`.
- [examples/file_read](examples/file_read): create an agent with a registered tool.

---

## Documentation

- [ARCHITECTURE.md](docs/ARCHITECTURE.md): core concepts, interfaces, directory layout.
- [CHARACTER.md](docs/CHARACTER.md): full YAML spec with examples.
- [ROADMAP.md](docs/ROADMAP.md): where we're headed.

---

## License

[MIT](LICENSE). Copyright (c) 2026 snowztech.
