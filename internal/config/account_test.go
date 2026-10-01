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

func TestGetAccountID_Cache(t *testing.T) {
	const (
		cfCache       = "node_modules/.cache/cloudflare/cloudflare-account.json"
		wranglerCache = "node_modules/.cache/wrangler/wrangler-account.json"
	)

	tests := []struct {
		name   string
		files  map[string]string
		wantID string
	}{
		{
			name:   "Wrangler のキャッシュ",
			files:  map[string]string{wranglerCache: `{"account": {"id": "wrangler-account", "name": "w"}}`},
			wantID: "wrangler-account",
		},
		{
			name:   "cf のキャッシュ",
			files:  map[string]string{cfCache: `{"account": {"id": "cf-account", "name": "c"}}`},
			wantID: "cf-account",
		},
		{
			name: "cf と Wrangler の両方のキャッシュ",
			files: map[string]string{
				cfCache:       `{"account": {"id": "cf-account", "name": "c"}}`,
				wranglerCache: `{"account": {"id": "wrangler-account", "name": "w"}}`,
			},
			wantID: "cf-account",
		},
		{
			name: "壊れた cf のキャッシュ",
			files: map[string]string{
				cfCache:       `{`,
				wranglerCache: `{"account": {"id": "wrangler-account", "name": "w"}}`,
			},
			wantID: "wrangler-account",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("CLOUDFLARE_ACCOUNT_ID", "")
			t.Setenv("WRANGLER_CACHE_DIR", "")

			root := t.TempDir()
			writeFiles(t, root, tt.files)
			t.Chdir(t.TempDir())

			gotID, gotHas := GetAccountID(&Config{Path: filepath.Join(root, "wrangler.jsonc")}, "")

			if gotID != tt.wantID || !gotHas {
				t.Errorf("GetAccountID() = (%q, %v), want (%q, true)", gotID, gotHas, tt.wantID)
			}
		})
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

func TestCacheFolder_CloudflareIgnoresWranglerCacheDir(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "node_modules"), 0o755); err != nil {
		t.Fatalf("テスト用ディレクトリの作成に失敗: %v", err)
	}
	t.Setenv("WRANGLER_CACHE_DIR", filepath.Join(root, "custom-cache"))

	got := cacheFolder(root, "cloudflare")

	if want := filepath.Join(root, "node_modules", ".cache", "cloudflare"); got != want {
		t.Errorf("cacheFolder() = %q, want %q", got, want)
	}
}
