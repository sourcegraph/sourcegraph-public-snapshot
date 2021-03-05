// Package envvar contains helpers for reading common environment variables.
package envvar

import (
	"strconv"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

var HTTPAddrInternal = env.Get(
	"SRC_HTTP_ADDR_INTERNAL",
	"0.0.0.0:3090",
	"HTTP listen address for internal HTTP API. This should never be exposed externally, as it lacks certain authz checks.",
)

var sourcegraphDotComMode, _ = strconv.ParseBool(env.Get("SOURCEGRAPHDOTCOM_MODE", "false", "run as Sourcegraph.com, with add'l marketing and redirects"))
var disableProfiler, _ = strconv.ParseBool(env.Get("SRC_DISABLE_PROFILER", "false", "Disable the gcloud profiler, for use when running locally without gcloud config"))

// SourcegraphDotComMode is true if this server is running Sourcegraph.com (solely by checking the
// SOURCEGRAPHDOTCOM_MODE env var). Sourcegraph.com shows add'l marketing and sets up some add'l
// redirects.
func SourcegraphDotComMode() bool {
	return sourcegraphDotComMode
}

// MockSourcegraphDotComMode is used by tests to mock the result of SourcegraphDotComMode.
func MockSourcegraphDotComMode(value bool) {
	sourcegraphDotComMode = value
}

func DisableProfiler() bool {
	return disableProfiler
}
