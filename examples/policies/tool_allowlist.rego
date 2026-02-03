package mcp.authz

import future.keywords.if
import future.keywords.in

# Default: deny everything
default allow := false

# Only allow these tools to be called at all.
allowed_tools := {"http_get", "echo"}

allow if {
    input.tool in allowed_tools
    tool_specific_allow
}

# Example of tool-specific constraints.
# You can remove this block if you only want a tool allowlist.

# Allow http_get for whitelisted domains.
allowed_domains := {"example.com", "google.com", "api.internal.corp"}

tool_specific_allow if {
    input.tool == "http_get"
    some domain in allowed_domains
    startswith(input.arguments.url, concat("", ["https://", domain]))
}

# Allow echo with any content (no content policy here).
# If you want content checks, see the main `policy.rego`.

tool_specific_allow if {
    input.tool == "echo"
}

# Reason for denials.

deny_reason := "Tool is not allowed." if {
    not input.tool in allowed_tools
}

deny_reason := "URL is not in the allowed whitelist." if {
    input.tool == "http_get"
    not tool_specific_allow
}
