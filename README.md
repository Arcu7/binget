# binget

`binget` is a simple Go CLI tool using [urfave/cli](urfave/cli) I created to install binaries from GitHub releases in a quick way. I created this tool to learn more about Go (specifically the stdlib packages like tar and new features like Go's iterator) and to help me easily install command-line tools that are distributed as GitHub releases, without having to manually download them with wget and extract/place them manually.

## What It Does

- Downloads tar.gz release binaries from a GitHub repository URL
- Supports installation for Linux and macOS (not tested) with different architectures (amd64, arm64)
- Supports user or system install mode
- Accepts a GitHub token to avoid API rate limits

## Future Plans

- Add support for ZIP releases (currently only tar.gz is supported)
- Add support for more platforms (e.g. Windows)

## Development Requirements

- Go 1.26+

Recommend using a Linux based environment and installing [mise](https://mise.jdx.dev/) to manage the dev tools automatically as there is a mise.toml that specifies the versions.

## Build

```bash
make build
```

This creates the binary at `bin/binget`.

## Usage

```bash
binget [options] <repo>
```

Example:

```bash
binget https://github.com/owner/repo
```

## Common Options

- `-m, --install-mode` : Install mode (`user` or `system`), default is `user`
- `-t, --token` : GitHub token (or use `BINGET_GITHUB_TOKEN`)
- `-v, --verbose` : Enable verbose logs

## Notes

- `user` install mode expects `~/.local/bin` in your `PATH`
- `system` install mode expects `/usr/local/bin` in your `PATH`
