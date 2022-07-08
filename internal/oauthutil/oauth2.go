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

type Endpoint = oauth2.Endpoint

// todo docstring
// todo Maybe can be deleted based on how do token refresh (i.e. need access to external service object anyway)
type Context struct {
	// ClientID is the application's ID.
	ClientID string
	// ClientSecret is the application's secret.
	ClientSecret string
	// Endpoint contains the resource server's token endpoint
	// URLs. These are constants specific to each server and are
	// often available via site-specific packages, such as
	// google.Endpoint or github.Endpoint.
	Endpoint Endpoint
	// Scope specifies optional requested permissions.
	Scopes []string
	// RefreshToken is a token that's used by the application
	// (as opposed to the user) to refresh the access token
	// if it expires.
	RefreshToken string
}

type oauthError struct {
	Err              string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func (e oauthError) Error() string {
	return fmt.Sprintf("OAuth response error %q description %q", e.Err, e.ErrorDescription)
}

// getOAuthErrorDetails only returns error if it's an OAuth error. For other
// errors like 404 we don't return error. We do this because this method is only
// intended to be used by oauth to refresh access token on expiration.
//
// When it's error like 404, GitLab API doesn't return it as error so we keep
// the similar behavior and let caller check the response status code.
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
type TokenRefresher func(ctx context.Context, doer httpcli.Doer, oauthCtx Context) (string, error)

// todo docstring
func DoRequest(ctx context.Context, doer httpcli.Doer, req *http.Request, auther *auth.OAuthBearerToken, oauthCtx Context, tokenRefresher TokenRefresher) (code int, header http.Header, body []byte, err error) {
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
				return 0, nil, nil, errors.Errorf("got unexpected OAuth error %T", err)
			}
		}
		return resp.StatusCode, resp.Header, body, nil
	}
	return 0, nil, nil, errors.Errorf("retries exceeded for OAuth refresher with status code %d and body %q", code, string(body))
}
