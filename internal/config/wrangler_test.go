package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadWranglerConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filename string
		content  string
		validate func(t *testing.T, cfg *WranglerConfig)
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
			validate: func(t *testing.T, cfg *WranglerConfig) {
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
			validate: func(t *testing.T, cfg *WranglerConfig) {
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
			validate: func(t *testing.T, cfg *WranglerConfig) {
				if cfg.Name != "toml-worker" {
					t.Errorf("Name = %q, want %q", cfg.Name, "toml-worker")
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
			validate: func(t *testing.T, cfg *WranglerConfig) {
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
			validate: func(t *testing.T, cfg *WranglerConfig) {
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
			validate: func(t *testing.T, cfg *WranglerConfig) {
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
			name:     "JSON で Workflows を含む設定",
			filename: "wrangler.json",
			content: `{
				"name": "workflow-worker",
				"workflows": [
					{"binding": "MY_WORKFLOW", "name": "my-workflow", "class_name": "MyWorkflow"}
				]
			}`,
			validate: func(t *testing.T, cfg *WranglerConfig) {
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
			name:     "JSON で Browser を含む設定",
			filename: "wrangler.json",
			content: `{
				"name": "browser-worker",
				"browser": {"binding": "MY_BROWSER"}
			}`,
			validate: func(t *testing.T, cfg *WranglerConfig) {
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
			name:     "JSON で VPC Services を含む設定",
			filename: "wrangler.json",
			content: `{
				"name": "vpc-worker",
				"vpc_services": [
					{"binding": "MY_VPC", "service_id": "vpc-service-id"}
				]
			}`,
			validate: func(t *testing.T, cfg *WranglerConfig) {
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
			validate: func(t *testing.T, cfg *WranglerConfig) {
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
			name:     "JSON で KV namespace を含む設定",
			filename: "wrangler.json",
			content: `{
				"name": "kv-worker",
				"kv_namespaces": [
					{"binding": "KV1", "id": "kv-id-1"},
					{"binding": "KV2", "id": "kv-id-2"}
				]
			}`,
			validate: func(t *testing.T, cfg *WranglerConfig) {
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
			validate: func(t *testing.T, cfg *WranglerConfig) {
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
			name:     "JSON で Pipelines を含む設定",
			filename: "wrangler.json",
			content: `{
				"name": "pipeline-worker",
				"pipelines": [
					{"binding": "MY_PIPELINE", "pipeline": "my-pipeline"}
				]
			}`,
			validate: func(t *testing.T, cfg *WranglerConfig) {
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
			name:     "JSON で Vectorize を含む設定",
			filename: "wrangler.json",
			content: `{
				"name": "vectorize-worker",
				"vectorize": [
					{"binding": "MY_VECTORIZE", "index_name": "my-index"}
				]
			}`,
			validate: func(t *testing.T, cfg *WranglerConfig) {
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
			validate: func(t *testing.T, cfg *WranglerConfig) {
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
			validate: func(t *testing.T, cfg *WranglerConfig) {
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
			validate: func(t *testing.T, cfg *WranglerConfig) {
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
			validate: func(t *testing.T, cfg *WranglerConfig) {
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
			validate: func(t *testing.T, cfg *WranglerConfig) {
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

			if err := os.WriteFile(configPath, []byte(tt.content), 0644); err != nil {
				t.Fatalf("テスト設定ファイルの書き込みに失敗: %v", err)
			}

			got, err := LoadWranglerConfig(configPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadWranglerConfig() error = %v, wantErr %v", err, tt.wantErr)
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

func TestLoadWranglerConfig_FileNotFound(t *testing.T) {
	t.Parallel()

	_, err := LoadWranglerConfig("/nonexistent/path/wrangler.json")
	if err == nil {
		t.Error("LoadWranglerConfig() expected error for nonexistent file, got nil")
	}
}

func TestLoadWranglerConfig_EmptyPath(t *testing.T) {
	t.Parallel()

	_, err := LoadWranglerConfig("")
	if err == nil {
		t.Error("LoadWranglerConfig() expected error for empty path with no wrangler config, got nil")
	}
}
