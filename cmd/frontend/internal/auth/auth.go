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
func NewAuthMiddleware(createCtx context.Context, appURL string) (*Middleware, error) {
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
		return newOIDCAuthMiddleware(createCtx, appURL, authProvider.Openidconnect)
	case authProvider.Saml != nil:
		log15.Info("SSO enabled", "protocol", "SAML 2.0")
		return newSAMLAuthMiddleware(createCtx, appURL, authProvider.Saml)
	case authProvider.HttpHeader != nil:
		log15.Info("SSO enabled", "protocol", "HTTP proxy header")
		// Same behavior for API and app.
		return &Middleware{API: httpHeaderAuthMiddleware, App: httpHeaderAuthMiddleware}, nil
	default:
		if conf.GetTODO().AuthPublic {
			// No auth is required.
			passThrough := func(h http.Handler) http.Handler { return h }
			return &Middleware{API: passThrough, App: passThrough}, nil
		}
		return newUserRequiredAuthzMiddleware(), nil
	}
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
