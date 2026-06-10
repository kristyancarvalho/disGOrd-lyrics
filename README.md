# DisGOrd Lyrics

[![License: MIT](https://img.shields.io/github/license/kristyancarvalho/disGOrd-lyrics)](LICENSE)
[![Go version](https://img.shields.io/github/go-mod/go-version/kristyancarvalho/disGOrd-lyrics)](go.mod)
[![Active milestones](https://img.shields.io/badge/milestones-active-2563eb)](https://github.com/kristyancarvalho/disGOrd-lyrics/milestones)

DisGOrd Lyrics is a cross-platform Go application that updates a Discord custom status with the active synchronized lyric line for the currently playing song.

The Go migration is in progress. The current project foundation provides the CLI structure, build metadata, configuration template, and development workflow. Media detection, lyrics lookup, and Discord status updates will be added in milestone-driven stages.

## Planned Platforms

- Windows amd64
- Linux amd64
- Linux arm64

## Installation

Versioned binaries will be published through [GitHub Releases](https://github.com/kristyancarvalho/disGOrd-lyrics/releases) after the release pipeline is implemented.

To build the current foundation from source:

```sh
git clone https://github.com/kristyancarvalho/disGOrd-lyrics.git
cd disGOrd-lyrics
make build
```

## Configuration

The planned default paths are:

- Linux: `~/.config/disgord-lyrics/config.toml`
- Windows: `%ProgramData%\DisGOrd Lyrics\config.toml`
- Windows fallback: `%APPDATA%\DisGOrd Lyrics\config.toml`

Use [config-example.toml](config-example.toml) as the configuration template.

The Discord credential must remain private. Never commit it, print it in logs, include it in screenshots, or share a populated configuration file.

Direct custom status updates may require a Discord user token and an unsupported API. This can violate Discord's Terms of Service and may put the account at risk. The implementation will be isolated and documented before it is enabled.

## Planned Usage

```sh
disgord-lyrics run
disgord-lyrics init
disgord-lyrics config-path
disgord-lyrics version
disgord-lyrics help
```

The current foundation implements `version` and `help`.

## Development

Requirements:

- Go 1.26 or newer
- GNU Make
- Git

```sh
go test ./...
make build
./bin/disgord-lyrics version
```

The initial package layout is:

| Path | Purpose |
|------|---------|
| `cmd/disgord-lyrics` | CLI entrypoint |
| `internal/app` | Command dispatch and runtime wiring |
| `internal/version` | Build-time version metadata |

Active work is tracked in [GitHub milestones](https://github.com/kristyancarvalho/disGOrd-lyrics/milestones). Development is integrated through `dev` from focused `stage/<issue-number>-<description>` branches.

## Releases

Releases will be created from `v*` tags after the release workflow is added. See [docs/release.md](docs/release.md) for the release checklist and planned artifacts.

## License

DisGOrd Lyrics is released under the [MIT License](LICENSE).
