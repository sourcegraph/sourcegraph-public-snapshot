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

	Discussion         = "discussion"
	DiscussionListDef  = "discussion.list.def"
	DiscussionListRepo = "discussion.list.repo"
	DiscussionCreate   = "discussion.create"
	DiscussionComment  = "discussion.comment"

	SearchTokens = "search.tokens"
	SearchText   = "search.text"

	AppdashUploadPageLoad = "appdash.upload-page-load"

	UserContentUpload = "usercontent.upload"
)

func New(base *mux.Router, isTest bool) *mux.Router {
	if base == nil {
		base = mux.NewRouter()
	}

	// m augments a list of HTTP request methods with an additional "POST" method
	// (in case it doesn't already exist) for testing.
	m := func(methods ...string) []string {
		if !isTest {
			return methods
		}
		for _, mm := range methods {
			if mm == "POST" {
				return methods
			}
		}
		return append(methods, "POST")
	}

	base.StrictSlash(true)

	repo := base.PathPrefix(`/` + routevar.Repo).Subrouter()

	repoRevPath := `/` + routevar.RepoRev
	repoRev := base.PathPrefix(repoRevPath).
		PostMatchFunc(routevar.FixRepoRevVars).
		BuildVarsFunc(routevar.PrepareRepoRevRouteVars).
		Subrouter()

	repoRev.Path("/.tree" + routevar.TreeEntryPath).
		Methods(m("GET")...).
		PostMatchFunc(routevar.FixTreeEntryVars).
		BuildVarsFunc(routevar.PrepareTreeEntryRouteVars).
		Name(RepoTree)

	repoRev.Path("/.filefinder").
		Methods(m("GET")...).
		Name(RepoFileFinder)

	defPath := "/" + routevar.Def

	repoRev.Path(defPath).
		Methods(m("GET")...).
		PostMatchFunc(routevar.FixDefUnitVars).
		BuildVarsFunc(routevar.PrepareDefRouteVars).
		Name(Definition)

	def := repoRev.PathPrefix(defPath).
		PostMatchFunc(routevar.FixDefUnitVars).
		BuildVarsFunc(routevar.PrepareDefRouteVars).
		Subrouter()

	def.Path("/.examples").
		Methods(m("GET")...).
		Name(DefExamples)

	def.Path("/.popover").
		Methods(m("GET")...).
		Name(DefPopover)

	repo.Path("/.discussion/{ID:[0-9]+}").
		Methods(m("GET")...).
		Name(Discussion)

	repo.Path("/.discussions").
		Methods(m("GET")...).
		Name(DiscussionListRepo)

	def.Path("/.discussions").
		Methods(m("GET")...).
		Name(DiscussionListDef)

	def.Path("/.discussions/create").
		Methods(m("POST")...).
		Name(DiscussionCreate)

	def.Path("/.discussions/{ID:[0-9]+}/.comment").
		Methods(m("POST")...).
		Name(DiscussionComment)

	repoRev.Path("/.search/tokens").
		Methods(m("GET")...).
		Name(SearchTokens)

	repoRev.Path("/.search/text").
		Methods(m("GET")...).
		Name(SearchText)

	base.Path("/.appdash/upload-page-load").
		Methods(m("POST")...).
		Name(AppdashUploadPageLoad)

	base.Path("/.usercontent").
		Methods(m("POST")...).
		Name(UserContentUpload)

	return base
}

// Rel is a relative url router, used for tests.
var Rel = app_router.Router{Router: *New(nil, false)}
