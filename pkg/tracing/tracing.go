// Package tracing configures OpenTelemetry to export traces via OTLP/HTTP.
//
// By default traces are sent to http://localhost:4318 (the standard OTLP HTTP
// port). Override the endpoint by setting OTEL_EXPORTER_OTLP_ENDPOINT.
//
// Optional environment variable:
//
//	OTEL_EXPORTER_OTLP_ENDPOINT – OTLP collector URL (default http://localhost:4318)
package tracing

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

const defaultEndpoint = "localhost:4318"

// Config holds the OTLP tracing settings.
type Config struct {
	Endpoint    string // OTLP HTTP host:port (default localhost:4318)
	ServiceName string // OpenTelemetry service name
	Insecure    bool   // Use plaintext HTTP (default true for localhost)
}

// ConfigFromEnv reads tracing configuration from environment variables.
// Tracing is always enabled; set OTEL_EXPORTER_OTLP_ENDPOINT to override the
// default endpoint, or pass serviceName to tag spans.
func ConfigFromEnv(serviceName string) *Config {
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		endpoint = defaultEndpoint
	}

	return &Config{
		Endpoint:    endpoint,
		ServiceName: serviceName,
		Insecure:    true,
	}
}

// Init sets up the OpenTelemetry TracerProvider exporting via OTLP/HTTP.
// Returns a shutdown function that must be called on application exit.
// If cfg is nil, tracing is disabled and a no-op provider is used.
func Init(ctx context.Context, cfg *Config) (shutdown func(context.Context) error, err error) {
	if cfg == nil {
		slog.Info("Tracing disabled (nil config)")
		return func(context.Context) error { return nil }, nil
	}

	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(cfg.Endpoint),
	}
	if cfg.Insecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}

	exporter, err := otlptracehttp.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporter)),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)

	// Log OTel errors (e.g. failed exports) so they're visible.
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		slog.Error("OpenTelemetry error", "error", err)
	}))

	slog.Info("OTLP tracing enabled", "endpoint", cfg.Endpoint)

	return func(ctx context.Context) error {
		slog.Info("Flushing and shutting down trace provider...")
		err := tp.Shutdown(ctx)
		if err != nil {
			slog.Error("Trace provider shutdown error", "error", err)
		} else {
			slog.Info("Trace provider shutdown complete, all spans flushed")
		}
		return err
	}, nil
}

// Tracer returns a named tracer from the global provider.
func Tracer(name string) trace.Tracer {
	return otel.Tracer(name)
}
