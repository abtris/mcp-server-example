package main

import (
	"context"
	"flag"
	"log/slog"
	"os"

	"github.com/abtris/mcp-server-2025/internal/config"
	"github.com/abtris/mcp-server-2025/internal/policy"
	"github.com/abtris/mcp-server-2025/internal/server"
	"github.com/abtris/mcp-server-2025/pkg/logger"
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

	// Load Configuration
	slog.Info("Loading configuration", "file", *configFile)
	cfg, err := config.Load(*configFile)
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Initialize Policy Engine
	slog.Info("Loading policy", "file", *policyFile)
	enforcer, err := policy.NewEnforcer(*policyFile)
	if err != nil {
		slog.Error("Failed to start policy engine", "error", err)
		os.Exit(1)
	}

	// Create and configure MCP Server
	srv := server.New(cfg, enforcer)
	srv.RegisterTools()

	// Start the Server (Stdio)
	if err := srv.Run(context.Background()); err != nil {
		slog.Error("Server error", "error", err)
		os.Exit(1)
	}
}
