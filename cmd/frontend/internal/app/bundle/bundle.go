//go:generate go run generate.go

// Package bundle contains the bundled assets for the vscode
// application.
//
// To fetch the prebuilt vscode package, run:
//
//   cmd/frontend/internal/app/bundle/fetch-and-generate.bash
//
// To publish a vscode package, follow the steps in vscode-private/SOURCEGRAPH.md
package bundle

import (
	"errors"
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/shurcooL/httpgzip"
)

var (
	cacheKey     string
	cacheControl string

	errNoApp = errors.New("vscode app is not enabled on this server")
)

// Handler handles HTTP requests for files in the bundle.
func Handler() http.Handler {
	if Data == nil {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, errNoApp.Error())
		})
	}

	fs := httpgzip.FileServer(Data, httpgzip.FileServerOptions{})

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If a cache key is specified for the bundle, then require it as a prefix on all
		// URL paths.
		if cacheKey != "" {
			prefix := "/" + cacheKey
			if p := strings.TrimPrefix(r.URL.Path, prefix); len(p) < len(r.URL.Path) {
				r.URL.Path = p
			} else {
				http.NotFound(w, r)
				return
			}
		}

		if cacheControl != "" {
			w.Header().Set("Cache-Control", cacheControl)
		}
		w.Header().Set("Vary", "Accept-Encoding")

		if name := path.Base(r.URL.Path); name == "index.html" || name == "webview.html" {
			// The UI uses iframes, so we need to allow iframes.
			w.Header().Set("X-Frame-Options", "SAMEORIGIN")
		}

		fs.ServeHTTP(w, r)
	})
}
