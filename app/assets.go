package app

import (
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"

	"golang.org/x/net/context"

	"src.sourcegraph.com/sourcegraph/conf"
	"src.sourcegraph.com/sourcegraph/sgx/buildvar"
)

var (
	// UseWebpackDevServer is whether to route asset requests through the
	// Webpack development server, which automatically rebuilds assets
	// when underlying SCSS/JS/etc. files change and autoreloads the
	// browser.
	UseWebpackDevServer, _ = strconv.ParseBool(os.Getenv("SG_USE_WEBPACK_DEV_SERVER"))

	// AssetsBasePath is the url path prefix for short term cached assets
	AssetsBasePath = "/assets/"

	// VersionedAssetsBasePath is the url path prefix for long term cached assets
	VersionedAssetsBasePath = "/versioned-assets/" + buildvar.Version + "/"
)

// assetURL returns a URL, possibly relative, to the asset at path
// p. If you need an absolute URL, use assetAbsURL.
func assetURL(p string) *url.URL {
	var base url.URL
	var basePath string
	if UseWebpackDevServer && !strings.Contains(p, "woff") {
		base.Scheme = "http"
		base.Host = "localhost:8080"
		basePath = AssetsBasePath
	} else {
		basePath = VersionedAssetsBasePath
	}
	return base.ResolveReference(&url.URL{Path: path.Join(basePath, p)})
}

// assetAbsURL returns an absolute URL to the asset at path p.
func assetAbsURL(ctx context.Context, p string) *url.URL {
	u := assetURL(p)
	if !u.IsAbs() {
		u.Scheme = conf.AppURL(ctx).Scheme
		u.Host = conf.AppURL(ctx).Host
	}
	return u
}
