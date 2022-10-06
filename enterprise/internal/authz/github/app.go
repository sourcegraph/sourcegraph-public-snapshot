package github

import (
	"context"
	"encoding/base64"
	"net/url"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/schema"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// newAppProvider creates a new authz Provider for GitHub App.
func newAppProvider(
	db database.DB,
	svc *types.ExternalService,
	urn string,
	baseURL *url.URL,
	appID string,
	privateKey string,
	installationID int64,
	cli httpcli.Doer,
) (*Provider, error) {
	pkey, err := base64.StdEncoding.DecodeString(privateKey)
	if err != nil {
		return nil, errors.Wrap(err, "decode private key")
	}

	apiURL, _ := github.APIRoot(baseURL)
	var installationAuther auth.Authenticator
	if svc != nil {
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

		apiURL, _ := github.APIRoot(baseURL)
		appClient := github.NewV3Client(
			log.Scoped("app", "github client for github app").
				With(log.String("appID", appID)),
			urn, apiURL, appAuther, cli)

		installationRefreshFunc := func(auther *github.GitHubAppInstallationAuthenticator) error {
			token, err := appClient.CreateAppInstallationAccessToken(context.Background(), installationID)
			if err != nil {
				return err
			}

			auther.InstallationAccessToken = token.GetToken()
			auther.Expiry = token.ExpiresAt

			rawConfig, err = jsonc.Edit(rawConfig, token.GetToken(), "token")
			if err != nil {
				return err
			}

			db.ExternalServices().Update(context.Background(),
				conf.Get().AuthProviders,
				svc.ID,
				&database.ExternalServiceUpdate{
					Config:         &rawConfig,
					TokenExpiresAt: token.ExpiresAt,
				},
			)

			return nil
		}

		installationAuther, err = github.NewGitHubAppInstallationAuthenticator(installationID, c.Token, svc.TokenExpiresAt, installationRefreshFunc)
		if err != nil {
			return nil, errors.Wrap(err, "new GitHub App installation auther")
		}
	}

	return &Provider{
		urn:      urn,
		codeHost: extsvc.NewCodeHost(baseURL, extsvc.TypeGitHub),
		client: func() (client, error) {
			logger := log.Scoped("installation", "github client for installation").
				With(log.String("appID", appID), log.Int64("installationID", installationID))

			return &ClientAdapter{
				V3Client: github.NewV3Client(logger, urn, apiURL, installationAuther, cli),
			}, nil
		},
		InstallationID: &installationID,
		db:             db,
	}, nil
}
