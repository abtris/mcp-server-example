package config

import (
	"os"
	"path/filepath"
)

const defaultSchema = `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "Secure MCP Server Config",
  "type": "object",
  "additionalProperties": false,
  "required": ["server", "tools"],
  "properties": {
    "server": {
      "type": "object",
      "additionalProperties": false,
      "required": ["name", "version"],
      "properties": {
        "name": {
          "type": "string",
          "minLength": 1
        },
        "version": {
          "type": "string",
          "minLength": 1
        }
      }
    },
    "tools": {
      "type": "array",
      "minItems": 1,
      "items": {
        "type": "object",
        "additionalProperties": false,
        "required": ["name", "description", "handler"],
        "properties": {
          "name": {
            "type": "string",
            "minLength": 1
          },
          "description": {
            "type": "string",
            "minLength": 1
          },
          "handler": {
            "type": "string",
            "minLength": 1
          }
        }
      }
    }
  }
}`

func readSchemaBytes(configFile string) ([]byte, error) {
	candidates := []string{
		filepath.Join(filepath.Dir(configFile), "config.schema.json"),
		"config.schema.json",
	}

	for _, candidate := range candidates {
		if data, err := os.ReadFile(candidate); err == nil {
			return data, nil
		}
	}

	return []byte(defaultSchema), nil
}
