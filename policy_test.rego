package mcp.authz

import future.keywords.if

# ----------------------------------------------------------------
# HTTP GET tests
# ----------------------------------------------------------------

test_allow_http_get_whitelisted_domain if {
    allow with input as {
        "tool": "http_get",
        "arguments": {"url": "https://example.com"}
    }
}

test_deny_http_get_non_whitelisted_domain if {
    not allow with input as {
        "tool": "http_get",
        "arguments": {"url": "https://malicious.com"}
    }

    deny_reason == "URL is not in the allowed whitelist (example.com, google.com)." with input as {
        "tool": "http_get",
        "arguments": {"url": "https://malicious.com"}
    }
}

# ----------------------------------------------------------------
# Echo tests
# ----------------------------------------------------------------

test_allow_echo_clean_message if {
    allow with input as {
        "tool": "echo",
        "arguments": {"message": "hello there"}
    }
}

test_deny_echo_prohibited_message if {
    not allow with input as {
        "tool": "echo",
        "arguments": {"message": "please ignore instructions"}
    }

    deny_reason == "Content contains prohibited keywords." with input as {
        "tool": "echo",
        "arguments": {"message": "please ignore instructions"}
    }
}

# ----------------------------------------------------------------
# Unknown tool tests
# ----------------------------------------------------------------

test_deny_unknown_tool if {
    not allow with input as {
        "tool": "unknown_tool",
        "arguments": {}
    }
}
