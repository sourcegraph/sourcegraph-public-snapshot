// Package bitbucketserver contains an authorization provider for Bitbucket Server.
package bitbucketserver

import (
	"context"
	"database/sql"
	"encoding/json"
	"strconv"
	"time"

	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	iauthz "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

// Provider is an implementation of AuthzProvider that provides repository permissions as
// determined from a Bitbucket Server instance API.
type Provider struct {
	client   *bitbucketserver.Client
	codeHost *extsvc.CodeHost
	pageSize int // Page size to use in paginated requests.
	store    *store
}

var _ authz.Provider = (*Provider)(nil)

var clock = func() time.Time { return time.Now().UTC().Truncate(time.Microsecond) }

// NewProvider returns a new Bitbucket Server authorization provider that uses
// the given bitbucketserver.Client to talk to a Bitbucket Server API that is
// the source of truth for permissions. It assumes usernames of Sourcegraph accounts
// match 1-1 with usernames of Bitbucket Server API users.
func NewProvider(cli *bitbucketserver.Client, db *sql.DB, ttl, hardTTL time.Duration) *Provider {
	return &Provider{
		client:   cli,
		codeHost: extsvc.NewCodeHost(cli.URL, bitbucketserver.ServiceType),
		pageSize: 1000,
		store:    newStore(db, ttl, hardTTL, clock),
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

// RepoPerms returns the permissions the given external account has in relation to the given set of repos.
// It performs a single HTTP request against the Bitbucket Server API which returns all repositories
// the authenticated user has permissions to read.
func (p *Provider) RepoPerms(ctx context.Context, acct *extsvc.ExternalAccount, repos []*types.Repo) (
	perms []authz.RepoPerms,
	err error,
) {
	var (
		userName string
		userID   int32
	)

	tr, ctx := trace.New(ctx, "bitbucket.authz.provider.RepoPerms", "")
	defer func() {
		tr.LogFields(
			otlog.String("user.name", userName),
			otlog.Int32("user.id", userID),
			otlog.Int("repos.count", len(repos)),
			otlog.Int("perms.count", len(perms)),
		)

		if err != nil {
			tr.SetError(err)
		}

		tr.Finish()
	}()

	if acct != nil && acct.ServiceID == p.codeHost.ServiceID && acct.ServiceType == p.codeHost.ServiceType {
		var user bitbucketserver.User
		if err := json.Unmarshal(*acct.AccountData, &user); err != nil {
			return nil, err
		}

		userID = acct.UserID
		userName = user.Name
	}

	ps := &iauthz.UserPermissions{
		UserID:   userID,
		Perm:     authz.Read,
		Type:     iauthz.PermRepos,
		Provider: iauthz.ProviderBitbucketServer,
	}

	err = p.store.LoadPermissions(ctx, ps, p.update(userName))
	if err != nil {
		return nil, err
	}

	return ps.AuthorizedRepos(repos), nil
}

// UpdatePermissions forces an update of the permissions of the given
// user.
func (p *Provider) UpdatePermissions(ctx context.Context, u *types.User) error {
	ps := &iauthz.UserPermissions{
		UserID:   u.ID,
		Perm:     authz.Read,
		Type:     iauthz.PermRepos,
		Provider: iauthz.ProviderBitbucketServer,
	}

	return p.store.UpdatePermissions(ctx, ps, p.update(u.Username))
}

// update returns a PermissionsUpdateFunc that fetches the IDs of
// all the repos the user with the given userName is authorized to
// see.
func (p *Provider) update(userName string) PermissionsUpdateFunc {
	return func(ctx context.Context) ([]uint32, *extsvc.CodeHost, error) {
		visible, err := p.repos(ctx, userName)
		if err != nil && err != errNoResults {
			return nil, p.codeHost, err
		}

		ids := make([]uint32, 0, len(visible))
		for _, r := range visible {
			ids = append(ids, uint32(r.ID))
		}

		return ids, p.codeHost, nil
	}
}

// FetchAccount satisfies the authz.Provider interface.
func (p *Provider) FetchAccount(ctx context.Context, user *types.User, _ []*extsvc.ExternalAccount) (acct *extsvc.ExternalAccount, err error) {
	if user == nil {
		return nil, nil
	}

	tr, ctx := trace.New(ctx, "bitbucket.authz.provider.FetchAccount", "")
	defer func() {
		tr.LogFields(
			otlog.String("user.name", user.Username),
			otlog.Int32("user.id", user.ID),
		)

		if err != nil {
			tr.SetError(err)
		}

		tr.Finish()
	}()

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
