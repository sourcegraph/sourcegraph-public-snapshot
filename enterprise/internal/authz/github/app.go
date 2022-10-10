package github

import (
	"encoding/base64"
	"net/url"

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

	var installationAuther auth.AuthenticatorWithRefresh
	if db != nil { // should only be nil when called by ValidateAuthz
		installationAuther, err = database.BuildGitHubAppInstallationAuther(db.ExternalServices(), appID, pkey, urn, apiURL, cli, installationID, svc)
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
