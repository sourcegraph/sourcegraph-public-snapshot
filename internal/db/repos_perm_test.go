package db

import (
	"context"
	"fmt"
	"net/url"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/schema"
)

func Benchmark_authzFilter(b *testing.B) {
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
				extAcct: &extsvc.Account{
					UserID: user.ID,
					AccountSpec: extsvc.AccountSpec{
						ServiceType: codeHost.ServiceType,
						ServiceID:   codeHost.ServiceID,
						AccountID:   "42_ext",
					},
					AccountData: extsvc.AccountData{Data: nil},
				},
			}
		}(),
		func() authz.Provider {
			baseURL, _ := url.Parse("https://github.com")
			codeHost := extsvc.NewCodeHost(baseURL, extsvc.TypeGitHub)
			return &fakeProvider{
				codeHost: codeHost,
				extAcct: &extsvc.Account{
					UserID: user.ID,
					AccountSpec: extsvc.AccountSpec{
						ServiceType: codeHost.ServiceType,
						ServiceID:   codeHost.ServiceID,
						AccountID:   "42_ext",
					},
					AccountData: extsvc.AccountData{Data: nil},
				},
			}
		}(),
	}

	{
		authzAllowByDefault, providers := authz.GetProviders()
		defer authz.SetProviders(authzAllowByDefault, providers)
	}

	authz.SetProviders(false, providers)

	serviceIDs := make([]string, 0, len(providers))
	for _, p := range providers {
		serviceIDs = append(serviceIDs, p.ServiceID())
	}

	rs := make([]types.Repo, 30000)
	for i := range rs {
		id := i + 1
		serviceID := serviceIDs[i%len(serviceIDs)]
		rs[i] = types.Repo{
			ID:           api.RepoID(id),
			ExternalRepo: api.ExternalRepoSpec{ServiceID: serviceID},
		}
	}

	repos := make([][]*types.Repo, b.N)
	for i := range repos {
		repos[i] = make([]*types.Repo, len(rs))
		for j := range repos[i] {
			repos[i][j] = &rs[j]
		}
	}

	ctx := context.Background()

	Mocks.ExternalAccounts.List = func(opt ExternalAccountsListOptions) (
		accts []*extsvc.Account,
		err error,
	) {
		for _, p := range providers {
			acct, _ := p.FetchAccount(ctx, user, nil)
			accts = append(accts, acct)
		}
		return accts, nil
	}
	defer func() { Mocks.ExternalAccounts.List = nil }()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := authzFilter(ctx, repos[i], authz.Read)
		if err != nil {
			b.Fatal(err)
		}
	}
}

type fakeProvider struct {
	codeHost *extsvc.CodeHost
	extAcct  *extsvc.Account
}

func (p *fakeProvider) FetchAccount(
	ctx context.Context,
	user *types.User,
	current []*extsvc.Account,
) (mine *extsvc.Account, err error) {
	return p.extAcct, nil
}

func (p *fakeProvider) ServiceType() string           { return p.codeHost.ServiceType }
func (p *fakeProvider) ServiceID() string             { return p.codeHost.ServiceID }
func (p *fakeProvider) URN() string                   { return extsvc.URN(p.codeHost.ServiceType, 0) }
func (p *fakeProvider) Validate() (problems []string) { return nil }

func (p *fakeProvider) FetchUserPerms(context.Context, *extsvc.Account) ([]extsvc.RepoID, error) {
	return nil, nil
}

func (p *fakeProvider) FetchRepoPerms(context.Context, *extsvc.Repository) ([]extsvc.AccountID, error) {
	return nil, nil
}

