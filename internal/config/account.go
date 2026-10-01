package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const defaultWranglerCachePath = "node_modules/.cache/wrangler/wrangler-account.json"

type AccountInfo struct {
	Account struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"account"`
}

func GetAccountID(config *Config, flagAccountID string) (string, bool) {
	if flagAccountID != "" {
		return flagAccountID, true
	}

	if accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID"); accountID != "" {
		return accountID, true
	}

	if config.AccountID != "" {
		return config.AccountID, true
	}

	if accountID := getAccountFromCache(filepath.Dir(config.Path)); accountID != "" {
		return accountID, true
	}

	return "", false
}

func getAccountFromCache(dir string) string {
	cacheFile := filepath.Join(dir, defaultWranglerCachePath)

	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return ""
	}

	var accountInfo AccountInfo
	if err := json.Unmarshal(data, &accountInfo); err != nil {
		return ""
	}

	return accountInfo.Account.ID
}
