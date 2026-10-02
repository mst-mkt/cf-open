package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Path string `json:"-" toml:"-"`

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
	Binding      string `json:"binding" toml:"binding"`
	BucketName   string `json:"bucket_name" toml:"bucket_name"`
	Jurisdiction string `json:"jurisdiction" toml:"jurisdiction"`
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
	Stream   string `json:"stream" toml:"stream"`
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

type Options struct {
	Path string
	Mode string
}

func Load(opts Options) (*Config, error) {
	configPath := opts.Path
	if configPath == "" {
		wd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get the working directory: %w", err)
		}

		configPath = findConfig(wd)
		if configPath == "" {
			return nil, errors.New("no config file found (cloudflare.config.ts, wrangler.jsonc, wrangler.json or wrangler.toml)")
		}
	}

	if strings.EqualFold(filepath.Base(configPath), "wrangler.config.ts") {
		return nil, fmt.Errorf("%s holds only Wrangler tooling settings; pass cloudflare.config.ts instead", configPath)
	}

	config, err := loadConfigFile(configPath, opts.Mode)
	if err != nil {
		return nil, err
	}

	config.Path = configPath

	return config, nil
}

func loadConfigFile(configPath, mode string) (*Config, error) {
	if strings.ToLower(filepath.Ext(configPath)) == ".ts" {
		if _, err := os.Stat(configPath); err != nil {
			return nil, fmt.Errorf("failed to find config file: %w", err)
		}

		return loadTypeScriptConfig(configPath, mode)
	}

	return loadWranglerConfig(configPath)
}

func findConfig(dir string) string {
	candidates := []string{
		"cloudflare.config.ts",
		"wrangler.jsonc",
		"wrangler.json",
		"wrangler.toml",
	}

	for {
		for _, candidate := range candidates {
			path := filepath.Join(dir, candidate)
			if info, err := os.Stat(path); err == nil && !info.IsDir() {
				return path
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}

		dir = parent
	}
}
