// +build !dist

package assets

import (
	"log"
	"net/http"
	"net/url"
	"os"
)

// Mount is a no-op (because this file is only used if the 'dist' tag
// is disabled). Static assets are served directly by Webpack (using
// the WEBPACK_DEV_SERVER_URL env var) in non-dist mode.
func Mount(mux *http.ServeMux) {}

func init() {
	if urlStr := os.Getenv("WEBPACK_DEV_SERVER_URL"); urlStr != "" {
		var err error
		baseURL, err = url.Parse(urlStr)
		if err != nil {
			log.Fatalf("WEBPACK_DEV_SERVER_URL: %s", err)
		}
	}
}

const mainJavaScriptBundlePath = "/main.js"
