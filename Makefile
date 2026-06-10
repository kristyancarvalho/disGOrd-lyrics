VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
MODULE := github.com/kristyancarvalho/disGOrd-lyrics
LDFLAGS ?= -s -w -X $(MODULE)/internal/version.Version=$(VERSION) -X $(MODULE)/internal/version.Commit=$(COMMIT) -X $(MODULE)/internal/version.Date=$(DATE)

.PHONY: build test clean version

build:
	go build -trimpath -ldflags="$(LDFLAGS)" -o bin/disgord-lyrics ./cmd/disgord-lyrics

test:
	go test ./...

clean:
	rm -rf bin/

version: build
	./bin/disgord-lyrics version
