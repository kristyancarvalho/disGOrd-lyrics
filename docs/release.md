# Release Process

DisGOrd Lyrics publishes GitHub Releases from `v*` tags.

## Requirements

- Go 1.26 or newer
- A clean `dev` branch synchronized with `origin/dev`
- Passing tests and local builds
- Permission to push version tags

## Checklist

1. Ensure `dev` is clean:

```sh
git checkout dev
git pull --ff-only origin dev
git status
```

2. Run tests:

```sh
go test ./...
```

3. Run the local build:

```sh
make build
```

4. Check version metadata:

```sh
make version
```

5. Build release-like archives and checksums:

```sh
make dist VERSION=vX.Y.Z
```

6. Confirm these files exist:

```text
dist/disgord-lyrics-vX.Y.Z-linux-amd64.tar.gz
dist/disgord-lyrics-vX.Y.Z-linux-arm64.tar.gz
dist/disgord-lyrics-vX.Y.Z-windows-amd64.zip
dist/checksums.txt
```

7. Inspect the archive contents:

```sh
tar -tzf dist/disgord-lyrics-vX.Y.Z-linux-amd64.tar.gz
tar -tzf dist/disgord-lyrics-vX.Y.Z-linux-arm64.tar.gz
bsdtar -tf dist/disgord-lyrics-vX.Y.Z-windows-amd64.zip
```

Each archive must contain the binary, `README.md`, `LICENSE`, `config-example.toml`, and the startup and release documentation under `docs/`.

8. Verify the checksums:

```sh
cd dist
sha256sum -c checksums.txt
cd ..
```

9. Create and push the tag:

```sh
git tag vX.Y.Z
git push origin vX.Y.Z
```

10. Confirm the `release` GitHub Actions workflow succeeds.
11. Confirm the GitHub Release contains all three archives and `checksums.txt`.
12. Download the assets and verify their checksums:

```sh
gh release download vX.Y.Z --dir /tmp/disgord-lyrics-release-check
cd /tmp/disgord-lyrics-release-check
sha256sum -c checksums.txt
```

13. Update `packaging/aur/disgord-lyrics-bin/PKGBUILD` with the release version and published Linux amd64 checksum.
14. Regenerate and validate AUR metadata:

```sh
cd packaging/aur/disgord-lyrics-bin
makepkg --printsrcinfo > .SRCINFO
makepkg -f
cd ../../..
```

15. Confirm the release archive is publicly downloadable without GitHub authentication.
16. Publish `PKGBUILD` and `.SRCINFO` to `ssh://aur@aur.archlinux.org/disgord-lyrics-bin.git`.

Private GitHub release assets cannot be used as AUR sources. See [aur.md](aur.md) for the validation and manual publishing commands.
