package mcp.authz

import future.keywords.if
import future.keywords.in

# Default: Deny everything
default allow := false
default deny_reason := null

# ----------------------------------------------------------------
# POLICY 1: HTTP GET Restrictions
# ----------------------------------------------------------------

# Whitelist of allowed domains
allowed_domains := {"example.com", "google.com", "api.internal.corp", "ifconfig.me"}

# Allow http_get ONLY if the domain is in the whitelist
allow if {
    input.tool == "http_get"
    url := input.arguments.url

    # Strip scheme if present
    stripped := trim_prefix(trim_prefix(url, "https://"), "http://")

    # Strip path if present
    host := split(stripped, "/")[0]

    some domain in allowed_domains
    host == domain
}

# Provide a specific reason if blocked
deny_reason := "URL is not in the allowed whitelist." if {
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
