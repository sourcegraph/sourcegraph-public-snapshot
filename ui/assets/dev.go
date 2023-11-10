package assets

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const assetsDir = "client/web/dist"

// UseDevAssetsProvider installs the development variant of the UseDevAssetsProvider
// which expects assets to be generated on the fly by an external web builder process.
func UseDevAssetsProvider() {
	Provider = DevProvider{assets: http.Dir(assetsDir)}
}

// DevProvider is the development variant of the UseDevAssetsProvider
// which expects assets to be generated on the fly by an external web builder process.
type DevProvider struct {
	assets http.FileSystem
}

func (p DevProvider) LoadWebBuildManifest() (*WebBuildManifest, error) {
	return loadWebBuildManifest()
}

func (p DevProvider) Assets() http.FileSystem {
	return p.assets
}

var MockLoadWebBuildManifest func() (*WebBuildManifest, error)

// loadWebBuildManifest uses a web builder manifest to extract hashed bundle names to
// serve to the client. In dev mode, we load this file from disk on demand, so it doesn't
// have to exist at compile time, to avoid a build dependency between frontend
// and client.
func loadWebBuildManifest() (m *WebBuildManifest, err error) {
	if MockLoadWebBuildManifest != nil {
		return MockLoadWebBuildManifest()
	}

	manifestContent, err := os.ReadFile(filepath.Join(assetsDir, "vite-manifest.json"))
	if err != nil {
		return nil, errors.Wrap(err, "loading web build manifest file from disk")
	}
	if err := json.Unmarshal(manifestContent, &m); err != nil {
		return nil, errors.Wrap(err, "parsing manifest json")
	}
	return m, nil
}
