package config

import (
	"embed"
	"os"
	"path/filepath"
)

//go:embed config.schema.json
var embeddedSchema embed.FS

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

	return embeddedSchema.ReadFile("config.schema.json")
}
