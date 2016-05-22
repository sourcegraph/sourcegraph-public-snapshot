package handlerutil

import (
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"gopkg.in/inconshreveable/log15.v2"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/util/httputil/httpctx"
	"sourcegraph.com/sqs/pbtypes"
)

type Middleware func(next http.Handler) http.Handler

func WithMiddleware(h http.Handler, mw ...Middleware) http.Handler {
	if len(mw) == 0 {
		return h
	}
	return mw[0](WithMiddleware(h, mw[1:]...))
}

// ActorMiddleware fetches the actor info and stores it in the
// context for downstream HTTP handlers. A middleware that calls
// sourcegraph.WithCredentials based on the request's auth must
// already have run for ActorMiddleware to work.
func ActorMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cl := Client(r)

		if cred := sourcegraph.CredentialsFromContext(ctx); cred != nil && fetchUserForCredentials(cred) {
			authInfo, err := cl.Auth.Identify(ctx, &pbtypes.Void{})
			if err == nil {
				ctx = auth.WithActor(ctx, auth.Actor{
					UID:   int(authInfo.UID),
					Login: authInfo.Login,
					Write: authInfo.Write,
					Admin: authInfo.Admin,
				})
				httpctx.SetForRequest(r, ctx)
			} else if err != nil {
				log15.Error("Auth.Identify in ActorMiddleware failed.", "err", err)
			}
		}

		next.ServeHTTP(w, r)
	})
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
