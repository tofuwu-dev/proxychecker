# Binary name and output directory
BINARY  := proxychecker
CMD     := ./cmd/proxychecker
OUTDIR  := ./dist

# Read version from git tag, fallback to "dev" if no tag exists yet.
# This gets embedded into the binary at build time.
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

.PHONY: build test lint vet clean release help

## build: compile binary for current OS/arch
build:
	go build $(LDFLAGS) -o $(BINARY) $(CMD)

## test: run unit tests (no network calls)
test:
	go test ./... -short -race

## test-integration: run all tests including real network calls
test-integration:
	go test ./... -race

## vet: run Go's built-in static analyzer
vet:
	go vet ./...

## lint: run golangci-lint (install: https://golangci-lint.run/usage/install)
lint:
	golangci-lint run ./...

## coverage: show test coverage report in browser
coverage:
	go test ./... -short -coverprofile=coverage.out
	go tool cover -html=coverage.out

## clean: remove build artifacts
clean:
	rm -f $(BINARY)
	rm -rf $(OUTDIR)
	rm -f coverage.out

## release: cross-compile for all platforms into dist/
release:
	mkdir -p $(OUTDIR)
	GOOS=linux   GOARCH=amd64 go build $(LDFLAGS) -o $(OUTDIR)/$(BINARY)-linux-amd64   $(CMD)
	GOOS=darwin  GOARCH=amd64 go build $(LDFLAGS) -o $(OUTDIR)/$(BINARY)-darwin-amd64  $(CMD)
	GOOS=darwin  GOARCH=arm64 go build $(LDFLAGS) -o $(OUTDIR)/$(BINARY)-darwin-arm64  $(CMD)
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(OUTDIR)/$(BINARY)-windows-amd64.exe $(CMD)
	@echo "\nBuilt binaries:"
	@ls -lh $(OUTDIR)

## help: list available targets
help:
	@grep -E '^##' Makefile | sed 's/## /  /'