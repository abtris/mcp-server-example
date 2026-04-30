package metrics

import (
	"testing"
)

func TestNew(t *testing.T) {
	m := New()
	if m == nil {
		t.Fatal("Expected non-nil metrics")
	}
	if m.totalRequests == nil {
		t.Error("Expected totalRequests to be initialized")
	}
	if m.toolCalls == nil {
		t.Error("Expected toolCalls to be initialized")
	}
	if m.policyEvaluations == nil {
		t.Error("Expected policyEvaluations to be initialized")
	}
}

func TestRecordRequest(t *testing.T) {
	m := New()
	m.RecordRequest()
	// No panic means success
}

func TestRecordToolCall(t *testing.T) {
	m := New()
	m.RecordToolCall("test_tool")
	// No panic means success
}

func TestRecordPolicyEvaluation(t *testing.T) {
	m := New()

	// Test allowed
	m.RecordPolicyEvaluation("test_tool", true, "", 0.001)

	// Test denied
	m.RecordPolicyEvaluation("test_tool", false, "test reason", 0.002)

	// No panic means success
}

func TestRecordPolicyError(t *testing.T) {
	m := New()
	m.RecordPolicyError()
	// No panic means success
}

func TestRecordToolDuration(t *testing.T) {
	m := New()
	m.RecordToolDuration("test_tool", 0.5)
	// No panic means success
}
