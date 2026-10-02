package config

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestLoad_TypeScript(t *testing.T) {
	t.Parallel()
	requireNode(t)

	tests := []struct {
		name    string
		content string
		files   map[string]string
		want    *Config
		wantErr string
	}{
		{
			name: "オブジェクト形式",
			content: `export default {
  accountId: 'acc-123',
  worker: {
    name: 'plain-worker',
    compatibilityDate: '2026-09-30',
    observability: { enabled: true },
    triggers: [{ type: 'scheduled', schedule: '0 * * * *' }],
    env: {
      DB: { type: 'd1', name: 'my-db', id: 'db-id' },
      SECRET: { type: 'secrets-store-secret', storeId: 'store-id', secretName: 'my-secret' },
      VPC: { type: 'vpc-service', id: 'vpc-id' },
    },
  },
}
`,
			want: &Config{
				Name:                "plain-worker",
				AccountID:           "acc-123",
				Observability:       &ObservabilityConfig{Enabled: true},
				Triggers:            &TriggersConfig{Crons: []string{"0 * * * *"}},
				VPCServices:         []VPCService{{Binding: "VPC", ServiceID: "vpc-id"}},
				D1Databases:         []D1Database{{Binding: "DB", DatabaseName: "my-db", DatabaseID: "db-id"}},
				SecretsStoreSecrets: []SecretsStoreSecret{{Binding: "SECRET", StoreID: "store-id", SecretName: "my-secret"}},
			},
		},
		{
			name: "関数形式の worker + entrypoint import",
			content: `import * as entrypoint from './src/index.ts' with { type: 'cf-worker' }

const worker = async () => ({
  name: 'function-worker',
  entrypoint,
  env: { DB: { type: 'd1', name: 'my-db', id: 'db-id' } },
})

export default () => ({ accountId: 'acc-123', worker })
`,
			files: map[string]string{
				"src/index.ts": "throw new Error('the entrypoint must not be evaluated')\n",
			},
			want: &Config{
				Name:        "function-worker",
				AccountID:   "acc-123",
				D1Databases: []D1Database{{Binding: "DB", DatabaseName: "my-db", DatabaseID: "db-id"}},
			},
		},
		{
			name:    "Promise 形式",
			content: "export default Promise.resolve({ worker: Promise.resolve({ name: 'promise-worker' }) })\n",
			want:    &Config{Name: "promise-worker"},
		},
		{
			name: "stdout に書き込む設定",
			content: `console.log('noise from the config')
export default { worker: { name: 'noisy-worker' } }
`,
			want: &Config{Name: "noisy-worker"},
		},
		{
			name: "イベントループを保持する設定",
			content: `setInterval(() => {}, 100000)
export default { worker: { name: 'lingering-worker' } }
`,
			want: &Config{Name: "lingering-worker"},
		},
		{
			name: "Workflow を定義する設定",
			content: `export default {
  worker: {
    name: 'workflow-worker',
    exports: { MyWorkflow: { type: 'workflow', name: 'my-workflow' } },
    env: { MY_WORKFLOW: { type: 'workflow', name: 'my-workflow', worker: 'workflow-worker', exportName: 'MyWorkflow' } },
  },
}
`,
			want: &Config{
				Name:      "workflow-worker",
				Workflows: []Workflow{{Binding: "MY_WORKFLOW", Name: "my-workflow", ClassName: "MyWorkflow"}},
			},
		},
		{
			name:    "containers だけの設定",
			content: "export default { accountId: 'acc-123', containers: [{ name: 'app', image: { dockerfile: './Dockerfile' } }] }\n",
			want:    &Config{AccountID: "acc-123"},
		},
		{
			name:    "worker が null の設定",
			content: "export default { worker: () => null }\n",
			want:    &Config{},
		},
		{
			name:    "default export のない設定",
			content: "export const worker = { name: 'named-only' }\n",
			wantErr: "cloudflare.config.ts: the config has no default export",
		},
		{
			name:    "オブジェクトでない default export",
			content: "export default 'not-a-config'\n",
			wantErr: "cloudflare.config.ts: the default export must be an object",
		},
		{
			name:    "オブジェクトでない worker",
			content: "export default { worker: () => 'not-a-worker' }\n",
			wantErr: "cloudflare.config.ts: the worker must be an object",
		},
		{
			name:    "文字列でない accountId",
			content: "export default { accountId: 123, worker: { name: 'w' } }\n",
			wantErr: "cloudflare.config.ts: accountId must be a string",
		},
		{
			name:    "型ストリップできない TypeScript 構文",
			content: "enum Mode { Prod }\nexport default { worker: { name: `w-${Mode.Prod}` } }\n",
			wantErr: "cloudflare.config.ts:",
		},
		{
			name:    "設定の評価中に投げられたエラー",
			content: "export default () => { throw new Error('missing FOO env var') }\n",
			wantErr: "cloudflare.config.ts: missing FOO env var",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			configPath := writeFixture(t, tt.content, tt.files)

			got, err := Load(Options{Path: configPath})

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("Load() error = nil, want an error containing %q", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("Load() error = %v, want it to contain %q", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}
			want := *tt.want
			want.Path = configPath
			if !reflect.DeepEqual(got, &want) {
				t.Errorf("Load()\n got = %+v\nwant = %+v", got, &want)
			}
		})
	}
}

