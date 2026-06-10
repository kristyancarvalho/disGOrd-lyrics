# Release Process

DisGOrd Lyrics will publish versioned GitHub Releases from `v*` tags after the release workflow is implemented.

## Requirements

- Go 1.26 or newer
- A clean `dev` branch
- Passing local tests
- A successful local build
- Permission to push version tags

## Checklist

1. Ensure `dev` is clean and synchronized with `origin/dev`.
2. Run `go test ./...`.
3. Run `make build`.
4. Run `make version` and confirm the version metadata.
5. Create a version tag:

```sh
git tag vX.Y.Z
```

6. Push the tag:

```sh
git push origin vX.Y.Z
```

7. Confirm the GitHub Release contains:

```text
disgord-lyrics-vX.Y.Z-linux-amd64.tar.gz
disgord-lyrics-vX.Y.Z-linux-arm64.tar.gz
disgord-lyrics-vX.Y.Z-windows-amd64.zip
checksums.txt
```

8. Verify the published checksums against the downloaded artifacts.
