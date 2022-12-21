package oauthutil

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/oauth2"

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
}

type oauthError struct {
	Err              string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func (e oauthError) Error() string {
	return fmt.Sprintf("OAuth response error %q description %q", e.Err, e.ErrorDescription)
}

// getOAuthErrorDetails is a method that only returns OAuth errors.
// It is intended to be used in the oauth flow, when refreshing an expired token.
func getOAuthErrorDetails(body []byte) error {
	var oe oauthError
	if err := json.Unmarshal(body, &oe); err != nil {
		// If we failed to unmarshal body with oauth error, it's not oauthError and we should return nil.
		return nil
	}

	// https://www.oauth.com/oauth2-servers/access-tokens/access-token-response/
	// {"error":"invalid_token","error_description":"Token is expired. You can either do re-authorization or token refresh."}
	if oe.Err == "invalid_token" && strings.Contains(oe.ErrorDescription, "expired") {
		return &oe
	}
	return nil
}

// TokenRefresher is a function to refresh and return the new OAuth token.
type TokenRefresher func(ctx context.Context, doer httpcli.Doer, oauthCtx OAuthContext) (*auth.OAuthBearerToken, error)

// DoRequest is a function that uses the httpcli.Doer interface to make HTTP
// requests and to handle "401 Unauthorized" errors. When the 401 error is due to
// a token being expired, it will use the supplied TokenRefresher function to
// update the token. If the token is updated successfully, the same request will
// be retried exactly once.
// autherWithRefresh should not be nil.
func DoRequest(ctx context.Context, doer httpcli.Doer, req *http.Request, autherWithRefresh auth.AuthenticatorWithRefresh) (resp *http.Response, err error) {
	// Do first request
	resp, err = authAndDoRequest(ctx, doer, req, autherWithRefresh)
	if err != nil {
		return resp, err
	}

	// If response is unauthorised, attempt a token refresh
	if resp.StatusCode == http.StatusUnauthorized {
		if err = autherWithRefresh.Refresh(ctx, doer); err != nil {
			// If the refresh failed, return the original response
			return resp, nil
		}
	}

	// Do second request with refreshed token
	return authAndDoRequest(ctx, doer, req, autherWithRefresh)
}

func authAndDoRequest(ctx context.Context, doer httpcli.Doer, req *http.Request, autherWithRefresh auth.AuthenticatorWithRefresh) (resp *http.Response, err error) {
	if err := autherWithRefresh.Authenticate(req); err != nil {
		return nil, errors.Wrap(err, "could not authenticate request")
	}

	return doer.Do(req.WithContext(ctx))
}
