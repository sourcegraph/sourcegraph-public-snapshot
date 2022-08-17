package oauthutil

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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
type TokenRefresher func(ctx context.Context, doer httpcli.Doer, oauthCtx OAuthContext) (string, error)

// DoRequest is a function that uses the httpcli.Doer interface to make HTTP
// requests and to handle "401 Unauthorized" errors. When the 401 error is due to
// a token being expired, it will use the supplied TokenRefresher function to
// update the token. If the token is updated successfully, the same request will
// be retried exactly once.
func DoRequest(ctx context.Context, doer httpcli.Doer, req *http.Request, auther *auth.OAuthBearerToken, tokenRefresher TokenRefresher, oauthCtx OAuthContext) (code int, header http.Header, body []byte, err error) {
	for i := 0; i < 2; i++ {
		if auther != nil {
			if err := auther.Authenticate(req); err != nil {
				return 0, nil, nil, errors.Wrap(err, "authenticate")
			}
		}

		resp, err := doer.Do(req.WithContext(ctx))
		if err != nil {
			return 0, nil, nil, errors.Wrap(err, "do request")
		}

		body, err = io.ReadAll(resp.Body)
		if err != nil {
			_ = resp.Body.Close()
			return 0, nil, nil, errors.Wrap(err, "read response body")
		}

		_ = resp.Body.Close()

		if resp.StatusCode == http.StatusUnauthorized && auther != nil {
			if err = getOAuthErrorDetails(body); err != nil {
				if _, ok := err.(*oauthError); ok {
					// Refresh the token
					newToken, err := tokenRefresher(ctx, doer, oauthCtx)
					if err != nil {
						return 0, nil, nil, errors.Wrap(err, "refresh token")
					}
					auther = auther.WithToken(newToken)
					continue
				}
				return 0, nil, nil, errors.Errorf("unexpected OAuth error %T", err)
			}
		}
		return resp.StatusCode, resp.Header, body, nil
	}
	return 0, nil, nil, errors.Errorf("retries exceeded with status code %d and body %q", code, string(body))
}
