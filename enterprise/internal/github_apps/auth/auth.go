package auth

import (
	"context"

	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	ghauth "github.com/sourcegraph/sourcegraph/internal/extsvc/github/auth"
	"github.com/sourcegraph/sourcegraph/schema"
)

// AutherFromConnection determines whether or not the provided GitHub connection is a
// GitHub App installation, and returns the appropriate Authenticator.
func AutherFromConnection(ctx context.Context, ossDB database.DB, c *schema.GitHubConnection) (auth.Authenticator, error) {
	if ghDetails := c.GitHubAppDetails; ghDetails != nil {
		enterpriseDB := edb.NewEnterpriseDB(ossDB)
		ghApp, err := enterpriseDB.GithubApps().GetByAppID(ctx, ghDetails.AppID, ghDetails.BaseURL)
		if err != nil {
			return nil, err
		}
		appAuther, err := ghauth.NewGitHubAppAuthenticator(ghApp.AppID, []byte(ghApp.PrivateKey))
		if err != nil {
			return nil, err
		}
		return ghauth.NewInstallationAccessToken(ghDetails.InstallationID, appAuther), nil
	}

	return &auth.OAuthBearerToken{Token: c.Token}, nil
}
