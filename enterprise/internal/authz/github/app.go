package github

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// todo
func newAppProvider(urn string, baseURL *url.URL, appID, privateKey string, installationID int64) (*Provider, error) {
	pkey, err := base64.StdEncoding.DecodeString(privateKey)
	if err != nil {
		return nil, errors.Wrap(err, "decode private key")
	}

	auther, err := auth.NewOAuthBearerTokenWithGitHubApp(appID, pkey)
	if err != nil {
		return nil, errors.Wrap(err, "new authenticator with GitHub App")
	}

	apiURL, _ := github.APIRoot(baseURL)
	appClient := github.NewV3Client(apiURL, auther, nil)
	return &Provider{
		urn:      urn,
		codeHost: extsvc.NewCodeHost(baseURL, extsvc.TypeGitHub),
		client: func() (client, error) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			defer cancel()

			// TODO(cloud-saas): Cache the installation access token until it expires.
			token, err := appClient.CreateAppInstallationAccessToken(ctx, installationID)
			if err != nil {
				return nil, errors.Wrap(err, "create app installation access token")
			}
			if token.Token == nil {
				return nil, errors.New("empty token returned")
			}

			fmt.Println("new token:", *token.Token)
			auther = &auth.OAuthBearerToken{Token: *token.Token}
			return &ClientAdapter{
				V3Client: github.NewV3Client(apiURL, auther, nil),
			}, nil
		},
	}, nil
}
