package azuredevops

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/sourcegraph/log"
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

var MOCK_API_URL string

type Provider struct {
	db database.DB

	urn      string
	codeHost *extsvc.CodeHost

	// orgs is the list of orgs as configured in the code host connection.
	orgs []string
	// projects is the list of projects as configured in the code host connection.
	projects []string
}

func (p *Provider) FetchAccount(ctx context.Context, user *types.User, _ []*extsvc.Account, _ []string) (*extsvc.Account, error) {
	return nil, nil
}

func (p *Provider) FetchUserPerms(ctx context.Context, account *extsvc.Account, _ authz.FetchPermsOptions) (*authz.ExternalUserPermissions, error) {
	l := log.Scoped("azuredevops.FetchuserPerms", "logger for azuredevops provider")

	l.Debug("starting FetchUserPerms", log.String("user ID", fmt.Sprintf("%#v", account.UserID)))

	_, token, err := azuredevops.GetExternalAccountData(ctx, &account.AccountData)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"failed to load external account data from database with external account with ID: %d",
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

	var apiURL string
	if MOCK_API_URL != "" {
		apiURL = MOCK_API_URL
	} else {
		apiURL = azuredevops.AZURE_DEV_OPS_API_URL
	}

	client, err := azuredevops.NewClient(
		p.ServiceID()+account.AccountID,
		apiURL,
		oauthToken,
		nil,
	)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"failed to create client for service ID: %q, account ID: %q", p.ServiceID(), account.AccountID,
		)
	}

	var repos []azuredevops.Repository

	seenOrgs := map[string]struct{}{}
	for _, org := range p.orgs {
		// The list of orgs may have duplicates if there are multiple Azure DevOps code host
		// connections that have the same org in their config. As a result, skip this org if we
		// already listed its repositories in this iteration of permissions syncing.
		if _, ok := seenOrgs[org]; ok {
			continue
		}

		seenOrgs[org] = struct{}{}

		l.Debug("listing repos",
			log.String("org", fmt.Sprintf("%#v", org)),
		)

		r, err := client.ListRepositoriesByProjectOrOrg(ctx, azuredevops.ListRepositoriesByProjectOrOrgArgs{
			ProjectOrOrgName: org,
		})
		if err != nil {
			if httpErr, ok := err.(*azuredevops.HTTPError); ok {
				// If the HTTPError is 401 / 403 / 404, this user does not have access to this org.
				// Skip and continue to the next.
				//
				// For orgs that don't exist, the API returns 404. For orgs that the user does not
				// have access to the API returns 401. We're not sure if the API might return 403
				// for some use case but we don't want to hard fail on that either.
				if httpErr.StatusCode == http.StatusUnauthorized || httpErr.StatusCode == http.StatusForbidden || httpErr.StatusCode == http.StatusNotFound {

					l.Debug("skipping org",
						log.String("org", org),
						log.Int("http status code", httpErr.StatusCode),
					)

					continue
				}
			}

			// For any other errors, we want to hard fail so that the issue can be identified.
			return nil, errors.Newf("failed to list repositories for org: %q with error: %q", org, err.Error())
		}

		l.Debug("adding repos", log.Int("count", len(r)))
		repos = append(repos, r...)
	}

	seenProjects := map[string]struct{}{}
	for _, project := range p.projects {
		// The list of projects may have duplicates if there are multiple Azure DevOps code host
		// connections that have the same project in their config. As a result, skip this project if
		// we already listed its repositories in this iteration of permissions syncing.
		if _, ok := seenProjects[project]; ok {
			continue
		}

		seenProjects[project] = struct{}{}

		r, err := client.ListRepositoriesByProjectOrOrg(ctx, azuredevops.ListRepositoriesByProjectOrOrgArgs{
			ProjectOrOrgName: project,
		})
		if err != nil {
			if httpErr, ok := err.(*azuredevops.HTTPError); ok {
				// If the HTTPError is 401 / 403 / 404, this user does not have access to this org.
				// Skip and continue to the next.
				//
				// For orgs/projects that don't exist, or the user does not have access to the API
				// returns 404. We're not sure if the API might return 401 or 403 for some use case
				// but we don't want to hard fail on that either.
				if httpErr.StatusCode == http.StatusUnauthorized || httpErr.StatusCode == http.StatusForbidden || httpErr.StatusCode == http.StatusNotFound {

					l.Debug("skipping project",
						log.String("project", project),
						log.Int("http status code", httpErr.StatusCode),
					)

					continue
				}
			}

			// For any other errors, we want to hard fail so that the issue can be identified.
			return nil, errors.Newf("failed to list repositories for project: %q with error: %q", project, err.Error())
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

func (p *Provider) FetchRepoPerms(ctx context.Context, repo *extsvc.Repository, opts authz.FetchPermsOptions) ([]extsvc.AccountID, error) {
	return nil, authz.ErrUnimplemented{Feature: "azuredevops.FetchRepoPerms"}
}

func (p *Provider) ServiceType() string {
	return p.codeHost.ServiceType
}

func (p *Provider) ServiceID() string {
	return p.codeHost.ServiceID
}

func (p *Provider) URN() string {
	return p.urn
}

// TODO: Implement this in a follow up PR.
func (p *Provider) ValidateConnection(ctx context.Context) error {
	return nil
}

func newAuthzProvider(db database.DB, orgs, projects []string) (*Provider, error) {

	if err := licensing.Check(licensing.FeatureACLs); err != nil {
		return nil, err
	}

	u, err := url.Parse(azuredevops.AZURE_DEV_OPS_API_URL)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse url: %q, this is likely a misconfigured URL in the constant azuredevops.AZURE_DEV_OPS_API_URL", azuredevops.AZURE_DEV_OPS_API_URL)
	}

	return &Provider{
		db:       db,
		urn:      "azuredevops:authzprovider",
		codeHost: extsvc.NewCodeHost(u, extsvc.TypeAzureDevOps),
		orgs:     orgs,
		projects: projects,
	}, nil
}

func NewAuthzProviders(db database.DB, conns []*types.AzureDevOpsConnection) *authztypes.ProviderInitResult {
	initResults := &authztypes.ProviderInitResult{}
	orgs, projects := []string{}, []string{}

	// Iterate over all Azure Dev Ops code host connections to make sure we sync permissions for all
	// orgs and projects in every permissions sync iteration.
	for _, c := range conns {
		if c.AzureDevOpsConnection == nil {
			continue
		}

		orgs = append(orgs, c.Orgs...)
		projects = append(projects, c.Projects...)
	}

	p, err := newAuthzProvider(db, orgs, projects)
	if err != nil {
		initResults.InvalidConnections = append(initResults.InvalidConnections, extsvc.TypeAzureDevOps)
		initResults.Problems = append(initResults.Problems, err.Error())
	} else if p != nil {
		initResults.Providers = append(initResults.Providers, p)
	}

	return initResults
}