func TestLoad_TypeScriptMode(t *testing.T) {
	requireNode(t)
	t.Setenv("CLOUDFLARE_ENV", "staging")

	content := "export default (ctx) => ({ worker: { name: `worker-${ctx.mode ?? 'none'}-${ctx.isPreview}` } })\n"
	configPath := writeFixture(t, content, nil)

	tests := []struct {
		name string
		mode string
		want string
	}{
		{name: "mode の指定", mode: "production", want: "worker-production-false"},
		{name: "mode の未指定", mode: "", want: "worker-none-false"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := Load(Options{Path: configPath, Mode: tt.mode})
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			if cfg.Name != tt.want {
				t.Errorf("Name = %q, want %q", cfg.Name, tt.want)
			}
		})
	}
}

func TestLoad_TypeScriptWithoutNode(t *testing.T) {
	t.Setenv("PATH", "")

	configPath := writeFixture(t, "export default { worker: { name: 'x' } }\n", nil)

	_, err := Load(Options{Path: configPath})

	if err == nil || !strings.Contains(err.Error(), "node not found in PATH") {
		t.Errorf("Load() error = %v, want a node-not-found error", err)
	}
}

func TestLoad_TypeScriptNotFound(t *testing.T) {
	t.Parallel()

	configPath := filepath.Join(t.TempDir(), "cloudflare.config.ts")

	_, err := Load(Options{Path: configPath})

	if err == nil || !strings.Contains(err.Error(), "failed to find config file") {
		t.Errorf("Load() error = %v, want a not-found error", err)
	}
}

func TestToConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		loaded string
		want   *Config
	}{
		{
			name: "全フィールドの取り込み",
			loaded: `{
				"worker": {
					"name": "my-worker",
					"observability": {"enabled": true},
					"triggers": [
						{"type": "scheduled", "schedule": "0 * * * *"},
						{"type": "fetch"},
						{"type": "queue", "name": "consumed-queue"},
						{"type": "scheduled", "schedule": "0 0 * * *"}
					],
					"env": {
						"DB": {"type": "d1", "name": "my-db", "id": "db-id"},
						"BUCKET": {"type": "r2", "name": "my-bucket", "jurisdiction": "eu"},
						"CACHE": {"type": "kv", "id": "kv-id"},
						"QUEUE": {"type": "queue", "name": "my-queue"},
						"INDEX": {"type": "vectorize", "name": "my-index"},
						"PIPE": {"type": "pipeline", "name": "my-pipeline"},
						"SECRET": {"type": "secrets-store-secret", "storeId": "store-id", "secretName": "my-secret"},
						"VPC": {"type": "vpc-service", "id": "vpc-id"},
						"BROWSER": {"type": "browser"},
						"IMAGES": {"type": "images"}
					}
				},
				"accountId": "acc-123"
			}`,
			want: &Config{
				Name:                "my-worker",
				AccountID:           "acc-123",
				Observability:       &ObservabilityConfig{Enabled: true},
				Triggers:            &TriggersConfig{Crons: []string{"0 * * * *", "0 0 * * *"}},
				Queues:              &QueuesConfig{Producers: []QueueProducer{{Binding: "QUEUE", Queue: "my-queue"}}, Consumers: []QueueConsumer{{Queue: "consumed-queue"}}},
				Browser:             &BrowserConfig{Binding: "BROWSER"},
				VPCServices:         []VPCService{{Binding: "VPC", ServiceID: "vpc-id"}},
				R2Buckets:           []R2Bucket{{Binding: "BUCKET", BucketName: "my-bucket", Jurisdiction: "eu"}},
				KVNamespaces:        []KVNamespace{{Binding: "CACHE", ID: "kv-id"}},
				D1Databases:         []D1Database{{Binding: "DB", DatabaseName: "my-db", DatabaseID: "db-id"}},
				Pipelines:           []Pipeline{{Binding: "PIPE", Pipeline: "my-pipeline"}},
				Vectorize:           []VectorizeIndex{{Binding: "INDEX", IndexName: "my-index"}},
				SecretsStoreSecrets: []SecretsStoreSecret{{Binding: "SECRET", StoreID: "store-id", SecretName: "my-secret"}},
				Images:              &ImagesConfig{Binding: "IMAGES"},
			},
		},
		{
			name: "URL に必要な値が欠けた binding の除外",
			loaded: `{
				"worker": {
					"name": "my-worker",
					"env": {
						"DB": {"type": "d1", "name": "id-less-db"},
						"BUCKET": {"type": "r2"},
						"CACHE": {"type": "kv"},
						"QUEUE": {"type": "queue"}
					}
				}
			}`,
			want: &Config{Name: "my-worker"},
		},
		{
			name: "未対応 binding の無視",
			loaded: `{
				"worker": {
					"name": "my-worker",
					"env": {
						"AI": {"type": "ai"},
						"HYPERDRIVE": {"type": "hyperdrive", "id": "hd-id"},
						"DO": {"type": "durable-object", "workerName": "w", "exportName": "MyDurableObject"},
						"DB": {"type": "d1", "id": "db-id"}
					}
				}
			}`,
			want: &Config{
				Name:        "my-worker",
				D1Databases: []D1Database{{Binding: "DB", DatabaseID: "db-id"}},
			},
		},
		{
			name: "同じ種別の binding の名前順",
			loaded: `{
				"worker": {
					"name": "my-worker",
					"env": {
						"ZED": {"type": "d1", "id": "zed-id"},
						"ALPHA": {"type": "d1", "id": "alpha-id"},
						"MID": {"type": "d1", "id": "mid-id"}
					}
				}
			}`,
			want: &Config{
				Name: "my-worker",
				D1Databases: []D1Database{
					{Binding: "ALPHA", DatabaseID: "alpha-id"},
					{Binding: "MID", DatabaseID: "mid-id"},
					{Binding: "ZED", DatabaseID: "zed-id"},
				},
			},
		},
		{
			name:   "scheduled trigger のない設定",
			loaded: `{"worker": {"name": "my-worker", "triggers": [{"type": "fetch"}]}}`,
			want:   &Config{Name: "my-worker"},
		},
		{
			name: "binding と exports からの Workflow の収集",
			loaded: `{
				"worker": {
					"name": "my-worker",
					"env": {
						"REMOTE": {"type": "workflow", "name": "remote-workflow", "worker": "other-worker", "exportName": "RemoteWorkflow"},
						"LOCAL": {"type": "workflow", "name": "local-workflow", "worker": "my-worker", "exportName": "LocalWorkflow"},
						"DUPLICATE": {"type": "workflow", "name": "local-workflow", "worker": "my-worker", "exportName": "LocalWorkflow"},
						"NAMELESS": {"type": "workflow", "worker": "my-worker", "exportName": "NamelessWorkflow"}
					},
					"exports": {
						"LocalWorkflow": {"type": "workflow", "name": "local-workflow"},
						"UnboundWorkflow": {"type": "workflow", "name": "unbound-workflow"},
						"Counter": {"type": "durable-object", "storage": "sqlite"},
						"Api": {"type": "worker"}
					}
				}
			}`,
			want: &Config{
				Name: "my-worker",
				Workflows: []Workflow{
					{Binding: "DUPLICATE", Name: "local-workflow", ClassName: "LocalWorkflow"},
					{Binding: "REMOTE", Name: "remote-workflow", ClassName: "RemoteWorkflow"},
					{Name: "unbound-workflow", ClassName: "UnboundWorkflow"},
				},
			},
		},
		{
			name:   "worker のない設定",
			loaded: `{"accountId": "acc-123"}`,
			want:   &Config{AccountID: "acc-123"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			loaded := &typeScriptConfig{}
			if err := json.Unmarshal([]byte(tt.loaded), loaded); err != nil {
				t.Fatalf("テスト入力のパースに失敗: %v", err)
			}

			got := toConfig(loaded.Worker, loaded.AccountID)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toConfig()\n got = %+v\nwant = %+v", got, tt.want)
			}
		})
	}
}

func writeFixture(t *testing.T, content string, files map[string]string) string {
	t.Helper()

	dir := t.TempDir()
	writeFiles(t, dir, files)

	configPath := filepath.Join(dir, "cloudflare.config.ts")
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatalf("テスト設定ファイルの書き込みに失敗: %v", err)
	}

	return configPath
}

func requireNode(t *testing.T) {
	t.Helper()

	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node が PATH にないためスキップ")
	}

	output, err := exec.Command(node, "--version").Output()
	if err != nil {
		t.Skipf("node のバージョンを取得できないためスキップ: %v", err)
	}

	version := strings.TrimSpace(string(output))
	if !supportsTypeStripping(version) {
		t.Skipf("node %s は v22.18.0 未満のためスキップ", version)
	}
}

func supportsTypeStripping(version string) bool {
	parts := strings.Split(strings.TrimPrefix(version, "v"), ".")
	if len(parts) < 2 {
		return false
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return false
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return false
	}

	return major > 22 || (major == 22 && minor >= 18)
}
