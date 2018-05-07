package auth

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

// authURLPrefix is the URL path prefix under which to attach authentication handlers
const authURLPrefix = "/.auth"

// Middleware groups two related middlewares (one for the API, one for the app).
type Middleware struct {
	// API is the middleware that performs authentication on the API handler.
	API func(http.Handler) http.Handler

	// App is the middleware that performs authentication on the app handler.
	App func(http.Handler) http.Handler
}

// Middlewares are the authentication middlewares. It is the composition of middleware for each
// authentication provider. Each middleware determines on a per-request basis whether it should be
// enabled (if not, it immediately delegates the request to the next middleware in the chain).
var Middlewares = composeMiddleware(
	requireAuthMiddleware,
	openIDConnectAuthMiddleware,
	samlAuthMiddleware,
	&Middleware{API: httpHeaderAuthMiddleware, App: httpHeaderAuthMiddleware},
	&Middleware{API: forbidAllAuthMiddleware, App: forbidAllAuthMiddleware},
)

// composeMiddleware returns a new Middleware that composes the middlewares together.
func composeMiddleware(middlewares ...*Middleware) *Middleware {
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
