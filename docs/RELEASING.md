# Releasing

Vikusha releases are tag-driven.

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

`go install` builds from source. The installed binary may print `dev` because Go does not pass the Git tag into `-ldflags` during `go install`.

## Install From GitHub Releases

Tagged releases build precompiled binaries for:

- Linux amd64/arm64
- macOS amd64/arm64
- Windows amd64

Download the matching archive from GitHub Releases, then verify it with `checksums.txt`.

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
git tag v0.0.2
git push origin v0.0.2
```

Pushing the tag runs `.github/workflows/release.yml`, which tests the project, builds release binaries, writes checksums, and creates a GitHub Release.
