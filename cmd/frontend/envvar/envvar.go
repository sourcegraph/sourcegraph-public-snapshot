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

var openGraphPreviewServiceURL = env.Get("OPENGRAPH_PREVIEW_SERVICE_URL", "", "The URL of the OpenGraph preview image generating service")
var extsvcConfigFile = env.Get("EXTSVC_CONFIG_FILE", "", "EXTSVC_CONFIG_FILE can contain configurations for multiple code host connections. See https://sourcegraph.com/docs/admin/config/advanced_config_file for details.")
var extsvcConfigAllowEdits, _ = strconv.ParseBool(env.Get("EXTSVC_CONFIG_ALLOW_EDITS", "false", "When EXTSVC_CONFIG_FILE is in use, allow edits in the application to be made which will be overwritten on next process restart"))

func OpenGraphPreviewServiceURL() string {
	return openGraphPreviewServiceURL
}

// ExtsvcConfigFile returns value of EXTSVC_CONFIG_FILE environment variable
func ExtsvcConfigFile() string {
	return extsvcConfigFile
}

// ExtsvcConfigAllowEdits returns boolean value of EXTSVC_CONFIG_ALLOW_EDITS
// environment variable.
func ExtsvcConfigAllowEdits() bool {
	return extsvcConfigAllowEdits
}
