// Package router is a URL router for the app UI handlers.
package router

import (
	"log"
	"net/url"
	"os"

	"github.com/sourcegraph/mux"
	"sourcegraph.com/sourcegraph/go-sourcegraph/routevar"
)

const (
	RepoTree = "repo.tree"

	RepoFileFinder = "repo.file.finder"

	ChangesetCreate       = "repo.changeset.create"
	ChangesetSubmitReview = "repo.changeset.submit-review"
	ChangesetUpdate       = "repo.changeset.update"

	Definition  = "def"
	DefList     = "def.list"
	DefPopover  = "def.popover"
	DefExamples = "def.examples"

	Discussion         = "discussion"
	DiscussionListDef  = "discussion.list.def"
	DiscussionListRepo = "discussion.list.repo"
	DiscussionCreate   = "discussion.create"
	DiscussionComment  = "discussion.comment"

	SearchTokens = "search.tokens"

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

	base.Path("/.defs").
		Methods(m("GET")...).
		Name(DefList)

	repo := base.PathPrefix(`/` + routevar.Repo).Subrouter()

	repo.Path("/.changesets/create").
		Methods(m("POST")...).
		Name(ChangesetCreate)

	repo.Path(`/.changesets/{ID:\d+}/submit-review`).
		Methods(m("POST")...).
		Name(ChangesetSubmitReview)

	repo.Path(`/.changesets/{ID:\d+}/update`).
		Methods(m("POST")...).
		Name(ChangesetUpdate)

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

	base.Path("/.appdash/upload-page-load").
		Methods(m("POST")...).
		Name(AppdashUploadPageLoad)

	base.Path("/.usercontent").
		Methods(m("POST")...).
		Name(UserContentUpload)

	return base
}

func urlToOrError(r *mux.Router, routeName string, params ...string) (*url.URL, error) {
	route := r.Get(routeName)
	if route == nil {
		log.Panicf("no such route: %q (params: %v)", routeName, params)
	}
	u, err := route.URL(params...)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func urlTo(r *mux.Router, routeName string, params ...string) *url.URL {
	u, err := urlToOrError(r, routeName, params...)
	if err != nil {
		if os.Getenv("STRICT_URL_GEN") != "" && *u == (url.URL{}) {
			log.Panicf("Failed to generate route. See log message above.")
		}
		log.Printf("Route error: failed to make URL for route %q (params: %v): %s", routeName, params, err)
		return &url.URL{}
	}
	return u
}

var rel = New(nil, false)

// RelURLTo, used for testing, returns a relative url to given route name.
func RelURLTo(routeName string, params ...string) *url.URL {
	return urlTo(rel, routeName, params...)
}
