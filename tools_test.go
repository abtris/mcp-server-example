package main

import (
	"testing"
)

func TestGetInput_Struct(t *testing.T) {
	// Test that GetInput struct can be created
	input := GetInput{
		URL: "https://example.com",
	}
	if input.URL != "https://example.com" {
		t.Errorf("Expected URL to be 'https://example.com', got '%s'", input.URL)
	}
}

func TestEchoInput_Struct(t *testing.T) {
	// Test that EchoInput struct can be created
	input := EchoInput{
		Message: "Hello, World!",
	}
	if input.Message != "Hello, World!" {
		t.Errorf("Expected Message to be 'Hello, World!', got '%s'", input.Message)
	}
}
