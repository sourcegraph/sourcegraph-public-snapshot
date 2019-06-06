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
	log15 "gopkg.in/inconshreveable/log15.v2"
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

// NewProvider returns a new Bitbucket Server authorization provider that uses
// the given bitbucketserver.Client to talk to a Bitbucket Server API that is
// the source of truth for permissions. It assumes usernames of Sourcegraph accounts
// match 1-1 with usernames of Bitbucket Server API users.
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

func (p *Provider) RepoPerms(ctx context.Context, acct *extsvc.ExternalAccount, repos map[authz.Repo]struct{}) (map[api.RepoName]map[authz.Perm]bool, error) {
	var user bitbucketserver.User
	if acct != nil && acct.ServiceID == p.codeHost.ServiceID &&
		acct.ServiceType == p.codeHost.ServiceType {
		if err := json.Unmarshal(*acct.AccountData, &user); err != nil {
			return nil, err
		}
	}

	unverified, _ := p.Repos(ctx, repos)
	perms := make(map[api.RepoName]map[authz.Perm]bool, len(unverified))

	for repo := range unverified {
		ok, err := p.authorized(ctx, &user, authz.Read, &repo)
		if err != nil {
			log15.Error(
				"Failed to verify authorization for Bitbucket user",
				"user", user.Name,
				"perm", authz.Read,
				"repo", repo.RepoName,
				"error", err,
			)
		}
		perms[repo.RepoName] = map[authz.Perm]bool{authz.Read: ok}
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

func (p *Provider) authorized(ctx context.Context, user *bitbucketserver.User, perm authz.Perm, repo *authz.Repo) (bool, error) {
	meta, _ := repo.Metadata.(*bitbucketserver.Repo)
	if user.Name == "" { // Anonymous user
		if meta == nil || !meta.Public { // No metadata synced yet by repo-updater.
			return false, nil
		}
		return true, nil
	}

	// Authenticated user
	_, err := p.user(ctx, user.Name,
		bitbucketserver.UserFilter{
			Permission: bitbucketserver.PermissionFilter{
				Root:         bitbucketServerPerm(perm),
				RepositoryID: repo.ExternalRepoSpec.ID,
			},
		},
	)

	if err == errNoResults {
		return false, nil
	}

	return err == nil, err
}

var errNoResults = errors.New("no user found matching the given filters")

func (p *Provider) user(ctx context.Context, username string, fs ...bitbucketserver.UserFilter) (*bitbucketserver.User, error) {
	t := &bitbucketserver.PageToken{Limit: 10000}
	fs = append(fs, bitbucketserver.UserFilter{Filter: username})

	for t.HasMore() {
		users, next, err := p.api.Users(ctx, t, fs...)
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

func bitbucketServerPerm(p authz.Perm) bitbucketserver.Perm {
	switch p {
	case authz.Read:
		return bitbucketserver.PermRepoRead
	}
	panic("unknown permission")
}
