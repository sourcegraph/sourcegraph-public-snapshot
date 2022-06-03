package repos

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/keegancsmith/sqlf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/log/logtest"
)

func TestStatusMessages(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := context.Background()
	db := database.NewDB(dbtest.NewDB(t))
	store := NewStore(logtest.Scoped(t), database.NewDB(db), sql.TxOptions{})

	admin, err := db.Users().Create(ctx, database.NewUser{
		Email:                 "a1@example.com",
		Username:              "a1",
		Password:              "p",
		EmailVerificationCode: "c",
	})
	require.NoError(t, err)

	nonAdmin, err := db.Users().Create(ctx, database.NewUser{
		Email:                 "u1@example.com",
		Username:              "u1",
		Password:              "p",
		EmailVerificationCode: "c",
	})
	require.NoError(t, err)

	siteLevelService := &types.ExternalService{
		ID:          1,
		Config:      `{}`,
		Kind:        extsvc.KindGitHub,
		DisplayName: "github.com - site",
	}
	err = db.ExternalServices().Upsert(ctx, siteLevelService)
	require.NoError(t, err)

	userService := &types.ExternalService{
		ID:              2,
		Config:          `{}`,
		Kind:            extsvc.KindGitHub,
		DisplayName:     "github.com - user",
		NamespaceUserID: nonAdmin.ID,
	}
	err = db.ExternalServices().Upsert(ctx, userService)
	require.NoError(t, err)

	testCases := []struct {
		name             string
		stored           types.Repos
		gitserverCloned  []string
		gitserverFailure map[string]bool
		sourcerErr       error
		res              []StatusMessage
		user             *types.User
		// maps repoName to external service
		repoOwner map[api.RepoName]*types.ExternalService
		err       string
	}{
		{
			name:            "all cloned",
			gitserverCloned: []string{"foobar"},
			stored:          []*types.Repo{{Name: "foobar"}},
			user:            admin,
			res:             nil,
		},
		{
			name:            "nothing cloned",
			stored:          []*types.Repo{{Name: "foobar"}},
			user:            admin,
			gitserverCloned: []string{},
			repoOwner: map[api.RepoName]*types.ExternalService{
				"foobar": siteLevelService,
			},
			res: []StatusMessage{
				{
					Cloning: &CloningProgress{
						Message: "Some repositories cloning...",
					},
				},
			},
		},
		{
			name:            "subset cloned",
			stored:          []*types.Repo{{Name: "foobar"}, {Name: "barfoo"}},
			user:            admin,
			gitserverCloned: []string{"foobar"},
			repoOwner: map[api.RepoName]*types.ExternalService{
				"foobar": siteLevelService,
				"barfoo": siteLevelService,
			},
			res: []StatusMessage{
				{
					Cloning: &CloningProgress{
						Message: "Some repositories cloning...",
					},
				},
			},
		},
		{
			name:   "non admin users should only count their own non cloned repos",
			stored: []*types.Repo{{Name: "foobar"}, {Name: "barfoo"}},
			repoOwner: map[api.RepoName]*types.ExternalService{
				"foobar": userService,
			},
			user: nonAdmin,
			res: []StatusMessage{
				{
					Cloning: &CloningProgress{
						Message: "Some repositories cloning...",
					},
				},
			},
		},
		{
			name:            "more cloned than stored",
			stored:          []*types.Repo{{Name: "foobar"}},
			user:            admin,
			gitserverCloned: []string{"foobar", "barfoo"},
			res:             nil,
		},
		{
			name:            "cloned different than stored",
			stored:          []*types.Repo{{Name: "foobar"}, {Name: "barfoo"}},
			user:            admin,
			gitserverCloned: []string{"one", "two", "three"},
			repoOwner: map[api.RepoName]*types.ExternalService{
				"foobar": siteLevelService,
				"barfoo": siteLevelService,
			},
			res: []StatusMessage{
				{
					Cloning: &CloningProgress{
						Message: "Some repositories cloning...",
					},
				},
			},
		},
		{
			name:             "one repo failed to sync",
			stored:           []*types.Repo{{Name: "foobar"}, {Name: "barfoo"}},
			user:             admin,
			gitserverCloned:  []string{"foobar", "barfoo"},
			gitserverFailure: map[string]bool{"foobar": true},
			repoOwner: map[api.RepoName]*types.ExternalService{
				"foobar": siteLevelService,
				"barfoo": siteLevelService,
			},
			res: []StatusMessage{
				{
					SyncError: &SyncError{
						Message: "Some repositories could not be synced",
					},
				},
			},
		},
		{
			name:             "two repos failed to sync",
			stored:           []*types.Repo{{Name: "foobar"}, {Name: "barfoo"}},
			user:             admin,
			gitserverCloned:  []string{"foobar", "barfoo"},
			gitserverFailure: map[string]bool{"foobar": true, "barfoo": true},
			repoOwner: map[api.RepoName]*types.ExternalService{
				"foobar": siteLevelService,
				"barfoo": siteLevelService,
			},
			res: []StatusMessage{
				{
					SyncError: &SyncError{
						Message: "Some repositories could not be synced",
					},
				},
			},
		},
		{
			name:            "case insensitivity",
			gitserverCloned: []string{"foobar"},
			stored:          []*types.Repo{{Name: "FOOBar"}},
			user:            admin,
			res:             nil,
		},
		{
			name:            "case insensitivity to gitserver names",
			gitserverCloned: []string{"FOOBar"},
			stored:          []*types.Repo{{Name: "FOOBar"}},
			user:            admin,
			res:             nil,
		},
		{
			name:       "one external service syncer err",
			user:       admin,
			sourcerErr: errors.New("github is down"),
			res: []StatusMessage{
				{
					ExternalServiceSyncError: &ExternalServiceSyncError{
						Message:           "github is down",
						ExternalServiceId: siteLevelService.ID,
					},
				},
			},
		},
	}

	logger := logtest.Scoped(t)
	for _, tc := range testCases {
		tc := tc
		ctx := context.Background()

		t.Run(tc.name, func(t *testing.T) {
			stored := tc.stored.Clone()
			for _, r := range stored {
				r.ExternalRepo = api.ExternalRepoSpec{
					ID:          uuid.New().String(),
					ServiceType: extsvc.TypeGitHub,
					ServiceID:   "https://github.com/",
				}
			}

			err := db.Repos().Create(ctx, stored...)
			require.NoError(t, err)

			t.Cleanup(func() {
				ids := make([]api.RepoID, 0, len(stored))
				for _, r := range stored {
					ids = append(ids, r.ID)
				}
				err := db.Repos().Delete(ctx, ids...)
				require.NoError(t, err)
			})

			idMapping := make(map[api.RepoName]api.RepoID)
			for _, r := range stored {
				lower := strings.ToLower(string(r.Name))
				idMapping[api.RepoName(lower)] = r.ID
			}

			// Add gitserver_repos rows
			for _, toClone := range tc.gitserverCloned {
				toClone = strings.ToLower(toClone)
				id := idMapping[api.RepoName(toClone)]
				if id == 0 {
					continue
				}
				lastError := ""
				if tc.gitserverFailure != nil && tc.gitserverFailure[toClone] {
					lastError = "Oops"
				}
				err := db.GitserverRepos().Upsert(ctx, &types.GitserverRepo{
					RepoID:      id,
					ShardID:     "test",
					CloneStatus: types.CloneStatusCloned,
					LastError:   lastError,
				})
				require.NoError(t, err)
			}
			t.Cleanup(func() {
				q := sqlf.Sprintf(`DELETE FROM gitserver_repos`)
				_, err = store.Handle().DB().ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
				require.NoError(t, err)
			})

			// Set up ownership of repos
			if tc.repoOwner != nil {
				for _, repo := range stored {
					svc, ok := tc.repoOwner[repo.Name]
					if !ok {
						continue
					}
					q := sqlf.Sprintf(`
						INSERT INTO external_service_repos(external_service_id, repo_id, user_id, clone_url)
						VALUES (%s, %s, NULLIF(%s, 0), 'example.com')
					`, svc.ID, repo.ID, svc.NamespaceUserID)
					_, err = store.Handle().DB().ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
					require.NoError(t, err)

					t.Cleanup(func() {
						q := sqlf.Sprintf(`DELETE FROM external_service_repos WHERE external_service_id = %s`, svc.ID)
						_, err = store.Handle().DB().ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
						require.NoError(t, err)
					})
				}
			}

			clock := timeutil.NewFakeClock(time.Now(), 0)
			syncer := &Syncer{
				Logger: logger,
				Store:  store,
				Now:    clock.Now,
			}

			mockDB := database.NewMockDBFrom(database.NewDB(db))
			if tc.sourcerErr != nil {
				sourcer := NewFakeSourcer(tc.sourcerErr, NewFakeSource(siteLevelService, nil))
				syncer.Sourcer = sourcer

				err = syncer.SyncExternalService(ctx, siteLevelService.ID, time.Millisecond)
				// In prod, SyncExternalService is kicked off by a worker queue. Any error
				// returned will be stored in the external_service_sync_jobs table, so we fake
				// that here.
				if err != nil {
					externalServices := database.NewMockExternalServiceStore()
					externalServices.GetAffiliatedSyncErrorsFunc.SetDefaultReturn(
						map[int64]string{
							siteLevelService.ID: err.Error(),
						},
						nil,
					)
					mockDB.ExternalServicesFunc.SetDefaultReturn(externalServices)
				}
			}

			if tc.err == "" {
				tc.err = "<nil>"
			}

			res, err := FetchStatusMessages(ctx, mockDB, tc.user)
			assert.Equal(t, tc.err, fmt.Sprint(err))
			assert.Equal(t, tc.res, res)
		})
	}
}
