package assets

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// UseDevAssetsProvider installs the development variant of the UseDevAssetsProvider
// which expects assets to be generated on the fly by an external web builder process.
func UseDevAssetsProvider() {
	devAssetsDir := "client/web/dist"
	Provider = DirProvider{dir: devAssetsDir, assets: http.Dir(devAssetsDir)}
}

func UseAssetsProviderForPath(path string) {
	Provider = DirProvider{dir: path, assets: http.Dir(path)}
}

// DirProvider provides assets from http filesystem rooted at the given path
type DirProvider struct {
	dir    string
	assets http.FileSystem
}

func (p DirProvider) LoadWebBuildManifest() (*WebBuildManifest, error) {
	return loadWebBuildManifest(p.dir)
}

func (p DirProvider) Assets() http.FileSystem {
	return p.assets
}

var MockLoadWebBuildManifest func() (*WebBuildManifest, error)

// loadWebBuildManifest uses a web builder manifest to extract hashed bundle names to
// serve to the client. In dev mode, we load this file from disk on demand, so it doesn't
// have to exist at compile time, to avoid a build dependency between frontend
// and client.
func loadWebBuildManifest(rootDir string) (m *WebBuildManifest, err error) {
	if MockLoadWebBuildManifest != nil {
		return MockLoadWebBuildManifest()
	}

	manifestContent, err := os.ReadFile(filepath.Join(rootDir, "web.manifest.json"))
	if err != nil {
		return nil, errors.Wrap(err, "loading web build manifest file from disk")
	}
	if err := json.Unmarshal(manifestContent, &m); err != nil {
		return nil, errors.Wrap(err, "parsing manifest json")
	}
	return m, nil
}
