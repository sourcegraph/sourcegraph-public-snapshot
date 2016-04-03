// Package router contains the URL router for the HTTP API.
//
// NOTE: It will likely be replaced with a codegenned router as part
// of the gRPC gateway transition.
package router

import (
	"fmt"
	"net/url"

	"github.com/gorilla/mux"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/routevar"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/spec"
)

const (
	Signup = "signup"
	Login  = "login"
	Logout = "logout"

	ForgotPassword = "forgot"
	ResetPassword  = "reset"

	Annotations      = "annotations"
	BlackHole        = "blackhole"
	Builds           = "builds"
	BuildTaskLog     = "build.task.log"
	Def              = "def"
	DefRefs          = "def.refs"
	Defs             = "defs"
	Home             = "home"
	Repo             = "repo"
	RepoBranches     = "repo.branches"
	RepoBuild        = "repo.build"
	RepoTree         = "repo.tree"
	RepoBuilds       = "repo.builds"
	RepoBuildTasks   = "build.tasks"
	RepoBuildsCreate = "repo.builds.create"
	RepoCommits      = "repo.commits"
	RepoTags         = "repo.tags"
	RepoTreeList     = "repo.tree-list"
	RepoTreeSearch   = "repo-tree.search"
	Repos            = "repos"
	SrclibImport     = "srclib.import"
	SrclibCoverage   = "srclib.coverage"
	SrclibDataVer    = "srclib.data-version"
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

	base.Path("/annotations").Methods("GET").Name(Annotations)

	base.Path("/builds").Methods("GET").Name(Builds)

	base.Path("/repos").Methods("GET").Name(Repos)

	// repo contains routes that are NOT specific to a revision. In these routes, the URL may not contain a revspec after the repo (that is, no "github.com/foo/bar@myrevspec").
	repoPath := `/repos/` + routevar.Repo
	base.Path(repoPath).Methods("GET").Name(Repo)
	repo := base.PathPrefix(repoPath + "/" + spec.RepoPathDelim + "/").Subrouter()
	repoRev := base.PathPrefix(repoPath + routevar.RepoRevSuffix + "/" + spec.RepoPathDelim + "/").Subrouter()
	repo.Path("/branches").Methods("GET").Name(RepoBranches)
	repo.Path("/commits").Methods("GET").Name(RepoCommits) // uses Head/Base query params, not {Rev} route var
	repoRev.Path("/tree-list").Methods("GET").Name(RepoTreeList)
	repoRev.Path("/tree-search").Methods("GET").Name(RepoTreeSearch)
	repoRev.Path("/tree{Path:.*}").Name(RepoTree)
	repo.Path("/tags").Methods("GET").Name(RepoTags)

	repo.Path("/builds").Methods("GET").Name(RepoBuilds)
	repo.Path("/builds").Methods("POST").Name(RepoBuildsCreate)
	buildPath := `/builds/{Build:\d+}`
	repo.Path(buildPath).Methods("GET").Name(RepoBuild)
	build := repo.PathPrefix(buildPath).Subrouter()
	build.Path("/tasks").Methods("GET").Name(RepoBuildTasks)
	build.Path(`/tasks/{Task:\d+}/log`).Methods("GET").Name(BuildTaskLog)

	base.Path("/defs").Methods("GET").Name(Defs)

	base.Path("/home").Methods("GET").Name(Home)

	repoRev.Path("/srclib-import").Methods("PUT").Name(SrclibImport)
	repoRev.Path("/srclib-coverage").Methods("PUT").Name(SrclibCoverage)
	repoRev.Path("/srclib-data-version").Methods("GET").Name(SrclibDataVer)

	// Old paths we used to support. Explicitly handle them to avoid bad
	// signal in no route logs
	base.Path("/ext/github/webhook").Methods("GET", "POST").Name(BlackHole)

	defPath := "/def/" + routevar.Def
	def := repoRev.PathPrefix(defPath + "/-/").Subrouter()
	def.Path("/refs").Methods("GET").Name(DefRefs)
	repoRev.Path(defPath).Methods("GET").Name(Def) // match subroutes first

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
