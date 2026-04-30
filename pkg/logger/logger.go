package logger

import (
	"context"
	"log/slog"
	"os"

	"go.opentelemetry.io/contrib/bridges/otelslog"
)

// Config holds logger configuration
type Config struct {
	Format string // "text" or "json"
	Level  string // "debug", "info", "warn", "error"
}

// New creates and configures a new slog logger
func New(cfg Config) *slog.Logger {
	handler := newStderrHandler(cfg)
	return slog.New(handler)
}

// NewWithOTel creates a logger that writes to both stderr and OTLP via the
// otelslog bridge. The OTel LoggerProvider must be initialised before calling
// this function.
func NewWithOTel(cfg Config, serviceName string) *slog.Logger {
	stderr := newStderrHandler(cfg)
	otelHandler := otelslog.NewHandler(serviceName)
	return slog.New(&multiHandler{handlers: []slog.Handler{stderr, otelHandler}})
}

// newStderrHandler builds the stderr slog.Handler according to Config.
func newStderrHandler(cfg Config) slog.Handler {
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

	opts := &slog.HandlerOptions{Level: level}
	if cfg.Format == "json" {
		return slog.NewJSONHandler(os.Stderr, opts)
	}
	return slog.NewTextHandler(os.Stderr, opts)
}

// SetDefault sets the default logger for the application
func SetDefault(logger *slog.Logger) {
	slog.SetDefault(logger)
}

// multiHandler fans out log records to multiple slog.Handler implementations.
type multiHandler struct {
	handlers []slog.Handler
}

func (m *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range m.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (m *multiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, h := range m.handlers {
		if h.Enabled(ctx, r.Level) {
			if err := h.Handle(ctx, r); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		handlers[i] = h.WithAttrs(attrs)
	}
	return &multiHandler{handlers: handlers}
}

func (m *multiHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		handlers[i] = h.WithGroup(name)
	}
	return &multiHandler{handlers: handlers}
}
