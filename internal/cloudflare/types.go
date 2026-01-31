package cloudflare

type ResourceType string

const (
	ResourceTypeWorker ResourceType = "worker"
	ResourceTypeKV     ResourceType = "kv"
	ResourceTypeD1     ResourceType = "d1"
	ResourceTypeR2     ResourceType = "r2"
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