package policy

import (
	"testing"
)

func TestNewEnforcer_Creation(t *testing.T) {
	// Test that we can create a policy enforcer with a valid policy file
	enforcer, err := NewEnforcer("../../policy.rego")
	if err != nil {
		t.Fatalf("Failed to create policy enforcer: %v", err)
	}
	if enforcer == nil {
		t.Fatal("Expected non-nil enforcer")
	}
}

func TestNewEnforcer_InvalidFile(t *testing.T) {
	// Test that we get an error with an invalid policy file
	_, err := NewEnforcer("nonexistent.rego")
	if err == nil {
		t.Fatal("Expected error for nonexistent policy file")
	}
}
