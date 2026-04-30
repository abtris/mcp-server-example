package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/abtris/mcp-server-example/internal/config"
	"github.com/abtris/mcp-server-example/internal/mcp_server"
	"github.com/abtris/mcp-server-example/internal/policy"
	"github.com/abtris/mcp-server-example/pkg/logger"
	"github.com/abtris/mcp-server-example/pkg/metrics"
	"github.com/abtris/mcp-server-example/pkg/tracing"
)

func main() {
	// Parse command line flags
	policyFile := flag.String("policy", "policy.rego", "Path to the OPA policy file")
	configFile := flag.String("config", "config.json", "Path to the server configuration file")
	logFormat := flag.String("log-format", "text", "Log format: text or json")
	logLevel := flag.String("log-level", "info", "Log level: debug, info, warn, error")
	flag.Parse()

	// Initialize structured logger
	log := logger.New(logger.Config{
		Format: *logFormat,
		Level:  *logLevel,
	})
	logger.SetDefault(log)

	// Initialize OpenTelemetry tracing (OTLP/HTTP)
	tracingCfg := tracing.ConfigFromEnv("mcp-server")
	tracingShutdown, err := tracing.Init(context.Background(), tracingCfg)
	if err != nil {
		slog.Error("Failed to initialize tracing", "error", err)
		os.Exit(1)
	}

	// Initialize OpenTelemetry metrics (OTLP/HTTP)
	meterShutdown, err := tracing.InitMeter(context.Background(), tracingCfg)
	if err != nil {
		slog.Error("Failed to initialize OTLP metrics", "error", err)
		os.Exit(1)
	}

	// Initialize OpenTelemetry logs (OTLP/HTTP)
	loggerShutdown, err := tracing.InitLogger(context.Background(), tracingCfg)
	if err != nil {
		slog.Error("Failed to initialize OTLP logs", "error", err)
		os.Exit(1)
	}

	// Re-create logger with OTel bridge now that LoggerProvider is ready
	log = logger.NewWithOTel(logger.Config{
		Format: *logFormat,
		Level:  *logLevel,
	}, "mcp-server")
	logger.SetDefault(log)

	// Initialize OTel metrics (must be after InitMeter so instruments use the real provider)
	m := metrics.New()
	slog.Info("Metrics initialized")

	// Load Configuration
	slog.Info("Loading configuration", "file", *configFile)
	cfg, err := config.Load(*configFile)
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Initialize Policy Engine
	slog.Info("Loading policy", "file", *policyFile)
	enforcer, err := policy.NewEnforcer(*policyFile, m)
	if err != nil {
		slog.Error("Failed to start policy engine", "error", err)
		os.Exit(1)
	}

	// Create and configure MCP Server
	mcp := mcp_server.New(cfg, enforcer, m)
	mcp.RegisterTools()

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start MCP server in background
	errChan := make(chan error, 1)
	go func() {
		slog.Info("Starting Secure MCP Server")
		errChan <- mcp.Run(ctx)
	}()

	// Wait for shutdown signal or error
	select {
	case <-sigChan:
		slog.Info("Received shutdown signal")
		cancel()

		// Graceful shutdown with timeout
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := loggerShutdown(shutdownCtx); err != nil {
			slog.Error("Error shutting down OTLP logs", "error", err)
		}

		if err := meterShutdown(shutdownCtx); err != nil {
			slog.Error("Error shutting down OTLP metrics", "error", err)
		}

		if err := tracingShutdown(shutdownCtx); err != nil {
			slog.Error("Error shutting down tracing", "error", err)
		}

		slog.Info("Shutdown complete")
	case err := <-errChan:
		if err != nil {
			slog.Error("Server error", "error", err)
			os.Exit(1)
		}
	}
}
