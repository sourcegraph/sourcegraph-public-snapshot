// +build dist

package ui

import "log"

var (
	bundleJS string // never changes because it's compiled into the "dist"-tagged binary
)

func init() {
	bundleJSBytes, err := readBundleJS()
	if err != nil {
		log.Fatalf("Failed to read bundle.js (for rendering React components): %s.", err)
	}
	bundleJS = string(bundleJSBytes)
}

func getBundleJS() (js, cacheKey string, err error) {
	return bundleJS, "_", nil
}
