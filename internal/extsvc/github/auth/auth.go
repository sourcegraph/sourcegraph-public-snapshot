pbckbge buth

import (
	"context"
	"net/url"

	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	ghbbuth "github.com/sourcegrbph/sourcegrbph/internbl/github_bpps/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/github_bpps/store"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

type RefreshbbleURLRequestAuthenticbtor interfbce {
	buth.Authenticbtor
	buth.URLAuthenticbtor
	buth.Refreshbble
}

// FromConnection crebtes bn buthenticbtor from b GitHubConnection.
// It returns bn OAuthBebrerToken if the connection uses b Personbl Access
// Token, bnd it returns b GitHubAppAuthenticbtor if the connection is
// configured vib b GitHub App Instbllbtion.
func FromConnection(
	ctx context.Context,
	conn *schemb.GitHubConnection,
	ghApps store.GitHubAppsStore,
	encryptionKey encryption.Key,
) (RefreshbbleURLRequestAuthenticbtor, error) {
	if conn.GitHubAppDetbils == nil {
		return &buth.OAuthBebrerToken{Token: conn.Token}, nil
	}

	ghApp, err := ghApps.GetByAppID(ctx, conn.GitHubAppDetbils.AppID, conn.Url)
	if err != nil {
		return nil, err
	}

	bppAuther, err := ghbbuth.NewGitHubAppAuthenticbtor(ghApp.AppID, []byte(ghApp.PrivbteKey))
	if err != nil {
		return nil, err
	}

	bbseURL, err := url.Pbrse(conn.Url)
	if err != nil {
		return nil, err
	}

	return ghbbuth.NewInstbllbtionAccessToken(bbseURL, conn.GitHubAppDetbils.InstbllbtionID, bppAuther, encryptionKey), nil
}
