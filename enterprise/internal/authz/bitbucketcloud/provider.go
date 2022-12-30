// Package bitbucketcloud contains an authorization provider for Bitbucket Cloud.
package bitbucketcloud

import (
	"context"
	"net/url"
	"strings"

	"github.com/ktrysmt/go-bitbucket"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketcloud"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Provider is an implementation of AuthzProvider that provides repository and
// user permissions as determined from Bitbucket Cloud.
type Provider struct {
	urn      string
	codeHost *extsvc.CodeHost
	client   *bitbucket.Client
	pageSize int // Page size to use in paginated requests.
}

var _ authz.Provider = (*Provider)(nil)

// NewProvider returns a new Bitbucket Cloud authorization provider that uses
// the given bitbucket.Client to talk to the Bitbucket Cloud API that is
// the source of truth for permissions. Sourcegraph users will need a valid
// Bitbucket Cloud external account for permissions to sync correctly.
func NewProvider(url *url.URL, urn string, client *bitbucket.Client) *Provider {
	return &Provider{
		urn:      urn,
		codeHost: extsvc.NewCodeHost(url, extsvc.TypeBitbucketCloud),
		client:   client,
		pageSize: 1000,
	}
}

// ValidateConnection validates that the Provider has access to the Bitbucket Server API
// with the OAuth credentials it was configured with.
func (p *Provider) ValidateConnection(ctx context.Context) []string {
	return nil
}

func (p *Provider) URN() string {
	return p.urn
}

// ServiceID returns the absolute URL that identifies the Bitbucket Server instance
// this provider is configured with.
func (p *Provider) ServiceID() string { return p.codeHost.ServiceID }

// ServiceType returns the type of this Provider, namely, "bitbucketCloud".
func (p *Provider) ServiceType() string { return p.codeHost.ServiceType }

// FetchAccount satisfies the authz.Provider interface.
func (p *Provider) FetchAccount(ctx context.Context, user *types.User, _ []*extsvc.Account, _ []string) (acct *extsvc.Account, err error) {
	return nil, nil
}

// FetchUserPerms returns a list of repository IDs (on code host) that the given account
// has read access on the code host. The repository ID has the same value as it would be
// used as api.ExternalRepoSpec.ID. The returned list only includes private repository IDs.
//
// This method may return partial but valid results in case of error, and it is up to
// callers to decide whether to discard.
//
// API docs: https://docs.atlassian.com/bitbucket-server/rest/5.16.0/bitbucket-rest.html#idm8296923984
func (p *Provider) FetchUserPerms(ctx context.Context, account *extsvc.Account, opts authz.FetchPermsOptions) (*authz.ExternalUserPermissions, error) {
	switch {
	case account == nil:
		return nil, errors.New("no account provided")
	case !extsvc.IsHostOfAccount(p.codeHost, account):
		return nil, errors.Errorf("not a code host of the account: want %q but have %q",
			p.codeHost.ServiceID, account.AccountSpec.ServiceID)
	case account.Data == nil:
		return nil, errors.New("no account data provided")
	}

	// secret := ""
	// for _, authProvider := range conf.SiteConfig().AuthProviders {
	// 	if authProvider.Bitbucketcloud != nil &&
	// 		authProvider.Bitbucketcloud.ClientKey == account.ClientID {
	// 		secret = authProvider.Bitbucketcloud.ClientSecret
	// 	}
	// }

	_, tok, err := bitbucketcloud.GetExternalAccountData(ctx, &account.AccountData)
	if err != nil {
		return nil, err
	}
	bbClient := bitbucket.NewOAuthbearerToken(tok.AccessToken)
	bbClient.Pagelen = 100
	bbClient.SetApiBaseURL(*p.codeHost.BaseURL)

	repos, err := bbClient.Repositories.ListForAccount(&bitbucket.RepositoriesOptions{
		Role: "member",
	})
	if err != nil {
		return nil, err
	}

	extIDs := make([]extsvc.RepoID, 0, len(repos.Items))
	for _, repo := range repos.Items {
		extIDs = append(extIDs, extsvc.RepoID(repo.Uuid))
	}

	return &authz.ExternalUserPermissions{
		Exacts: extIDs,
	}, err
}

// FetchRepoPerms returns a list of user IDs (on code host) who have read access to
// the given repo on the code host. The user ID has the same value as it would
// be used as extsvc.Account.AccountID. The returned list includes both direct access
// and inherited from the group membership.
//
// This method may return partial but valid results in case of error, and it is up to
// callers to decide whether to discard.
//
// API docs: https://docs.atlassian.com/bitbucket-server/rest/5.16.0/bitbucket-rest.html#idm8283203728
func (p *Provider) FetchRepoPerms(ctx context.Context, repo *extsvc.Repository, opts authz.FetchPermsOptions) ([]extsvc.AccountID, error) {
	repoNameParts := strings.Split(repo.URI, "/")
	repoOwner := repoNameParts[1]
	repoName := repoNameParts[2]
	perms, err := p.client.Repositories.Repository.ListUserPermissions(&bitbucket.RepositoryOptions{
		Owner:    repoOwner,
		RepoSlug: repoName,
	})
	if err != nil {
		return nil, err
	}

	owner, err := p.client.User.Profile()
	if err != nil {
		return nil, err
	}

	userIDs := make([]extsvc.AccountID, 0, len(perms.UserPermissions)+1)
	for i := range perms.UserPermissions {
		userIDs = append(userIDs, extsvc.AccountID(perms.UserPermissions[i].User.AccountId))
	}

	userIDs = append(userIDs, extsvc.AccountID(owner.AccountId))

	return userIDs, nil
}
