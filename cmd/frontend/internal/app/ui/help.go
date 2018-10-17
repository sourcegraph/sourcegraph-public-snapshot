package ui

import (
	"net/http"

	"github.com/sourcegraph/docsite"
	frontenddocsite "github.com/sourcegraph/sourcegraph/cmd/frontend/docsite"
)

var (
	docsiteHandler  = frontenddocsite.Site.Handler()
	docsitePassThru = serveBasicPageString("Help - Sourcegraph")
)

func serveHelp(w http.ResponseWriter, r *http.Request) {
	if docsite.IsContentAsset(r.URL.Path) {
		// Serve content assets (such as images referenced in .md files).
		docsiteHandler.ServeHTTP(w, r)
	} else {
		// The UI renders content pages (.md files).
		docsitePassThru(w, r)
	}
}
