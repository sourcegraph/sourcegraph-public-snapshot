// Package assetsutil is a utils package for static files.
package assetsutil

import (
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/assets"
	"github.com/sourcegraph/sourcegraph/pkg/env"
	"github.com/sqs/httpgzip"
)

// Mount mounts the static asset handler.
func Mount(mux *http.ServeMux) {
	const urlPathPrefix = "/.assets"

	fs := httpgzip.FileServer(assets.Assets, httpgzip.FileServerOptions{})
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
	if strings.Contains(path, "phabricator.bundle.js") {
		return true
	}
	if strings.Contains(path, "sgdev.bundle.sj") {
		return true
	}
	if strings.Contains(path, "umami.bundle.sj") {
		return true
	}
	return false
}
