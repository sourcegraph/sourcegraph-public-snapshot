package assets

import (
	"net/http"
	"path/filepath"

	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/gzipfileserver"
)

func AssetFS() http.Handler {
	fs := gzipfileserver.New(Assets)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Kludge to set proper MIME type. Automatic MIME detection somehow detects text/xml under
		// circumstances that couldn't be reproduced
		if filepath.Ext(r.URL.Path) == ".svg" {
			w.Header().Set("Content-Type", "image/svg+xml")
		}
		w.Header().Set("Cache-Control", "max-age=31556926, public")
		fs.ServeHTTP(w, r)
	})
}
