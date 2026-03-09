PROTO_DIR   := api
PROTO_FILE  := $(PROTO_DIR)/img.proto
GO          := go
GOPATH_BIN  := $(shell go env GOPATH)/bin

.PHONY: all build test lint proto clean

all: build

## build: compile both binaries
build:
	$(GO) build -o bin/server ./cmd/server
	$(GO) build -o bin/worker ./cmd/worker

## test: run all tests with race detector
test:
	$(GO) test -race -count=1 ./...

## test/cover: run tests and open HTML coverage report
test/cover:
	$(GO) test -race -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out

## proto: regenerate Go code from .proto files (requires protoc + plugins)
proto:
	PATH="$$PATH:$(GOPATH_BIN)" protoc \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		$(PROTO_FILE)

## lint: run go vet
lint:
	$(GO) vet ./...

## clean: remove build artifacts
clean:
	rm -rf bin/ coverage.out

## help: print this help message
help:
	@grep -E '^## ' Makefile | sed 's/## //'
