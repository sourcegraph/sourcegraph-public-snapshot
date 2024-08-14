package authz

import (
	"context"
	"net/http"
	"strconv"
	"testing"
	"time"

	mockrequire "github.com/derision-test/go-mockgen/v2/testutil/require"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type mockProvider struct {
	id          int64
	serviceType string
	serviceID   string

	fetchUserPerms        func(context.Context, *extsvc.Account) (*authz.ExternalUserPermissions, error)
	fetchUserPermsByToken func(ctx context.Context, token string) (*authz.ExternalUserPermissions, error)
	fetchRepoPerms        func(ctx context.Context, repo *extsvc.Repository, opts authz.FetchPermsOptions) ([]extsvc.AccountID, error)
	fetchAccount          func(ctx context.Context, user *types.User) (*extsvc.Account, error)
}

func (p *mockProvider) FetchAccount(ctx context.Context, user *types.User) (*extsvc.Account, error) {
	if p.fetchAccount != nil {
		return p.fetchAccount(ctx, user)
	}
	return nil, nil
}

func (p *mockProvider) ServiceType() string { return p.serviceType }
func (p *mockProvider) ServiceID() string   { return p.serviceID }
func (p *mockProvider) URN() string         { return extsvc.URN(p.serviceType, p.id) }

func (*mockProvider) ValidateConnection(context.Context) error { return nil }

func (p *mockProvider) FetchUserPerms(ctx context.Context, acct *extsvc.Account, _ authz.FetchPermsOptions) (*authz.ExternalUserPermissions, error) {
	return p.fetchUserPerms(ctx, acct)
}

func (p *mockProvider) FetchUserPermsByToken(ctx context.Context, token string, _ authz.FetchPermsOptions) (*authz.ExternalUserPermissions, error) {
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

	extAccount := extsvc.Account{
		AccountSpec: extsvc.AccountSpec{
			ServiceType: p.ServiceType(),
			ServiceID:   p.ServiceID(),
		},
	}

	users := dbmocks.NewMockUserStore()
	users.GetByIDFunc.SetDefaultHook(func(ctx context.Context, id int32) (*types.User, error) {
		return &types.User{ID: id}, nil
	})

	mockRepos := dbmocks.NewMockRepoStore()
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

	userEmails := dbmocks.NewMockUserEmailsStore()

	externalAccounts := dbmocks.NewMockUserExternalAccountsStore()
	externalAccounts.ListFunc.SetDefaultHook(func(_ context.Context, opts database.ExternalAccountsListOptions) ([]*extsvc.Account, error) {
		if opts.OnlyExpired {
			return []*extsvc.Account{}, nil
		}
		return []*extsvc.Account{&extAccount}, nil
	})

	featureFlags := dbmocks.NewMockFeatureFlagStore()

	syncJobs := dbmocks.NewMockPermissionSyncJobStore()
	syncJobs.GetLatestFinishedSyncJobFunc.SetDefaultReturn(nil, nil)

	db := dbmocks.NewMockDB()
	db.UsersFunc.SetDefaultReturn(users)
	db.ReposFunc.SetDefaultReturn(mockRepos)
	db.UserEmailsFunc.SetDefaultReturn(userEmails)
	db.UserExternalAccountsFunc.SetDefaultReturn(externalAccounts)
	db.FeatureFlagsFunc.SetDefaultReturn(featureFlags)
	db.PermissionSyncJobsFunc.SetDefaultReturn(syncJobs)

	reposStore := repos.NewMockStoreFrom(repos.NewStore(logtest.Scoped(t), db))
	reposStore.RepoStoreFunc.SetDefaultReturn(mockRepos)

	perms := dbmocks.NewMockPermsStore()
	perms.SetUserExternalAccountPermsFunc.SetDefaultHook(func(_ context.Context, _ authz.UserIDWithExternalAccountID, repoIDs []int32, source authz.PermsSource) (*database.SetPermissionsResult, error) {
		wantIDs := []int32{1, 2, 3, 4}
		assert.Equal(t, wantIDs, repoIDs)
		return &database.SetPermissionsResult{}, nil
	})

	s := newPermsSyncer(logtest.Scoped(t), db, reposStore, perms, timeutil.Now)
	s.providerFactory = func(context.Context) []authz.Provider {
		return []authz.Provider{p}
	}

	p.fetchUserPerms = func(context.Context, *extsvc.Account) (*authz.ExternalUserPermissions, error) {
		return &authz.ExternalUserPermissions{
			Exacts: []extsvc.RepoID{"1", "2", "3", "4"},
		}, nil
	}

	_, providers, err := s.syncUserPerms(context.Background(), 1, true, authz.FetchPermsOptions{})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, database.CodeHostStatusesSet{{
		ProviderID:   "https://gitlab.com/",
		ProviderType: "gitlab",
		Status:       database.CodeHostStatusSuccess,
		Message:      "FetchUserPerms",
	}}, providers)
}

func TestPermsSyncer_syncUserPerms_listExternalAccountsError(t *testing.T) {
	p := &mockProvider{
		id:          1,
		serviceType: extsvc.TypeGitLab,
		serviceID:   "https://gitlab.com/",
	}

	users := dbmocks.NewMockUserStore()
	users.GetByIDFunc.SetDefaultHook(func(ctx context.Context, id int32) (*types.User, error) {
		return &types.User{ID: id}, nil
	})

	mockRepos := dbmocks.NewMockRepoStore()
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

	externalAccounts := dbmocks.NewMockUserExternalAccountsStore()
	externalAccounts.ListFunc.SetDefaultHook(func(ctx context.Context, options database.ExternalAccountsListOptions) ([]*extsvc.Account, error) {
		// Force an error here to bail out of fetchUserPermsViaExternalAccounts
		return nil, errors.New("forced error")
	})

	userEmails := dbmocks.NewMockUserEmailsStore()
	userEmails.ListByUserFunc.SetDefaultHook(func(ctx context.Context, options database.UserEmailsListOptions) ([]*database.UserEmail, error) {
		return []*database.UserEmail{}, nil
	})

	db := dbmocks.NewMockDB()
	db.UsersFunc.SetDefaultReturn(users)
	db.ReposFunc.SetDefaultReturn(mockRepos)
	db.UserExternalAccountsFunc.SetDefaultReturn(externalAccounts)
	db.UserEmailsFunc.SetDefaultReturn(userEmails)

	reposStore := repos.NewMockStoreFrom(repos.NewStore(logtest.Scoped(t), db))
	reposStore.RepoStoreFunc.SetDefaultReturn(mockRepos)

	perms := dbmocks.NewMockPermsStore()
	perms.SetUserExternalAccountPermsFunc.SetDefaultHook(func(_ context.Context, user authz.UserIDWithExternalAccountID, repoIDs []int32, source authz.PermsSource) (*database.SetPermissionsResult, error) {
		assert.Equal(t, []int32{1, 2, 3, 4, 5}, repoIDs)
		return &database.SetPermissionsResult{}, nil
	})

	s := newPermsSyncer(logtest.Scoped(t), db, reposStore, perms, timeutil.Now)
	s.providerFactory = func(context.Context) []authz.Provider {
		return []authz.Provider{p}
	}

	t.Run("fetchUserPermsViaExternalAccounts", func(t *testing.T) {
		_, _, err := s.syncUserPerms(context.Background(), 1, true, authz.FetchPermsOptions{})
		require.Error(t, err, "expected error")
	})
}

