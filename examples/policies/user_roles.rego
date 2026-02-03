package mcp.authz

import future.keywords.if
import future.keywords.in

# Default: deny everything
default allow := false

# This example requires the server to pass `input.actor` to OPA.
# Example: { "id": "user_123", "role": "admin" }

role_permissions := {
    "admin": {"http_get", "echo"},
    "analyst": {"http_get"},
    "guest": {"echo"},
}

allow if {
    role := input.actor.role
    tool := input.tool
    role in role_permissions
    tool in role_permissions[role]
}

deny_reason := "Missing actor information." if {
    not input.actor
}

deny_reason := "Role is not permitted for this tool." if {
    input.actor
    role := input.actor.role
    tool := input.tool
    not (role in role_permissions and tool in role_permissions[role])
}
