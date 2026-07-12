# Releasing

Vikusha releases are tag-driven.

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

From a clean `main`:

```bash
git pull --ff-only origin main
go test ./...
git tag v0.0.3
git push origin v0.0.3
```

Pushing the tag runs `.github/workflows/release.yml`, which tests the project, builds release binaries, writes checksums, and creates a GitHub Release.
