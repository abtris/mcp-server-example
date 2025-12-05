package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/open-policy-agent/opa/rego"
)

// PolicyEnforcer wraps the OPA policy engine for evaluating tool access
type PolicyEnforcer struct {
	query rego.PreparedEvalQuery
}

// NewPolicyEnforcer loads the Rego file and prepares the OPA engine
func NewPolicyEnforcer(regoFile string) (*PolicyEnforcer, error) {
	ctx := context.Background()

	bs, err := os.ReadFile(regoFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read policy file: %w", err)
	}

	// Prepare Rego query.
	// We query for 'x' which contains both the boolean decision and the reason.
	r := rego.New(
		rego.Query("x = { \"allow\": data.mcp.authz.allow, \"reason\": data.mcp.authz.deny_reason }"),
		rego.Module("policy.rego", string(bs)),
	)

	query, err := r.PrepareForEval(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare rego: %w", err)
	}

	return &PolicyEnforcer{query: query}, nil
}

// Enforce is a Generic Middleware for the official SDK.
// It intercepts the tool call, converts the typed input into a map,
// evaluates it against OPA, and either returns an error or calls the actual handler.
func Enforce[In any, Out any](
	pe *PolicyEnforcer,
	toolName string,
	handler func(context.Context, *mcp.CallToolRequest, In) (*mcp.CallToolResult, Out, error),
) func(context.Context, *mcp.CallToolRequest, In) (*mcp.CallToolResult, Out, error) {

	return func(ctx context.Context, request *mcp.CallToolRequest, input In) (*mcp.CallToolResult, Out, error) {
		var zeroOut Out // Zero value for Output to return on blocked requests

		// 1. Convert Typed Input to Map for OPA
		// OPA requires map[string]interface{}, but our Input is a struct.
		// We use JSON marshal/unmarshal to converting the struct to a generic map.
		inputMap := make(map[string]interface{})
		if b, err := json.Marshal(input); err == nil {
			var m map[string]interface{}
			if err := json.Unmarshal(b, &m); err == nil {
				inputMap = m
			}
		}

		opaInput := map[string]interface{}{
			"tool":      toolName,
			"arguments": inputMap,
		}

		// 2. Evaluate Policy
		results, err := pe.query.Eval(ctx, rego.EvalInput(opaInput))
		if err != nil {
			// Fail closed if policy engine fails
			return errorResult("Policy Engine Error: " + err.Error()), zeroOut, nil
		}

		if len(results) == 0 {
			return errorResult("Policy Engine returned no results"), zeroOut, nil
		}

		// 3. Parse Decision
		bindings := results[0].Bindings["x"].(map[string]interface{})
		allowed, ok := bindings["allow"].(bool)

		if !ok || !allowed {
			reason := "Action blocked by security policy."
			if r, exists := bindings["reason"]; exists && r != nil {
				reason = fmt.Sprintf("Blocked: %v", r)
			}
			log.Printf("[BLOCK] Tool: %s | Reason: %s", toolName, reason)

			// Return an MCP "Error" Result. This tells the LLM the tool call failed.
			return errorResult(reason), zeroOut, nil
		}

		log.Printf("[ALLOW] Tool: %s", toolName)

		// 4. If Allowed, execute the actual tool logic
		return handler(ctx, request, input)
	}
}

// errorResult creates a standard MCP error response object
func errorResult(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: msg},
		},
		IsError: true,
	}
}
