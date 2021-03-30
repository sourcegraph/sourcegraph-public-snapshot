package assets

import (
	_ "embed"
)

// We use Webpack manifest to extract hashed bundle names to serve to the client
// https://webpack.js.org/concepts/manifest/

//go:embed webpack.manifest.json
var WebpackManifestJSON []byte
