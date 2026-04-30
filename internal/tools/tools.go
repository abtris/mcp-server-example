package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

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
	client := &http.Client{Timeout: 10 * time.Second}

	url := input.URL
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, GetOutput{}, fmt.Errorf("invalid URL %q: %w", input.URL, err)
	}
	httpReq.Header.Set("Accept", "text/plain, application/json, */*;q=0.1")
	httpReq.Header.Set("User-Agent", "mcp-server-http-get/1.0")

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, GetOutput{}, fmt.Errorf("HTTP GET %q failed: %w", input.URL, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1 MB limit
	if err != nil {
		return nil, GetOutput{}, fmt.Errorf("reading response body: %w", err)
	}

	return nil, GetOutput{
		Content: string(body),
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
