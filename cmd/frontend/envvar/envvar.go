package envvar

import (
	"strconv"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"

	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/env"
)

var sourcegraphDotComMode, _ = strconv.ParseBool(env.Get("SOURCEGRAPHDOTCOM_MODE", "false", "run as Sourcegraph.com, with add'l marketing and redirects"))

// SourcegraphDotComMode is true if this server is running Sourcegraph.com. It shows
// add'l marketing and sets up some add'l redirects.
func SourcegraphDotComMode() bool {
	u := globals.AppURL.String()
	return sourcegraphDotComMode || u == "https://sourcegraph.com" || u == "https://sourcegraph.com/"
}

// HasCodeIntelligence reports whether the site has enabled code intelligence.
func HasCodeIntelligence() bool {
	return len(conf.EnabledLangservers()) > 0
}
