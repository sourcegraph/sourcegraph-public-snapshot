// +build dist

package assets

import "github.com/sourcegraph/sourcegraph/cmd/frontend/external/assets"

func init() {
	assets.Assets = DistAssets
}
