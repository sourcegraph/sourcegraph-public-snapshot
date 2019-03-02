package ui

import (
	"fmt"
	"html"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/pkg/version"
)

// serveHelp redirects to documentation pages on https://docs.sourcegraph.com for the current
// product version, i.e., /help/PATH -> https://docs.sourcegraph.com/@VERSION/PATH. In unreleased
// development builds (whose docs aren't necessarily available on https://docs.sourcegraph.com, it
// shows a message with instructions on how to see the docs.)
func serveHelp(w http.ResponseWriter, r *http.Request) {
	page := strings.TrimPrefix(r.URL.Path, "/help")
	versionStr := version.Version()

	// For release builds, use the version string. Otherwise, don't use any version string because:
	//
	// - For unreleased dev builds, there is no version.
	// - Sourcegraph.com users probably want the latest docs on the default branch.
	var docRevPrefix string
	if !version.IsDev(versionStr) && !envvar.SourcegraphDotComMode() {
		docRevPrefix = "@v" + versionStr
	}

	// Note that the URI fragment (e.g., #some-section-in-doc) *should* be preserved by most user
	// agents even though the Location HTTP response header omits it. See
	// https://stackoverflow.com/a/2305927.
	dest := &url.URL{
		Scheme: "https",
		Host:   "docs.sourcegraph.com",
		Path:   path.Join("/", docRevPrefix, page),
	}

	if version.IsDev(versionStr) && !envvar.SourcegraphDotComMode() {
		// Local changes to the docs (which is what devs likely expect to see) aren't visible on
		// https://docs.sourcegraph.com, so we can't redirect there.
		//
		// TODO(sqs): Make this redirect to http://localhost:5080 and run docsite locally there (in
		// the Procfile). This also requires moving docs.sourcegraph.com resources into the
		// sourcegraph repository.
		w.WriteHeader(http.StatusNotImplemented)
		fmt.Fprintln(w, "<p>Unable to redirect to published documentation page for an unreleased development build (because your local changes may not be committed and pushed).</p>")
		fmt.Fprintln(w)
		fmt.Fprintln(w, "<p>To read the published documentation page (if it exists) for the default branch, visit:</p>")
		fmt.Fprintln(w)
		escapedDest := html.EscapeString(dest.String())
		fmt.Fprintf(w, "<blockquote><a href=\"%s\" target=\"_blank\">%s</a></blockquote>", escapedDest, escapedDest)
		return
	}

	// Use temporary, not permanent, redirect, because the destination URL changes (depending on the
	// current product version).
	http.Redirect(w, r, dest.String(), http.StatusTemporaryRedirect)
}
