package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Path string `json:"-" toml:"-"`

	Name                string         `json:"name" toml:"name"`
	AccountID           string         `json:"account_id" toml:"account_id"`
	CompatibilityDate   string         `json:"compatibility_date" toml:"compatibility_date"`
	PagesBuildOutputDir string         `json:"pages_build_output_dir" toml:"pages_build_output_dir"`
	Vars                map[string]any `json:"vars" toml:"vars"`

	Observability           *ObservabilityConfig     `json:"observability" toml:"observability"`
	Logpush                 bool                     `json:"logpush" toml:"logpush"`
	Triggers                *TriggersConfig          `json:"triggers" toml:"triggers"`
	Routes                  []Route                  `json:"routes" toml:"routes"`
	Route                   *Route                   `json:"route" toml:"route"`
	Addresses               []string                 `json:"addresses" toml:"addresses"`
	Queues                  *QueuesConfig            `json:"queues" toml:"queues"`
	Workflows               []Workflow               `json:"workflows" toml:"workflows"`
	DurableObjects          *DurableObjectsConfig    `json:"durable_objects" toml:"durable_objects"`
	Containers              []Container              `json:"containers" toml:"containers"`
	Browser                 *BrowserConfig           `json:"browser" toml:"browser"`
	AI                      *AIConfig                `json:"ai" toml:"ai"`
	Stream                  *StreamConfig            `json:"stream" toml:"stream"`
	AISearch                []AISearchInstance       `json:"ai_search" toml:"ai_search"`
	AISearchNamespaces      []AISearchNamespace      `json:"ai_search_namespaces" toml:"ai_search_namespaces"`
	Artifacts               []Artifacts              `json:"artifacts" toml:"artifacts"`
	AnalyticsEngineDatasets []AnalyticsEngineDataset `json:"analytics_engine_datasets" toml:"analytics_engine_datasets"`
	SendEmail               []SendEmail              `json:"send_email" toml:"send_email"`
	VPCServices             []VPCService             `json:"vpc_services" toml:"vpc_services"`
	VPCNetworks             []VPCNetwork             `json:"vpc_networks" toml:"vpc_networks"`
	Flagship                []Flagship               `json:"flagship" toml:"flagship"`
	Services                []Service                `json:"services" toml:"services"`
	TailConsumers           []TailConsumer           `json:"tail_consumers" toml:"tail_consumers"`
	StreamingTailConsumers  []TailConsumer           `json:"streaming_tail_consumers" toml:"streaming_tail_consumers"`
	R2Buckets               []R2Bucket               `json:"r2_buckets" toml:"r2_buckets"`
	KVNamespaces            []KVNamespace            `json:"kv_namespaces" toml:"kv_namespaces"`
	D1Databases             []D1Database             `json:"d1_databases" toml:"d1_databases"`
	Hyperdrive              []Hyperdrive             `json:"hyperdrive" toml:"hyperdrive"`
	Pipelines               []Pipeline               `json:"pipelines" toml:"pipelines"`
	Vectorize               []VectorizeIndex         `json:"vectorize" toml:"vectorize"`
	SecretsStoreSecrets     []SecretsStoreSecret     `json:"secrets_store_secrets" toml:"secrets_store_secrets"`
	Images                  *ImagesConfig            `json:"images" toml:"images"`
}

type ObservabilityConfig struct {
	Enabled bool `json:"enabled" toml:"enabled"`
}

type TriggersConfig struct {
	Crons []string `json:"crons" toml:"crons"`
}

type Route struct {
	Pattern      string `json:"pattern" toml:"pattern"`
	ZoneName     string `json:"zone_name" toml:"zone_name"`
	CustomDomain bool   `json:"custom_domain" toml:"custom_domain"`
}

func (r *Route) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &r.Pattern); err == nil {
		return nil
	}

	type route Route
	return json.Unmarshal(data, (*route)(r))
}

func (r *Route) UnmarshalTOML(data any) error {
	switch value := data.(type) {
	case string:
		r.Pattern = value
	case map[string]any:
		r.Pattern, _ = value["pattern"].(string)
		r.ZoneName, _ = value["zone_name"].(string)
		r.CustomDomain, _ = value["custom_domain"].(bool)
	}

	return nil
}

type QueuesConfig struct {
	Producers []QueueProducer `json:"producers" toml:"producers"`
	Consumers []QueueConsumer `json:"consumers" toml:"consumers"`
}

type QueueProducer struct {
	Binding string `json:"binding" toml:"binding"`
	Queue   string `json:"queue" toml:"queue"`
}

type QueueConsumer struct {
	Queue string `json:"queue" toml:"queue"`
}

type Workflow struct {
	Binding   string `json:"binding" toml:"binding"`
	Name      string `json:"name" toml:"name"`
	ClassName string `json:"class_name" toml:"class_name"`
}

type DurableObjectsConfig struct {
	Bindings []DurableObjectBinding `json:"bindings" toml:"bindings"`
}

type DurableObjectBinding struct {
	Name      string `json:"name" toml:"name"`
	ClassName string `json:"class_name" toml:"class_name"`
}

type Container struct {
	Name      string `json:"name" toml:"name"`
	ClassName string `json:"class_name" toml:"class_name"`
}

type BrowserConfig struct {
	Binding string `json:"binding" toml:"binding"`
}

type AIConfig struct {
	Binding string `json:"binding" toml:"binding"`
}

type StreamConfig struct {
	Binding string `json:"binding" toml:"binding"`
}

type AISearchInstance struct {
	Binding      string `json:"binding" toml:"binding"`
	InstanceName string `json:"instance_name" toml:"instance_name"`
}

type AISearchNamespace struct {
	Binding   string `json:"binding" toml:"binding"`
	Namespace string `json:"namespace" toml:"namespace"`
}

type Artifacts struct {
	Binding   string `json:"binding" toml:"binding"`
	Namespace string `json:"namespace" toml:"namespace"`
}

type SendEmail struct {
	Name string `json:"name" toml:"name"`
}

type AnalyticsEngineDataset struct {
	Binding string `json:"binding" toml:"binding"`
	Dataset string `json:"dataset" toml:"dataset"`
}

type VPCService struct {
	Binding   string `json:"binding" toml:"binding"`
	ServiceID string `json:"service_id" toml:"service_id"`
}

type VPCNetwork struct {
	Binding   string `json:"binding" toml:"binding"`
	TunnelID  string `json:"tunnel_id" toml:"tunnel_id"`
	NetworkID string `json:"network_id" toml:"network_id"`
}

type Flagship struct {
	Binding string `json:"binding" toml:"binding"`
	AppID   string `json:"app_id" toml:"app_id"`
}

type Service struct {
	Binding string `json:"binding" toml:"binding"`
	Service string `json:"service" toml:"service"`
}

type TailConsumer struct {
	Service string `json:"service" toml:"service"`
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

type Hyperdrive struct {
	Binding string `json:"binding" toml:"binding"`
	ID      string `json:"id" toml:"id"`
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
