package logger

import (
	"log/slog"
	"testing"
)

func TestNew_TextFormat(t *testing.T) {
	cfg := Config{
		Format: "text",
		Level:  "info",
	}
	logger := New(cfg)
	if logger == nil {
		t.Fatal("Expected non-nil logger")
	}
}

func TestNew_JSONFormat(t *testing.T) {
	cfg := Config{
		Format: "json",
		Level:  "debug",
	}
	logger := New(cfg)
	if logger == nil {
		t.Fatal("Expected non-nil logger")
	}
}

func TestNew_InvalidLevel(t *testing.T) {
	cfg := Config{
		Format: "text",
		Level:  "invalid",
	}
	logger := New(cfg)
	if logger == nil {
		t.Fatal("Expected non-nil logger with default level")
	}
}

func TestSetDefault(t *testing.T) {
	cfg := Config{
		Format: "text",
		Level:  "info",
	}
	logger := New(cfg)
	SetDefault(logger)

	// Verify default logger is set
	if slog.Default() == nil {
		t.Fatal("Expected default logger to be set")
	}
}
