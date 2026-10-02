package cloudflare

import (
	"fmt"
	"strings"

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

	// Logpush
	if cfg.Logpush {
		logpushURL := "logs"
		resources = append(resources, Resource{
			Type:        ResourceTypeLogpush,
			Name:        "logpush",
			ID:          "logpush",
			Description: "Logpush",
			URL:         BuildDashboardURL(accountID, logpushURL, hasAccount),
		})
	}

	// Workers Routes
	routes := cfg.Routes
	if cfg.Route != nil {
		routes = append(routes, *cfg.Route)
	}

	if cfg.Name != "" && !isPages && len(routes) > 0 {
		triggersURL := fmt.Sprintf("workers/services/view/%s/production/triggers", cfg.Name)
		resources = append(resources, Resource{
			Type:        ResourceTypeRoutes,
			Name:        cfg.Name,
			ID:          cfg.Name,
			Description: fmt.Sprintf("Routes: %s", cfg.Name),
			URL:         BuildDashboardURL(accountID, triggersURL, hasAccount),
		})
	}

	seenZones := make(map[string]bool)
	for _, route := range routes {
		if route.ZoneName == "" || route.CustomDomain || seenZones[route.ZoneName] {
			continue
		}
		seenZones[route.ZoneName] = true

		routesURL := fmt.Sprintf("%s/workers", route.ZoneName)
		resources = append(resources, Resource{
			Type:        ResourceTypeWorkersRoutes,
			Name:        route.ZoneName,
			ID:          route.ZoneName,
			Description: fmt.Sprintf("Workers Routes: %s", route.ZoneName),
			URL:         BuildDashboardURL(accountID, routesURL, hasAccount),
		})
	}

	// Email Routing
	if len(cfg.Addresses) > 0 {
		emailRoutingURL := "email-service/routing"
		resources = append(resources, Resource{
			Type:        ResourceTypeEmailRouting,
			Name:        "email-routing",
			ID:          "email-routing",
			Description: "Email Routing",
			URL:         BuildDashboardURL(accountID, emailRoutingURL, hasAccount),
		})
	}

	seenDomains := make(map[string]bool)
	for _, address := range cfg.Addresses {
		_, domain, ok := strings.Cut(address, "@")
		if !ok || domain == "" || seenDomains[domain] {
			continue
		}
		seenDomains[domain] = true

		domainURL := fmt.Sprintf("%s/email/routing", domain)
		resources = append(resources, Resource{
			Type:        ResourceTypeEmailRoutingDomain,
			Name:        domain,
			ID:          domain,
			Description: fmt.Sprintf("Email Routing: %s", domain),
			URL:         BuildDashboardURL(accountID, domainURL, hasAccount),
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

	// Containers
	// The detail page is addressed by application ID, not the name in the config, so open the list.
	if len(cfg.Containers) > 0 {
		containersURL := "workers/containers"
		resources = append(resources, Resource{
			Type:        ResourceTypeContainers,
			Name:        "containers",
			ID:          "containers",
			Description: "Containers",
			URL:         BuildDashboardURL(accountID, containersURL, hasAccount),
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

	// Stream
	if cfg.Stream != nil && cfg.Stream.Binding != "" {
		streamURL := "stream/videos"
		resources = append(resources, Resource{
			Type:        ResourceTypeStream,
			Name:        cfg.Stream.Binding,
			ID:          "stream",
			Description: "Stream",
			URL:         BuildDashboardURL(accountID, streamURL, hasAccount),
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

	// Artifacts
	for _, artifacts := range cfg.Artifacts {
		artifactsURL := fmt.Sprintf("workers/artifacts/namespaces/%s", artifacts.Namespace)
		resources = append(resources, Resource{
			Type:        ResourceTypeArtifacts,
			Name:        artifacts.Binding,
			ID:          artifacts.Namespace,
			Description: fmt.Sprintf("Artifacts: %s", artifacts.Namespace),
			URL:         BuildDashboardURL(accountID, artifactsURL, hasAccount),
		})
	}

	// Analytics Engine
	// Datasets have no page of their own, so open the list.
	if len(cfg.AnalyticsEngineDatasets) > 0 {
		analyticsEngineURL := "workers/analytics-engine"
		resources = append(resources, Resource{
			Type:        ResourceTypeAnalyticsEngine,
			Name:        "analytics-engine",
			ID:          "analytics-engine",
			Description: "Analytics Engine",
			URL:         BuildDashboardURL(accountID, analyticsEngineURL, hasAccount),
		})
	}

	// Email Sending
	// The detail page is addressed by zone and domain IDs, which the config does not have, so open the list.
	if len(cfg.SendEmail) > 0 {
		sendEmailURL := "email-service/sending"
		resources = append(resources, Resource{
			Type:        ResourceTypeEmailSending,
			Name:        "email-sending",
			ID:          "email-sending",
			Description: "Email Sending",
			URL:         BuildDashboardURL(accountID, sendEmailURL, hasAccount),
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

	// VPC Networks
	if len(cfg.VPCNetworks) > 0 {
		vpcNetworksURL := "workers/vpc/networks"
		resources = append(resources, Resource{
			Type:        ResourceTypeVPCNetworks,
			Name:        "vpc-networks",
			ID:          "vpc-networks",
			Description: "VPC Networks",
			URL:         BuildDashboardURL(accountID, vpcNetworksURL, hasAccount),
		})
	}

	// Tunnels
	for _, network := range cfg.VPCNetworks {
		if network.TunnelID == "" {
			continue
		}

		tunnelURL := fmt.Sprintf("tunnels/%s/overview", network.TunnelID)
		resources = append(resources, Resource{
			Type:        ResourceTypeTunnel,
			Name:        network.Binding,
			ID:          network.TunnelID,
			Description: fmt.Sprintf("Tunnel: %s", network.TunnelID),
			URL:         BuildDashboardURL(accountID, tunnelURL, hasAccount),
		})
	}

	// Flagship
	for _, flagship := range cfg.Flagship {
		flagshipURL := "flagship"
		description := "Flagship"
		if flagship.AppID != "" {
			flagshipURL = fmt.Sprintf("flagship/applications/%s/overview", flagship.AppID)
			description = fmt.Sprintf("Flagship: %s", flagship.AppID)
		}

		resources = append(resources, Resource{
			Type:        ResourceTypeFlagship,
			Name:        flagship.Binding,
			ID:          flagship.AppID,
			Description: description,
			URL:         BuildDashboardURL(accountID, flagshipURL, hasAccount),
		})
	}

	// Service Bindings
	seenServices := make(map[string]bool)
	for _, service := range cfg.Services {
		if service.Service == "" || seenServices[service.Service] {
			continue
		}
		seenServices[service.Service] = true

		serviceURL := fmt.Sprintf("workers/services/view/%s/production", service.Service)
		resources = append(resources, Resource{
			Type:        ResourceTypeService,
			Name:        service.Binding,
			ID:          service.Service,
			Description: fmt.Sprintf("Service: %s", service.Service),
			URL:         BuildDashboardURL(accountID, serviceURL, hasAccount),
		})
	}

	// Tail Workers
	seenTailWorkers := make(map[string]bool)
	for _, consumer := range append(cfg.TailConsumers, cfg.StreamingTailConsumers...) {
		if consumer.Service == "" || seenTailWorkers[consumer.Service] {
			continue
		}
		seenTailWorkers[consumer.Service] = true

		tailWorkerURL := fmt.Sprintf("workers/services/view/%s/production", consumer.Service)
		resources = append(resources, Resource{
			Type:        ResourceTypeTailWorker,
			Name:        consumer.Service,
			ID:          consumer.Service,
			Description: fmt.Sprintf("Tail Worker: %s", consumer.Service),
			URL:         BuildDashboardURL(accountID, tailWorkerURL, hasAccount),
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

	// K2
	for _, k2 := range cfg.K2 {
		k2URL := fmt.Sprintf("k2/%s", k2.Stream)
		resources = append(resources, Resource{
			Type:        ResourceTypeK2,
			Name:        k2.Binding,
			ID:          k2.Stream,
			Description: fmt.Sprintf("K2: %s", k2.Stream),
			URL:         BuildDashboardURL(accountID, k2URL, hasAccount),
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
		imagesURL := "images/hosted"
		resources = append(resources, Resource{
			Type:        ResourceTypeImages,
			Name:        cfg.Images.Binding,
			ID:          "images",
			Description: "Images",
			URL:         BuildDashboardURL(accountID, imagesURL, hasAccount),
		})
	}

	// Media
	if cfg.Media != nil && cfg.Media.Binding != "" {
		mediaURL := "media/transformations"
		resources = append(resources, Resource{
			Type:        ResourceTypeMedia,
			Name:        cfg.Media.Binding,
			ID:          "media",
			Description: "Media",
			URL:         BuildDashboardURL(accountID, mediaURL, hasAccount),
		})
	}

	return resources
}
