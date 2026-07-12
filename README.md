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

An assistant has personality, tools, memory, and a transport. You define one in YAML, run it from the terminal today, and later attach transports like Discord or Slack. Same binary, different YAMLs: a coding assistant, a support bot, an ops assistant.

## What is Vikusha?

Vikusha is the harness your assistants run on. It handles the agent loop, context engineering, prompt caching, memory, tool execution, and transport wiring. You write the character YAML, the harness does the rest.

The long-term direction is always-on assistants: define an assistant once, then run it wherever people need it.

- **Providers**: Anthropic, OpenAI-compatible, Ollama.
- **Tools**: bash, file, web search, grep, glob.
- **Memory**: file, SQLite, pgvector.
- **Transports**: terminal, Discord, Slack, Telegram.
- **Isolation**: separate workspace and secrets per assistant.
- **Scaffolder**: `vikusha create writer` creates a named assistant.
- **Observability**: structured logs, tokens, cost, duration.

---

## Quickstart

Install the CLI:

```bash
curl -sSL https://raw.githubusercontent.com/snowztech/vikusha/main/install.sh | bash
vikusha version
```

Go users can also install from source:

```bash
go install github.com/snowztech/vikusha/cmd/vikusha@latest
```

### Option 1: YAML

Create `character.yaml`.

```yaml
name: Helper
model: gpt-4o-mini
system_prompt: You are a concise assistant.
provider:
  name: openai
  api_key_env: OPENAI_API_KEY
```

```bash
export OPENAI_API_KEY=...
vikusha chat character.yaml
```

Named agents are loaded from `~/.vikusha/agents/<name>/character.yaml`.

```bash
vikusha create writer
vikusha chat writer
```

Named agents get their own `workspace/`. Built-in file tools read from that workspace and reject paths outside it.

For structured turn logs, pass `-log-json`:

```bash
vikusha chat -log-json writer
```

You can load the same YAML from Go.

```go
a, err := vikusha.LoadAgent("character.yaml", vikusha.Options{})
reply, err := a.Chat(ctx, "lucas", "hello")
```

### Option 2: Go

Use `agent.New` when you want to pass the provider and tools yourself.

```go
reg := tool.NewRegistry()
reg.Register(file.NewList())
reg.Register(file.NewRead())

a, err := agent.New(agent.Options{
	Name:         "Helper",
	Model:        "gpt-4o-mini",
	SystemPrompt: "You are helpful.",
	Provider:     llm.NewOpenAI(os.Getenv("OPENAI_API_KEY")),
	Tools:        reg,
})
```

Both paths return an `*agent.Agent`; call `Chat(ctx, userID, msg)` to run a turn.

## Character YAML

The fields implemented today are:

```yaml
name: Helper
model: gpt-4o-mini
system_prompt: You are helpful.
provider:
  name: openai # openai, openrouter, anthropic
  api_key_env: OPENAI_API_KEY
  # base_url: http://localhost:1234/v1
memory:
  backend: file
  path: .vikusha/helper/memory
tools:
  - file_list
  - file_read
```

If `provider` is omitted, Vikusha infers Anthropic for models beginning with `claude`; otherwise it uses OpenAI-compatible chat completions. The default env vars are `ANTHROPIC_API_KEY`, `OPENAI_API_KEY`, `OPENROUTER_API_KEY`, and `GROQ_API_KEY`, depending on the provider.

See:

- [examples/from_yaml](examples/from_yaml): create an agent from YAML.
- [examples/hello](examples/hello): create the smallest agent with `agent.New`.
- [examples/file_read](examples/file_read): create an agent with registered file tools.

---

## Documentation

- [NORTHSTAR.md](docs/NORTHSTAR.md): product direction and target experience.
- [ARCHITECTURE.md](docs/ARCHITECTURE.md): core concepts, interfaces, directory layout.
- [CHARACTER.md](docs/CHARACTER.md): full YAML spec with examples.
- [RELEASING.md](docs/RELEASING.md): release and install process.
- [ROADMAP.md](docs/ROADMAP.md): where we're headed.

---

## License

[MIT](LICENSE). Copyright (c) 2026 snowztech.
