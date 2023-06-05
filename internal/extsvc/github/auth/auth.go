package auth

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/schema"
)

type RefreshableURLRequestAuthenticator interface {
	auth.Authenticator
	auth.URLAuthenticator
	auth.Refreshable
}

type FromConnectionFunc func(context.Context, *schema.GitHubConnection) (RefreshableURLRequestAuthenticator, error)

// FromConnection creates an authenticator from a GitHubConnection.,
// This function gets replaced in Enterprise services that support GitHub App.
var FromConnection = func(_ context.Context, conn *schema.GitHubConnection) (RefreshableURLRequestAuthenticator, error) {
	return &auth.OAuthBearerToken{Token: conn.Token}, nil
}
