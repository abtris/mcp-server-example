package logger

import (
	"log/slog"
	"os"
)

// Config holds logger configuration
type Config struct {
	Format string // "text" or "json"
	Level  string // "debug", "info", "warn", "error"
}

// New creates and configures a new slog logger
func New(cfg Config) *slog.Logger {
	var handler slog.Handler
	var level slog.Level

	// Parse log level
	switch cfg.Level {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	// Create handler based on format
	opts := &slog.HandlerOptions{Level: level}
	if cfg.Format == "json" {
		handler = slog.NewJSONHandler(os.Stderr, opts)
	} else {
		handler = slog.NewTextHandler(os.Stderr, opts)
	}

	return slog.New(handler)
}

// SetDefault sets the default logger for the application
func SetDefault(logger *slog.Logger) {
	slog.SetDefault(logger)
}
