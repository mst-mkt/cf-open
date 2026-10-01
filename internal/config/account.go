package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

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
	if accountID := readAccountCache(filepath.Join(cacheFolder(dir, "cloudflare"), "cloudflare-account.json")); accountID != "" {
		return accountID
	}

	return readAccountCache(filepath.Join(cacheFolder(dir, "wrangler"), "wrangler-account.json"))
}

func readAccountCache(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	var accountInfo AccountInfo
	if err := json.Unmarshal(data, &accountInfo); err != nil {
		return ""
	}

	return accountInfo.Account.ID
}

func cacheFolder(dir, namespace string) string {
	if namespace == "wrangler" {
		if envCacheDir := os.Getenv("WRANGLER_CACHE_DIR"); envCacheDir != "" {
			return envCacheDir
		}
	}

	if absDir, err := filepath.Abs(dir); err == nil {
		dir = absDir
	}

	localCache := filepath.Join(dir, "."+namespace, "cache")

	nodeModules := findDirectoryUp(dir, "node_modules")
	if nodeModules == "" {
		return localCache
	}

	nodeModulesCache := filepath.Join(nodeModules, ".cache", namespace)
	if exists(nodeModulesCache) || !exists(localCache) {
		return nodeModulesCache
	}

	return localCache
}

func findDirectoryUp(dir, name string) string {
	for {
		path := filepath.Join(dir, name)
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			return path
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}

		dir = parent
	}
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
