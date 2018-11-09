// Package globals exports symbols from frontend/globals. See the parent
// package godoc for more information.
package globals

import (
	"net/url"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
)

func AppURL() *url.URL {
	return globals.AppURL
}
