package config

import (
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

	root := t.TempDir()
	writeFiles(t, root, map[string]string{
		defaultWranglerCachePath: `{"account": {"id": "cached-account-123", "name": "cached"}}`,
	})
	t.Chdir(t.TempDir())

	gotID, gotHas := GetAccountID(&Config{Path: filepath.Join(root, "wrangler.jsonc")}, "")

	if gotID != "cached-account-123" || !gotHas {
		t.Errorf("GetAccountID() = (%q, %v), want (%q, true)", gotID, gotHas, "cached-account-123")
	}
}
