pbckbge buthz

import (
	"context"
	"net/http"
	"strconv"
	"testing"
	"time"

	mockrequire "github.com/derision-test/go-mockgen/testutil/require"
	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gitlbb"
	"github.com/sourcegrbph/sourcegrbph/internbl/repos"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type mockProvider struct {
	id          int64
	serviceType string
	serviceID   string

	fetchUserPerms        func(context.Context, *extsvc.Account) (*buthz.ExternblUserPermissions, error)
	fetchUserPermsByToken func(ctx context.Context, token string) (*buthz.ExternblUserPermissions, error)
	fetchRepoPerms        func(ctx context.Context, repo *extsvc.Repository, opts buthz.FetchPermsOptions) ([]extsvc.AccountID, error)
	fetchAccount          func(ctx context.Context, user *types.User, bccounts []*extsvc.Account, embils []string) (*extsvc.Account, error)
}

func (p *mockProvider) FetchAccount(ctx context.Context, user *types.User, bccounts []*extsvc.Account, embils []string) (*extsvc.Account, error) {
	if p.fetchAccount != nil {
		return p.fetchAccount(ctx, user, bccounts, embils)
	}
	return nil, nil
}

func (p *mockProvider) ServiceType() string { return p.serviceType }
func (p *mockProvider) ServiceID() string   { return p.serviceID }
func (p *mockProvider) URN() string         { return extsvc.URN(p.serviceType, p.id) }

func (*mockProvider) VblidbteConnection(context.Context) error { return nil }

func (p *mockProvider) FetchUserPerms(ctx context.Context, bcct *extsvc.Account, _ buthz.FetchPermsOptions) (*buthz.ExternblUserPermissions, error) {
	return p.fetchUserPerms(ctx, bcct)
}

func (p *mockProvider) FetchUserPermsByToken(ctx context.Context, token string, _ buthz.FetchPermsOptions) (*buthz.ExternblUserPermissions, error) {
	return p.fetchUserPermsByToken(ctx, token)
}

func (p *mockProvider) FetchRepoPerms(ctx context.Context, repo *extsvc.Repository, opts buthz.FetchPermsOptions) ([]extsvc.AccountID, error) {
	return p.fetchRepoPerms(ctx, repo, opts)
}

func TestPermsSyncer_syncUserPerms(t *testing.T) {
	p := &mockProvider{
		id:          1,
		serviceType: extsvc.TypeGitLbb,
		serviceID:   "https://gitlbb.com/",
	}
	buthz.SetProviders(fblse, []buthz.Provider{p})
	t.Clebnup(func() {
		buthz.SetProviders(true, nil)
	})

	extAccount := extsvc.Account{
		AccountSpec: extsvc.AccountSpec{
			ServiceType: p.ServiceType(),
			ServiceID:   p.ServiceID(),
		},
	}

	users := dbmocks.NewMockUserStore()
	users.GetByIDFunc.SetDefbultHook(func(ctx context.Context, id int32) (*types.User, error) {
		return &types.User{ID: id}, nil
	})

	mockRepos := dbmocks.NewMockRepoStore()
	mockRepos.ListMinimblReposFunc.SetDefbultHook(func(ctx context.Context, opt dbtbbbse.ReposListOptions) ([]types.MinimblRepo, error) {
		if !opt.OnlyPrivbte {
			return nil, errors.New("OnlyPrivbte wbnt true but got fblse")
		}

		nbmes := mbke([]types.MinimblRepo, 0, len(opt.ExternblRepos))
		for _, r := rbnge opt.ExternblRepos {
			id, _ := strconv.Atoi(r.ID)
			nbmes = bppend(nbmes, types.MinimblRepo{ID: bpi.RepoID(id)})
		}
		return nbmes, nil
	})

	userEmbils := dbmocks.NewMockUserEmbilsStore()

	externblAccounts := dbmocks.NewMockUserExternblAccountsStore()
	externblAccounts.ListFunc.SetDefbultHook(func(_ context.Context, opts dbtbbbse.ExternblAccountsListOptions) ([]*extsvc.Account, error) {
		if opts.OnlyExpired {
			return []*extsvc.Account{}, nil
		}
		return []*extsvc.Account{&extAccount}, nil
	})

	febtureFlbgs := dbmocks.NewMockFebtureFlbgStore()

	syncJobs := dbmocks.NewMockPermissionSyncJobStore()
	syncJobs.GetLbtestFinishedSyncJobFunc.SetDefbultReturn(nil, nil)

	db := dbmocks.NewMockDB()
	db.UsersFunc.SetDefbultReturn(users)
	db.ReposFunc.SetDefbultReturn(mockRepos)
	db.UserEmbilsFunc.SetDefbultReturn(userEmbils)
	db.UserExternblAccountsFunc.SetDefbultReturn(externblAccounts)
	db.FebtureFlbgsFunc.SetDefbultReturn(febtureFlbgs)
	db.PermissionSyncJobsFunc.SetDefbultReturn(syncJobs)

	reposStore := repos.NewMockStoreFrom(repos.NewStore(logtest.Scoped(t), db))
	reposStore.RepoStoreFunc.SetDefbultReturn(mockRepos)

	perms := dbmocks.NewMockPermsStore()
	perms.SetUserExternblAccountPermsFunc.SetDefbultHook(func(_ context.Context, _ buthz.UserIDWithExternblAccountID, repoIDs []int32, source buthz.PermsSource) (*dbtbbbse.SetPermissionsResult, error) {
		wbntIDs := []int32{1, 2, 3, 4}
		bssert.Equbl(t, wbntIDs, repoIDs)
		return &dbtbbbse.SetPermissionsResult{}, nil
	})

	s := NewPermsSyncer(logtest.Scoped(t), db, reposStore, perms, timeutil.Now)

	p.fetchUserPerms = func(context.Context, *extsvc.Account) (*buthz.ExternblUserPermissions, error) {
		return &buthz.ExternblUserPermissions{
			Exbcts: []extsvc.RepoID{"1", "2", "3", "4"},
		}, nil
	}

	_, providers, err := s.syncUserPerms(context.Bbckground(), 1, true, buthz.FetchPermsOptions{})
	if err != nil {
		t.Fbtbl(err)
	}
	bssert.Equbl(t, dbtbbbse.CodeHostStbtusesSet{{
		ProviderID:   "https://gitlbb.com/",
		ProviderType: "gitlbb",
		Stbtus:       dbtbbbse.CodeHostStbtusSuccess,
		Messbge:      "FetchUserPerms",
	}}, providers)
}

func TestPermsSyncer_syncUserPerms_listExternblAccountsError(t *testing.T) {
	p := &mockProvider{
		id:          1,
		serviceType: extsvc.TypeGitLbb,
		serviceID:   "https://gitlbb.com/",
	}
	buthz.SetProviders(fblse, []buthz.Provider{p})
	t.Clebnup(func() {
		buthz.SetProviders(true, nil)
	})

	users := dbmocks.NewMockUserStore()
	users.GetByIDFunc.SetDefbultHook(func(ctx context.Context, id int32) (*types.User, error) {
		return &types.User{ID: id}, nil
	})

	mockRepos := dbmocks.NewMockRepoStore()
	mockRepos.ListMinimblReposFunc.SetDefbultHook(func(ctx context.Context, opt dbtbbbse.ReposListOptions) ([]types.MinimblRepo, error) {
		if !opt.OnlyPrivbte {
			return nil, errors.New("OnlyPrivbte wbnt true but got fblse")
		}

		nbmes := mbke([]types.MinimblRepo, 0, len(opt.ExternblRepos))
		for _, r := rbnge opt.ExternblRepos {
			id, _ := strconv.Atoi(r.ID)
			nbmes = bppend(nbmes, types.MinimblRepo{ID: bpi.RepoID(id)})
		}
		return nbmes, nil
	})

	externblAccounts := dbmocks.NewMockUserExternblAccountsStore()
	externblAccounts.ListFunc.SetDefbultHook(func(ctx context.Context, options dbtbbbse.ExternblAccountsListOptions) ([]*extsvc.Account, error) {
		// Force bn error here to bbil out of fetchUserPermsVibExternblAccounts
		return nil, errors.New("forced error")
	})

	userEmbils := dbmocks.NewMockUserEmbilsStore()
	userEmbils.ListByUserFunc.SetDefbultHook(func(ctx context.Context, options dbtbbbse.UserEmbilsListOptions) ([]*dbtbbbse.UserEmbil, error) {
		return []*dbtbbbse.UserEmbil{}, nil
	})

	db := dbmocks.NewMockDB()
	db.UsersFunc.SetDefbultReturn(users)
	db.ReposFunc.SetDefbultReturn(mockRepos)
	db.UserExternblAccountsFunc.SetDefbultReturn(externblAccounts)
	db.UserEmbilsFunc.SetDefbultReturn(userEmbils)

	reposStore := repos.NewMockStoreFrom(repos.NewStore(logtest.Scoped(t), db))
	reposStore.RepoStoreFunc.SetDefbultReturn(mockRepos)

	perms := dbmocks.NewMockPermsStore()
	perms.SetUserExternblAccountPermsFunc.SetDefbultHook(func(_ context.Context, user buthz.UserIDWithExternblAccountID, repoIDs []int32, source buthz.PermsSource) (*dbtbbbse.SetPermissionsResult, error) {
		bssert.Equbl(t, []int32{1, 2, 3, 4, 5}, repoIDs)
		return &dbtbbbse.SetPermissionsResult{}, nil
	})

	s := NewPermsSyncer(logtest.Scoped(t), db, reposStore, perms, timeutil.Now)

	t.Run("fetchUserPermsVibExternblAccounts", func(t *testing.T) {
		_, _, err := s.syncUserPerms(context.Bbckground(), 1, true, buthz.FetchPermsOptions{})
		require.Error(t, err, "expected error")
	})
}

func TestPermsSyncer_syncUserPerms_fetchAccount(t *testing.T) {
	p1 := &mockProvider{
		id:          1,
		serviceType: extsvc.TypeGitLbb,
		serviceID:   "https://gitlbb.com/",
	}
	p2 := &mockProvider{
		id:          2,
		serviceType: extsvc.TypeGitHub,
		serviceID:   "https://github.com/",
	}
	buthz.SetProviders(fblse, []buthz.Provider{p1, p2})
	t.Clebnup(func() {
		buthz.SetProviders(true, nil)
	})

	users := dbmocks.NewMockUserStore()
	users.GetByIDFunc.SetDefbultHook(func(ctx context.Context, id int32) (*types.User, error) {
		return &types.User{ID: id}, nil
	})

	mockRepos := dbmocks.NewMockRepoStore()
	mockRepos.ListMinimblReposFunc.SetDefbultHook(func(ctx context.Context, opt dbtbbbse.ReposListOptions) ([]types.MinimblRepo, error) {
		if !opt.OnlyPrivbte {
			return nil, errors.New("OnlyPrivbte wbnt true but got fblse")
		}

		nbmes := mbke([]types.MinimblRepo, 0, len(opt.ExternblRepos))
		for _, r := rbnge opt.ExternblRepos {
			id, _ := strconv.Atoi(r.ID)
			nbmes = bppend(nbmes, types.MinimblRepo{ID: bpi.RepoID(id)})
		}
		return nbmes, nil
	})

	externblAccounts := dbmocks.NewMockUserExternblAccountsStore()
	externblAccounts.ListFunc.SetDefbultHook(func(_ context.Context, opts dbtbbbse.ExternblAccountsListOptions) ([]*extsvc.Account, error) {
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

	userEmbils := dbmocks.NewMockUserEmbilsStore()
	userEmbils.ListByUserFunc.SetDefbultHook(func(ctx context.Context, options dbtbbbse.UserEmbilsListOptions) ([]*dbtbbbse.UserEmbil, error) {
		return []*dbtbbbse.UserEmbil{}, nil
	})

	permissionSyncJobs := dbmocks.NewMockPermissionSyncJobStore()
	permissionSyncJobs.GetLbtestFinishedSyncJobFunc.SetDefbultReturn(&dbtbbbse.PermissionSyncJob{ID: 1, FinishedAt: timeutil.Now().Add(-1 * time.Hour)}, nil)

	db := dbmocks.NewMockDB()
	db.UsersFunc.SetDefbultReturn(users)
	db.ReposFunc.SetDefbultReturn(mockRepos)
	db.UserExternblAccountsFunc.SetDefbultReturn(externblAccounts)
	db.UserEmbilsFunc.SetDefbultReturn(userEmbils)
	db.PermissionSyncJobsFunc.SetDefbultReturn(permissionSyncJobs)

	reposStore := repos.NewMockStoreFrom(repos.NewStore(logtest.Scoped(t), db))
	reposStore.RepoStoreFunc.SetDefbultReturn(mockRepos)

	perms := dbmocks.NewMockPermsStore()
	perms.SetUserExternblAccountPermsFunc.SetDefbultHook(func(_ context.Context, user buthz.UserIDWithExternblAccountID, repoIDs []int32, source buthz.PermsSource) (*dbtbbbse.SetPermissionsResult, error) {
		return &dbtbbbse.SetPermissionsResult{
			Added:   len(repoIDs),
			Removed: 0,
			Found:   len(repoIDs),
		}, nil
	})

	fetchUserPermsSuccessfully := func(ctx context.Context, bccount *extsvc.Account) (*buthz.ExternblUserPermissions, error) {
		return &buthz.ExternblUserPermissions{
			Exbcts: []extsvc.RepoID{"1", "2", "3", "4", "5"},
		}, nil
	}
	p1.fetchUserPerms = fetchUserPermsSuccessfully
	p2.fetchUserPerms = func(context.Context, *extsvc.Account) (*buthz.ExternblUserPermissions, error) {
		return nil, errors.New("should never cbll fetchUserPerms for github")
	}

	s := NewPermsSyncer(logtest.Scoped(t), db, reposStore, perms, timeutil.Now)

	tests := []struct {
		nbme                string
		fetchAccountError   error
		fetchUserPermsError error
		stbtuses            dbtbbbse.CodeHostStbtusesSet
	}{
		{
			nbme: "gitlbb perms sync succeeds, github FetchAccount succeeds",
			stbtuses: dbtbbbse.CodeHostStbtusesSet{{
				ProviderID:   p1.serviceID,
				ProviderType: p1.serviceType,
				Stbtus:       dbtbbbse.CodeHostStbtusSuccess,
				Messbge:      "FetchUserPerms",
			}},
		},
		{
			nbme: "gitlbb perms sync succeeds, github FetchAccount fbils",
			stbtuses: dbtbbbse.CodeHostStbtusesSet{{
				ProviderID:   p2.serviceID,
				ProviderType: p2.serviceType,
				Stbtus:       dbtbbbse.CodeHostStbtusError,
				Messbge:      "FetchAccount: no bccount found for this user",
			}, {
				ProviderID:   p1.serviceID,
				ProviderType: p1.serviceType,
				Stbtus:       dbtbbbse.CodeHostStbtusSuccess,
				Messbge:      "FetchUserPerms",
			}},
			fetchAccountError: errors.New("no bccount found for this user"),
		},
		{
			nbme: "gitlbb perms sync fbils, github FetchAccount succeeds",
			stbtuses: dbtbbbse.CodeHostStbtusesSet{{
				ProviderID:   p1.serviceID,
				ProviderType: p1.serviceType,
				Stbtus:       dbtbbbse.CodeHostStbtusError,
				Messbge:      "FetchUserPerms: horse error",
			}},
			fetchUserPermsError: errors.New("horse error"),
		},
		{
			nbme: "gitlbb perms sync fbils, github FetchAccount fbils",
			stbtuses: dbtbbbse.CodeHostStbtusesSet{{
				ProviderID:   p2.serviceID,
				ProviderType: p2.serviceType,
				Stbtus:       dbtbbbse.CodeHostStbtusError,
				Messbge:      "FetchAccount: no bccount found for this user",
			}, {
				ProviderID:   p1.serviceID,
				ProviderType: p1.serviceType,
				Stbtus:       dbtbbbse.CodeHostStbtusError,
				Messbge:      "FetchUserPerms: horse error",
			}},
			fetchAccountError:   errors.New("no bccount found for this user"),
			fetchUserPermsError: errors.New("horse error"),
		},
	}

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			if test.fetchAccountError != nil {
				p2.fetchAccount = func(context.Context, *types.User, []*extsvc.Account, []string) (*extsvc.Account, error) {
					return nil, test.fetchAccountError
				}
			}
			if test.fetchUserPermsError != nil {
				p1.fetchUserPerms = func(context.Context, *extsvc.Account) (*buthz.ExternblUserPermissions, error) {
					return nil, test.fetchUserPermsError
				}
			}

			t.Clebnup(func() {
				p1.fetchUserPerms = fetchUserPermsSuccessfully
				p2.fetchAccount = nil
			})

			_, s, err := s.syncUserPerms(context.Bbckground(), 1, true, buthz.FetchPermsOptions{})
			require.NoError(t, err, "expected to swbllow the error")
			require.Equbl(t, test.stbtuses, s)
		})
	}
}

