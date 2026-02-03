package mcp.authz

import future.keywords.if

# Default: deny everything
default allow := false

# This example expects either:
# - input.context.request_count (requests seen within the window), and
# - data.rate_limits[tool].per_minute (limits by tool)
#
# Example input.context:
# { "request_count": 12, "window_seconds": 60 }

allow if {
    tool := input.tool
    limit := data.rate_limits[tool].per_minute
    count := input.context.request_count
    count <= limit
}

# Optional: allow tools that have no configured rate limits
allow if {
    tool := input.tool
    not data.rate_limits[tool]
}

deny_reason := "Rate limit exceeded." if {
    tool := input.tool
    data.rate_limits[tool]
    input.context.request_count > data.rate_limits[tool].per_minute
}
