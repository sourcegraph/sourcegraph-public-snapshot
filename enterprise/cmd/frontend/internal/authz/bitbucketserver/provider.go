// Package bitbucketserver contains an authorization provider for Bitbucket Server.
package bitbucketserver

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/bitbucketserver"
)

// Provider is an implementation of AuthzProvider that provides repository permissions as
// determined from a Bitbucket Server instance API.
type Provider struct {
	client   *bitbucketserver.Client
	codeHost *extsvc.CodeHost
	pageSize int // Page size to use in paginated requests.
}

var _ authz.Provider = ((*Provider)(nil))

// NewProvider returns a new Bitbucket Server authorization provider that uses
// the given bitbucketserver.Client to talk to a Bitbucket Server API that is
// the source of truth for permissions. It assumes usernames of Sourcegraph accounts
// match 1-1 with usernames of Bitbucket Server API users.
func NewProvider(cli *bitbucketserver.Client) *Provider {
	return &Provider{
		client:   cli,
		codeHost: extsvc.NewCodeHost(cli.URL, bitbucketserver.ServiceType),
	}
}

func (p *Provider) Validate() []string  { return nil }
func (p *Provider) ServiceID() string   { return p.codeHost.ServiceID }
func (p *Provider) ServiceType() string { return p.codeHost.ServiceType }

func (p *Provider) Repos(ctx context.Context, repos map[authz.Repo]struct{}) (mine map[authz.Repo]struct{}, others map[authz.Repo]struct{}) {
	return authz.GetCodeHostRepos(p.codeHost, repos)
}

func (p *Provider) RepoPerms(ctx context.Context, acct *extsvc.ExternalAccount, repos map[authz.Repo]struct{}) (map[api.RepoName]map[authz.Perm]bool, error) {
	var user bitbucketserver.User
	if acct != nil && acct.ServiceID == p.codeHost.ServiceID &&
		acct.ServiceType == p.codeHost.ServiceType {
		if err := json.Unmarshal(*acct.AccountData, &user); err != nil {
			return nil, err
		}
	}

	unverified := make(map[string]api.RepoName, len(repos))
	for repo := range repos {
		unverified[repo.ExternalRepoSpec.ID] = repo.RepoName
	}

	authorized, err := p.repos(ctx, user.Name)
	if err != nil && err != errNoResults {
		return nil, err
	}

	perms := make(map[api.RepoName]map[authz.Perm]bool, len(authorized))
	for _, r := range authorized {
		if name, ok := unverified[strconv.Itoa(r.ID)]; ok {
			perms[name] = map[authz.Perm]bool{authz.Read: true}
		}
	}

	return perms, nil
}

// FetchAccount satisfies the authz.Provider interface.
func (p *Provider) FetchAccount(ctx context.Context, user *types.User, _ []*extsvc.ExternalAccount) (*extsvc.ExternalAccount, error) {
	if user == nil {
		return nil, nil
	}

	bitbucketUser, err := p.user(ctx, user.Username)
	if err != nil {
		return nil, err
	}

	accountData, err := json.Marshal(bitbucketUser)
	if err != nil {
		return nil, err
	}

	return &extsvc.ExternalAccount{
		UserID: user.ID,
		ExternalAccountSpec: extsvc.ExternalAccountSpec{
			ServiceType: p.codeHost.ServiceType,
			ServiceID:   p.codeHost.ServiceID,
			AccountID:   strconv.Itoa(bitbucketUser.ID),
		},
		ExternalAccountData: extsvc.ExternalAccountData{
			AccountData: (*json.RawMessage)(&accountData),
		},
	}, nil
}

var errNoResults = errors.New("no results returned by the Bitbucket Server API")

// repos returns all repositories for which the given user has the permission to read.
// when no username is given, only public repos are returned.
func (p *Provider) repos(ctx context.Context, username string) (all []*bitbucketserver.Repo, err error) {
	t := &bitbucketserver.PageToken{Limit: p.pageSize}
	c := p.client

	var filters []string
	if username == "" {
		filters = append(filters, "?visibility=public")
	} else {
		c = c.Sudo(username)
	}

	for t.HasMore() {
		repos, next, err := c.Repos(ctx, t, filters...)
		if err != nil {
			return nil, err
		}
		all = append(all, repos...)
		t = next
	}

	if len(all) == 0 {
		err = errNoResults
	}

	return all, err
}

func (p *Provider) user(ctx context.Context, username string, fs ...bitbucketserver.UserFilter) (*bitbucketserver.User, error) {
	t := &bitbucketserver.PageToken{Limit: p.pageSize}
	fs = append(fs, bitbucketserver.UserFilter{Filter: username})

	for t.HasMore() {
		users, next, err := p.client.Users(ctx, t, fs...)
		if err != nil {
			return nil, err
		}

		for _, u := range users {
			if u.Name == username {
				return u, nil
			}
		}

		t = next
	}

	return nil, errNoResults
}
