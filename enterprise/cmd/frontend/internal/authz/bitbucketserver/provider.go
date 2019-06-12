// Package bitbucketserver contains an authorization provider for Bitbucket Server.
package bitbucketserver

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/RoaringBitmap/roaring"
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
	store    store
	ttl      time.Duration
	clock    func() time.Time
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

// Validate validates that the Provider has access to the Bitbucket Server API
// with the OAuth credentials it was configured with.
func (p *Provider) Validate() []string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := p.client.UserPermissions(ctx, p.client.Username)
	if err != nil {
		return []string{err.Error()}
	}

	return nil
}

// ServiceID returns the absolute URL that identifies the Bitbucket Server instance
// this provider is configured with.
func (p *Provider) ServiceID() string { return p.codeHost.ServiceID }

// ServiceType returns the type of this Provider, namely, "bitbucketServer".
func (p *Provider) ServiceType() string { return p.codeHost.ServiceType }

// Repos returns all Bitbucket Server repos as mine, and all others as others.
func (p *Provider) Repos(ctx context.Context, repos map[authz.Repo]struct{}) (mine map[authz.Repo]struct{}, others map[authz.Repo]struct{}) {
	return authz.GetCodeHostRepos(p.codeHost, repos)
}

// RepoPerms returns the permissions the given external account has in relation to the given set of repos.
// It performs a single HTTP request against the Bitbucket Server API which returns all repositories
// the authenticated user has permissions to read.
func (p *Provider) RepoPerms(ctx context.Context, acct *extsvc.ExternalAccount, repos map[authz.Repo]struct{}) (map[api.RepoName]map[authz.Perm]bool, error) {
	var (
		user   bitbucketserver.User
		acctID int32
	)

	if acct != nil && acct.ServiceID == p.codeHost.ServiceID &&
		acct.ServiceType == p.codeHost.ServiceType {
		acctID = acct.ID
		if err := json.Unmarshal(*acct.AccountData, &user); err != nil {
			return nil, err
		}
	}

	var unverified roaring.Bitmap
	names := make(map[uint32]api.RepoName, len(repos))
	for r := range repos {
		id, err := strconv.ParseUint(r.ExternalRepoSpec.ID, 10, 32)
		if err != nil {
			return nil, err
		}

		unverified.Add(uint32(id))
		names[uint32(id)] = r.RepoName
	}

	authorized, err := p.authorized(ctx, acctID, user.Name, &unverified)
	if err != nil && err != errNoResults {
		return nil, err
	}

	iter := authorized.Iterator()
	perms := make(map[api.RepoName]map[authz.Perm]bool, authorized.GetCardinality())

	for iter.HasNext() {
		perms[names[iter.Next()]] =
			map[authz.Perm]bool{authz.Read: true}
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

func (p *Provider) authorized(ctx context.Context, acctID int32, username string, unverified *roaring.Bitmap) (*roaring.Bitmap, error) {
	perms := Permissions{
		AccountID: acctID,
		Perm:      authz.Read,
		Type:      "repos",
	}

	// TODO: Wrap in serializable transaction to control cache filling concurrency.

	if err := p.store.LoadPermissions(ctx, &perms); err != nil {
		return nil, err
	}

	now := p.now()

	if now.Equal(perms.ExpiredAt) || now.After(perms.ExpiredAt) { // Cache expired
		repos, err := p.repos(ctx, username)
		if err != nil {
			return nil, err
		}

		perms.IDs = roaring.NewBitmap()
		for _, r := range repos {
			perms.IDs.Add(uint32(r.ID))
		}

		perms.UpdatedAt = now
		perms.ExpiredAt = now.Add(p.ttl)

		if err := p.store.UpsertPermissions(ctx, &perms); err != nil {
			return nil, err
		}
	}

	return roaring.And(unverified, perms.IDs), nil
}

var errNoResults = errors.New("no results returned by the Bitbucket Server API")

// repos returns all repositories for which the given user has the permission to read from
// the Bitbucket Server API. when no username is given, only public repos are returned.
func (p *Provider) repos(ctx context.Context, username string) (all []*bitbucketserver.Repo, err error) {
	t := &bitbucketserver.PageToken{Limit: p.pageSize}
	c := p.client

	var filters []string
	if username == "" {
		filters = append(filters, "?visibility=public")
	} else if c, err = c.Sudo(username); err != nil {
		return nil, err
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

func (p *Provider) now() time.Time {
	if p.clock != nil {
		return p.clock()
	}
	return time.Now()
}
