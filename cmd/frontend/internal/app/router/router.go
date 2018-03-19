// Package router contains the URL router for the frontend app.
//
// It is in a separate package from app so that other packages may use it to generate URLs without resulting in Go
// import cycles.
package router

import (
	"github.com/gorilla/mux"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/routevar"
)

const (
	RobotsTxt = "robots-txt"
	Favicon   = "favicon"

	OpenSearch = "opensearch"

	RepoBadge          = "repo.badge"
	RepoExternalCommit = "repo.external.commit"

	Logout = "logout"

	SignIn            = "sign-in"
	SignOut           = "sign-out"
	SignUp            = "sign-up"
	SiteInit          = "site-init"
	VerifyEmail       = "verify-email"
	ResetPasswordInit = "reset-password.init"
	ResetPassword     = "reset-password"

	OldToolsRedirect = "old-tools-redirect"
	OldTreeRedirect  = "old-tree-redirect"

	GDDORefs = "gddo.refs"
	Editor   = "editor"

	Debug        = "debug"
	DebugHeaders = "debug.headers"

	GopherconLiveBlog = "gophercon.live.blog"

	GoSymbolURL = "go-symbol-url"

	UI = "ui"
)

// Router returns the frontend app router.
func Router() *mux.Router { return router }

var router = newRouter()

func newRouter() *mux.Router {
	base := mux.NewRouter()

	base.StrictSlash(true)

	base.Path("/robots.txt").Methods("GET").Name(RobotsTxt)
	base.Path("/favicon.ico").Methods("GET").Name(Favicon)
	base.Path("/opensearch.xml").Methods("GET").Name(OpenSearch)

	base.Path("/-/logout").Methods("GET").Name(Logout)

	base.Path("/-/sign-up").Methods("POST").Name(SignUp)
	base.Path("/-/site-init").Methods("POST").Name(SiteInit)
	base.Path("/-/verify-email").Methods("GET").Name(VerifyEmail)
	base.Path("/-/sign-in").Methods("POST").Name(SignIn)
	base.Path("/-/sign-out").Methods("GET").Name(SignOut)
	base.Path("/-/reset-password-init").Methods("POST").Name(ResetPasswordInit)
	base.Path("/-/reset-password").Methods("POST").Name(ResetPassword)

	base.Path("/-/godoc/refs").Methods("GET").Name(GDDORefs)
	base.Path("/-/editor").Methods("GET").Name(Editor)

	base.Path("/-/debug/headers").Methods("GET").Name(DebugHeaders)
	base.PathPrefix("/-/debug").Name(Debug)

	base.Path("/gophercon").Methods("GET").Name(GopherconLiveBlog)

	addOldTreeRedirectRoute(base)
	base.Path("/tools").Methods("GET").Name(OldToolsRedirect)

	base.PathPrefix("/go/").Methods("GET").Name(GoSymbolURL)

	repoPath := `/` + routevar.Repo
	repo := base.PathPrefix(repoPath + "/" + routevar.RepoPathDelim + "/").Subrouter()
	repo.Path("/badge.svg").Methods("GET").Name(RepoBadge)

	repoExternal := repo.PathPrefix("/external/").Subrouter()
	repoExternal.Path("/commit/{commit}").Methods("GET").Name(RepoExternalCommit)

	// Must come last
	base.PathPrefix("/").Methods("GET").Name(UI)

	return base
}
