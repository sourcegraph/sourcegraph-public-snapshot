package app

import (
	"fmt"
	"html"
	"net/http"

	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/oauth2client"
	"sourcegraph.com/sourcegraph/sourcegraph/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
)

// serveShowAuth prints the user's auth cookie. Local tools that need
// to access Sourcegraph as a certain user can direct the user to
// http://sourcegraph.com/-/show-auth to let the user obtain their own
// auth cookie.
func serveShowAuth(w http.ResponseWriter, r *http.Request) error {
	sessionCookie := auth.SessionCookie(r)

	// If the user isn't logged in, redirect them to log in and come
	// back here.
	if sessionCookie == "" {
		returnTo := router.Rel.URLTo(router.ShowAuth)
		return oauth2client.GitHubOAuth2Initiate(w, r, nil, returnTo.String(), returnTo.String())
	}

	// sessionCookie is the value of the user's cookie. If an attacker
	// can set the user's cookie to an arbitrary value, then they
	// could inject content into this HTTP response. We mitigate this
	// risk by (1) setting the Content-Type to text/plain to reduce
	// the chance it's interpreted as a script, and (2) HTML-escaping
	// the string. In practice, if an attacker could SET a user's
	// cookie, they could probably do much more damage than anything
	// this endpoint would allow.
	w.Header().Set("content-type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "To authenticate with Sourcegraph:\n\nexport ZAP_AUTH_COOKIE=sg-session=%s", html.EscapeString(sessionCookie))
	return nil
}
