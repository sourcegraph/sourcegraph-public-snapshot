// Package router is a URL router for the app UI handlers.
package router

import (
	"github.com/gorilla/mux"
	app_router "sourcegraph.com/sourcegraph/sourcegraph/app/router"
)

const (
	References = "def.refs"

	AppdashUploadPageLoad = "appdash.upload-page-load"

	UserInviteBulk = "user.invite.bulk"
)

func New(base *mux.Router) *mux.Router {
	if base == nil {
		base = mux.NewRouter()
	}

	base.StrictSlash(true)

	base.Path("/.appdash/upload-page-load").
		Methods("POST").
		Name(AppdashUploadPageLoad)

	return base
}

// Rel is a relative url router, used for tests.
var Rel = app_router.Router{Router: *New(nil)}
