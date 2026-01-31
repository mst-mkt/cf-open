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
			name:       "Account ID あり",
			accountID:  "abc123",
			path:       "workers/services/view/my-worker/production",
			hasAccount: true,
			want:       "https://dash.cloudflare.com/abc123/workers/services/view/my-worker/production",
		},
		{
			name:       "Account ID なし",
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
		name      string
		config    *config.WranglerConfig
		wantTypes []ResourceType
		wantURLs  map[ResourceType]string
	}{
		{
			name:      "空の設定",
			config:    &config.WranglerConfig{},
			wantTypes: nil,
			wantURLs:  nil,
		},
		{
			name: "Worker のみ",
			config: &config.WranglerConfig{
				Name: "my-worker",
			},
			wantTypes: []ResourceType{ResourceTypeWorker},
			wantURLs: map[ResourceType]string{
				ResourceTypeWorker: "https://dash.cloudflare.com/acc/workers/services/view/my-worker/production",
			},
		},
		{
			name: "Worker + Observability",
			config: &config.WranglerConfig{
				Name:          "my-worker",
				Observability: &config.ObservabilityConfig{Enabled: true},
			},
			wantTypes: []ResourceType{ResourceTypeWorker, ResourceTypeObservability},
			wantURLs: map[ResourceType]string{
				ResourceTypeWorker:        "https://dash.cloudflare.com/acc/workers/services/view/my-worker/production",
				ResourceTypeObservability: "https://dash.cloudflare.com/acc/workers/services/view/my-worker/production/observability",
			},
		},
		{
			name: "KV Namespace",
			config: &config.WranglerConfig{
				KVNamespaces: []config.KVNamespace{
					{Binding: "MY_KV", ID: "kv-id-123"},
				},
			},
			wantTypes: []ResourceType{ResourceTypeKV},
			wantURLs: map[ResourceType]string{
				ResourceTypeKV: "https://dash.cloudflare.com/acc/workers/kv/namespaces/kv-id-123/metrics",
			},
		},
		{
			name: "D1 Database",
			config: &config.WranglerConfig{
				D1Databases: []config.D1Database{
					{Binding: "MY_DB", DatabaseName: "my-db", DatabaseID: "d1-id-456"},
				},
			},
			wantTypes: []ResourceType{ResourceTypeD1},
			wantURLs: map[ResourceType]string{
				ResourceTypeD1: "https://dash.cloudflare.com/acc/workers/d1/databases/d1-id-456/metrics",
			},
		},
		{
			name: "R2 Bucket",
			config: &config.WranglerConfig{
				R2Buckets: []config.R2Bucket{
					{Binding: "MY_BUCKET", BucketName: "my-bucket"},
				},
			},
			wantTypes: []ResourceType{ResourceTypeR2},
			wantURLs: map[ResourceType]string{
				ResourceTypeR2: "https://dash.cloudflare.com/acc/r2/default/buckets/my-bucket",
			},
		},
		{
			name: "Queue",
			config: &config.WranglerConfig{
				Queues: &config.QueuesConfig{
					Producers: []config.QueueProducer{
						{Binding: "MY_QUEUE", Queue: "my-queue"},
					},
				},
			},
			wantTypes: []ResourceType{ResourceTypeQueue},
			wantURLs: map[ResourceType]string{
				ResourceTypeQueue: "https://dash.cloudflare.com/acc/workers/queues/my-queue/metrics",
			},
		},
		{
			name: "Workflow",
			config: &config.WranglerConfig{
				Workflows: []config.Workflow{
					{Binding: "MY_WORKFLOW", Name: "my-workflow", ClassName: "MyWorkflow"},
				},
			},
			wantTypes: []ResourceType{ResourceTypeWorkflow},
			wantURLs: map[ResourceType]string{
				ResourceTypeWorkflow: "https://dash.cloudflare.com/acc/workers/workflows/my-workflow/instances",
			},
		},
		{
			name: "Vectorize",
			config: &config.WranglerConfig{
				Vectorize: []config.VectorizeIndex{
					{Binding: "MY_VECTORIZE", IndexName: "my-index"},
				},
			},
			wantTypes: []ResourceType{ResourceTypeVectorize},
			wantURLs: map[ResourceType]string{
				ResourceTypeVectorize: "https://dash.cloudflare.com/acc/ai/vectorize/my-index",
			},
		},
		{
			name: "Pipeline",
			config: &config.WranglerConfig{
				Pipelines: []config.Pipeline{
					{Binding: "MY_PIPELINE", Pipeline: "my-pipeline"},
				},
			},
			wantTypes: []ResourceType{ResourceTypePipeline},
			wantURLs: map[ResourceType]string{
				ResourceTypePipeline: "https://dash.cloudflare.com/acc/pipelines/my-pipeline/overview",
			},
		},
		{
			name: "Secrets Store",
			config: &config.WranglerConfig{
				SecretsStoreSecrets: []config.SecretsStoreSecret{
					{Binding: "MY_SECRET", StoreID: "store-id", SecretName: "my-secret"},
				},
			},
			wantTypes: []ResourceType{ResourceTypeSecretsStore},
			wantURLs: map[ResourceType]string{
				ResourceTypeSecretsStore: "https://dash.cloudflare.com/acc/secrets-store/store-id",
			},
		},
		{
			name: "Secrets Store - 同じ Store ID をまとめる",
			config: &config.WranglerConfig{
				SecretsStoreSecrets: []config.SecretsStoreSecret{
					{Binding: "SECRET1", StoreID: "store-id", SecretName: "secret-1"},
					{Binding: "SECRET2", StoreID: "store-id", SecretName: "secret-2"},
				},
			},
			wantTypes: []ResourceType{ResourceTypeSecretsStore},
			wantURLs: map[ResourceType]string{
				ResourceTypeSecretsStore: "https://dash.cloudflare.com/acc/secrets-store/store-id",
			},
		},
		{
			name: "Browser Rendering",
			config: &config.WranglerConfig{
				Browser: &config.BrowserConfig{Binding: "MY_BROWSER"},
			},
			wantTypes: []ResourceType{ResourceTypeBrowserRendering},
			wantURLs: map[ResourceType]string{
				ResourceTypeBrowserRendering: "https://dash.cloudflare.com/acc/workers/browser-rendering/overview",
			},
		},
		{
			name: "Images",
			config: &config.WranglerConfig{
				Images: &config.ImagesConfig{Binding: "MY_IMAGES"},
			},
			wantTypes: []ResourceType{ResourceTypeImages},
			wantURLs: map[ResourceType]string{
				ResourceTypeImages: "https://dash.cloudflare.com/acc/images",
			},
		},
		{
			name: "VPC Services",
			config: &config.WranglerConfig{
				VPCServices: []config.VPCService{
					{Binding: "MY_VPC", ServiceID: "vpc-id"},
				},
			},
			wantTypes: []ResourceType{ResourceTypeVPC},
			wantURLs: map[ResourceType]string{
				ResourceTypeVPC: "https://dash.cloudflare.com/acc/workers/vpc/services",
			},
		},
		{
			name: "Cron Triggers",
			config: &config.WranglerConfig{
				Name: "my-worker",
				Triggers: &config.TriggersConfig{
					Crons: []string{"0 * * * *"},
				},
			},
			wantTypes: []ResourceType{ResourceTypeWorker, ResourceTypeCronTriggers},
			wantURLs: map[ResourceType]string{
				ResourceTypeWorker:       "https://dash.cloudflare.com/acc/workers/services/view/my-worker/production",
				ResourceTypeCronTriggers: "https://dash.cloudflare.com/acc/workers/services/view/my-worker/production/settings#trigger-events",
			},
		},
		{
			name: "Cron Triggers - Worker 名なしの場合は表示しない",
			config: &config.WranglerConfig{
				Triggers: &config.TriggersConfig{
					Crons: []string{"0 * * * *"},
				},
			},
			wantTypes: nil,
			wantURLs:  nil,
		},
		{
			name: "全リソース",
			config: &config.WranglerConfig{
				Name:          "my-worker",
				Observability: &config.ObservabilityConfig{Enabled: true},
				KVNamespaces:  []config.KVNamespace{{Binding: "KV", ID: "kv-id"}},
				D1Databases:   []config.D1Database{{Binding: "DB", DatabaseName: "db", DatabaseID: "d1-id"}},
				R2Buckets:     []config.R2Bucket{{Binding: "R2", BucketName: "bucket"}},
				Queues:        &config.QueuesConfig{Producers: []config.QueueProducer{{Binding: "Q", Queue: "queue"}}},
				Workflows:     []config.Workflow{{Binding: "WF", Name: "workflow", ClassName: "WF"}},
				Vectorize:     []config.VectorizeIndex{{Binding: "VEC", IndexName: "index"}},
				Pipelines:     []config.Pipeline{{Binding: "PIPE", Pipeline: "pipeline"}},
				SecretsStoreSecrets: []config.SecretsStoreSecret{
					{Binding: "SEC", StoreID: "store", SecretName: "secret"},
				},
				Browser:     &config.BrowserConfig{Binding: "BROWSER"},
				Images:      &config.ImagesConfig{Binding: "IMAGES"},
				VPCServices: []config.VPCService{{Binding: "VPC", ServiceID: "vpc"}},
				Triggers:    &config.TriggersConfig{Crons: []string{"* * * * *"}},
			},
			wantTypes: []ResourceType{
				ResourceTypeWorker,
				ResourceTypeObservability,
				ResourceTypeCronTriggers,
				ResourceTypeQueue,
				ResourceTypeWorkflow,
				ResourceTypeBrowserRendering,
				ResourceTypeVPC,
				ResourceTypeR2,
				ResourceTypeKV,
				ResourceTypeD1,
				ResourceTypePipeline,
				ResourceTypeVectorize,
				ResourceTypeSecretsStore,
				ResourceTypeImages,
			},
			wantURLs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			resources := GetResourcesFromConfig(tt.config, "acc", true)

			if len(resources) != len(tt.wantTypes) {
				t.Errorf("リソース数 = %d, want %d", len(resources), len(tt.wantTypes))
				return
			}

			for i, wantType := range tt.wantTypes {
				if resources[i].Type != wantType {
					t.Errorf("resources[%d].Type = %q, want %q", i, resources[i].Type, wantType)
				}
			}

			for resType, wantURL := range tt.wantURLs {
				for _, r := range resources {
					if r.Type == resType {
						if r.URL != wantURL {
							t.Errorf("%s の URL = %q, want %q", resType, r.URL, wantURL)
						}
						break
					}
				}
			}
		})
	}
}

func TestGetResourcesFromConfig_NoAccountID(t *testing.T) {
	t.Parallel()

	cfg := &config.WranglerConfig{
		Name: "my-worker",
		KVNamespaces: []config.KVNamespace{
			{Binding: "KV", ID: "kv-id"},
		},
	}

	resources := GetResourcesFromConfig(cfg, "", false)

	expectedURLs := map[ResourceType]string{
		ResourceTypeWorker: "https://dash.cloudflare.com/?to=/:account/workers/services/view/my-worker/production",
		ResourceTypeKV:     "https://dash.cloudflare.com/?to=/:account/workers/kv/namespaces/kv-id/metrics",
	}

	for _, r := range resources {
		if expected, ok := expectedURLs[r.Type]; ok {
			if r.URL != expected {
				t.Errorf("%s の URL = %q, want %q", r.Type, r.URL, expected)
			}
		}
	}
}
