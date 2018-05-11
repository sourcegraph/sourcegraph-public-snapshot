package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/sourcegraph/sourcegraph/pkg/conf"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// authURLPrefix is the URL path prefix under which to attach authentication handlers
const authURLPrefix = "/.auth"

var (
	initialized   bool
	initializedMu sync.Mutex
)

// Middleware holds the middlewares that perform authentication. See NewAuthMiddleware for more
// information.
type Middleware struct {
	// API is the middleware that performs authentication on the API handler. See NewAuthMiddleware
	// for more information.
	API func(http.Handler) http.Handler

	// App is the middleware that performs authentication on the app handler. See NewAuthMiddleware
	// for more information.
	App func(http.Handler) http.Handler
}

// ComposeMiddleware returns a new Middleware that composes the middlewares together.
func ComposeMiddleware(middlewares ...*Middleware) *Middleware {
	return &Middleware{
		API: func(h http.Handler) http.Handler {
			for _, m := range middlewares {
				h = m.API(h)
			}
			return h
		},
		App: func(h http.Handler) http.Handler {
			for _, m := range middlewares {
				h = m.App(h)
			}
			return h
		},
	}
}

// NewAuthMiddleware returns middlewares that perform authentication based on the currently configured
// authentication provider (OpenID, SAML, builtin, etc.) This will expose endpoints necessary for
// the login flow and require login for all other endpoints.
//
// It returns two middlewares that have slightly different behaviors: the apiMiddleware will just
// return an HTTP 401 Unauthorized if unauthenticated; the appMiddleware will redirect the user to
// the login flow or show a nice error page if unauthenticated (depending on the auth provider)
//
// Note: this should only be called at most once (there is implicit shared state on the backend via the session store
// and the frontend via cookies). This function will return an error if called more than once.
func NewAuthMiddleware(ctx context.Context, appURL string) (*Middleware, error) {
	mw, err := createAuthMiddleware(ctx, appURL)
	if err != nil {
		return nil, err
	}

	// Prefer using middleware that is always active and enables/disables itself per-request by
	// checking config.
	//
	// TODO(sqs): Migrate all auth middlewares to this design.
	return ComposeMiddleware(mw,
		requireAuthMiddleware,
		openIDConnectAuthMiddleware,
		samlAuthMiddleware,
		&Middleware{API: httpHeaderAuthMiddleware, App: httpHeaderAuthMiddleware},
	), nil
}

func createAuthMiddleware(createCtx context.Context, appURL string) (*Middleware, error) {
	initializedMu.Lock()
	defer initializedMu.Unlock()
	if initialized {
		return nil, errors.New("NewAuthMiddleware was invoked more than once")
	}
	initialized = true

	authProvider := conf.AuthProvider()
	switch {
	case authProvider.Openidconnect != nil:
		log15.Info("SSO enabled", "protocol", "OpenID Connect")
		// The openIDConnectAuthMiddleware is always present, so no need to add it here.
		return passThrough, nil
	case authProvider.Saml != nil:
		log15.Info("SSO enabled", "protocol", "SAML 2.0")
		// The samlAuthMiddleware is always present, so no need to add it here.
		return passThrough, nil
	case authProvider.HttpHeader != nil:
		log15.Info("SSO enabled", "protocol", "HTTP proxy header")
		// The httpHeaderAuthMiddleware is always present, so no need to add it here.
		return passThrough, nil
	default:
		// The requireAuthMiddleware is always present, so no need to add it here.
		return passThrough, nil
	}
}

var passThrough = &Middleware{
	API: func(h http.Handler) http.Handler { return h },
	App: func(h http.Handler) http.Handler { return h },
}

// NormalizeUsername normalizes a proposed username into a format that meets Sourcegraph's
// username formatting rules (consistent with
// https://help.github.com/enterprise/2.11/admin/guides/user-management/using-ldap/#username-considerations-with-ldap):
//
// - Any portion of the username after a '@' character is removed
// - Any characters not in `[a-zA-Z0-9-]` are replaced with `-`
// - Usernames with consecutive '-' characters are not allowed
// - Usernames that start or end with '-' are not allowed
//
// Usernames that could not be converted return an error.
func NormalizeUsername(name string) (string, error) {
	origName := name
	if i := strings.Index(name, "@"); i != -1 && i == strings.LastIndex(name, "@") {
		name = name[:i]
	}
	name = disallowedCharacter.ReplaceAllString(name, "-")
	if strings.HasPrefix(name, "-") || strings.HasSuffix(name, "-") || strings.Index(name, "--") != -1 {
		return "", fmt.Errorf("username %q could not be normalized to acceptable format", origName)
	}
	return name, nil
}

var disallowedCharacter = regexp.MustCompile(`[^a-zA-Z0-9\-]`)

const couldNotGetUserDescription = "This occurs most frequently when there is an existing user account with the same username or email that was created from a different authentication provider."
