package main

import (
	"log"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

var (
	rawBundleManagerURL = env.Get("PRECISE_CODE_INTEL_BUNDLE_MANAGER_URL", "", "HTTP address for internal LSIF bundle manager server.")
	rawPollInterval     = env.Get("PRECISE_CODE_INTEL_POLL_INTERVAL", "1s", "Interval between queries to the work queue.")
)

// mustGet returns the non-empty version of the given raw value fatally logs on failure.
func mustGet(rawValue, name string) string {
	if rawValue == "" {
		log.Fatalf("invalid value %q for %s: no value supplied", rawValue, name)
	}

	return rawValue
}

// mustParseInterval returns the interval version of the given raw value fatally logs on failure.
func mustParseInterval(rawValue, name string) time.Duration {
	d, err := time.ParseDuration(rawValue)
	if err != nil {
		log.Fatalf("invalid duration %q for %s: %s", rawValue, name, err)
	}

	return d
}
