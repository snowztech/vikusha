<div align="center">
    <img src="assets/vika.png" alt="Vika" width="180" style="border-radius: 10px; border-width: 1px;">
    <h1>Vika</h1>
    <h3><em>Go framework for AI assistants that run where you need them.</em></h3>
    <p>
    <a href="https://github.com/snowztech/vika/stargazers"><img src="https://img.shields.io/github/stars/snowztech/vika?style=flat&logo=github" alt="Stars"></a>
    <a href="https://github.com/snowztech/vika/network/members"><img src="https://img.shields.io/github/forks/snowztech/vika?style=flat&logo=github" alt="Forks"></a>
    <a href="https://github.com/snowztech/vika/issues"><img src="https://img.shields.io/github/issues/snowztech/vika?style=flat&logo=github" alt="Issues"></a>
    <a href="https://github.com/snowztech/vika/graphs/contributors"><img src="https://img.shields.io/github/contributors/snowztech/vika?style=flat&logo=github" alt="Contributors"></a>
    <a href="LICENSE"><img src="https://img.shields.io/badge/License-MIT-green?style=flat" alt="License"></a>
    </p>
</div>

One YAML defines the assistant: personality, tools, memory, and how it connects to users. Same binary, different YAML—a coding assistant, a support bot, a Discord mascot.

Vika ships with a default assistant ready to use.

## What is Vika?

Vika is the harness your assistants run on. It handles the agent loop, context engineering, prompt caching, memory, tool execution, and transport wiring. You write the character YAML, the harness does the rest.

- **Providers**: Anthropic, OpenAI-compatible, Ollama.
- **Tools**: bash, file, web search, grep, glob.
- **Memory**: file, SQLite, pgvector.
- **Transports**: terminal, Discord, Slack, Telegram.
- **Per-agent isolation**: workspace, memory, and secrets per assistant.
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
