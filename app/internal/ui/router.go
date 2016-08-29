package ui

import (
	"strings"

	"github.com/gorilla/mux"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/routevar"
)

const (
	routeBlob        = "blob"
	routeBuild       = "build"
	routeDef         = "def"
	routeDefInfo     = "def.info"
	routeDefLanding  = "def.landing"
	routeRepo        = "repo"
	routeRepoBuilds  = "repo.builds"
	routeRepoLanding = "repo.landing"
	routeTree        = "tree"

	routeTopLevel = "toplevel" // non-repo top-level routes
)

var router = newRouter()

func newRouter() *mux.Router {
	m := mux.NewRouter()

	m.StrictSlash(true)

	// Special top-level routes that do NOT refer to repos.
	//
	// NOTE: Keep in sync with routePatterns.tsx. See the NOTE in that
	// file for more information.
	topLevel := []string{
		// These all omit the leading "/".
		"about",
		"about/browser-ext-faqs",
		"beta",
		"contact",
		"coverage",
		"forgot",
		"join",
		"legal",
		"login",
		"pricing",
		"privacy",
		"settings/repos",
		"reset",
		"search",
		"security",
		"styleguide",
		"terms",
		"tools",
		"tools/editor",
		"tools/browser",
		"tools",
	}
	m.Path("/{Path:(?:" + strings.Join(topLevel, "|") + ")}").Methods("GET").Name(routeTopLevel)

	// Repo
	repoPath := "/" + routevar.Repo
	repo := m.PathPrefix(repoPath + "/" + routevar.RepoPathDelim).Subrouter()
	repo.Path("/builds").Methods("GET").Name(routeRepoBuilds)
	repo.Path(`/builds/{Build:\d+}`).Methods("GET").Name(routeBuild)

	// RepoRev
	repoRevPath := repoPath + routevar.RepoRevSuffix
	m.Path(repoRevPath).Methods("GET").Name(routeRepo)
	repoRev := m.PathPrefix(repoRevPath + "/" + routevar.RepoPathDelim).Subrouter()
	repoRev.Path("/tree{Path:.*}").Methods("GET").Name(routeTree)
	repoRev.Path("/blob{Path:.*}").Methods("GET").Name(routeBlob)

	// Def
	repoRev.Path("/def/" + routevar.Def).Methods("GET").Name(routeDef)
	repoRev.Path("/info/" + routevar.Def).Methods("GET").Name(routeDefInfo)
	repoRev.Path("/land/" + routevar.Def).Methods("GET").Name(routeDefLanding)
	repoRev.Path("/land").Methods("GET").Name(routeRepoLanding)

	return m
}
