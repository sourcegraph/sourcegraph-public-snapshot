// Package router contains the URL router for the HTTP API.
package router

import (
	"fmt"
	"net/url"

	"github.com/gorilla/mux"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/routevar"
)

const (
	GraphQL = "graphql"
	XLang   = "xlang"
	LSP     = "lsp"

	GlobalSearch     = "global.search"
	RepoShield       = "repo.shield"
	BetaSubscription = "beta-subscription"
	SubmitForm       = "submit-form"
	Telemetry        = "telemetry"

	DefsRefreshIndex           = "internal.defs.refresh-index"
	ReposGetByURI              = "internal.repos.get-by-uri"
	ReposCreateIfNotExists     = "internal.repos.create-if-not-exists"
	ReposUpdateIndex           = "internal.repos.update-index"
	ReposUnindexedDependencies = "internal.repos.unindexed-dependencies"
	ReposInventoryUncached     = "internal.repos.inventory-uncached"
	PhabricatorRepoCreate      = "internal.phabricator.repo.create"
	GitoliteUpdateRepos        = "internal.gitolite.update-repos"
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

	base.Path("/beta-subscription").Methods("POST").Name(BetaSubscription)
	base.Path("/submit-form").Methods("POST").Name(SubmitForm)

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
	base.Path("/repos/get-by-uri").Methods("POST").Name(ReposGetByURI)
	base.Path("/repos/update-index").Methods("POST").Name(ReposUpdateIndex)
	base.Path("/repos/create-if-not-exists").Methods("POST").Name(ReposCreateIfNotExists)
	base.Path("/repos/unindexed-dependencies").Methods("POST").Name(ReposUnindexedDependencies)
	base.Path("/repos/inventory-uncached").Methods("POST").Name(ReposInventoryUncached)
	base.Path("/repos/{RepoURI:.*}").Methods("GET").Name(ReposGetByURI)

	return base
}

var rel = New(nil)

// URL generates a relative URL for the given route, route variables,
// and querystring options. The returned URL will contain only path
// and querystring components (and will not be an absolute URL).
func URL(route string, routeVars map[string]string) (*url.URL, error) {
	rt := rel.Get(route)
	if rt == nil {
		return nil, fmt.Errorf("no API route named %q", route)
	}

	routeVarsList := make([]string, 2*len(routeVars))
	i := 0
	for name, val := range routeVars {
		routeVarsList[i*2] = name
		routeVarsList[i*2+1] = val
		i++
	}
	url, err := rt.URL(routeVarsList...)
	if err != nil {
		return nil, err
	}

	return url, nil
}
