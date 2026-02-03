# Policy Library Examples

This folder contains example policies for common MCP tool-authorization patterns.
Each file is self-contained and can be adapted into your `policy.rego`.

Important notes:
- The server currently passes only `input.tool` and `input.arguments` to OPA.
- Some examples show additional inputs (like `input.actor` or `input.context`).
  If you want to use those, extend the OPA input in `internal/policy/policy.go`.

## Examples

- `tool_allowlist.rego`:
  Allow only specific tools (and example argument checks).
- `user_roles.rego`:
  Per-tool access based on user role (requires `input.actor`).
- `rate_limit.rego`:
  Simple rate limiting based on request counts (requires `input.context`
  or external `data` updates).

## Suggested Data Shapes

If you extend the OPA input or provide external data, these shapes match the examples:

```json
// input.actor
{ "id": "user_123", "role": "admin" }

// input.context
{ "request_count": 12, "window_seconds": 60 }

// data.rate_limits
{
  "http_get": { "per_minute": 30 },
  "echo": { "per_minute": 120 }
}
```
