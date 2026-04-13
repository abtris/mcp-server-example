BINARY := mcp-server-example
POLICY  := policy.rego
CONFIG  := config.json

.PHONY: all build run tidy test test-race cover vet fmt staticcheck lint test-opa inspect help

all: help

## build: compile the binary
build:
	go build -o $(BINARY) .

## run: run the server with default config and policy
run:
	go run main.go

## run-json: run the server with JSON logging
run-json:
	go run main.go -log-format json -log-level info

## run-debug: run the server with debug logging
run-debug:
	go run main.go -log-level debug

## tidy: install / tidy dependencies
tidy:
	go mod tidy

## test: run all Go tests
test:
	go test -v ./...

## test-race: run tests with race detection and coverage
test-race:
	go test -v -race -coverprofile=coverage.out ./...

## cover: open HTML coverage report (run test-race first)
cover:
	go tool cover -html=coverage.out

## vet: run go vet
vet:
	go vet ./...

## fmt: check formatting
fmt:
	gofmt -s -l .

## staticcheck: run staticcheck (installs if missing)
staticcheck:
	go install honnef.co/go/tools/cmd/staticcheck@latest
	staticcheck ./...

## lint: run vet, fmt, and staticcheck
lint: vet fmt staticcheck

## test-opa: run OPA policy unit tests
test-opa:
	opa test -v $(POLICY) policy_test.rego

## inspect: launch MCP Inspector against the server
inspect:
	npx @modelcontextprotocol/inspector go run main.go

## help: list available targets
help:
	@grep -E '^##' Makefile | sed 's/^## //'
