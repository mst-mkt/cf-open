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

func GetResourcesFromConfig(cfg *config.Config, accountID string, hasAccount bool) []Resource {
	var resources []Resource

	isPages := cfg.PagesBuildOutputDir != ""

	// Pages
	if cfg.Name != "" && isPages {
		pagesURL := fmt.Sprintf("pages/view/%s", cfg.Name)
		resources = append(resources, Resource{
			Type:        ResourceTypePages,
			Name:        cfg.Name,
			ID:          cfg.Name,
			Description: fmt.Sprintf("Pages: %s", cfg.Name),
			URL:         BuildDashboardURL(accountID, pagesURL, hasAccount),
		})
	}

	// Workers
	if cfg.Name != "" && !isPages {
		workerURL := fmt.Sprintf("workers/services/view/%s/production", cfg.Name)
		resources = append(resources, Resource{
			Type:        ResourceTypeWorker,
			Name:        cfg.Name,
			ID:          cfg.Name,
			Description: fmt.Sprintf("Worker: %s", cfg.Name),
			URL:         BuildDashboardURL(accountID, workerURL, hasAccount),
		})
	}

	// Workers Observability
	if cfg.Name != "" && !isPages && cfg.Observability != nil {
		observabilityURL := fmt.Sprintf("workers/services/view/%s/production/observability", cfg.Name)
		resources = append(resources, Resource{
			Type:        ResourceTypeObservability,
			Name:        cfg.Name,
			ID:          cfg.Name,
			Description: fmt.Sprintf("Observability: %s", cfg.Name),
			URL:         BuildDashboardURL(accountID, observabilityURL, hasAccount),
		})
	}

	// Workers Cron Triggers
	if cfg.Name != "" && !isPages && cfg.Triggers != nil && len(cfg.Triggers.Crons) > 0 {
		cronURL := fmt.Sprintf("workers/services/view/%s/production/settings#trigger-events", cfg.Name)
		resources = append(resources, Resource{
			Type:        ResourceTypeCronTriggers,
			Name:        cfg.Name,
			ID:          cfg.Name,
			Description: fmt.Sprintf("Cron Triggers: %s", cfg.Name),
			URL:         BuildDashboardURL(accountID, cronURL, hasAccount),
		})
	}

	// Queues
	// The detail page is addressed by queue ID, not the name in the config, so open the list.
	if cfg.Queues != nil && (len(cfg.Queues.Producers) > 0 || len(cfg.Queues.Consumers) > 0) {
		queueURL := "workers/queues"
		resources = append(resources, Resource{
			Type:        ResourceTypeQueue,
			Name:        "queues",
			ID:          "queues",
			Description: "Queues",
			URL:         BuildDashboardURL(accountID, queueURL, hasAccount),
		})
	}

	// Workflows
	for _, workflow := range cfg.Workflows {
		workflowURL := fmt.Sprintf("workers/workflows/%s/instances", workflow.Name)
		resources = append(resources, Resource{
			Type:        ResourceTypeWorkflow,
			Name:        workflow.Binding,
			ID:          workflow.Name,
			Description: fmt.Sprintf("Workflow: %s", workflow.Name),
			URL:         BuildDashboardURL(accountID, workflowURL, hasAccount),
		})
	}

	// Durable Objects
	// The detail page is addressed by namespace ID, not the class name in the config, so open the list.
	if cfg.DurableObjects != nil && len(cfg.DurableObjects.Bindings) > 0 {
		durableObjectsURL := "workers/durable-objects"
		resources = append(resources, Resource{
			Type:        ResourceTypeDurableObjects,
			Name:        "durable-objects",
			ID:          "durable-objects",
			Description: "Durable Objects",
			URL:         BuildDashboardURL(accountID, durableObjectsURL, hasAccount),
		})
	}

	// Browser Run
	if cfg.Browser != nil && cfg.Browser.Binding != "" {
		browserURL := "workers/browser-run"
		resources = append(resources, Resource{
			Type:        ResourceTypeBrowserRun,
			Name:        cfg.Browser.Binding,
			ID:          "browser-run",
			Description: "Browser Run",
			URL:         BuildDashboardURL(accountID, browserURL, hasAccount),
		})
	}

	// Workers AI
	if cfg.AI != nil && cfg.AI.Binding != "" {
		aiURL := "ai/workers-ai"
		resources = append(resources, Resource{
			Type:        ResourceTypeWorkersAI,
			Name:        cfg.AI.Binding,
			ID:          "workers-ai",
			Description: "Workers AI",
			URL:         BuildDashboardURL(accountID, aiURL, hasAccount),
		})
	}

	// AI Search
	for _, instance := range cfg.AISearch {
		aiSearchURL := fmt.Sprintf("ai/ai-search/namespace/default/instance/%s/overview", instance.InstanceName)
		resources = append(resources, Resource{
			Type:        ResourceTypeAISearch,
			Name:        instance.Binding,
			ID:          instance.InstanceName,
			Description: fmt.Sprintf("AI Search: %s", instance.InstanceName),
			URL:         BuildDashboardURL(accountID, aiSearchURL, hasAccount),
		})
	}

	// AI Search Namespaces
	for _, namespace := range cfg.AISearchNamespaces {
		namespaceURL := fmt.Sprintf("ai/ai-search?namespace=%s", namespace.Namespace)
		resources = append(resources, Resource{
			Type:        ResourceTypeAISearchNamespace,
			Name:        namespace.Binding,
			ID:          namespace.Namespace,
			Description: fmt.Sprintf("AI Search Namespace: %s", namespace.Namespace),
			URL:         BuildDashboardURL(accountID, namespaceURL, hasAccount),
		})
	}

	// VPC
	for _, service := range cfg.VPCServices {
		vpcURL := fmt.Sprintf("workers/vpc/services/%s", service.ServiceID)
		resources = append(resources, Resource{
			Type:        ResourceTypeVPC,
			Name:        service.Binding,
			ID:          service.ServiceID,
			Description: fmt.Sprintf("VPC Service: %s", service.ServiceID),
			URL:         BuildDashboardURL(accountID, vpcURL, hasAccount),
		})
	}

	// R2 Object Storage
	for _, bucket := range cfg.R2Buckets {
		jurisdiction := bucket.Jurisdiction
		if jurisdiction == "" {
			jurisdiction = "default"
		}
		r2URL := fmt.Sprintf("r2/%s/buckets/%s", jurisdiction, bucket.BucketName)
		resources = append(resources, Resource{
			Type:        ResourceTypeR2,
			Name:        bucket.Binding,
			ID:          bucket.BucketName,
			Description: fmt.Sprintf("R2: %s", bucket.BucketName),
			URL:         BuildDashboardURL(accountID, r2URL, hasAccount),
		})
	}

	// Workers KV
	for _, kv := range cfg.KVNamespaces {
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
	for _, db := range cfg.D1Databases {
		d1URL := fmt.Sprintf("workers/d1/databases/%s/metrics", db.DatabaseID)
		resources = append(resources, Resource{
			Type:        ResourceTypeD1,
			Name:        db.Binding,
			ID:          db.DatabaseID,
			Description: fmt.Sprintf("D1: %s (%s)", db.DatabaseName, db.DatabaseID),
			URL:         BuildDashboardURL(accountID, d1URL, hasAccount),
		})
	}

	// Hyperdrive
	for _, hyperdrive := range cfg.Hyperdrive {
		hyperdriveURL := fmt.Sprintf("workers/hyperdrive/%s", hyperdrive.ID)
		resources = append(resources, Resource{
			Type:        ResourceTypeHyperdrive,
			Name:        hyperdrive.Binding,
			ID:          hyperdrive.ID,
			Description: fmt.Sprintf("Hyperdrive: %s (%s)", hyperdrive.Binding, hyperdrive.ID),
			URL:         BuildDashboardURL(accountID, hyperdriveURL, hasAccount),
		})
	}

	// Pipelines
	for _, pipeline := range cfg.Pipelines {
		streamID := pipeline.Stream
		if streamID == "" {
			streamID = pipeline.Pipeline
		}
		streamURL := fmt.Sprintf("pipelines/streams/%s", streamID)
		resources = append(resources, Resource{
			Type:        ResourceTypePipeline,
			Name:        pipeline.Binding,
			ID:          streamID,
			Description: fmt.Sprintf("Stream: %s", streamID),
			URL:         BuildDashboardURL(accountID, streamURL, hasAccount),
		})
	}

	// Vectorize
	for _, vectorize := range cfg.Vectorize {
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
	for _, secret := range cfg.SecretsStoreSecrets {
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
	if cfg.Images != nil && cfg.Images.Binding != "" {
		imagesURL := "images"
		resources = append(resources, Resource{
			Type:        ResourceTypeImages,
			Name:        cfg.Images.Binding,
			ID:          "images",
			Description: "Images",
			URL:         BuildDashboardURL(accountID, imagesURL, hasAccount),
		})
	}

	return resources
}
