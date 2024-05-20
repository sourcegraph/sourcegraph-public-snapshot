package sams

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Scope string

const (
	ScopeDotcom Scope = "client.dotcom"
	// more granualar scopes can be added here later if required
)

// TokenIntrospection is the response from the OAuth spec. More information
// can be found at: https://www.oauth.com/oauth2-servers/token-introspection-endpoint/
type TokenIntrospection struct {
	// This is a boolean value of whether or not the presented token is
	// currently active. The value should be “true” if the token has been
	// issued by this authorization server, has not been revoked by the user,
	// and has not expired.
	Active bool `json:"active"`

	// A space-separated list of scopes associated with this token.
	Scope string `json:"scope"`

	// The client identifier for the OAuth 2.0 client that the token was issued to.
	ClientID string `json:"client_id"`

	// A human-readable identifier for the user who authorized this token.
	Username string `json:"username"`

	// The unix timestamp (integer timestamp, number of seconds since January 1, 1970 UTC)
	// indicating when this token will expire.
	Expiration int64 `json:"exp"`
}

// Client defines a basic SAMS client for verifying access tokens. All uses should
// be replaced by the proper sourcegraph-accounts-sdk-go client.
type Client interface {
	IntrospectToken(ctx context.Context, token string) (*TokenIntrospection, error)
}

type samsClient struct {
	server                  string
	clientCredentialsConfig clientcredentials.Config
	tokenSource             oauth2.TokenSource
}

func (c *samsClient) IntrospectToken(ctx context.Context, token string) (*TokenIntrospection, error) {
	// https://www.oauth.com/oauth2-servers/token-introspection-endpoint/
	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/oauth/introspect", c.server),
		strings.NewReader("token="+token),
	)
	if err != nil {
		return nil, errors.Wrap(err, "create introspection request")
	}

	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Create the OAuth2 HTTP client. This will handle setting up the HTTP headers based on the token source.
	// (And potentially issue a new token as needed.)
	httpClient := oauth2.NewClient(ctx, c.tokenSource)
	response, err := httpClient.Do(request)
	if err != nil {
		return nil, errors.Wrap(err, "introspecting SAMS token")
	}

	defer response.Body.Close()

	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, errors.Wrap(err, "reading response")
	}

	if response.StatusCode != http.StatusOK {
		return nil, errors.Errorf("unexpected status code %d from SAMS", response.StatusCode)
	}

	var introspectionResponse TokenIntrospection
	err = json.Unmarshal(bodyBytes, &introspectionResponse)
	if err != nil {
		return nil, errors.Wrap(err, "parse introspection response body")
	}
	return &introspectionResponse, nil
}

func NewClient(samsServer string, clientCredentialsConfig clientcredentials.Config) Client {
	// Provide a default value for the required TokenURL field, in case the caller forgot to set it.
	if clientCredentialsConfig.TokenURL == "" {
		clientCredentialsConfig.TokenURL = fmt.Sprintf("%s/oauth/token", samsServer)
	}

	return &samsClient{
		server:                  samsServer,
		clientCredentialsConfig: clientCredentialsConfig,
		tokenSource:             clientCredentialsConfig.TokenSource(context.Background()),
	}
}
