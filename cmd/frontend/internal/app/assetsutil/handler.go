// Package assetsutil is a utils package for static files.
package assetsutil

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/shurcooL/httpgzip"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/ui/assets"
)

// NewAssetHandler creates the static asset handler. The handler should be wrapped into a middleware
// that enables cross-origin requests to allow the loading of the Phabricator native extension assets.
func NewAssetHandler(mux *http.ServeMux) http.Handler {
	fs := httpgzip.FileServer(assets.Provider.Assets(), httpgzip.FileServerOptions{DisableDirListing: true})

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Kludge to set proper MIME type. Automatic MIME detection somehow detects text/xml under
		// circumstances that couldn't be reproduced
		if filepath.Ext(r.URL.Path) == ".svg" {
			w.Header().Set("Content-Type", "image/svg+xml")
		}
		// Required for phabricator integration, some browser extensions block
		// unless the mime type on externally loaded JS is set
		if filepath.Ext(r.URL.Path) == ".js" {
			w.Header().Set("Content-Type", "application/javascript")
		}

		// Allow extensionHostFrame to be rendered in an iframe on trusted origins
		corsOrigin := conf.Get().CorsOrigin
		if filepath.Base(r.URL.Path) == "extensionHostFrame.html" && corsOrigin != "" {
			w.Header().Set("Content-Security-Policy", "frame-ancestors "+corsOrigin)
			w.Header().Set("X-Frame-Options", "allow-from "+corsOrigin)
		}

		// Only cache if the file is found. This avoids a race
		// condition during deployment where a 404 for a
		// not-fully-propagated asset can get cached by Cloudflare and
		// prevent any users from entire geographic regions from ever
		// being able to load that asset.
		//
		// Assets is backed by in-memory byte arrays, so this is a
		// cheap operation.
		f, err := assets.Provider.Assets().Open(r.URL.Path)
		if f != nil {
			defer f.Close()
		}
		if err == nil {
			if isPhabricatorAsset(r.URL.Path) {
				w.Header().Set("Cache-Control", "max-age=300, public")
			} else {
				w.Header().Set("Cache-Control", "immutable, max-age=31536000, public")
			}
		}

		fs.ServeHTTP(w, r)
	})
}

func isPhabricatorAsset(path string) bool {
	return strings.Contains(path, "phabricator.bundle.js")
}
