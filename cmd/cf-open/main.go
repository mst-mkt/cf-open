package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/mst-mkt/cf-open/internal"
	"github.com/mst-mkt/cf-open/internal/cloudflare"
	"github.com/mst-mkt/cf-open/internal/config"
)

var (
	wranglerConfigPath string
	accountID          string
	openAll            bool
	printOnly          bool
)

var version = "dev"

var rootCmd = &cobra.Command{
	Use:     "cf-open",
	Short:   "Open Cloudflare dashboard for your project from CLI",
	Version: version,
	RunE: func(cmd *cobra.Command, args []string) error {
		wranglerConfig, err := config.LoadWranglerConfig(wranglerConfigPath)
		if err != nil {
			return fmt.Errorf("failed to load wrangler config: %w", err)
		}

		resolvedAccountID, hasAccount := config.GetAccountID(wranglerConfig, accountID)

		resources := cloudflare.GetResourcesFromConfig(wranglerConfig, resolvedAccountID, hasAccount)
		if len(resources) == 0 {
			return fmt.Errorf("no resources found in wrangler config")
		}

		if openAll {
			urls := make([]string, len(resources))
			for i, r := range resources {
				urls[i] = r.URL
			}
			if printOnly {
				for _, url := range urls {
					fmt.Println(url)
				}
				return nil
			}
			return internal.OpenURLs(urls)
		}

		selectedResource, err := internal.SelectResource(resources)
		if err != nil {
			return fmt.Errorf("failed to select resource: %w", err)
		}

		if printOnly {
			fmt.Println(selectedResource.URL)
			return nil
		}

		return internal.OpenURL(selectedResource.URL)
	},
}

func init() {
	rootCmd.Flags().StringVar(&wranglerConfigPath, "wrangler-config", "", "Path to wrangler configuration file")
	rootCmd.Flags().StringVar(&accountID, "account-id", "", "Cloudflare account ID")
	rootCmd.Flags().BoolVarP(&openAll, "all", "a", false, "Open all resources in the browser")
	rootCmd.Flags().BoolVarP(&printOnly, "print", "p", false, "Print URL to stdout instead of opening in browser")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
