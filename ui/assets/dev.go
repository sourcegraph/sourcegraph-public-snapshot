package assets

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type DevProvider struct {
	assets http.FileSystem
}

func (p DevProvider) LoadWebpackManifest() (*WebpackManifest, error) {
	return LoadWebpackManifest()
}

func (p DevProvider) Assets() http.FileSystem {
	return p.assets
}

var MockLoadWebpackManifest func() (*WebpackManifest, error)

// LoadWebpackManifest uses Webpack manifest to extract hashed bundle names to
// serve to the client, see https://webpack.js.org/concepts/manifest/ for
// details. In dev mode, we load this file from disk on demand, so it doesn't
// have to exist at compile time, to avoid a build dependency between frontend
// and client.
func LoadWebpackManifest() (m *WebpackManifest, err error) {
	if MockLoadWebpackManifest != nil {
		return MockLoadWebpackManifest()
	}

	manifestContent, err := os.ReadFile("./ui/assets/webpack.manifest.json")
	if err != nil {
		return nil, errors.Wrap(err, "loading webpack manifest file from disk")
	}
	if err := json.Unmarshal(manifestContent, &m); err != nil {
		return nil, errors.Wrap(err, "parsing manifest json")
	}
	return m, nil
}
