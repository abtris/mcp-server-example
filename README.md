# Secure MCP Server (Official Go SDK)

This project demonstrates a **Model Context Protocol (MCP)** server written using the official Go SDK ([modelcontextprotocol/go-sdk](https://github.com/modelcontextprotocol/go-sdk)). It integrates **Open Policy Agent (OPA)** to enforce security and compliance rules on tool calls.

## Architecture

- **Transport**: JSON-RPC over Stdio
- **SDK**: `github.com/modelcontextprotocol/go-sdk` v1.1.0
- **Policy Engine**: OPA (`github.com/open-policy-agent/opa` v0.68.0) running in-process
- **Go Version**: 1.25.5
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

- **Go 1.23+** (tested with Go 1.25.5)

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
./mcp-server-2025
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
./mcp-server-2025
# Output:
# time=2025-12-05T15:20:24.004+01:00 level=INFO msg="Loading configuration" file=config.json
# time=2025-12-05T15:20:24.006+01:00 level=INFO msg="Creating MCP server" name=SecureGoMCP version=1.0.0
```

**JSON format (for log aggregation):**
```bash
./mcp-server-2025 -log-format json
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
./mcp-server-2025 -log-level warn

# Show all debug information
./mcp-server-2025 -log-level debug

# JSON format with debug level
./mcp-server-2025 -log-format json -log-level debug
```

## Metrics

The server exposes Prometheus metrics on an HTTP endpoint for monitoring tool calls, policy evaluations, and performance.

### Metrics Endpoint

By default, metrics are available at `http://localhost:9090/metrics`. You can change the port with the `-metrics-port` flag.

```bash
# Start server with custom metrics port
./mcp-server-2025 -metrics-port 8080

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

## Development

### Continuous Integration

This project uses GitHub Actions for CI/CD. The workflow automatically:

- **Validates Go build** across multiple Go versions (1.23, 1.24, 1.25)
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
