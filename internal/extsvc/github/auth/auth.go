package auth

import (
	"context"
	"net/url"

	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	ghaauth "github.com/sourcegraph/sourcegraph/internal/github_apps/auth"
	"github.com/sourcegraph/sourcegraph/internal/github_apps/store"
	"github.com/sourcegraph/sourcegraph/schema"
)

type RefreshableURLRequestAuthenticator interface {
	auth.Authenticator
	auth.URLAuthenticator
	auth.Refreshable
}

// FromConnection creates an authenticator from a GitHubConnection.
// It returns an OAuthBearerToken if the connection uses a Personal Access
// Token, and it returns a GitHubAppAuthenticator if the connection is
// configured via a GitHub App Installation.
func FromConnection(
	ctx context.Context,
	conn *schema.GitHubConnection,
	ghApps store.GitHubAppsStore,
	encryptionKey encryption.Key,
) (RefreshableURLRequestAuthenticator, error) {
	if conn.GitHubAppDetails == nil {
		return &auth.OAuthBearerToken{Token: conn.Token}, nil
	}

	ghApp, err := ghApps.GetByAppID(ctx, conn.GitHubAppDetails.AppID, conn.Url)
	if err != nil {
		return nil, err
	}

	appAuther, err := ghaauth.NewGitHubAppAuthenticator(ghApp.AppID, []byte(ghApp.PrivateKey))
	if err != nil {
		return nil, err
	}

	baseURL, err := url.Parse(conn.Url)
	if err != nil {
		return nil, err
	}

	return ghaauth.NewInstallationAccessToken(baseURL, conn.GitHubAppDetails.InstallationID, appAuther, encryptionKey), nil
}
