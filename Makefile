BINARY := omadev
PKG := github.com/c3nk/omadev
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X $(PKG)/cmd.version=$(VERSION)

.PHONY: build test vet fmt tidy release clean

## build: compile a static binary with the version stamped in
build:
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o $(BINARY) .

## test: run all unit tests
test:
	go test ./...

## vet: run go vet
vet:
	go vet ./...

## fmt: format the tree in place
fmt:
	gofmt -l -w .

## tidy: tidy module dependencies
tidy:
	go mod tidy

## release: build release artifacts (implemented in the M2 milestone)
release:
	@echo "release target is defined in the M2 milestone (static binaries + checksums)"

## clean: remove build artifacts
clean:
	rm -f $(BINARY)
	rm -rf dist
