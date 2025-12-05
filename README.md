# Secure MCP Server (Official Go SDK)

This project demonstrates a **Model Context Protocol (MCP)** server written using the official Go SDK ([modelcontextprotocol/go-sdk](https://github.com/modelcontextprotocol/go-sdk)). It integrates **Open Policy Agent (OPA)** to enforce security and compliance rules on tool calls.

## Architecture

- **Transport**: JSON-RPC over Stdio
- **SDK**: `github.com/modelcontextprotocol/go-sdk` v1.1.0
- **Policy Engine**: OPA (`github.com/open-policy-agent/opa` v0.68.0) running in-process
- **Go Version**: 1.25.5

## How It Works

When an LLM calls a tool (e.g., `http_get`):

1. The `Enforce` generic middleware intercepts the call
2. It converts the strongly-typed input struct (e.g., `GetInput`) into a map
3. It sends the tool name and arguments to the OPA engine
4. OPA evaluates `policy.rego`
5. If `allow` is `false`, the tool execution is skipped, and an error is returned to the LLM

## Prerequisites

- **Go 1.23+** (tested with Go 1.25.5)

## Setup

The module is already initialized. To install dependencies:

```bash
go mod tidy
```

## Running the Server

```bash
go run main.go
```

The server will start and listen for MCP protocol messages on stdin/stdout.

## Testing with MCP Inspector

You can use the official MCP Inspector to test the policies:

```bash
npx @modelcontextprotocol/inspector go run main.go
```

## Test Scenarios

### ✅ Allowed URL

- **Tool**: `http_get`
- **Args**: `{"url": "https://example.com"}`
- **Result**: Success (domain is whitelisted)

### ❌ Blocked URL

- **Tool**: `http_get`
- **Args**: `{"url": "https://malicious.com"}`
- **Result**: Error - "Blocked: URL is not in the allowed whitelist..."

### ❌ Toxic Content

- **Tool**: `echo`
- **Args**: `{"message": "I want to hack the mainframe"}`
- **Result**: Error - "Blocked: Content contains prohibited keywords."

## Available Tools

### `http_get`

Simulates an HTTP GET request with domain whitelisting.

**Allowed domains** (defined in `policy.rego`):
- `example.com`
- `google.com`
- `api.internal.corp`

### `echo`

Echoes a message back with content filtering.

**Blocked keywords** (defined in `policy.rego`):
- `hack`
- `ignore instructions`
- `bypass`

## Policy Configuration

Edit `policy.rego` to customize:
- Allowed domains for `http_get`
- Prohibited keywords for `echo`
- Add new policy rules for additional tools