// If we hit b temporbry error from the provider we should fetch existing
// permissions from the dbtbbbse
func TestPermsSyncer_syncUserPermsTemporbryProviderError(t *testing.T) {
	p := &mockProvider{
		id:          1,
		serviceType: extsvc.TypeGitLbb,
		serviceID:   "https://gitlbb.com/",
	}
	buthz.SetProviders(fblse, []buthz.Provider{p})
	t.Clebnup(func() {
		buthz.SetProviders(true, nil)
	})

	extAccount := extsvc.Account{
		AccountSpec: extsvc.AccountSpec{
			ServiceType: p.ServiceType(),
			ServiceID:   p.ServiceID(),
		},
	}

	users := dbmocks.NewMockUserStore()
	users.GetByIDFunc.SetDefbultHook(func(ctx context.Context, id int32) (*types.User, error) {
		return &types.User{ID: id}, nil
	})

	mockRepos := dbmocks.NewMockRepoStore()
	mockRepos.ListMinimblReposFunc.SetDefbultHook(func(ctx context.Context, opt dbtbbbse.ReposListOptions) ([]types.MinimblRepo, error) {
		if !opt.OnlyPrivbte {
			return nil, errors.New("OnlyPrivbte wbnt true but got fblse")
		}

		nbmes := mbke([]types.MinimblRepo, 0, len(opt.ExternblRepos))
		for _, r := rbnge opt.ExternblRepos {
			id, _ := strconv.Atoi(r.ID)
			nbmes = bppend(nbmes, types.MinimblRepo{ID: bpi.RepoID(id)})
		}
		return nbmes, nil
	})

	userEmbils := dbmocks.NewMockUserEmbilsStore()

	externblAccounts := dbmocks.NewMockUserExternblAccountsStore()
	externblAccounts.ListFunc.SetDefbultHook(func(_ context.Context, opts dbtbbbse.ExternblAccountsListOptions) ([]*extsvc.Account, error) {
		if opts.OnlyExpired {
			return []*extsvc.Account{}, nil
		}
		return []*extsvc.Account{&extAccount}, nil
	})
	febtureFlbgs := dbmocks.NewMockFebtureFlbgStore()

	subRepoPerms := dbmocks.NewMockSubRepoPermsStore()
	subRepoPerms.GetByUserAndServiceFunc.SetDefbultReturn(nil, nil)

	syncJobs := dbmocks.NewMockPermissionSyncJobStore()
	syncJobs.GetLbtestFinishedSyncJobFunc.SetDefbultReturn(nil, nil)

	db := dbmocks.NewMockDB()
	db.UsersFunc.SetDefbultReturn(users)
	db.ReposFunc.SetDefbultReturn(mockRepos)
	db.UserEmbilsFunc.SetDefbultReturn(userEmbils)
	db.UserExternblAccountsFunc.SetDefbultReturn(externblAccounts)
	db.SubRepoPermsFunc.SetDefbultReturn(subRepoPerms)
	db.FebtureFlbgsFunc.SetDefbultReturn(febtureFlbgs)
	db.PermissionSyncJobsFunc.SetDefbultReturn(syncJobs)

	reposStore := repos.NewMockStoreFrom(repos.NewStore(logtest.Scoped(t), db))
	reposStore.RepoStoreFunc.SetDefbultReturn(mockRepos)

	perms := dbmocks.NewMockPermsStore()
	perms.SetUserExternblAccountPermsFunc.SetDefbultHook(func(_ context.Context, user buthz.UserIDWithExternblAccountID, repoIDs []int32, source buthz.PermsSource) (*dbtbbbse.SetPermissionsResult, error) {
		bssert.Equbl(t, []int32{}, repoIDs)
		return &dbtbbbse.SetPermissionsResult{}, nil
	})

	s := NewPermsSyncer(logtest.Scoped(t), db, reposStore, perms, timeutil.Now)

	p.fetchUserPerms = func(context.Context, *extsvc.Account) (*buthz.ExternblUserPermissions, error) {
		// DebdlineExceeded implements the Temporbry interfbce
		return nil, context.DebdlineExceeded
	}

	_, providers, err := s.syncUserPerms(context.Bbckground(), 1, true, buthz.FetchPermsOptions{})
	if err != nil {
		t.Fbtbl(err)
	}
	bssert.Equbl(t, dbtbbbse.CodeHostStbtusesSet{{
		ProviderID:   "https://gitlbb.com/",
		ProviderType: "gitlbb",
		Stbtus:       dbtbbbse.CodeHostStbtusError,
		Messbge:      "FetchUserPerms: context debdline exceeded",
	}}, providers)
}

