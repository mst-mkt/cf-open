package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadWranglerConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		filename    string
		content     string
		wantName    string
		wantKVCount int
		wantD1Count int
		wantR2Count int
		wantErr     bool
	}{
		{
			name:     "有効な JSON 設定",
			filename: "wrangler.json",
			content: `{
				"name": "my-worker",
				"account_id": "abc123",
				"compatibility_date": "2024-01-01"
			}`,
			wantName: "my-worker",
			wantErr:  false,
		},
		{
			name:     "コメント付き JSONC",
			filename: "wrangler.jsonc",
			content: `{
				// This is a comment
				"name": "worker-with-comment",
				"account_id": "xyz789"
			}`,
			wantName: "worker-with-comment",
			wantErr:  false,
		},
		{
			name:     "有効な TOML 設定",
			filename: "wrangler.toml",
			content: `
name = "toml-worker"
account_id = "abc123"
compatibility_date = "2024-01-01"
`,
			wantName: "toml-worker",
			wantErr:  false,
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
			wantName:    "kv-toml-worker",
			wantKVCount: 2,
			wantErr:     false,
		},
		{
			name:     "TOML で D1 データベースを含む設定",
			filename: "wrangler.toml",
			content: `
name = "d1-toml-worker"

[[d1_databases]]
binding = "DB"
database_name = "test-db"
database_id = "db-123"
`,
			wantName:    "d1-toml-worker",
			wantD1Count: 1,
			wantErr:     false,
		},
		{
			name:     "TOML で R2 バケットを含む設定",
			filename: "wrangler.toml",
			content: `
name = "r2-toml-worker"

[[r2_buckets]]
binding = "BUCKET"
bucket_name = "my-bucket"
`,
			wantName:    "r2-toml-worker",
			wantR2Count: 1,
			wantErr:     false,
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
			wantName:    "kv-worker",
			wantKVCount: 2,
			wantErr:     false,
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
			wantName:    "d1-worker",
			wantD1Count: 1,
			wantErr:     false,
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
			wantName:    "r2-worker",
			wantR2Count: 1,
			wantErr:     false,
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

			if got.Name != tt.wantName {
				t.Errorf("config.Name = %q, want %q", got.Name, tt.wantName)
			}
			if len(got.KVNamespaces) != tt.wantKVCount {
				t.Errorf("len(config.KVNamespaces) = %d, want %d", len(got.KVNamespaces), tt.wantKVCount)
			}
			if len(got.D1Databases) != tt.wantD1Count {
				t.Errorf("len(config.D1Databases) = %d, want %d", len(got.D1Databases), tt.wantD1Count)
			}
			if len(got.R2Buckets) != tt.wantR2Count {
				t.Errorf("len(config.R2Buckets) = %d, want %d", len(got.R2Buckets), tt.wantR2Count)
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
