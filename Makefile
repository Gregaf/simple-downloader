VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE    ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS = -ldflags="-s -w -X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.Date=$(DATE)"

.PHONY: all help build run clean test

all: help

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^## [-a-zA-Z0-9_]+:' Makefile | sed 's/## //g' | awk -F: '{printf "  %-10s %s\n", $$1, $$2}'

## build: Build the application with metadata
build:
	go build $(LDFLAGS) -o dwnld ./cmd/dwnld

## clean: Remove built binaries
clean:
	rm -f dwnld

## test: Run tests
test:
	go test ./...
