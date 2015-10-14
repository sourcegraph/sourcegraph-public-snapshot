// Package authutil contains authentication-related utilities for the
// app.
package authutil

import (
	"net/http"

	"src.sourcegraph.com/sourcegraph/app/internal"
	"src.sourcegraph.com/sourcegraph/app/internal/returnto"
	"src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/auth/authutil"
)

func init() {
	internal.UnauthorizedErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) error {
		return RedirectToLogIn(w, r)
	}
}

// RedirectToOAuth2Initiate is a workaround to avoid an import
// cycle. It points to oauth2client.ServeOAuth2Initiate. It is
// defined as a variable here and set at init time in package
// oauth2client.
var RedirectToOAuth2Initiate func(http.ResponseWriter, *http.Request) error

// RedirectToLogIn issues an HTTP redirect to begin the login
// process. It redirects to either the OAuth2 authorization flow or
// the local login page, depending on the configuration.
func RedirectToLogIn(w http.ResponseWriter, r *http.Request) error {
	switch authutil.ActiveFlags.Source {
	case "oauth", "oauth2":
		return RedirectToOAuth2Initiate(w, r)
	case "local":
		u := router.Rel.URLTo(router.LogIn)
		returnTo, err := returnto.BestGuess(r)
		if err != nil {
			return err
		}
		returnto.SetOnURL(u, returnTo)
		http.Redirect(w, r, u.String(), http.StatusSeeOther)
		return nil
	}
	http.Error(w, "login is not enabled", http.StatusMethodNotAllowed)
	return nil
}
