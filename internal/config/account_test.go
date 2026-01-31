package config

import "testing"

func TestGetAccountID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		config  *WranglerConfig
		wantID  string
		wantHas bool
	}{
		{
			name: "設定に account_id がある場合",
			config: &WranglerConfig{
				AccountID: "config-account-123",
			},
			wantID:  "config-account-123",
			wantHas: true,
		},
		{
			name: "設定に account_id がなく `wrangler-account.json` にもない場合",
			config: &WranglerConfig{
				AccountID: "",
			},
			wantID:  "",
			wantHas: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotID, gotHas := GetAccountID(tt.config)
			if gotID != tt.wantID {
				t.Errorf("GetAccountID() id = %q, want %q", gotID, tt.wantID)
			}
			if gotHas != tt.wantHas {
				t.Errorf("GetAccountID() hasAccount = %v, want %v", gotHas, tt.wantHas)
			}
		})
	}
}
