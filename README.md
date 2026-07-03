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

```bash
go install github.com/snowztech/vikusha/cmd/vikusha@latest
vikusha chat character.yaml
```

Or embed the harness in your own Go binary:

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/snowztech/vikusha"
)

func main() {
	a, err := vikusha.LoadAgent("character.yaml", vikusha.Options{})
	if err != nil {
		log.Fatal(err)
	}

	reply, err := a.Chat(context.Background(), "lucas", "hello")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(reply)
}
```

---

## Documentation

- [ARCHITECTURE.md](docs/ARCHITECTURE.md): core concepts, interfaces, directory layout.
- [CHARACTER.md](docs/CHARACTER.md): full YAML spec with examples.
- [ROADMAP.md](docs/ROADMAP.md): where we're headed.

---

## License

[MIT](LICENSE). Copyright (c) 2026 snowztech.
