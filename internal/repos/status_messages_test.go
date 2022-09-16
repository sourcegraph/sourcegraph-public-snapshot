package repos

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/keegancsmith/sqlf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestStatusMessages(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Background()

	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := NewStore(logtest.Scoped(t), db)

	extSvc := &types.ExternalService{
		ID:          1,
		Config:      extsvc.NewEmptyConfig(),
		Kind:        extsvc.KindGitHub,
		DisplayName: "github.com - site",
	}
	err := db.ExternalServices().Upsert(ctx, extSvc)
	require.NoError(t, err)

	testCases := []struct {
		name  string
		repos types.Repos
		// maps repoName to CloneStatus
		cloneStatus      map[string]types.CloneStatus
		gitserverFailure map[string]bool
		sourcerErr       error
		res              []StatusMessage
		err              string
	}{
		{
			name:        "site-admin: all cloned",
			cloneStatus: map[string]types.CloneStatus{"foobar": types.CloneStatusCloned},
			repos:       []*types.Repo{{Name: "foobar"}},
			res:         nil,
		},
		{
			name:        "site-admin: one repository not cloned",
			repos:       []*types.Repo{{Name: "foobar"}},
			cloneStatus: map[string]types.CloneStatus{},
			res: []StatusMessage{
				{
					Cloning: &CloningProgress{
						Message: "1 repository enqueued for cloning.",
					},
				},
			},
		},
		{
			name:        "site-admin: one repository cloning",
			repos:       []*types.Repo{{Name: "foobar"}},
			cloneStatus: map[string]types.CloneStatus{"foobar": types.CloneStatusCloning},
			res: []StatusMessage{
				{
					Cloning: &CloningProgress{
						Message: "1 repository currently cloning...",
					},
				},
			},
		},
		{
			name:        "site-admin: one not cloned, one cloning",
			repos:       []*types.Repo{{Name: "foobar"}, {Name: "barfoo"}},
			cloneStatus: map[string]types.CloneStatus{"foobar": types.CloneStatusCloning, "barfoo": types.CloneStatusNotCloned},
			res: []StatusMessage{
				{
					Cloning: &CloningProgress{
						Message: "1 repository enqueued for cloning. 1 repository currently cloning...",
					},
				},
			},
		},
		{
			name: "site-admin: multiple not cloned, multiple cloning, multiple cloned",
			repos: []*types.Repo{
				{Name: "repo-1"},
				{Name: "repo-2"},
				{Name: "repo-3"},
				{Name: "repo-4"},
				{Name: "repo-5"},
				{Name: "repo-6"},
			},
			cloneStatus: map[string]types.CloneStatus{
				"repo-1": types.CloneStatusCloning,
				"repo-2": types.CloneStatusCloning,
				"repo-3": types.CloneStatusNotCloned,
				"repo-4": types.CloneStatusNotCloned,
				"repo-5": types.CloneStatusCloned,
				"repo-6": types.CloneStatusCloned,
			},
			res: []StatusMessage{
				{
					Cloning: &CloningProgress{
						Message: "2 repositories enqueued for cloning. 2 repositories currently cloning...",
					},
				},
			},
		},
		{
			name:        "site-admin: subset cloned",
			repos:       []*types.Repo{{Name: "foobar"}, {Name: "barfoo"}},
			cloneStatus: map[string]types.CloneStatus{"foobar": types.CloneStatusCloned},
			res: []StatusMessage{
				{
					Cloning: &CloningProgress{
						Message: "1 repository enqueued for cloning.",
					},
				},
			},
		},
		{
			name:  "site-admin: more cloned than stored",
			repos: []*types.Repo{{Name: "foobar"}},
			cloneStatus: map[string]types.CloneStatus{
				"foobar": types.CloneStatusCloned,
				"barfoo": types.CloneStatusCloned,
			},
			res: nil,
		},
		{
			name:  "site-admin: cloned different than stored",
			repos: []*types.Repo{{Name: "foobar"}, {Name: "barfoo"}},
			cloneStatus: map[string]types.CloneStatus{
				"one":   types.CloneStatusCloned,
				"two":   types.CloneStatusCloned,
				"three": types.CloneStatusCloned,
			},
			res: []StatusMessage{
				{
					Cloning: &CloningProgress{
						Message: "2 repositories enqueued for cloning.",
					},
				},
			},
		},
		{
			name:  "site-admin: one repo failed to sync",
			repos: []*types.Repo{{Name: "foobar"}, {Name: "barfoo"}},
			cloneStatus: map[string]types.CloneStatus{
				"foobar": types.CloneStatusCloned,
				"barfoo": types.CloneStatusCloned,
			},
			gitserverFailure: map[string]bool{"foobar": true},
			res: []StatusMessage{
				{
					SyncError: &SyncError{
						Message: "1 repository failed last attempt to sync content from code host",
					},
				},
			},
		},
		{
			name:  "site-admin: two repos failed to sync",
			repos: []*types.Repo{{Name: "foobar"}, {Name: "barfoo"}},
			cloneStatus: map[string]types.CloneStatus{
				"foobar": types.CloneStatusCloned,
				"barfoo": types.CloneStatusCloned,
			},
			gitserverFailure: map[string]bool{"foobar": true, "barfoo": true},
			res: []StatusMessage{
				{
					SyncError: &SyncError{
						Message: "2 repositories failed last attempt to sync content from code host",
					},
				},
			},
		},
		{
			name:       "one external service syncer err",
			sourcerErr: errors.New("github is down"),
			res: []StatusMessage{
				{
					ExternalServiceSyncError: &ExternalServiceSyncError{
						Message:           "github is down",
						ExternalServiceId: extSvc.ID,
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		ctx := context.Background()

		t.Run(tc.name, func(t *testing.T) {
			stored := tc.repos.Clone()
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
			for repoName, cloneStatus := range tc.cloneStatus {
				id := idMapping[api.RepoName(repoName)]
				if id == 0 {
					continue
				}
				lastError := ""
				if tc.gitserverFailure != nil && tc.gitserverFailure[repoName] {
					lastError = "Oops"
				}
				err := db.GitserverRepos().Update(ctx, &types.GitserverRepo{
					RepoID:      id,
					ShardID:     "test",
					CloneStatus: cloneStatus,
					LastError:   lastError,
				})
				require.NoError(t, err)
			}

			// Set up ownership of repos
			for _, repo := range stored {
				q := sqlf.Sprintf(`
						INSERT INTO external_service_repos(external_service_id, repo_id, clone_url)
						VALUES (%s, %s, 'example.com')
					`, extSvc.ID, repo.ID)
				_, err = store.Handle().ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
				require.NoError(t, err)

				t.Cleanup(func() {
					q := sqlf.Sprintf(`DELETE FROM external_service_repos WHERE external_service_id = %s`, extSvc.ID)
					_, err = store.Handle().ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
					require.NoError(t, err)
				})
			}

			clock := timeutil.NewFakeClock(time.Now(), 0)
			syncer := &Syncer{
				Logger: logger,
				Store:  store,
				Now:    clock.Now,
			}

			mockDB := database.NewMockDBFrom(db)
			if tc.sourcerErr != nil {
				sourcer := NewFakeSourcer(tc.sourcerErr, NewFakeSource(extSvc, nil))
				syncer.Sourcer = sourcer

				err = syncer.SyncExternalService(ctx, extSvc.ID, time.Millisecond)
				// In prod, SyncExternalService is kicked off by a worker queue. Any error
				// returned will be stored in the external_service_sync_jobs table, so we fake
				// that here.
				if err != nil {
					externalServices := database.NewMockExternalServiceStore()
					externalServices.GetLatestSyncErrorsFunc.SetDefaultReturn(
						map[int64]string{
							extSvc.ID: err.Error(),
						},
						nil,
					)
					mockDB.ExternalServicesFunc.SetDefaultReturn(externalServices)
				}
			}

			if tc.err == "" {
				tc.err = "<nil>"
			}

			res, err := FetchStatusMessages(ctx, mockDB)
			assert.Equal(t, tc.err, fmt.Sprint(err))
			assert.Equal(t, tc.res, res)
		})
	}
}