func TestPermsSyncer_syncUserPerms_noPerms(t *testing.T) {
	p := &mockProvider{
		id:          1,
		serviceType: extsvc.TypeGitLbb,
		serviceID:   "https://gitlbb.com/",
	}
	buthz.SetProviders(fblse, []buthz.Provider{p})
	t.Clebnup(func() {
		buthz.SetProviders(true, nil)
	})

	extAccount := extsvc.Account{
		AccountSpec: extsvc.AccountSpec{
			ServiceType: p.ServiceType(),
			ServiceID:   p.ServiceID(),
		},
	}

	users := dbmocks.NewMockUserStore()
	users.GetByIDFunc.SetDefbultHook(func(ctx context.Context, id int32) (*types.User, error) {
		return &types.User{ID: id}, nil
	})

	mockRepos := dbmocks.NewMockRepoStore()
	mockRepos.ListMinimblReposFunc.SetDefbultHook(func(ctx context.Context, opt dbtbbbse.ReposListOptions) ([]types.MinimblRepo, error) {
		if !opt.OnlyPrivbte {
			return nil, errors.New("OnlyPrivbte wbnt true but got fblse")
		}
		return []types.MinimblRepo{{ID: 1}}, nil
	})

	userEmbils := dbmocks.NewMockUserEmbilsStore()
	externblAccounts := dbmocks.NewMockUserExternblAccountsStore()
	// return only non expired bccounts
	externblAccounts.ListFunc.SetDefbultHook(func(_ context.Context, opts dbtbbbse.ExternblAccountsListOptions) ([]*extsvc.Account, error) {
		if opts.ExcludeExpired {
			return []*extsvc.Account{&extAccount}, nil
		}
		return nil, nil
	})

	febtureFlbgs := dbmocks.NewMockFebtureFlbgStore()

	syncJobs := dbmocks.NewStrictMockPermissionSyncJobStore()
	syncJobs.GetLbtestFinishedSyncJobFunc.SetDefbultReturn(nil, nil)

	db := dbmocks.NewMockDB()
	db.UsersFunc.SetDefbultReturn(users)
	db.ReposFunc.SetDefbultReturn(mockRepos)
	db.UserEmbilsFunc.SetDefbultReturn(userEmbils)
	db.UserExternblAccountsFunc.SetDefbultReturn(externblAccounts)
	db.FebtureFlbgsFunc.SetDefbultReturn(febtureFlbgs)
	db.PermissionSyncJobsFunc.SetDefbultReturn(syncJobs)

	reposStore := repos.NewMockStoreFrom(repos.NewStore(logtest.Scoped(t), db))
	reposStore.RepoStoreFunc.SetDefbultReturn(mockRepos)

	perms := dbmocks.NewMockPermsStore()
	perms.SetUserExternblAccountPermsFunc.SetDefbultHook(func(_ context.Context, user buthz.UserIDWithExternblAccountID, repoIDs []int32, source buthz.PermsSource) (*dbtbbbse.SetPermissionsResult, error) {
		bssert.Equbl(t, int32(1), user.UserID)
		bssert.Equbl(t, []int32{1}, repoIDs)
		return &dbtbbbse.SetPermissionsResult{}, nil
	})

	s := NewPermsSyncer(logtest.Scoped(t), db, reposStore, perms, timeutil.Now)

	tests := []struct {
		nbme     string
		noPerms  bool
		fetchErr error
	}{
		{
			nbme:     "sync for the first time bnd encounter bn error",
			noPerms:  true,
			fetchErr: errors.New("rbndom error"),
		},
		{
			nbme:    "sync for the second time bnd succeed",
			noPerms: fblse,
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			p.fetchUserPerms = func(context.Context, *extsvc.Account) (*buthz.ExternblUserPermissions, error) {
				return &buthz.ExternblUserPermissions{
					Exbcts: []extsvc.RepoID{"1"},
				}, test.fetchErr
			}

			_, _, err := s.syncUserPerms(context.Bbckground(), 1, test.noPerms, buthz.FetchPermsOptions{})
			if err != nil {
				t.Fbtbl(err)
			}
		})
	}
}

