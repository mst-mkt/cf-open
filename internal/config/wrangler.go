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
	Vars              map[string]any `json:"vars" toml:"vars"`

	Observability       *ObservabilityConfig `json:"observability" toml:"observability"`
	Triggers            *TriggersConfig      `json:"triggers" toml:"triggers"`
	Queues              *QueuesConfig        `json:"queues" toml:"queues"`
	Workflows           []Workflow           `json:"workflows" toml:"workflows"`
	Browser             *BrowserConfig       `json:"browser" toml:"browser"`
	VPCServices         []VPCService         `json:"vpc_services" toml:"vpc_services"`
	R2Buckets           []R2Bucket           `json:"r2_buckets" toml:"r2_buckets"`
	KVNamespaces        []KVNamespace        `json:"kv_namespaces" toml:"kv_namespaces"`
	D1Databases         []D1Database         `json:"d1_databases" toml:"d1_databases"`
	Pipelines           []Pipeline           `json:"pipelines" toml:"pipelines"`
	Vectorize           []VectorizeIndex     `json:"vectorize" toml:"vectorize"`
	SecretsStoreSecrets []SecretsStoreSecret `json:"secrets_store_secrets" toml:"secrets_store_secrets"`
	Images              *ImagesConfig        `json:"images" toml:"images"`
}

type ObservabilityConfig struct {
	Enabled bool `json:"enabled" toml:"enabled"`
}

type TriggersConfig struct {
	Crons []string `json:"crons" toml:"crons"`
}

type QueuesConfig struct {
	Producers []QueueProducer `json:"producers" toml:"producers"`
}

type QueueProducer struct {
	Binding string `json:"binding" toml:"binding"`
	Queue   string `json:"queue" toml:"queue"`
}

type Workflow struct {
	Binding   string `json:"binding" toml:"binding"`
	Name      string `json:"name" toml:"name"`
	ClassName string `json:"class_name" toml:"class_name"`
}

type BrowserConfig struct {
	Binding string `json:"binding" toml:"binding"`
}

type VPCService struct {
	Binding   string `json:"binding" toml:"binding"`
	ServiceID string `json:"service_id" toml:"service_id"`
}

type R2Bucket struct {
	Binding    string `json:"binding" toml:"binding"`
	BucketName string `json:"bucket_name" toml:"bucket_name"`
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

type Pipeline struct {
	Binding  string `json:"binding" toml:"binding"`
	Pipeline string `json:"pipeline" toml:"pipeline"`
}

type VectorizeIndex struct {
	Binding   string `json:"binding" toml:"binding"`
	IndexName string `json:"index_name" toml:"index_name"`
}

type SecretsStoreSecret struct {
	Binding    string `json:"binding" toml:"binding"`
	StoreID    string `json:"store_id" toml:"store_id"`
	SecretName string `json:"secret_name" toml:"secret_name"`
}

type ImagesConfig struct {
	Binding string `json:"binding" toml:"binding"`
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
