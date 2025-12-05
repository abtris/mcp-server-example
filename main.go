package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/abtris/mcp-server-2025/internal/config"
	"github.com/abtris/mcp-server-2025/internal/mcp_server"
	"github.com/abtris/mcp-server-2025/internal/policy"
	"github.com/abtris/mcp-server-2025/pkg/logger"
	"github.com/abtris/mcp-server-2025/pkg/metrics"
)

func main() {
	// Parse command line flags
	policyFile := flag.String("policy", "policy.rego", "Path to the OPA policy file")
	configFile := flag.String("config", "config.json", "Path to the server configuration file")
	logFormat := flag.String("log-format", "text", "Log format: text or json")
	logLevel := flag.String("log-level", "info", "Log level: debug, info, warn, error")
	metricsPort := flag.Int("metrics-port", 9090, "Port for Prometheus metrics endpoint")
	flag.Parse()

	// Initialize structured logger
	log := logger.New(logger.Config{
		Format: *logFormat,
		Level:  *logLevel,
	})
	logger.SetDefault(log)

	// Initialize Prometheus metrics
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

	// Start metrics HTTP server in background
	metricsServer := metrics.NewServer(*metricsPort)
	go func() {
		if err := metricsServer.Start(); err != nil {
			slog.Error("Metrics server error", "error", err)
		}
	}()

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

		if err := metricsServer.Shutdown(shutdownCtx); err != nil {
			slog.Error("Error shutting down metrics server", "error", err)
		}

		slog.Info("Shutdown complete")
	case err := <-errChan:
		if err != nil {
			slog.Error("Server error", "error", err)
			os.Exit(1)
		}
	}
}
