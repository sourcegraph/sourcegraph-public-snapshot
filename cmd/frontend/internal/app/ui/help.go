package ui

import (
	"net/http"
	"strings"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/docs"
)

// serveHelp redirects to documentation pages on https://sourcegraph.com/docs for the current
// product version, i.e., /help/PATH -> https://sourcegraph.com/docs/v/VERSION/PATH. In unreleased
// development builds (whose docs aren't necessarily available on https://sourcegraph.com/docs), it
// shows a message with instructions on how to see the docs.
func serveHelp(w http.ResponseWriter, r *http.Request) {
	page := strings.TrimPrefix(r.URL.Path, "/help")

	dest := docs.URL(page)
	log.Scoped("serveHelp").Info("redirecting to docs", log.String("url", dest.String()))

	// Use temporary, not permanent, redirect, because the destination URL changes (depending on the
	// current product version).
	http.Redirect(w, r, dest.String(), http.StatusTemporaryRedirect)
}
