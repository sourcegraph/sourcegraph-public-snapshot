package ui

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
	"runtime"
	"strings"

	"github.com/coreos/go-semver/semver"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/internal/conf/deploy"
	"github.com/sourcegraph/sourcegraph/internal/version"
)

// serveHelp redirects to documentation pages on https://docs.sourcegraph.com for the current
// product version, i.e., /help/PATH -> https://docs.sourcegraph.com/@VERSION/PATH. In unreleased
// development builds (whose docs aren't necessarily available on https://docs.sourcegraph.com, it
// shows a message with instructions on how to see the docs.)
func serveHelp(w http.ResponseWriter, r *http.Request) {
	page := strings.TrimPrefix(r.URL.Path, "/help")
	versionStr := version.Version()
	sourcegraphAppMode := deploy.IsApp()

	// For release builds, use the version string. Otherwise, don't use any
	// version string because:
	//
	// - For unreleased dev builds, we serve the contents from the working tree.
	// - Sourcegraph.com users probably want the latest docs on the default
	//   branch.
	// - For Sourcegraph App users we also want to show the latest docs,
	//   but we add the app version as a query param.
	var docRevPrefix string
	if !version.IsDev(versionStr) && !envvar.SourcegraphDotComMode() && !sourcegraphAppMode {
		v, err := semver.NewVersion(versionStr)
		if err != nil {
			// If not a semver, just use the version string and hope for the best
			docRevPrefix = "@" + versionStr
		} else {
			// Otherwise, send viewer to the major.minor branch of this version
			docRevPrefix = fmt.Sprintf("@%d.%d", v.Major, v.Minor)
		}
	}

	// Note that the URI fragment (e.g., #some-section-in-doc) *should* be preserved by most user
	// agents even though the Location HTTP response header omits it. See
	// https://stackoverflow.com/a/2305927.
	dest := &url.URL{
		Path: path.Join("/", docRevPrefix, page),
	}
	if version.IsDev(versionStr) && !envvar.SourcegraphDotComMode() && !sourcegraphAppMode {
		dest.Scheme = "http"
		dest.Host = "localhost:5080" // local documentation server (defined in Procfile) -- CI:LOCALHOST_OK
	} else {
		dest.Scheme = "https"
		dest.Host = "docs.sourcegraph.com"
	}

	// For App, add UTM parameters to the docs url.
	if sourcegraphAppMode {
		q := dest.Query()
		q.Set("utm_source", "sg_app")
		q.Set("utm_medium", "referral")

		// App OS and version
		os := runtime.GOOS
		if os == "darwin" {
			// Use a more common name for mac because it'll be used for analytics.
			os = "mac"
		}

		q.Set("app_os", os)
		q.Set("app_version", versionStr)

		dest.RawQuery = q.Encode()
	}

	// Use temporary, not permanent, redirect, because the destination URL changes (depending on the
	// current product version).
	http.Redirect(w, r, dest.String(), http.StatusTemporaryRedirect)
}
