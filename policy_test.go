package main

import (
	"testing"
)

func TestPolicyEnforcer_Creation(t *testing.T) {
	// Test that we can create a policy enforcer with a valid policy file
	enforcer, err := NewPolicyEnforcer("policy.rego")
	if err != nil {
		t.Fatalf("Failed to create policy enforcer: %v", err)
	}
	if enforcer == nil {
		t.Fatal("Expected non-nil enforcer")
	}
}

func TestPolicyEnforcer_InvalidFile(t *testing.T) {
	// Test that we get an error with an invalid policy file
	_, err := NewPolicyEnforcer("nonexistent.rego")
	if err == nil {
		t.Fatal("Expected error for nonexistent policy file")
	}
}
