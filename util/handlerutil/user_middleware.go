package handlerutil

import (
	"log"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sqs/pbtypes"
	appauth "src.sourcegraph.com/sourcegraph/app/auth"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
)

// This file contains getters and middleware setters for data that
// should be in the context during HTTP handler execution.

type contextKey int

const (
	userKey contextKey = iota
)

// UserMiddleware fetches the user object and stores it in the context
// for downstream HTTP handlers. The CookieMiddleware must already
// have run (or something else that calls sourcegraph.WithCredentials
// based on the request's auth).
func UserMiddleware(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	ctx := httpctx.FromRequest(r)

	cred := sourcegraph.CredentialsFromContext(ctx)
	if cred != nil && UserFromRequest(r) == nil && fetchUserForCredentials(cred) {
		cl := sourcegraph.NewClientFromContext(ctx)

		// Call to Identify will be authenticated with the
		// session's access token (because of previous middleware).
		authInfo, err := cl.Auth.Identify(ctx, &pbtypes.Void{})
		if err != nil {
			log.Printf("warning: identifying current user failed: %s (continuing, deleting cookie)", err)
			appauth.DeleteSessionCookie(w)
		} else if authInfo.UID != 0 {
			// Fetch user.
			user, err := cl.Users.Get(ctx, authInfo.UserSpec())
			if err == nil {
				ctx = WithUser(ctx, user)
			} else if grpc.Code(err) != codes.Unimplemented && grpc.Code(err) != codes.Unauthenticated {
				log.Printf("warning: fetching user failed: %s (continuing, deleting cookie)", err)
				appauth.DeleteSessionCookie(w)
			}
		} else {
			// The cookie was probably created by another server; delete it.
			log.Printf("warning: credentials don't identify a user on this server (continuing, deleting cookie)")
			appauth.DeleteSessionCookie(w)
		}
	}

	httpctx.SetForRequest(r, ctx)
	next(w, r)
}

// fetchUserForCredentials is whether UserMiddleware should try to
// fetch the user object, given the specified credentials. It returns
// true if cred represents a user. If it just represents an authed
// client (or nothing), it returns false.
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

// UserFromContext returns the request's context's authenticated user (if
// any).
//
// TODO(sqs): rename to CurrentUser or something -- "FromContext"
// implies it takes a context.Context, which it does not.
func UserFromRequest(r *http.Request) *sourcegraph.User {
	return UserFromContext(httpctx.FromRequest(r))
}

// userFromContext returns the context's authenticated user (if any).
func UserFromContext(ctx context.Context) *sourcegraph.User {
	user, _ := ctx.Value(userKey).(*sourcegraph.User)
	return user
}

// WithUser returns a copy of the context with the user added to it
// (and available via UserFromContext). Generally you should use
// UserMiddleware to set it in the context; WithUser is probably most
// useful for tests where you want to inject a specific user.
func WithUser(ctx context.Context, user *sourcegraph.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}
