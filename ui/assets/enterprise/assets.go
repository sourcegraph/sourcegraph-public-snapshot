package enterprise

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

//go:embed *
var assetsFS embed.FS
var afs fs.FS = assetsFS

var Assets http.FileSystem

var (
	webpackManifestOnce sync.Once
	assetsOnce          sync.Once
	webpackManifest     *assets.WebpackManifest
	webpackManifestErr  error
)

func init() {
	// Sets the global assets provider.
	assets.Provider = Provider{}
}

type Provider struct{}

func (p Provider) LoadWebpackManifest() (*assets.WebpackManifest, error) {
	webpackManifestOnce.Do(func() {
		f, err := afs.Open("webpack.manifest.json")
		if err != nil {
			webpackManifestErr = errors.Wrap(err, "read manifest file")
			return
		}
		defer f.Close()

		manifestContent, err := io.ReadAll(f)
		if err != nil {
			webpackManifestErr = errors.Wrap(err, "read manifest file")
			return
		}

		if err := json.Unmarshal(manifestContent, &webpackManifest); err != nil {
			webpackManifestErr = errors.Wrap(err, "unmarshal manifest json")
			return
		}
	})
	return webpackManifest, webpackManifestErr
}

func (p Provider) Assets() http.FileSystem {
	assetsOnce.Do(func() {
		// When we're building this package with Bazel, we cannot directly output the files in this current folder, because
		// it's already containing other files known to Bazel. So instead we put those into the dist folder.
		// If we do detect a dist folder when running this code, we immediately substitute the root to that dist folder.
		//
		// Therefore, this code works with both the traditionnal build approach and when built with Bazel.
		if _, err := assetsFS.ReadDir("dist"); err == nil {
			var err error
			afs, err = fs.Sub(assetsFS, "dist")
			if err != nil {
				panic("incorrect embed")
			}
		}
		Assets = http.FS(afs)
	})

	return Assets
}
