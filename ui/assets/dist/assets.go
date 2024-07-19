package dist

import (
	"os"

	"github.com/sourcegraph/sourcegraph/ui/assets"
)

const DefaultAssetPath = "dist"

func init() {
	path := os.Getenv("SRC_ASSETS_DIR")
	if path == "" {
		path = DefaultAssetPath
	}

	// _, err := os.Stat(path)
	// if err != nil {
	// 	panic(err)
	// }
	assets.UseAssetsProviderForPath(path)
}
