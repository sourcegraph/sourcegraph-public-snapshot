package auth

import (
	"context"
	"errors"
	"net/http"
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
