package docs

import (
	"fmt"
	"net/url"
	"path"

	"github.com/coreos/go-semver/semver"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/internal/version"
)

// URL returns the docs URL for the given docs page. This should be used instead of manually constructing a docs URL
// because it points to the docs for the current version of the instance.
func URL(page string) *url.URL {
	versionStr := version.Version()

	// For release builds, use the version string. Otherwise, don't use any
	// version string because:
	//
	// - For unreleased dev builds, we serve the contents from the working tree.
	// - Sourcegraph.com users probably want the latest docs on the default
	//   branch.
	var docRevPrefix string
	if !version.IsDev(versionStr) && !envvar.SourcegraphDotComMode() {
		v, err := semver.NewVersion(versionStr)
		if err == nil {
			docRevPrefix = fmt.Sprintf("v/%d.%d", v.Major, v.Minor)
		}
		// In the case of an error, just redirect to the latest version of the docs
	}

	// Note that the URI fragment (e.g., #some-section-in-doc) *should* be preserved by most user
	// agents even though the Location HTTP response header omits it. See
	// https://stackoverflow.com/a/2305927.
	if version.IsDev(versionStr) && !envvar.SourcegraphDotComMode() {
		return &url.URL{
			Scheme: "http",
			Host:   "localhost:5080", // local documentation server (defined in Procfile) -- CI:LOCALHOST_OK
			Path:   path.Join("/", docRevPrefix, page),
		}
	} else {
		return &url.URL{
			Scheme: "https",
			Host:   "sourcegraph.com",
			Path:   path.Join("/", "docs", docRevPrefix, page),
		}
	}
}
