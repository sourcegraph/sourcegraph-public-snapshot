package main

import (
	"fmt"
	"log"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

var (
	envPrefix           = "PRECISE_CODE_INTEL"
	rawBundleManagerURL = envGet("BUNDLE_MANAGER_URL", "", "HTTP address for internal LSIF bundle manager server.")
	rawJanitorInterval  = envGet("JANITOR_INTERVAL", "1m", "Interval between cleanup runs.")
)

// envGet is like env.Get but prefixes all envvars
func envGet(name, defaultValue, description string) string {
	return env.Get(fmt.Sprintf("%s_%s", envPrefix, name), defaultValue, description)
}

// mustGet returns the non-empty version of the given raw value fatally logs on failure.
func mustGet(rawValue, name string) string {
	if rawValue == "" {
		log.Fatalf("invalid value %q for %s_%s: no value supplied", rawValue, envPrefix, name)
	}

	return rawValue
}

// mustParseInterval returns the interval version of the given raw value fatally logs on failure.
func mustParseInterval(rawValue, name string) time.Duration {
	d, err := time.ParseDuration(rawValue)
	if err != nil {
		log.Fatalf("invalid duration %q for %s_%s: %s", rawValue, envPrefix, name, err)
	}

	return d
}
