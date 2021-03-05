// Package assetsutil is a utils package for static files.
package assetsutil

import (
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/shurcooL/httpgzip"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/assets"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/env"
)

// Mount mounts the static asset handler.
func Mount(mux *http.ServeMux) {
	const urlPathPrefix = "/.assets"

	fs := httpgzip.FileServer(assets.Assets, httpgzip.FileServerOptions{DisableDirListing: true})
	mux.Handle(urlPathPrefix+"/", http.StripPrefix(urlPathPrefix, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		f, err := assets.Assets.Open(r.URL.Path)
		if f != nil {
			defer f.Close()
		}
		if err == nil {
			if isPhabricatorAsset(r.URL.Path) {
				w.Header().Set("Cache-Control", "max-age=300, public")
			} else {
				w.Header().Set("Cache-Control", "immutable, max-age=172800, public")
			}
		}

		fs.ServeHTTP(w, r)
	})))
}

var assetsRoot = env.Get("ASSETS_ROOT", "/.assets", "URL to web assets")

func init() {
	var err error
	baseURL, err = url.Parse(assetsRoot)
	if err != nil {
		log.Fatalln("Parsing ASSETS_ROOT failed:", err)
	}
}

func isPhabricatorAsset(path string) bool {
	return strings.Contains(path, "phabricator.bundle.js")
}
