package bitbucketserver

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/bitbucketserver"
)

func TestProvider_RepoPerms(t *testing.T) {
	codeHost := extsvc.CodeHost{
		ServiceType: bitbucketserver.ServiceType,
		ServiceID:   "https://bitbucketserver.mycorp.com",
	}

	externalRepo := func(id string) api.ExternalRepoSpec {
		return api.ExternalRepoSpec{
			ServiceType: codeHost.ServiceType,
			ServiceID:   codeHost.ServiceID,
			ID:          id,
		}
	}

	externalAccount := func(u *bitbucketserver.User) *extsvc.ExternalAccount {
		bs := marshalJSON(u)
		return &extsvc.ExternalAccount{
			ExternalAccountSpec: extsvc.ExternalAccountSpec{
				ServiceType: codeHost.ServiceType,
				ServiceID:   codeHost.ServiceID,
				AccountID:   strconv.Itoa(u.ID),
			},
			ExternalAccountData: extsvc.ExternalAccountData{
				AccountData: (*json.RawMessage)(&bs),
			},
		}
	}

	for _, tc := range []struct {
		name  string
		api   fakeBitbucketServerAPI
		ctx   context.Context
		acct  *extsvc.ExternalAccount
		repos []authz.Repo
		perms map[api.RepoName]map[authz.Perm]bool
		err   string
	}{
		{
			name: "anonymous user",
			acct: nil,
			repos: []authz.Repo{
				{
					RepoName:         "foo",
					ExternalRepoSpec: externalRepo("foo"),
				},
				{
					RepoName:         "bar",
					ExternalRepoSpec: externalRepo("bar"),
					Metadata:         &bitbucketserver.Repo{Public: false},
				},
				{
					RepoName:         "baz",
					ExternalRepoSpec: externalRepo("baz"),
					Metadata:         &bitbucketserver.Repo{Public: true},
				},
			},
			perms: map[api.RepoName]map[authz.Perm]bool{
				"foo": {authz.Read: false},
				"bar": {authz.Read: false},
				"baz": {authz.Read: true},
			},
		},
		{
			name: "authenticated user",
			api: fakeBitbucketServerAPI{
				perms: map[string]map[string]bitbucketserver.Perm{
					"john": {
						"foo": bitbucketserver.PermRepoRead,
						"bar": bitbucketserver.PermRepoRead,
					},
				},
				users: []*bitbucketserver.User{
					{ID: 1, EmailAddress: "john.doe@mycorp.com"},
					{ID: 2, DisplayName: "mr. john"},
					{ID: 3, Name: "john-doe"},
					{ID: 4, Name: "john"}, // This one should be returned.
				},
			},
			acct: externalAccount(&bitbucketserver.User{ID: 4, Name: "john"}),
			repos: []authz.Repo{
				{
					RepoName:         "foo",
					ExternalRepoSpec: externalRepo("foo"),
				},
				{
					RepoName:         "bar",
					ExternalRepoSpec: externalRepo("bar"),
				},
				{
					RepoName:         "baz",
					ExternalRepoSpec: externalRepo("baz"),
				},
			},
			perms: map[api.RepoName]map[authz.Perm]bool{
				"foo": {authz.Read: true},
				"bar": {authz.Read: true},
				"baz": {authz.Read: false},
			},
		},
		{
			// When failing to contact the API, block access by default.
			name: "authenticated user - errors",
			api: fakeBitbucketServerAPI{
				err: errors.New("boom"),
				perms: map[string]map[string]bitbucketserver.Perm{
					"john": {
						"foo": bitbucketserver.PermRepoRead,
						"bar": bitbucketserver.PermRepoRead,
					},
				},
				users: []*bitbucketserver.User{{ID: 4, Name: "john"}},
			},
			acct: externalAccount(&bitbucketserver.User{ID: 4, Name: "john"}),
			repos: []authz.Repo{
				{
					RepoName:         "foo",
					ExternalRepoSpec: externalRepo("foo"),
				},
				{
					RepoName:         "bar",
					ExternalRepoSpec: externalRepo("bar"),
				},
				{
					RepoName:         "baz",
					ExternalRepoSpec: externalRepo("baz"),
				},
			},
			perms: map[api.RepoName]map[authz.Perm]bool{
				"foo": {authz.Read: false},
				"bar": {authz.Read: false},
				"baz": {authz.Read: false},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.ctx == nil {
				tc.ctx = context.Background()
			}

			if tc.err == "" {
				tc.err = "<nil>"
			}

			repos := make(map[authz.Repo]struct{}, len(tc.repos))
			for _, r := range tc.repos {
				repos[r] = struct{}{}
			}

			p := Provider{api: tc.api, codeHost: &codeHost}
			perms, err := p.RepoPerms(tc.ctx, tc.acct, repos)

			if have, want := fmt.Sprint(err), tc.err; have != want {
				t.Errorf("error:\nhave: %q\nwant: %q", have, want)
			}

			if have, want := perms, tc.perms; !reflect.DeepEqual(have, want) {
				t.Error(cmp.Diff(have, want))
			}
		})
	}
}

