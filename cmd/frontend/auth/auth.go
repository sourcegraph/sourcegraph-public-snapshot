// Package auth contains auth related code for the frontend.
package auth

import (
	"math/rand"
	"net/http"
	"strings"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/suspiciousnames"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/lib/errors"
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
// username formatting rules (based on, but not identical to
// https://web.archive.org/web/20180215000330/https://help.github.com/enterprise/2.11/admin/guides/user-management/using-ldap):
//
// - Any characters not in `[a-zA-Z0-9-._]` are replaced with `-`
// - Usernames with exactly one `@` character are interpreted as an email address, so the username will be extracted by truncating at the `@` character.
// - Usernames with two or more `@` characters are not considered an email address, so the `@` will be treated as a non-standard character and be replaced with `-`
// - Usernames with consecutive `-` or `.` characters are not allowed, so they are replaced with a single `-` or `.`
// - Usernames that start with `.` or `-` are not allowed, starting periods and dashes are removed
// - Usernames that end with `.` are not allowed, ending periods are removed
//
// Usernames that could not be converted return an error.
//
// Note: Do not forget to change database constraints on "users" and "orgs" tables.
func NormalizeUsername(name string) (string, error) {
	origName := name

	// If the username is an email address, extract the username part.
	if i := strings.Index(name, "@"); i != -1 && i == strings.LastIndex(name, "@") {
		name = name[:i]
	}

	// Replace all non-alphanumeric characters with a dash.
	name = disallowedCharacter.ReplaceAllString(name, "-")

	// Replace all consecutive dashes and periods with a single dash.
	name = consecutivePeriodsDashes.ReplaceAllString(name, "-")

	// Trim leading and trailing dashes and periods.
	name = sequencesToTrim.ReplaceAllString(name, "")

	if name == "" {
		return "", errors.Errorf("username %q could not be normalized to acceptable format", origName)
	}

	if err := suspiciousnames.CheckNameAllowedForUserOrOrganization(name); err != nil {
		return "", err
	}

	return name, nil
}

var (
	disallowedCharacter      = lazyregexp.New(`[^\w\-\.]`)
	consecutivePeriodsDashes = lazyregexp.New(`[\-\.]{2,}`)
	sequencesToTrim          = lazyregexp.New(`(^[\-\.])|(\.$)|`)
)

// AddRandomSuffix appends a random 5-character lowercase alphabetical suffix (like "-lbwwt")
// to the username to avoid collisions. If the username already ends with a dash, it is not
// added again.
func AddRandomSuffix(username string) (string, error) {
	b := make([]byte, 5)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	for i, c := range b {
		b[i] = "abcdefghijklmnopqrstuvwxyz"[c%26]
	}
	if len(username) == 0 || username[len(username)-1] == '-' {
		return username + string(b), nil
	}
	return username + "-" + string(b), nil
}

// Equivalent to `^\w(?:\w|[-.](?=\w))*-?$` which we have in the DB constraint, but without a lookahead
var validUsername = lazyregexp.New(`^\w(?:(?:[\w.-]\w|\w)*-?|)$`)

// IsValidUsername returns true if the username matches the constraints in the database.
func IsValidUsername(name string) bool {
	return validUsername.MatchString(name) && len(name) <= 255
}
