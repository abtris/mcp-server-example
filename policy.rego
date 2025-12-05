package mcp.authz

import future.keywords.if
import future.keywords.in

# Default: Deny everything
default allow := false

# ----------------------------------------------------------------
# POLICY 1: HTTP GET Restrictions
# ----------------------------------------------------------------

# Whitelist of allowed domains
allowed_domains := {"example.com", "google.com", "api.internal.corp"}

# Allow http_get ONLY if the domain is in the whitelist
allow if {
    input.tool == "http_get"
    some domain in allowed_domains
    startswith(input.arguments.url, concat("", ["https://", domain]))
}

# Provide a specific reason if blocked
deny_reason := "URL is not in the allowed whitelist (example.com, google.com)." if {
    input.tool == "http_get"
    not allow
}

# ----------------------------------------------------------------
# POLICY 2: Behavioral Checks (The "Be Nice" Policy)
# ----------------------------------------------------------------

# A list of prohibited words (e.g., trying to prompt inject or be rude)
prohibited_words := ["hack", "ignore instructions", "bypass"]

# Check the 'message' argument in the 'echo' tool
has_prohibited_content if {
    some word in prohibited_words
    contains(lower(input.arguments.message), word)
}

# Allow echo only if no prohibited words found
allow if {
    input.tool == "echo"
    not has_prohibited_content
}

deny_reason := "Content contains prohibited keywords." if {
    input.tool == "echo"
    has_prohibited_content
}
