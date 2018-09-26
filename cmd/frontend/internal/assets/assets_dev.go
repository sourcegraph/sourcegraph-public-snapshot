// +build !dist

package assets

import (
	"net/http"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/assets"
)

func init() {
	assets.Assets = http.Dir("./ui/assets")
}
