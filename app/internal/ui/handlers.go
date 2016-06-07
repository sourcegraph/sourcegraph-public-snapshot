// Package ui contains Go code to route, preload, and SEO the
// JavaScript (React) UI.
package ui

import (
	"encoding/json"
	"net/http"

	"sourcegraph.com/sourcegraph/sourcegraph/app/internal"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/tmpl"
	"sourcegraph.com/sourcegraph/sourcegraph/app/router"
)

func init() {
	internal.Handlers[router.UI] = internal.Handler(serve)
}

func serve(w http.ResponseWriter, r *http.Request) error {
	return tmpl.Exec(r, w, "ui.html", http.StatusOK, nil, &struct {
		tmpl.Common
		Stores *json.RawMessage
	}{})
}
