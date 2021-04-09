package assets

import (
	_ "embed"
	"encoding/json"
	"io/ioutil"

	"github.com/pkg/errors"
)

// We use Webpack manifest to extract hashed bundle names to serve to the client
// https://webpack.js.org/concepts/manifest/
func LoadWebpackManifest() (m *WebpackManifest, err error) {
	manifestContent, err := ioutil.ReadFile("./ui/assets/webpack.manifest.json")
	if err != nil {
		return nil, errors.Wrap(err, "loading webpack manifest file from disk")
	}
	if err := json.Unmarshal(manifestContent, &m); err != nil {
		return nil, errors.Wrap(err, "parsing manifest json")
	}
	return m, nil
}
