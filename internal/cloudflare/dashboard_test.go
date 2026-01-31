package cloudflare

import (
	"testing"

	"github.com/mst-mkt/cf-open/internal/config"
)

func TestBuildDashboardURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		accountID  string
		path       string
		hasAccount bool
		want       string
	}{
		{
			name:       "Account ID があれば直接指定した URL を返す",
			accountID:  "abc123",
			path:       "workers/services/view/my-worker/production",
			hasAccount: true,
			want:       "https://dash.cloudflare.com/abc123/workers/services/view/my-worker/production",
		},
		{
			name:       "Account ID がなければ ?to=/:account を使った URL を返す",
			accountID:  "",
			path:       "workers/services/view/my-worker/production",
			hasAccount: false,
			want:       "https://dash.cloudflare.com/?to=/:account/workers/services/view/my-worker/production",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := BuildDashboardURL(tt.accountID, tt.path, tt.hasAccount)
			if got != tt.want {
				t.Errorf("BuildDashboardURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetResourcesFromConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		config     *config.WranglerConfig
		accountID  string
		hasAccount bool
		wantLen    int
		wantTypes  []ResourceType
	}{
		{
			name: "設定が空の場合",
			config: &config.WranglerConfig{
				Name: "",
			},
			accountID:  "",
			hasAccount: false,
			wantLen:    0,
			wantTypes:  nil,
		},
		{
			name: "設定が Worker のみ",
			config: &config.WranglerConfig{
				Name: "my-worker",
			},
			accountID:  "abc123",
			hasAccount: true,
			wantLen:    1,
			wantTypes:  []ResourceType{ResourceTypeWorker},
		},
		{
			name: "設定が全てのリソースを含む",
			config: &config.WranglerConfig{
				Name: "my-worker",
				KVNamespaces: []config.KVNamespace{
					{Binding: "KV_BINDING", ID: "kv-id-123"},
				},
				D1Databases: []config.D1Database{
					{Binding: "DB", DatabaseName: "my-db", DatabaseID: "d1-id-456"},
				},
				R2Buckets: []config.R2Bucket{
					{Binding: "BUCKET", BucketName: "my-bucket"},
				},
			},
			accountID:  "abc123",
			hasAccount: true,
			wantLen:    4,
			wantTypes:  []ResourceType{ResourceTypeWorker, ResourceTypeKV, ResourceTypeD1, ResourceTypeR2},
		},
		{
			name: "設定が複数のKV namespaceを含む",
			config: &config.WranglerConfig{
				KVNamespaces: []config.KVNamespace{
					{Binding: "KV1", ID: "kv-1"},
					{Binding: "KV2", ID: "kv-2"},
				},
			},
			accountID:  "abc123",
			hasAccount: true,
			wantLen:    2,
			wantTypes:  []ResourceType{ResourceTypeKV, ResourceTypeKV},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := GetResourcesFromConfig(tt.config, tt.accountID, tt.hasAccount)

			if len(got) != tt.wantLen {
				t.Errorf("GetResourcesFromConfig() returned %d resources, want %d", len(got), tt.wantLen)
			}

			for i, wantType := range tt.wantTypes {
				if i >= len(got) {
					break
				}
				if got[i].Type != wantType {
					t.Errorf("resource[%d].Type = %q, want %q", i, got[i].Type, wantType)
				}
			}
		})
	}
}

func TestGetResourcesFromConfig_URLs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		config     *config.WranglerConfig
		accountID  string
		hasAccount bool
		wantURLs   map[string]string
	}{
		{
			name: "Account ID ありで全リソースの URL を生成",
			config: &config.WranglerConfig{
				Name: "my-worker",
				KVNamespaces: []config.KVNamespace{
					{Binding: "KV1", ID: "kv-namespace-id-1"},
					{Binding: "KV2", ID: "kv-namespace-id-2"},
				},
				D1Databases: []config.D1Database{
					{Binding: "DB", DatabaseName: "test-db", DatabaseID: "d1-db-id"},
				},
				R2Buckets: []config.R2Bucket{
					{Binding: "BUCKET", BucketName: "test-bucket"},
				},
			},
			accountID:  "account123",
			hasAccount: true,
			wantURLs: map[string]string{
				"my-worker":         "https://dash.cloudflare.com/account123/workers/services/view/my-worker/production",
				"kv-namespace-id-1": "https://dash.cloudflare.com/account123/workers/kv/namespaces/kv-namespace-id-1",
				"kv-namespace-id-2": "https://dash.cloudflare.com/account123/workers/kv/namespaces/kv-namespace-id-2",
				"d1-db-id":          "https://dash.cloudflare.com/account123/workers/d1/databases/d1-db-id",
				"test-bucket":       "https://dash.cloudflare.com/account123/r2/default/buckets/test-bucket",
			},
		},
		{
			name: "Account ID なしで ?to=/:account を使った URL を生成",
			config: &config.WranglerConfig{
				Name: "my-worker",
				KVNamespaces: []config.KVNamespace{
					{Binding: "KV", ID: "kv-id"},
				},
			},
			accountID:  "",
			hasAccount: false,
			wantURLs: map[string]string{
				"my-worker": "https://dash.cloudflare.com/?to=/:account/workers/services/view/my-worker/production",
				"kv-id":     "https://dash.cloudflare.com/?to=/:account/workers/kv/namespaces/kv-id",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			resources := GetResourcesFromConfig(tt.config, tt.accountID, tt.hasAccount)

			if len(resources) != len(tt.wantURLs) {
				t.Errorf("GetResourcesFromConfig() returned %d resources, want %d", len(resources), len(tt.wantURLs))
			}

			for _, r := range resources {
				expected, ok := tt.wantURLs[r.ID]
				if !ok {
					t.Errorf("予期しないリソース ID: %q", r.ID)
					continue
				}
				if r.URL != expected {
					t.Errorf("resource %s (ID: %s) URL = %q, want %q", r.Type, r.ID, r.URL, expected)
				}
			}
		})
	}
}
