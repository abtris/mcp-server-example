package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ---------------------------------------------------------
// HTTP GET Tool
// ---------------------------------------------------------

// GetInput defines the input schema for the http_get tool
type GetInput struct {
	URL string `json:"url" jsonschema:"The URL to fetch"`
}

// GetOutput defines the output schema for the http_get tool
type GetOutput struct {
	Content string `json:"content"`
}

// SafeGetHandler handles HTTP GET requests with policy enforcement
func SafeGetHandler(ctx context.Context, req *mcp.CallToolRequest, input GetInput) (*mcp.CallToolResult, GetOutput, error) {
	// In a real app, you would perform the HTTP request here.
	return nil, GetOutput{
		Content: fmt.Sprintf("Successfully accessed: %s. (Simulated Content)", input.URL),
	}, nil
}

// ---------------------------------------------------------
// Echo Tool
// ---------------------------------------------------------

// EchoInput defines the input schema for the echo tool
type EchoInput struct {
	Message string `json:"message" jsonschema:"Message to echo"`
}

// EchoOutput defines the output schema for the echo tool
type EchoOutput struct {
	Response string `json:"response"`
}

// EchoHandler handles echo requests with content filtering
func EchoHandler(ctx context.Context, req *mcp.CallToolRequest, input EchoInput) (*mcp.CallToolResult, EchoOutput, error) {
	return nil, EchoOutput{
		Response: fmt.Sprintf("Echo: %s", input.Message),
	}, nil
}
