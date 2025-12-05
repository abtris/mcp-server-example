package metrics

import (
	"sync"
	"testing"
)

var (
	testMetrics     *Metrics
	testMetricsOnce sync.Once
)

func getTestMetrics() *Metrics {
	testMetricsOnce.Do(func() {
		testMetrics = New()
	})
	return testMetrics
}

func TestNew(t *testing.T) {
	m := getTestMetrics()
	if m == nil {
		t.Fatal("Expected non-nil metrics")
	}
	if m.TotalRequests == nil {
		t.Error("Expected TotalRequests to be initialized")
	}
	if m.ToolCalls == nil {
		t.Error("Expected ToolCalls to be initialized")
	}
	if m.PolicyEvaluations == nil {
		t.Error("Expected PolicyEvaluations to be initialized")
	}
}

func TestRecordRequest(t *testing.T) {
	m := getTestMetrics()
	m.RecordRequest()
	// No panic means success
}

func TestRecordToolCall(t *testing.T) {
	m := getTestMetrics()
	m.RecordToolCall("test_tool")
	// No panic means success
}

func TestRecordPolicyEvaluation(t *testing.T) {
	m := getTestMetrics()

	// Test allowed
	m.RecordPolicyEvaluation("test_tool", true, "", 0.001)

	// Test denied
	m.RecordPolicyEvaluation("test_tool", false, "test reason", 0.002)

	// No panic means success
}

func TestRecordPolicyError(t *testing.T) {
	m := getTestMetrics()
	m.RecordPolicyError()
	// No panic means success
}

func TestRecordToolDuration(t *testing.T) {
	m := getTestMetrics()
	m.RecordToolDuration("test_tool", 0.5)
	// No panic means success
}
