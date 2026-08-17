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

func TestLoadWranglerConfig_TypeScript(t *testing.T) {
	t.Parallel()
	requireNode(t)

	tests := []struct {
		name    string
		content string
		files   map[string]string
		want    *WranglerConfig
		wantErr string
	}{
		{
			name: "プレーンオブジェクト形式",
			content: `export default {
  type: 'worker',
  name: 'plain-worker',
  observability: { enabled: true },
  triggers: [{ type: 'scheduled', schedule: '0 * * * *' }],
  env: {
    DB: { type: 'd1', name: 'my-db', id: 'db-id' },
    SECRET: { type: 'secrets-store-secret', storeId: 'store-id', secretName: 'my-secret' },
    VPC: { type: 'vpc-service', id: 'vpc-id' },
  },
}
`,
			want: &WranglerConfig{
				Name:                "plain-worker",
				Observability:       &ObservabilityConfig{Enabled: true},
				Triggers:            &TriggersConfig{Crons: []string{"0 * * * *"}},
				VPCServices:         []VPCService{{Binding: "VPC", ServiceID: "vpc-id"}},
				D1Databases:         []D1Database{{Binding: "DB", DatabaseName: "my-db", DatabaseID: "db-id"}},
				SecretsStoreSecrets: []SecretsStoreSecret{{Binding: "SECRET", StoreID: "store-id", SecretName: "my-secret"}},
			},
		},
		{
			name: "defineWorker ヘルパー形式 + entrypoint import",
			content: `import * as entrypoint from './src/index.ts' with { type: 'cf-worker' }

const DEFINITION = Symbol.for('@cloudflare/config:definition')
const defineWorker = (config) => ({ [DEFINITION]: { config, type: 'worker' } })
const defineSettings = (config) => ({ [DEFINITION]: { config, type: 'settings' } })

export const settings = defineSettings({ accountId: 'acc-123' })

export default defineWorker(async () => ({
  name: 'helper-worker',
  entrypoint,
  env: { DB: { type: 'd1', name: 'my-db', id: 'db-id' } },
}))
`,
			files: map[string]string{
				"src/index.ts": "throw new Error('the entrypoint must not be evaluated')\n",
			},
			want: &WranglerConfig{
				Name:        "helper-worker",
				AccountID:   "acc-123",
				D1Databases: []D1Database{{Binding: "DB", DatabaseName: "my-db", DatabaseID: "db-id"}},
			},
		},
		{
			name:    "Promise 形式",
			content: "export default Promise.resolve({ type: 'worker', name: 'promise-worker' })\n",
			want:    &WranglerConfig{Name: "promise-worker"},
		},
		{
			name: "stdout に書き込む設定",
			content: `console.log('noise from the config')
export default { type: 'worker', name: 'noisy-worker' }
`,
			want: &WranglerConfig{Name: "noisy-worker"},
		},
		{
			name: "イベントループを保持する設定",
			content: `setInterval(() => {}, 100000)
export default { type: 'worker', name: 'lingering-worker' }
`,
			want: &WranglerConfig{Name: "lingering-worker"},
		},
		{
			name:    "worker でない default export",
			content: "export default { name: 'no-type' }\n",
			wantErr: "cloudflare.config.ts: the default export must be a worker",
		},
		{
			name:    "型ストリップできない TypeScript 構文",
			content: "enum Mode { Prod }\nexport default { type: 'worker', name: `w-${Mode.Prod}` }\n",
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

			got, err := LoadWranglerConfig(configPath)

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("LoadWranglerConfig() error = nil, want an error containing %q", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("LoadWranglerConfig() error = %v, want it to contain %q", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("LoadWranglerConfig() error = %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LoadWranglerConfig()\n got = %+v\nwant = %+v", got, tt.want)
			}
		})
	}
}

func TestLoadWranglerConfig_TypeScriptMode(t *testing.T) {
	requireNode(t)
	t.Setenv("CLOUDFLARE_ENV", "staging")

	content := "export default (ctx) => ({ type: 'worker', name: `worker-${ctx.mode ?? 'none'}` })\n"
	configPath := writeFixture(t, content, nil)

	cfg, err := LoadWranglerConfig(configPath)
	if err != nil {
		t.Fatalf("LoadWranglerConfig() error = %v", err)
	}

	if cfg.Name != "worker-staging" {
		t.Errorf("Name = %q, want %q", cfg.Name, "worker-staging")
	}
}

func TestLoadWranglerConfig_TypeScriptWithoutNode(t *testing.T) {
	t.Setenv("PATH", "")

	configPath := writeFixture(t, "export default { type: 'worker', name: 'x' }\n", nil)

	_, err := LoadWranglerConfig(configPath)

	if err == nil || !strings.Contains(err.Error(), "node not found in PATH") {
		t.Errorf("LoadWranglerConfig() error = %v, want a node-not-found error", err)
	}
}

