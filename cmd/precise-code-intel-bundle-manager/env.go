package main

import (
	"log"
	"strconv"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

var (
	rawBundleDir            = env.Get("PRECISE_CODE_INTEL_BUNDLE_DIR", "/lsif-storage", "Root dir containing uploads and converted bundles.")
	rawDatabaseCacheSize    = env.Get("PRECISE_CODE_INTEL_CONNECTION_CACHE_CAPACITY", "100", "Maximum number of SQLite connections that can be open at once.")
	rawDocumentCacheSize    = env.Get("PRECISE_CODE_INTEL_DOCUMENT_CACHE_CAPACITY", "1000000", "Size of decoded document cache. A document's cost is the number of fields.")
	rawResultChunkCacheSize = env.Get("PRECISE_CODE_INTEL_RESULT_CHUNK_CACHE_CAPACITY", "1000000", "Size of decoded result chunk cache. A result chunk's cost is the number of fields.")
	rawDesiredPercentFree   = env.Get("PRECISE_CODE_INTEL_DESIRED_PERCENT_FREE", "10", "Target percentage of free space on disk.")
	rawJanitorInterval      = env.Get("PRECISE_CODE_INTEL_JANITOR_INTERVAL", "1m", "Interval between cleanup runs.")
	rawMaxUploadAge         = env.Get("PRECISE_CODE_INTEL_MAX_UPLOAD_AGE", "24h", "The maximum time an upload can sit on disk.")
)

// mustGet returns the non-empty version of the given raw value fatally logs on failure.
func mustGet(rawValue, name string) string {
	if rawValue == "" {
		log.Fatalf("invalid value %q for %s: no value supplied", rawValue, name)
	}

	return rawValue
}

// mustParseInt returns the integer version of the given raw value fatally logs on failure.
func mustParseInt(rawValue, name string) int {
	i, err := strconv.ParseInt(rawValue, 10, 64)
	if err != nil {
		log.Fatalf("invalid int %q for %s: %s", rawValue, name, err)
	}

	return int(i)
}

// mustParsePercent returns the integer percent (in range [0, 100]) version of the given raw
// value fatally logs on failure.
func mustParsePercent(rawValue, name string) int {
	p := mustParseInt(rawValue, name)
	if p < 0 || p > 100 {
		log.Fatalf("invalid percent %q for %s: must be 0 <= p <= 100", rawValue, name)
	}

	return p
}

// mustParseInterval returns the interval version of the given raw value fatally logs on failure.
func mustParseInterval(rawValue, name string) time.Duration {
	d, err := time.ParseDuration(rawValue)
	if err != nil {
		log.Fatalf("invalid duration %q for %s: %s", rawValue, name, err)
	}

	return d
}
