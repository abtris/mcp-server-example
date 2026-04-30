package metrics

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const meterName = "mcp-server"

// Metrics holds all OpenTelemetry metrics for the MCP server.
type Metrics struct {
	totalRequests        metric.Int64Counter
	toolCalls            metric.Int64Counter
	toolCallDuration     metric.Float64Histogram
	policyEvaluations    metric.Int64Counter
	policyEvaluationTime metric.Float64Histogram
	policyDenials        metric.Int64Counter
	policyErrors         metric.Int64Counter
}

// New creates and registers all OTel metrics using the global MeterProvider.
// The MeterProvider must be initialised (e.g. via tracing.InitMeter) before
// calling New; otherwise instruments record to a no-op provider.
func New() *Metrics {
	meter := otel.Meter(meterName)

	totalRequests, _ := meter.Int64Counter("mcp.requests.total",
		metric.WithDescription("Total number of MCP requests received"))

	toolCalls, _ := meter.Int64Counter("mcp.tool_calls.total",
		metric.WithDescription("Total number of tool calls by tool name"))

	toolCallDuration, _ := meter.Float64Histogram("mcp.tool_call.duration",
		metric.WithDescription("Duration of tool calls in seconds"),
		metric.WithUnit("s"))

	policyEvaluations, _ := meter.Int64Counter("mcp.policy_evaluations.total",
		metric.WithDescription("Total number of policy evaluations by tool and result"))

	policyEvaluationTime, _ := meter.Float64Histogram("mcp.policy_evaluation.duration",
		metric.WithDescription("Duration of policy evaluations in seconds"),
		metric.WithUnit("s"))

	policyDenials, _ := meter.Int64Counter("mcp.policy_denials.total",
		metric.WithDescription("Total number of policy denials by tool and reason"))

	policyErrors, _ := meter.Int64Counter("mcp.policy_errors.total",
		metric.WithDescription("Total number of policy evaluation errors"))

	return &Metrics{
		totalRequests:        totalRequests,
		toolCalls:            toolCalls,
		toolCallDuration:     toolCallDuration,
		policyEvaluations:    policyEvaluations,
		policyEvaluationTime: policyEvaluationTime,
		policyDenials:        policyDenials,
		policyErrors:         policyErrors,
	}
}

// RecordRequest increments the total request counter.
func (m *Metrics) RecordRequest() {
	m.totalRequests.Add(context.Background(), 1)
}

// RecordToolCall increments the tool call counter for a specific tool.
func (m *Metrics) RecordToolCall(toolName string) {
	m.toolCalls.Add(context.Background(), 1,
		metric.WithAttributes(attribute.String("tool", toolName)))
}

// RecordPolicyEvaluation records a policy evaluation with its result.
func (m *Metrics) RecordPolicyEvaluation(toolName string, allowed bool, reason string, durationSeconds float64) {
	ctx := context.Background()
	m.policyEvaluationTime.Record(ctx, durationSeconds,
		metric.WithAttributes(attribute.String("tool", toolName)))

	if allowed {
		m.policyEvaluations.Add(ctx, 1,
			metric.WithAttributes(
				attribute.String("tool", toolName),
				attribute.String("result", "allowed"),
			))
	} else {
		m.policyEvaluations.Add(ctx, 1,
			metric.WithAttributes(
				attribute.String("tool", toolName),
				attribute.String("result", "denied"),
			))
		if reason != "" {
			m.policyDenials.Add(ctx, 1,
				metric.WithAttributes(
					attribute.String("tool", toolName),
					attribute.String("reason", reason),
				))
		}
	}
}

// RecordPolicyError records a policy evaluation error.
func (m *Metrics) RecordPolicyError() {
	m.policyErrors.Add(context.Background(), 1)
}

// RecordToolDuration records the execution duration of a tool.
func (m *Metrics) RecordToolDuration(toolName string, durationSeconds float64) {
	m.toolCallDuration.Record(context.Background(), durationSeconds,
		metric.WithAttributes(attribute.String("tool", toolName)))
}
