//go:generate go run generate.go

// Package bundle contains the bundled assets for the vscode
// application.
//
// To fetch the prebuilt vscode package, run:
//
//   cmd/frontend/internal/app/bundle/fetch-and-generate.bash
//
// To publish a vscode package, run the following:
//
//   # first, in vscode:
//   gulp vscode-browser # or vscode-browser-min
//
//   # then, in sourcegraph:
//   cmd/frontend/internal/app/bundle/publish-package.bash ~/src/github.com/Microsoft/VSCode-browser $VERSION
//
// The $VERSION is currently chosen manually and should be
// unique. Update fetch-and-generate.bash's version number when you
// publish a new package.
package bundle

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"strconv"

	"github.com/shurcooL/httpgzip"
)

var (
	devMode, _ = strconv.ParseBool(os.Getenv("VSCODE_DEV"))
	errNoApp   = errors.New("vscode app is not enabled on this server")
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
		// TODO(sqs): implement Cache-Control: immutable, and add a
		// version identifier to the URL path.
		if devMode {
			w.Header().Set("Cache-Control", "public, must-revalidate")
		} else {
			w.Header().Set("Cache-Control", "public, max-age=300")
		}

		if path.Base(r.URL.Path) == "index.html" {
			// The UI uses iframes, so we need to allow iframes.
			w.Header().Set("X-Frame-Options", "SAMEORIGIN")
		}

		fs.ServeHTTP(w, r)
	})
}
