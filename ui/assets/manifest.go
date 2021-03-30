package assets

import (
	_ "embed"
)

//go:embed webpack.manifest.json
var WebpackManifestJSON []byte