func TestPermsSyncer_syncUserPerms_fetchAccount(t *testing.T) {
	p1 := &mockProvider{
		id:          1,
		serviceType: extsvc.TypeGitLab,
		serviceID:   "https://gitlab.com/",
	}
	p2 := &mockProvider{
		id:          2,
		serviceType: extsvc.TypeGitHub,
		serviceID:   "https://github.com/",
	}

	users := dbmocks.NewMockUserStore()
	users.GetByIDFunc.SetDefaultHook(func(ctx context.Context, id int32) (*types.User, error) {
		return &types.User{ID: id}, nil
	})

	mockRepos := dbmocks.NewMockRepoStore()
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

	externalAccounts := dbmocks.NewMockUserExternalAccountsStore()
	externalAccounts.ListFunc.SetDefaultHook(func(_ context.Context, opts database.ExternalAccountsListOptions) ([]*extsvc.Account, error) {
		if opts.OnlyExpired {
			return nil, nil
		}

		return []*extsvc.Account{{
			UserID: 1,
			AccountSpec: extsvc.AccountSpec{
				ServiceType: p1.serviceType,
				ServiceID:   p1.serviceID,
				AccountID:   "1",
			},
		}}, nil
	})

	userEmails := dbmocks.NewMockUserEmailsStore()
	userEmails.ListByUserFunc.SetDefaultHook(func(ctx context.Context, options database.UserEmailsListOptions) ([]*database.UserEmail, error) {
		return []*database.UserEmail{}, nil
	})

	permissionSyncJobs := dbmocks.NewMockPermissionSyncJobStore()
	permissionSyncJobs.GetLatestFinishedSyncJobFunc.SetDefaultReturn(&database.PermissionSyncJob{ID: 1, FinishedAt: timeutil.Now().Add(-1 * time.Hour)}, nil)

	db := dbmocks.NewMockDB()
	db.UsersFunc.SetDefaultReturn(users)
	db.ReposFunc.SetDefaultReturn(mockRepos)
	db.UserExternalAccountsFunc.SetDefaultReturn(externalAccounts)
	db.UserEmailsFunc.SetDefaultReturn(userEmails)
	db.PermissionSyncJobsFunc.SetDefaultReturn(permissionSyncJobs)

	reposStore := repos.NewMockStoreFrom(repos.NewStore(logtest.Scoped(t), db))
	reposStore.RepoStoreFunc.SetDefaultReturn(mockRepos)

	perms := dbmocks.NewMockPermsStore()
	perms.SetUserExternalAccountPermsFunc.SetDefaultHook(func(_ context.Context, user authz.UserIDWithExternalAccountID, repoIDs []int32, source authz.PermsSource) (*database.SetPermissionsResult, error) {
		return &database.SetPermissionsResult{
			Added:   len(repoIDs),
			Removed: 0,
			Found:   len(repoIDs),
		}, nil
	})

	fetchUserPermsSuccessfully := func(ctx context.Context, account *extsvc.Account) (*authz.ExternalUserPermissions, error) {
		return &authz.ExternalUserPermissions{
			Exacts: []extsvc.RepoID{"1", "2", "3", "4", "5"},
		}, nil
	}
	p1.fetchUserPerms = fetchUserPermsSuccessfully
	p2.fetchUserPerms = func(context.Context, *extsvc.Account) (*authz.ExternalUserPermissions, error) {
		return nil, errors.New("should never call fetchUserPerms for github")
	}

	s := newPermsSyncer(logtest.Scoped(t), db, reposStore, perms, timeutil.Now)
	s.providerFactory = func(context.Context) []authz.Provider {
		return []authz.Provider{p1, p2}
	}

	tests := []struct {
		name                string
		fetchAccountError   error
		fetchUserPermsError error
		statuses            database.CodeHostStatusesSet
	}{
		{
			name: "gitlab perms sync succeeds, github FetchAccount succeeds",
			statuses: database.CodeHostStatusesSet{{
				ProviderID:   p1.serviceID,
				ProviderType: p1.serviceType,
				Status:       database.CodeHostStatusSuccess,
				Message:      "FetchUserPerms",
			}},
		},
		{
			name: "gitlab perms sync succeeds, github FetchAccount fails",
			statuses: database.CodeHostStatusesSet{{
				ProviderID:   p2.serviceID,
				ProviderType: p2.serviceType,
				Status:       database.CodeHostStatusError,
				Message:      "FetchAccount: no account found for this user",
			}, {
				ProviderID:   p1.serviceID,
				ProviderType: p1.serviceType,
				Status:       database.CodeHostStatusSuccess,
				Message:      "FetchUserPerms",
			}},
			fetchAccountError: errors.New("no account found for this user"),
		},
		{
			name: "gitlab perms sync fails, github FetchAccount succeeds",
			statuses: database.CodeHostStatusesSet{{
				ProviderID:   p1.serviceID,
				ProviderType: p1.serviceType,
				Status:       database.CodeHostStatusError,
				Message:      "FetchUserPerms: horse error",
			}},
			fetchUserPermsError: errors.New("horse error"),
		},
		{
			name: "gitlab perms sync fails, github FetchAccount fails",
			statuses: database.CodeHostStatusesSet{{
				ProviderID:   p2.serviceID,
				ProviderType: p2.serviceType,
				Status:       database.CodeHostStatusError,
				Message:      "FetchAccount: no account found for this user",
			}, {
				ProviderID:   p1.serviceID,
				ProviderType: p1.serviceType,
				Status:       database.CodeHostStatusError,
				Message:      "FetchUserPerms: horse error",
			}},
			fetchAccountError:   errors.New("no account found for this user"),
			fetchUserPermsError: errors.New("horse error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.fetchAccountError != nil {
				p2.fetchAccount = func(context.Context, *types.User) (*extsvc.Account, error) {
					return nil, test.fetchAccountError
				}
			}
			if test.fetchUserPermsError != nil {
				p1.fetchUserPerms = func(context.Context, *extsvc.Account) (*authz.ExternalUserPermissions, error) {
					return nil, test.fetchUserPermsError
				}
			}

			t.Cleanup(func() {
				p1.fetchUserPerms = fetchUserPermsSuccessfully
				p2.fetchAccount = nil
			})

			_, s, err := s.syncUserPerms(context.Background(), 1, true, authz.FetchPermsOptions{})
			require.NoError(t, err, "expected to swallow the error")
			require.Equal(t, test.statuses, s)
		})
	}
}

