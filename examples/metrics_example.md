# Prometheus Metrics Example

This example demonstrates how to use the Prometheus metrics endpoint.

## Starting the Server

Start the MCP server with metrics enabled:

```bash
./mcp-server-example -metrics-port 9090
```

The server will:
1. Start the MCP server on stdio
2. Start the metrics HTTP server on port 9090
3. Expose metrics at `http://localhost:9090/metrics`

## Querying Metrics

In another terminal, query the metrics endpoint:

```bash
curl http://localhost:9090/metrics
```

## Example Metrics Output

When the server starts, you'll see initial metrics like:

```
# HELP mcp_requests_total Total number of MCP requests
# TYPE mcp_requests_total counter
mcp_requests_total 0

# HELP mcp_tool_calls_total Total number of tool calls by tool name
# TYPE mcp_tool_calls_total counter
mcp_tool_calls_total{tool="echo"} 0
mcp_tool_calls_total{tool="http_get"} 0

# HELP mcp_policy_evaluations_total Total number of policy evaluations
# TYPE mcp_policy_evaluations_total counter
mcp_policy_evaluations_total{result="allowed",tool="echo"} 0
mcp_policy_evaluations_total{result="denied",tool="echo"} 0

# HELP mcp_policy_errors_total Total number of policy evaluation errors
# TYPE mcp_policy_errors_total counter
mcp_policy_errors_total 0
```

After making some tool calls, the counters will increment:

```
mcp_requests_total 5
mcp_tool_calls_total{tool="echo"} 3
mcp_tool_calls_total{tool="http_get"} 2
mcp_policy_evaluations_total{result="allowed",tool="echo"} 3
mcp_policy_evaluations_total{result="denied",tool="http_get"} 1
```

## Prometheus Configuration

To scrape these metrics with Prometheus, add this to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'mcp-server'
    static_configs:
      - targets: ['localhost:9090']
    scrape_interval: 15s
```

## Grafana Dashboard

Create a Grafana dashboard with panels like:

**Tool Call Rate:**
```promql
rate(mcp_tool_calls_total[5m])
```

**Policy Denial Rate:**
```promql
rate(mcp_policy_evaluations_total{result="denied"}[5m])
```

**Average Tool Duration:**
```promql
rate(mcp_tool_call_duration_seconds_sum[5m]) / rate(mcp_tool_call_duration_seconds_count[5m])
```

**Policy Evaluation Latency (95th percentile):**
```promql
histogram_quantile(0.95, rate(mcp_policy_evaluation_duration_seconds_bucket[5m]))
```

