# Releasing

Vikusha releases are automated with Release Please.

Release Please watches conventional commits on `main`, opens or updates a release PR, maintains `CHANGELOG.md`, and creates a GitHub Release when the release PR is merged.

## Install the CLI

Install the latest release binary:

```bash
curl -sSL https://raw.githubusercontent.com/snowztech/vikusha/main/install.sh | bash
vikusha version
```

The install script downloads the latest GitHub Release binary for your OS and architecture.

## Install From Source

Install the latest released CLI with Go:

```bash
go install github.com/snowztech/vikusha/cmd/vikusha@latest
vikusha version
```

Install a specific version:

```bash
go install github.com/snowztech/vikusha/cmd/vikusha@v0.0.2
vikusha version
```

`go install` builds from source. The installed binary prints `dev` because Go does not pass the Git tag into `-ldflags` during `go install`.

## Install From GitHub Releases

Tagged releases build precompiled binaries for:

- Linux amd64/arm64
- macOS amd64/arm64
- Windows amd64

Download the matching binary from GitHub Releases, then verify it with `checksums.txt`.

Binary names are stable so `install.sh` can fetch them:

- `vikusha-linux-amd64`
- `vikusha-linux-arm64`
- `vikusha-darwin-amd64`
- `vikusha-darwin-arm64`
- `vikusha-windows-amd64.exe`

Release binaries are built with:

```bash
-ldflags "-s -w -X main.version=${VERSION}"
```

So `vikusha version` prints the release tag.

## Cutting a Release

Use conventional commit messages:

```bash
feat(agent): inject memory into prompts
fix(cli): handle missing character path
docs: clarify install instructions
```

Version bumps follow Release Please's conventional-commit rules:

- `fix:` creates a patch release, for example `v0.0.3` -> `v0.0.4`.
- `feat:` creates a minor release once the project is `v1.0.0` or later.
- Before `v1.0.0`, `feat:` is configured to create a patch release so early versions stay conservative.
- Breaking changes create a major release after `v1.0.0`.
- Before `v1.0.0`, breaking changes create a minor release.

Breaking changes can be written as:

```bash
feat!: change character config schema
```

or with a commit footer:

```text
BREAKING CHANGE: character config schema changed
```

After commits land on `main`, `.github/workflows/release-please.yml` opens or updates a release PR. Merging that PR:

1. Updates `CHANGELOG.md`.
2. Creates the release tag.
3. Creates the GitHub Release.
4. Builds release binaries.
5. Uploads binaries and `checksums.txt` to the release.

## Release Flow

Use one release path:

1. Merge normal feature/fix/docs PRs into `main`.
2. Release Please opens or updates a release PR.
3. Merge the release PR when you want to cut a version.
4. Release Please creates the tag and GitHub Release.
5. The release workflow builds binaries and uploads `checksums.txt`.

Do not push release tags manually during normal development. Let Release Please own tags, changelog updates, and GitHub Release creation.
