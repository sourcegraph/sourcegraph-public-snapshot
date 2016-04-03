// Package authutil contains authentication-related utilities for the
// app.
package authutil

import (
	"net/http"

	"sourcegraph.com/sourcegraph/sourcegraph/app/internal"
	"sourcegraph.com/sourcegraph/sourcegraph/util/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/util/httputil/httpctx"
)

func init() {
	internal.UnauthorizedErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) error {
		// Remove any user and credentials from the request context
		// to prevent any subsequent gRPC requests from hitting the
		// same unauthorized error (eg. if the token has expired).
		ctx := httpctx.FromRequest(r)
		ctx = handlerutil.ClearUser(ctx)
		httpctx.SetForRequest(r, ctx)
		return RedirectToLogIn(w, r)
	}
}

// RedirectToLogIn issues an HTTP redirect to begin the login
// process.
func RedirectToLogIn(w http.ResponseWriter, r *http.Request) error {
	http.Redirect(w, r, "/login", http.StatusSeeOther)
	return nil
}
