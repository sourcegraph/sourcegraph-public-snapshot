package bitbucketserver

import (
	"context"
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

func TestProvider_Validate(t *testing.T) {
	for _, tc := range []struct {
		name     string
		client   func(*bitbucketserver.Client)
		problems []string
	}{
		{
			name: "no-problems-when-authenticated-as-admin",
		},
		{
			name:   "problems-when-authenticated-as-non-admin",
			client: func(c *bitbucketserver.Client) { c.Oauth = nil },
			problems: []string{
				"unexpected 401 response from BitBucket Server REST API at http://127.0.0.1:7990/rest/api/1.0/admin/permissions/users?filter=",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p, save := newProvider(t, "Validate/"+tc.name)
			defer save()

			if tc.client != nil {
				tc.client(p.client)
			}

			problems := p.Validate()
			if have, want := problems, tc.problems; !reflect.DeepEqual(have, want) {
				t.Error(cmp.Diff(have, want))
			}
		})
	}
}

func TestProvider_RepoPerms(t *testing.T) {
	p, save := newProvider(t, "RepoPerms")
	defer save()

	f := newFixtures()
	f.load(t, p.client)

	h := codeHost{CodeHost: p.codeHost}

	for _, tc := range []struct {
		name  string
		ctx   context.Context
		acct  *extsvc.ExternalAccount
		perms map[api.RepoName]map[authz.Perm]bool
		err   string
	}{
		{
			name: "anonymous user",
			acct: nil,
			perms: map[api.RepoName]map[authz.Perm]bool{
				// Because repo is public
				"public-repo": {authz.Read: true},
			},
		},
		{
			name: "authenticated user: engineer1",
			acct: h.externalAccount(2, f.users["engineer1"]),
			perms: map[api.RepoName]map[authz.Perm]bool{
				// Because repo is public
				"public-repo": {authz.Read: true},
				// Because of engineer1 has a secret-project group membership
				// and secret-project group has PROJECT_READ perm on SECRET project
				// which secret-repo belongs to.
				"secret-repo": {authz.Read: true},
				// Because engineers group has PROJECT_WRITE perm on PRIVATE project
				// which private-repo belongs to.
				"private-repo": {authz.Read: true},
			},
		},
		{
			name: "authenticated user: engineer2",
			acct: h.externalAccount(3, f.users["engineer2"]),
			perms: map[api.RepoName]map[authz.Perm]bool{
				// Because repo is public
				"public-repo": {authz.Read: true},
				// Because engineers group has PROJECT_WRITE perm on PRIVATE project
				// which private-repo belongs to.
				"private-repo": {authz.Read: true}, // Because of engineers group membership
			},
		},
		{
			name: "authenticated user: scientist",
			acct: h.externalAccount(4, f.users["scientist"]),
			perms: map[api.RepoName]map[authz.Perm]bool{
				// Because repo is public
				"public-repo": {authz.Read: true},
				// Because of scientist1 has a secret-project group membership
				// and secret-project group has PROJECT_READ perm on SECRET project
				// which secret-repo belongs to.
				"secret-repo": {authz.Read: true},
				// Because scientists group has PROJECT_READ perm on PRIVATE project
				// which private-repo belongs to.
				"private-repo": {authz.Read: true},
			},
		},
		{
			name: "authenticated user: ceo",
			acct: h.externalAccount(5, f.users["ceo"]),
			perms: map[api.RepoName]map[authz.Perm]bool{
				// Because repo is public
				"public-repo": {authz.Read: true},
				// Because management group has PROJECT_READ perm on PRIVATE project
				// which private-repo belongs to.
				"secret-repo": {authz.Read: true},
				// Because ceo has REPO_WRITE perm on super-secret-repo.
				"super-secret-repo": {authz.Read: true},
				// Because management group has PROJECT_READ perm on PRIVATE project
				// which private-repo belongs to.
				"private-repo": {authz.Read: true},
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

			repos := make(map[authz.Repo]struct{}, len(f.repos))
			for _, r := range f.repos {
				repos[authz.Repo{
					RepoName:         api.RepoName(r.Name),
					ExternalRepoSpec: h.externalRepo(r),
				}] = struct{}{}
			}

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
	p, save := newProvider(t, "FetchAccount")
	defer save()

	f := newFixtures()
	f.load(t, p.client)

	h := codeHost{CodeHost: p.codeHost}

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
			err:  `no results returned by the Bitbucket Server API`,
		},
		{
			name: "user found by exact username match",
			user: &types.User{ID: 42, Username: "ceo"},
			acct: h.externalAccount(42, f.users["ceo"]),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.ctx == nil {
				tc.ctx = context.Background()
			}

			if tc.err == "" {
				tc.err = "<nil>"
			}

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

// fixtures contains the data we need loaded in Bitbucket Server API
// to run the Provider tests. Because we use VCR recordings, we don't
// need a Bitbucket Server API up and running to run those tests. But if
// you want to work on these tests / code, you need to start a new instance
// of Bitbucket Server with docker, create an Application Link as per
// https://docs.sourcegraph.com/admin/external_service/bitbucket_server, and
// then run the tests with -update=true.
type fixtures struct {
	users             map[string]*bitbucketserver.User
	groups            map[string]*bitbucketserver.Group
	projects          map[string]*bitbucketserver.Project
	repos             map[string]*bitbucketserver.Repo
	groupProjectPerms []*bitbucketserver.GroupProjectPermission
	userRepoPerms     []*bitbucketserver.UserRepoPermission
}

func (f fixtures) load(t *testing.T, cli *bitbucketserver.Client) {
	ctx := context.Background()

	for _, u := range f.users {
		u.Password = "password"
		u.Slug = u.Name

		if err := cli.LoadUser(ctx, u); err != nil {
			t.Log(err)
			if err := cli.CreateUser(ctx, u); err != nil {
				t.Error(err)
			}
		}
	}

	for _, g := range f.groups {
		if err := cli.LoadGroup(ctx, g); err != nil {
			t.Log(err)

			if err := cli.CreateGroup(ctx, g); err != nil {
				t.Error(err)
			}

			if err := cli.CreateGroupMembership(ctx, g); err != nil {
				t.Error(err)
			}
		}
	}

	for _, p := range f.projects {
		if err := cli.LoadProject(ctx, p); err != nil {
			t.Log(err)
			if err := cli.CreateProject(ctx, p); err != nil {
				t.Error(err)
			}
		}
	}

	for _, r := range f.repos {
		repo, err := cli.Repo(ctx, r.Project.Key, r.Slug)
		if err != nil {
			t.Log(err)
			if err := cli.CreateRepo(ctx, r); err != nil {
				t.Error(err)
			}
		} else {
			*r = *repo
		}
	}

	for _, p := range f.groupProjectPerms {
		if err := cli.CreateGroupProjectPermission(ctx, p); err != nil {
			t.Error(err)
		}
	}

	for _, p := range f.userRepoPerms {
		if err := cli.CreateUserRepoPermission(ctx, p); err != nil {
			t.Error(err)
		}
	}
}

func newFixtures() *fixtures {
	users := map[string]*bitbucketserver.User{
		"engineer1": {Name: "engineer1", DisplayName: "Mr. Engineer 1", EmailAddress: "engineer1@mycorp.com"},
		"engineer2": {Name: "engineer2", DisplayName: "Mr. Engineer 2", EmailAddress: "engineer2@mycorp.com"},
		"scientist": {Name: "scientist", DisplayName: "Ms. Scientist", EmailAddress: "scientist@mycorp.com"},
		"ceo":       {Name: "ceo", DisplayName: "Mrs. CEO", EmailAddress: "ceo@mycorp.com"},
	}

	groups := map[string]*bitbucketserver.Group{
		"engineers":      {Name: "engineers", Users: []string{"engineer1", "engineer2"}},
		"scientists":     {Name: "scientists", Users: []string{"scientist"}},
		"secret-project": {Name: "secret-project", Users: []string{"engineer1", "scientist"}},
		"management":     {Name: "management", Users: []string{"ceo"}},
	}

	projects := map[string]*bitbucketserver.Project{
		"SECRET":      {Name: "Secret", Key: "SECRET", Public: false},
		"SUPERSECRET": {Name: "Super Secret", Key: "SUPERSECRET", Public: false},
		"PRIVATE":     {Name: "Private", Key: "PRIVATE", Public: false},
		"PUBLIC":      {Name: "Public", Key: "PUBLIC", Public: true},
	}

	repos := map[string]*bitbucketserver.Repo{
		"secret-repo":       {Slug: "secret-repo", Name: "secret-repo", Project: projects["SECRET"]},
		"super-secret-repo": {Slug: "super-secret-repo", Name: "super-secret-repo", Project: projects["SUPERSECRET"]},
		"public-repo":       {Slug: "public-repo", Name: "public-repo", Project: projects["PUBLIC"]},
		"private-repo":      {Slug: "private-repo", Name: "private-repo", Project: projects["PRIVATE"]},
	}

	groupProjectPerms := []*bitbucketserver.GroupProjectPermission{
		{
			Group:   groups["engineers"],
			Perm:    bitbucketserver.PermProjectWrite,
			Project: projects["PRIVATE"],
		},
		{
			Group:   groups["engineers"],
			Perm:    bitbucketserver.PermProjectWrite,
			Project: projects["PUBLIC"],
		},
		{
			Group:   groups["scientists"],
			Perm:    bitbucketserver.PermProjectRead,
			Project: projects["PRIVATE"],
		},
		{
			Group:   groups["secret-project"],
			Perm:    bitbucketserver.PermProjectWrite,
			Project: projects["SECRET"],
		},
		{
			Group:   groups["management"],
			Perm:    bitbucketserver.PermProjectRead,
			Project: projects["SECRET"],
		},
		{
			Group:   groups["management"],
			Perm:    bitbucketserver.PermProjectRead,
			Project: projects["PRIVATE"],
		},
	}

	userRepoPerms := []*bitbucketserver.UserRepoPermission{
		{
			User: users["ceo"],
			Perm: bitbucketserver.PermRepoWrite,
			Repo: repos["super-secret-repo"],
		},
	}

	return &fixtures{
		users:             users,
		groups:            groups,
		projects:          projects,
		repos:             repos,
		groupProjectPerms: groupProjectPerms,
		userRepoPerms:     userRepoPerms,
	}
}

type codeHost struct {
	*extsvc.CodeHost
}

func (h codeHost) externalRepo(r *bitbucketserver.Repo) api.ExternalRepoSpec {
	return api.ExternalRepoSpec{
		ServiceType: h.ServiceType,
		ServiceID:   h.ServiceID,
		ID:          strconv.Itoa(r.ID),
	}
}

func (h codeHost) externalAccount(userID int32, u *bitbucketserver.User) *extsvc.ExternalAccount {
	bs := marshalJSON(u)
	return &extsvc.ExternalAccount{
		UserID: userID,
		ExternalAccountSpec: extsvc.ExternalAccountSpec{
			ServiceType: h.ServiceType,
			ServiceID:   h.ServiceID,
			AccountID:   strconv.Itoa(u.ID),
		},
		ExternalAccountData: extsvc.ExternalAccountData{
			AccountData: (*json.RawMessage)(&bs),
		},
	}
}

func newProvider(t *testing.T, name string) (*Provider, func()) {
	cli, save := bitbucketserver.NewTestClient(t, name, *update)

	signingKey := os.Getenv("BITBUCKET_SERVER_SIGNING_KEY")
	if signingKey == "" {
		// Bogus key
		signingKey = `LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlCUEFJQkFBSkJBUEpIaWprdG1UMUlLYUd0YTVFZXAzQVo5Q2VPZUw4alBESUZUN3dRZ0tabXQzRUZxRGhCCk93bitRVUhKdUs5Zm92UkROSmVWTDJvWTVCT0l6NHJ3L0cwQ0F3RUFBUUpCQU1BK0o5Mks0d2NQVllsbWMrM28KcHU5NmlKTkNwMmp5Nm5hK1pEQlQzK0VvSUo1VFJGdnN3R2kvTHUzZThYUWwxTDNTM21ub0xPSlZNcTF0bUxOMgpIY0VDSVFEK3daeS83RlYxUEFtdmlXeWlYVklETzJnNWJOaUJlbmdKQ3hFa3Nia1VtUUloQVBOMlZaczN6UFFwCk1EVG9vTlJXcnl0RW1URERkamdiOFpzTldYL1JPRGIxQWlCZWNKblNVQ05TQllLMXJ5VTFmNURTbitoQU9ZaDkKWDFBMlVnTDE3bWhsS1FJaEFPK2JMNmRDWktpTGZORWxmVnRkTUtxQnFjNlBIK01heFU2VzlkVlFvR1dkQWlFQQptdGZ5cE9zYTFiS2hFTDg0blovaXZFYkJyaVJHalAya3lERHYzUlg0V0JrPQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo=`
	}

	consumerKey := os.Getenv("BITBUCKET_SERVER_CONSUMER_KEY")
	if consumerKey == "" {
		consumerKey = "sourcegraph"
	}

	if err := cli.SetOAuth(consumerKey, signingKey); err != nil {
		t.Fatal(err)
	}

	codeHost := extsvc.CodeHost{
		ServiceType: bitbucketserver.ServiceType,
		ServiceID:   os.Getenv("BITBUCKET_SERVER_URL"),
	}

	if codeHost.ServiceID == "" {
		codeHost.ServiceID = "http://127.0.0.1:7990"
	}

	return &Provider{
		client:   cli,
		codeHost: &codeHost,
		pageSize: 1, // Exercise pagination
	}, save
}
