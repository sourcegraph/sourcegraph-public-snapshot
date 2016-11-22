// +build !dist

package assets

import (
	"log"
	"net/http"
	"net/url"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
)

// Mount is a no-op (because this file is only used if the 'dist' tag
// is disabled). Static assets are served directly by Webpack (using
// the WEBPACK_DEV_SERVER_URL env var) in non-dist mode.
func Mount(mux *http.ServeMux) {}

func init() {
	if urlStr := env.Get("WEBPACK_DEV_SERVER_URL", "", "the URL of Webpack serving the assets"); urlStr != "" {
		var err error
		baseURL, err = url.Parse(urlStr)
		if err != nil {
			log.Fatalf("WEBPACK_DEV_SERVER_URL: %s", err)
		}
	}
}

const mainJavaScriptBundlePath = "/main.js"
