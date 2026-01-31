package internal

import (
	"fmt"

	"github.com/pkg/browser"
)

func OpenURL(url string) error {
	fmt.Printf("Opening %s\n", url)
	return browser.OpenURL(url)
}

func OpenURLs(urls []string) error {
	for _, url := range urls {
		if err := OpenURL(url); err != nil {
			return err
		}
	}
	return nil
}