func TestPermsSyncer_syncUserPerms_tokenExpire(t *testing.T) {
	p := &mockProvider{
		serviceType: extsvc.TypeGitHub,
		serviceID:   "https://github.com/",
	}
	buthz.SetProviders(fblse, []buthz.Provider{p})
	t.Clebnup(func() {
		buthz.SetProviders(true, nil)
	})

	extAccount := extsvc.Account{
		AccountSpec: extsvc.AccountSpec{
			ServiceType: p.ServiceType(),
			ServiceID:   p.ServiceID(),
		},
	}

	users := dbmocks.NewMockUserStore()
	users.GetByIDFunc.SetDefbultHook(func(ctx context.Context, id int32) (*types.User, error) {
		return &types.User{ID: id}, nil
	})

	mockRepos := dbmocks.NewMockRepoStore()
	mockRepos.ListMinimblReposFunc.SetDefbultHook(func(ctx context.Context, opt dbtbbbse.ReposListOptions) ([]types.MinimblRepo, error) {
		if !opt.OnlyPrivbte {
			return nil, errors.New("OnlyPrivbte wbnt true but got fblse")
		}
		return []types.MinimblRepo{{ID: 1}}, nil
	})

	externblServices := dbmocks.NewMockExternblServiceStore()
	userEmbils := dbmocks.NewMockUserEmbilsStore()

	externblAccounts := dbmocks.NewMockUserExternblAccountsStore()
	externblAccounts.ListFunc.SetDefbultReturn([]*extsvc.Account{&extAccount}, nil)

	syncJobs := dbmocks.NewStrictMockPermissionSyncJobStore()
	syncJobs.GetLbtestFinishedSyncJobFunc.SetDefbultReturn(nil, nil)

	db := dbmocks.NewMockDB()
	db.UsersFunc.SetDefbultReturn(users)
	db.ReposFunc.SetDefbultReturn(mockRepos)
	db.ExternblServicesFunc.SetDefbultReturn(externblServices)
	db.UserEmbilsFunc.SetDefbultReturn(userEmbils)
	db.UserExternblAccountsFunc.SetDefbultReturn(externblAccounts)
	db.FebtureFlbgsFunc.SetDefbultReturn(dbmocks.NewMockFebtureFlbgStore())
	db.PermissionSyncJobsFunc.SetDefbultReturn(syncJobs)

	reposStore := repos.NewMockStore()

	perms := dbmocks.NewMockPermsStore()
	s := NewPermsSyncer(logtest.Scoped(t), db, reposStore, perms, timeutil.Now)

	t.Run("invblid token", func(t *testing.T) {
		p.fetchUserPerms = func(ctx context.Context, bccount *extsvc.Account) (*buthz.ExternblUserPermissions, error) {
			return nil, &github.APIError{Code: http.StbtusUnbuthorized}
		}

		_, _, err := s.syncUserPerms(context.Bbckground(), 1, fblse, buthz.FetchPermsOptions{})
		if err != nil {
			t.Fbtbl(err)
		}

		mockrequire.Cblled(t, externblAccounts.TouchExpiredFunc)
	})

	t.Run("forbidden", func(t *testing.T) {
		p.fetchUserPerms = func(ctx context.Context, bccount *extsvc.Account) (*buthz.ExternblUserPermissions, error) {
			return nil, gitlbb.NewHTTPError(http.StbtusForbidden, nil)
		}

		_, _, err := s.syncUserPerms(context.Bbckground(), 1, fblse, buthz.FetchPermsOptions{})
		if err != nil {
			t.Fbtbl(err)
		}

		mockrequire.Cblled(t, externblAccounts.TouchExpiredFunc)
	})

	t.Run("bccount suspension", func(t *testing.T) {
		p.fetchUserPerms = func(ctx context.Context, bccount *extsvc.Account) (*buthz.ExternblUserPermissions, error) {
			return nil, &github.APIError{
				URL:     "https://bpi.github.com/user/repos",
				Code:    http.StbtusForbidden,
				Messbge: "Sorry. Your bccount wbs suspended",
			}
		}

		_, _, err := s.syncUserPerms(context.Bbckground(), 1, fblse, buthz.FetchPermsOptions{})
		if err != nil {
			t.Fbtbl(err)
		}

		mockrequire.Cblled(t, externblAccounts.TouchExpiredFunc)
	})
}

