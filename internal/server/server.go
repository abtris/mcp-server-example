package server

import (
	"context"
	"log/slog"

	"github.com/abtris/mcp-server-2025/internal/config"
	"github.com/abtris/mcp-server-2025/internal/policy"
	"github.com/abtris/mcp-server-2025/internal/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Server wraps the MCP server with configuration
type Server struct {
	config   *config.ServerConfig
	enforcer *policy.Enforcer
	mcp      *mcp.Server
}

// New creates a new MCP server with the given configuration and policy enforcer
func New(cfg *config.ServerConfig, enforcer *policy.Enforcer) *Server {
	mcpServer := mcp.NewServer(&mcp.Implementation{
		Name:    cfg.Server.Name,
		Version: cfg.Server.Version,
	}, nil)

	slog.Info("Creating MCP server", "name", cfg.Server.Name, "version", cfg.Server.Version)

	return &Server{
		config:   cfg,
		enforcer: enforcer,
		mcp:      mcpServer,
	}
}

// RegisterTools registers all tools from the configuration
func (s *Server) RegisterTools() {
	for _, tool := range s.config.Tools {
		slog.Info("Registering tool", "name", tool.Name, "handler", tool.Handler)

		switch tool.Handler {
		case "http_get":
			mcp.AddTool(s.mcp, &mcp.Tool{
				Name:        tool.Name,
				Description: tool.Description,
			}, policy.Enforce(s.enforcer, tool.Name, tools.SafeGetHandler))
		case "echo":
			mcp.AddTool(s.mcp, &mcp.Tool{
				Name:        tool.Name,
				Description: tool.Description,
			}, policy.Enforce(s.enforcer, tool.Name, tools.EchoHandler))
		default:
			slog.Warn("Unknown handler, skipping tool", "handler", tool.Handler, "tool", tool.Name)
		}
	}
}

// Run starts the MCP server with stdio transport
func (s *Server) Run(ctx context.Context) error {
	slog.Info("Starting Secure MCP Server")
	return s.mcp.Run(ctx, &mcp.StdioTransport{})
}
