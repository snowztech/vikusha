<p align="center">
  <img src="assets/vika.png" width="200" alt="Vika" />
</p>

# Vika

**A Go framework to build always-on AI assistants. One YAML, one binary, deploy anywhere.**

You define an assistant in YAML, run it, and use it on Discord or Slack.
Vika ships with a default assistant ready to use.

## The idea

An assistant has personality, tools, memory, and a transport. Same binary,
different YAMLs. One for a coding assistant, one for a support bot,
one for a Discord mascot. Configure in YAML, run, done.

---

## Quickstart

```bash
go install github.com/snowztech/vika/cmd/vika@latest
vika setup
vika run
```

---

## What is Vika?

- Define assistants in YAML
- Connect via Discord or Slack
- Built-in tools: bash, file, web search, GitHub
- Per-agent workspace isolation
- Chat via terminal with `vika chat`

---

## Documentation

- [ARCHITECTURE.md](docs/ARCHITECTURE.md) — core concepts, interfaces, directory layout
- [CHARACTER.md](docs/CHARACTER.md) — full YAML spec with examples
- [ROADMAP.md](docs/ROADMAP.md) - where we're headed

---

## License

MIT
