package tracing

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// InitMeter sets up the OpenTelemetry MeterProvider exporting via OTLP/HTTP.
// Returns a shutdown function that must be called on application exit.
// If cfg is nil, metrics OTLP export is disabled.
func InitMeter(ctx context.Context, cfg *Config) (shutdown func(context.Context) error, err error) {
	if cfg == nil {
		slog.Info("OTLP metrics disabled (nil config)")
		return func(context.Context) error { return nil }, nil
	}

	opts := []otlpmetrichttp.Option{
		otlpmetrichttp.WithEndpoint(cfg.Endpoint),
	}
	if cfg.Insecure {
		opts = append(opts, otlpmetrichttp.WithInsecure())
	}

	exporter, err := otlpmetrichttp.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP metric exporter: %w", err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter)),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(mp)

	slog.Info("OTLP metrics enabled", "endpoint", cfg.Endpoint)

	return func(ctx context.Context) error {
		slog.Info("Shutting down meter provider...")
		err := mp.Shutdown(ctx)
		if err != nil {
			slog.Error("Meter provider shutdown error", "error", err)
		} else {
			slog.Info("Meter provider shutdown complete")
		}
		return err
	}, nil
}
