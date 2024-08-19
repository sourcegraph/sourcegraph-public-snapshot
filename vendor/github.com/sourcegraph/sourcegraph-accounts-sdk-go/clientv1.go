package sams

import (
	"context"
	"fmt"
	"strings"

	"connectrpc.com/connect"
	"connectrpc.com/otelconnect"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	clientsv1 "github.com/sourcegraph/sourcegraph-accounts-sdk-go/clients/v1"
	"github.com/sourcegraph/sourcegraph-accounts-sdk-go/scopes"
)

// ClientV1 provides helpers to talk to a SAMS instance via Clients API v1.
type ClientV1 struct {
	rootURL     string
	tokenSource oauth2.TokenSource

	// defaultInterceptors is a list of default interceptors to use with all
	// clients, generally providing enhanced diagnostics.
	defaultInterceptors []connect.Interceptor
}

type ClientV1Config struct {
	ConnConfig
	// TokenSource is the OAuth2 token source to use for authentication. It MUST be
	// based on a per-client token that is on behalf of a SAMS client (i.e. Clients
	// Credentials).
	//
	// The oauth2.TokenSource abstraction will take care of creating short-lived
	// access tokens as needed.
	TokenSource oauth2.TokenSource
}

func (c ClientV1Config) Validate() error {
	if err := c.ConnConfig.Validate(); err != nil {
		return errors.Wrap(err, "invalid ConnConfig")
	}
	if c.TokenSource == nil {
		return errors.New("token source is required")
	}
	return nil
}

// NewClientV1 returns a new SAMS client for interacting with Clients API v1
// using the given client credentials, and the scopes are used to as requested
// scopes for access tokens that are issued to this client.
func NewClientV1(config ClientV1Config) (*ClientV1, error) {
	if err := config.Validate(); err != nil {
		return nil, errors.Wrap(err, "ClientV1ConnectionConfig is invalid")
	}

	otelinterceptor, err := otelconnect.NewInterceptor(
		// Start with simple, lower-cardinality metrics
		otelconnect.WithoutServerPeerAttributes(),
		// Start with lower-volume trace data
		otelconnect.WithoutTraceEvents())
	if err != nil {
		return nil, errors.Wrap(err, "initiate OTEL interceptor")
	}

	apiURL := config.getAPIURL()
	return &ClientV1{
		rootURL:             strings.TrimSuffix(apiURL, "/"),
		tokenSource:         config.TokenSource,
		defaultInterceptors: []connect.Interceptor{otelinterceptor},
	}, nil
}

func parseResponseAndError[T any](resp *connect.Response[T], err error) (*connect.Response[T], error) {
	var connectErr *connect.Error
	if !errors.As(err, &connectErr) {
		// Not an error that we can extract information from.
		return resp, err
	}

	if connectErr.Code() == connect.CodeNotFound {
		return nil, ErrNotFound
	}

	// Cannot determine action solely based on status code, let's look at the error
	// details.
	for _, detail := range connectErr.Details() {
		value, err := detail.Value()
		if err != nil {
			return nil, errors.Wrap(err, "extract error detail value")
		}

		switch value.(type) {
		case *clientsv1.ErrorRecordMismatch:
			return nil, ErrRecordMismatch
		}
	}

	// Nothing juicy, return as-is.
	return resp, err
}

func (c *ClientV1) gRPCURL() string {
	return c.rootURL + "/api/grpc"
}

// Users returns a client handler to interact with the UsersServiceV1 API.
func (c *ClientV1) Users() *UsersServiceV1 {
	return &UsersServiceV1{client: c}
}

// Sessions returns a client handler to interact with the SessionsServiceV1 API.
func (c *ClientV1) Sessions() *SessionsServiceV1 {
	return &SessionsServiceV1{client: c}
}

// Tokens returns a client handler to interact with the TokensServiceV1 API.
func (c *ClientV1) Tokens() *TokensServiceV1 {
	return &TokensServiceV1{client: c}
}

var (
	ErrNotFound       = errors.New("not found")
	ErrRecordMismatch = errors.New("record mismatch")
)

// ClientCredentialsTokenSource returns a TokenSource that generates an access
// token using the client credentials flow. Internally, the token returned is
// reused. So that new tokens are only created when needed. (Provided this
// `Client` is long-lived.)
//
// The `requestScopes` is a list of scopes to be requested for this token
// source. Scopes should be defined using the available scopes package. All
// requested scopes must be allowed by the registered client - see:
// https://sourcegraph.notion.site/6cc4a1bd9cb247eea9674dbf9d5ce8c3
func ClientCredentialsTokenSource(conn ConnConfig, clientID, clientSecret string, requestScopes []scopes.Scope) oauth2.TokenSource {
	config := clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     fmt.Sprintf("%s/oauth/token", conn.getAPIURL()),
		Scopes:       scopes.ToStrings(requestScopes),
	}
	return config.TokenSource(context.Background())
}
