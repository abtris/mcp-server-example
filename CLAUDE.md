# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

A Model Context Protocol (MCP) server (Go, `github.com/abtris/mcp-server-example`, Go 1.26) demonstrating policy-gated tool execution. Tools registered with the official `modelcontextprotocol/go-sdk` are wrapped in a generic middleware that evaluates every call against an in-process OPA policy engine before the handler runs. Telemetry (traces, metrics, logs) is emitted over OTLP/HTTP.

Note: the README still describes a Prometheus `/metrics` HTTP endpoint and a `-metrics-port` flag — those are obsolete. Metrics are now OpenTelemetry-only and there is no metrics HTTP server. Only `-policy`, `-config`, `-log-format`, `-log-level` flags exist in `main.go`.

## Common commands

The Makefile is the canonical entry point — `make help` lists targets.

- `make build` — compile binary `mcp-server-example`
- `make run` / `make run-debug` / `make run-json` — run via `go run main.go` with different log configurations
- `make test` — `go test -v ./...`
- `make test-race` — race detector + `coverage.out`
- `make cover` — open HTML coverage (after `test-race`)
- `make lint` — `go vet`, `gofmt -s -l .`, and `staticcheck` (installs staticcheck if missing)
- `make test-opa` — `opa test -v policy.rego policy_test.rego` (requires `opa` binary)
- `make inspect` — launch `npx @modelcontextprotocol/inspector` against the server

Single test: `go test -v -run TestName ./internal/policy` (substitute package and test name).

OTLP endpoint defaults to `localhost:4318`; override with `OTEL_EXPORTER_OTLP_ENDPOINT`. For local debugging the README recommends `otelite serve`.

## Architecture

Request flow on a tool call:

1. SDK dispatches the call to a handler that was wrapped at registration time by `policy.Enforce[In, Out]` (`internal/policy/policy.go`).
2. The middleware starts an OTel span `tool-call/<name>`, JSON-marshals the typed `In` struct into a `map[string]interface{}`, and packs it as OPA input `{tool, arguments}`.
3. OPA evaluates a prepared query against `policy.rego` and returns `{allow, reason}` (binding `x`). Failures, missing results, or `allow=false` short-circuit and return an `mcp.CallToolResult{IsError: true}` — the actual handler is never invoked. The pattern is **fail-closed**.
4. On allow, the handler runs inside a `tool-execute/<name>` child span and durations are recorded via the OTel metrics in `pkg/metrics`.

Key wiring points:

- `main.go` initializes tracing, the OTel meter, the OTel log bridge, then builds `logger.NewWithOTel(...)` so `slog` records also flow as OTLP logs. Shutdown is ordered: logs → metrics → traces with a 10s timeout.
- `internal/mcp_server/mcp_server.go` reads the parsed `config.ServerConfig` and, for each tool entry, looks up the handler by `tool.Handler` string (`"http_get"` or `"echo"`) and registers the policy-wrapped version. Unknown handlers are skipped with a warning — adding a new tool requires a new `case` in `RegisterTools` plus a handler in `internal/tools/`.
- `internal/config/` validates `config.json` against the embedded `config.schema.json` (overridable by placing a sibling schema file next to the config) before unmarshaling.
- Transport is stdio (`mcp.StdioTransport`) — there is no HTTP listener.

The OPA policy (`policy.rego`, package `mcp.authz`) is the single source of authorization truth. `default allow := false` means new tools are denied until an `allow if { input.tool == "..." ... }` rule is added. `deny_reason` strings surface to the LLM as the error body. `examples/policies/` has copy-paste patterns for allowlists, RBAC, and rate limiting.

## Adding a tool

1. Define `Input`/`Output` structs and a `Handler(ctx, *mcp.CallToolRequest, In) (*mcp.CallToolResult, Out, error)` function in `internal/tools/tools.go`.
2. Add a `case "<handler>"` in `MCPServer.RegisterTools` (`internal/mcp_server/mcp_server.go`) that calls `mcp.AddTool(s.mcp, &mcp.Tool{...}, policy.Enforce(s.enforcer, tool.Name, tools.YourHandler))`.
3. Add the tool entry to `config.json` (or a custom config).
4. Add an `allow if { input.tool == "<name>" ... }` rule plus a `deny_reason` to `policy.rego`, and a corresponding test in `policy_test.rego`.
