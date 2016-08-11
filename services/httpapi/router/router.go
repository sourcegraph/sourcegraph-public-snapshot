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
	Signup = "signup"
	Login  = "login"
	Logout = "logout"

	ForgotPassword = "forgot"
	ResetPassword  = "reset"

	Annotations              = "annotations"
	AuthInfo                 = "auth-info"
	Builds                   = "builds"
	BuildTaskLog             = "build.task.log"
	ChannelListen            = "channel.listen"
	ChannelSend              = "channel.send"
	Commit                   = "commit"
	Coverage                 = "coverage"
	Def                      = "def"
	DefRefs                  = "def.refs"
	DefRefLocations          = "def.ref.locations"
	DefExamples              = "def.examples"
	DefAuthors               = "def.authors"
	Defs                     = "defs"
	DeltaFiles               = "delta.files"
	GlobalSearch             = "global.search"
	Repo                     = "repo"
	RepoJumpDef              = "repo.jump-def"
	RepoResolve              = "repo.resolve"
	RepoCreate               = "repo.create"
	RepoRefresh              = "repo.refresh"
	RepoInventory            = "repo.inventory"
	RepoBranches             = "repo.branches"
	RepoBuild                = "repo.build"
	RepoTree                 = "repo.tree"
	RepoBuilds               = "repo.builds"
	RepoBuildTasks           = "build.tasks"
	RepoBuildsCreate         = "repo.builds.create"
	RepoCommits              = "repo.commits"
	RepoResolveRev           = "repo.resolve-rev"
	RepoTags                 = "repo.tags"
	RepoTreeList             = "repo.tree-list"
	RepoTreeSearch           = "repo-tree.search"
	RepoHoverInfo            = "repo.hover-info"
	Repos                    = "repos"
	SourcegraphDesktop       = "sourcegraph-desktop"
	SrclibImport             = "srclib.import"
	SrclibDataVer            = "srclib.data-version"
	User                     = "user"
	UserEmails               = "user.emails"
	ResolveCustomImportsInfo = "resolve-custom-import.info"
	ResolveCustomImportsTree = "resolve-custom-import.tree"

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

	base.Path("/join").Methods("POST").Name(Signup)
	base.Path("/login").Methods("POST").Name(Login)
	base.Path("/logout").Methods("POST").Name(Logout)
	base.Path("/forgot").Methods("POST").Name(ForgotPassword)
	base.Path("/reset").Methods("POST").Name(ResetPassword)

	base.Path("/beta-subscription").Methods("POST").Name(BetaSubscription)

	base.Path("/annotations").Methods("GET").Name(Annotations)

	base.Path("/builds").Methods("GET").Name(Builds)
	base.Path("/coverage").Methods("GET").Name(Coverage)

	base.Path("/repos").Methods("GET").Name(Repos)
	base.Path("/repos").Methods("POST").Name(RepoCreate)

	base.Path("/global-search").Methods("GET").Name(GlobalSearch)

	base.Path("/internal/appdash/record-span").Methods("POST").Name(InternalAppdashRecordSpan)

	base.Path("/auth-info").Methods("GET").Name(AuthInfo)
	userPath := "/users/" + routevar.User
	base.Path(userPath).Methods("GET").Name(User)
	user := base.PathPrefix(userPath + "/").Subrouter()
	user.Path("/emails").Methods("GET").Name(UserEmails)

	// repo contains routes that are NOT specific to a revision. In these routes, the URL may not contain a revspec after the repo (that is, no "github.com/foo/bar@myrevspec").
	repoPath := `/repos/` + routevar.Repo
	base.Path(repoPath).Methods("GET").Name(Repo)

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
	repoRev.Path("/tree-search").Methods("GET").Name(RepoTreeSearch)
	repoRev.Path("/tree{Path:.*}").Name(RepoTree)
	repoRev.Path("/hover-info").Methods("GET").Name(RepoHoverInfo)
	repoRev.Path("/jump-def").Methods("GET").Name(RepoJumpDef)
	repo.Path("/tags").Methods("GET").Name(RepoTags)

	repo.Path("/builds").Methods("GET").Name(RepoBuilds)
	repo.Path("/builds").Methods("POST").Name(RepoBuildsCreate)
	buildPath := `/builds/{Build:\d+}`
	repo.Path(buildPath).Methods("GET").Name(RepoBuild)
	build := repo.PathPrefix(buildPath).Subrouter()
	build.Path("/tasks").Methods("GET").Name(RepoBuildTasks)
	build.Path(`/tasks/{Task:\d+}/log`).Methods("GET").Name(BuildTaskLog)

	base.Path("/defs").Methods("GET").Name(Defs)

	repoRev.Path("/srclib-import").Methods("PUT").Name(SrclibImport)
	repoRev.Path("/srclib-data-version").Methods("GET").Name(SrclibDataVer)

	defPath := "/def/" + routevar.Def
	def := repoRev.PathPrefix(defPath + "/-/").Subrouter()
	def.Path("/authors").Methods("GET").Name(DefAuthors)
	def.Path("/refs").Methods("GET").Name(DefRefs)
	def.Path("/ref-locations").Methods("GET").Name(DefRefLocations)
	def.Path("/examples").Methods("GET").Name(DefExamples)
	repoRev.Path(defPath).Methods("GET").Name(Def) // match subroutes first

	base.Path("/channel/{Channel}").Methods("GET").Name(ChannelListen)
	base.Path("/channel/{Channel}").Methods("POST").Name(ChannelSend)

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
