//go:build dist
// +build dist

package assets

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed *
var assetsFS embed.FS
var afs fs.FS = assetsFS

var Assets http.FileSystem

func init() {
	// When we're building this package with Bazel, we cannot directly output the files in this current folder, because
	// it's already containing other files known to Bazel. So instead we put those into the dist folder.
	// If we do detect a dist folder when running this code, we immediately substitute the root to that dist folder.
	//
	// Therefore, this code works with both the traditionnal build approach and when built with Bazel.
	if fs.ValidPath("dist") {
		var err error
		afs, err = fs.Sub(assetsFS, "dist")
		if err != nil {
			panic("incorrect embed")
		}
	}

	Assets = http.FS(afs)
}
