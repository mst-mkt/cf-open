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
		config    *config.Config
		wantTypes []ResourceType
		wantURLs  map[ResourceType]string
	}{
		{
			name:      "空の設定",
			config:    &config.Config{},
			wantTypes: nil,
			wantURLs:  nil,
		},
		{
			name: "Worker のみ",
			config: &config.Config{
				Name: "my-worker",
			},
			wantTypes: []ResourceType{ResourceTypeWorker},
			wantURLs: map[ResourceType]string{
				ResourceTypeWorker: "https://dash.cloudflare.com/acc/workers/services/view/my-worker/production",
			},
		},
		{
			name: "Pages",
			config: &config.Config{
				Name:                "my-project",
				PagesBuildOutputDir: "./dist",
			},
			wantTypes: []ResourceType{ResourceTypePages},
			wantURLs: map[ResourceType]string{
				ResourceTypePages: "https://dash.cloudflare.com/acc/pages/view/my-project",
			},
		},
		{
			name: "Pages - Observability と Cron Triggers は表示しない",
			config: &config.Config{
				Name:                "my-project",
				PagesBuildOutputDir: "./dist",
				Observability:       &config.ObservabilityConfig{Enabled: true},
				Triggers:            &config.TriggersConfig{Crons: []string{"0 * * * *"}},
			},
			wantTypes: []ResourceType{ResourceTypePages},
			wantURLs:  nil,
		},
		{
			name: "Worker + Observability",
			config: &config.Config{
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
			config: &config.Config{
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
			config: &config.Config{
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
			name: "Hyperdrive",
			config: &config.Config{
				Hyperdrive: []config.Hyperdrive{
					{Binding: "MY_HYPERDRIVE", ID: "hd-id"},
				},
			},
			wantTypes: []ResourceType{ResourceTypeHyperdrive},
			wantURLs: map[ResourceType]string{
				ResourceTypeHyperdrive: "https://dash.cloudflare.com/acc/workers/hyperdrive/hd-id",
			},
		},
		{
			name: "R2 Bucket",
			config: &config.Config{
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
			name: "R2 Bucket - jurisdiction 付き",
			config: &config.Config{
				R2Buckets: []config.R2Bucket{
					{Binding: "MY_BUCKET", BucketName: "my-bucket", Jurisdiction: "eu"},
				},
			},
			wantTypes: []ResourceType{ResourceTypeR2},
			wantURLs: map[ResourceType]string{
				ResourceTypeR2: "https://dash.cloudflare.com/acc/r2/eu/buckets/my-bucket",
			},
		},
		{
			name: "Queue",
			config: &config.Config{
				Queues: &config.QueuesConfig{
					Producers: []config.QueueProducer{
						{Binding: "MY_QUEUE", Queue: "my-queue"},
					},
				},
			},
			wantTypes: []ResourceType{ResourceTypeQueue},
			wantURLs: map[ResourceType]string{
				ResourceTypeQueue: "https://dash.cloudflare.com/acc/workers/queues",
			},
		},
		{
			name: "Queue - 複数の queue をまとめる",
			config: &config.Config{
				Queues: &config.QueuesConfig{
					Producers: []config.QueueProducer{
						{Binding: "QUEUE1", Queue: "queue-1"},
						{Binding: "QUEUE2", Queue: "queue-2"},
					},
				},
			},
			wantTypes: []ResourceType{ResourceTypeQueue},
			wantURLs: map[ResourceType]string{
				ResourceTypeQueue: "https://dash.cloudflare.com/acc/workers/queues",
			},
		},
		{
			name: "Queue - consumer のみ",
			config: &config.Config{
				Queues: &config.QueuesConfig{
					Consumers: []config.QueueConsumer{
						{Queue: "my-queue"},
					},
				},
			},
			wantTypes: []ResourceType{ResourceTypeQueue},
			wantURLs: map[ResourceType]string{
				ResourceTypeQueue: "https://dash.cloudflare.com/acc/workers/queues",
			},
		},
		{
			name: "Workflow",
			config: &config.Config{
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
			name: "Durable Objects - 複数の class をまとめる",
			config: &config.Config{
				DurableObjects: &config.DurableObjectsConfig{
					Bindings: []config.DurableObjectBinding{
						{Name: "OBJECT1", ClassName: "Object1"},
						{Name: "OBJECT2", ClassName: "Object2"},
					},
				},
			},
			wantTypes: []ResourceType{ResourceTypeDurableObjects},
			wantURLs: map[ResourceType]string{
				ResourceTypeDurableObjects: "https://dash.cloudflare.com/acc/workers/durable-objects",
			},
		},
		{
			name: "Containers - 複数の container をまとめる",
			config: &config.Config{
				Containers: []config.Container{
					{Name: "app-1"},
					{Name: "app-2"},
				},
			},
			wantTypes: []ResourceType{ResourceTypeContainers},
			wantURLs: map[ResourceType]string{
				ResourceTypeContainers: "https://dash.cloudflare.com/acc/workers/containers",
			},
		},
		{
			name: "Vectorize",
			config: &config.Config{
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
			config: &config.Config{
				Pipelines: []config.Pipeline{
					{Binding: "MY_PIPELINE", Stream: "my-stream"},
				},
			},
			wantTypes: []ResourceType{ResourceTypePipeline},
			wantURLs: map[ResourceType]string{
				ResourceTypePipeline: "https://dash.cloudflare.com/acc/pipelines/streams/my-stream",
			},
		},
		{
			name: "Pipeline - 非推奨の pipeline キー",
			config: &config.Config{
				Pipelines: []config.Pipeline{
					{Binding: "MY_PIPELINE", Pipeline: "my-stream"},
				},
			},
			wantTypes: []ResourceType{ResourceTypePipeline},
			wantURLs: map[ResourceType]string{
				ResourceTypePipeline: "https://dash.cloudflare.com/acc/pipelines/streams/my-stream",
			},
		},
		{
			name: "Secrets Store",
			config: &config.Config{
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
			config: &config.Config{
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
			name: "Browser Run",
			config: &config.Config{
				Browser: &config.BrowserConfig{Binding: "MY_BROWSER"},
			},
			wantTypes: []ResourceType{ResourceTypeBrowserRun},
			wantURLs: map[ResourceType]string{
				ResourceTypeBrowserRun: "https://dash.cloudflare.com/acc/workers/browser-run",
			},
		},
		{
			name: "Workers AI",
			config: &config.Config{
				AI: &config.AIConfig{Binding: "AI"},
			},
			wantTypes: []ResourceType{ResourceTypeWorkersAI},
			wantURLs: map[ResourceType]string{
				ResourceTypeWorkersAI: "https://dash.cloudflare.com/acc/ai/workers-ai",
			},
		},
		{
			name: "Stream",
			config: &config.Config{
				Stream: &config.StreamConfig{Binding: "STREAM"},
			},
			wantTypes: []ResourceType{ResourceTypeStream},
			wantURLs: map[ResourceType]string{
				ResourceTypeStream: "https://dash.cloudflare.com/acc/stream/videos",
			},
		},
		{
			name: "AI Search",
			config: &config.Config{
				AISearch: []config.AISearchInstance{
					{Binding: "SEARCH", InstanceName: "my-instance"},
				},
			},
			wantTypes: []ResourceType{ResourceTypeAISearch},
			wantURLs: map[ResourceType]string{
				ResourceTypeAISearch: "https://dash.cloudflare.com/acc/ai/ai-search/namespace/default/instance/my-instance/overview",
			},
		},
		{
			name: "AI Search Namespace",
			config: &config.Config{
				AISearchNamespaces: []config.AISearchNamespace{
					{Binding: "SEARCH_NAMESPACE", Namespace: "my-namespace"},
				},
			},
			wantTypes: []ResourceType{ResourceTypeAISearchNamespace},
			wantURLs: map[ResourceType]string{
				ResourceTypeAISearchNamespace: "https://dash.cloudflare.com/acc/ai/ai-search?namespace=my-namespace",
			},
		},
		{
			name: "Artifacts",
			config: &config.Config{
				Artifacts: []config.Artifacts{
					{Binding: "ARTIFACTS", Namespace: "my-artifacts"},
				},
			},
			wantTypes: []ResourceType{ResourceTypeArtifacts},
			wantURLs: map[ResourceType]string{
				ResourceTypeArtifacts: "https://dash.cloudflare.com/acc/workers/artifacts/namespaces/my-artifacts",
			},
		},
		{
			name: "Analytics Engine - 複数の dataset をまとめる",
			config: &config.Config{
				AnalyticsEngineDatasets: []config.AnalyticsEngineDataset{
					{Binding: "EVENTS1", Dataset: "dataset-1"},
					{Binding: "EVENTS2", Dataset: "dataset-2"},
				},
			},
			wantTypes: []ResourceType{ResourceTypeAnalyticsEngine},
			wantURLs: map[ResourceType]string{
				ResourceTypeAnalyticsEngine: "https://dash.cloudflare.com/acc/workers/analytics-engine",
			},
		},
		{
			name: "Email Sending - 複数の binding をまとめる",
			config: &config.Config{
				SendEmail: []config.SendEmail{
					{Name: "EMAIL1"},
					{Name: "EMAIL2"},
				},
			},
			wantTypes: []ResourceType{ResourceTypeEmailSending},
			wantURLs: map[ResourceType]string{
				ResourceTypeEmailSending: "https://dash.cloudflare.com/acc/email-service/sending",
			},
		},
		{
			name: "Images",
			config: &config.Config{
				Images: &config.ImagesConfig{Binding: "MY_IMAGES"},
			},
			wantTypes: []ResourceType{ResourceTypeImages},
			wantURLs: map[ResourceType]string{
				ResourceTypeImages: "https://dash.cloudflare.com/acc/images",
			},
		},
		{
			name: "VPC Services",
			config: &config.Config{
				VPCServices: []config.VPCService{
					{Binding: "MY_VPC", ServiceID: "vpc-id"},
				},
			},
			wantTypes: []ResourceType{ResourceTypeVPC},
			wantURLs: map[ResourceType]string{
				ResourceTypeVPC: "https://dash.cloudflare.com/acc/workers/vpc/services/vpc-id",
			},
		},
		{
			name: "VPC Networks - tunnel_id の Tunnel も表示する",
			config: &config.Config{
				VPCNetworks: []config.VPCNetwork{
					{Binding: "TUNNEL_NETWORK", TunnelID: "tunnel-id"},
					{Binding: "MESH_NETWORK", NetworkID: "network-id"},
				},
			},
			wantTypes: []ResourceType{ResourceTypeVPCNetworks, ResourceTypeTunnel},
			wantURLs: map[ResourceType]string{
				ResourceTypeVPCNetworks: "https://dash.cloudflare.com/acc/workers/vpc/networks",
				ResourceTypeTunnel:      "https://dash.cloudflare.com/acc/tunnels/tunnel-id/overview",
			},
		},
		{
			name: "Flagship",
			config: &config.Config{
				Flagship: []config.Flagship{
					{Binding: "FLAGS", AppID: "app-id"},
				},
			},
			wantTypes: []ResourceType{ResourceTypeFlagship},
			wantURLs: map[ResourceType]string{
				ResourceTypeFlagship: "https://dash.cloudflare.com/acc/flagship/applications/app-id/overview",
			},
		},
		{
			name: "Flagship - app_id なしの場合は一覧を開く",
			config: &config.Config{
				Flagship: []config.Flagship{
					{Binding: "FLAGS"},
				},
			},
			wantTypes: []ResourceType{ResourceTypeFlagship},
			wantURLs: map[ResourceType]string{
				ResourceTypeFlagship: "https://dash.cloudflare.com/acc/flagship",
			},
		},
		{
			name: "Service Binding - 同じ参照先をまとめる",
			config: &config.Config{
				Services: []config.Service{
					{Binding: "API", Service: "api-worker"},
					{Binding: "API_ADMIN", Service: "api-worker"},
				},
			},
			wantTypes: []ResourceType{ResourceTypeService},
			wantURLs: map[ResourceType]string{
				ResourceTypeService: "https://dash.cloudflare.com/acc/workers/services/view/api-worker/production",
			},
		},
		{
			name: "Tail Worker - 通常とストリーミングをまとめる",
			config: &config.Config{
				TailConsumers:          []config.TailConsumer{{Service: "tail-worker"}},
				StreamingTailConsumers: []config.TailConsumer{{Service: "tail-worker"}},
			},
			wantTypes: []ResourceType{ResourceTypeTailWorker},
			wantURLs: map[ResourceType]string{
				ResourceTypeTailWorker: "https://dash.cloudflare.com/acc/workers/services/view/tail-worker/production",
			},
		},
		{
			name: "Cron Triggers",
			config: &config.Config{
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
			config: &config.Config{
				Triggers: &config.TriggersConfig{
					Crons: []string{"0 * * * *"},
				},
			},
			wantTypes: nil,
			wantURLs:  nil,
		},
		{
			name: "Workers Routes - ゾーン名のあるルートをゾーンごとにまとめる",
			config: &config.Config{
				Routes: []config.Route{
					{Pattern: "example.com/*"},
					{Pattern: "example.com/api/*", ZoneName: "example.com"},
					{Pattern: "example.com/app/*", ZoneName: "example.com"},
					{Pattern: "api.example.net", ZoneName: "example.net", CustomDomain: true},
				},
				Route: &config.Route{Pattern: "example.org/*", ZoneName: "example.org"},
			},
			wantTypes: []ResourceType{ResourceTypeWorkersRoutes, ResourceTypeWorkersRoutes},
			wantURLs:  nil,
		},
		{
			name: "Logpush",
			config: &config.Config{
				Logpush: true,
			},
			wantTypes: []ResourceType{ResourceTypeLogpush},
			wantURLs: map[ResourceType]string{
				ResourceTypeLogpush: "https://dash.cloudflare.com/acc/logs",
			},
		},
		{
			name: "Email Routing - 一覧とドメインごとのページ",
			config: &config.Config{
				Addresses: []string{"*@example.com", "admin@example.com", "invalid"},
			},
			wantTypes: []ResourceType{ResourceTypeEmailRouting, ResourceTypeEmailRoutingDomain},
			wantURLs: map[ResourceType]string{
				ResourceTypeEmailRouting:       "https://dash.cloudflare.com/acc/email-service/routing",
				ResourceTypeEmailRoutingDomain: "https://dash.cloudflare.com/acc/example.com/email/routing",
			},
		},
		{
			name: "Routes - Worker のルートとゾーンの Workers Routes",
			config: &config.Config{
				Name: "my-worker",
				Routes: []config.Route{
					{Pattern: "example.com/*", ZoneName: "example.com"},
				},
			},
			wantTypes: []ResourceType{ResourceTypeWorker, ResourceTypeRoutes, ResourceTypeWorkersRoutes},
			wantURLs: map[ResourceType]string{
				ResourceTypeRoutes:        "https://dash.cloudflare.com/acc/workers/services/view/my-worker/production/triggers",
				ResourceTypeWorkersRoutes: "https://dash.cloudflare.com/acc/example.com/workers",
			},
		},
		{
			name: "全リソース",
			config: &config.Config{
				Name:          "my-worker",
				Observability: &config.ObservabilityConfig{Enabled: true},
				KVNamespaces:  []config.KVNamespace{{Binding: "KV", ID: "kv-id"}},
				D1Databases:   []config.D1Database{{Binding: "DB", DatabaseName: "db", DatabaseID: "d1-id"}},
				Hyperdrive:    []config.Hyperdrive{{Binding: "HD", ID: "hd-id"}},
				R2Buckets:     []config.R2Bucket{{Binding: "R2", BucketName: "bucket"}},
				Queues:        &config.QueuesConfig{Producers: []config.QueueProducer{{Binding: "Q", Queue: "queue"}}},
				Workflows:     []config.Workflow{{Binding: "WF", Name: "workflow", ClassName: "WF"}},
				Vectorize:     []config.VectorizeIndex{{Binding: "VEC", IndexName: "index"}},
				Pipelines:     []config.Pipeline{{Binding: "PIPE", Pipeline: "pipeline"}},
				SecretsStoreSecrets: []config.SecretsStoreSecret{
					{Binding: "SEC", StoreID: "store", SecretName: "secret"},
				},
				DurableObjects: &config.DurableObjectsConfig{
					Bindings: []config.DurableObjectBinding{{Name: "DO", ClassName: "DO"}},
				},
				Containers:              []config.Container{{Name: "app"}},
				Browser:                 &config.BrowserConfig{Binding: "BROWSER"},
				AI:                      &config.AIConfig{Binding: "AI"},
				Stream:                  &config.StreamConfig{Binding: "STREAM"},
				AISearch:                []config.AISearchInstance{{Binding: "SEARCH", InstanceName: "instance"}},
				AISearchNamespaces:      []config.AISearchNamespace{{Binding: "SEARCH_NAMESPACE", Namespace: "namespace"}},
				AnalyticsEngineDatasets: []config.AnalyticsEngineDataset{{Binding: "EVENTS", Dataset: "dataset"}},
				SendEmail:               []config.SendEmail{{Name: "EMAIL"}},
				Images:                  &config.ImagesConfig{Binding: "IMAGES"},
				VPCServices:             []config.VPCService{{Binding: "VPC", ServiceID: "vpc"}},
				VPCNetworks:             []config.VPCNetwork{{Binding: "NETWORK", TunnelID: "tunnel-id"}},
				Flagship:                []config.Flagship{{Binding: "FLAGS", AppID: "app-id"}},
				Triggers:                &config.TriggersConfig{Crons: []string{"* * * * *"}},
			},
			wantTypes: []ResourceType{
				ResourceTypeWorker,
				ResourceTypeObservability,
				ResourceTypeCronTriggers,
				ResourceTypeQueue,
				ResourceTypeWorkflow,
				ResourceTypeDurableObjects,
				ResourceTypeContainers,
				ResourceTypeBrowserRun,
				ResourceTypeWorkersAI,
				ResourceTypeStream,
				ResourceTypeAISearch,
				ResourceTypeAISearchNamespace,
				ResourceTypeAnalyticsEngine,
				ResourceTypeEmailSending,
				ResourceTypeVPC,
				ResourceTypeVPCNetworks,
				ResourceTypeTunnel,
				ResourceTypeFlagship,
				ResourceTypeR2,
				ResourceTypeKV,
				ResourceTypeD1,
				ResourceTypeHyperdrive,
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

	cfg := &config.Config{
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
