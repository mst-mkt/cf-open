package internal

import (
	"fmt"

	"github.com/manifoldco/promptui"
	"github.com/mst-mkt/cf-open/internal/cloudflare"
)

func SelectResource(resources []cloudflare.Resource) (*cloudflare.Resource, error) {
	if len(resources) == 0 {
		return nil, fmt.Errorf("no resources found")
	}

	if len(resources) == 1 {
		return &resources[0], nil
	}

	items := make([]string, len(resources))
	for i, resource := range resources {
		items[i] = resource.Display()
	}

	prompt := promptui.Select{
		Label: "Select a resource to open",
		Items: items,
	}

	index, _, err := prompt.Run()
	if err != nil {
		return nil, fmt.Errorf("selection cancelled: %w", err)
	}

	return &resources[index], nil
}
