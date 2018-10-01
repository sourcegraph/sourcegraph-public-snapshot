package globals

import (
	"net/url"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
)

func AppURL() *url.URL {
	return globals.AppURL
}
