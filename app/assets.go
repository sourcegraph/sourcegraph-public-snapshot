package app

import (
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"

	"golang.org/x/net/context"

	"src.sourcegraph.com/sourcegraph/app/assets"
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

func AssetsMiddleware() func(http.ResponseWriter, *http.Request, http.HandlerFunc) {
	handlers := []struct {
		PathPrefix string
		Handler    http.Handler
	}{
		{
			AssetsBasePath,
			http.StripPrefix(AssetsBasePath, assets.AssetFS(assets.ShortTermCache)),
		},
		{
			VersionedAssetsBasePath,
			http.StripPrefix(VersionedAssetsBasePath, assets.AssetFS(assets.LongTermCache)),
		},
		{
			"/versioned-assets/",
			http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				// This handler redirects paths to old versions of assets to
				// the latest asset. This should rarely happen, but in
				// production there is a race condition while deploying a new
				// version
				p := strings.SplitN(req.URL.Path, "/", 4)
				// We require len(p) == 4 since that implies we have something
				// after the version part of the path
				if len(p) >= 3 {
					http.Redirect(w, req, VersionedAssetsBasePath+p[len(p)-1], 303)
				} else {
					http.NotFound(w, req)
				}
			}),
		},
		{
			"/robots.txt",
			http.HandlerFunc(robotsTxt),
		},
		{
			"/favicon.ico",
			http.HandlerFunc(favicon),
		},
	}
	return func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		for _, h := range handlers {
			if strings.HasPrefix(r.URL.Path, h.PathPrefix) {
				h.Handler.ServeHTTP(w, r)
				return
			}
		}
		next(w, r)
	}
}
