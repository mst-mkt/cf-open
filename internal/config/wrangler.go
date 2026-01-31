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

type WranglerConfig struct {
	Name              string         `json:"name" toml:"name"`
	AccountID         string         `json:"account_id" toml:"account_id"`
	CompatibilityDate string         `json:"compatibility_date" toml:"compatibility_date"`
	KVNamespaces      []KVNamespace  `json:"kv_namespaces" toml:"kv_namespaces"`
	D1Databases       []D1Database   `json:"d1_databases" toml:"d1_databases"`
	R2Buckets         []R2Bucket     `json:"r2_buckets" toml:"r2_buckets"`
	Vars              map[string]any `json:"vars" toml:"vars"`
}

type KVNamespace struct {
	Binding string `json:"binding" toml:"binding"`
	ID      string `json:"id" toml:"id"`
}

type D1Database struct {
	Binding      string `json:"binding" toml:"binding"`
	DatabaseName string `json:"database_name" toml:"database_name"`
	DatabaseID   string `json:"database_id" toml:"database_id"`
}

type R2Bucket struct {
	Binding    string `json:"binding" toml:"binding"`
	BucketName string `json:"bucket_name" toml:"bucket_name"`
}

func LoadWranglerConfig(configPath string) (*WranglerConfig, error) {
	if configPath == "" {
		configPath = findWranglerConfig()
		if configPath == "" {
			return nil, fmt.Errorf("wrangler config file not found")
		}
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	config := &WranglerConfig{}
	ext := strings.ToLower(filepath.Ext(configPath))

	switch ext {
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

func findWranglerConfig() string {
	candidates := []string{
		"wrangler.jsonc",
		"wrangler.json",
		"wrangler.toml",
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	return ""
}
