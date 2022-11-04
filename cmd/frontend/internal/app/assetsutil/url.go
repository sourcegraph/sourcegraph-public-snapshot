package assetsutil

import (
	"fmt"
	"net/url"
	"path"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

var (
	assetsRoot = env.Get("ASSETS_ROOT", "/.assets", "URL to web assets")

	// baseURL is the path prefix under which static assets should
	// be served.
	baseURL = &url.URL{}
)

func init() {
	var err error
	baseURL, err = url.Parse(assetsRoot)
	if err != nil {
		panic(fmt.Sprintf("Parsing ASSETS_ROOT failed: %s", err))
	}
}

// URL returns a URL, possibly relative, to the asset at path
// p.
func URL(p string) *url.URL {
	return baseURL.ResolveReference(&url.URL{Path: path.Join(baseURL.Path, p)})
}
