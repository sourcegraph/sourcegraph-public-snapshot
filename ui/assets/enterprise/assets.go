package enterprise

import (
	"embed"
	"encoding/json"
	"net/http"
	"sync"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/ui/assets"
)

//go:embed *
var assetsFS embed.FS

var Assets = http.FS(assetsFS)

var (
	webpackManifestOnce sync.Once
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

func (p Provider) Assets() http.FileSystem {
	return Assets
}
