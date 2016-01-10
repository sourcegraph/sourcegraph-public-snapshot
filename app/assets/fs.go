package assets

import (
	"net/http"
	"path/filepath"

	"src.sourcegraph.com/sourcegraph/app/internal/gzipfileserver"
)

type cacheStrategy uint

const (
	// shortTermCache is a caching strategy that informs a client to
	// cache for a short amount of time
	shortTermCache cacheStrategy = iota

	// longTermCache is a caching strategy that informs a client to cache
	// for a long time
	longTermCache
)

func AssetFS(cacheStrategy cacheStrategy) http.Handler {
	fs := gzipfileserver.New(Assets)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Kludge to set proper MIME type. Automatic MIME detection somehow detects text/xml under
		// circumstances that couldn't be reproduced
		if filepath.Ext(r.URL.Path) == ".svg" {
			w.Header().Set("Content-Type", "image/svg+xml")
		}
		switch cacheStrategy {
		case shortTermCache:
			w.Header().Set("Cache-Control", "max-age=300, public")
		case longTermCache:
			w.Header().Set("Cache-Control", "max-age=31556926, public")
		}
		fs.ServeHTTP(w, r)
	})
}
