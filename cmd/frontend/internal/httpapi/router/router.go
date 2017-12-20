// Package router contains the URL router for the HTTP API.
package router

import (
	"github.com/gorilla/mux"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/routevar"
)

const (
	GraphQL = "graphql"
	XLang   = "xlang"
	LSP     = "lsp"

	RepoShield = "repo.shield"
	Telemetry  = "telemetry"

	DefsRefreshIndex           = "internal.defs.refresh-index"
	GitoliteUpdateRepos        = "internal.gitolite.update-repos"
	PhabricatorRepoCreate      = "internal.phabricator.repo.create"
	ReposCreateIfNotExists     = "internal.repos.create-if-not-exists"
	ReposGetByURI              = "internal.repos.get-by-uri"
	ReposInventoryUncached     = "internal.repos.inventory-uncached"
	ReposUnindexedDependencies = "internal.repos.unindexed-dependencies"
	ReposUpdateIndex           = "internal.repos.update-index"
)

// New creates a new API router with route URL pattern definitions but
// no handlers attached to the routes.
func New(base *mux.Router) *mux.Router {
	if base == nil {
		base = mux.NewRouter()
	}

	base.StrictSlash(true)

	base.Path("/graphql").Methods("GET", "POST").Name(GraphQL)
	base.Path("/xlang/{LSPMethod:.*}").Methods("POST").Name(XLang)
	base.Path("/lsp").Methods("GET").Name(LSP)

	base.Path("/telemetry/{TelemetryPath:.*}").Methods("POST").Name(Telemetry)

	// repo contains routes that are NOT specific to a revision. In these routes, the URL may not contain a revspec after the repo (that is, no "github.com/foo/bar@myrevspec").
	repoPath := `/repos/` + routevar.Repo

	// Additional paths added will be treated as a repo. To add a new path that should not be treated as a repo
	// add above repo paths.
	repo := base.PathPrefix(repoPath + "/" + routevar.RepoPathDelim + "/").Subrouter()
	repo.Path("/shield").Methods("GET").Name(RepoShield)

	return base
}

// NewInternal creates a new API router for internal endpoints.
func NewInternal(base *mux.Router) *mux.Router {
	if base == nil {
		base = mux.NewRouter()
	}
	base.StrictSlash(true)
	// Internal API endpoints should only be served on the internal Handler
	base.Path("/defs/refresh-index").Methods("POST").Name(DefsRefreshIndex)
	base.Path("/gitolite/update-repos").Methods("POST").Name(GitoliteUpdateRepos)
	base.Path("/phabricator/repo-create").Methods("POST").Name(PhabricatorRepoCreate)
	base.Path("/repos/create-if-not-exists").Methods("POST").Name(ReposCreateIfNotExists)
	base.Path("/repos/get-by-uri").Methods("POST").Name(ReposGetByURI)
	base.Path("/repos/inventory-uncached").Methods("POST").Name(ReposInventoryUncached)
	base.Path("/repos/unindexed-dependencies").Methods("POST").Name(ReposUnindexedDependencies)
	base.Path("/repos/update-index").Methods("POST").Name(ReposUpdateIndex)
	base.Path("/repos/{RepoURI:.*}").Methods("GET").Name(ReposGetByURI)

	return base
}