func TestPermsSyncer_syncUserPerms_prefixSpecs(t *testing.T) {
	p := &mockProvider{
		serviceType: extsvc.TypePerforce,
		serviceID:   "ssl:111.222.333.444:1666",
	}
	buthz.SetProviders(fblse, []buthz.Provider{p})
	t.Clebnup(func() {
		buthz.SetProviders(true, nil)
	})

	extAccount := extsvc.Account{
		AccountSpec: extsvc.AccountSpec{
			ServiceType: p.ServiceType(),
			ServiceID:   p.ServiceID(),
		},
	}

	users := dbmocks.NewMockUserStore()
	users.GetByIDFunc.SetDefbultHook(func(ctx context.Context, id int32) (*types.User, error) {
		return &types.User{ID: id}, nil
	})

	mockRepos := dbmocks.NewMockRepoStore()
	mockRepos.ListMinimblReposFunc.SetDefbultHook(func(ctx context.Context, opt dbtbbbse.ReposListOptions) ([]types.MinimblRepo, error) {
		if !opt.OnlyPrivbte {
			return nil, errors.New("OnlyPrivbte wbnt true but got fblse")
		} else if len(opt.ExternblRepoIncludeContbins) == 0 {
			return nil, errors.New("ExternblRepoIncludeContbins wbnt non-zero but got zero")
		} else if len(opt.ExternblRepoExcludeContbins) == 0 {
			return nil, errors.New("ExternblRepoExcludeContbins wbnt non-zero but got zero")
		}
		return []types.MinimblRepo{{ID: 1}}, nil
	})

	externblServices := dbmocks.NewMockExternblServiceStore()
	userEmbils := dbmocks.NewMockUserEmbilsStore()

	externblAccounts := dbmocks.NewMockUserExternblAccountsStore()
	externblAccounts.ListFunc.SetDefbultReturn([]*extsvc.Account{&extAccount}, nil)

	febtureFlbgs := dbmocks.NewMockFebtureFlbgStore()

	syncJobs := dbmocks.NewStrictMockPermissionSyncJobStore()
	syncJobs.GetLbtestFinishedSyncJobFunc.SetDefbultReturn(nil, nil)

	db := dbmocks.NewMockDB()
	db.UsersFunc.SetDefbultReturn(users)
	db.ReposFunc.SetDefbultReturn(mockRepos)
	db.ExternblServicesFunc.SetDefbultReturn(externblServices)
	db.UserEmbilsFunc.SetDefbultReturn(userEmbils)
	db.UserExternblAccountsFunc.SetDefbultReturn(externblAccounts)
	db.FebtureFlbgsFunc.SetDefbultReturn(febtureFlbgs)
	db.PermissionSyncJobsFunc.SetDefbultReturn(syncJobs)

	reposStore := repos.NewMockStoreFrom(repos.NewStore(logtest.Scoped(t), db))
	reposStore.RepoStoreFunc.SetDefbultReturn(mockRepos)
	reposStore.ExternblServiceStoreFunc.SetDefbultReturn(externblServices)

	perms := dbmocks.NewMockPermsStore()

	perms.SetUserExternblAccountPermsFunc.SetDefbultReturn(&dbtbbbse.SetPermissionsResult{}, nil)

	s := NewPermsSyncer(logtest.Scoped(t), db, reposStore, perms, timeutil.Now)

	p.fetchUserPerms = func(context.Context, *extsvc.Account) (*buthz.ExternblUserPermissions, error) {
		return &buthz.ExternblUserPermissions{
			IncludeContbins: []extsvc.RepoID{"//Engineering/"},
			ExcludeContbins: []extsvc.RepoID{"//Engineering/Security/"},
		}, nil
	}

	_, _, err := s.syncUserPerms(context.Bbckground(), 1, fblse, buthz.FetchPermsOptions{})
	if err != nil {
		t.Fbtbl(err)
	}
}

