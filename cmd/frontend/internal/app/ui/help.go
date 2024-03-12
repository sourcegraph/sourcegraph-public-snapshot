package ui

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/coreos/go-semver/semver"

	sglog "github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/dotcom"
	"github.com/sourcegraph/sourcegraph/internal/version"
)

// serveHelp redirects to documentation pages on https://sourcegraph.com/docs for the current
// product version, i.e., /help/PATH -> https://sourcegraph.com/docs/v/VERSION/PATH. In unreleased
// development builds (whose docs aren't necessarily available on https://sourcegraph.com/docs), it
// shows a message with instructions on how to see the docs.
func serveHelp(w http.ResponseWriter, r *http.Request) {
	page := strings.TrimPrefix(r.URL.Path, "/help")
	versionStr := version.Version()

	logger := sglog.Scoped("serveHelp")
	logger.Info("redirecting to docs", sglog.String("page", page), sglog.String("versionStr", versionStr))

	// For release builds, use the version string. Otherwise, don't use any
	// version string because:
	//
	// - For unreleased dev builds, we serve the contents from the working tree.
	// - Sourcegraph.com users probably want the latest docs on the default
	//   branch.
	var docRevPrefix string
	if !version.IsDev(versionStr) && !dotcom.SourcegraphDotComMode() {
		v, err := semver.NewVersion(versionStr)
		if err == nil {
			docRevPrefix = fmt.Sprintf("v/%d.%d", v.Major, v.Minor)
		}
		// In the case of an error, just redirect to the latest version of the docs
	}

	// Note that the URI fragment (e.g., #some-section-in-doc) *should* be preserved by most user
	// agents even though the Location HTTP response header omits it. See
	// https://stackoverflow.com/a/2305927.
	var dest *url.URL
	if version.IsDev(versionStr) && !dotcom.SourcegraphDotComMode() {
		dest = &url.URL{
			Scheme: "http",
			Host:   "localhost:3000", // local documentation server (defined in Procfile) -- CI:LOCALHOST_OK
			Path:   path.Join("/", docRevPrefix, page),
		}
	} else {
		dest = &url.URL{
			Scheme: "https",
			Host:   "sourcegraph.com",
			Path:   path.Join("/", "docs", docRevPrefix, page),
		}
	}

	// Use temporary, not permanent, redirect, because the destination URL changes (depending on the
	// current product version).
	http.Redirect(w, r, dest.String(), http.StatusTemporaryRedirect)
}
