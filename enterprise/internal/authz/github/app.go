package github

import (
	"context"
	"encoding/base64"
	"net/url"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// newAppProvider creates a new authz Provider for GitHub App.
func newAppProvider(
	externalServicesStore database.ExternalServiceStore,
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

	auther, err := auth.NewOAuthBearerTokenWithGitHubApp(appID, pkey)
	if err != nil {
		return nil, errors.Wrap(err, "new authenticator with GitHub App")
	}

	apiURL, _ := github.APIRoot(baseURL)
	appClient := github.NewV3Client(urn, apiURL, auther, cli)
	return &Provider{
		urn:      urn,
		codeHost: extsvc.NewCodeHost(baseURL, extsvc.TypeGitHub),
		client: func() (client, error) {
			token, err := repos.GetOrRenewGitHubAppInstallationAccessToken(context.Background(), externalServicesStore, svc, appClient, installationID)
			if err != nil {
				return nil, errors.Wrap(err, "get or renew GitHub App installation access token")
			}

			auther = &auth.OAuthBearerToken{Token: token}
			return &ClientAdapter{
				V3Client: github.NewV3Client(urn, apiURL, auther, cli),
			}, nil
		},
	}, nil
}
