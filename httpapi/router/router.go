// Package router contains the URL router for the HTTP API.
//
// NOTE: It will likely be replaced with a codegenned router as part
// of the gRPC gateway transition.
package router

import (
	"fmt"
	"net/url"

	"github.com/sourcegraph/mux"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/routevar"
)

const (
	Annotations      = "annotations"
	BlackHole        = "blackhole"
	Build            = "build"
	Builds           = "builds"
	Def              = "def"
	Defs             = "defs"
	Repo             = "repo"
	RepoBranches     = "repo.branches"
	RepoTree         = "repo.tree"
	RepoBuild        = "repo.build"
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

	base.Path("/annotations").Methods("GET").Name(Annotations)

	base.Path("/builds").Methods("GET").Name(Builds)

	base.Path("/repos").Methods("GET").Name(Repos)

	repoRev := base.PathPrefix(`/repos/` + routevar.RepoRev).PostMatchFunc(routevar.FixRepoRevVars).BuildVarsFunc(routevar.PrepareRepoRevRouteVars).Subrouter()

	// repo contains routes that are NOT specific to a revision. In these routes, the URL may not contain a revspec after the repo (that is, no "github.com/foo/bar@myrevspec").
	repoPath := `/repos/` + routevar.Repo
	base.Path(repoPath).Methods("GET").Name(Repo)
	repo := base.PathPrefix(repoPath).Subrouter()
	repo.Path("/.branches").Methods("GET").Name(RepoBranches)
	repo.Path("/.commits").Methods("GET").Name(RepoCommits)
	repoRev.Path("/.tree" + routevar.TreeEntryPath).PostMatchFunc(routevar.FixTreeEntryVars).BuildVarsFunc(routevar.PrepareTreeEntryRouteVars).Name(RepoTree)
	repoRev.Path("/.tree-list").Methods("GET").Name(RepoTreeList)
	repoRev.Path("/.tree-search").Methods("GET").Name(RepoTreeSearch)
	repo.Path("/.tags").Methods("GET").Name(RepoTags)

	repoRev.Path("/.build").Methods("GET").Name(RepoBuild)
	repo.Path("/.builds").Methods("POST").Name(RepoBuildsCreate)
	buildPath := `/.builds/{Build:\d+}`
	repo.Path(buildPath).Methods("GET").Name(Build)
	build := repo.PathPrefix(buildPath).Subrouter()
	build.Path("/.tasks").Methods("GET").Name(RepoBuildTasks)

	base.Path("/.defs").Methods("GET").Name(Defs)

	repoRev.Path("/.srclib-import").Methods("PUT").Name(SrclibImport)
	repoRev.Path("/.srclib-coverage").Methods("PUT").Name(SrclibCoverage)
	repoRev.Path("/.srclib-data-version").Methods("GET").Name(SrclibDataVer)

	// Old paths we used to support. Explicitly handle them to avoid bad
	// signal in no route logs
	base.Path("/ext/github/webhook").Methods("GET", "POST").Name(BlackHole)

	// See router_util/def_route.go for an explanation of how we match def
	// routes.
	repoRev.Path("/" + routevar.Def).Methods("GET").PostMatchFunc(routevar.FixDefUnitVars).BuildVarsFunc(routevar.PrepareDefRouteVars).Name(Def)

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
