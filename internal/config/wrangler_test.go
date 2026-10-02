package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filename string
		content  string
		validate func(t *testing.T, cfg *Config)
		wantErr  bool
	}{
		{
			name:     "有効な JSON 設定",
			filename: "wrangler.json",
			content: `{
				"name": "my-worker",
				"account_id": "abc123",
				"compatibility_date": "2024-01-01"
			}`,
			validate: func(t *testing.T, cfg *Config) {
				if cfg.Name != "my-worker" {
					t.Errorf("Name = %q, want %q", cfg.Name, "my-worker")
				}
			},
		},
		{
			name:     "コメント付き JSONC",
			filename: "wrangler.jsonc",
			content: `{
				// This is a comment
				"name": "worker-with-comment",
				"account_id": "xyz789"
			}`,
			validate: func(t *testing.T, cfg *Config) {
				if cfg.Name != "worker-with-comment" {
					t.Errorf("Name = %q, want %q", cfg.Name, "worker-with-comment")
				}
			},
		},
		{
			name:     "有効な TOML 設定",
			filename: "wrangler.toml",
			content: `
name = "toml-worker"
account_id = "abc123"
compatibility_date = "2024-01-01"
`,
			validate: func(t *testing.T, cfg *Config) {
				if cfg.Name != "toml-worker" {
					t.Errorf("Name = %q, want %q", cfg.Name, "toml-worker")
				}
			},
		},
		{
			name:     "JSON で Pages の設定",
			filename: "wrangler.json",
			content: `{
				"name": "my-project",
				"pages_build_output_dir": "./dist"
			}`,
			validate: func(t *testing.T, cfg *Config) {
				if cfg.PagesBuildOutputDir != "./dist" {
					t.Errorf("PagesBuildOutputDir = %q, want %q", cfg.PagesBuildOutputDir, "./dist")
				}
			},
		},
		{
			name:     "JSON で Observability を含む設定",
			filename: "wrangler.json",
			content: `{
				"name": "obs-worker",
				"observability": {"enabled": true}
			}`,
			validate: func(t *testing.T, cfg *Config) {
				if cfg.Observability == nil {
					t.Error("Observability is nil")
					return
				}
				if !cfg.Observability.Enabled {
					t.Error("Observability.Enabled = false, want true")
				}
			},
		},
		{
			name:     "JSON で Triggers を含む設定",
			filename: "wrangler.json",
			content: `{
				"name": "cron-worker",
				"triggers": {"crons": ["0 * * * *", "0 0 * * *"]}
			}`,
			validate: func(t *testing.T, cfg *Config) {
				if cfg.Triggers == nil {
					t.Error("Triggers is nil")
					return
				}
				if len(cfg.Triggers.Crons) != 2 {
					t.Errorf("len(Triggers.Crons) = %d, want 2", len(cfg.Triggers.Crons))
				}
			},
		},
		{
			name:     "JSON で Queues を含む設定",
			filename: "wrangler.json",
			content: `{
				"name": "queue-worker",
				"queues": {
					"producers": [
						{"binding": "MY_QUEUE", "queue": "my-queue"}
					]
				}
			}`,
			validate: func(t *testing.T, cfg *Config) {
				if cfg.Queues == nil {
					t.Error("Queues is nil")
					return
				}
				if len(cfg.Queues.Producers) != 1 {
					t.Errorf("len(Queues.Producers) = %d, want 1", len(cfg.Queues.Producers))
				}
				if cfg.Queues.Producers[0].Queue != "my-queue" {
					t.Errorf("Queues.Producers[0].Queue = %q, want %q", cfg.Queues.Producers[0].Queue, "my-queue")
				}
			},
		},
		{
			name:     "JSON で Queue consumer を含む設定",
			filename: "wrangler.json",
			content: `{
				"name": "queue-worker",
				"queues": {
					"consumers": [
						{"queue": "my-queue"}
					]
				}
			}`,
			validate: func(t *testing.T, cfg *Config) {
				if cfg.Queues == nil {
					t.Error("Queues is nil")
					return
				}
				if len(cfg.Queues.Consumers) != 1 {
					t.Errorf("len(Queues.Consumers) = %d, want 1", len(cfg.Queues.Consumers))
					return
				}
				if cfg.Queues.Consumers[0].Queue != "my-queue" {
					t.Errorf("Queues.Consumers[0].Queue = %q, want %q", cfg.Queues.Consumers[0].Queue, "my-queue")
				}
			},
		},
		{
			name:     "JSON で Workflows を含む設定",
			filename: "wrangler.json",
			content: `{
				"name": "workflow-worker",
				"workflows": [
					{"binding": "MY_WORKFLOW", "name": "my-workflow", "class_name": "MyWorkflow"}
				]
			}`,
			validate: func(t *testing.T, cfg *Config) {
				if len(cfg.Workflows) != 1 {
					t.Errorf("len(Workflows) = %d, want 1", len(cfg.Workflows))
					return
				}
				if cfg.Workflows[0].Name != "my-workflow" {
					t.Errorf("Workflows[0].Name = %q, want %q", cfg.Workflows[0].Name, "my-workflow")
				}
			},
		},
		{
			name:     "JSON で Containers を含む設定",
			filename: "wrangler.json",
			content: `{
				"name": "container-worker",
				"containers": [
					{"class_name": "MyContainer", "image": "./Dockerfile"}
				]
			}`,
			validate: func(t *testing.T, cfg *Config) {
				if len(cfg.Containers) != 1 {
					t.Errorf("len(Containers) = %d, want 1", len(cfg.Containers))
					return
				}
				if cfg.Containers[0].ClassName != "MyContainer" {
					t.Errorf("Containers[0].ClassName = %q, want %q", cfg.Containers[0].ClassName, "MyContainer")
				}
			},
		},
		{
			name:     "JSON で Durable Objects を含む設定",
			filename: "wrangler.json",
			content: `{
				"name": "do-worker",
				"durable_objects": {
					"bindings": [
						{"name": "MY_OBJECT", "class_name": "MyObject"}
					]
				}
			}`,
			validate: func(t *testing.T, cfg *Config) {
				if cfg.DurableObjects == nil {
					t.Error("DurableObjects is nil")
					return
				}
				if len(cfg.DurableObjects.Bindings) != 1 {
					t.Errorf("len(DurableObjects.Bindings) = %d, want 1", len(cfg.DurableObjects.Bindings))
					return
				}
				if cfg.DurableObjects.Bindings[0].ClassName != "MyObject" {
					t.Errorf("DurableObjects.Bindings[0].ClassName = %q, want %q", cfg.DurableObjects.Bindings[0].ClassName, "MyObject")
				}
			},
		},
		{
			name:     "JSON で Browser を含む設定",
			filename: "wrangler.json",
			content: `{
				"name": "browser-worker",
				"browser": {"binding": "MY_BROWSER"}
			}`,
			validate: func(t *testing.T, cfg *Config) {
				if cfg.Browser == nil {
					t.Error("Browser is nil")
					return
				}
				if cfg.Browser.Binding != "MY_BROWSER" {
					t.Errorf("Browser.Binding = %q, want %q", cfg.Browser.Binding, "MY_BROWSER")
				}
			},
		},
		{
			name:     "JSON で Workers AI を含む設定",
			filename: "wrangler.json",
			content: `{
				"name": "ai-worker",
				"ai": {"binding": "AI"}
			}`,
			validate: func(t *testing.T, cfg *Config) {
				if cfg.AI == nil {
					t.Error("AI is nil")
					return
				}
				if cfg.AI.Binding != "AI" {
					t.Errorf("AI.Binding = %q, want %q", cfg.AI.Binding, "AI")
				}
			},
		},
		{
			name:     "JSON で send_email を含む設定",
			filename: "wrangler.json",
			content: `{
				"name": "email-worker",
				"send_email": [
					{"name": "EMAIL"}
				]
			}`,
			validate: func(t *testing.T, cfg *Config) {
				if len(cfg.SendEmail) != 1 {
					t.Errorf("len(SendEmail) = %d, want 1", len(cfg.SendEmail))
					return
				}
				if cfg.SendEmail[0].Name != "EMAIL" {
					t.Errorf("SendEmail[0].Name = %q, want %q", cfg.SendEmail[0].Name, "EMAIL")
				}
			},
		},
		{
			name:     "JSON で Analytics Engine を含む設定",
			filename: "wrangler.json",
			content: `{
				"name": "analytics-worker",
				"analytics_engine_datasets": [
					{"binding": "EVENTS", "dataset": "my-dataset"}
				]
			}`,
			validate: func(t *testing.T, cfg *Config) {
				if len(cfg.AnalyticsEngineDatasets) != 1 {
					t.Errorf("len(AnalyticsEngineDatasets) = %d, want 1", len(cfg.AnalyticsEngineDatasets))
					return
				}
				if cfg.AnalyticsEngineDatasets[0].Dataset != "my-dataset" {
					t.Errorf("AnalyticsEngineDatasets[0].Dataset = %q, want %q", cfg.AnalyticsEngineDatasets[0].Dataset, "my-dataset")
				}
			},
		},
		{
			name:     "JSON で AI Search を含む設定",
			filename: "wrangler.json",
			content: `{
				"name": "ai-search-worker",
				"ai_search": [
					{"binding": "SEARCH", "instance_name": "my-instance"}
				],
				"ai_search_namespaces": [
					{"binding": "SEARCH_NAMESPACE", "namespace": "my-namespace"}
				]
			}`,
			validate: func(t *testing.T, cfg *Config) {
				if len(cfg.AISearch) != 1 || len(cfg.AISearchNamespaces) != 1 {
					t.Errorf("len(AISearch) = %d, len(AISearchNamespaces) = %d, want 1 and 1", len(cfg.AISearch), len(cfg.AISearchNamespaces))
					return
				}
				if cfg.AISearch[0].InstanceName != "my-instance" {
					t.Errorf("AISearch[0].InstanceName = %q, want %q", cfg.AISearch[0].InstanceName, "my-instance")
				}
				if cfg.AISearchNamespaces[0].Namespace != "my-namespace" {
					t.Errorf("AISearchNamespaces[0].Namespace = %q, want %q", cfg.AISearchNamespaces[0].Namespace, "my-namespace")
				}
			},
		},
		{
			name:     "JSON で VPC Services を含む設定",
			filename: "wrangler.json",
			content: `{
				"name": "vpc-worker",
				"vpc_services": [
					{"binding": "MY_VPC", "service_id": "vpc-service-id"}
				]
			}`,
			validate: func(t *testing.T, cfg *Config) {
				if len(cfg.VPCServices) != 1 {
					t.Errorf("len(VPCServices) = %d, want 1", len(cfg.VPCServices))
					return
				}
				if cfg.VPCServices[0].ServiceID != "vpc-service-id" {
					t.Errorf("VPCServices[0].ServiceID = %q, want %q", cfg.VPCServices[0].ServiceID, "vpc-service-id")
				}
			},
		},
		{
			name:     "JSON で R2 バケットを含む設定",
			filename: "wrangler.json",
			content: `{
				"name": "r2-worker",
				"r2_buckets": [
					{"binding": "BUCKET", "bucket_name": "my-bucket"}
				]
			}`,
			validate: func(t *testing.T, cfg *Config) {
				if len(cfg.R2Buckets) != 1 {
					t.Errorf("len(R2Buckets) = %d, want 1", len(cfg.R2Buckets))
					return
				}
				if cfg.R2Buckets[0].BucketName != "my-bucket" {
					t.Errorf("R2Buckets[0].BucketName = %q, want %q", cfg.R2Buckets[0].BucketName, "my-bucket")
				}
			},
		},
		{
			name:     "JSON で jurisdiction 付きの R2 バケットを含む設定",
			filename: "wrangler.json",
			content: `{
				"name": "r2-worker",
				"r2_buckets": [
					{"binding": "BUCKET", "bucket_name": "my-bucket", "jurisdiction": "eu"}
				]
			}`,
			validate: func(t *testing.T, cfg *Config) {
				if len(cfg.R2Buckets) != 1 {
					t.Errorf("len(R2Buckets) = %d, want 1", len(cfg.R2Buckets))
					return
				}
				if cfg.R2Buckets[0].Jurisdiction != "eu" {
					t.Errorf("R2Buckets[0].Jurisdiction = %q, want %q", cfg.R2Buckets[0].Jurisdiction, "eu")
				}
			},
		},
		{
			name:     "JSON で KV namespace を含む設定",
			filename: "wrangler.json",
			content: `{
				"name": "kv-worker",
				"kv_namespaces": [
					{"binding": "KV1", "id": "kv-id-1"},
					{"binding": "KV2", "id": "kv-id-2"}
				]
			}`,
			validate: func(t *testing.T, cfg *Config) {
				if len(cfg.KVNamespaces) != 2 {
					t.Errorf("len(KVNamespaces) = %d, want 2", len(cfg.KVNamespaces))
				}
			},
		},
		{
			name:     "JSON で D1 データベースを含む設定",
			filename: "wrangler.json",
			content: `{
				"name": "d1-worker",
				"d1_databases": [
					{"binding": "DB", "database_name": "test-db", "database_id": "db-123"}
				]
			}`,
			validate: func(t *testing.T, cfg *Config) {
				if len(cfg.D1Databases) != 1 {
					t.Errorf("len(D1Databases) = %d, want 1", len(cfg.D1Databases))
					return
				}
				if cfg.D1Databases[0].DatabaseID != "db-123" {
					t.Errorf("D1Databases[0].DatabaseID = %q, want %q", cfg.D1Databases[0].DatabaseID, "db-123")
				}
			},
		},
		{
			name:     "JSON で Hyperdrive を含む設定",
			filename: "wrangler.json",
			content: `{
				"name": "hyperdrive-worker",
				"hyperdrive": [
					{"binding": "HYPERDRIVE", "id": "hd-123"}
				]
			}`,
			validate: func(t *testing.T, cfg *Config) {
				if len(cfg.Hyperdrive) != 1 {
					t.Errorf("len(Hyperdrive) = %d, want 1", len(cfg.Hyperdrive))
					return
				}
				if cfg.Hyperdrive[0].ID != "hd-123" {
					t.Errorf("Hyperdrive[0].ID = %q, want %q", cfg.Hyperdrive[0].ID, "hd-123")
				}
			},
		},
		{
			name:     "JSON で Pipelines を含む設定",
			filename: "wrangler.json",
			content: `{
				"name": "pipeline-worker",
				"pipelines": [
					{"binding": "MY_PIPELINE", "pipeline": "my-pipeline"}
				]
			}`,
			validate: func(t *testing.T, cfg *Config) {
				if len(cfg.Pipelines) != 1 {
					t.Errorf("len(Pipelines) = %d, want 1", len(cfg.Pipelines))
					return
				}
				if cfg.Pipelines[0].Pipeline != "my-pipeline" {
					t.Errorf("Pipelines[0].Pipeline = %q, want %q", cfg.Pipelines[0].Pipeline, "my-pipeline")
				}
			},
		},
		{
			name:     "JSON で stream キーの Pipelines を含む設定",
			filename: "wrangler.json",
			content: `{
				"name": "pipeline-worker",
				"pipelines": [
					{"binding": "MY_PIPELINE", "stream": "my-stream"}
				]
			}`,
			validate: func(t *testing.T, cfg *Config) {
				if len(cfg.Pipelines) != 1 {
					t.Errorf("len(Pipelines) = %d, want 1", len(cfg.Pipelines))
					return
				}
				if cfg.Pipelines[0].Stream != "my-stream" {
					t.Errorf("Pipelines[0].Stream = %q, want %q", cfg.Pipelines[0].Stream, "my-stream")
				}
			},
		},
		{
			name:     "JSON で Vectorize を含む設定",
			filename: "wrangler.json",
			content: `{
				"name": "vectorize-worker",
				"vectorize": [
					{"binding": "MY_VECTORIZE", "index_name": "my-index"}
				]
			}`,
			validate: func(t *testing.T, cfg *Config) {
				if len(cfg.Vectorize) != 1 {
					t.Errorf("len(Vectorize) = %d, want 1", len(cfg.Vectorize))
					return
				}
				if cfg.Vectorize[0].IndexName != "my-index" {
					t.Errorf("Vectorize[0].IndexName = %q, want %q", cfg.Vectorize[0].IndexName, "my-index")
				}
			},
		},
		{
			name:     "JSON で Secrets Store を含む設定",
			filename: "wrangler.json",
			content: `{
				"name": "secrets-worker",
				"secrets_store_secrets": [
					{"binding": "MY_SECRET", "store_id": "store-123", "secret_name": "my-secret"}
				]
			}`,
			validate: func(t *testing.T, cfg *Config) {
				if len(cfg.SecretsStoreSecrets) != 1 {
					t.Errorf("len(SecretsStoreSecrets) = %d, want 1", len(cfg.SecretsStoreSecrets))
					return
				}
				if cfg.SecretsStoreSecrets[0].StoreID != "store-123" {
					t.Errorf("SecretsStoreSecrets[0].StoreID = %q, want %q", cfg.SecretsStoreSecrets[0].StoreID, "store-123")
				}
			},
		},
		{
			name:     "JSON で Images を含む設定",
			filename: "wrangler.json",
			content: `{
				"name": "images-worker",
				"images": {"binding": "MY_IMAGES"}
			}`,
			validate: func(t *testing.T, cfg *Config) {
				if cfg.Images == nil {
					t.Error("Images is nil")
					return
				}
				if cfg.Images.Binding != "MY_IMAGES" {
					t.Errorf("Images.Binding = %q, want %q", cfg.Images.Binding, "MY_IMAGES")
				}
			},
		},
		{
			name:     "TOML で KV namespace を含む設定",
			filename: "wrangler.toml",
			content: `
name = "kv-toml-worker"

[[kv_namespaces]]
binding = "KV1"
id = "kv-id-1"

[[kv_namespaces]]
binding = "KV2"
id = "kv-id-2"
`,
			validate: func(t *testing.T, cfg *Config) {
				if len(cfg.KVNamespaces) != 2 {
					t.Errorf("len(KVNamespaces) = %d, want 2", len(cfg.KVNamespaces))
				}
			},
		},
		{
			name:     "TOML で Queues を含む設定",
			filename: "wrangler.toml",
			content: `
name = "queue-toml-worker"

[queues]
[[queues.producers]]
binding = "MY_QUEUE"
queue = "my-queue"
`,
			validate: func(t *testing.T, cfg *Config) {
				if cfg.Queues == nil {
					t.Error("Queues is nil")
					return
				}
				if len(cfg.Queues.Producers) != 1 {
					t.Errorf("len(Queues.Producers) = %d, want 1", len(cfg.Queues.Producers))
				}
			},
		},
		{
			name:     "TOML で Triggers を含む設定",
			filename: "wrangler.toml",
			content: `
name = "cron-toml-worker"

[triggers]
crons = ["0 * * * *"]
`,
			validate: func(t *testing.T, cfg *Config) {
				if cfg.Triggers == nil {
					t.Error("Triggers is nil")
					return
				}
				if len(cfg.Triggers.Crons) != 1 {
					t.Errorf("len(Triggers.Crons) = %d, want 1", len(cfg.Triggers.Crons))
				}
			},
		},
		{
			name:     "無効な JSON",
			filename: "wrangler.json",
			content:  `{invalid json}`,
			wantErr:  true,
		},
		{
			name:     "無効な TOML",
			filename: "wrangler.toml",
			content:  `name = "unclosed`,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, tt.filename)

			if err := os.WriteFile(configPath, []byte(tt.content), 0o644); err != nil {
				t.Fatalf("テスト設定ファイルの書き込みに失敗: %v", err)
			}

			got, err := Load(Options{Path: configPath})
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if tt.validate != nil {
				tt.validate(t, got)
			}
		})
	}
}
