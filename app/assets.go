package app

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"strings"

	"golang.org/x/net/context"

	"src.sourcegraph.com/sourcegraph/app/appconf"
	"src.sourcegraph.com/sourcegraph/app/assets"
	"src.sourcegraph.com/sourcegraph/conf"
	"src.sourcegraph.com/sourcegraph/sgx/buildvar"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
)

var (
	// versionedAssetsBasePath is the url path prefix for long term cached assets
	versionedAssetsBasePath = "/versioned-assets/" + buildvar.Version + "/"
)

// assetURL returns a URL, possibly relative, to the asset at path
// p. If you need an absolute URL, use assetAbsURL.
func assetURL(p string) *url.URL {
	return &url.URL{Path: path.Join(versionedAssetsBasePath, p)}
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
	var assetHandler http.Handler
	if urlStr := appconf.Flags.WebpackDevServerURL; urlStr != "" {
		url, err := url.Parse(urlStr)
		if err != nil {
			log.Fatalf("Error parsing Webpack dev server URL %q: %s.", urlStr, err)
		}

		// Dev: proxy asset requests to Webpack
		proxy := httputil.NewSingleHostReverseProxy(url)
		origDirector := proxy.Director
		proxy.Director = func(r *http.Request) {
			r.URL.Path = path.Join("/assets", r.URL.Path)
			origDirector(r)
		}
		assetHandler = proxy
	} else {
		// Production: serve assets directly from bundled assets
		assetHandler = assets.AssetFS(assets.LongTermCache)
	}

	handlers := []struct {
		PathPrefix string
		Handler    http.Handler
	}{
		{
			// Fonts are fetched from /assets/... paths, not
			// /versioned-assets/... paths, for two reasons: (1) their
			// paths are defined in SCSS not Go templates, which makes
			// it harder to change their paths from Go code; (2) they
			// change very frequently, so we actually prefer to cache
			// them across version upgrades.
			"/assets/",
			http.StripPrefix("/assets/", assetHandler),
		},
		{
			versionedAssetsBasePath,
			http.StripPrefix(versionedAssetsBasePath, assetHandler),
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
					http.Redirect(w, req, versionedAssetsBasePath+p[len(p)-1], 303)
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
				httpctx.SetRouteName(r, "asset")
				h.Handler.ServeHTTP(w, r)
				return
			}
		}
		next(w, r)
	}
}
