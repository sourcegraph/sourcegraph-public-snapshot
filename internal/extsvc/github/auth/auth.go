package auth

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/schema"
)

type FromConnectionFunc func(context.Context, *schema.GitHubConnection) (auth.Authenticator, error)

// FromConnection creates an authenticator from a GitHubConnection.,
// This function gets replaced in Enterprise services that support GitHub App.
var FromConnection = func(_ context.Context, conn *schema.GitHubConnection) (auth.Authenticator, error) {
	return &auth.OAuthBearerToken{Token: conn.Token}, nil
}
