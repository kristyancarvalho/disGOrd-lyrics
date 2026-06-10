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

7. Create and push the tag:

```sh
git tag vX.Y.Z
git push origin vX.Y.Z
```

8. Confirm the `release` GitHub Actions workflow succeeds.
9. Confirm the GitHub Release contains all three archives and `checksums.txt`.
10. Download the assets and verify their checksums:

```sh
sha256sum -c checksums.txt
```
