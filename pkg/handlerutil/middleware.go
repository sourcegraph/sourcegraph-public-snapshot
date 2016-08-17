package handlerutil

import (
	"net/http"
	"regexp"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/csrf"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/app/auth"
)

type Middleware func(next http.Handler) http.Handler

func WithMiddleware(h http.Handler, mw ...Middleware) http.Handler {
	if len(mw) == 0 {
		return h
	}
	return mw[0](WithMiddleware(h, mw[1:]...))
}

// fetchUserForCredentials is whether ActorMiddleware should try to
// fetch the actor, given the specified credentials. It returns true
// if cred represents a user. If it just represents an authed client
// (or nothing), it returns false.
func fetchUserForCredentials(cred sourcegraph.Credentials) bool {
	tok0, err := cred.Token()
	if err != nil {
		// Return true so it tries to use these creds and deletes them
		// from the session if they are invalid.
		return true
	}
	tok, _ := jwt.Parse(tok0.AccessToken, func(*jwt.Token) (interface{}, error) { return nil, nil })
	if tok == nil {
		return false
	}
	_, hasUID := tok.Claims["UID"]
	return hasUID
}

var skipCSRFPattern = regexp.MustCompile("^/login/oauth/|git-[\\w-]+$")

// NewHandlerWithCSRFProtection creates a new handler that uses the provided
// handler. It additionally adds support for cross-site request forgery. To make
// your forms compliant you will have to include a hidden input which contains
// the CSRFToken that is made available to you in the template via tmpl.Common.
//
// Example:
// 	<input type="hidden" name="csrf_token" value="{{$.CSRFToken}}">
//
func NewHandlerWithCSRFProtection(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if skipCSRFPattern.MatchString(r.URL.Path) {
			handler.ServeHTTP(w, r)
			return
		}

		p := csrf.Protect(
			[]byte("e953612ddddcdd5ec60d74e07d40218c"),
			csrf.CookieName("csrf_token"),
			csrf.Secure(auth.OnlySecureCookies(r.Context())),
		)
		p(handler).ServeHTTP(w, r)
	})
}
