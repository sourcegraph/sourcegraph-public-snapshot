//go:build !dist
// +build !dist

package assets

import (
	"encoding/json"
	"os"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

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
