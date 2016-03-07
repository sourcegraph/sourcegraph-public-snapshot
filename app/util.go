package app

import (
	"html/template"
	"net/http"

	"github.com/microcosm-cc/bluemonday"
)

func sanitizeHTML(html string) template.HTML {
	return template.HTML(bluemonday.UGCPolicy().Sanitize(html))
}

func copyRequest(r *http.Request) *http.Request {
	rCopy := *r
	urlCopy := *r.URL
	rCopy.URL = &urlCopy
	return &rCopy
}
