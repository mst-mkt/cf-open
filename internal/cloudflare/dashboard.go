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

	// Workers
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

	// Workers Observability
	if config.Name != "" && config.Observability != nil {
		observabilityURL := fmt.Sprintf("workers/services/view/%s/production/observability", config.Name)
		resources = append(resources, Resource{
			Type:        ResourceTypeObservability,
			Name:        config.Name,
			ID:          config.Name,
			Description: fmt.Sprintf("Observability: %s", config.Name),
			URL:         BuildDashboardURL(accountID, observabilityURL, hasAccount),
		})
	}

	// Workers Cron Triggers
	if config.Name != "" && config.Triggers != nil && len(config.Triggers.Crons) > 0 {
		cronURL := fmt.Sprintf("workers/services/view/%s/production/settings#trigger-events", config.Name)
		resources = append(resources, Resource{
			Type:        ResourceTypeCronTriggers,
			Name:        config.Name,
			ID:          config.Name,
			Description: fmt.Sprintf("Cron Triggers: %s", config.Name),
			URL:         BuildDashboardURL(accountID, cronURL, hasAccount),
		})
	}

	// Queues
	if config.Queues != nil {
		for _, producer := range config.Queues.Producers {
			queueURL := fmt.Sprintf("workers/queues/%s/metrics", producer.Queue)
			resources = append(resources, Resource{
				Type:        ResourceTypeQueue,
				Name:        producer.Binding,
				ID:          producer.Queue,
				Description: fmt.Sprintf("Queue: %s", producer.Queue),
				URL:         BuildDashboardURL(accountID, queueURL, hasAccount),
			})
		}
	}

	// Workflows
	for _, workflow := range config.Workflows {
		workflowURL := fmt.Sprintf("workers/workflows/%s/instances", workflow.Name)
		resources = append(resources, Resource{
			Type:        ResourceTypeWorkflow,
			Name:        workflow.Binding,
			ID:          workflow.Name,
			Description: fmt.Sprintf("Workflow: %s", workflow.Name),
			URL:         BuildDashboardURL(accountID, workflowURL, hasAccount),
		})
	}

	// Browser Rendering
	if config.Browser != nil && config.Browser.Binding != "" {
		browserURL := "workers/browser-rendering/overview"
		resources = append(resources, Resource{
			Type:        ResourceTypeBrowserRendering,
			Name:        config.Browser.Binding,
			ID:          "browser-rendering",
			Description: "Browser Rendering",
			URL:         BuildDashboardURL(accountID, browserURL, hasAccount),
		})
	}

	// VPC
	if len(config.VPCServices) > 0 {
		vpcURL := "workers/vpc/services"
		resources = append(resources, Resource{
			Type:        ResourceTypeVPC,
			Name:        "vpc",
			ID:          "vpc",
			Description: "VPC Services",
			URL:         BuildDashboardURL(accountID, vpcURL, hasAccount),
		})
	}

	// R2 Object Storage
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

	// Workers KV
	for _, kv := range config.KVNamespaces {
		kvURL := fmt.Sprintf("workers/kv/namespaces/%s/metrics", kv.ID)
		resources = append(resources, Resource{
			Type:        ResourceTypeKV,
			Name:        kv.Binding,
			ID:          kv.ID,
			Description: fmt.Sprintf("KV: %s (%s)", kv.Binding, kv.ID),
			URL:         BuildDashboardURL(accountID, kvURL, hasAccount),
		})
	}

	// D1 SQL Database
	for _, db := range config.D1Databases {
		d1URL := fmt.Sprintf("workers/d1/databases/%s/metrics", db.DatabaseID)
		resources = append(resources, Resource{
			Type:        ResourceTypeD1,
			Name:        db.Binding,
			ID:          db.DatabaseID,
			Description: fmt.Sprintf("D1: %s (%s)", db.DatabaseName, db.DatabaseID),
			URL:         BuildDashboardURL(accountID, d1URL, hasAccount),
		})
	}

	// Pipelines
	for _, pipeline := range config.Pipelines {
		pipelineURL := fmt.Sprintf("pipelines/%s/overview", pipeline.Pipeline)
		resources = append(resources, Resource{
			Type:        ResourceTypePipeline,
			Name:        pipeline.Binding,
			ID:          pipeline.Pipeline,
			Description: fmt.Sprintf("Pipeline: %s", pipeline.Pipeline),
			URL:         BuildDashboardURL(accountID, pipelineURL, hasAccount),
		})
	}

	// Vectorize
	for _, vectorize := range config.Vectorize {
		vectorizeURL := fmt.Sprintf("ai/vectorize/%s", vectorize.IndexName)
		resources = append(resources, Resource{
			Type:        ResourceTypeVectorize,
			Name:        vectorize.Binding,
			ID:          vectorize.IndexName,
			Description: fmt.Sprintf("Vectorize: %s", vectorize.IndexName),
			URL:         BuildDashboardURL(accountID, vectorizeURL, hasAccount),
		})
	}

	// Secrets Store
	seenStoreIDs := make(map[string]bool)
	for _, secret := range config.SecretsStoreSecrets {
		if seenStoreIDs[secret.StoreID] {
			continue
		}
		seenStoreIDs[secret.StoreID] = true

		secretsStoreURL := fmt.Sprintf("secrets-store/%s", secret.StoreID)
		resources = append(resources, Resource{
			Type:        ResourceTypeSecretsStore,
			Name:        secret.StoreID,
			ID:          secret.StoreID,
			Description: fmt.Sprintf("Secrets Store: %s", secret.StoreID),
			URL:         BuildDashboardURL(accountID, secretsStoreURL, hasAccount),
		})
	}

	// Images
	if config.Images != nil && config.Images.Binding != "" {
		imagesURL := "images"
		resources = append(resources, Resource{
			Type:        ResourceTypeImages,
			Name:        config.Images.Binding,
			ID:          "images",
			Description: "Images",
			URL:         BuildDashboardURL(accountID, imagesURL, hasAccount),
		})
	}

	return resources
}
