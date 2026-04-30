package policy

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/abtris/mcp-server-example/pkg/metrics"
	"github.com/abtris/mcp-server-example/pkg/tracing"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/open-policy-agent/opa/v1/rego"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Enforcer wraps the OPA policy engine for evaluating tool access
type Enforcer struct {
	query   rego.PreparedEvalQuery
	metrics *metrics.Metrics
}

// NewEnforcer loads the Rego file and prepares the OPA engine
func NewEnforcer(regoFile string, m *metrics.Metrics) (*Enforcer, error) {
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

	return &Enforcer{
		query:   query,
		metrics: m,
	}, nil
}

// Enforce is a Generic Middleware for the official SDK.
// It intercepts the tool call, converts the typed input into a map,
// evaluates it against OPA, and either returns an error or calls the actual handler.
func Enforce[In any, Out any](
	pe *Enforcer,
	toolName string,
	handler func(context.Context, *mcp.CallToolRequest, In) (*mcp.CallToolResult, Out, error),
) func(context.Context, *mcp.CallToolRequest, In) (*mcp.CallToolResult, Out, error) {

	return func(ctx context.Context, request *mcp.CallToolRequest, input In) (*mcp.CallToolResult, Out, error) {
		tracer := tracing.Tracer("mcp-server")
		ctx, span := tracer.Start(ctx, "tool-call/"+toolName,
			trace.WithAttributes(attribute.String("mcp.tool", toolName)),
		)
		defer func() {
			span.End()
			slog.Info("Tracing span ended, queued for export",
				"tool", toolName,
				"trace_id", span.SpanContext().TraceID().String(),
			)
		}()

		sc := span.SpanContext()
		slog.Info("Tracing span started",
			"tool", toolName,
			"trace_id", sc.TraceID().String(),
			"span_id", sc.SpanID().String(),
			"is_recording", span.IsRecording(),
			"is_valid", sc.IsValid(),
		)

		var zeroOut Out // Zero value for Output to return on blocked requests

		// Record the request
		if pe.metrics != nil {
			pe.metrics.RecordRequest()
			pe.metrics.RecordToolCall(toolName)
		}

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
		_, policySpan := tracer.Start(ctx, "policy-evaluation",
			trace.WithAttributes(attribute.String("mcp.tool", toolName)),
		)
		startTime := time.Now()
		results, err := pe.query.Eval(ctx, rego.EvalInput(opaInput))
		evalDuration := time.Since(startTime)
		policySpan.SetAttributes(
			attribute.Float64("policy.duration_ms", float64(evalDuration.Milliseconds())),
		)

		if err != nil {
			// Fail closed if policy engine fails
			policySpan.SetStatus(codes.Error, err.Error())
			policySpan.End()
			span.SetStatus(codes.Error, "policy engine error")
			if pe.metrics != nil {
				pe.metrics.RecordPolicyError()
			}
			slog.Error("Policy engine error", "tool", toolName, "error", err)
			return ErrorResult("Policy Engine Error: " + err.Error()), zeroOut, nil
		}

		if len(results) == 0 {
			policySpan.SetStatus(codes.Error, "no results")
			policySpan.End()
			span.SetStatus(codes.Error, "policy engine returned no results")
			if pe.metrics != nil {
				pe.metrics.RecordPolicyError()
			}
			slog.Error("Policy engine returned no results", "tool", toolName)
			return ErrorResult("Policy Engine returned no results"), zeroOut, nil
		}

		// 3. Parse Decision
		bindings := results[0].Bindings["x"].(map[string]interface{})
		allowed, ok := bindings["allow"].(bool)

		if !ok || !allowed {
			reason := "Action blocked by security policy."
			if r, exists := bindings["reason"]; exists && r != nil {
				reason = fmt.Sprintf("Blocked: %v", r)
			}

			policySpan.SetAttributes(
				attribute.Bool("policy.allowed", false),
				attribute.String("policy.deny_reason", reason),
			)
			policySpan.SetStatus(codes.Error, "denied")
			policySpan.End()
			span.SetStatus(codes.Error, "policy denied")

			// Record policy denial
			if pe.metrics != nil {
				pe.metrics.RecordPolicyEvaluation(toolName, false, reason, evalDuration.Seconds())
			}

			slog.Warn("Policy blocked action", "tool", toolName, "reason", reason)

			// Return an MCP "Error" Result. This tells the LLM the tool call failed.
			return ErrorResult(reason), zeroOut, nil
		}

		policySpan.SetAttributes(attribute.Bool("policy.allowed", true))
		policySpan.End()

		// Record policy approval
		if pe.metrics != nil {
			pe.metrics.RecordPolicyEvaluation(toolName, true, "", evalDuration.Seconds())
		}

		slog.Info("Policy allowed action", "tool", toolName)

		// 4. If Allowed, execute the actual tool logic
		ctx, toolSpan := tracer.Start(ctx, "tool-execute/"+toolName)
		toolStartTime := time.Now()
		result, output, err := handler(ctx, request, input)
		toolDuration := time.Since(toolStartTime)
		toolSpan.SetAttributes(
			attribute.Float64("tool.duration_ms", float64(toolDuration.Milliseconds())),
		)
		if err != nil {
			toolSpan.SetStatus(codes.Error, err.Error())
			span.SetStatus(codes.Error, "tool execution error")
		}
		toolSpan.End()

		if pe.metrics != nil {
			pe.metrics.RecordToolDuration(toolName, toolDuration.Seconds())
		}

		return result, output, err
	}
}

// ErrorResult creates a standard MCP error response object
func ErrorResult(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: msg},
		},
		IsError: true,
	}
}