// 🚨 SECURITY: Tests are necessary to ensure security.
func TestAuthzQueryConds(t *testing.T) {
	cmpOpts := cmp.AllowUnexported(sqlf.Query{})

	t.Run("Conflict with permissions user mapping", func(t *testing.T) {
		before := globals.PermissionsUserMapping()
		globals.SetPermissionsUserMapping(&schema.PermissionsUserMapping{Enabled: true})
		defer globals.SetPermissionsUserMapping(before)

		authz.SetProviders(false, []authz.Provider{&fakeProvider{}})
		defer authz.SetProviders(true, nil)

		_, err := authzQueryConds(context.Background())
		gotErr := fmt.Sprintf("%v", err)
		if diff := cmp.Diff(errPermissionsUserMappingConflict.Error(), gotErr); diff != "" {
			t.Fatalf("Mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("When permissions user mapping is enabled", func(t *testing.T) {
		before := globals.PermissionsUserMapping()
		globals.SetPermissionsUserMapping(&schema.PermissionsUserMapping{Enabled: true})
		defer globals.SetPermissionsUserMapping(before)

		got, err := authzQueryConds(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		want := sqlf.Sprintf(authzQueryCondsFmtstr, false, true, int32(0), authz.Read.String())
		if diff := cmp.Diff(want, got, cmpOpts); diff != "" {
			t.Fatalf("Mismatch (-want +got):\n%s", diff)
		}
	})

	defer func() {
		Mocks.Users.GetByCurrentAuthUser = nil
	}()
	tests := []struct {
		name                string
		setup               func() context.Context
		authzAllowByDefault bool
		wantQuery           *sqlf.Query
	}{
		{
			name: "internal actor bypass checks",
			setup: func() context.Context {
				return actor.WithInternalActor(context.Background())
			},
			wantQuery: sqlf.Sprintf(authzQueryCondsFmtstr, true, false, int32(0), authz.Read.String()),
		},

		{
			name: "no authz provider and not allow by default",
			setup: func() context.Context {
				return context.Background()
			},
			wantQuery: sqlf.Sprintf(authzQueryCondsFmtstr, false, false, int32(0), authz.Read.String()),
		},
		{
			name: "no authz provider but allow by default",
			setup: func() context.Context {
				return context.Background()
			},
			authzAllowByDefault: true,
			wantQuery:           sqlf.Sprintf(authzQueryCondsFmtstr, true, false, int32(0), authz.Read.String()),
		},

		{
			name: "authenticated user is a site admin",
			setup: func() context.Context {
				Mocks.Users.GetByCurrentAuthUser = func(ctx context.Context) (*types.User, error) {
					return &types.User{ID: 1, SiteAdmin: true}, nil
				}
				return actor.WithActor(context.Background(), &actor.Actor{UID: 1})
			},
			wantQuery: sqlf.Sprintf(authzQueryCondsFmtstr, true, false, int32(1), authz.Read.String()),
		},
		{
			name: "authenticated user is not a site admin",
			setup: func() context.Context {
				Mocks.Users.GetByCurrentAuthUser = func(ctx context.Context) (*types.User, error) {
					return &types.User{ID: 1}, nil
				}
				return actor.WithActor(context.Background(), &actor.Actor{UID: 1})
			},
			wantQuery: sqlf.Sprintf(authzQueryCondsFmtstr, false, false, int32(1), authz.Read.String()),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			authz.SetProviders(test.authzAllowByDefault, nil)
			defer authz.SetProviders(true, nil)

			q, err := authzQueryConds(test.setup())
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(test.wantQuery, q, cmpOpts); diff != "" {
				t.Fatalf("Mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// 🚨 SECURITY: test necessary to ensure security
func Test_getBySQL_permissionsCheck(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	defer func() { MockAuthzFilter = nil }()

	dbtesting.SetupGlobalTestDB(t)
	ctx := context.Background()
	ctx = actor.WithActor(ctx, &actor.Actor{UID: 1, Internal: true})

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
		MockAuthzFilter = func(ctx context.Context, repos []*types.Repo, p authz.Perms) ([]*types.Repo, error) {
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
		MockAuthzFilter = func(ctx context.Context, repos []*types.Repo, p authz.Perms) ([]*types.Repo, error) {
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
		MockAuthzFilter = func(ctx context.Context, repos []*types.Repo, p authz.Perms) ([]*types.Repo, error) {
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
