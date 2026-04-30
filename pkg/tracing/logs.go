package tracing

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/log/global"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// InitLogger sets up the OpenTelemetry LoggerProvider exporting via OTLP/HTTP.
// Returns a shutdown function that must be called on application exit.
// If cfg is nil, OTLP log export is disabled.
func InitLogger(ctx context.Context, cfg *Config) (shutdown func(context.Context) error, err error) {
	if cfg == nil {
		slog.Info("OTLP logs disabled (nil config)")
		return func(context.Context) error { return nil }, nil
	}

	opts := []otlploghttp.Option{
		otlploghttp.WithEndpoint(cfg.Endpoint),
	}
	if cfg.Insecure {
		opts = append(opts, otlploghttp.WithInsecure())
	}

	exporter, err := otlploghttp.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP log exporter: %w", err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	lp := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)),
		sdklog.WithResource(res),
	)
	global.SetLoggerProvider(lp)

	slog.Info("OTLP logs enabled", "endpoint", cfg.Endpoint)

	return func(ctx context.Context) error {
		slog.Info("Shutting down logger provider...")
		err := lp.Shutdown(ctx)
		if err != nil {
			slog.Error("Logger provider shutdown error", "error", err)
		} else {
			slog.Info("Logger provider shutdown complete")
		}
		return err
	}, nil
}
