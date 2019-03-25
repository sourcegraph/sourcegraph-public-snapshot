// Package envvar contains helpers for reading common environment variables.
package envvar

import (
	"strconv"

	"github.com/sourcegraph/sourcegraph/pkg/env"
)

var sourcegraphDotComMode, _ = strconv.ParseBool(env.Get("SOURCEGRAPHDOTCOM_MODE", "false", "run as Sourcegraph.com, with add'l marketing and redirects"))
var e2eTest, _ = strconv.ParseBool(env.Get("E2E_TEST", "false", "whether the Sourcegraph instance is being run as an end-to-end test"))

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

// E2eTestMode indicates whether the Sourcegraph instance is being run as an end-to-end test.
func E2eTestMode() bool {
	return e2eTest
}
