package envvar

import (
	"os"
	"strconv"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/globals"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
)

var sourcegraphDotComMode, _ = strconv.ParseBool(env.Get("SOURCEGRAPHDOTCOM_MODE", "false", "run as Sourcegraph.com, with add'l marketing and redirects"))

// SourcegraphDotComMode is true if this server is running Sourcegraph.com. It shows
// add'l marketing and sets up some add'l redirects.
func SourcegraphDotComMode() bool {
	u := globals.AppURL.String()
	return sourcegraphDotComMode || u == "https://sourcegraph.com" || u == "https://sourcegraph.com/"
}

var debugMode, _ = strconv.ParseBool(env.Get("DEBUG", "false", "debug mode"))

// DebugMode is true if and only if the application is running in debug mode. In
// this mode, the application should display more verbose and informative errors
// in the UI. It should also show all features (as possible). Debug should NEVER
// be true in production.
func DebugMode() bool { return debugMode }

// HasCodeIntelligence reports whether the site has enabled code intelligence.
func HasCodeIntelligence() bool {
	addr := os.Getenv("LSP_PROXY")
	if addr == "" {
		return false
	}
	return len(conf.Get().Langservers) > 0
}
