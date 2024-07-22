package dist

import (
	"os"

	"github.com/sourcegraph/sourcegraph/ui/assets"
)

// DefaultAssetPath is the default path where assets should be loaded from. It is primarily used when
// * AssetPath is empty
// * env var SRC_ASSETS_DIR is empty
const DefaultAssetPath = "dist"

// AssetsPath is absolute path where assets should be loaded from.
// * During init if it's value is empty the value of the environment variable `SRC_ASSETS_DIR` is used
// * If the environment variable is ALSO empty, value of the constant `DefaultAssetPath` is used
var assetsPath = ""

func init() {
	// If AssetsPath is empty try:
	// * Getting the env var value of `SRC_ASSETS_DIR`
	// * Settle for DefaultAssetPath
	if assetsPath == "" {
		path := os.Getenv("SRC_ASSETS_DIR")
		if path == "" {
			path = DefaultAssetPath
		}
		assetsPath = path
	}
	assets.UseAssetsProviderForPath(assetsPath)
}
