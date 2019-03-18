package auth

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/suspiciousnames"
)

// AuthURLPrefix is the URL path prefix under which to attach authentication handlers
const AuthURLPrefix = "/.auth"

// Middleware groups two related middlewares (one for the API, one for the app).
type Middleware struct {
	// API is the middleware that performs authentication on the API handler.
	API func(http.Handler) http.Handler

	// App is the middleware that performs authentication on the app handler.
	App func(http.Handler) http.Handler
}

var extraAuthMiddlewares []*Middleware

// RegisterMiddlewares registers additional authentication middlewares. Currently this is used to
// register enterprise-only SSO middleware. This should only be called from an init function.
func RegisterMiddlewares(m ...*Middleware) {
	extraAuthMiddlewares = append(extraAuthMiddlewares, m...)
}

// AuthMiddleware returns the authentication middleware that combines all authentication middlewares
// that have been registered.
func AuthMiddleware() *Middleware {
	m := make([]*Middleware, 0, 1+len(extraAuthMiddlewares))
	m = append(m, RequireAuthMiddleware)
	m = append(m, extraAuthMiddlewares...)
	return composeMiddleware(m...)
}

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
	if strings.HasPrefix(name, "-") || strings.HasSuffix(name, "-") || strings.Contains(name, "--") {
		return "", fmt.Errorf("username %q could not be normalized to acceptable format", origName)
	}
	if err := suspiciousnames.CheckNameAllowedForUserOrOrganization(name); err != nil {
		return "", err
	}
	return name, nil
}

var disallowedCharacter = regexp.MustCompile(`[^a-zA-Z0-9\-]`)
