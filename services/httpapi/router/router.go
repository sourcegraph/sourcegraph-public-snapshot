// Package router contains the URL router for the HTTP API.
//
// NOTE: It will likely be replaced with a codegenned router as part
// of the gRPC gateway transition.
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

	GlobalSearch        = "global.search"
	RepoCreate          = "repo.create"
	RepoRefresh         = "repo.refresh"
	RepoResolveRev      = "repo.resolve-rev"
	RepoDefLanding      = "repo.def-landing"
	Repos               = "repos"
	RepoShield          = "repo.shield"
	AsyncRefreshIndexes = "async.refresh-indexes"
	BetaSubscription    = "beta-subscription"
	Orgs                = "orgs"
	OrgMembers          = "org-members"
	OrgInvites          = "org-invites"
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

	base.Path("/beta-subscription").Methods("POST").Name(BetaSubscription)
	base.Path("/orgs").Methods("POST").Name(Orgs)
	base.Path("/org-members").Methods("POST").Name(OrgMembers)
	base.Path("/org-invites").Methods("POST").Name(OrgInvites)

	base.Path("/repos").Methods("GET").Name(Repos)
	base.Path("/repos").Methods("POST").Name(RepoCreate)

	// repo contains routes that are NOT specific to a revision. In these routes, the URL may not contain a revspec after the repo (that is, no "github.com/foo/bar@myrevspec").
	repoPath := `/repos/` + routevar.Repo

	// Additional paths added will be treated as a repo. To add a new path that should not be treated as a repo
	// add above repo paths.
	repo := base.PathPrefix(repoPath + "/" + routevar.RepoPathDelim + "/").Subrouter()
	repoRev := base.PathPrefix(repoPath + routevar.RepoRevSuffix + "/" + routevar.RepoPathDelim + "/").Subrouter()
	repo.Path("/refresh").Methods("POST").Name(RepoRefresh)
	repoRev.Path("/rev").Methods("GET").Name(RepoResolveRev)
	repoRev.Path("/def-landing").Methods("GET").Name(RepoDefLanding)
	repoRev.Path("/shield").Methods("GET").Name(RepoShield)

	repo.Path("/async-refresh-indexes").Methods("POST").Name(AsyncRefreshIndexes)

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
