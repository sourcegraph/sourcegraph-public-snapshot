package authz

import (
	"context"
	"database/sql"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/cockroachdb/errors"
	mockrequire "github.com/derision-test/go-mockgen/testutil/require"
	"github.com/google/go-cmp/cmp"

	eauthz "github.com/sourcegraph/sourcegraph/enterprise/internal/authz"
	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestPermsSyncer_ScheduleUsers(t *testing.T) {
	authz.SetProviders(true, []authz.Provider{&mockProvider{}})
	defer authz.SetProviders(true, nil)

	s := NewPermsSyncer(nil, nil, nil, nil, nil)
	s.ScheduleUsers(context.Background(), authz.FetchPermsOptions{}, 1)

	expHeap := []*syncRequest{
		{requestMeta: &requestMeta{
			Priority: priorityHigh,
			Type:     requestTypeUser,
			ID:       1,
		}, acquired: false, index: 0},
	}
	if diff := cmp.Diff(expHeap, s.queue.heap, cmpOpts); diff != "" {
		t.Fatalf("heap: %v", diff)
	}
}

func TestPermsSyncer_ScheduleRepos(t *testing.T) {
	authz.SetProviders(true, []authz.Provider{&mockProvider{}})
	defer authz.SetProviders(true, nil)

	s := NewPermsSyncer(nil, nil, nil, nil, nil)
	s.ScheduleRepos(context.Background(), 1)

	expHeap := []*syncRequest{
		{requestMeta: &requestMeta{
			Priority: priorityHigh,
			Type:     requestTypeRepo,
			ID:       1,
		}, acquired: false, index: 0},
	}
	if diff := cmp.Diff(expHeap, s.queue.heap, cmpOpts); diff != "" {
		t.Fatalf("heap: %v", diff)
	}
}

type mockProvider struct {
	id          int64
	serviceType string
	serviceID   string

	fetchUserPerms        func(context.Context, *extsvc.Account) (*authz.ExternalUserPermissions, error)
	fetchUserPermsByToken func(ctx context.Context, token string) (*authz.ExternalUserPermissions, error)
	fetchRepoPerms        func(ctx context.Context, repo *extsvc.Repository, opts authz.FetchPermsOptions) ([]extsvc.AccountID, error)
}

func (*mockProvider) FetchAccount(context.Context, *types.User, []*extsvc.Account, []string) (*extsvc.Account, error) {
	return nil, nil
}

func (p *mockProvider) ServiceType() string { return p.serviceType }
func (p *mockProvider) ServiceID() string   { return p.serviceID }
func (p *mockProvider) URN() string         { return extsvc.URN(p.serviceType, p.id) }
func (*mockProvider) Validate() []string    { return nil }

func (p *mockProvider) FetchUserPerms(ctx context.Context, acct *extsvc.Account, opts authz.FetchPermsOptions) (*authz.ExternalUserPermissions, error) {
	return p.fetchUserPerms(ctx, acct)
}

func (p *mockProvider) FetchUserPermsByToken(ctx context.Context, token string, opts authz.FetchPermsOptions) (*authz.ExternalUserPermissions, error) {
	return p.fetchUserPermsByToken(ctx, token)
}

func (p *mockProvider) FetchRepoPerms(ctx context.Context, repo *extsvc.Repository, opts authz.FetchPermsOptions) ([]extsvc.AccountID, error) {
	return p.fetchRepoPerms(ctx, repo, opts)
}

func TestPermsSyncer_syncUserPerms(t *testing.T) {
	p := &mockProvider{
		id:          1,
		serviceType: extsvc.TypeGitLab,
		serviceID:   "https://gitlab.com/",
	}
	authz.SetProviders(false, []authz.Provider{p})
	defer authz.SetProviders(true, nil)

	extAccount := extsvc.Account{
		AccountSpec: extsvc.AccountSpec{
			ServiceType: p.ServiceType(),
			ServiceID:   p.ServiceID(),
		},
	}
	extService := &types.ExternalService{
		ID:              1,
		Kind:            extsvc.KindGitLab,
		DisplayName:     "GITLAB1",
		Config:          `{"token": "limited", "authorization": {}}`,
		NamespaceUserID: 1,
	}

	users := database.NewMockUserStore()
	users.GetByIDFunc.SetDefaultHook(func(ctx context.Context, id int32) (*types.User, error) {
		return &types.User{ID: id}, nil
	})

	mockRepos := database.NewMockRepoStore()
	mockRepos.ListMinimalReposFunc.SetDefaultHook(func(ctx context.Context, opt database.ReposListOptions) ([]types.MinimalRepo, error) {
		if !opt.OnlyPrivate {
			return nil, errors.New("OnlyPrivate want true but got false")
		}

		names := make([]types.MinimalRepo, 0, len(opt.ExternalRepos))
		for _, r := range opt.ExternalRepos {
			id, _ := strconv.Atoi(r.ID)
			names = append(names, types.MinimalRepo{ID: api.RepoID(id)})
		}
		return names, nil
	})

	externalServices := database.NewMockExternalServiceStore()
	externalServices.ListFunc.SetDefaultReturn([]*types.ExternalService{extService}, nil)

	userEmails := database.NewMockUserEmailsStore()
	externalAccounts := database.NewMockUserExternalAccountsStore()

	db := database.NewMockDB()
	db.UsersFunc.SetDefaultReturn(users)
	db.ReposFunc.SetDefaultReturn(mockRepos)
	db.ExternalServicesFunc.SetDefaultReturn(externalServices)
	db.UserEmailsFunc.SetDefaultReturn(userEmails)
	db.UserExternalAccountsFunc.SetDefaultReturn(externalAccounts)

	edb.Mocks.Perms.ListExternalAccounts = func(context.Context, int32) ([]*extsvc.Account, error) {
		return []*extsvc.Account{&extAccount}, nil
	}
	edb.Mocks.Perms.SetUserPermissions = func(_ context.Context, p *authz.UserPermissions) error {
		wantIDs := []uint32{1, 2, 3, 4, 5}
		if diff := cmp.Diff(wantIDs, p.IDs.ToArray()); diff != "" {
			return errors.Errorf("IDs mismatch (-want +got):\n%s", diff)
		}
		return nil
	}
	edb.Mocks.Perms.UserIsMemberOfOrgHasCodeHostConnection = func(context.Context, int32) (bool, error) {
		return true, nil
	}
	repos.Mocks.ListExternalServiceUserIDsByRepoID = func(ctx context.Context, repoID api.RepoID) ([]int32, error) {
		return []int32{1}, nil
	}
	repos.Mocks.ListExternalServiceRepoIDsByUserID = func(ctx context.Context, userID int32) ([]api.RepoID, error) {
		return []api.RepoID{2, 3, 4}, nil
	}
	eauthz.MockProviderFromExternalService = func(siteConfig schema.SiteConfiguration, svc *types.ExternalService) (authz.Provider, error) {
		return p, nil
	}
	defer func() {
		database.Mocks = database.MockStores{}
		edb.Mocks.Perms = edb.MockPerms{}
		eauthz.MockProviderFromExternalService = nil
	}()

	reposStore := repos.NewStore(db, sql.TxOptions{})
	reposStore.RepoStore = mockRepos
	permsStore := edb.Perms(db, timeutil.Now)
	s := NewPermsSyncer(db, reposStore, permsStore, timeutil.Now, nil)

	p.fetchUserPerms = func(context.Context, *extsvc.Account) (*authz.ExternalUserPermissions, error) {
		return &authz.ExternalUserPermissions{
			Exacts: []extsvc.RepoID{"1"},
		}, nil
	}
	p.fetchUserPermsByToken = func(ctx context.Context, s string) (*authz.ExternalUserPermissions, error) {
		return &authz.ExternalUserPermissions{
			Exacts: []extsvc.RepoID{"5"},
		}, nil
	}

	err := s.syncUserPerms(context.Background(), 1, true, authz.FetchPermsOptions{})
	if err != nil {
		t.Fatal(err)
	}
}

func TestPermsSyncer_syncUserPerms_noPerms(t *testing.T) {
	p := &mockProvider{
		id:          1,
		serviceType: extsvc.TypeGitLab,
		serviceID:   "https://gitlab.com/",
	}
	authz.SetProviders(false, []authz.Provider{p})
	defer authz.SetProviders(true, nil)

	extAccount := extsvc.Account{
		AccountSpec: extsvc.AccountSpec{
			ServiceType: p.ServiceType(),
			ServiceID:   p.ServiceID(),
		},
	}

	users := database.NewMockUserStore()
	users.GetByIDFunc.SetDefaultHook(func(ctx context.Context, id int32) (*types.User, error) {
		return &types.User{ID: id}, nil
	})

	mockRepos := database.NewMockRepoStore()
	mockRepos.ListMinimalReposFunc.SetDefaultHook(func(ctx context.Context, opt database.ReposListOptions) ([]types.MinimalRepo, error) {
		if !opt.OnlyPrivate {
			return nil, errors.New("OnlyPrivate want true but got false")
		}
		return []types.MinimalRepo{{ID: 1}}, nil
	})

	externalServices := database.NewMockExternalServiceStore()
	externalServices.ListFunc.SetDefaultHook(func(ctx context.Context, opt database.ExternalServicesListOptions) ([]*types.ExternalService, error) {
		return []*types.ExternalService{
			{
				ID:              1,
				Kind:            extsvc.KindGitHub,
				DisplayName:     "GITHUB #1",
				Config:          `{"token": "mytoken"}`,
				NamespaceUserID: opt.NamespaceUserID,
			},
		}, nil
	})

	userEmails := database.NewMockUserEmailsStore()
	externalAccounts := database.NewMockUserExternalAccountsStore()

	db := database.NewMockDB()
	db.UsersFunc.SetDefaultReturn(users)
	db.ReposFunc.SetDefaultReturn(mockRepos)
	db.ExternalServicesFunc.SetDefaultReturn(externalServices)
	db.UserEmailsFunc.SetDefaultReturn(userEmails)
	db.UserExternalAccountsFunc.SetDefaultReturn(externalAccounts)

	edb.Mocks.Perms.ListExternalAccounts = func(context.Context, int32) ([]*extsvc.Account, error) {
		return []*extsvc.Account{&extAccount}, nil
	}
	edb.Mocks.Perms.SetUserPermissions = func(_ context.Context, p *authz.UserPermissions) error {
		if p.UserID != 1 {
			return errors.Errorf("UserID: want 1 but got %d", p.UserID)
		}

		wantIDs := []uint32{1}
		if diff := cmp.Diff(wantIDs, p.IDs.ToArray()); diff != "" {
			return errors.Errorf("IDs mismatch (-want +got):\n%s", diff)
		}
		return nil
	}
	edb.Mocks.Perms.UserIsMemberOfOrgHasCodeHostConnection = func(context.Context, int32) (bool, error) {
		return true, nil
	}
	repos.Mocks.ListExternalServiceRepoIDsByUserID = func(ctx context.Context, userID int32) ([]api.RepoID, error) {
		return []api.RepoID{}, nil
	}
	defer func() {
		database.Mocks = database.MockStores{}
		edb.Mocks.Perms = edb.MockPerms{}
	}()

	reposStore := repos.NewStore(db, sql.TxOptions{})
	reposStore.RepoStore = mockRepos
	permsStore := edb.Perms(db, timeutil.Now)
	s := NewPermsSyncer(db, reposStore, permsStore, timeutil.Now, nil)

	tests := []struct {
		name     string
		noPerms  bool
		fetchErr error
	}{
		{
			name:     "sync for the first time and encounter an error",
			noPerms:  true,
			fetchErr: errors.New("random error"),
		},
		{
			name:    "sync for the second time and succeed",
			noPerms: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p.fetchUserPerms = func(context.Context, *extsvc.Account) (*authz.ExternalUserPermissions, error) {
				return &authz.ExternalUserPermissions{
					Exacts: []extsvc.RepoID{"1"},
				}, test.fetchErr
			}
			p.fetchUserPermsByToken = func(ctx context.Context, token string) (*authz.ExternalUserPermissions, error) {
				return &authz.ExternalUserPermissions{
					Exacts: []extsvc.RepoID{"2"},
				}, nil
			}

			err := s.syncUserPerms(context.Background(), 1, test.noPerms, authz.FetchPermsOptions{})
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestPermsSyncer_syncUserPerms_tokenExpire(t *testing.T) {
	p := &mockProvider{
		serviceType: extsvc.TypeGitHub,
		serviceID:   "https://github.com/",
	}
	authz.SetProviders(false, []authz.Provider{p})
	defer authz.SetProviders(true, nil)

	extAccount := extsvc.Account{
		AccountSpec: extsvc.AccountSpec{
			ServiceType: p.ServiceType(),
			ServiceID:   p.ServiceID(),
		},
	}

	users := database.NewMockUserStore()
	users.GetByIDFunc.SetDefaultHook(func(ctx context.Context, id int32) (*types.User, error) {
		return &types.User{ID: id}, nil
	})

	mockRepos := database.NewMockRepoStore()
	mockRepos.ListMinimalReposFunc.SetDefaultHook(func(ctx context.Context, opt database.ReposListOptions) ([]types.MinimalRepo, error) {
		if !opt.OnlyPrivate {
			return nil, errors.New("OnlyPrivate want true but got false")
		}
		return []types.MinimalRepo{{ID: 1}}, nil
	})

	externalServices := database.NewMockExternalServiceStore()
	userEmails := database.NewMockUserEmailsStore()
	externalAccounts := database.NewMockUserExternalAccountsStore()

	db := database.NewMockDB()
	db.UsersFunc.SetDefaultReturn(users)
	db.ReposFunc.SetDefaultReturn(mockRepos)
	db.ExternalServicesFunc.SetDefaultReturn(externalServices)
	db.UserEmailsFunc.SetDefaultReturn(userEmails)
	db.UserExternalAccountsFunc.SetDefaultReturn(externalAccounts)

	edb.Mocks.Perms.ListExternalAccounts = func(context.Context, int32) ([]*extsvc.Account, error) {
		return []*extsvc.Account{&extAccount}, nil
	}
	edb.Mocks.Perms.SetUserPermissions = func(_ context.Context, p *authz.UserPermissions) error {
		return nil
	}
	edb.Mocks.Perms.UserIsMemberOfOrgHasCodeHostConnection = func(context.Context, int32) (bool, error) {
		return true, nil
	}
	repos.Mocks.ListExternalServiceRepoIDsByUserID = func(ctx context.Context, userID int32) ([]api.RepoID, error) {
		return []api.RepoID{}, nil
	}
	defer func() {
		database.Mocks = database.MockStores{}
		edb.Mocks.Perms = edb.MockPerms{}
	}()

	reposStore := repos.NewStore(db, sql.TxOptions{})
	reposStore.RepoStore = mockRepos
	permsStore := edb.Perms(db, timeutil.Now)
	s := NewPermsSyncer(db, reposStore, permsStore, timeutil.Now, nil)

	t.Run("invalid token", func(t *testing.T) {
		p.fetchUserPerms = func(ctx context.Context, account *extsvc.Account) (*authz.ExternalUserPermissions, error) {
			return nil, &github.APIError{Code: http.StatusUnauthorized}
		}

		err := s.syncUserPerms(context.Background(), 1, false, authz.FetchPermsOptions{})
		if err != nil {
			t.Fatal(err)
		}

		mockrequire.Called(t, externalAccounts.TouchExpiredFunc)
	})

	t.Run("forbidden", func(t *testing.T) {
		p.fetchUserPerms = func(ctx context.Context, account *extsvc.Account) (*authz.ExternalUserPermissions, error) {
			return nil, gitlab.NewHTTPError(http.StatusForbidden, nil)
		}

		err := s.syncUserPerms(context.Background(), 1, false, authz.FetchPermsOptions{})
		if err != nil {
			t.Fatal(err)
		}

		mockrequire.Called(t, externalAccounts.TouchExpiredFunc)
	})

	t.Run("account suspension", func(t *testing.T) {
		p.fetchUserPerms = func(ctx context.Context, account *extsvc.Account) (*authz.ExternalUserPermissions, error) {
			return nil, &github.APIError{
				URL:     "https://api.github.com/user/repos",
				Code:    http.StatusForbidden,
				Message: "Sorry. Your account was suspended",
			}
		}

		err := s.syncUserPerms(context.Background(), 1, false, authz.FetchPermsOptions{})
		if err != nil {
			t.Fatal(err)
		}

		mockrequire.Called(t, externalAccounts.TouchExpiredFunc)
	})
}

func TestPermsSyncer_syncUserPerms_prefixSpecs(t *testing.T) {
	p := &mockProvider{
		serviceType: extsvc.TypePerforce,
		serviceID:   "ssl:111.222.333.444:1666",
	}
	authz.SetProviders(false, []authz.Provider{p})
	defer authz.SetProviders(true, nil)

	extAccount := extsvc.Account{
		AccountSpec: extsvc.AccountSpec{
			ServiceType: p.ServiceType(),
			ServiceID:   p.ServiceID(),
		},
	}

	users := database.NewMockUserStore()
	users.GetByIDFunc.SetDefaultHook(func(ctx context.Context, id int32) (*types.User, error) {
		return &types.User{ID: id}, nil
	})

	mockRepos := database.NewMockRepoStore()
	mockRepos.ListMinimalReposFunc.SetDefaultHook(func(ctx context.Context, opt database.ReposListOptions) ([]types.MinimalRepo, error) {
		if !opt.OnlyPrivate {
			return nil, errors.New("OnlyPrivate want true but got false")
		} else if len(opt.ExternalRepoIncludeContains) == 0 {
			return nil, errors.New("ExternalRepoIncludeContains want non-zero but got zero")
		} else if len(opt.ExternalRepoExcludeContains) == 0 {
			return nil, errors.New("ExternalRepoExcludeContains want non-zero but got zero")
		}
		return []types.MinimalRepo{{ID: 1}}, nil
	})

	externalServices := database.NewMockExternalServiceStore()
	userEmails := database.NewMockUserEmailsStore()
	externalAccounts := database.NewMockUserExternalAccountsStore()

	db := database.NewMockDB()
	db.UsersFunc.SetDefaultReturn(users)
	db.ReposFunc.SetDefaultReturn(mockRepos)
	db.ExternalServicesFunc.SetDefaultReturn(externalServices)
	db.UserEmailsFunc.SetDefaultReturn(userEmails)
	db.UserExternalAccountsFunc.SetDefaultReturn(externalAccounts)

	edb.Mocks.Perms.ListExternalAccounts = func(context.Context, int32) ([]*extsvc.Account, error) {
		return []*extsvc.Account{&extAccount}, nil
	}
	edb.Mocks.Perms.SetUserPermissions = func(_ context.Context, p *authz.UserPermissions) error {
		return nil
	}
	edb.Mocks.Perms.UserIsMemberOfOrgHasCodeHostConnection = func(context.Context, int32) (bool, error) {
		return true, nil
	}
	repos.Mocks.ListExternalServiceRepoIDsByUserID = func(ctx context.Context, userID int32) ([]api.RepoID, error) {
		return []api.RepoID{}, nil
	}
	defer func() {
		database.Mocks = database.MockStores{}
		edb.Mocks.Perms = edb.MockPerms{}
	}()

	reposStore := repos.NewStore(db, sql.TxOptions{})
	reposStore.RepoStore = mockRepos
	permsStore := edb.Perms(db, timeutil.Now)
	s := NewPermsSyncer(db, reposStore, permsStore, timeutil.Now, nil)

	p.fetchUserPerms = func(context.Context, *extsvc.Account) (*authz.ExternalUserPermissions, error) {
		return &authz.ExternalUserPermissions{
			IncludeContains: []extsvc.RepoID{"//Engineering/"},
			ExcludeContains: []extsvc.RepoID{"//Engineering/Security/"},
		}, nil
	}

	err := s.syncUserPerms(context.Background(), 1, false, authz.FetchPermsOptions{})
	if err != nil {
		t.Fatal(err)
	}
}

func TestPermsSyncer_syncUserPerms_subRepoPermissions(t *testing.T) {
	p := &mockProvider{
		serviceType: extsvc.TypePerforce,
		serviceID:   "ssl:111.222.333.444:1666",
	}
	authz.SetProviders(false, []authz.Provider{p})
	defer authz.SetProviders(true, nil)

	extAccount := extsvc.Account{
		AccountSpec: extsvc.AccountSpec{
			ServiceType: p.ServiceType(),
			ServiceID:   p.ServiceID(),
		},
	}

	users := database.NewMockUserStore()
	users.GetByIDFunc.SetDefaultHook(func(ctx context.Context, id int32) (*types.User, error) {
		return &types.User{ID: id}, nil
	})

	mockRepos := database.NewMockRepoStore()
	mockRepos.ListMinimalReposFunc.SetDefaultHook(func(ctx context.Context, opt database.ReposListOptions) ([]types.MinimalRepo, error) {
		if !opt.OnlyPrivate {
			return nil, errors.New("OnlyPrivate want true but got false")
		} else if len(opt.ExternalRepoIncludeContains) == 0 {
			return nil, errors.New("ExternalRepoIncludeContains want non-zero but got zero")
		} else if len(opt.ExternalRepoExcludeContains) == 0 {
			return nil, errors.New("ExternalRepoExcludeContains want non-zero but got zero")
		}
		return []types.MinimalRepo{{ID: 1}}, nil
	})

	externalServices := database.NewMockExternalServiceStore()
	userEmails := database.NewMockUserEmailsStore()
	externalAccounts := database.NewMockUserExternalAccountsStore()
	subRepoPerms := database.NewMockSubRepoPermsStore()

	db := database.NewMockDB()
	db.UsersFunc.SetDefaultReturn(users)
	db.ReposFunc.SetDefaultReturn(mockRepos)
	db.ExternalServicesFunc.SetDefaultReturn(externalServices)
	db.UserEmailsFunc.SetDefaultReturn(userEmails)
	db.UserExternalAccountsFunc.SetDefaultReturn(externalAccounts)
	db.SubRepoPermsFunc.SetDefaultReturn(subRepoPerms)

	edb.Mocks.Perms.ListExternalAccounts = func(context.Context, int32) ([]*extsvc.Account, error) {
		return []*extsvc.Account{&extAccount}, nil
	}
	edb.Mocks.Perms.SetUserPermissions = func(_ context.Context, p *authz.UserPermissions) error {
		return nil
	}
	edb.Mocks.Perms.UserIsMemberOfOrgHasCodeHostConnection = func(context.Context, int32) (bool, error) {
		return false, nil
	}
	repos.Mocks.ListExternalServiceRepoIDsByUserID = func(ctx context.Context, userID int32) ([]api.RepoID, error) {
		return []api.RepoID{}, nil
	}
	defer func() {
		database.Mocks = database.MockStores{}
		edb.Mocks.Perms = edb.MockPerms{}
	}()

	reposStore := repos.NewStore(db, sql.TxOptions{})
	reposStore.RepoStore = mockRepos
	permsStore := edb.Perms(db, timeutil.Now)
	s := NewPermsSyncer(db, reposStore, permsStore, timeutil.Now, nil)

	p.fetchUserPerms = func(context.Context, *extsvc.Account) (*authz.ExternalUserPermissions, error) {
		return &authz.ExternalUserPermissions{
			IncludeContains: []extsvc.RepoID{"//Engineering/"},
			ExcludeContains: []extsvc.RepoID{"//Engineering/Security/"},

			SubRepoPermissions: map[extsvc.RepoID]*authz.SubRepoPermissions{
				"abc": {
					PathIncludes: []string{"include1", "include2"},
					PathExcludes: []string{"exclude1", "exclude2"},
				},
				"def": {
					PathIncludes: []string{"include1", "include2"},
					PathExcludes: []string{"exclude1", "exclude2"},
				},
			},
		}, nil
	}

	err := s.syncUserPerms(context.Background(), 1, false, authz.FetchPermsOptions{})
	if err != nil {
		t.Fatal(err)
	}

	mockrequire.CalledN(t, subRepoPerms.UpsertWithSpecFunc, 2)
}

func TestPermsSyncer_syncRepoPerms(t *testing.T) {
	mockRepos := database.NewMockRepoStore()
	db := database.NewMockDB()
	db.ReposFunc.SetDefaultReturn(mockRepos)

	newPermsSyncer := func() *PermsSyncer {
		reposStore := repos.NewStore(db, sql.TxOptions{})
		reposStore.RepoStore = mockRepos
		permsStore := edb.Perms(db, timeutil.Now)
		return NewPermsSyncer(db, reposStore, permsStore, timeutil.Now, nil)
	}

	t.Run("TouchRepoPermissions is called when no authz provider", func(t *testing.T) {
		mockRepos.ListFunc.SetDefaultReturn(
			[]*types.Repo{
				{
					ID:      1,
					Private: true,
					ExternalRepo: api.ExternalRepoSpec{
						ServiceID: "https://gitlab.com/",
					},
					Sources: map[string]*types.SourceInfo{
						extsvc.URN(extsvc.TypeGitLab, 0): {},
					},
				},
			},
			nil,
		)

		calledTouchRepoPermissions := false
		edb.Mocks.Perms.TouchRepoPermissions = func(ctx context.Context, repoID int32) error {
			calledTouchRepoPermissions = true
			return nil
		}
		repos.Mocks.ListExternalServiceUserIDsByRepoID = func(ctx context.Context, repoID api.RepoID) ([]int32, error) {
			return []int32{}, nil
		}
		defer func() {
			edb.Mocks.Perms = edb.MockPerms{}
			repos.Mocks = repos.ReposMocks{}
		}()

		s := newPermsSyncer()

		err := s.syncRepoPerms(context.Background(), 1, false, authz.FetchPermsOptions{})
		if err != nil {
			t.Fatal(err)
		}

		if !calledTouchRepoPermissions {
			t.Fatal("!calledTouchRepoPermissions")
		}
	})

	t.Run("identify authz provider by URN", func(t *testing.T) {
		// Even though both p1 and p2 are pointing to the same code host,
		// but p2 should not be used because it is not responsible for listing
		// test repository.
		p1 := &mockProvider{
			id:          1,
			serviceType: extsvc.TypeGitLab,
			serviceID:   "https://gitlab.com/",
			fetchRepoPerms: func(ctx context.Context, repo *extsvc.Repository, opts authz.FetchPermsOptions) ([]extsvc.AccountID, error) {
				return []extsvc.AccountID{"user"}, nil
			},
		}
		p2 := &mockProvider{
			id:          2,
			serviceType: extsvc.TypeGitLab,
			serviceID:   "https://gitlab.com/",
			fetchRepoPerms: func(ctx context.Context, repo *extsvc.Repository, opts authz.FetchPermsOptions) ([]extsvc.AccountID, error) {
				return nil, errors.New("not supposed to be called")
			},
		}
		authz.SetProviders(false, []authz.Provider{p1, p2})
		defer authz.SetProviders(true, nil)

		mockRepos.ListFunc.SetDefaultReturn(
			[]*types.Repo{
				{
					ID:      1,
					Private: true,
					ExternalRepo: api.ExternalRepoSpec{
						ServiceID: p1.ServiceID(),
					},
					Sources: map[string]*types.SourceInfo{
						p1.URN(): {},
					},
				},
			},
			nil,
		)

		edb.Mocks.Perms.Transact = func(context.Context) (edb.PermsStore, error) {
			return edb.Perms(nil, nil), nil
		}
		edb.Mocks.Perms.GetUserIDsByExternalAccounts = func(context.Context, *extsvc.Accounts) (map[string]int32, error) {
			return map[string]int32{"user": 1}, nil
		}
		edb.Mocks.Perms.SetRepoPermissions = func(_ context.Context, p *authz.RepoPermissions) error {
			if p.RepoID != 1 {
				return errors.Errorf("RepoID: want 1 but got %d", p.RepoID)
			}

			wantUserIDs := []uint32{1}
			if diff := cmp.Diff(wantUserIDs, p.UserIDs.ToArray()); diff != "" {
				return errors.Errorf("UserIDs mismatch (-want +got):\n%s", diff)
			}
			return nil
		}
		edb.Mocks.Perms.SetRepoPendingPermissions = func(ctx context.Context, accounts *extsvc.Accounts, p *authz.RepoPermissions) error {
			return nil
		}
		repos.Mocks.ListExternalServiceUserIDsByRepoID = func(ctx context.Context, repoID api.RepoID) ([]int32, error) {
			return []int32{}, nil
		}
		defer func() {
			edb.Mocks.Perms = edb.MockPerms{}
			repos.Mocks = repos.ReposMocks{}
		}()

		s := newPermsSyncer()

		err := s.syncRepoPerms(context.Background(), 1, false, authz.FetchPermsOptions{})
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("repo sync with external service userid but no providers", func(t *testing.T) {
		mockRepos.ListFunc.SetDefaultReturn(
			[]*types.Repo{
				{
					ID:      1,
					Private: true,
				},
			},
			nil,
		)

		edb.Mocks.Perms.Transact = func(context.Context) (edb.PermsStore, error) {
			return edb.Perms(nil, nil), nil
		}
		edb.Mocks.Perms.GetUserIDsByExternalAccounts = func(context.Context, *extsvc.Accounts) (map[string]int32, error) {
			return map[string]int32{"user": 1}, nil
		}
		edb.Mocks.Perms.SetRepoPermissions = func(_ context.Context, p *authz.RepoPermissions) error {
			if p.RepoID != 1 {
				return errors.Errorf("RepoID: want 1 but got %d", p.RepoID)
			}

			wantUserIDs := []uint32{1}
			if diff := cmp.Diff(wantUserIDs, p.UserIDs.ToArray()); diff != "" {
				return errors.Errorf("UserIDs mismatch (-want +got):\n%s", diff)
			}
			return nil
		}
		edb.Mocks.Perms.SetRepoPendingPermissions = func(ctx context.Context, accounts *extsvc.Accounts, p *authz.RepoPermissions) error {
			return errors.Errorf("SetRepoPendingPermissions should not be invoked in this test case")
		}
		repos.Mocks.ListExternalServiceUserIDsByRepoID = func(ctx context.Context, repoID api.RepoID) ([]int32, error) {
			return []int32{1}, nil
		}
		defer func() {
			edb.Mocks.Perms = edb.MockPerms{}
			repos.Mocks = repos.ReposMocks{}
		}()

		s := newPermsSyncer()

		err := s.syncRepoPerms(context.Background(), 1, false, authz.FetchPermsOptions{})
		if err != nil {
			t.Fatal(err)
		}
	})

	p := &mockProvider{
		serviceType: extsvc.TypeGitLab,
		serviceID:   "https://gitlab.com/",
	}
	authz.SetProviders(false, []authz.Provider{p})
	defer authz.SetProviders(true, nil)

	mockRepos.ListFunc.SetDefaultReturn(
		[]*types.Repo{
			{
				ID:      1,
				Private: true,
				ExternalRepo: api.ExternalRepoSpec{
					ServiceID: p.ServiceID(),
				},
				Sources: map[string]*types.SourceInfo{
					p.URN(): {},
				},
			},
		},
		nil,
	)

	edb.Mocks.Perms.Transact = func(context.Context) (edb.PermsStore, error) {
		return edb.Perms(nil, nil), nil
	}
	edb.Mocks.Perms.GetUserIDsByExternalAccounts = func(context.Context, *extsvc.Accounts) (map[string]int32, error) {
		return map[string]int32{"user": 1}, nil
	}
	edb.Mocks.Perms.SetRepoPermissions = func(_ context.Context, p *authz.RepoPermissions) error {
		if p.RepoID != 1 {
			return errors.Errorf("RepoID: want 1 but got %d", p.RepoID)
		}

		wantUserIDs := []uint32{1}
		if diff := cmp.Diff(wantUserIDs, p.UserIDs.ToArray()); diff != "" {
			return errors.Errorf("UserIDs mismatch (-want +got):\n%s", diff)
		}
		return nil
	}
	edb.Mocks.Perms.SetRepoPendingPermissions = func(_ context.Context, accounts *extsvc.Accounts, _ *authz.RepoPermissions) error {
		wantAccounts := &extsvc.Accounts{
			ServiceType: p.ServiceType(),
			ServiceID:   p.ServiceID(),
			AccountIDs:  []string{"pending_user"},
		}
		if diff := cmp.Diff(wantAccounts, accounts); diff != "" {
			return errors.Errorf("accounts mismatch (-want +got):\n%s", diff)
		}
		return nil
	}
	repos.Mocks.ListExternalServiceUserIDsByRepoID = func(ctx context.Context, repoID api.RepoID) ([]int32, error) {
		return []int32{}, nil
	}
	defer func() {
		edb.Mocks.Perms = edb.MockPerms{}
		repos.Mocks = repos.ReposMocks{}
	}()

	s := newPermsSyncer()

	tests := []struct {
		name     string
		noPerms  bool
		fetchErr error
	}{
		{
			name:     "sync for the first time and encounter an error",
			noPerms:  true,
			fetchErr: errors.New("random error"),
		},
		{
			name:    "sync for the second time and succeed",
			noPerms: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p.fetchRepoPerms = func(context.Context, *extsvc.Repository, authz.FetchPermsOptions) ([]extsvc.AccountID, error) {
				return []extsvc.AccountID{"user", "pending_user"}, test.fetchErr
			}

			err := s.syncRepoPerms(context.Background(), 1, test.noPerms, authz.FetchPermsOptions{})
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestPermsSyncer_waitForRateLimit(t *testing.T) {
	ctx := context.Background()
	t.Run("no rate limit registry", func(t *testing.T) {
		s := NewPermsSyncer(nil, nil, nil, nil, nil)

		ctx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		err := s.waitForRateLimit(ctx, "https://github.com/", 100000)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("enough quota available", func(t *testing.T) {
		rateLimiterRegistry := ratelimit.NewRegistry()
		s := NewPermsSyncer(nil, nil, nil, nil, rateLimiterRegistry)

		ctx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		err := s.waitForRateLimit(ctx, "https://github.com/", 1)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("not enough quota available", func(t *testing.T) {
		rateLimiterRegistry := ratelimit.NewRegistry()
		l := rateLimiterRegistry.Get("https://github.com/")
		l.SetLimit(1)
		l.SetBurst(1)
		s := NewPermsSyncer(nil, nil, nil, nil, rateLimiterRegistry)

		ctx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		err := s.waitForRateLimit(ctx, "https://github.com/", 10)
		if err == nil {
			t.Fatalf("err: want %v but got nil", context.Canceled)
		}
	})
}

func TestPermsSyncer_syncPerms(t *testing.T) {
	request := &syncRequest{
		requestMeta: &requestMeta{
			Type: 3,
			ID:   1,
		},
		acquired: true,
	}

	// Request should be removed from the queue even if error occurred.
	s := NewPermsSyncer(nil, nil, nil, nil, nil)
	s.queue.Push(request)

	expErr := "unexpected request type: 3"
	err := s.syncPerms(context.Background(), request)
	if err == nil || err.Error() != expErr {
		t.Fatalf("err: want %q but got %v", expErr, err)
	}

	if s.queue.Len() != 0 {
		t.Fatalf("queue length: want 0 but got %d", s.queue.Len())
	}
}
