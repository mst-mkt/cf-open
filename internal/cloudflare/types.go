package cloudflare

type ResourceType string

const (
	ResourceTypeWorker           ResourceType = "worker"
	ResourceTypeObservability    ResourceType = "observability"
	ResourceTypeCronTriggers     ResourceType = "cron_triggers"
	ResourceTypeQueue            ResourceType = "queue"
	ResourceTypeWorkflow         ResourceType = "workflow"
	ResourceTypeBrowserRendering ResourceType = "browser_rendering"
	ResourceTypeVPC              ResourceType = "vpc"
	ResourceTypeR2               ResourceType = "r2"
	ResourceTypeKV               ResourceType = "kv"
	ResourceTypeD1               ResourceType = "d1"
	ResourceTypePipeline         ResourceType = "pipeline"
	ResourceTypeVectorize        ResourceType = "vectorize"
	ResourceTypeSecretsStore     ResourceType = "secrets_store"
	ResourceTypeImages           ResourceType = "images"
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
