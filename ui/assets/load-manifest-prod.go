// +build dist

package assets

import (
	_ "embed"
	"encoding/json"
	"sync"

	"github.com/pkg/errors"
)

//go:embed webpack.manifest.json
var manifestContent []byte

var (
	once               sync.Once
	webpackManifest    *WebpackManifest
	webpackManifestErr error
)

// We use Webpack manifest to extract hashed bundle names to serve to the client
// https://webpack.js.org/concepts/manifest/
func LoadWebpackManifest() (*WebpackManifest, error) {
	once.Do(func() {
		if err := json.Unmarshal(manifestContent, &webpackManifest); err != nil {
			webpackManifestErr = errors.Wrap(err, "parsing manifest json")
		}
	})
	return webpackManifest, webpackManifestErr
}
