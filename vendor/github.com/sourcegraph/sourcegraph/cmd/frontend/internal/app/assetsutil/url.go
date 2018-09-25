package assetsutil

import (
	"net/url"
	"path"
)

// baseURL is the path prefix under which static assets should
// be served.
var baseURL = &url.URL{}

// URL returns a URL, possibly relative, to the asset at path
// p.
func URL(p string) *url.URL {
	return baseURL.ResolveReference(&url.URL{Path: path.Join(baseURL.Path, p)})
}
