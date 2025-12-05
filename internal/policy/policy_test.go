package policy

import (
	"sync"
	"testing"

	"github.com/abtris/mcp-server-2025/pkg/metrics"
)

var (
	testMetrics     *metrics.Metrics
	testMetricsOnce sync.Once
)

func getTestMetrics() *metrics.Metrics {
	testMetricsOnce.Do(func() {
		testMetrics = metrics.New()
	})
	return testMetrics
}

func TestNewEnforcer_Creation(t *testing.T) {
	// Test that we can create a policy enforcer with a valid policy file
	m := getTestMetrics()
	enforcer, err := NewEnforcer("../../policy.rego", m)
	if err != nil {
		t.Fatalf("Failed to create policy enforcer: %v", err)
	}
	if enforcer == nil {
		t.Fatal("Expected non-nil enforcer")
	}
}

func TestNewEnforcer_InvalidFile(t *testing.T) {
	// Test that we get an error with an invalid policy file
	m := getTestMetrics()
	_, err := NewEnforcer("nonexistent.rego", m)
	if err == nil {
		t.Fatal("Expected error for nonexistent policy file")
	}
}