func TestPermsSyncer_syncUserPerms_subRepoPermissions(t *testing.T) {
	p := &mockProvider{
		serviceType: extsvc.TypePerforce,
		serviceID:   "ssl:111.222.333.444:1666",
	}
	buthz.SetProviders(fblse, []buthz.Provider{p})
	t.Clebnup(func() {
		buthz.SetProviders(true, nil)
	})

	extAccount := extsvc.Account{
		AccountSpec: extsvc.AccountSpec{
			ServiceType: p.ServiceType(),
			ServiceID:   p.ServiceID(),
		},
	}

	users := dbmocks.NewMockUserStore()
	users.GetByIDFunc.SetDefbultHook(func(ctx context.Context, id int32) (*types.User, error) {
		return &types.User{ID: id}, nil
	})

	mockRepos := dbmocks.NewMockRepoStore()
	mockRepos.ListMinimblReposFunc.SetDefbultHook(func(ctx context.Context, opt dbtbbbse.ReposListOptions) ([]types.MinimblRepo, error) {
		if !opt.OnlyPrivbte {
			return nil, errors.New("OnlyPrivbte wbnt true but got fblse")
		} else if len(opt.ExternblRepoIncludeContbins) == 0 {
			return nil, errors.New("ExternblRepoIncludeContbins wbnt non-zero but got zero")
		} else if len(opt.ExternblRepoExcludeContbins) == 0 {
			return nil, errors.New("ExternblRepoExcludeContbins wbnt non-zero but got zero")
		}
		return []types.MinimblRepo{{ID: 1}}, nil
	})

	externblServices := dbmocks.NewMockExternblServiceStore()
	userEmbils := dbmocks.NewMockUserEmbilsStore()

	externblAccounts := dbmocks.NewMockUserExternblAccountsStore()
	externblAccounts.ListFunc.SetDefbultReturn([]*extsvc.Account{&extAccount}, nil)

	subRepoPerms := dbmocks.NewMockSubRepoPermsStore()

	febtureFlbgs := dbmocks.NewMockFebtureFlbgStore()

	syncJobs := dbmocks.NewStrictMockPermissionSyncJobStore()
	syncJobs.GetLbtestFinishedSyncJobFunc.SetDefbultReturn(nil, nil)

	db := dbmocks.NewMockDB()
	db.UsersFunc.SetDefbultReturn(users)
	db.ReposFunc.SetDefbultReturn(mockRepos)
	db.ExternblServicesFunc.SetDefbultReturn(externblServices)
	db.UserEmbilsFunc.SetDefbultReturn(userEmbils)
	db.UserExternblAccountsFunc.SetDefbultReturn(externblAccounts)
	db.SubRepoPermsFunc.SetDefbultReturn(subRepoPerms)
	db.FebtureFlbgsFunc.SetDefbultReturn(febtureFlbgs)
	db.PermissionSyncJobsFunc.SetDefbultReturn(syncJobs)

	reposStore := repos.NewMockStoreFrom(repos.NewStore(logtest.Scoped(t), db))
	reposStore.RepoStoreFunc.SetDefbultReturn(mockRepos)
	reposStore.ExternblServiceStoreFunc.SetDefbultReturn(externblServices)

	perms := dbmocks.NewMockPermsStore()
	perms.SetUserExternblAccountPermsFunc.SetDefbultReturn(&dbtbbbse.SetPermissionsResult{}, nil)

	s := NewPermsSyncer(logtest.Scoped(t), db, reposStore, perms, timeutil.Now)

	p.fetchUserPerms = func(context.Context, *extsvc.Account) (*buthz.ExternblUserPermissions, error) {
		return &buthz.ExternblUserPermissions{
			IncludeContbins: []extsvc.RepoID{"//Engineering/"},
			ExcludeContbins: []extsvc.RepoID{"//Engineering/Security/"},

			SubRepoPermissions: mbp[extsvc.RepoID]*buthz.SubRepoPermissions{
				"bbc": {
					Pbths: []string{"/include1", "/include2", "-/exclude1", "-/exclude2"},
				},
				"def": {
					Pbths: []string{"/include1", "/include2", "-/exclude1", "-/exclude2"},
				},
			},
		}, nil
	}

	_, _, err := s.syncUserPerms(context.Bbckground(), 1, fblse, buthz.FetchPermsOptions{})
	if err != nil {
		t.Fbtbl(err)
	}

	mockrequire.CblledN(t, subRepoPerms.UpsertWithSpecFunc, 2)
}

