package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all Prometheus metrics for the MCP server
type Metrics struct {
	// Request metrics
	TotalRequests prometheus.Counter

	// Tool call metrics
	ToolCalls        *prometheus.CounterVec
	ToolCallDuration *prometheus.HistogramVec

	// Policy engine metrics
	PolicyEvaluations    *prometheus.CounterVec
	PolicyEvaluationTime prometheus.Histogram
	PolicyDenials        *prometheus.CounterVec
	PolicyErrors         prometheus.Counter
}

// New creates and registers all Prometheus metrics
func New() *Metrics {
	return &Metrics{
		TotalRequests: promauto.NewCounter(prometheus.CounterOpts{
			Name: "mcp_requests_total",
			Help: "Total number of MCP requests received",
		}),

		ToolCalls: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "mcp_tool_calls_total",
			Help: "Total number of tool calls by tool name",
		}, []string{"tool"}),

		ToolCallDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "mcp_tool_call_duration_seconds",
			Help:    "Duration of tool calls in seconds",
			Buckets: prometheus.DefBuckets,
		}, []string{"tool"}),

		PolicyEvaluations: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "mcp_policy_evaluations_total",
			Help: "Total number of policy evaluations by tool and result",
		}, []string{"tool", "result"}),

		PolicyEvaluationTime: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "mcp_policy_evaluation_duration_seconds",
			Help:    "Duration of policy evaluations in seconds",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
		}),

		PolicyDenials: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "mcp_policy_denials_total",
			Help: "Total number of policy denials by tool and reason",
		}, []string{"tool", "reason"}),

		PolicyErrors: promauto.NewCounter(prometheus.CounterOpts{
			Name: "mcp_policy_errors_total",
			Help: "Total number of policy evaluation errors",
		}),
	}
}

// RecordRequest increments the total request counter
func (m *Metrics) RecordRequest() {
	m.TotalRequests.Inc()
}

// RecordToolCall increments the tool call counter for a specific tool
func (m *Metrics) RecordToolCall(toolName string) {
	m.ToolCalls.WithLabelValues(toolName).Inc()
}

// RecordPolicyEvaluation records a policy evaluation with its result
func (m *Metrics) RecordPolicyEvaluation(toolName string, allowed bool, reason string, durationSeconds float64) {
	m.PolicyEvaluationTime.Observe(durationSeconds)

	if allowed {
		m.PolicyEvaluations.WithLabelValues(toolName, "allowed").Inc()
	} else {
		m.PolicyEvaluations.WithLabelValues(toolName, "denied").Inc()
		if reason != "" {
			m.PolicyDenials.WithLabelValues(toolName, reason).Inc()
		}
	}
}

// RecordPolicyError records a policy evaluation error
func (m *Metrics) RecordPolicyError() {
	m.PolicyErrors.Inc()
}

// RecordToolDuration records the execution duration of a tool
func (m *Metrics) RecordToolDuration(toolName string, durationSeconds float64) {
	m.ToolCallDuration.WithLabelValues(toolName).Observe(durationSeconds)
}

