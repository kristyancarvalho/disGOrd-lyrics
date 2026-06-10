VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
MODULE := github.com/kristyancarvalho/disGOrd-lyrics
BINARY := disgord-lyrics
RELEASE_VERSION := $(if $(filter v%,$(VERSION)),$(VERSION),v0.0.0-$(VERSION))
LDFLAGS := -s -w -X $(MODULE)/internal/version.Version=$(VERSION) -X $(MODULE)/internal/version.Commit=$(COMMIT) -X $(MODULE)/internal/version.Date=$(DATE)

.PHONY: test build clean version build-linux-amd64 build-linux-arm64 build-windows-amd64 dist checksums

test:
	go test ./...

build:
	go build -trimpath -ldflags="$(LDFLAGS)" -o bin/$(BINARY) ./cmd/$(BINARY)

clean:
	rm -rf bin/ dist/

version: build
	./bin/$(BINARY) version

build-linux-amd64:
	mkdir -p dist/linux-amd64
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="$(LDFLAGS)" -o dist/linux-amd64/$(BINARY) ./cmd/$(BINARY)

build-linux-arm64:
	mkdir -p dist/linux-arm64
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -ldflags="$(LDFLAGS)" -o dist/linux-arm64/$(BINARY) ./cmd/$(BINARY)

build-windows-amd64:
	mkdir -p dist/windows-amd64
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="$(LDFLAGS)" -o dist/windows-amd64/$(BINARY).exe ./cmd/$(BINARY)

dist:
	rm -rf dist/
	$(MAKE) build-linux-amd64 build-linux-arm64 build-windows-amd64 VERSION="$(VERSION)" COMMIT="$(COMMIT)" DATE="$(DATE)"
	cp README.md LICENSE config-example.toml dist/linux-amd64/
	cp README.md LICENSE config-example.toml dist/linux-arm64/
	cp README.md LICENSE config-example.toml dist/windows-amd64/
	tar -czf dist/$(BINARY)-$(RELEASE_VERSION)-linux-amd64.tar.gz -C dist/linux-amd64 .
	tar -czf dist/$(BINARY)-$(RELEASE_VERSION)-linux-arm64.tar.gz -C dist/linux-arm64 .
	if command -v zip >/dev/null 2>&1; then cd dist/windows-amd64 && zip -q -r ../$(BINARY)-$(RELEASE_VERSION)-windows-amd64.zip .; else bsdtar -a -cf dist/$(BINARY)-$(RELEASE_VERSION)-windows-amd64.zip -C dist/windows-amd64 .; fi
	$(MAKE) checksums

checksums:
	cd dist && sha256sum *.tar.gz *.zip > checksums.txt
