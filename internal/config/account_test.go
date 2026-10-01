package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetAccountID(t *testing.T) {
	tests := []struct {
		name          string
		config        *Config
		flagAccountID string
		envAccountID  string
		wantID        string
		wantHas       bool
	}{
		{
			name: "フラグで account_id が指定されている場合",
			config: &Config{
				AccountID: "config-account-123",
			},
			flagAccountID: "flag-account-456",
			envAccountID:  "env-account-789",
			wantID:        "flag-account-456",
			wantHas:       true,
		},
		{
			name: "環境変数で account_id が指定されている場合",
			config: &Config{
				AccountID: "config-account-123",
			},
			flagAccountID: "",
			envAccountID:  "env-account-789",
			wantID:        "env-account-789",
			wantHas:       true,
		},
		{
			name: "設定に account_id がある場合",
			config: &Config{
				AccountID: "config-account-123",
			},
			flagAccountID: "",
			envAccountID:  "",
			wantID:        "config-account-123",
			wantHas:       true,
		},
		{
			name: "設定に account_id がなく `wrangler-account.json` にもない場合",
			config: &Config{
				AccountID: "",
			},
			flagAccountID: "",
			envAccountID:  "",
			wantID:        "",
			wantHas:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("CLOUDFLARE_ACCOUNT_ID", tt.envAccountID)
			t.Setenv("WRANGLER_CACHE_DIR", "")
			t.Chdir(t.TempDir())

			gotID, gotHas := GetAccountID(tt.config, tt.flagAccountID)
			if gotID != tt.wantID {
				t.Errorf("GetAccountID() id = %q, want %q", gotID, tt.wantID)
			}
			if gotHas != tt.wantHas {
				t.Errorf("GetAccountID() hasAccount = %v, want %v", gotHas, tt.wantHas)
			}
		})
	}
}

func TestGetAccountID_CacheNextToConfig(t *testing.T) {
	t.Setenv("CLOUDFLARE_ACCOUNT_ID", "")
	t.Setenv("WRANGLER_CACHE_DIR", "")

	root := t.TempDir()
	writeFiles(t, root, map[string]string{
		"node_modules/.cache/wrangler/wrangler-account.json": `{"account": {"id": "cached-account-123", "name": "cached"}}`,
	})
	t.Chdir(t.TempDir())

	gotID, gotHas := GetAccountID(&Config{Path: filepath.Join(root, "wrangler.jsonc")}, "")

	if gotID != "cached-account-123" || !gotHas {
		t.Errorf("GetAccountID() = (%q, %v), want (%q, true)", gotID, gotHas, "cached-account-123")
	}
}

func TestCacheFolder(t *testing.T) {
	tests := []struct {
		name     string
		dirs     []string
		start    string
		envCache string
		want     string
	}{
		{
			name:  "node_modules のキャッシュ",
			dirs:  []string{"node_modules/.cache/wrangler", ".wrangler/cache"},
			start: ".",
			want:  "node_modules/.cache/wrangler",
		},
		{
			name:  "親ディレクトリの node_modules のキャッシュ",
			dirs:  []string{"node_modules/.cache/wrangler", "packages/app"},
			start: "packages/app",
			want:  "node_modules/.cache/wrangler",
		},
		{
			name:  "node_modules にキャッシュがなくローカルにある場合",
			dirs:  []string{"node_modules", ".wrangler/cache"},
			start: ".",
			want:  ".wrangler/cache",
		},
		{
			name:  "どちらにもキャッシュがない場合",
			dirs:  []string{"node_modules"},
			start: ".",
			want:  "node_modules/.cache/wrangler",
		},
		{
			name:  "node_modules がない場合",
			dirs:  []string{},
			start: ".",
			want:  ".wrangler/cache",
		},
		{
			name:     "環境変数でのキャッシュの指定",
			dirs:     []string{"node_modules/.cache/wrangler"},
			start:    ".",
			envCache: "custom-cache",
			want:     "custom-cache",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			for _, dir := range tt.dirs {
				if err := os.MkdirAll(filepath.Join(root, filepath.FromSlash(dir)), 0o755); err != nil {
					t.Fatalf("テスト用ディレクトリの作成に失敗: %v", err)
				}
			}

			envCache := ""
			want := filepath.Join(root, filepath.FromSlash(tt.want))
			if tt.envCache != "" {
				envCache = filepath.Join(root, tt.envCache)
				want = envCache
			}
			t.Setenv("WRANGLER_CACHE_DIR", envCache)

			got := cacheFolder(filepath.Join(root, filepath.FromSlash(tt.start)), "wrangler")

			if got != want {
				t.Errorf("cacheFolder() = %q, want %q", got, want)
			}
		})
	}
}
