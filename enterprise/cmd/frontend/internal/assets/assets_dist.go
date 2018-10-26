// +build dist

package assets

import "github.com/sourcegraph/sourcegraph/cmd/frontend/assets"

func init() {
	assets.Assets = DistAssets
}
