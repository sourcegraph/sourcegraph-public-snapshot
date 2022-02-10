//go:build dist
// +build dist

package assets

import (
	_ "embed"
	"encoding/json"
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
		manifestContent, err := assetsFS.ReadFile("webpack.manifest.json")
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