// If we hit a temporary error from the provider we should fetch existing
// permissions from the database
func TestPermsSyncer_syncUserPermsTemporaryProviderError(t *testing.T) {
	t.Run("no existing permissions", func(t *testing.T) {
		p := &mockProvider{
			id:          1,
			serviceType: extsvc.TypeGitLab,
			serviceID:   "https://gitlab.com/",
		}

		extAccount := extsvc.Account{
			AccountSpec: extsvc.AccountSpec{
				ServiceType: p.ServiceType(),
				ServiceID:   p.ServiceID(),
			},
		}

		users := dbmocks.NewMockUserStore()
		users.GetByIDFunc.SetDefaultHook(func(ctx context.Context, id int32) (*types.User, error) {
			return &types.User{ID: id}, nil
		})

		mockRepos := dbmocks.NewMockRepoStore()
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

		userEmails := dbmocks.NewMockUserEmailsStore()

		externalAccounts := dbmocks.NewMockUserExternalAccountsStore()
		externalAccounts.ListFunc.SetDefaultHook(func(_ context.Context, opts database.ExternalAccountsListOptions) ([]*extsvc.Account, error) {
			if opts.OnlyExpired {
				return []*extsvc.Account{}, nil
			}
			return []*extsvc.Account{&extAccount}, nil
		})
		featureFlags := dbmocks.NewMockFeatureFlagStore()

		subRepoPerms := dbmocks.NewMockSubRepoPermsStore()
		subRepoPerms.GetByUserAndServiceFunc.SetDefaultReturn(nil, nil)

		syncJobs := dbmocks.NewMockPermissionSyncJobStore()
		syncJobs.GetLatestFinishedSyncJobFunc.SetDefaultReturn(nil, nil)

		db := dbmocks.NewMockDB()
		db.UsersFunc.SetDefaultReturn(users)
		db.ReposFunc.SetDefaultReturn(mockRepos)
		db.UserEmailsFunc.SetDefaultReturn(userEmails)
		db.UserExternalAccountsFunc.SetDefaultReturn(externalAccounts)
		db.SubRepoPermsFunc.SetDefaultReturn(subRepoPerms)
		db.FeatureFlagsFunc.SetDefaultReturn(featureFlags)
		db.PermissionSyncJobsFunc.SetDefaultReturn(syncJobs)

		reposStore := repos.NewMockStoreFrom(repos.NewStore(logtest.Scoped(t), db))
		reposStore.RepoStoreFunc.SetDefaultReturn(mockRepos)

		perms := dbmocks.NewMockPermsStore()
		perms.SetUserExternalAccountPermsFunc.SetDefaultHook(func(_ context.Context, user authz.UserIDWithExternalAccountID, repoIDs []int32, source authz.PermsSource) (*database.SetPermissionsResult, error) {
			assert.Equal(t, []int32{}, repoIDs)
			return &database.SetPermissionsResult{}, nil
		})

		s := newPermsSyncer(logtest.Scoped(t), db, reposStore, perms, timeutil.Now)
		s.providerFactory = func(context.Context) []authz.Provider {
			return []authz.Provider{p}
		}

		p.fetchUserPerms = func(context.Context, *extsvc.Account) (*authz.ExternalUserPermissions, error) {
			// DeadlineExceeded implements the Temporary interface
			return nil, context.DeadlineExceeded
		}

		_, providers, err := s.syncUserPerms(context.Background(), 1, true, authz.FetchPermsOptions{})
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, database.CodeHostStatusesSet{{
			ProviderID:   "https://gitlab.com/",
			ProviderType: "gitlab",
			Status:       database.CodeHostStatusError,
			Message:      "FetchUserPerms: context deadline exceeded",
		}}, providers)
	})

	t.Run("reinsert permissions with IP address on temporary error", func(t *testing.T) {
		p := &mockProvider{
			id:          1,
			serviceType: extsvc.TypeGitLab,
			serviceID:   "https://gitlab.com/",
		}

		extAccount := extsvc.Account{
			AccountSpec: extsvc.AccountSpec{
				ServiceType: p.ServiceType(),
				ServiceID:   p.ServiceID(),
			},
		}

		users := dbmocks.NewMockUserStore()
		users.GetByIDFunc.SetDefaultHook(func(ctx context.Context, id int32) (*types.User, error) {
			return &types.User{ID: id}, nil
		})

		mockRepos := dbmocks.NewMockRepoStore()
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

		userEmails := dbmocks.NewMockUserEmailsStore()

		externalAccounts := dbmocks.NewMockUserExternalAccountsStore()
		externalAccounts.ListFunc.SetDefaultHook(func(_ context.Context, opts database.ExternalAccountsListOptions) ([]*extsvc.Account, error) {
			if opts.OnlyExpired {
				return []*extsvc.Account{}, nil
			}
			return []*extsvc.Account{&extAccount}, nil
		})
		featureFlags := dbmocks.NewMockFeatureFlagStore()

		subRepoPerms := dbmocks.NewMockSubRepoPermsStore()

		// Set up initial state with permissions entry including IP address
		initialPerms := map[api.ExternalRepoSpec]authz.SubRepoPermissionsWithIPs{
			{ID: "repo1", ServiceType: p.ServiceType(), ServiceID: p.ServiceID()}: {
				Paths: []authz.PathWithIP{{Path: "/include1", IP: "1.1.1.1"}},
			},
		}
		subRepoPerms.GetByUserAndServiceWithIPsFunc.SetDefaultReturn(initialPerms, nil)

		// Set up spy for UpsertWithSpecWithIPs
		var upsertCallCount int

		subRepoPerms.UpsertWithSpecWithIPsFunc.SetDefaultHook(func(ctx context.Context, userID int32, spec api.ExternalRepoSpec, perms authz.SubRepoPermissionsWithIPs) error {
			upsertCallCount++

			// Check that we're re-inserting the same permissions
			assert.Equal(t, int32(1), userID, "Incorrect user ID passed to UpsertWithSpecWithIPs")
			assert.Equal(t, api.ExternalRepoSpec{ID: "repo1", ServiceType: p.ServiceType(), ServiceID: p.ServiceID()}, spec, "Incorrect spec passed to UpsertWithSpecWithIPs")
			assert.Equal(t, initialPerms[spec], perms, "Incorrect permissions passed to UpsertWithSpecWithIPs")
			return nil
		})

		syncJobs := dbmocks.NewMockPermissionSyncJobStore()
		syncJobs.GetLatestFinishedSyncJobFunc.SetDefaultReturn(nil, nil)

		db := dbmocks.NewMockDB()
		db.UsersFunc.SetDefaultReturn(users)
		db.ReposFunc.SetDefaultReturn(mockRepos)
		db.UserEmailsFunc.SetDefaultReturn(userEmails)
		db.UserExternalAccountsFunc.SetDefaultReturn(externalAccounts)
		db.SubRepoPermsFunc.SetDefaultReturn(subRepoPerms)
		db.FeatureFlagsFunc.SetDefaultReturn(featureFlags)
		db.PermissionSyncJobsFunc.SetDefaultReturn(syncJobs)

		reposStore := repos.NewMockStoreFrom(repos.NewStore(logtest.Scoped(t), db))
		reposStore.RepoStoreFunc.SetDefaultReturn(mockRepos)

		perms := dbmocks.NewMockPermsStore()
		perms.SetUserExternalAccountPermsFunc.SetDefaultHook(func(_ context.Context, user authz.UserIDWithExternalAccountID, repoIDs []int32, source authz.PermsSource) (*database.SetPermissionsResult, error) {
			assert.Equal(t, []int32{}, repoIDs)
			return &database.SetPermissionsResult{}, nil
		})

		s := newPermsSyncer(logtest.Scoped(t), db, reposStore, perms, timeutil.Now)
		s.providerFactory = func(context.Context) []authz.Provider {
			return []authz.Provider{p}
		}

		// Set up initial state with permissions entry including IP address

		p.fetchUserPerms = func(context.Context, *extsvc.Account) (*authz.ExternalUserPermissions, error) {
			return nil, context.DeadlineExceeded
		}

		_, providers, err := s.syncUserPerms(context.Background(), 1, true, authz.FetchPermsOptions{})
		if err != nil {
			t.Fatal(err)
		}

		// Verify that the code host status is set to error
		assert.Equal(t, database.CodeHostStatusesSet{{
			ProviderID:   "https://gitlab.com/",
			ProviderType: "gitlab",
			Status:       database.CodeHostStatusError,
			Message:      "FetchUserPerms: context deadline exceeded",
		}}, providers)

		// Verify that the UpsertWithSpecWithIPs was called
		assert.Equal(t, 1, upsertCallCount, "UpsertWithSpecWithIPs should have been called once")
	})

	t.Run("reinsert permissions without address on temporary error", func(t *testing.T) {
		p := &mockProvider{
			id:          1,
			serviceType: extsvc.TypeGitLab,
			serviceID:   "https://gitlab.com/",
		}

		extAccount := extsvc.Account{
			AccountSpec: extsvc.AccountSpec{
				ServiceType: p.ServiceType(),
				ServiceID:   p.ServiceID(),
			},
		}

		users := dbmocks.NewMockUserStore()
		users.GetByIDFunc.SetDefaultHook(func(ctx context.Context, id int32) (*types.User, error) {
			return &types.User{ID: id}, nil
		})

		mockRepos := dbmocks.NewMockRepoStore()
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

		userEmails := dbmocks.NewMockUserEmailsStore()

		externalAccounts := dbmocks.NewMockUserExternalAccountsStore()
		externalAccounts.ListFunc.SetDefaultHook(func(_ context.Context, opts database.ExternalAccountsListOptions) ([]*extsvc.Account, error) {
			if opts.OnlyExpired {
				return []*extsvc.Account{}, nil
			}
			return []*extsvc.Account{&extAccount}, nil
		})
		featureFlags := dbmocks.NewMockFeatureFlagStore()

		subRepoPerms := dbmocks.NewMockSubRepoPermsStore()

		// Set up a fake implementation of IP version of the sub repo permissions getter
		// that simulates a database entry without an IP address
		var getByUserAndServiceWithIPsCalled bool
		subRepoPerms.GetByUserAndServiceWithIPsFunc.SetDefaultHook(func(ctx context.Context, userID int32, serviceType string, serviceID string, backfillWithWildcardIP bool) (map[api.ExternalRepoSpec]authz.SubRepoPermissionsWithIPs, error) {
			getByUserAndServiceWithIPsCalled = true

			assert.Equal(t, int32(1), userID, "Incorrect user ID passed to UpsertWithSpec")
			assert.Equal(t, p.ServiceType(), serviceType, "Incorrect service type passed to UpsertWithSpec")
			assert.Equal(t, p.ServiceID(), serviceID, "Incorrect service ID passed to UpsertWithSpec")
			assert.False(t, backfillWithWildcardIP, "backfillWithWildcardIP should be false since we don't want a fake IP address")

			return nil, database.IPsNotSyncedError
		})

		// Set up initial state with permissions entry without an IP address
		initialPerms := map[api.ExternalRepoSpec]authz.SubRepoPermissions{
			{ID: "repo1", ServiceType: p.ServiceType(), ServiceID: p.ServiceID()}: {
				Paths: []string{"/include1"},
			},
		}
		subRepoPerms.GetByUserAndServiceFunc.SetDefaultReturn(initialPerms, nil)

		// Set up spy for UpsertWithSpec

		var upsertCallCount int

		subRepoPerms.UpsertWithSpecFunc.SetDefaultHook(func(ctx context.Context, userID int32, spec api.ExternalRepoSpec, perms authz.SubRepoPermissions) error {
			upsertCallCount++

			// Check that we're re-inserting the same permissions
			assert.Equal(t, int32(1), userID, "Incorrect user ID passed to UpsertWithSpec")
			assert.Equal(t, api.ExternalRepoSpec{ID: "repo1", ServiceType: p.ServiceType(), ServiceID: p.ServiceID()}, spec, "Incorrect spec passed to UpsertWithSpec")
			assert.Equal(t, initialPerms[spec], perms, "Incorrect permissions passed to UpsertWithSpec")
			return nil
		})

		syncJobs := dbmocks.NewMockPermissionSyncJobStore()
		syncJobs.GetLatestFinishedSyncJobFunc.SetDefaultReturn(nil, nil)

		db := dbmocks.NewMockDB()
		db.UsersFunc.SetDefaultReturn(users)
		db.ReposFunc.SetDefaultReturn(mockRepos)
		db.UserEmailsFunc.SetDefaultReturn(userEmails)
		db.UserExternalAccountsFunc.SetDefaultReturn(externalAccounts)
		db.SubRepoPermsFunc.SetDefaultReturn(subRepoPerms)
		db.FeatureFlagsFunc.SetDefaultReturn(featureFlags)
		db.PermissionSyncJobsFunc.SetDefaultReturn(syncJobs)

		reposStore := repos.NewMockStoreFrom(repos.NewStore(logtest.Scoped(t), db))
		reposStore.RepoStoreFunc.SetDefaultReturn(mockRepos)

		perms := dbmocks.NewMockPermsStore()
		perms.SetUserExternalAccountPermsFunc.SetDefaultHook(func(_ context.Context, user authz.UserIDWithExternalAccountID, repoIDs []int32, source authz.PermsSource) (*database.SetPermissionsResult, error) {
			assert.Equal(t, []int32{}, repoIDs)
			return &database.SetPermissionsResult{}, nil
		})

		s := newPermsSyncer(logtest.Scoped(t), db, reposStore, perms, timeutil.Now)
		s.providerFactory = func(context.Context) []authz.Provider {
			return []authz.Provider{p}
		}

		// Set up initial state with permissions entry including IP address

		p.fetchUserPerms = func(context.Context, *extsvc.Account) (*authz.ExternalUserPermissions, error) {
			return nil, context.DeadlineExceeded
		}

		_, providers, err := s.syncUserPerms(context.Background(), 1, true, authz.FetchPermsOptions{})
		if err != nil {
			t.Fatal(err)
		}

		// Verify that the code host status is set to error
		assert.Equal(t, database.CodeHostStatusesSet{{
			ProviderID:   "https://gitlab.com/",
			ProviderType: "gitlab",
			Status:       database.CodeHostStatusError,
			Message:      "FetchUserPerms: context deadline exceeded",
		}}, providers)

		// Verify that fake version of the sub repo permissions getter that forced a fallback was called
		assert.True(t, getByUserAndServiceWithIPsCalled, "getByUserWithIPsFunc should have been called")

		// Verify that the UpsertWithSpec (non IP version) was called once
		assert.Equal(t, 1, upsertCallCount, "UpsertWithSpec should have been called once")
	})
}