func TestLoadWranglerConfig_TypeScriptNotFound(t *testing.T) {
	t.Parallel()

	configPath := filepath.Join(t.TempDir(), "cloudflare.config.ts")

	_, err := LoadWranglerConfig(configPath)

	if err == nil || !strings.Contains(err.Error(), "failed to find config file") {
		t.Errorf("LoadWranglerConfig() error = %v, want a not-found error", err)
	}
}

func TestToWranglerConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		loaded string
		want   *WranglerConfig
	}{
		{
			name: "全フィールドの取り込み",
			loaded: `{
				"worker": {
					"type": "worker",
					"name": "my-worker",
					"observability": {"enabled": true},
					"triggers": [
						{"type": "scheduled", "schedule": "0 * * * *"},
						{"type": "fetch"},
						{"type": "scheduled", "schedule": "0 0 * * *"}
					],
					"env": {
						"DB": {"type": "d1", "name": "my-db", "id": "db-id"},
						"BUCKET": {"type": "r2", "name": "my-bucket"},
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
				"settings": {"type": "settings", "accountId": "acc-123"}
			}`,
			want: &WranglerConfig{
				Name:                "my-worker",
				AccountID:           "acc-123",
				Observability:       &ObservabilityConfig{Enabled: true},
				Triggers:            &TriggersConfig{Crons: []string{"0 * * * *", "0 0 * * *"}},
				Queues:              &QueuesConfig{Producers: []QueueProducer{{Binding: "QUEUE", Queue: "my-queue"}}},
				Browser:             &BrowserConfig{Binding: "BROWSER"},
				VPCServices:         []VPCService{{Binding: "VPC", ServiceID: "vpc-id"}},
				R2Buckets:           []R2Bucket{{Binding: "BUCKET", BucketName: "my-bucket"}},
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
					"type": "worker",
					"name": "my-worker",
					"env": {
						"DB": {"type": "d1", "name": "id-less-db"},
						"BUCKET": {"type": "r2"},
						"CACHE": {"type": "kv"},
						"QUEUE": {"type": "queue"}
					}
				}
			}`,
			want: &WranglerConfig{Name: "my-worker"},
		},
		{
			name: "未対応 binding の無視",
			loaded: `{
				"worker": {
					"type": "worker",
					"name": "my-worker",
					"env": {
						"AI": {"type": "ai"},
						"HYPERDRIVE": {"type": "hyperdrive", "id": "hd-id"},
						"DO": {"type": "durable-object", "workerName": "w", "exportName": "MyDurableObject"},
						"DB": {"type": "d1", "id": "db-id"}
					}
				}
			}`,
			want: &WranglerConfig{
				Name:        "my-worker",
				D1Databases: []D1Database{{Binding: "DB", DatabaseID: "db-id"}},
			},
		},
		{
			name: "同じ種別の binding の名前順",
			loaded: `{
				"worker": {
					"type": "worker",
					"name": "my-worker",
					"env": {
						"ZED": {"type": "d1", "id": "zed-id"},
						"ALPHA": {"type": "d1", "id": "alpha-id"},
						"MID": {"type": "d1", "id": "mid-id"}
					}
				}
			}`,
			want: &WranglerConfig{
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
			loaded: `{"worker": {"type": "worker", "name": "my-worker", "triggers": [{"type": "fetch"}]}}`,
			want:   &WranglerConfig{Name: "my-worker"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			loaded := &typeScriptConfig{}
			if err := json.Unmarshal([]byte(tt.loaded), loaded); err != nil {
				t.Fatalf("テスト入力のパースに失敗: %v", err)
			}

			got := toWranglerConfig(loaded.Worker, loaded.Settings)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toWranglerConfig()\n got = %+v\nwant = %+v", got, tt.want)
			}
		})
	}
}

func TestFindWranglerConfig_Priority(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"cloudflare.config.ts", "wrangler.jsonc"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("{}"), 0o644); err != nil {
			t.Fatalf("テスト設定ファイルの書き込みに失敗: %v", err)
		}
	}
	t.Chdir(dir)

	got := findWranglerConfig()

	if got != "cloudflare.config.ts" {
		t.Errorf("findWranglerConfig() = %q, want %q", got, "cloudflare.config.ts")
	}
}

func writeFixture(t *testing.T, content string, files map[string]string) string {
	t.Helper()

	dir := t.TempDir()
	for name, body := range files {
		path := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("テスト用ディレクトリの作成に失敗: %v", err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatalf("テスト用ファイルの書き込みに失敗: %v", err)
		}
	}

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
