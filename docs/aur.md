# AUR Package

The `disgord-lyrics-bin` package metadata is stored in `packaging/aur/disgord-lyrics-bin`.

## Current Limitation

The GitHub repository is private. GitHub Release asset URLs return HTTP 404 without authentication, so an AUR build cannot download the source archive.

Do not publish the package to the AUR until the release archive is publicly downloadable:

```sh
curl -fIL https://github.com/kristyancarvalho/disGOrd-lyrics/releases/download/v0.1.0/disgord-lyrics-v0.1.0-linux-amd64.tar.gz
```

## Local Validation

Download and rename the authenticated release asset:

```sh
cd packaging/aur/disgord-lyrics-bin
gh release download v0.1.0 --repo kristyancarvalho/disGOrd-lyrics --pattern 'disgord-lyrics-v0.1.0-linux-amd64.tar.gz'
mv disgord-lyrics-v0.1.0-linux-amd64.tar.gz disgord-lyrics-bin-0.1.0.tar.gz
makepkg --printsrcinfo > .SRCINFO
makepkg -Ccf
```

`namcap` can be run when installed:

```sh
namcap PKGBUILD
namcap disgord-lyrics-bin-0.1.0-1-x86_64.pkg.tar.zst
```

## Publishing

After making the repository and release public, validate the unauthenticated URL and rebuild without a predownloaded archive:

```sh
cd packaging/aur/disgord-lyrics-bin
rm -f disgord-lyrics-bin-0.1.0.tar.gz
makepkg -Ccf
makepkg --printsrcinfo > .SRCINFO
```

Publish with the configured AUR SSH account:

```sh
rm -rf /tmp/disgord-lyrics-bin-aur
git clone ssh://aur@aur.archlinux.org/disgord-lyrics-bin.git /tmp/disgord-lyrics-bin-aur
cp PKGBUILD .SRCINFO /tmp/disgord-lyrics-bin-aur/
cd /tmp/disgord-lyrics-bin-aur
git add PKGBUILD .SRCINFO
git commit -m "Initial import of disgord-lyrics-bin"
git push
```
