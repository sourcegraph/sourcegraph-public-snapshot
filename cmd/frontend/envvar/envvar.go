// Package envvar contains helpers for reading common environment variables.
package envvar

import (
	"strconv"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

var HTTPAddrInternal = env.Get(
	"SRC_HTTP_ADDR_INTERNAL",
	func() string {
		if env.InsecureDev {
			return "127.0.0.1:3090"
		}
		return "0.0.0.0:3090"
	}(),
	"HTTP listen address for internal HTTP API. This should never be exposed externally, as it lacks certain authz checks.",
)

var sourcegraphDotComMode, _ = strconv.ParseBool(env.Get("SOURCEGRAPHDOTCOM_MODE", "false", "run as Sourcegraph.com, with add'l marketing and redirects"))
var openGraphPreviewServiceURL = env.Get("OPENGRAPH_PREVIEW_SERVICE_URL", "", "The URL of the OpenGraph preview image generating service")

// SourcegraphDotComMode is true if this server is running Sourcegraph.com
// (solely by checking the SOURCEGRAPHDOTCOM_MODE env var). Sourcegraph.com shows
// additional marketing and sets up some additional redirects.
func SourcegraphDotComMode() bool {
	return sourcegraphDotComMode
}

// MockSourcegraphDotComMode is used by tests to mock the result of SourcegraphDotComMode.
func MockSourcegraphDotComMode(value bool) {
	sourcegraphDotComMode = value
}

func OpenGraphPreviewServiceURL() string {
	return openGraphPreviewServiceURL
}
