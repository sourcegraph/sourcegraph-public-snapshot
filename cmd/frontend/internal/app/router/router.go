// Package router contains the URL router for the frontend app.
package router

import (
	"log"
	"net/url"

	"github.com/gorilla/mux"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/routevar"
)

const (
	RobotsTxt = "robots-txt"
	Favicon   = "favicon"

	SitemapIndex = "sitemap-index"
	RepoSitemap  = "repo.sitemap"
	RepoBadge    = "repo.badge"

	Logout = "logout"

	GitHubOAuth2Initiate = "github-oauth2.initiate"
	GitHubOAuth2Receive  = "github-oauth2.receive"

	GoogleOAuth2Initiate = "google-oauth2.initiate"
	GoogleOAuth2Receive  = "google-oauth2.receive"

	InstallZap = "install.zap"

	OldDefRedirect   = "old-def-redirect"
	OldToolsRedirect = "old-tools-redirect"
	OldTreeRedirect  = "old-tree-redirect"

	GDDORefs = "gddo.refs"
	Editor   = "editor"

	ShowAuth = "show-auth"

	GoSymbolURL = "go-symbol-url"

	UI = "ui"
)

// Router is an app URL router.
type Router struct{ mux.Router }

// New creates a new app router with route URL pattern definitions but
// no handlers attached to the routes.
//
// It is in a separate package from app so that other packages may use it to
// generate URLs without resulting in Go import cycles (and so we can release
// the router as open-source to support our client library).
func New() *Router {
	base := mux.NewRouter()

	base.StrictSlash(true)

	base.Path("/robots.txt").Methods("GET").Name(RobotsTxt)
	base.Path("/favicon.ico").Methods("GET").Name(Favicon)

	base.Path("/sitemap.xml").Methods("GET").Name(SitemapIndex)

	base.Path("/-/logout").Methods("GET").Name(Logout)

	base.Path("/-/github-oauth/initiate").Methods("POST").Name(GitHubOAuth2Initiate)
	base.Path("/-/github-oauth/receive").Methods("GET", "POST").Name(GitHubOAuth2Receive)

	base.Path("/-/google-oauth/initiate").Methods("POST").Name(GoogleOAuth2Initiate)
	base.Path("/-/google-oauth/receive").Methods("GET", "POST").Name(GoogleOAuth2Receive)

	base.Path("/-/godoc/refs").Methods("GET").Name(GDDORefs)
	base.Path("/-/editor").Methods("GET").Name(Editor)

	base.Path("/-/show-auth").Methods("GET").Name(ShowAuth)

	addOldTreeRedirectRoute(&Router{*base}, base)
	base.Path("/tools").Methods("GET").Name(OldToolsRedirect)

	base.Path("/install/zap").Methods("GET").Name(InstallZap)

	base.PathPrefix("/go/").Methods("GET").Name(GoSymbolURL)

	repoPath := `/` + routevar.Repo
	repo := base.PathPrefix(repoPath + "/" + routevar.RepoPathDelim + "/").Subrouter()
	repo.Path("/sitemap.xml").Methods("GET").Name(RepoSitemap)
	repo.Path("/badge.svg").Methods("GET").Name(RepoBadge)

	// Must come last
	base.PathPrefix("/").Methods("GET").Name(UI)

	return &Router{*base}
}

func (r *Router) URLToOrError(routeName string, params ...string) (*url.URL, error) {
	route := r.Get(routeName)
	if route == nil {
		log.Panicf("no such route: %q (params: %v)", routeName, params)
	}
	u, err := route.URL(params...)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *Router) URLTo(routeName string, params ...string) *url.URL {
	u, err := r.URLToOrError(routeName, params...)
	if err != nil {
		log.Printf("Route error: failed to make URL for route %q (params: %v): %s", routeName, params, err)
		return &url.URL{}
	}
	return u
}

var Rel = New()
