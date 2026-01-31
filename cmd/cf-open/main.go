package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/mst-mkt/cf-open/internal"
	"github.com/mst-mkt/cf-open/internal/cloudflare"
	"github.com/mst-mkt/cf-open/internal/config"
)

var version = "dev"

type options struct {
	wranglerConfig string
	accountID      string
	all            bool
	print          bool
}

var opts options

var rootCmd = &cobra.Command{
	Use:     "cf-open",
	Short:   "Open Cloudflare dashboard for your project from CLI",
	Version: version,
	RunE: func(cmd *cobra.Command, args []string) error {
		return run(opts)
	},
}

func run(opts options) error {
	wranglerConfig, err := config.LoadWranglerConfig(opts.wranglerConfig)
	if err != nil {
		return fmt.Errorf("failed to load wrangler config: %w", err)
	}

	accountID, hasAccount := config.GetAccountID(wranglerConfig, opts.accountID)

	resources := cloudflare.GetResourcesFromConfig(wranglerConfig, accountID, hasAccount)
	if len(resources) == 0 {
		return fmt.Errorf("no resources found in wrangler config")
	}

	urls, err := selectURLs(resources, opts.all)
	if err != nil {
		return err
	}

	return outputURLs(urls, opts.print)
}

func selectURLs(resources []cloudflare.Resource, all bool) ([]string, error) {
	// `--all` が指定された場合はすべてのリソースの URL を返す
	if all {
		urls := make([]string, len(resources))
		for i, r := range resources {
			urls[i] = r.URL
		}
		return urls, nil
	}

	// 通常はユーザーに選択させそのリソースの URL を返す
	selected, err := internal.SelectResource(resources)
	if err != nil {
		return nil, fmt.Errorf("failed to select resource: %w", err)
	}
	return []string{selected.URL}, nil
}

func outputURLs(urls []string, printOnly bool) error {
	// `--print` が指定された場合は URL を標準出力に出力する
	if printOnly {
		for _, url := range urls {
			fmt.Println(url)
		}
		return nil
	}

	return internal.OpenURLs(urls)
}

func init() {
	rootCmd.Flags().StringVar(&opts.wranglerConfig, "wrangler-config", "", "Path to wrangler configuration file")
	rootCmd.Flags().StringVar(&opts.accountID, "account-id", "", "Cloudflare account ID")
	rootCmd.Flags().BoolVarP(&opts.all, "all", "a", false, "Open all resources in the browser")
	rootCmd.Flags().BoolVarP(&opts.print, "print", "p", false, "Print URL to stdout instead of opening in browser")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
