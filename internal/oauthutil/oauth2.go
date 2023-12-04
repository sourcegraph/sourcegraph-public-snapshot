package oauthutil

import (
	"bytes"
	"context"
	"io"
	"net/http"

	"golang.org/x/oauth2"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Endpoint represents an OAuth 2.0 provider's authorization and token endpoint
// URLs.
type Endpoint = oauth2.Endpoint

// OAuthContext contains the configuration used in the requests to get a new
// token.
type OAuthContext struct {
	// ClientID is the application's ID.
	ClientID string
	// ClientSecret is the application's secret.
	ClientSecret string
	// Endpoint contains the resource server's token endpoint URLs.
	Endpoint Endpoint
	// Scope specifies optional requested permissions.
	Scopes []string

	// CustomQueryParams are URL parameters which may be needed by a specific provider that may have a custom OAuth2 implementation.
	CustomQueryParams map[string]string
}

// TokenRefresher is a function to refresh and return the new OAuth token.
type TokenRefresher func(ctx context.Context, doer httpcli.Doer, oauthCtx OAuthContext) (*auth.OAuthBearerToken, error)

// DoRequest is a function that uses the httpcli.Doer interface to make HTTP
// requests. It authenticates the request using the supplied Authenticator.
// If the Authenticator implements the AuthenticatorWithRefresh interface,
// it will also attempt to refresh the token in case of a 401 response.
// If the token is updated successfully, the same request will be retried exactly once.
func DoRequest(ctx context.Context, logger log.Logger, doer httpcli.Doer, req *http.Request, auther auth.Authenticator) (*http.Response, error) {
	if auther == nil {
		return doer.Do(req.WithContext(ctx))
	}

	// Try a pre-emptive token refresh in case we know it is definitely expired
	autherWithRefresh, ok := auther.(auth.AuthenticatorWithRefresh)
	if ok && autherWithRefresh.NeedsRefresh() {
		if err := autherWithRefresh.Refresh(ctx, doer); err != nil {
			logger.Warn("doRequest: token refresh failed", log.Error(err))
		}
	}

	var err error
	if err = auther.Authenticate(req); err != nil {
		return nil, errors.Wrap(err, "authenticating request")
	}

	// Store the body first in case we need to retry the request
	var reqBody []byte
	if req.Body != nil {
		reqBody, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
	}
	req.Body = io.NopCloser(bytes.NewBuffer(reqBody))
	// Do first request
	resp, err := doer.Do(req.WithContext(ctx))
	if err != nil {
		return resp, err
	}

	// If the response was unauthorised, try to refresh the token
	if resp.StatusCode == http.StatusUnauthorized && ok {
		if err = autherWithRefresh.Refresh(ctx, doer); err != nil {
			// If the refresh failed, return the original response
			return resp, nil
		}
		// Re-authorize the request and re-do the request
		if err = autherWithRefresh.Authenticate(req); err != nil {
			return nil, errors.Wrap(err, "authenticating request after token refresh")
		}
		// We need to reset the body before retrying the request
		req.Body = io.NopCloser(bytes.NewBuffer(reqBody))
		resp, err = doer.Do(req.WithContext(ctx))
	}

	return resp, err
}
