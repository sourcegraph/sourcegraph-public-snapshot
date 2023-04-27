//go:build dist
// +build dist

package assets

import (
	_ "embed"
	"encoding/json"
	"io"
	"sync"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	webpackManifestOnce sync.Once
	webpackManifest     *WebpackManifest
	webpackManifestErr  error
)

// LoadWebpackManifest uses Webpack manifest to extract hashed bundle names to
// serve to the client, see https://webpack.js.org/concepts/manifest/ for
// details.
func LoadWebpackManifest() (*WebpackManifest, error) {
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