func TestPermsSyncer_syncUserPerms_noPerms(t *testing.T) {
	p := &mockProvider{
		id:          1,
		serviceType: extsvc.TypeGitLab,
		serviceID:   "https://gitlab.com/",
	}

	extAccount := extsvc.Account{
		AccountSpec: extsvc.AccountSpec{
			ServiceType: p.ServiceType(),
			ServiceID:   p.ServiceID(),
		},
	}

	users := dbmocks.NewMockUserStore()
	users.GetByIDFunc.SetDefaultHook(func(ctx context.Context, id int32) (*types.User, error) {
		return &types.User{ID: id}, nil
	})

	mockRepos := dbmocks.NewMockRepoStore()
	mockRepos.ListMinimalReposFunc.SetDefaultHook(func(ctx context.Context, opt database.ReposListOptions) ([]types.MinimalRepo, error) {
		if !opt.OnlyPrivate {
			return nil, errors.New("OnlyPrivate want true but got false")
		}
		return []types.MinimalRepo{{ID: 1}}, nil
	})

	userEmails := dbmocks.NewMockUserEmailsStore()
	externalAccounts := dbmocks.NewMockUserExternalAccountsStore()
	// return only non expired accounts
	externalAccounts.ListFunc.SetDefaultHook(func(_ context.Context, opts database.ExternalAccountsListOptions) ([]*extsvc.Account, error) {
		if opts.ExcludeExpired {
			return []*extsvc.Account{&extAccount}, nil
		}
		return nil, nil
	})

	featureFlags := dbmocks.NewMockFeatureFlagStore()

	syncJobs := dbmocks.NewStrictMockPermissionSyncJobStore()
	syncJobs.GetLatestFinishedSyncJobFunc.SetDefaultReturn(nil, nil)

	db := dbmocks.NewMockDB()
	db.UsersFunc.SetDefaultReturn(users)
	db.ReposFunc.SetDefaultReturn(mockRepos)
	db.UserEmailsFunc.SetDefaultReturn(userEmails)
	db.UserExternalAccountsFunc.SetDefaultReturn(externalAccounts)
	db.FeatureFlagsFunc.SetDefaultReturn(featureFlags)
	db.PermissionSyncJobsFunc.SetDefaultReturn(syncJobs)

	reposStore := repos.NewMockStoreFrom(repos.NewStore(logtest.Scoped(t), db))
	reposStore.RepoStoreFunc.SetDefaultReturn(mockRepos)

	perms := dbmocks.NewMockPermsStore()
	perms.SetUserExternalAccountPermsFunc.SetDefaultHook(func(_ context.Context, user authz.UserIDWithExternalAccountID, repoIDs []int32, source authz.PermsSource) (*database.SetPermissionsResult, error) {
		assert.Equal(t, int32(1), user.UserID)
		assert.Equal(t, []int32{1}, repoIDs)
		return &database.SetPermissionsResult{}, nil
	})

	s := newPermsSyncer(logtest.Scoped(t), db, reposStore, perms, timeutil.Now)
	s.providerFactory = func(context.Context) []authz.Provider {
		return []authz.Provider{p}
	}

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

			_, _, err := s.syncUserPerms(context.Background(), 1, test.noPerms, authz.FetchPermsOptions{})
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

	extAccount := extsvc.Account{
		AccountSpec: extsvc.AccountSpec{
			ServiceType: p.ServiceType(),
			ServiceID:   p.ServiceID(),
		},
	}

	users := dbmocks.NewMockUserStore()
	users.GetByIDFunc.SetDefaultHook(func(ctx context.Context, id int32) (*types.User, error) {
		return &types.User{ID: id}, nil
	})

	mockRepos := dbmocks.NewMockRepoStore()
	mockRepos.ListMinimalReposFunc.SetDefaultHook(func(ctx context.Context, opt database.ReposListOptions) ([]types.MinimalRepo, error) {
		if !opt.OnlyPrivate {
			return nil, errors.New("OnlyPrivate want true but got false")
		}
		return []types.MinimalRepo{{ID: 1}}, nil
	})

	externalServices := dbmocks.NewMockExternalServiceStore()
	userEmails := dbmocks.NewMockUserEmailsStore()

	externalAccounts := dbmocks.NewMockUserExternalAccountsStore()
	externalAccounts.ListFunc.SetDefaultReturn([]*extsvc.Account{&extAccount}, nil)

	syncJobs := dbmocks.NewStrictMockPermissionSyncJobStore()
	syncJobs.GetLatestFinishedSyncJobFunc.SetDefaultReturn(nil, nil)

	db := dbmocks.NewMockDB()
	db.UsersFunc.SetDefaultReturn(users)
	db.ReposFunc.SetDefaultReturn(mockRepos)
	db.ExternalServicesFunc.SetDefaultReturn(externalServices)
	db.UserEmailsFunc.SetDefaultReturn(userEmails)
	db.UserExternalAccountsFunc.SetDefaultReturn(externalAccounts)
	db.FeatureFlagsFunc.SetDefaultReturn(dbmocks.NewMockFeatureFlagStore())
	db.PermissionSyncJobsFunc.SetDefaultReturn(syncJobs)

	reposStore := repos.NewMockStore()

	perms := dbmocks.NewMockPermsStore()
	s := newPermsSyncer(logtest.Scoped(t), db, reposStore, perms, timeutil.Now)
	s.providerFactory = func(context.Context) []authz.Provider {
		return []authz.Provider{p}
	}

	t.Run("invalid token", func(t *testing.T) {
		p.fetchUserPerms = func(ctx context.Context, account *extsvc.Account) (*authz.ExternalUserPermissions, error) {
			return nil, &github.APIError{Code: http.StatusUnauthorized}
		}

		_, _, err := s.syncUserPerms(context.Background(), 1, false, authz.FetchPermsOptions{})
		if err != nil {
			t.Fatal(err)
		}

		mockrequire.Called(t, externalAccounts.TouchExpiredFunc)
	})

	t.Run("forbidden", func(t *testing.T) {
		p.fetchUserPerms = func(ctx context.Context, account *extsvc.Account) (*authz.ExternalUserPermissions, error) {
			return nil, gitlab.NewHTTPError(http.StatusForbidden, nil)
		}

		_, _, err := s.syncUserPerms(context.Background(), 1, false, authz.FetchPermsOptions{})
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

		_, _, err := s.syncUserPerms(context.Background(), 1, false, authz.FetchPermsOptions{})
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

	extAccount := extsvc.Account{
		AccountSpec: extsvc.AccountSpec{
			ServiceType: p.ServiceType(),
			ServiceID:   p.ServiceID(),
		},
	}

	users := dbmocks.NewMockUserStore()
	users.GetByIDFunc.SetDefaultHook(func(ctx context.Context, id int32) (*types.User, error) {
		return &types.User{ID: id}, nil
	})

	mockRepos := dbmocks.NewMockRepoStore()
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

	externalServices := dbmocks.NewMockExternalServiceStore()
	userEmails := dbmocks.NewMockUserEmailsStore()

	externalAccounts := dbmocks.NewMockUserExternalAccountsStore()
	externalAccounts.ListFunc.SetDefaultReturn([]*extsvc.Account{&extAccount}, nil)

	featureFlags := dbmocks.NewMockFeatureFlagStore()

	syncJobs := dbmocks.NewStrictMockPermissionSyncJobStore()
	syncJobs.GetLatestFinishedSyncJobFunc.SetDefaultReturn(nil, nil)

	db := dbmocks.NewMockDB()
	db.UsersFunc.SetDefaultReturn(users)
	db.ReposFunc.SetDefaultReturn(mockRepos)
	db.ExternalServicesFunc.SetDefaultReturn(externalServices)
	db.UserEmailsFunc.SetDefaultReturn(userEmails)
	db.UserExternalAccountsFunc.SetDefaultReturn(externalAccounts)
	db.FeatureFlagsFunc.SetDefaultReturn(featureFlags)
	db.PermissionSyncJobsFunc.SetDefaultReturn(syncJobs)

	reposStore := repos.NewMockStoreFrom(repos.NewStore(logtest.Scoped(t), db))
	reposStore.RepoStoreFunc.SetDefaultReturn(mockRepos)
	reposStore.ExternalServiceStoreFunc.SetDefaultReturn(externalServices)

	perms := dbmocks.NewMockPermsStore()

	perms.SetUserExternalAccountPermsFunc.SetDefaultReturn(&database.SetPermissionsResult{}, nil)

	s := newPermsSyncer(logtest.Scoped(t), db, reposStore, perms, timeutil.Now)
	s.providerFactory = func(context.Context) []authz.Provider {
		return []authz.Provider{p}
	}

	p.fetchUserPerms = func(context.Context, *extsvc.Account) (*authz.ExternalUserPermissions, error) {
		return &authz.ExternalUserPermissions{
			IncludeContains: []extsvc.RepoID{"//Engineering/"},
			ExcludeContains: []extsvc.RepoID{"//Engineering/Security/"},
		}, nil
	}

	_, _, err := s.syncUserPerms(context.Background(), 1, false, authz.FetchPermsOptions{})
	if err != nil {
		t.Fatal(err)
	}
}

func TestPermsSyncer_syncUserPerms_subRepoPermissions(t *testing.T) {
	p := &mockProvider{
		serviceType: extsvc.TypePerforce,
		serviceID:   "ssl:111.222.333.444:1666",
	}

	users := dbmocks.NewMockUserStore()
	users.GetByIDFunc.SetDefaultHook(func(ctx context.Context, id int32) (*types.User, error) {
		return &types.User{ID: id}, nil
	})

	mockRepos := dbmocks.NewMockRepoStore()
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

	externalServices := dbmocks.NewMockExternalServiceStore()
	userEmails := dbmocks.NewMockUserEmailsStore()

	externalAccounts := dbmocks.NewMockUserExternalAccountsStore()
	externalAccounts.ListFunc.SetDefaultReturn([]*extsvc.Account{}, nil)
	externalAccounts.UpsertFunc.SetDefaultReturn(&extsvc.Account{
		ID: 1,
		AccountSpec: extsvc.AccountSpec{
			ServiceType: p.ServiceType(),
			ServiceID:   p.ServiceID(),
		},
	}, nil)

	subRepoPerms := dbmocks.NewMockSubRepoPermsStore()

	featureFlags := dbmocks.NewMockFeatureFlagStore()

	syncJobs := dbmocks.NewStrictMockPermissionSyncJobStore()
	syncJobs.GetLatestFinishedSyncJobFunc.SetDefaultReturn(nil, nil)

	db := dbmocks.NewMockDB()
	db.UsersFunc.SetDefaultReturn(users)
	db.ReposFunc.SetDefaultReturn(mockRepos)
	db.ExternalServicesFunc.SetDefaultReturn(externalServices)
	db.UserEmailsFunc.SetDefaultReturn(userEmails)
	db.UserExternalAccountsFunc.SetDefaultReturn(externalAccounts)
	db.SubRepoPermsFunc.SetDefaultReturn(subRepoPerms)
	db.FeatureFlagsFunc.SetDefaultReturn(featureFlags)
	db.PermissionSyncJobsFunc.SetDefaultReturn(syncJobs)

	reposStore := repos.NewMockStoreFrom(repos.NewStore(logtest.Scoped(t), db))
	reposStore.RepoStoreFunc.SetDefaultReturn(mockRepos)
	reposStore.ExternalServiceStoreFunc.SetDefaultReturn(externalServices)

	perms := dbmocks.NewMockPermsStore()
	perms.SetUserExternalAccountPermsFunc.PushHook(func(ctx context.Context, ids authz.UserIDWithExternalAccountID, i []int32, ps authz.PermsSource) (*database.SetPermissionsResult, error) {
		if ids.ExternalAccountID == 0 {
			t.Fatal("ExternalAccountID should not be 0")
		}
		return &database.SetPermissionsResult{}, nil
	})

	s := newPermsSyncer(logtest.Scoped(t), db, reposStore, perms, timeutil.Now)
	s.providerFactory = func(context.Context) []authz.Provider {
		return []authz.Provider{p}
	}

	p.fetchUserPerms = func(context.Context, *extsvc.Account) (*authz.ExternalUserPermissions, error) {
		return &authz.ExternalUserPermissions{
			IncludeContains: []extsvc.RepoID{"//Engineering/"},
			ExcludeContains: []extsvc.RepoID{"//Engineering/Security/"},

			SubRepoPermissions: map[extsvc.RepoID]*authz.SubRepoPermissionsWithIPs{
				"abc": {
					Paths: []authz.PathWithIP{
						{Path: "/include1", IP: "1.1.1.1"},
						{Path: "/include2", IP: "1.1.1.1"},
						{Path: "-/exclude1", IP: "1.1.1.1"},
						{Path: "-/exclude2", IP: "1.1.1.1"},
					},
				},
				"def": {
					Paths: []authz.PathWithIP{
						{Path: "/include1", IP: "1.1.1.1"},
						{Path: "/include2", IP: "1.1.1.1"},
						{Path: "-/exclude1", IP: "1.1.1.1"},
						{Path: "-/exclude2", IP: "1.1.1.1"},
					},
				},
			},
		}, nil
	}
	p.fetchAccount = func(ctx context.Context, user *types.User) (*extsvc.Account, error) {
		return &extsvc.Account{
			AccountSpec: extsvc.AccountSpec{
				ServiceType: p.ServiceType(),
				ServiceID:   p.ServiceID(),
			},
		}, nil
	}

	_, _, err := s.syncUserPerms(context.Background(), 1, false, authz.FetchPermsOptions{})
	if err != nil {
		t.Fatal(err)
	}

	mockrequire.CalledN(t, subRepoPerms.UpsertWithSpecWithIPsFunc, 2)
}

func TestPermsSyncer_syncRepoPerms(t *testing.T) {
	mockRepos := dbmocks.NewMockRepoStore()
	mockFeatureFlags := dbmocks.NewMockFeatureFlagStore()
	mockSyncJobs := dbmocks.NewMockPermissionSyncJobStore()
	mockSyncJobs.GetLatestFinishedSyncJobFunc.SetDefaultReturn(&database.PermissionSyncJob{FinishedAt: timeutil.Now()}, nil)

	db := dbmocks.NewMockDB()
	db.ReposFunc.SetDefaultReturn(mockRepos)
	db.FeatureFlagsFunc.SetDefaultReturn(mockFeatureFlags)
	db.PermissionSyncJobsFunc.SetDefaultReturn(mockSyncJobs)

	newPermsSyncer := func(reposStore repos.Store, perms database.PermsStore) *permsSyncerImpl {
		return newPermsSyncer(logtest.Scoped(t), db, reposStore, perms, timeutil.Now)
	}

	t.Run("Err is nil when no authz provider", func(t *testing.T) {
		mockRepos.GetFunc.SetDefaultReturn(
			&types.Repo{
				ID:      1,
				Private: true,
				ExternalRepo: api.ExternalRepoSpec{
					ServiceID: "https://gitlab.com/",
				},
				Sources: map[string]*types.SourceInfo{
					extsvc.URN(extsvc.TypeGitLab, 0): {},
				},
			},
			nil,
		)

		reposStore := repos.NewMockStoreFrom(repos.NewStore(logtest.Scoped(t), db))
		reposStore.RepoStoreFunc.SetDefaultReturn(mockRepos)

		perms := dbmocks.NewMockPermsStore()
		s := newPermsSyncer(reposStore, perms)
		s.providerFactory = func(context.Context) []authz.Provider {
			return []authz.Provider{}
		}

		// error should be nil in this case
		_, _, err := s.syncRepoPerms(context.Background(), 1, false, authz.FetchPermsOptions{})
		if err != nil {
			t.Fatal(err)
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

		reposStore := repos.NewMockStoreFrom(repos.NewStore(logtest.Scoped(t), db))
		reposStore.RepoStoreFunc.SetDefaultReturn(mockRepos)

		perms := dbmocks.NewMockPermsStore()
		perms.TransactFunc.SetDefaultReturn(perms, nil)
		perms.GetUserIDsByExternalAccountsFunc.SetDefaultReturn(map[string]authz.UserIDWithExternalAccountID{"user": {UserID: 1, ExternalAccountID: 1}}, nil)
		perms.SetRepoPermsFunc.SetDefaultHook(func(_ context.Context, repoID int32, userIDs []authz.UserIDWithExternalAccountID, source authz.PermsSource) (*database.SetPermissionsResult, error) {
			assert.Equal(t, int32(1), repoID)
			assert.Equal(t, []authz.UserIDWithExternalAccountID{{UserID: 1, ExternalAccountID: 1}}, userIDs)
			assert.Equal(t, authz.SourceRepoSync, source)
			return nil, nil
		})

		s := newPermsSyncer(reposStore, perms)
		s.providerFactory = func(context.Context) []authz.Provider {
			return []authz.Provider{p1, p2}
		}

		_, _, err := s.syncRepoPerms(context.Background(), 1, false, authz.FetchPermsOptions{})
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

		reposStore := repos.NewMockStoreFrom(repos.NewStore(logtest.Scoped(t), db))
		reposStore.RepoStoreFunc.SetDefaultReturn(mockRepos)

		perms := dbmocks.NewMockPermsStore()
		perms.TransactFunc.SetDefaultReturn(perms, nil)
		perms.GetUserIDsByExternalAccountsFunc.SetDefaultReturn(map[string]authz.UserIDWithExternalAccountID{"user": {UserID: 1, ExternalAccountID: 1}}, nil)
		perms.SetRepoPermsFunc.SetDefaultHook(func(_ context.Context, repoID int32, userIDs []authz.UserIDWithExternalAccountID, source authz.PermsSource) (*database.SetPermissionsResult, error) {
			assert.Equal(t, int32(1), repoID)
			assert.Equal(t, []authz.UserIDWithExternalAccountID{{UserID: 1, ExternalAccountID: 1}}, userIDs)
			assert.Equal(t, authz.SourceRepoSync, source)
			return nil, nil
		})

		s := newPermsSyncer(reposStore, perms)
		s.providerFactory = func(context.Context) []authz.Provider {
			return []authz.Provider{}
		}

		_, _, err := s.syncRepoPerms(context.Background(), 1, false, authz.FetchPermsOptions{})
		if err != nil {
			t.Fatal(err)
		}

		mockrequire.NotCalled(t, perms.SetRepoPendingPermissionsFunc)
	})

	t.Run("repo sync that returns 404 does not have error in provider status", func(t *testing.T) {
		p := &mockProvider{
			id:          1,
			serviceType: extsvc.TypeGitHub,
			serviceID:   "https://github.com/",
			fetchRepoPerms: func(ctx context.Context, repo *extsvc.Repository, opts authz.FetchPermsOptions) ([]extsvc.AccountID, error) {
				return []extsvc.AccountID{}, &github.APIError{Code: http.StatusNotFound}
			},
		}

		mockRepos.GetFunc.SetDefaultReturn(
			&types.Repo{
				ID:      1,
				Private: true,
				ExternalRepo: api.ExternalRepoSpec{
					ServiceID: p.ServiceID(),
				},
				Sources: map[string]*types.SourceInfo{
					p.URN(): {},
				},
			},
			nil,
		)

		reposStore := repos.NewMockStoreFrom(repos.NewStore(logtest.Scoped(t), db))
		reposStore.RepoStoreFunc.SetDefaultReturn(mockRepos)

		perms := dbmocks.NewMockPermsStore()
		perms.TransactFunc.SetDefaultReturn(perms, nil)
		perms.GetUserIDsByExternalAccountsFunc.SetDefaultReturn(map[string]authz.UserIDWithExternalAccountID{"user": {UserID: 1, ExternalAccountID: 1}}, nil)
		perms.SetRepoPermsFunc.SetDefaultHook(func(_ context.Context, repoID int32, userIDs []authz.UserIDWithExternalAccountID, source authz.PermsSource) (*database.SetPermissionsResult, error) {
			assert.Equal(t, int32(1), repoID)
			assert.Equal(t, []authz.UserIDWithExternalAccountID{{UserID: 1, ExternalAccountID: 1}}, userIDs)
			assert.Equal(t, authz.SourceRepoSync, source)
			return nil, nil
		})

		s := newPermsSyncer(reposStore, perms)
		s.providerFactory = func(context.Context) []authz.Provider {
			return []authz.Provider{p}
		}

		_, providerStates, err := s.syncRepoPerms(context.Background(), 1, false, authz.FetchPermsOptions{})
		if err != nil {
			t.Fatal(err)
		}

		assert.Greater(t, len(providerStates), 0)
		for _, ps := range providerStates {
			if ps.Status == database.CodeHostStatusError {
				t.Fatal("Did not expect provider status of ERROR")
			}
		}
	})

	p := &mockProvider{
		serviceType: extsvc.TypeGitLab,
		serviceID:   "https://gitlab.com/",
	}

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

	reposStore := repos.NewMockStoreFrom(repos.NewStore(logtest.Scoped(t), db))
	reposStore.RepoStoreFunc.SetDefaultReturn(mockRepos)

	perms := dbmocks.NewMockPermsStore()
	perms.TransactFunc.SetDefaultReturn(perms, nil)
	perms.GetUserIDsByExternalAccountsFunc.SetDefaultReturn(map[string]authz.UserIDWithExternalAccountID{"user": {UserID: 1, ExternalAccountID: 1}}, nil)
	perms.SetRepoPermsFunc.SetDefaultHook(func(_ context.Context, repoID int32, userIDs []authz.UserIDWithExternalAccountID, source authz.PermsSource) (*database.SetPermissionsResult, error) {
		assert.Equal(t, int32(1), repoID)
		assert.Equal(t, []authz.UserIDWithExternalAccountID{{UserID: 1, ExternalAccountID: 1}}, userIDs)
		assert.Equal(t, authz.SourceRepoSync, source)
		return nil, nil
	})
	perms.SetRepoPendingPermissionsFunc.SetDefaultHook(func(_ context.Context, accounts *extsvc.Accounts, _ *authz.RepoPermissions) error {
		wantAccounts := &extsvc.Accounts{
			ServiceType: p.ServiceType(),
			ServiceID:   p.ServiceID(),
			AccountIDs:  []string{"pending_user"},
		}
		assert.Equal(t, wantAccounts, accounts)
		return nil
	})

	s := newPermsSyncer(reposStore, perms)
	s.providerFactory = func(context.Context) []authz.Provider {
		return []authz.Provider{p}
	}

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

			_, _, err := s.syncRepoPerms(context.Background(), 1, test.noPerms, authz.FetchPermsOptions{})
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}
