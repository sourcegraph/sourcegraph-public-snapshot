package database

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func GetAppInstallationRefreshFunc(externalServiceStore ExternalServiceStore, installationID int64, svc *types.ExternalService, appClient *github.V3Client) func(auther *github.GitHubAppInstallationAuthenticator) error {
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
