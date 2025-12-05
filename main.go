package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	// Parse command line flags
	policyFile := flag.String("policy", "policy.rego", "Path to the OPA policy file")
	flag.Parse()

	// Initialize Policy Engine
	log.Printf("Loading policy from: %s", *policyFile)
	enforcer, err := NewPolicyEnforcer(*policyFile)
	if err != nil {
		log.Fatalf("Failed to start policy engine: %v", err)
	}

	// Create MCP Server
	s := mcp.NewServer(&mcp.Implementation{
		Name:    "SecureGoMCP",
		Version: "1.0.0",
	}, nil)

	// Register "http_get" with Policy Middleware
	mcp.AddTool(s, &mcp.Tool{
		Name:        "http_get",
		Description: "Fetch a website. Subject to strict domain policies.",
	}, Enforce(enforcer, "http_get", safeGetHandler))

	// Register "echo" with Policy Middleware
	mcp.AddTool(s, &mcp.Tool{
		Name:        "echo",
		Description: "Echo a message back.",
	}, Enforce(enforcer, "echo", echoHandler))

	// Start the Server (Stdio)
	fmt.Fprintln(os.Stderr, "Starting Secure MCP Server (Official SDK)...")
	if err := s.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
	}
}
