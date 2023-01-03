// Package bitbucketcloud contains an authorization provider for Bitbucket Cloud.
package bitbucketcloud

import (
	"context"
	"net/url"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketcloud"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Provider is an implementation of AuthzProvider that provides repository and
// user permissions as determined from Bitbucket Cloud.
type Provider struct {
	urn      string
	codeHost *extsvc.CodeHost
	client   bitbucketcloud.Client
	pageSize int // Page size to use in paginated requests.
}

type ProviderOptions struct {
	BitbucketCloudClient bitbucketcloud.Client
}

var _ authz.Provider = (*Provider)(nil)

// NewProvider returns a new Bitbucket Cloud authorization provider that uses
// the given bitbucket.Client to talk to the Bitbucket Cloud API that is
// the source of truth for permissions. Sourcegraph users will need a valid
// Bitbucket Cloud external account for permissions to sync correctly.
func NewProvider(conn *types.BitbucketCloudConnection, opts ProviderOptions) *Provider {
	baseURL, err := url.Parse(conn.Url)
	if err != nil {
		return nil
	}

	if opts.BitbucketCloudClient == nil {
		opts.BitbucketCloudClient, err = bitbucketcloud.NewClient(conn.Url, conn.BitbucketCloudConnection, httpcli.ExternalClient)
	}

	return &Provider{
		urn:      conn.URN,
		codeHost: extsvc.NewCodeHost(baseURL, extsvc.TypeBitbucketCloud),
		client:   opts.BitbucketCloudClient,
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
	oauthToken := &auth.OAuthBearerToken{
		Token:        tok.AccessToken,
		RefreshToken: tok.RefreshToken,
		Expiry:       tok.Expiry,
	}
	client := p.client.WithAuthenticator(oauthToken)
	repos, next, err := client.Repos(ctx, nil, "", &bitbucketcloud.ReposOptions{Role: "member"})
	if err != nil {
		return nil, err
	}
	for next.HasMore() {
		var nextRepos []*bitbucketcloud.Repo
		nextRepos, next, err = client.Repos(ctx, next, "", &bitbucketcloud.ReposOptions{Role: "member"})
		if err != nil {
			return nil, err
		}
		repos = append(repos, nextRepos...)
	}

	extIDs := make([]extsvc.RepoID, 0, len(repos))
	for _, repo := range repos {
		extIDs = append(extIDs, extsvc.RepoID(repo.UUID))
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
	users, next, err := p.client.ListExplicitUserPermsForRepo(ctx, nil, repoOwner, repoName)
	if err != nil {
		return nil, err
	}

	for next.HasMore() {
		var nextUsers []*bitbucketcloud.Account
		nextUsers, next, err = p.client.ListExplicitUserPermsForRepo(ctx, next, repoOwner, repoName)
		if err != nil {
			return nil, err
		}
		users = append(users, nextUsers...)
	}

	// Bitbucket Cloud API does not return the owner of the repository as part
	// of the explicit permissions list, so we need to fetch and add them.
	bbCloudRepo, err := p.client.Repo(ctx, repoOwner, repoName)
	if err != nil {
		return nil, err
	}

	userIDs := make([]extsvc.AccountID, 0, len(users)+1)
	for i := range users {
		userIDs = append(userIDs, extsvc.AccountID(users[i].UUID))
	}

	userIDs = append(userIDs, extsvc.AccountID(bbCloudRepo.Owner.UUID))

	return userIDs, nil
}
