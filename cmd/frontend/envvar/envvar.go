// Package envvar contains helpers for reading common environment variables.
package envvar

import (
	"strconv"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

var sourcegraphDotComMode, _ = strconv.ParseBool(env.Get("SOURCEGRAPHDOTCOM_MODE", "false", "run as Sourcegraph.com, with add'l marketing and redirects"))

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
