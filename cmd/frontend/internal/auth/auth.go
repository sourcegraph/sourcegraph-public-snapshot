package auth

import (
	"context"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"sync"

	log15 "gopkg.in/inconshreveable/log15.v2"
)

// authURLPrefix is the URL path prefix under which to attach authentication handlers
const authURLPrefix = "/.auth"

var (
	initialized   bool
	initializedMu sync.Mutex
)

// NewSSOAuthHandler wraps the passed in handler with the appropriate authentication protocol (either OIDC or SAML)
// based on what environment variables are set. This will expose endpoints necessary for the login flow and require
// login for all other endpoints.
//
// Note: this should only be called at most once (there is implicit shared state on the backend via the session store
// and the frontend via cookies). This function will return an error if called more than once.
func NewSSOAuthHandler(createCtx context.Context, handler http.Handler, appURL string) (http.Handler, error) {
	initializedMu.Lock()
	defer initializedMu.Unlock()
	if initialized {
		return nil, errors.New("NewSSOAuthHandler was invoked more than once")
	}
	initialized = true

	if oidcIDProvider != "" {
		log15.Info("SSO enabled", "protocol", "OpenID Connect")
		return newOIDCAuthHandler(createCtx, handler, appURL)
	}
	if samlIDPMetadataURL != "" {
		log15.Info("SSO enabled", "protocol", "SAML 2.0")
		return newSAMLAuthHandler(createCtx, handler, appURL)
	}
	log15.Info("SSO disabled")
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
	if i := strings.Index(name, "@"); i != -1 && i == strings.LastIndex(name, "@") {
		name = name[:i]
	}
	name = disallowedCharacter.ReplaceAllString(name, "-")
	if strings.HasPrefix(name, "-") || strings.HasSuffix(name, "-") || strings.Index(name, "--") != -1 {
		return "", errors.New("username could not be normalized to acceptable format")
	}
	return name, nil
}

var disallowedCharacter = regexp.MustCompile(`[^a-zA-Z0-9\-]`)
