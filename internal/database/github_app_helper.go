package database

import (
	"context"
	"net/url"
	"time"

	"github.com/sourcegraph/sourcegraph/schema"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func getGitHubAppInstallationRefreshFunc(externalServiceStore ExternalServiceStore, installationID int64, svc *types.ExternalService, appClient *github.V3Client) func(context.Context, httpcli.Doer) (string, time.Time, error) {
	return func(ctx context.Context, cli httpcli.Doer) (string, time.Time, error) {
		token, err := appClient.CreateAppInstallationAccessToken(ctx, installationID)
		if err != nil {
			return "", time.Time{}, err
		}

		rawConfig, err := svc.Config.Decrypt(context.Background())
		if err != nil {
			return "", time.Time{}, err
		}

		rawConfig, err = jsonc.Edit(rawConfig, token.GetToken(), "token")
		if err != nil {
			return "", time.Time{}, err
		}

		externalServiceStore.Update(context.Background(),
			conf.Get().AuthProviders,
			svc.ID,
			&ExternalServiceUpdate{
				Config:         &rawConfig,
				TokenExpiresAt: token.ExpiresAt,
			},
		)

		return *token.Token, token.GetExpiresAt(), nil
	}
}

// BuildGitHubAppInstallationAuther builds a GitHub App Installation authenticator.
// First it creates a GitHub App Authenticator, as it is required to create App Installation
// Access Tokens. The App Authenticator is used in the refresh function of the
// Installation Authenticator, as installation tokens cannot be refreshed on their own.
// The App Authenticator is used to generate a new token once it expires.
func BuildGitHubAppInstallationAuther(
	externalServiceStore ExternalServiceStore,
	appID string,
	pkey []byte,
	urn string,
	apiURL *url.URL,
	cli httpcli.Doer,
	installationID int64,
	svc *types.ExternalService,
) (auth.AuthenticatorWithRefresh, error) {
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

	expiry := time.Time{}
	if svc.TokenExpiresAt != nil {
		expiry = *svc.TokenExpiresAt
	}

	return github.NewGitHubAppInstallationAuthenticator(
		installationID,
		c.Token,
		expiry,
		getGitHubAppInstallationRefreshFunc(externalServiceStore, installationID, svc, appClient),
	)
}
