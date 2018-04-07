package assets

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/sqs/httpgzip"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
)

// Mount mounts the static asset handler.
func Mount(mux *http.ServeMux) {
	const urlPathPrefix = "/.assets"

	if urlStr := os.Getenv("WEBPACK_DEV_SERVER_URL"); urlStr != "" {
		// When using the Webpack dev server, we need to proxy assets so they live on the same
		// origin, because WebWorker scripts (required by the Monaco editor) can't be loaded
		// cross-origin.
		webpackDevServerURL, err := url.Parse(urlStr)
		if err != nil {
			log.Fatalln("WEBPACK_DEV_SERVER_URL:", err)
		}
		proxy := httputil.NewSingleHostReverseProxy(webpackDevServerURL)
		mux.Handle(urlPathPrefix+"/", &httputil.ReverseProxy{
			Director: func(r *http.Request) {
				r.Host = webpackDevServerURL.Host
				r.URL.Path = strings.TrimPrefix(r.URL.Path, urlPathPrefix)
				proxy.Director(r)
			},
		})
		return
	}

	fs := httpgzip.FileServer(Assets, httpgzip.FileServerOptions{})
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
		f, err := Assets.Open(r.URL.Path)
		if f != nil {
			defer f.Close()
		}
		if err == nil {
			if isPhabricatorAsset(r.URL.Path) {
				w.Header().Set("Cache-Control", "max-age=300, public")
			} else {
				w.Header().Set("Cache-Control", "max-age=25200, public")
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