func TestProvider_FetchAccount(t *testing.T) {
	codeHost := extsvc.CodeHost{
		ServiceType: bitbucketserver.ServiceType,
		ServiceID:   "https://bitbucketserver.mycorp.com",
	}

	for _, tc := range []struct {
		name string
		api  fakeBitbucketServerAPI
		ctx  context.Context
		user *types.User
		acct *extsvc.ExternalAccount
		err  string
	}{
		{
			name: "no user given",
			user: nil,
			acct: nil,
		},
		{
			name: "user not found",
			api: fakeBitbucketServerAPI{
				users: []*bitbucketserver.User{
					{EmailAddress: "john.doe@mycorp.com"},
					{DisplayName: "mr. john"},
					{Name: "john-doe"},
				},
			},
			user: &types.User{Username: "john"},
			acct: nil,
			err:  `no user found matching the given filters`,
		},
		{
			name: "user found by exact username match",
			api: fakeBitbucketServerAPI{
				users: []*bitbucketserver.User{
					{ID: 1, EmailAddress: "john.doe@mycorp.com"},
					{ID: 2, DisplayName: "mr. john"},
					{ID: 3, Name: "john-doe"},
					{ID: 4, Name: "john"}, // This one should be returned.
				},
			},
			user: &types.User{ID: 42, Username: "john"},
			acct: &extsvc.ExternalAccount{
				UserID: 42,
				ExternalAccountSpec: extsvc.ExternalAccountSpec{
					ServiceType: codeHost.ServiceType,
					ServiceID:   codeHost.ServiceID,
					AccountID:   strconv.Itoa(4),
				},
				ExternalAccountData: extsvc.ExternalAccountData{
					AccountData: func() *json.RawMessage {
						bs, err := json.Marshal(&bitbucketserver.User{ID: 4, Name: "john"})
						if err != nil {
							panic(err)
						}
						return (*json.RawMessage)(&bs)
					}(),
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.ctx == nil {
				tc.ctx = context.Background()
			}

			if tc.err == "" {
				tc.err = "<nil>"
			}

			p := Provider{api: tc.api, codeHost: &codeHost}
			acct, err := p.FetchAccount(tc.ctx, tc.user, nil)

			if have, want := fmt.Sprint(err), tc.err; have != want {
				t.Errorf("error:\nhave: %q\nwant: %q", have, want)
			}

			if have, want := acct, tc.acct; !reflect.DeepEqual(have, want) {
				t.Error(cmp.Diff(have, want))
			}
		})
	}
}

type fakeBitbucketServerAPI struct {
	// username -> repo id -> permission
	perms map[string]map[string]bitbucketserver.Perm
	users []*bitbucketserver.User
	err   error // Error to be returned in a Users call
}

func (api fakeBitbucketServerAPI) Users(
	ctx context.Context,
	pt *bitbucketserver.PageToken,
	fs ...bitbucketserver.UserFilter,
) (
	users []*bitbucketserver.User,
	_ *bitbucketserver.PageToken,
	err error,
) {
	for _, u := range api.users {
		ok := true
		for _, f := range fs {
			ok = ok && api.match(f, u)
		}
		if ok {
			users = append(users, u)
		}
	}

	// Pretend like the maxium allowed limit is 1 so that
	// we exercise the pagination logic.
	next := &bitbucketserver.PageToken{Limit: 1}

	lo, hi := pt.NextPageStart, pt.NextPageStart+1
	if lo > len(users) {
		next.IsLastPage = true
		return nil, next, nil
	}

	if hi > len(users) {
		hi = len(users)
		next.IsLastPage = true
	} else {
		next.NextPageStart = hi
	}

	page := users[lo:hi]
	next.Size = len(page)
	next.Start = lo

	return page, next, api.err
}

func (api fakeBitbucketServerAPI) match(f bitbucketserver.UserFilter, u *bitbucketserver.User) bool {
	if f.Filter != "" {
		return strings.Contains(u.Name, f.Filter) ||
			strings.Contains(u.EmailAddress, f.Filter) ||
			strings.Contains(u.DisplayName, f.Filter)
	}

	if f.Permission != (bitbucketserver.PermissionFilter{}) {
		if repos := api.perms[u.Name]; repos != nil {
			return repos[f.Permission.RepositoryID] == f.Permission.Root
		}
		return false
	}

	panic("filter not handled")
}

func marshalJSON(v interface{}) []byte {
	bs, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return bs
}
