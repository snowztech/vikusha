# Character file

A vika assistant is defined by a single YAML file. Drop a different YAML, get a different assistant — same binary.

## Minimal example

```yaml
name: Coach
model: claude-sonnet-4-6
system_prompt: You are Coach, a calm fitness assistant.
transports:
  - discord
```

That's enough to run. Memory defaults to file backend, tools default to an empty set.

## Full spec

```yaml
# Required
name: Coach                          # assistant name, used in logs and default greetings
model: claude-sonnet-4-6             # model ID; provider is inferred from ID or explicit
system_prompt: |                     # system prompt — defines persona and rules
  You are Coach, a calm and helpful fitness assistant.
  You answer concisely. You never apologize for existing.

# Optional — personality helpers appended to system_prompt if set
bio:                                 # short factual bio
  - "Built by Lucas to help with fitness."
  - "Runs on a 4GB VPS."
lore:                                # deeper backstory, injected when relevant
  - "Coach reads memory before answering."
style:                               # response style guidelines
  all:
    - "Short sentences. No fluff."
    - "When unsure, say so."
  chat:
    - "Use bullets for multi-step answers."
topics:                              # things the assistant cares about
  - "fitness"
  - "nutrition"
  - "workout programming"

# Optional — explicit provider (inferred from model ID by default)
provider:
  name: anthropic
  api_key_env: ANTHROPIC_API_KEY
  # base_url: https://api.anthropic.com  # override for compat endpoints

# Optional — tools enabled for this assistant. Defaults to empty.
tools:
  - bash
  - file_read
  - file_edit
  - web_search
  - web_fetch
  - github

# Optional — tool-specific config
tool_config:
  bash:
    allowlist:
      - "git *"
      - "ls *"
      - "cat *"
    timeout: 30s
  web_search:
    provider: tavily
    api_key_env: TAVILY_API_KEY

# Optional — transports. Multiple can run at once.
transports:
  - discord:
      token_env: DISCORD_BOT_TOKEN
      owner_id_env: DISCORD_OWNER_ID   # restrict to this user; omit for open
  - telegram:
      token_env: TELEGRAM_BOT_TOKEN
      owner_id_env: TELEGRAM_OWNER_ID
  - cli: {}                             # local REPL, no config
  - http:
      bind: 127.0.0.1:7070
      auth_token_env: VIKA_HTTP_TOKEN

# Optional — memory backend. Defaults to file.
memory:
  backend: file                        # file | sqlite | pgvector
  path: ~/.vika/coach/memory
  workspace: ~/.vika/coach/workspace   # files this agent can access
  max_entries: 200                     # cap before trimming
  # For pgvector / sqlite backends:
  # dsn_env: DATABASE_URL

# Optional — RAG pipeline. Off by default.
rag:
  enabled: true
  sources:
    - path: ~/docs/fitness-notes       # directory, file, or URL
      glob: "**/*.md"
  embedding:
    provider: openai
    model: text-embedding-3-small
    api_key_env: OPENAI_API_KEY
  retrieval:
    top_k: 5
    min_score: 0.4

# Optional — context budget overrides
context:
  max_history_tokens: 30000
  tool_result_cap: 4000
  summary_on_trim: true

# Optional — logging
logging:
  level: info                          # debug | info | warn | error
  format: text                         # text | json
  file: ~/.vika/coach/logs
```

## Validation

`vika run character.yaml` validates on startup:

- Required fields present (`name`, `model`, `system_prompt`).
- Model ID matches a known provider (or `provider:` is explicit).
- All referenced tools are registered at build time.
- All referenced env vars exist.
- Transport config matches the transport's schema.

Invalid config exits non-zero with a clear error listing all problems, not just the first one.

## Hot reload

Out of scope for v1. Stop + start the process. Fast because Go.

## Multiple assistants in one process

```bash
vika run characters/vika.yaml characters/coach.yaml characters/koda.yaml
```

Each runs independently. Each needs its own Discord/Telegram token (different bot accounts).