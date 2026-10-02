package cloudflare

type ResourceType string

const (
	ResourceTypePages             ResourceType = "pages"
	ResourceTypeWorker            ResourceType = "worker"
	ResourceTypeObservability     ResourceType = "observability"
	ResourceTypeCronTriggers      ResourceType = "cron_triggers"
	ResourceTypeRoutes            ResourceType = "routes"
	ResourceTypeWorkersRoutes     ResourceType = "workers_routes"
	ResourceTypeQueue             ResourceType = "queue"
	ResourceTypeWorkflow          ResourceType = "workflow"
	ResourceTypeDurableObjects    ResourceType = "durable_objects"
	ResourceTypeContainers        ResourceType = "containers"
	ResourceTypeBrowserRun        ResourceType = "browser_run"
	ResourceTypeWorkersAI         ResourceType = "workers_ai"
	ResourceTypeStream            ResourceType = "stream"
	ResourceTypeAISearch          ResourceType = "ai_search"
	ResourceTypeAISearchNamespace ResourceType = "ai_search_namespace"
	ResourceTypeAnalyticsEngine   ResourceType = "analytics_engine"
	ResourceTypeEmailSending      ResourceType = "email_sending"
	ResourceTypeVPC               ResourceType = "vpc"
	ResourceTypeVPCNetworks       ResourceType = "vpc_networks"
	ResourceTypeTunnel            ResourceType = "tunnel"
	ResourceTypeFlagship          ResourceType = "flagship"
	ResourceTypeR2                ResourceType = "r2"
	ResourceTypeKV                ResourceType = "kv"
	ResourceTypeD1                ResourceType = "d1"
	ResourceTypeHyperdrive        ResourceType = "hyperdrive"
	ResourceTypePipeline          ResourceType = "pipeline"
	ResourceTypeVectorize         ResourceType = "vectorize"
	ResourceTypeSecretsStore      ResourceType = "secrets_store"
	ResourceTypeImages            ResourceType = "images"
)

type Resource struct {
	Type        ResourceType
	Name        string
	ID          string
	Description string
	URL         string
}

func (r Resource) Display() string {
	if r.Description != "" {
		return r.Description
	}
	return r.Name
}
