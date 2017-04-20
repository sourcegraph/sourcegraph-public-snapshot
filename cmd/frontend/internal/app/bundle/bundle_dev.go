// +build !dist,!distbundle

package bundle

import (
	"net/http"
	"os"
)

// BaseDir is the path to the VSCode-browser or VSCode-browser-min
// directory built by running `gulp vscode-browser` or `gulp
// vscode-browser-min` in the Sourcegraph fork of vscode.
//
// It is used by Handler when running in dev mode (neither the build
// tag "dist" nor "distbundle" is satisfied), and it's used when
// running `go generate` on this package.
//
// If empty at dev time, the vscode app will not be available on this
// server.
var BaseDir = os.Getenv("VSCODE_BROWSER_PKG")

var Data http.FileSystem

func init() {
	if BaseDir != "" {
		Data = http.Dir(BaseDir)
	}
}
