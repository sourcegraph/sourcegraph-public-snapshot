package assets

import (
	_ "embed"
)

// 🚨 WARNING: This is just a shim implementation that
// hardcodes the app bundle name.
//
// TODO: The webpack configuration will need to be adjusted so that
// it outputs its manifest file in this folder.

//go:embed webpack.manifest.json
var WebpackManifestJSON []byte
