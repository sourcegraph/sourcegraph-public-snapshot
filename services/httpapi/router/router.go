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
	Annotations              = "annotations"
	AuthInfo                 = "auth-info"
	Commit                   = "commit"
	Coverage                 = "coverage"
	DefLocalRefLocations     = "def.local.ref.locations"
	DeltaFiles               = "delta.files"
	GitHubToken              = "github-token"
	GlobalSearch             = "global.search"
	Repo                     = "repo"
	RepoJumpDef              = "repo.jump-def"
	RepoSymbols              = "repo.symbols"
	RepoResolve              = "repo.resolve"
	RepoCreate               = "repo.create"
	RepoRefresh              = "repo.refresh"
	RepoInventory            = "repo.inventory"
	RepoBranches             = "repo.branches"
	RepoTree                 = "repo.tree"
	RepoCommits              = "repo.commits"
	RepoResolveRev           = "repo.resolve-rev"
	RepoTags                 = "repo.tags"
	RepoTreeList             = "repo.tree-list"
	RepoHoverInfo            = "repo.hover-info"
	RepoWebhookEnable        = "repo.webhook-enable"
	RepoWebhookCallback      = "repo.webhook-callback"
	Repos                    = "repos"
	SourcegraphDesktop       = "sourcegraph-desktop"
	SrclibImport             = "srclib.import"
	SrclibDataVer            = "srclib.data-version"
	AsyncRefreshIndexes      = "async.refresh-indexes"
	ResolveCustomImportsInfo = "resolve-custom-import.info"
	ResolveCustomImportsTree = "resolve-custom-import.tree"

	XLang = "xlang"

	InternalAppdashRecordSpan = "internal.appdash.record-span"
	BetaSubscription          = "beta-subscription"
)

// New creates a new API router with route URL pattern definitions but
// no handlers attached to the routes.
func New(base *mux.Router) *mux.Router {
	if base == nil {
		base = mux.NewRouter()
	}

	base.StrictSlash(true)

	base.Path("/beta-subscription").Methods("POST").Name(BetaSubscription)

	base.Path("/annotations").Methods("GET").Name(Annotations)

	base.Path("/coverage").Methods("GET").Name(Coverage)

	base.Path("/repos").Methods("GET").Name(Repos)
	base.Path("/repos").Methods("POST").Name(RepoCreate)
	base.Path("/webhook/enable").Methods("GET").Name(RepoWebhookEnable)
	base.Path("/webhook/callback").Methods("POST").Name(RepoWebhookCallback)

	base.Path("/global-search").Methods("GET").Name(GlobalSearch)

	base.Path("/internal/appdash/record-span").Methods("POST").Name(InternalAppdashRecordSpan)

	base.Path("/auth-info").Methods("GET").Name(AuthInfo)
	base.Path("/github-token").Methods("GET").Name(GitHubToken)

	// repo contains routes that are NOT specific to a revision. In these routes, the URL may not contain a revspec after the repo (that is, no "github.com/foo/bar@myrevspec").
	repoPath := `/repos/` + routevar.Repo
	base.Path(repoPath).Methods("GET").Name(Repo)

	base.Path("/xlang/{LSPMethod:.*}").Methods("POST").Name(XLang)

	base.Path("/sourcegraph-desktop/").Methods("GET").Name(SourcegraphDesktop)
	// Additional paths added will be treated as a repo. To add a new path that should not be treated as a repo
	// add above repo paths.
	repo := base.PathPrefix(repoPath + "/" + routevar.RepoPathDelim + "/").Subrouter()
	repoRev := base.PathPrefix(repoPath + routevar.RepoRevSuffix + "/" + routevar.RepoPathDelim + "/").Subrouter()
	repo.Path("/resolve").Methods("GET").Name(RepoResolve)
	repo.Path("/refresh").Methods("POST").Name(RepoRefresh)
	repo.Path("/branches").Methods("GET").Name(RepoBranches)
	repo.Path("/commits").Methods("GET").Name(RepoCommits) // uses Head/Base query params, not {Rev} route var
	repoRev.Path("/tree-list").Methods("GET").Name(RepoTreeList)
	repoRev.Path("/rev").Methods("GET").Name(RepoResolveRev)
	repoRev.Path("/commit").Methods("GET").Name(Commit)
	repoRev.Path("/delta/{DeltaBaseRev}/-/files").Methods("GET").Name(DeltaFiles)
	repoRev.Path("/inventory").Methods("GET").Name(RepoInventory)
	repoRev.Path("/tree{Path:.*}").Name(RepoTree)
	repoRev.Path("/hover-info").Methods("GET").Name(RepoHoverInfo)
	repoRev.Path("/jump-def").Methods("GET").Name(RepoJumpDef)
	repoRev.Path("/symbols").Methods("GET").Name(RepoSymbols)
	repo.Path("/tags").Methods("GET").Name(RepoTags)

	repoRev.Path("/srclib-import").Methods("PUT").Name(SrclibImport)
	repoRev.Path("/srclib-data-version").Methods("GET").Name(SrclibDataVer)
	repo.Path("/async-refresh-indexes").Methods("POST").Name(AsyncRefreshIndexes)

	defPath := "/def/" + routevar.Def
	def := repoRev.PathPrefix(defPath + "/-/").Subrouter()
	def.Path("/local-refs").Methods("GET").Name(DefLocalRefLocations)

	base.Path("/resolve-custom-import/info").Methods("GET").Name(ResolveCustomImportsInfo)
	base.Path("/resolve-custom-import/tree").Methods("GET").Name(ResolveCustomImportsTree)

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
