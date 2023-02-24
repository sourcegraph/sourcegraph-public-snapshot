package azuredevops

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	authztypes "github.com/sourcegraph/sourcegraph/enterprise/internal/authz/types"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/azuredevops"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Provider struct {
	db database.DB

	urn      string
	codeHost *extsvc.CodeHost

	orgs     []string
	projects []string
}

func (p *Provider) FetchAccount(ctx context.Context, user *types.User, _ []*extsvc.Account, _ []string) (*extsvc.Account, error) {
	return nil, nil
}

func (p *Provider) FetchUserPerms(ctx context.Context, account *extsvc.Account, opts authz.FetchPermsOptions) (*authz.ExternalUserPermissions, error) {
	if account == nil {
		return nil, errors.New("skipping user permissions sync for nil account, this is likely a bug in fetching the user external accounts, see call site of FetchUserPerms")
	}

	_, token, err := azuredevops.GetExternalAccountData(ctx, &account.AccountData)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"failed to get external account data external account with ID: %d, this might be related to a bad JSON in the database table user_external_accounts for this ID or a misconfigured encryption key",
			account.ID,
		)
	}

	oauthToken := &auth.OAuthBearerToken{
		Token:              token.AccessToken,
		RefreshToken:       token.RefreshToken,
		Expiry:             token.Expiry,
		NeedsRefreshBuffer: 5,
	}

	oauthContext, err := azuredevops.GetOAuthContext(token.RefreshToken)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to generate oauth context, this is likely a misconfiguration with the Azure OAuth provider (bad URL?), please check the auth.providers configuration in your site config")
	}

	oauthToken.RefreshFunc = database.GetAccountRefreshAndStoreOAuthTokenFunc(p.db, account.ID, oauthContext)

	client, err := azuredevops.NewClient(
		p.ServiceID()+account.AccountID,
		azuredevops.AZURE_DEV_OPS_API_URL,
		oauthToken,
		nil,
	)

	var repos []azuredevops.Repository

	for _, org := range p.orgs {
		r, err := client.ListRepositoriesByProjectOrOrg(ctx, azuredevops.ListRepositoriesByProjectOrOrgArgs{
			ProjectOrOrgName: org,
		})
		if err != nil {
			// TODO: Test error handling. And do the same for projects.
			msg := err.Error()
			if httpErr, ok := err.(*azuredevops.HTTPError); ok {
				// The Azure DevOps API returns HTML on 4xx errors. We don't want to dump HTML into
				// the logs / errors as this will bubble up everywhere and clutter all user
				// interfaces. It destroys everything that it touches.
				//
				// Drop the body and only write the status code and URL in the error.
				msg = fmt.Sprintf("HTTP status: %d, URL: %q", httpErr.StatusCode, httpErr.URL.String())

				if httpErr.StatusCode == http.StatusUnauthorized || httpErr.StatusCode == http.StatusForbidden {
					continue
				}
			}

			return nil, errors.Newf("failed to list repositories for org: %q with error: %q", org, msg)
			// TODO: log and continue or hard fail?
			// TODO: Maybe hard fail on 4xx errs only.
		}

		repos = append(repos, r...)
	}

	for _, project := range p.projects {
		r, err := client.ListRepositoriesByProjectOrOrg(ctx, azuredevops.ListRepositoriesByProjectOrOrgArgs{
			ProjectOrOrgName: project,
		})
		if err != nil {
			return nil, errors.Newf("failed to list repositories for project: %q", project)
			// TODO: log and continue or hard fail?
			// TODO: Maybe hard fail on 4xx errs only.
		}

		repos = append(repos, r...)
	}

	extIDs := make([]extsvc.RepoID, 0, len(repos))
	for _, repo := range repos {
		extIDs = append(extIDs, extsvc.RepoID(repo.ID))
	}

	return &authz.ExternalUserPermissions{
		Exacts: extIDs,
	}, err
}

// FIXME: Implement
func (p *Provider) FetchRepoPerms(ctx context.Context, repo *extsvc.Repository, opts authz.FetchPermsOptions) ([]extsvc.AccountID, error) {
	// return nil, nil
	return nil, authz.ErrUnimplemented{Feature: "azuredevops.FetchRepoPerms"}
}

func (p *Provider) ServiceType() string {
	return p.codeHost.ServiceType
}

func (p *Provider) ServiceID() string {
	return p.codeHost.ServiceID
	// return azuredevops.AZURE_DEV_OPS_API_URL
}

func (p *Provider) URN() string {
	return p.urn
}

// FIXME: Implement
func (p *Provider) ValidateConnection(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return nil
}

func newAuthzProvider(db database.DB, conn *types.AzureDevOpsConnection) (*Provider, error) {
	if conn.AzureDevOpsConnection == nil {
		return nil, nil
	}

	u, err := url.Parse(azuredevops.AZURE_DEV_OPS_API_URL)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse url: %q, this is likely a misconfigured URL in the constant azuredevops.AZURE_DEV_OPS_API_URL", azuredevops.AZURE_DEV_OPS_API_URL)
	}

	return &Provider{
		db:       db,
		urn:      conn.URN,
		codeHost: extsvc.NewCodeHost(u, extsvc.TypeAzureDevOps),
		orgs:     conn.Orgs,
		projects: conn.Projects,
	}, nil
}

func NewAuthzProviders(db database.DB, conns []*types.AzureDevOpsConnection) *authztypes.ProviderInitResult {
	initResults := &authztypes.ProviderInitResult{}

	// TODO: Confirm if it's okay to check this once for all connections or if we need to check this
	// for each connection individually, like we do in other places like the gitlab.NewAuthzProviders.
	if err := licensing.Check(licensing.FeatureACLs); err != nil {
		initResults.InvalidConnections = append(initResults.InvalidConnections, extsvc.TypeAzureDevOps)
		initResults.Problems = append(initResults.Problems, err.Error())

		return initResults
	}

	for _, c := range conns {
		p, err := newAuthzProvider(db, c)
		if err != nil {
			initResults.InvalidConnections = append(initResults.InvalidConnections, extsvc.TypeAzureDevOps)
			initResults.Problems = append(initResults.Problems, err.Error())
		} else if p != nil {
			initResults.Providers = append(initResults.Providers, p)
		}
	}

	return initResults

}
