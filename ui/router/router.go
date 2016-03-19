// Package router is a URL router for the app UI handlers.
package router

import (
	"github.com/sourcegraph/mux"
	app_router "sourcegraph.com/sourcegraph/sourcegraph/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/routevar"
)

const (
	References            = "def.refs"
	AppdashUploadPageLoad = "appdash.upload-page-load"

	UserContentUpload = "usercontent.upload"

	UserInviteBulk = "user.invite.bulk"
)

func New(base *mux.Router) *mux.Router {
	if base == nil {
		base = mux.NewRouter()
	}

	base.StrictSlash(true)

	repoRev := base.PathPrefix(`/` + routevar.RepoRev).
		PostMatchFunc(routevar.FixRepoRevVars).
		BuildVarsFunc(routevar.PrepareRepoRevRouteVars).
		Subrouter()

	def := repoRev.PathPrefix("/" + routevar.Def).
		PostMatchFunc(routevar.FixDefUnitVars).
		BuildVarsFunc(routevar.PrepareDefRouteVars).
		Subrouter()

	def.Path("/.refs").
		Methods("GET").
		Name(References)

	base.Path("/.appdash/upload-page-load").
		Methods("POST").
		Name(AppdashUploadPageLoad)

	base.Path("/.usercontent").
		Methods("POST").
		Name(UserContentUpload)

	base.Path("/.invite-bulk").
		Methods("POST").
		Name(UserInviteBulk)

	return base
}

// Rel is a relative url router, used for tests.
var Rel = app_router.Router{Router: *New(nil)}
