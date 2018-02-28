package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"

	log15 "gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
)

// authURLPrefix is the URL path prefix under which to attach authentication handlers
const authURLPrefix = "/.auth"

var (
	initialized   bool
	initializedMu sync.Mutex
)

// NewAuthHandler wraps the passed in handler with the appropriate authentication protocol (either OIDC or SAML)
// based on what environment variables are set. This will expose endpoints necessary for the login flow and require
// login for all other endpoints.
//
// Note: this should only be called at most once (there is implicit shared state on the backend via the session store
// and the frontend via cookies). This function will return an error if called more than once.
func NewAuthHandler(createCtx context.Context, handler http.Handler, appURL string) (http.Handler, error) {
	initializedMu.Lock()
	defer initializedMu.Unlock()
	if initialized {
		return nil, errors.New("NewAuthHandler was invoked more than once")
	}
	initialized = true

	if oidcProvider != nil {
		log15.Info("SSO enabled", "protocol", "OpenID Connect")
		return newOIDCAuthHandler(createCtx, handler, appURL)
	}
	if samlProvider != nil {
		log15.Info("SSO enabled", "protocol", "SAML 2.0")
		return newSAMLAuthHandler(createCtx, handler, appURL)
	}

	// Note: auth.public should only have an effect when auth.provider == "builtin".
	//
	// This is important to check here because there is no
	// auth.provider == "http-header" case above, and requiring builtin auth
	// for that provider would effectively break it.
	if conf.GetTODO().AuthProvider == "builtin" && !conf.GetTODO().AuthPublic {
		return newUserRequiredAuthzHandler(handler), nil
	}

	return handler, nil
}

// NormalizeUsername normalizes a proposed username into a format that meets Sourcegraph's
// username formatting rules (consistent with
// https://help.github.com/enterprise/2.11/admin/guides/user-management/using-ldap/#username-considerations-with-ldap):
//
// - Email address after the '@' is discarded
// - Any characters that are non-alphanumeric and not a dash are coverted to dashes
// - Usernames with consecutive dashes are not allowed
// - Usernames that start or end with a dash are not allowed
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
