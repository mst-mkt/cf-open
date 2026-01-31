package cloudflare

import (
	"fmt"

	"github.com/mst-mkt/cf-open/internal/config"
)

const baseURL = "https://dash.cloudflare.com"

func BuildDashboardURL(accountID, path string, hasAccount bool) string {
	if !hasAccount {
		return fmt.Sprintf("%s/?to=/:account/%s", baseURL, path)
	}
	return fmt.Sprintf("%s/%s/%s", baseURL, accountID, path)
}

func GetResourcesFromConfig(config *config.WranglerConfig, accountID string, hasAccount bool) []Resource {
	var resources []Resource

	if config.Name != "" {
		workerURL := fmt.Sprintf("workers/services/view/%s/production", config.Name)
		resources = append(resources, Resource{
			Type:        ResourceTypeWorker,
			Name:        config.Name,
			ID:          config.Name,
			Description: fmt.Sprintf("Worker: %s", config.Name),
			URL:         BuildDashboardURL(accountID, workerURL, hasAccount),
		})
	}

	for _, kv := range config.KVNamespaces {
		kvURL := fmt.Sprintf("workers/kv/namespaces/%s", kv.ID)
		resources = append(resources, Resource{
			Type:        ResourceTypeKV,
			Name:        kv.Binding,
			ID:          kv.ID,
			Description: fmt.Sprintf("KV: %s (%s)", kv.Binding, kv.ID),
			URL:         BuildDashboardURL(accountID, kvURL, hasAccount),
		})
	}

	for _, db := range config.D1Databases {
		d1URL := fmt.Sprintf("workers/d1/databases/%s", db.DatabaseID)
		resources = append(resources, Resource{
			Type:        ResourceTypeD1,
			Name:        db.Binding,
			ID:          db.DatabaseID,
			Description: fmt.Sprintf("D1: %s (%s)", db.DatabaseName, db.DatabaseID),
			URL:         BuildDashboardURL(accountID, d1URL, hasAccount),
		})
	}

	for _, bucket := range config.R2Buckets {
		r2URL := fmt.Sprintf("r2/default/buckets/%s", bucket.BucketName)
		resources = append(resources, Resource{
			Type:        ResourceTypeR2,
			Name:        bucket.Binding,
			ID:          bucket.BucketName,
			Description: fmt.Sprintf("R2: %s", bucket.BucketName),
			URL:         BuildDashboardURL(accountID, r2URL, hasAccount),
		})
	}

	return resources
}