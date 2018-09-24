package globals

import (
	"net/url"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/globals"
)

func AppURL() *url.URL {
	return globals.AppURL
}
