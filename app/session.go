package app

import (
	"net/http"

	appauth "src.sourcegraph.com/sourcegraph/app/auth"
	"src.sourcegraph.com/sourcegraph/app/internal/tmpl"
	"src.sourcegraph.com/sourcegraph/util/eventsutil"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
)

func serveLogOut(w http.ResponseWriter, r *http.Request) error {
	appauth.DeleteSessionCookie(w)

	ctx := httpctx.FromRequest(r)
	eventsutil.LogSignOut(ctx)

	currentUser := handlerutil.UserFromContext(ctx)
	if currentUser != nil {
		// If a user was logged in prior to navigating here, clear the user in
		// the request context so that we don't show the logout page with the
		// user's info.
		ctx = handlerutil.ClearUser(ctx)
		httpctx.SetForRequest(r, ctx)
	}

	return tmpl.Exec(r, w, "user/logged_out.html", http.StatusOK, nil, &struct {
		tmpl.Common
	}{})
}
