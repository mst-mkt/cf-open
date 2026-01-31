package config

import (
	"encoding/json"
	"os"
)

const defaultWranglerCachePath = "node_modules/.cache/wrangler/wrangler-account.json"

type AccountInfo struct {
	Account struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"account"`
}

func GetAccountID(config *WranglerConfig) (string, bool) {
	if config.AccountID != "" {
		return config.AccountID, true
	}

	if accountID := getAccountFromCache(); accountID != "" {
		return accountID, true
	}

	return "", false
}

func getAccountFromCache() string {
	cacheFile := defaultWranglerCachePath

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