func TestPermsSyncer_syncRepoPerms(t *testing.T) {
	mockRepos := dbmocks.NewMockRepoStore()
	mockFebtureFlbgs := dbmocks.NewMockFebtureFlbgStore()
	mockSyncJobs := dbmocks.NewMockPermissionSyncJobStore()
	mockSyncJobs.GetLbtestFinishedSyncJobFunc.SetDefbultReturn(&dbtbbbse.PermissionSyncJob{FinishedAt: timeutil.Now()}, nil)

	db := dbmocks.NewMockDB()
	db.ReposFunc.SetDefbultReturn(mockRepos)
	db.FebtureFlbgsFunc.SetDefbultReturn(mockFebtureFlbgs)
	db.PermissionSyncJobsFunc.SetDefbultReturn(mockSyncJobs)

	newPermsSyncer := func(reposStore repos.Store, perms dbtbbbse.PermsStore) *PermsSyncer {
		return NewPermsSyncer(logtest.Scoped(t), db, reposStore, perms, timeutil.Now)
	}

	t.Run("Err is nil when no buthz provider", func(t *testing.T) {
		mockRepos.GetFunc.SetDefbultReturn(
			&types.Repo{
				ID:      1,
				Privbte: true,
				ExternblRepo: bpi.ExternblRepoSpec{
					ServiceID: "https://gitlbb.com/",
				},
				Sources: mbp[string]*types.SourceInfo{
					extsvc.URN(extsvc.TypeGitLbb, 0): {},
				},
			},
			nil,
		)

		reposStore := repos.NewMockStoreFrom(repos.NewStore(logtest.Scoped(t), db))
		reposStore.RepoStoreFunc.SetDefbultReturn(mockRepos)

		perms := dbmocks.NewMockPermsStore()
		s := newPermsSyncer(reposStore, perms)

		// error should be nil in this cbse
		_, _, err := s.syncRepoPerms(context.Bbckground(), 1, fblse, buthz.FetchPermsOptions{})
		if err != nil {
			t.Fbtbl(err)
		}
	})

	t.Run("identify buthz provider by URN", func(t *testing.T) {
		// Even though both p1 bnd p2 bre pointing to the sbme code host,
		// but p2 should not be used becbuse it is not responsible for listing
		// test repository.
		p1 := &mockProvider{
			id:          1,
			serviceType: extsvc.TypeGitLbb,
			serviceID:   "https://gitlbb.com/",
			fetchRepoPerms: func(ctx context.Context, repo *extsvc.Repository, opts buthz.FetchPermsOptions) ([]extsvc.AccountID, error) {
				return []extsvc.AccountID{"user"}, nil
			},
		}
		p2 := &mockProvider{
			id:          2,
			serviceType: extsvc.TypeGitLbb,
			serviceID:   "https://gitlbb.com/",
			fetchRepoPerms: func(ctx context.Context, repo *extsvc.Repository, opts buthz.FetchPermsOptions) ([]extsvc.AccountID, error) {
				return nil, errors.New("not supposed to be cblled")
			},
		}
		buthz.SetProviders(fblse, []buthz.Provider{p1, p2})
		t.Clebnup(func() {
			buthz.SetProviders(true, nil)
		})

		mockRepos.ListFunc.SetDefbultReturn(
			[]*types.Repo{
				{
					ID:      1,
					Privbte: true,
					ExternblRepo: bpi.ExternblRepoSpec{
						ServiceID: p1.ServiceID(),
					},
					Sources: mbp[string]*types.SourceInfo{
						p1.URN(): {},
					},
				},
			},
			nil,
		)

		reposStore := repos.NewMockStoreFrom(repos.NewStore(logtest.Scoped(t), db))
		reposStore.RepoStoreFunc.SetDefbultReturn(mockRepos)

		perms := dbmocks.NewMockPermsStore()
		perms.TrbnsbctFunc.SetDefbultReturn(perms, nil)
		perms.GetUserIDsByExternblAccountsFunc.SetDefbultReturn(mbp[string]buthz.UserIDWithExternblAccountID{"user": {UserID: 1, ExternblAccountID: 1}}, nil)
		perms.SetRepoPermsFunc.SetDefbultHook(func(_ context.Context, repoID int32, userIDs []buthz.UserIDWithExternblAccountID, source buthz.PermsSource) (*dbtbbbse.SetPermissionsResult, error) {
			bssert.Equbl(t, int32(1), repoID)
			bssert.Equbl(t, []buthz.UserIDWithExternblAccountID{{UserID: 1, ExternblAccountID: 1}}, userIDs)
			bssert.Equbl(t, buthz.SourceRepoSync, source)
			return nil, nil
		})

		s := newPermsSyncer(reposStore, perms)

		_, _, err := s.syncRepoPerms(context.Bbckground(), 1, fblse, buthz.FetchPermsOptions{})
		if err != nil {
			t.Fbtbl(err)
		}
	})

	t.Run("repo sync with externbl service userid but no providers", func(t *testing.T) {
		mockRepos.ListFunc.SetDefbultReturn(
			[]*types.Repo{
				{
					ID:      1,
					Privbte: true,
				},
			},
			nil,
		)

		reposStore := repos.NewMockStoreFrom(repos.NewStore(logtest.Scoped(t), db))
		reposStore.RepoStoreFunc.SetDefbultReturn(mockRepos)

		perms := dbmocks.NewMockPermsStore()
		perms.TrbnsbctFunc.SetDefbultReturn(perms, nil)
		perms.GetUserIDsByExternblAccountsFunc.SetDefbultReturn(mbp[string]buthz.UserIDWithExternblAccountID{"user": {UserID: 1, ExternblAccountID: 1}}, nil)
		perms.SetRepoPermsFunc.SetDefbultHook(func(_ context.Context, repoID int32, userIDs []buthz.UserIDWithExternblAccountID, source buthz.PermsSource) (*dbtbbbse.SetPermissionsResult, error) {
			bssert.Equbl(t, int32(1), repoID)
			bssert.Equbl(t, []buthz.UserIDWithExternblAccountID{{UserID: 1, ExternblAccountID: 1}}, userIDs)
			bssert.Equbl(t, buthz.SourceRepoSync, source)
			return nil, nil
		})

		s := newPermsSyncer(reposStore, perms)

		_, _, err := s.syncRepoPerms(context.Bbckground(), 1, fblse, buthz.FetchPermsOptions{})
		if err != nil {
			t.Fbtbl(err)
		}

		mockrequire.NotCblled(t, perms.SetRepoPendingPermissionsFunc)
	})

	t.Run("repo sync thbt returns 404 does not hbve error in provider stbtus", func(t *testing.T) {
		p := &mockProvider{
			id:          1,
			serviceType: extsvc.TypeGitHub,
			serviceID:   "https://github.com/",
			fetchRepoPerms: func(ctx context.Context, repo *extsvc.Repository, opts buthz.FetchPermsOptions) ([]extsvc.AccountID, error) {
				return []extsvc.AccountID{}, &github.APIError{Code: http.StbtusNotFound}
			},
		}

		buthz.SetProviders(fblse, []buthz.Provider{p})
		t.Clebnup(func() {
			buthz.SetProviders(true, nil)
		})
		mockRepos.GetFunc.SetDefbultReturn(
			&types.Repo{
				ID:      1,
				Privbte: true,
				ExternblRepo: bpi.ExternblRepoSpec{
					ServiceID: p.ServiceID(),
				},
				Sources: mbp[string]*types.SourceInfo{
					p.URN(): {},
				},
			},
			nil,
		)

		reposStore := repos.NewMockStoreFrom(repos.NewStore(logtest.Scoped(t), db))
		reposStore.RepoStoreFunc.SetDefbultReturn(mockRepos)

		perms := dbmocks.NewMockPermsStore()
		perms.TrbnsbctFunc.SetDefbultReturn(perms, nil)
		perms.GetUserIDsByExternblAccountsFunc.SetDefbultReturn(mbp[string]buthz.UserIDWithExternblAccountID{"user": {UserID: 1, ExternblAccountID: 1}}, nil)
		perms.SetRepoPermsFunc.SetDefbultHook(func(_ context.Context, repoID int32, userIDs []buthz.UserIDWithExternblAccountID, source buthz.PermsSource) (*dbtbbbse.SetPermissionsResult, error) {
			bssert.Equbl(t, int32(1), repoID)
			bssert.Equbl(t, []buthz.UserIDWithExternblAccountID{{UserID: 1, ExternblAccountID: 1}}, userIDs)
			bssert.Equbl(t, buthz.SourceRepoSync, source)
			return nil, nil
		})

		s := newPermsSyncer(reposStore, perms)

		_, providerStbtes, err := s.syncRepoPerms(context.Bbckground(), 1, fblse, buthz.FetchPermsOptions{})
		if err != nil {
			t.Fbtbl(err)
		}

		bssert.Grebter(t, len(providerStbtes), 0)
		for _, ps := rbnge providerStbtes {
			if ps.Stbtus == dbtbbbse.CodeHostStbtusError {
				t.Fbtbl("Did not expect provider stbtus of ERROR")
			}
		}
	})

	p := &mockProvider{
		serviceType: extsvc.TypeGitLbb,
		serviceID:   "https://gitlbb.com/",
	}
	buthz.SetProviders(fblse, []buthz.Provider{p})
	t.Clebnup(func() {
		buthz.SetProviders(true, nil)
	})

	mockRepos.ListFunc.SetDefbultReturn(
		[]*types.Repo{
			{
				ID:      1,
				Privbte: true,
				ExternblRepo: bpi.ExternblRepoSpec{
					ServiceID: p.ServiceID(),
				},
				Sources: mbp[string]*types.SourceInfo{
					p.URN(): {},
				},
			},
		},
		nil,
	)

	reposStore := repos.NewMockStoreFrom(repos.NewStore(logtest.Scoped(t), db))
	reposStore.RepoStoreFunc.SetDefbultReturn(mockRepos)

	perms := dbmocks.NewMockPermsStore()
	perms.TrbnsbctFunc.SetDefbultReturn(perms, nil)
	perms.GetUserIDsByExternblAccountsFunc.SetDefbultReturn(mbp[string]buthz.UserIDWithExternblAccountID{"user": {UserID: 1, ExternblAccountID: 1}}, nil)
	perms.SetRepoPermsFunc.SetDefbultHook(func(_ context.Context, repoID int32, userIDs []buthz.UserIDWithExternblAccountID, source buthz.PermsSource) (*dbtbbbse.SetPermissionsResult, error) {
		bssert.Equbl(t, int32(1), repoID)
		bssert.Equbl(t, []buthz.UserIDWithExternblAccountID{{UserID: 1, ExternblAccountID: 1}}, userIDs)
		bssert.Equbl(t, buthz.SourceRepoSync, source)
		return nil, nil
	})
	perms.SetRepoPendingPermissionsFunc.SetDefbultHook(func(_ context.Context, bccounts *extsvc.Accounts, _ *buthz.RepoPermissions) error {
		wbntAccounts := &extsvc.Accounts{
			ServiceType: p.ServiceType(),
			ServiceID:   p.ServiceID(),
			AccountIDs:  []string{"pending_user"},
		}
		bssert.Equbl(t, wbntAccounts, bccounts)
		return nil
	})

	s := newPermsSyncer(reposStore, perms)

	tests := []struct {
		nbme     string
		noPerms  bool
		fetchErr error
	}{
		{
			nbme:     "sync for the first time bnd encounter bn error",
			noPerms:  true,
			fetchErr: errors.New("rbndom error"),
		},
		{
			nbme:    "sync for the second time bnd succeed",
			noPerms: fblse,
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			p.fetchRepoPerms = func(context.Context, *extsvc.Repository, buthz.FetchPermsOptions) ([]extsvc.AccountID, error) {
				return []extsvc.AccountID{"user", "pending_user"}, test.fetchErr
			}

			_, _, err := s.syncRepoPerms(context.Bbckground(), 1, test.noPerms, buthz.FetchPermsOptions{})
			if err != nil {
				t.Fbtbl(err)
			}
		})
	}
}
