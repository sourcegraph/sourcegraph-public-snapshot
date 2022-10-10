package database

import (
	"context"
	"net/url"

	"github.com/sourcegraph/sourcegraph/schema"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func getGitHubAppInstallationRefreshFunc(externalServiceStore ExternalServiceStore, installationID int64, svc *types.ExternalService, appClient *github.V3Client) func(auther *github.GitHubAppInstallationAuthenticator) error {
	return func(auther *github.GitHubAppInstallationAuthenticator) error {
		token, err := appClient.CreateAppInstallationAccessToken(context.Background(), installationID)
		if err != nil {
			return err
		}

		auther.InstallationAccessToken = token.GetToken()
		auther.Expiry = token.ExpiresAt

		rawConfig, err := svc.Config.Decrypt(context.Background())
		if err != nil {
			return err
		}

		rawConfig, err = jsonc.Edit(rawConfig, token.GetToken(), "token")
		if err != nil {
			return err
		}

		externalServiceStore.Update(context.Background(),
			conf.Get().AuthProviders,
			svc.ID,
			&ExternalServiceUpdate{
				Config:         &rawConfig,
				TokenExpiresAt: token.ExpiresAt,
			},
		)

		return nil
	}
}

// BuildGitHubAppInstallationAuther builds a GitHub App Installation authenticator.
// First it creates a GitHub App Authenticator, as it is required to create App Installation
// Access Tokens. The App Authenticator is used in the refresh function of the
// Installation Authenticator, as installation tokens cannot be refreshed on their own.
// The App Authenticator is used to generate a new token once it expires.
func BuildGitHubAppInstallationAuther(externalServiceStore ExternalServiceStore, appID string, pkey []byte, urn string, apiURL *url.URL, cli httpcli.Doer, installationID int64, svc *types.ExternalService) (*github.GitHubAppInstallationAuthenticator, error) {
	if svc == nil {
		return nil, nil
	}

	rawConfig, err := svc.Config.Decrypt(context.Background())
	if err != nil {
		return nil, errors.Errorf("external service id=%d config error: %s", svc.ID, err)
	}
	var c schema.GitHubConnection
	if err := jsonc.Unmarshal(rawConfig, &c); err != nil {
		return nil, errors.Errorf("external service id=%d config error: %s", svc.ID, err)
	}

	appAuther, err := github.NewGitHubAppAuthenticator(appID, pkey)
	if err != nil {
		return nil, errors.Wrap(err, "new authenticator with GitHub App")
	}

	appClient := github.NewV3Client(
		log.Scoped("app", "github client for github app").
			With(log.String("appID", appID)),
		urn, apiURL, appAuther, cli)

	github.NewGitHubAppInstallationAuthenticator(installationID, c.Token, svc.TokenExpiresAt, getGitHubAppInstallationRefreshFunc(externalServiceStore, installationID, svc, appClient))
	return nil, nil
}
