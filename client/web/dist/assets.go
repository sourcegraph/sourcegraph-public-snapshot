package dist

import (
	"embed"
	"encoding/json"
	"io"
	"io/fs"
	"net/http"
	"sync"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/ui/assets"
)

// By default, `go:embed *` will ignore files that start with `.` or `_`. `all:*` on the other
// hand will truly include all files.
//
//go:embed all:*
var assetsFS embed.FS

var (
	afs          fs.FS = assetsFS
	assetsHTTPFS http.FileSystem
)

var (
	webBuildManifestOnce sync.Once
	webBuildManifest     *assets.WebBuildManifest
	webBuildManifestErr  error
	assetsOnce           sync.Once
)

func init() {
	// Sets the global assets provider.
	assets.Provider = Provider{}
}

type Provider struct{}

func (p Provider) LoadWebBuildManifest() (*assets.WebBuildManifest, error) {
	webBuildManifestOnce.Do(func() {
		f, err := afs.Open("vite-manifest.json")
		if err != nil {
			webBuildManifestErr = errors.Wrap(err, "read manifest file")
			return
		}
		defer f.Close()

		manifestContent, err := io.ReadAll(f)
		if err != nil {
			webBuildManifestErr = errors.Wrap(err, "read manifest file")
			return
		}

		if err := json.Unmarshal(manifestContent, &webBuildManifest); err != nil {
			webBuildManifestErr = errors.Wrap(err, "unmarshal manifest json")
			return
		}
	})
	return webBuildManifest, webBuildManifestErr
}

func (p Provider) Assets() http.FileSystem {
	assetsOnce.Do(func() {
		// When we're building this package with Bazel, we cannot directly output the files in this current folder, because
		// it's already containing other files known to Bazel. So instead we put those into the dist folder.
		// If we do detect a dist folder when running this code, we immediately substitute the root to that dist folder.
		//
		// Therefore, this code works with both the traditional build approach and when built with Bazel.
		if _, err := assetsFS.ReadDir("dist"); err == nil {
			var err error
			afs, err = fs.Sub(assetsFS, "dist")
			if err != nil {
				panic("incorrect embed")
			}
		}
		assetsHTTPFS = http.FS(afs)
	})
	return assetsHTTPFS
}
