package app

import (
	"encoding/json"
	"html/template"
	"net/http"

	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/tmpl"
)

func serveUI(w http.ResponseWriter, r *http.Request) error {
	var header http.Header
	var data struct {
		tmpl.Common
		Body   template.HTML
		Stores *json.RawMessage
	}

	return tmpl.Exec(r, w, "ui.html", http.StatusOK, header, &data)
}
