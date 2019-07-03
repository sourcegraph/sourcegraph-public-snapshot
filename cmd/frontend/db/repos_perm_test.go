package db

import (
	"context"
	"fmt"
	"net/url"
	"testing"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbtesting"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc"
)

func Benchmark_authzFilter(b *testing.B) {
	repos := make([]*types.Repo, 30000)
	for i := range repos {
		id := i + 1
		repos[i] = &types.Repo{
			ID:   api.RepoID(id),
			Name: api.RepoName(fmt.Sprintf("github.com/organization/repository-%d", i)),
			ExternalRepo: api.ExternalRepoSpec{
				ID:          "a-long-and-random-github-external-id",
				ServiceType: "github",
				ServiceID:   "https://github.com/",
			},
		}
	}

	user := &types.User{
		ID:        42,
		Username:  "john.doe",
		SiteAdmin: false,
	}

	Mocks.Users.GetByCurrentAuthUser = func(ctx context.Context) (*types.User, error) {
		return user, nil
	}
	defer func() { Mocks.Users.GetByCurrentAuthUser = nil }()

	providers := []authz.Provider{
		func() authz.Provider {
			baseURL, _ := url.Parse("http://fake.provider")
			codeHost := extsvc.NewCodeHost(baseURL, "fake")
			return &fakeProvider{
				codeHost: codeHost,
				extAcct: &extsvc.ExternalAccount{
					UserID: user.ID,
					ExternalAccountSpec: extsvc.ExternalAccountSpec{
						ServiceType: codeHost.ServiceType,
						ServiceID:   codeHost.ServiceID,
						AccountID:   "42_ext",
					},
					ExternalAccountData: extsvc.ExternalAccountData{AccountData: nil},
				},
			}
		}(),
		func() authz.Provider {
			baseURL, _ := url.Parse("https://github.com")
			codeHost := extsvc.NewCodeHost(baseURL, "github")
			return &fakeProvider{
				codeHost: codeHost,
				extAcct: &extsvc.ExternalAccount{
					UserID: user.ID,
					ExternalAccountSpec: extsvc.ExternalAccountSpec{
						ServiceType: codeHost.ServiceType,
						ServiceID:   codeHost.ServiceID,
						AccountID:   "42_ext",
					},
					ExternalAccountData: extsvc.ExternalAccountData{AccountData: nil},
				},
			}
		}(),
	}

	authz.SetProviders(false, providers)

	Mocks.ExternalAccounts.List = func(opt ExternalAccountsListOptions) (
		accts []*extsvc.ExternalAccount,
		err error,
	) {
		for _, p := range providers {
			acct, _ := p.FetchAccount(context.Background(), user, nil)
			accts = append(accts, acct)
		}
		return accts, nil
	}
	defer func() { Mocks.ExternalAccounts.List = nil }()

	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := authzFilter(ctx, repos, authz.Read)
		if err != nil {
			b.Fatal(err)
		}
	}
}

type fakeProvider struct {
	codeHost *extsvc.CodeHost
	extAcct  *extsvc.ExternalAccount
}

func (f fakeProvider) Repos(ctx context.Context, repos map[authz.Repo]struct{}) (
	mine map[authz.Repo]struct{},
	others map[authz.Repo]struct{},
) {
	return authz.GetCodeHostRepos(f.codeHost, repos)
}

func (f fakeProvider) RepoPerms(
	ctx context.Context,
	userAccount *extsvc.ExternalAccount,
	repos map[authz.Repo]struct{},
) (map[api.RepoName]map[authz.Perm]bool, error) {
	authorized := make(map[api.RepoName]map[authz.Perm]bool, len(repos))
	for repo := range repos {
		authorized[repo.RepoName] = map[authz.Perm]bool{authz.Read: true}
	}
	return authorized, nil
}

func (f fakeProvider) FetchAccount(
	ctx context.Context,
	user *types.User,
	current []*extsvc.ExternalAccount,
) (mine *extsvc.ExternalAccount, err error) {
	return f.extAcct, nil
}

func (f fakeProvider) ServiceType() string           { return f.codeHost.ServiceType }
func (f fakeProvider) ServiceID() string             { return f.codeHost.ServiceID }
func (f fakeProvider) Validate() (problems []string) { return nil }

// 🚨 SECURITY: test necessary to ensure security
func Test_getBySQL_permissionsCheck(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	defer func() { mockAuthzFilter = nil }()

	ctx := dbtesting.TestContext(t)

	allRepos := mustCreate(ctx, t,
		&types.Repo{

			Name: "r0",
			ExternalRepo: api.ExternalRepoSpec{
				ID:          "a0",
				ServiceType: "b0",
				ServiceID:   "c0",
			}},

		&types.Repo{

			Name: "r1",
			ExternalRepo: api.ExternalRepoSpec{
				ID:          "a1",
				ServiceType: "b1",
				ServiceID:   "c1",
			}},
	)
	{
		calledFilter := false
		mockAuthzFilter = func(ctx context.Context, repos []*types.Repo, p authz.Perm) ([]*types.Repo, error) {
			calledFilter = true
			return repos, nil
		}

		gotRepos, err := Repos.getBySQL(ctx, sqlf.Sprintf("true"))
		if err != nil {
			t.Fatal(err)
		}
		if !jsonEqual(t, gotRepos, allRepos) {
			t.Errorf("got %v, want %v", gotRepos, allRepos)
		}
		if !calledFilter {
			t.Error("did not call authzFilter (SECURITY)")
		}
	}
	{
		calledFilter := false
		mockAuthzFilter = func(ctx context.Context, repos []*types.Repo, p authz.Perm) ([]*types.Repo, error) {
			calledFilter = true
			return nil, nil
		}

		gotRepos, err := Repos.getBySQL(ctx, sqlf.Sprintf("true"))
		if err != nil {
			t.Fatal(err)
		}
		if !jsonEqual(t, gotRepos, nil) {
			t.Errorf("got %v, want %v", gotRepos, nil)
		}
		if !calledFilter {
			t.Error("did not call authzFilter (SECURITY)")
		}
	}
	{
		calledFilter := false
		filteredRepos := allRepos[0:1]
		mockAuthzFilter = func(ctx context.Context, repos []*types.Repo, p authz.Perm) ([]*types.Repo, error) {
			calledFilter = true
			return filteredRepos, nil
		}

		gotRepos, err := Repos.getBySQL(ctx, sqlf.Sprintf("true"))
		if err != nil {
			t.Fatal(err)
		}
		if !jsonEqual(t, gotRepos, filteredRepos) {
			t.Errorf("got %v, want %v", gotRepos, filteredRepos)
		}
		if !calledFilter {
			t.Error("did not call authzFilter (SECURITY)")
		}
	}
}
