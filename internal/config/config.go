package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/xeipuuv/gojsonschema"
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

	if err := validateConfig(data, configFile); err != nil {
		return nil, err
	}

	var cfg ServerConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

func validateConfig(data []byte, configFile string) error {
	schemaBytes, err := readSchemaBytes(configFile)
	if err != nil {
		return fmt.Errorf("failed to load config schema: %w", err)
	}

	schemaLoader := gojsonschema.NewBytesLoader(schemaBytes)
	documentLoader := gojsonschema.NewBytesLoader(data)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return fmt.Errorf("failed to validate config against schema: %w", err)
	}

	if result.Valid() {
		return nil
	}

	var messages []string
	for _, schemaErr := range result.Errors() {
		messages = append(messages, schemaErr.String())
	}

	return fmt.Errorf("config validation failed:\n- %s", strings.Join(messages, "\n- "))
}
