package bitbucketserver

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/bitbucketserver"
)

var update = flag.Bool("update", false, "update testdata")

func TestProvider_RepoPerms(t *testing.T) {
	cli, save := bitbucketserver.NewTestClient(t, "RepoPerms", *update)
	defer save()

	key, err := base64.StdEncoding.DecodeString(os.Getenv("BITBUCKET_SERVER_KEY"))
	if err == nil {
		if err := cli.SetOAuth("sourcegraph", key); err != nil {
			t.Fatal(err)
		}
	}

	codeHost := extsvc.CodeHost{
		ServiceType: bitbucketserver.ServiceType,
		ServiceID:   os.Getenv("BITBUCKET_SERVER_URL"),
	}

	externalRepo := func(r *bitbucketserver.Repo) api.ExternalRepoSpec {
		return api.ExternalRepoSpec{
			ServiceType: codeHost.ServiceType,
			ServiceID:   codeHost.ServiceID,
			ID:          strconv.Itoa(r.ID),
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

	ctx := context.Background()
	users := map[string]*bitbucketserver.User{}

	for _, u := range []*bitbucketserver.User{
		{Name: "john1", DisplayName: "Mr. John 1", EmailAddress: "john1@mycorp.com"},
		{Name: "john2", DisplayName: "Mr. John 2", EmailAddress: "john2@mycorp.com"},
		{Name: "john3", DisplayName: "Mr. John 3", EmailAddress: "john3@mycorp.com"},
	} {
		u.Password = "password"
		if err := cli.CreateUser(ctx, u); err != nil {
			t.Fatal(err)
		}
		users[u.Name] = u
	}

	// projects := map[string]*bitbucketserver.Project{}
	repos := map[string]*bitbucketserver.Repo{}

	for _, tc := range []struct {
		name  string
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
					ExternalRepoSpec: externalRepo(repos["foo"]),
				},
				{
					RepoName:         "bar",
					ExternalRepoSpec: externalRepo(repos["bar"]),
				},
				{
					RepoName:         "baz",
					ExternalRepoSpec: externalRepo(repos["baz"]),
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
			acct: externalAccount(users["john4"]),
			repos: []authz.Repo{
				{
					RepoName:         "foo",
					ExternalRepoSpec: externalRepo(repos["foo"]),
				},
				{
					RepoName:         "bar",
					ExternalRepoSpec: externalRepo(repos["bar"]),
				},
				{
					RepoName:         "baz",
					ExternalRepoSpec: externalRepo(repos["baz"]),
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
			acct: externalAccount(users["john4"]),
			repos: []authz.Repo{
				{
					RepoName:         "foo",
					ExternalRepoSpec: externalRepo(repos["foo"]),
				},
				{
					RepoName:         "bar",
					ExternalRepoSpec: externalRepo(repos["bar"]),
				},
				{
					RepoName:         "baz",
					ExternalRepoSpec: externalRepo(repos["baz"]),
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

			p := Provider{client: cli, codeHost: &codeHost}
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
	cli, save := bitbucketserver.NewTestClient(t, "FetchAccount", *update)
	defer save()

	codeHost := extsvc.CodeHost{
		ServiceType: bitbucketserver.ServiceType,
		ServiceID:   "https://bitbucketserver.mycorp.com",
	}

	for _, tc := range []struct {
		name string
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
			user: &types.User{Username: "john"},
			acct: nil,
			err:  `no user found matching the given filters`,
		},
		{
			name: "user found by exact username match",
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

			p := Provider{client: cli, codeHost: &codeHost}
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

func marshalJSON(v interface{}) []byte {
	bs, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return bs
}
