package cloudflare

type ResourceType string

const (
	ResourceTypePages          ResourceType = "pages"
	ResourceTypeWorker         ResourceType = "worker"
	ResourceTypeObservability  ResourceType = "observability"
	ResourceTypeCronTriggers   ResourceType = "cron_triggers"
	ResourceTypeQueue          ResourceType = "queue"
	ResourceTypeWorkflow       ResourceType = "workflow"
	ResourceTypeDurableObjects ResourceType = "durable_objects"
	ResourceTypeBrowserRun     ResourceType = "browser_run"
	ResourceTypeVPC            ResourceType = "vpc"
	ResourceTypeR2             ResourceType = "r2"
	ResourceTypeKV             ResourceType = "kv"
	ResourceTypeD1             ResourceType = "d1"
	ResourceTypeHyperdrive     ResourceType = "hyperdrive"
	ResourceTypePipeline       ResourceType = "pipeline"
	ResourceTypeVectorize      ResourceType = "vectorize"
	ResourceTypeSecretsStore   ResourceType = "secrets_store"
	ResourceTypeImages         ResourceType = "images"
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
