# DisGOrd Lyrics

[![Release workflow](https://github.com/kristyancarvalho/disGOrd-lyrics/actions/workflows/release.yml/badge.svg)](https://github.com/kristyancarvalho/disGOrd-lyrics/actions/workflows/release.yml)
[![License: MIT](https://img.shields.io/github/license/kristyancarvalho/disGOrd-lyrics)](LICENSE)
[![Go version](https://img.shields.io/github/go-mod/go-version/kristyancarvalho/disGOrd-lyrics)](go.mod)
[![Latest release](https://img.shields.io/github/v/release/kristyancarvalho/disGOrd-lyrics)](https://github.com/kristyancarvalho/disGOrd-lyrics/releases)

DisGOrd Lyrics reads the active desktop media session, finds synchronized lyrics, and updates the user's Discord custom status with the current lyric line.

The application is implemented in Go and has no Python runtime dependency. Linux media detection uses MPRIS over D-Bus. Windows binaries build and provide the CLI and configuration commands, but Windows media detection is not available in this release.

## Supported Platforms

| Platform | Architecture | Media Detection |
|----------|--------------|-----------------|
| Linux | amd64 | MPRIS |
| Linux | arm64 | MPRIS |
| Windows | amd64 | Not currently supported |

## Discord Account Risk

Discord does not provide an official API for applications to update a user's custom status. DisGOrd Lyrics uses the undocumented user settings endpoint with a user token.

Using a user token for automation may violate Discord's Terms of Service, can stop working without notice, and may put the Discord account at risk. Use this application only after understanding and accepting that risk. The token is isolated in the configuration and Discord packages, but it is still a sensitive account credential.

Never commit, log, publish, screenshot, or share the token or a populated configuration file.

## Installation

Download the archive for the target platform from [GitHub Releases](https://github.com/kristyancarvalho/disGOrd-lyrics/releases):

```text
disgord-lyrics-vX.Y.Z-linux-amd64.tar.gz
disgord-lyrics-vX.Y.Z-linux-arm64.tar.gz
disgord-lyrics-vX.Y.Z-windows-amd64.zip
checksums.txt
```

Verify the archive before extracting:

```sh
sha256sum -c checksums.txt
```

On Linux, install the binary in the user path:

```sh
install -Dm755 disgord-lyrics ~/.local/bin/disgord-lyrics
```

On Windows, extract `disgord-lyrics.exe` to a stable user-owned directory.

## Configuration

Create the configuration:

```sh
disgord-lyrics init
```

Print its location:

```sh
disgord-lyrics config-path
```

Default locations:

- Linux: `~/.config/disgord-lyrics/config.toml`
- Windows: `%ProgramData%\DisGOrd Lyrics\config.toml`
- Windows fallback: `%APPDATA%\DisGOrd Lyrics\config.toml`

Edit the generated file and set `discord.token`. The default configuration is:

```toml
[discord]
token = ""

[status]
prefix = "🎵 "
max_length = 70
clear_on_pause = true
clear_on_exit = true

[lyrics]
provider = "lrclib"
offset_ms = 500

[polling]
interval_ms = 300

[logging]
level = "info"
```

`init` does not overwrite an existing file. Use `disgord-lyrics init --force` only when replacing it is intentional.

## Usage

Start the runtime:

```sh
disgord-lyrics run
```

Other commands:

```sh
disgord-lyrics init [--force]
disgord-lyrics config-path
disgord-lyrics version
disgord-lyrics help
```

The runtime clears the Discord custom status at startup, when media is paused, stopped, unavailable, or invalid, and on exit according to the configuration. Duplicate lyric lines do not produce duplicate Discord requests.

## Startup

- [Linux user systemd service](docs/linux-startup.md)
- [Windows Startup folder and Task Scheduler](docs/windows-startup.md)

The Windows startup instructions are provided for future media support and for CLI availability. The current Windows runtime exits with a clear unsupported-media error.

## Troubleshooting

### Config file not found

Run `disgord-lyrics init`, then use `disgord-lyrics config-path` to locate the generated file.

### Discord token is required

Set a non-empty `discord.token` value. The application never prints the configured value.

### No media is detected on Linux

Confirm the player exposes an MPRIS service:

```sh
busctl --user list | grep org.mpris.MediaPlayer2
```

The application must run inside the same graphical user session and D-Bus session as the media player.

### No synchronized lyrics are found

LRCLIB may not have a synchronized record for the exact title and artist reported by the player. Instrumental and unsynchronized records are ignored.

### Discord returns an HTTP error

The token may be invalid, Discord may have changed the undocumented endpoint, or the account may be rate limited. The response body and token are not logged.

### Windows reports unsupported media detection

The Windows binary is intentionally buildable, but the current provider returns an explicit unsupported error. Windows Runtime global media session integration remains a known limitation.

## Development

Requirements:

- Go 1.26 or newer
- GNU Make
- Git
- `tar`, `sha256sum`, and either `zip` or `bsdtar` for release packaging

```sh
go test ./...
make build
make version
make dist
```

Package layout:

| Path | Purpose |
|------|---------|
| `cmd/disgord-lyrics` | Minimal CLI entrypoint |
| `internal/app` | Commands, runtime wiring, polling, and shutdown |
| `internal/config` | TOML defaults, loading, validation, and initialization |
| `internal/discord` | Isolated custom status HTTP client |
| `internal/lyrics` | LRCLIB client, cache, LRC parsing, and active-line selection |
| `internal/media` | Platform-neutral media types and provider interface |
| `internal/media/linux` | Linux MPRIS provider |
| `internal/media/windows` | Explicit unsupported Windows provider |
| `internal/status` | Lyric cleaning, formatting, and duplicate suppression |
| `internal/version` | Build-time metadata |

## Releases

The `release` GitHub Actions workflow runs for `v*` tags, tests the project, creates Linux and Windows archives, generates `checksums.txt`, and publishes a GitHub Release with generated notes.

See [docs/release.md](docs/release.md) for the release checklist.

## License

DisGOrd Lyrics is released under the [MIT License](LICENSE).
