// +build dist

package assets

import (
	"net/http"
	"net/url"
	"path/filepath"

	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/gzipfileserver"
)

// Mount mounts the static asset handler.
func Mount(mux *http.ServeMux) {
	const urlPathPrefix = "/.assets"
	fs := gzipfileserver.New(Assets)
	mux.Handle(urlPathPrefix+"/", http.StripPrefix(urlPathPrefix, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Kludge to set proper MIME type. Automatic MIME detection somehow detects text/xml under
		// circumstances that couldn't be reproduced
		if filepath.Ext(r.URL.Path) == ".svg" {
			w.Header().Set("Content-Type", "image/svg+xml")
		}
		w.Header().Set("Cache-Control", "max-age=31556926, public")
		fs.ServeHTTP(w, r)
	})))
}

func init() {
	baseURL = &url.URL{Path: "/.assets"}
}
