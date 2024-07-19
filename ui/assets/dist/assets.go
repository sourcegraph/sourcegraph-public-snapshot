package dist

import (
	"os"

	"github.com/sourcegraph/sourcegraph/ui/assets"
)

func init() {
	path := os.Getenv("SRC_ASSETS_DIR")
	if path == "" {
		panic("SRC_ASSETS_DIR not set")
	}

	_, err := os.Stat(path)
	if err != nil {
		panic(err)
	}
	assets.UseAssetsProviderForPath(path)
}
