package handlerutil

import (
	"net/http"

	"github.com/dgrijalva/jwt-go"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
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
