# Secure MCP Server (Official Go SDK)

This project demonstrates a **Model Context Protocol (MCP)** server written using the official Go SDK ([modelcontextprotocol/go-sdk](https://github.com/modelcontextprotocol/go-sdk)). It integrates **Open Policy Agent (OPA)** to enforce security and compliance rules on tool calls.

## Architecture

- **Transport**: JSON-RPC over Stdio
- **SDK**: `github.com/modelcontextprotocol/go-sdk`
- **Policy Engine**: OPA (`github.com/open-policy-agent/opa`) running in-process
- **Logging**: Structured logging with `log/slog` (JSON and text formats)

### Package Organization

The project follows Go best practices with a clear separation of concerns:

- **`main.go`** - Application entry point, CLI flag parsing, and initialization
- **`internal/`** - Internal packages (not importable by external projects)
  - **`config/`** - Configuration loading and management
  - **`policy/`** - OPA policy enforcement and middleware
  - **`server/`** - MCP server setup and tool registration
  - **`tools/`** - Tool implementations (http_get, echo)
- **`pkg/`** - Public packages (can be imported by other projects)
  - **`logger/`** - Structured logging configuration

### Project Structure

```
.
├── main.go                      # Application entry point
├── config.json                  # Server and tool configuration
├── config-minimal.json          # Example minimal configuration
├── policy.rego                  # OPA policy rules
├── internal/                    # Internal packages
│   ├── config/                  # Configuration management
│   │   ├── config.go
│   │   └── config_test.go
│   ├── policy/                  # Policy enforcement
│   │   ├── policy.go
│   │   └── policy_test.go
│   ├── mcp_server/              # MCP server implementation
│   │   └── server.go
│   └── tools/                   # Tool definitions and handlers
│       ├── tools.go
│       └── tools_test.go
├── pkg/                         # Public packages
│   └── logger/                  # Structured logging
│       ├── logger.go
│       └── logger_test.go
└── .github/
    └── workflows/
        └── go.yml               # CI/CD workflow
```

## How It Works

When an LLM calls a tool (e.g., `http_get`):

1. The `Enforce` generic middleware intercepts the call
2. It converts the strongly-typed input struct (e.g., `GetInput`) into a map
3. It sends the tool name and arguments to the OPA engine
4. OPA evaluates `policy.rego`
5. If `allow` is `false`, the tool execution is skipped, and an error is returned to the LLM

## Prerequisites

- **Go** (see `go.mod` for minimum version)

## Setup

The module is already initialized. To install dependencies:

```bash
go mod tidy
```

(Optional) Customize the configuration by editing `config.json` to enable/disable tools or change server settings.

## Running the Server

### Basic Usage

```bash
# Run with default policy file (policy.rego)
go run main.go

# Or use the compiled binary
./mcp-server-example
```

### Command Line Options

```bash
# Specify a custom policy file
go run main.go -policy /path/to/custom-policy.rego

# Specify a custom configuration file
go run main.go -config /path/to/custom-config.json

# Use both custom config and policy
go run main.go -config config-minimal.json -policy policy.rego

# Use JSON logging for production
go run main.go -log-format json -log-level info

# Debug mode with detailed logging
go run main.go -log-level debug

# Quiet mode (warnings and errors only)
go run main.go -log-level warn

# Show help
go run main.go -h
```

**Available flags:**
- `-config` - Path to the server configuration file (default: `config.json`)
- `-policy` - Path to the OPA policy file (default: `policy.rego`)
- `-log-format` - Log format: `text` or `json` (default: `text`)
- `-log-level` - Log level: `debug`, `info`, `warn`, `error` (default: `info`)
- `-metrics-port` - Port for Prometheus metrics endpoint (default: `9090`)

The server will start and listen for MCP protocol messages on stdin/stdout.

### Logging Options

The server uses structured logging with `slog` (Go's standard structured logging library).

**Text format (default):**
```bash
./mcp-server-example
# Output:
# time=2025-12-05T15:20:24.004+01:00 level=INFO msg="Loading configuration" file=config.json
# time=2025-12-05T15:20:24.006+01:00 level=INFO msg="Creating MCP server" name=SecureGoMCP version=1.0.0
```

**JSON format (for log aggregation):**
```bash
./mcp-server-example -log-format json
# Output:
# {"time":"2025-12-05T15:21:15.131+01:00","level":"INFO","msg":"Loading configuration","file":"config.json"}
# {"time":"2025-12-05T15:21:15.133+01:00","level":"INFO","msg":"Creating MCP server","name":"SecureGoMCP","version":"1.0.0"}
```

**Log levels:**
- `debug` - Detailed debugging information
- `info` - General informational messages (default)
- `warn` - Warning messages and policy blocks
- `error` - Error messages only

```bash
# Only show warnings and errors
./mcp-server-example -log-level warn

# Show all debug information
./mcp-server-example -log-level debug

# JSON format with debug level
./mcp-server-example -log-format json -log-level debug
```

## Metrics

The server exposes Prometheus metrics on an HTTP endpoint for monitoring tool calls, policy evaluations, and performance.

### Metrics Endpoint

By default, metrics are available at `http://localhost:9090/metrics`. You can change the port with the `-metrics-port` flag.

```bash
# Start server with custom metrics port
./mcp-server-example -metrics-port 8080

# In another terminal, query metrics
curl http://localhost:8080/metrics
```

### Available Metrics

**Request Metrics:**
- `mcp_requests_total` - Total number of MCP requests received

**Tool Call Metrics:**
- `mcp_tool_calls_total{tool="..."}` - Total number of tool calls by tool name
- `mcp_tool_call_duration_seconds{tool="..."}` - Duration of tool calls in seconds (histogram)

**Policy Engine Metrics:**
- `mcp_policy_evaluations_total{tool="...", result="allowed|denied"}` - Total policy evaluations by result
- `mcp_policy_evaluation_duration_seconds` - Duration of policy evaluations in seconds (histogram)
- `mcp_policy_denials_total{tool="...", reason="..."}` - Total policy denials by tool and reason
- `mcp_policy_errors_total` - Total number of policy evaluation errors

### Example Prometheus Queries

```promql
# Rate of tool calls per second
rate(mcp_tool_calls_total[5m])

# Policy denial rate
rate(mcp_policy_evaluations_total{result="denied"}[5m])

# Average policy evaluation time
rate(mcp_policy_evaluation_duration_seconds_sum[5m]) / rate(mcp_policy_evaluation_duration_seconds_count[5m])

# 95th percentile tool call duration
histogram_quantile(0.95, rate(mcp_tool_call_duration_seconds_bucket[5m]))
```

### Grafana Dashboard

You can create a Grafana dashboard to visualize these metrics. Import the metrics into Prometheus and configure Grafana to query your Prometheus instance.

## Debugging with otelite

[otelite](https://github.com/planetf1/otelite) is a lightweight, single-binary OpenTelemetry receiver with a built-in web dashboard and terminal UI. It stores traces, metrics, and logs in SQLite — no Jaeger, Prometheus, or other infrastructure required.

### Install

```bash
# macOS (Homebrew)
brew install planetf1/tap/otelite

# Or with Cargo
cargo install otelite
```

### Quick start

Start the otelite server (receives OTLP on ports 4317/4318, dashboard on port 3000):

```bash
otelite serve
```

In a separate terminal, run the MCP server (or use the Inspector):

```bash
go run main.go
```

Open http://localhost:3000 in your browser to view traces, metrics, and logs.

### CLI and TUI

```bash
# List recent traces
otelite traces list

# Show trace details
otelite traces show <trace-id>

# List recent logs
otelite logs list --severity ERROR --since 1h

# Search logs
otelite logs search "policy"

# List metrics
otelite metrics list

# Launch the terminal UI
otelite tui
```

### Overriding the endpoint

By default the MCP server sends OTLP data to `localhost:4318`. To use a different collector, set the standard environment variable:

```bash
OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4320 go run main.go
```

## Configuration

### Server Configuration (`config.json`)

The server configuration is defined in a JSON file that specifies server metadata and available tools:

```json
{
  "server": {
    "name": "SecureGoMCP",
    "version": "1.0.0"
  },
  "tools": [
    {
      "name": "http_get",
      "description": "Fetch a website. Subject to strict domain policies.",
      "handler": "http_get"
    },
    {
      "name": "echo",
      "description": "Echo a message back.",
      "handler": "echo"
    }
  ]
}
```

**Configuration fields:**
- `server.name` - Name of the MCP server
- `server.version` - Version of the MCP server
- `tools` - Array of tool definitions
  - `name` - Tool name (must be unique)
  - `description` - Tool description shown to clients
  - `handler` - Handler function name (`http_get` or `echo`)

You can create custom configuration files to enable/disable tools or change server metadata without modifying code.

The config file is validated at startup against `internal/config/config.schema.json`
(embedded into the binary), with clear validation errors if required fields
are missing or malformed. You can also place a `config.schema.json` next to
your config file to override the embedded schema.

## Testing with MCP Inspector

You can use the official MCP Inspector to test the policies:

```bash
# With default policy
npx @modelcontextprotocol/inspector go run main.go

# With custom policy
npx @modelcontextprotocol/inspector go run main.go -policy custom-policy.rego
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

## Policy Library Examples

See `examples/policies/` for ready-to-copy policy patterns, including:
- Tool allowlists and argument constraints
- Role-based access control
- Rate limiting (with external data or input context)

Each example includes notes about any extra input fields you may need to pass
to OPA.

## Development

### Continuous Integration

This project uses GitHub Actions for CI/CD. The workflow automatically:

- **Validates Go build**
- **Runs tests** with race detection and coverage reporting
- **Performs code quality checks**:
  - `go vet` - examines Go source code and reports suspicious constructs
  - `go fmt` - checks code formatting
  - `staticcheck` - advanced static analysis

The CI workflow runs on:
- Push to `main` or `develop` branches
- Pull requests targeting `main` or `develop` branches

### Running Tests Locally

```bash
# Run all tests
go test -v ./...

# Run tests with race detection and coverage
go test -v -race -coverprofile=coverage.out ./...

# View coverage report
go tool cover -html=coverage.out
```

### Policy Tests (OPA)

The OPA policy has unit tests in `policy_test.rego`.

```bash
# Run policy tests
opa test -v policy.rego policy_test.rego
```

### Code Quality Checks

```bash
# Run go vet
go vet ./...

# Check formatting
gofmt -s -l .

# Install and run staticcheck
go install honnef.co/go/tools/cmd/staticcheck@latest
staticcheck ./...
```
