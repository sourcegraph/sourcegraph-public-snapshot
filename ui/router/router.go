// Package router is a URL router for the app UI handlers.
package router

import (
	"github.com/sourcegraph/mux"
	app_router "src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/routevar"
)

const (
	RepoTree = "repo.tree"

	RepoFileFinder = "repo.file.finder"

	Definition  = "def"
	DefPopover  = "def.popover"
	DefExamples = "def.examples"

	RepoCommits = "repo.commits"

	SearchTokens = "search.tokens"
	SearchText   = "search.text"

	AppdashUploadPageLoad = "appdash.upload-page-load"

	UserContentUpload = "usercontent.upload"

	UserInvite = "user.invite"
	UserKeys   = "user.keys"
)

func New(base *mux.Router) *mux.Router {
	if base == nil {
		base = mux.NewRouter()
	}

	base.StrictSlash(true)

	repoRevPath := `/` + routevar.RepoRev
	repoRev := base.PathPrefix(repoRevPath).
		PostMatchFunc(routevar.FixRepoRevVars).
		BuildVarsFunc(routevar.PrepareRepoRevRouteVars).
		Subrouter()

	repoRev.Path("/.tree" + routevar.TreeEntryPath).
		Methods("GET").
		PostMatchFunc(routevar.FixTreeEntryVars).
		BuildVarsFunc(routevar.PrepareTreeEntryRouteVars).
		Name(RepoTree)

	repoRev.Path("/.filefinder").
		Methods("GET").
		Name(RepoFileFinder)

	defPath := "/" + routevar.Def

	repoRev.Path(defPath).
		Methods("GET").
		PostMatchFunc(routevar.FixDefUnitVars).
		BuildVarsFunc(routevar.PrepareDefRouteVars).
		Name(Definition)

	def := repoRev.PathPrefix(defPath).
		PostMatchFunc(routevar.FixDefUnitVars).
		BuildVarsFunc(routevar.PrepareDefRouteVars).
		Subrouter()

	def.Path("/.examples").
		Methods("GET").
		Name(DefExamples)

	def.Path("/.popover").
		Methods("GET").
		Name(DefPopover)

	repoRev.Path("/.search/tokens").
		Methods("GET").
		Name(SearchTokens)

	repoRev.Path("/.search/text").
		Methods("GET").
		Name(SearchText)

	repo := base.PathPrefix(`/` + routevar.Repo).Subrouter()

	repo.Path("/.commits").
		Methods("GET").
		Name(RepoCommits)

	base.Path("/.appdash/upload-page-load").
		Methods("POST").
		Name(AppdashUploadPageLoad)

	base.Path("/.usercontent").
		Methods("POST").
		Name(UserContentUpload)

	base.Path("/.invite").
		Methods("POST").
		Name(UserInvite)

	base.Path("/.user/keys").
		Methods("POST", "GET", "DELETE").
		Name(UserKeys)

	return base
}

// Rel is a relative url router, used for tests.
var Rel = app_router.Router{Router: *New(nil)}
