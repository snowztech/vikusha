# Character File

A Vikusha assistant is defined by a YAML file. The file describes the assistant's name, model, system prompt, provider, memory, and enabled tools.

The loader is strict: unknown fields are rejected so typos do not silently change how an assistant runs.

## Minimal Example

```yaml
name: Helper
model: gpt-4o-mini
system_prompt: You are a concise assistant.
provider:
  name: openai
  api_key_env: OPENAI_API_KEY
```

Run it with:

```bash
vikusha chat character.yaml
```

Or create a named agent:

```bash
vikusha create helper
vikusha chat helper
```

## Current Schema

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
  - file_read
```

## Fields

`name` is required. It is the assistant name shown by the CLI and used by the runtime.

`model` is required. If `provider.name` is omitted, models beginning with `claude` use Anthropic; every other model uses the OpenAI-compatible provider.

`system_prompt` is required. This is the assistant's persona and instruction block.

`provider` is optional. Supported names today are `anthropic`, `openai`, `openrouter`, `groq`, `lmstudio`, and OpenAI-compatible aliases. `base_url` is used for compatible endpoints.

`memory` is optional. The implemented backend today is `file`, backed by JSONL files at `memory.path`.

`tools` is optional. The implemented built-in tool today is `file_read`.

## Validation

`vikusha chat character.yaml` validates on startup:

- Required fields are present: `name`, `model`, `system_prompt`.
- Unknown YAML fields are rejected.
- Provider and memory names are supported.
- Tool names are registered in the current build.
- Required provider API key env vars are set before the agent starts.

Invalid config exits non-zero with a clear error.

## Target Runtime

The roadmap adds named agents, transports, richer tool config, context budgets, logging, and RAG. Those fields are intentionally not accepted yet; they will be added to this schema as they become implemented.
