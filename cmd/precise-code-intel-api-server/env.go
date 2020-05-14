package main

import (
	"log"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

var (
	rawBundleManagerURL = env.Get("PRECISE_CODE_INTEL_BUNDLE_MANAGER_URL", "", "HTTP address for internal LSIF bundle manager server.")
)

// mustGet returns the non-empty version of the given raw value fatally logs on failure.
func mustGet(rawValue, name string) string {
	if rawValue == "" {
		log.Fatalf("invalid value %q for %s: no value supplied", rawValue, name)
	}

	return rawValue
}
