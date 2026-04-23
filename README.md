<div align="center">
    <img src="assets/vika.png" alt="Vika" width="150" style="border-radius: 10px; border-width: 1px;">
    <h1>Vika</h1>
    <h3><em>A Go framework to build AI assistants. One YAML, one binary, deploy anywhere.</em></h3>
    <p>
    <img src="https://img.shields.io/github/contributors/snowztech/vika?style=plastic" alt="Contributors">
    <img src="https://img.shields.io/github/forks/snowztech/vika" alt="Forks">
    <img src="https://img.shields.io/github/stars/snowztech/vika" alt="Stars">
    <img src="https://img.shields.io/github/issues/snowztech/vika" alt="Issues">
    <img src="https://img.shields.io/github/repo-size/snowztech/vika" alt="Repository Size">
    <a href="LICENSE">
      <img src="https://img.shields.io/badge/License-MIT-green.svg" alt="MIT License">
    </a>
    <a href="https://github.com/sponsors/lucasnevespereira">
      <img src="https://img.shields.io/badge/Sponsor-GitHub-333333?style=flat&logo=github&logoColor=white" alt="Sponsor">
    </a>
    </p>
</div>

---

An assistant has personality, tools, memory, and a transport. You define one in YAML, run it, and use it on Discord or Slack. Same binary, different YAMLs: a coding assistant, a support bot, a Discord mascot.

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
