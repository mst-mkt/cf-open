package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/mst-mkt/cf-open/internal"
	"github.com/mst-mkt/cf-open/internal/cloudflare"
	"github.com/mst-mkt/cf-open/internal/config"
)

var wranglerConfigPath string

var rootCmd = &cobra.Command{
	Use: "cf-open",
	RunE: func(cmd *cobra.Command, args []string) error {
		wranglerConfig, err := config.LoadWranglerConfig(wranglerConfigPath)
		if err != nil {
			return fmt.Errorf("failed to load wrangler config: %w", err)
		}

		accountID, hasAccount := config.GetAccountID(wranglerConfig)

		resources := cloudflare.GetResourcesFromConfig(wranglerConfig, accountID, hasAccount)
		if len(resources) == 0 {
			return fmt.Errorf("no resources found in wrangler config")
		}

		selectedResource, err := internal.SelectResource(resources)
		if err != nil {
			return fmt.Errorf("failed to select resource: %w", err)
		}

		return internal.OpenURL(selectedResource.URL)
	},
}

func init() {
	rootCmd.Flags().StringVar(&wranglerConfigPath, "wrangler-config", "", "Path to wrangler configuration file")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
