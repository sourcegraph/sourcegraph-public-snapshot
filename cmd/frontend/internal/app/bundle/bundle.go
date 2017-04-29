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
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"

	"github.com/shurcooL/httpgzip"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/jscontext"
)

// This list should be periodically updated to be in sync with the
// unpushed resources loaded over the network when a browser loads the
// app.
//
// TODO(sqs): It would be nice but might not be worth the effort to
// generate this list automatically.
var pushResources = []string{
	"out/vs/launcher/browser/bootstrap/index.js",

	"out/vs/workbench/browser/bootstrap/config.js",
	"out/vs/workbench/browser/bootstrap/index.js",
	"out/vs/loader.js",
	"out/vs/code/browser/main.js",
	"out/vs/code/browser/main.css",
	"out/vs/code/browser/main.nls.js",

	"extensions/diff/package.json",
	"extensions/diff/language-configuration.json",
	"extensions/docker/package.json",
	"extensions/file-links/package.json",
	"extensions/gitsyntax/package.json",
	"extensions/go/package.json",
	"extensions/json/package.json",
	"extensions/lsp/package.json",
	"extensions/markdown/package.json",
	"extensions/theme-abyss/package.json",
	"extensions/theme-defaults/package.json",
	"extensions/theme-kimbie-dark/package.json",
	"extensions/theme-monokai/package.json",
	"extensions/theme-monokai-dimmed/package.json",
	"extensions/theme-quietlight/package.json",
	"extensions/theme-red/package.json",
	"extensions/theme-seti/package.json",
	"extensions/theme-solarized-dark/package.json",
	"extensions/theme-solarized-light/package.json",
	"extensions/theme-sourcegraph/package.json",
	"extensions/theme-tomorrow-night-blue/package.json",

	"out/vs/workbench/browser/extensionHostProcess.js",
	"out/vs/workbench/browser/extensionHostProcess.nls.js",
	"out/browser_modules/lsp.js",
}

const rootPath = "/out/vs/launcher/browser/bootstrap/index.html"

var devMode, _ = strconv.ParseBool(os.Getenv("VSCODE_DEV"))

// Handler handles HTTP requests for files in the bundle.
func Handler() http.Handler {
	if Data == nil {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "vscode app is not enabled on this server")
		})
	}

	fs := httpgzip.FileServer(Data, httpgzip.FileServerOptions{})
	file, err := Data.Open(rootPath)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	rootData, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO(sqs): implement Cache-Control: immutable, and add a
		// version identifier to the URL path.
		if devMode {
			w.Header().Set("Cache-Control", "public, must-revalidate")
		} else {
			w.Header().Set("Cache-Control", "public, max-age=300")
		}
		if pusher, ok := w.(http.Pusher); ok {
			if path.Base(r.URL.Path) == "index.html" {
				opt := &http.PushOptions{
					Header: http.Header{
						"Accept":          r.Header["Accept"],
						"Accept-Encoding": r.Header["Accept-Encoding"],
						"Cookie":          r.Header["Cookie"],
						"Authorization":   r.Header["Authorization"],
					},
				}
				for _, r := range pushResources {
					p := path.Join("/main", r)
					if err := pusher.Push(p, opt); err != nil {
						log.Printf("warning: HTTP/2 push %q failed: %s", p, err)
						break
					}
				}
			}
		}

		// The UI uses iframes, so we need to allow iframes.
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")

		if r.URL.Path == rootPath {
			// Inject the user's JSContext.
			jsctx, err := json.Marshal(jscontext.NewJSContextFromRequest(r))
			if err == nil {
				// This is safe because Go's json.Marshal escapes '<'
				// and '>' so that if jsctx contains "</script>"
				// substrings or similar, the browser won't interpret
				// them as HTML closing tags.
				inject := `<script type="text/javascript">window.__sourcegraphJSContext = ` + string(jsctx) + ";</script>"
				data := bytes.Replace(rootData, []byte("<!-- INSERT SOURCEGRAPH CONTEXT -->"), []byte(inject), 1)
				_, _ = w.Write(data)
			} else {
				log.Println(err)
			}
			return
		}

		fs.ServeHTTP(w, r)
	})
}
