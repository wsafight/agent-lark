APP       := agent-lark
PKG       := ./cmd/$(APP)
VERSION   ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS   := -s -w -X main.Version=$(VERSION)

.PHONY: build install test lint cover clean

## build: compile binary to ./agent-lark
build:
	go build -ldflags="$(LDFLAGS)" -o $(APP) $(PKG)

## install: install to $GOPATH/bin
install:
	go install -ldflags="$(LDFLAGS)" $(PKG)

## test: run all tests
test:
	go test ./...

## lint: run golangci-lint
lint:
	golangci-lint run ./...

## cover: run tests with coverage report
cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	@rm -f coverage.out

## clean: remove build artifacts
clean:
	rm -f $(APP) coverage.out
	rm -rf dist/

## help: show this help
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/^## //' | column -t -s ':'
