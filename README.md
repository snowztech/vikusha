<p align="center">
  <img src="assets/vika.png" width="200" alt="Vika" />
</p>

<h1 align="center">Vika</h1>

<p align="center">
  <i>A Go framework to build always-on AI assistants. One YAML, one binary, deploy anywhere.</i>
</p>

You define an assistant in YAML, run it, and use it on Discord or Slack.
Vika ships with a default assistant ready to use.

An assistant has personality, tools, memory, and a transport. Same binary,
different YAMLs. One for a coding assistant, one for a support bot,
one for a Discord mascot. Configure in YAML, run, done.

## What is Vika?

Vika is the harness your assistants run on. It handles the agent loop, context engineering, prompt caching, memory, tool execution, and transport wiring. You write the character YAML, the harness does the rest.

- **Providers**: Anthropic, OpenAI-compatible, Ollama.
- **Tools**: bash, file, web search, grep, glob.
- **Memory**: file, SQLite, pgvector.
- **Transports**: terminal, Discord, Slack, Telegram.
- **Per-agent isolation**: workspace, memory, and secrets per assistant.
- **Observability**: structured per-turn logs with tokens, cache hits, and cost.
- **Scaffolder**: `vika create` spins up a new assistant.

---

## Quickstart

```bash
go install github.com/snowztech/vika/cmd/vika@latest
vika setup
vika run
```

---

## Documentation

- [ARCHITECTURE.md](docs/ARCHITECTURE.md): core concepts, interfaces, directory layout.
- [CHARACTER.md](docs/CHARACTER.md): full YAML spec with examples.
- [ROADMAP.md](docs/ROADMAP.md): where we're headed.

---

## License

[MIT](LICENSE). Copyright (c) 2026 snowztech.
