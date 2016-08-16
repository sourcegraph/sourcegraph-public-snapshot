package assets

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"strings"

	"context"

	"github.com/gorilla/mux"
	"sourcegraph.com/sourcegraph/sourcegraph/app/appconf"
	"sourcegraph.com/sourcegraph/sourcegraph/cli/buildvar"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
)

const URLPathPrefix = "/.assets"

var (
	// versionedAssetsBasePath is the url path prefix for long term cached assets
	versionedAssetsBasePath = "/v/" + buildvar.Version + "/"
)

// URL returns a URL, possibly relative, to the asset at path
// p. If you need an absolute URL, use AbsURL.
func URL(p string) *url.URL {
	return &url.URL{Path: path.Join(URLPathPrefix, versionedAssetsBasePath, p)}
}

// AbsURL returns an absolute URL to the asset at path p.
func AbsURL(ctx context.Context, p string) *url.URL {
	u := URL(p)
	if !u.IsAbs() {
		u.Scheme = conf.AppURL(ctx).Scheme
		u.Host = conf.AppURL(ctx).Host
	}
	return u
}

func NewHandler(m *mux.Router) http.Handler {
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
			if strings.Contains(r.URL.Path, "img/") {
				r.URL.Path = path.Join("/assets", r.URL.Path)
			} else {
				r.URL.Path = path.Join("/", r.URL.Path)
			}
			origDirector(r)
		}
		assetHandler = proxy
	} else {
		// Production: serve assets directly from bundled assets
		assetHandler = AssetFS()
	}

	// Fonts are fetched from /font/... paths, not
	// /v/VERSION/... paths, for two reasons: (1) their
	// paths are defined in SCSS not Go templates, which makes
	// it harder to change their paths from Go code; (2) they
	// change very frequently, so we actually prefer to cache
	// them across version upgrades.
	m.PathPrefix("/font/").Handler(assetHandler).Name("assets.font")

	m.PathPrefix(versionedAssetsBasePath).Handler(http.StripPrefix(versionedAssetsBasePath, assetHandler)).Name("assets.versioned")

	m.PathPrefix("/v/").HandlerFunc(serveRedirectFromOldVersion).Name("assets.redirect-from-old-version")

	return m
}

func serveRedirectFromOldVersion(w http.ResponseWriter, req *http.Request) {
	// This handler redirects paths to old versions of assets to
	// the latest asset. This should rarely happen, but in
	// production there is a race condition while deploying a new
	// version
	p := strings.SplitN(req.URL.Path, "/", 3)
	// We require len(p) >= 3 since that implies we have something
	// after the version part of the path (/v/MYVER/SOMETHING).
	if len(p) >= 3 {
		http.Redirect(w, req, URL(path.Base(req.URL.Path)).String(), http.StatusSeeOther)
	} else {
		http.NotFound(w, req)
	}
}
