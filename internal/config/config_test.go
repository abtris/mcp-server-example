package config

import (
	"os"
	"testing"
)

func TestLoad_ValidFile(t *testing.T) {
	// Test loading the default config file
	cfg, err := Load("../../config.json")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Server.Name != "SecureGoMCP" {
		t.Errorf("Expected server name 'SecureGoMCP', got '%s'", cfg.Server.Name)
	}

	if cfg.Server.Version != "1.0.0" {
		t.Errorf("Expected server version '1.0.0', got '%s'", cfg.Server.Version)
	}

	if len(cfg.Tools) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(cfg.Tools))
	}
}

func TestLoad_MinimalFile(t *testing.T) {
	// Test loading the minimal config file
	cfg, err := Load("../../config-minimal.json")
	if err != nil {
		t.Fatalf("Failed to load minimal config: %v", err)
	}

	if cfg.Server.Name != "MinimalMCP" {
		t.Errorf("Expected server name 'MinimalMCP', got '%s'", cfg.Server.Name)
	}

	if len(cfg.Tools) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(cfg.Tools))
	}

	if cfg.Tools[0].Name != "echo" {
		t.Errorf("Expected tool name 'echo', got '%s'", cfg.Tools[0].Name)
	}
}

func TestLoad_InvalidFile(t *testing.T) {
	// Test that we get an error with a non-existent file
	_, err := Load("nonexistent.json")
	if err == nil {
		t.Fatal("Expected error for nonexistent config file")
	}
}

func TestLoad_InvalidSchema(t *testing.T) {
	// Missing server.name and tool handler should fail schema validation
	invalidJSON := `{
  "server": { "version": "1.0.0" },
  "tools": [
    { "name": "echo", "description": "Echo a message back." }
  ]
}`

	tmpFile, err := os.CreateTemp("", "invalid-config-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(invalidJSON); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	_, err = Load(tmpFile.Name())
	if err == nil {
		t.Fatal("Expected error for invalid config schema")
	}
}
