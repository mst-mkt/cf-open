package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/tidwall/jsonc"
)

type WranglerConfig struct {
	Name              string                 `json:"name"`
	AccountID         string                 `json:"account_id"`
	CompatibilityDate string                 `json:"compatibility_date"`
	KVNamespaces      []KVNamespace          `json:"kv_namespaces"`
	D1Databases       []D1Database           `json:"d1_databases"`
	R2Buckets         []R2Bucket             `json:"r2_buckets"`
	Vars              map[string]interface{} `json:"vars"`
}

type KVNamespace struct {
	Binding string `json:"binding"`
	ID      string `json:"id"`
}

type D1Database struct {
	Binding      string `json:"binding"`
	DatabaseName string `json:"database_name"`
	DatabaseID   string `json:"database_id"`
}

type R2Bucket struct {
	Binding    string `json:"binding"`
	BucketName string `json:"bucket_name"`
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
	if err := json.Unmarshal(jsonc.ToJSON(data), config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}

func findWranglerConfig() string {
	candidates := []string{
		"wrangler.jsonc",
		"wrangler.json",
		// "wrangler.toml",
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	return ""
}
