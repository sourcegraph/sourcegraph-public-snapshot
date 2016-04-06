// +build dist

package ui

import "log"

var (
	bundleJS []byte // never changes because it's compiled into the "dist"-tagged binary
)

func init() {
	var err error
	bundleJS, err = readBundleJS()
	if err != nil {
		log.Fatalf("Failed to read server bundle JS (for rendering React components): %s.", err)
	}
}

func getBundleJS() (js []byte, cacheKey string, err error) {
	return bundleJS, "_", nil
}
