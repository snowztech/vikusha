# Changelog

## [0.0.5](https://github.com/snowztech/vikusha/compare/v0.0.4...v0.0.5) (2026-07-14)


### Features

* **agent:** add per-tool execution config ([#31](https://github.com/snowztech/vikusha/issues/31)) ([06b2fef](https://github.com/snowztech/vikusha/commit/06b2fef03f7efbab0dc7349899c6af20cc9ca610))
* **agent:** add token-budgeted history ([#30](https://github.com/snowztech/vikusha/issues/30)) ([d891e1d](https://github.com/snowztech/vikusha/commit/d891e1d9356f38a341181c98a45225d892e5fe32))
* **agent:** cancel active user turns ([3df3fb6](https://github.com/snowztech/vikusha/commit/3df3fb65c72261caf934ab0c95e3a150f3274504))
* **agent:** cap tool results ([9c0aa10](https://github.com/snowztech/vikusha/commit/9c0aa100b26647b8b1e9fc8689b93ba918f431ff))
* **agent:** log structured turn events ([d697de7](https://github.com/snowztech/vikusha/commit/d697de7ab42d27c03a723979419a2caf232898a2))
* **agent:** log token usage ([1e9bf90](https://github.com/snowztech/vikusha/commit/1e9bf90079eb29c02bad8e7385cdd9473fe00051))
* **api:** clarify character-first agent creation ([#29](https://github.com/snowztech/vikusha/issues/29)) ([32c209f](https://github.com/snowztech/vikusha/commit/32c209f32f16c941be7b45a9d6362d38a44f10bd))
* **character:** configure turn logging ([#36](https://github.com/snowztech/vikusha/issues/36)) ([5469e81](https://github.com/snowztech/vikusha/commit/5469e81b064059f7cac8b11cc207f4e472d00c47))
* **cli:** add terminal turn logger ([#34](https://github.com/snowztech/vikusha/issues/34)) ([9d76478](https://github.com/snowztech/vikusha/commit/9d764787bf36d5d892e60d21b743d9f3dedb4017))
* **cli:** write named agent turn logs ([9b1fc92](https://github.com/snowztech/vikusha/commit/9b1fc92bfd4c77a13222afc0b286b6afe609c40c))
* **llm:** parse openai cache usage ([#33](https://github.com/snowztech/vikusha/issues/33)) ([5b58541](https://github.com/snowztech/vikusha/commit/5b585412376087ba65ec6dddc63dfb42582172e4))
* **llm:** retry transient provider failures ([#32](https://github.com/snowztech/vikusha/issues/32)) ([55d28ca](https://github.com/snowztech/vikusha/commit/55d28ca04922fa96f85942d06d85b7b669866abc))
* **tools:** add workspace file list ([2fd5abb](https://github.com/snowztech/vikusha/commit/2fd5abbff4cf57716b56e9cd33296e1a83926e70))
* **tools:** add workspace-scoped file edit ([#35](https://github.com/snowztech/vikusha/issues/35)) ([12daf17](https://github.com/snowztech/vikusha/commit/12daf17875794f3f1d4203e9c69500ef139601b6))
* **tools:** scope file reads to workspace ([8b643ca](https://github.com/snowztech/vikusha/commit/8b643caf72bb5a396c690b371fce929b7ff64b9c))

## [0.0.4](https://github.com/snowztech/vikusha/compare/v0.0.3...v0.0.4) (2026-07-12)


### Features

* **agent:** serialize turns per user ([4e76b23](https://github.com/snowztech/vikusha/commit/4e76b23e346140d2204c33059767a763e3457265))
* **character:** reject unknown yaml fields ([d0955e7](https://github.com/snowztech/vikusha/commit/d0955e797e142597aebdeb3af611b6cc01316de7))
* **cli:** resolve named agents ([29bbce1](https://github.com/snowztech/vikusha/commit/29bbce17130ded728b42c1f9969cd62cc079a4fa))
* **cli:** scaffold named agents ([e613333](https://github.com/snowztech/vikusha/commit/e6133335f22147fe59ec6ead010b46239899cafe))

## [0.0.3](https://github.com/snowztech/vikusha/compare/v0.0.2...v0.0.3) (2026-07-12)

### Features

- Inject configured memory entries into agent prompts.
- Add file-backed memory config to character YAML.
- Add release binary installer and stable release asset names.

### Documentation

- Clarify north star, architecture flow, install, and release process.
