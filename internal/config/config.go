package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// ServerConfig represents the MCP server configuration
type ServerConfig struct {
	Server ServerInfo `json:"server"`
	Tools  []ToolInfo `json:"tools"`
}

// ServerInfo contains server metadata
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ToolInfo contains tool configuration
type ToolInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Handler     string `json:"handler"`
}

// Load loads the configuration from a JSON file
func Load(configFile string) (*ServerConfig, error) {
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg ServerConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}
