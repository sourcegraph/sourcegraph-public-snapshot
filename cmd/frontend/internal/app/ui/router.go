package ui

import (
	"strings"

	"github.com/gorilla/mux"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/routevar"
)

const (
	routeLangsIndex = "page.index.langs"
	routeReposIndex = "page.index.repos"

	routeBlob          = "page.blob"
	routeDefLanding    = "page.def.landing"
	oldRouteDefLanding = "page.def.landing.old"
	routeRepo          = "page.repo"
	routeRepoLanding   = "page.repo.landing"
	routeTree          = "page.tree"

	routeJobs           = "page.jobs"
	routeHomePage       = "page.home"
	routeAboutSubdomain = "about.sourcegraph.com" // top level redirects to about.sourcegraph.com
	routeAppPaths       = "app.paths"             // paths that are handled by our javascript app and should not 404

	routeDefRedirectToDefLanding = "page.def.redirect"
)

var router = newRouter()

func newRouter() *mux.Router {
	m := mux.NewRouter()

	m.StrictSlash(true)

	// Top level paths that should redirect to
	// about.sourcegraph.com/$PATH
	aboutPaths := []string{
		// These all omit the leading "/".
		"about",
		"plan",
		"beta",
		"beta/zap",
		"contact",
		"coverage",
		"customers/twitter",
		"docs",
		"enterprise",
		"forgot",
		"join",
		"legal",
		"pricing",
		"privacy",
		"reset",
		"security",
		"settings",
		"settings/organization",
		"styleguide",
		"terms",
		"tools",
		"tools/editor",
		"tools/browser",
		"zap/beta",
	}

	// Top level paths served by the app.
	// This is just so we don't return a 404 status.
	appPaths := []string{
		"login",
		"help",
		"help/.*",
	}
	m.Path("/sitemap").Methods("GET").Name(routeLangsIndex)
	m.Path("/sitemap/{Lang:.*}").Methods("GET").Name(routeReposIndex)
	m.Path("/{Path:(?:jobs|careers)}").Methods("GET").Name(routeJobs)
	m.Path("/{Path:(?:" + strings.Join(aboutPaths, "|") + ")}").Methods("GET").Name(routeAboutSubdomain)
	m.Path("/{Path:(?:" + strings.Join(appPaths, "|") + ")}").Methods("GET").Name(routeAppPaths)
	m.Path("/").Methods("GET").Name(routeHomePage)

	// Repo
	repoPath := "/" + routevar.Repo
	repo := m.PathPrefix(repoPath + "/" + routevar.RepoPathDelim).Subrouter()
	repo.Path("/info").Methods("GET").Name(routeRepoLanding)

	// RepoRev
	repoRevPath := repoPath + routevar.RepoRevSuffix
	m.Path(repoRevPath).Methods("GET").Name(routeRepo)
	repoRev := m.PathPrefix(repoRevPath + "/" + routevar.RepoPathDelim).Subrouter()
	repoRev.Path("/tree{Path:.*}").Methods("GET").Name(routeTree)
	repoRev.Path("/blob{Path:.*}").Methods("GET").Name(routeBlob)

	// Def
	repoRev.Path("/{dummy:def|refs}/" + routevar.Def).Methods("GET").Name(routeDefRedirectToDefLanding)
	repoRev.Path("/info/" + routevar.Def).Methods("GET").Name(routeDefLanding)
	repoRev.Path("/land/" + routevar.Def).Methods("GET").Name(oldRouteDefLanding)

	return m
}
