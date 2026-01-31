package internal

import (
	"fmt"

	"github.com/pkg/browser"
)

func OpenURL(url string) error {
	fmt.Printf("Opening %s\n", url)
	return browser.OpenURL(url)
}
