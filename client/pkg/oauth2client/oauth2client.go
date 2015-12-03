// Package oauth2client has helpers for implementing OAuth2 client
// applications.
package oauth2client

import (
	"errors"
	"net/http"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/errcode"
	"src.sourcegraph.com/sourcegraph/fed"
)

type contextKey int

const clientIDKey = iota

// ClientIDFromContext retrieves the OAuth2 client ID in ctx that was
// previously stored by WithClientID.
func ClientIDFromContext(ctx context.Context) string {
	clientID, _ := ctx.Value(clientIDKey).(string)
	return clientID
}

// WithClientID returns a copy of ctx that has the specified OAuth2
// client ID.
func WithClientID(ctx context.Context, clientID string) context.Context {
	return context.WithValue(ctx, clientIDKey, clientID)
}

// Config returns the OAuth2 configuration, based on information
// stored in the context.
func Config(ctx context.Context) (*oauth2.Config, error) {
	clientID := ClientIDFromContext(ctx)
	if clientID == "" {
		return nil, ErrClientNotRegistered
	}

	// Provider endpoints
	authURL, err := router.Rel.URLToOrError(router.OAuth2ServerAuthorize)
	if err != nil {
		return nil, err
	}

	return &oauth2.Config{
		ClientID: clientID,
		Endpoint: oauth2.Endpoint{
			AuthURL:  fed.Config.RootURL().ResolveReference(authURL).String(),
			TokenURL: TokenURL(),
		},
	}, nil
}

// ErrClientNotRegistered occurs when an operation that requires a
// registered client ID is required, and none is present in the
// context.
var ErrClientNotRegistered = &errcode.HTTPErr{
	Status: http.StatusForbidden,
	Err:    errors.New("no OAuth2 client ID configured"),
}

func TokenURL() string {
	if fed.Config.IsRoot {
		return ""
	}

	tokenURL, err := router.Rel.URLToOrError(router.OAuth2ServerToken)
	if err != nil {
		panic(err)
	}
	return fed.Config.RootURL().ResolveReference(tokenURL).String()
}
