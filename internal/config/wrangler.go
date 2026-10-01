package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/tidwall/jsonc"
)

func loadWranglerConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	config := &Config{}

	switch ext := strings.ToLower(filepath.Ext(configPath)); ext {
	case ".toml":
		if err := toml.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("failed to parse TOML config file: %w", err)
		}
	case ".json", ".jsonc":
		if err := json.Unmarshal(jsonc.ToJSON(data), config); err != nil {
			return nil, fmt.Errorf("failed to parse JSON config file: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported config file format: %s", ext)
	}

	return config, nil
}
