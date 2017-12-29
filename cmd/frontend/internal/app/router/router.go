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

	OpenSearch = "opensearch"

	RepoBadge = "repo.badge"

	Logout = "logout"

	SignIn2           = "sign-in2"
	SignOut           = "sign-out"
	SignUp            = "sign-up"
	VerifyEmail       = "verify-email"
	ResetPasswordInit = "reset-password.init"
	ResetPassword     = "reset-password"

	OldToolsRedirect = "old-tools-redirect"
	OldTreeRedirect  = "old-tree-redirect"

	GDDORefs = "gddo.refs"
	Editor   = "editor"

	DebugHeaders = "debug.headers"

	GopherconLiveBlog = "gophercon.live.blog"

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
	base.Path("/opensearch.xml").Methods("GET").Name(OpenSearch)

	base.Path("/-/logout").Methods("GET").Name(Logout)

	base.Path("/-/sign-up").Methods("POST").Name(SignUp)
	base.Path("/-/verify-email").Methods("GET").Name(VerifyEmail)
	base.Path("/-/sign-in-2").Methods("POST").Name(SignIn2)
	base.Path("/-/sign-out").Methods("GET").Name(SignOut)
	base.Path("/-/reset-password-init").Methods("POST").Name(ResetPasswordInit)
	base.Path("/-/reset-password").Methods("POST").Name(ResetPassword)

	base.Path("/-/godoc/refs").Methods("GET").Name(GDDORefs)
	base.Path("/-/editor").Methods("GET").Name(Editor)

	base.Path("/-/debug/headers").Methods("GET").Name(DebugHeaders)

	base.Path("/gophercon").Methods("GET").Name(GopherconLiveBlog)

	addOldTreeRedirectRoute(&Router{*base}, base)
	base.Path("/tools").Methods("GET").Name(OldToolsRedirect)

	base.PathPrefix("/go/").Methods("GET").Name(GoSymbolURL)

	repoPath := `/` + routevar.Repo
	repo := base.PathPrefix(repoPath + "/" + routevar.RepoPathDelim + "/").Subrouter()
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
