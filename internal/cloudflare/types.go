package cloudflare

type ResourceType string

const (
	ResourceTypePages              ResourceType = "pages"
	ResourceTypeWorker             ResourceType = "worker"
	ResourceTypeObservability      ResourceType = "observability"
	ResourceTypeCronTriggers       ResourceType = "cron_triggers"
	ResourceTypeLogpush            ResourceType = "logpush"
	ResourceTypeRoutes             ResourceType = "routes"
	ResourceTypeWorkersRoutes      ResourceType = "workers_routes"
	ResourceTypeEmailRouting       ResourceType = "email_routing"
	ResourceTypeEmailRoutingDomain ResourceType = "email_routing_domain"
	ResourceTypeQueue              ResourceType = "queue"
	ResourceTypeWorkflow           ResourceType = "workflow"
	ResourceTypeDurableObjects     ResourceType = "durable_objects"
	ResourceTypeContainers         ResourceType = "containers"
	ResourceTypeBrowserRun         ResourceType = "browser_run"
	ResourceTypeWorkersAI          ResourceType = "workers_ai"
	ResourceTypeStream             ResourceType = "stream"
	ResourceTypeAISearch           ResourceType = "ai_search"
	ResourceTypeAISearchNamespace  ResourceType = "ai_search_namespace"
	ResourceTypeArtifacts          ResourceType = "artifacts"
	ResourceTypeAnalyticsEngine    ResourceType = "analytics_engine"
	ResourceTypeEmailSending       ResourceType = "email_sending"
	ResourceTypeVPC                ResourceType = "vpc"
	ResourceTypeVPCNetworks        ResourceType = "vpc_networks"
	ResourceTypeTunnel             ResourceType = "tunnel"
	ResourceTypeFlagship           ResourceType = "flagship"
	ResourceTypeService            ResourceType = "service"
	ResourceTypeTailWorker         ResourceType = "tail_worker"
	ResourceTypeR2                 ResourceType = "r2"
	ResourceTypeKV                 ResourceType = "kv"
	ResourceTypeD1                 ResourceType = "d1"
	ResourceTypeHyperdrive         ResourceType = "hyperdrive"
	ResourceTypePipeline           ResourceType = "pipeline"
	ResourceTypeK2                 ResourceType = "k2"
	ResourceTypeVectorize          ResourceType = "vectorize"
	ResourceTypeSecretsStore       ResourceType = "secrets_store"
	ResourceTypeImages             ResourceType = "images"
	ResourceTypeMedia              ResourceType = "media"
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
