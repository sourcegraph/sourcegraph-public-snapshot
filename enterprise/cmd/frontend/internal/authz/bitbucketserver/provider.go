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
	api      bitbucketServerAPI
	codeHost *extsvc.CodeHost
}

// bitbucketServerAPI captures the client interface to a Bitbucket Server API.
type bitbucketServerAPI interface {
	// Users retrieves a page of Bitbucket Server users matching
	// the given filters.
	Users(context.Context, *bitbucketserver.PageToken, ...bitbucketserver.UserFilter) (
		[]*bitbucketserver.User, *bitbucketserver.PageToken, error)
}

var _ authz.Provider = ((*Provider)(nil))

func NewProvider(cli *bitbucketserver.Client) *Provider {
	return &Provider{
		api:      cli,
		codeHost: extsvc.NewCodeHost(cli.URL, bitbucketserver.ServiceType),
	}
}

func (p *Provider) Validate() (problems []string) {
	// TODO(tsenart): Validate that the access token has the right permissions with the API
	return nil
}

func (p *Provider) ServiceID() string   { return p.codeHost.ServiceID }
func (p *Provider) ServiceType() string { return p.codeHost.ServiceType }

func (p *Provider) Repos(ctx context.Context, repos map[authz.Repo]struct{}) (mine map[authz.Repo]struct{}, others map[authz.Repo]struct{}) {
	return authz.GetCodeHostRepos(p.codeHost, repos)
}

func (p *Provider) RepoPerms(ctx context.Context, account *extsvc.ExternalAccount, repos map[authz.Repo]struct{}) (map[api.RepoName]map[authz.Perm]bool, error) {
	return nil, nil
}

// repos returns all Bitbucket Server repos that for which the given user has the given permission.
func (p *Provider) repos(username string, perm bitbucketserver.Perm) ([]*bitbucketserver.Repo, error) {
	return nil, nil
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

func (p *Provider) user(ctx context.Context, username string) (*bitbucketserver.User, error) {
	t := &bitbucketserver.PageToken{Limit: 10000}
	f := bitbucketserver.UserFilter{Filter: username}

	for t.HasMore() {
		users, next, err := p.api.Users(ctx, t, f)
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

	return nil, errors.Errorf("Bitbucket Server user with username=%q not found", username)
}
