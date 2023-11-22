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
	"github.com/sourcegraph/zoekt"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestStatusMessages(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Background()

	db := database.NewDB(logger, dbtest.NewDB(t))
	store := NewStore(logtest.Scoped(t), db)

	mockGitserverClient := gitserver.NewMockClient()

	extSvc := &types.ExternalService{
		ID:          1,
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "token": "beef", "repos": ["owner/name"]}`),
		Kind:        extsvc.KindGitHub,
		DisplayName: "github.com - site",
	}
	err := db.ExternalServices().Upsert(ctx, extSvc)
	require.NoError(t, err)

	testCases := []struct {
		testSetup func()
		name      string
		repos     types.Repos
		// maps repoName to CloneStatus
		cloneStatus map[string]types.CloneStatus
		// indexed is list of repo names that are indexed
		indexed          []string
		gitserverFailure map[string]bool
		sourcerErr       error
		res              []StatusMessage
		err              string
	}{
		{
			testSetup: func() {
				conf.Mock(&conf.Unified{
					SiteConfiguration: schema.SiteConfiguration{
						DisableAutoGitUpdates: true,
					},
				})
			},
			name: "disableAutoGitUpdates set to true",
			res: []StatusMessage{
				{
					GitUpdatesDisabled: &GitUpdatesDisabled{
						Message: "Repositories will not be cloned or updated.",
					},
				},
				{
					NoRepositoriesDetected: &NoRepositoriesDetected{
						Message: "No repositories have been added to Sourcegraph.",
					},
				},
			},
		},
		{
			name:        "site-admin: all cloned and indexed",
			cloneStatus: map[string]types.CloneStatus{"foobar": types.CloneStatusCloned},
			indexed:     []string{"foobar"},
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
				{
					Indexing: &IndexingProgress{NotIndexed: 1},
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
				{
					Indexing: &IndexingProgress{NotIndexed: 1},
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
				{
					Indexing: &IndexingProgress{NotIndexed: 2},
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
			indexed: []string{"repo-6"},
			res: []StatusMessage{
				{
					Cloning: &CloningProgress{
						Message: "2 repositories enqueued for cloning. 2 repositories currently cloning...",
					},
				},
				{
					Indexing: &IndexingProgress{Indexed: 1, NotIndexed: 5},
				},
			},
		},
		{
			name:       "site-admin: no repos detected",
			repos:      []*types.Repo{},
			sourcerErr: nil,
			res: []StatusMessage{
				{
					NoRepositoriesDetected: &NoRepositoriesDetected{
						Message: "No repositories have been added to Sourcegraph.",
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
			indexed:          []string{"foobar", "barfoo"},
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
			indexed:          []string{"foobar", "barfoo"},
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
		{
			testSetup: func() {
				conf.Mock(&conf.Unified{
					SiteConfiguration: schema.SiteConfiguration{
						GitserverDiskUsageWarningThreshold: pointers.Ptr(10),
					},
				})

				mockGitserverClient.SystemsInfoFunc.SetDefaultReturn([]protocol.SystemInfo{
					{
						Address:     "gitserver-0",
						PercentUsed: 75.10345,
					},
				}, nil)

			},
			name:        "site-admin: gitserver disk threshold reached (configured threshold)",
			cloneStatus: map[string]types.CloneStatus{"foobar": types.CloneStatusCloned},
			indexed:     []string{"foobar"},
			repos:       []*types.Repo{{Name: "foobar"}},
			res: []StatusMessage{
				{
					GitserverDiskThresholdReached: &GitserverDiskThresholdReached{
						Message: "The disk usage on gitserver \"gitserver-0\" is over 10% (75.10% used).",
					},
				},
			},
		},
		{
			testSetup: func() {
				mockGitserverClient.SystemsInfoFunc.SetDefaultReturn([]protocol.SystemInfo{
					{
						Address:     "gitserver-0",
						PercentUsed: 95.10345,
					},
				}, nil)

			},
			name:        "site-admin: gitserver disk threshold reached (default threshold)",
			cloneStatus: map[string]types.CloneStatus{"foobar": types.CloneStatusCloned},
			indexed:     []string{"foobar"},
			repos:       []*types.Repo{{Name: "foobar"}},
			res: []StatusMessage{
				{
					GitserverDiskThresholdReached: &GitserverDiskThresholdReached{
						Message: "The disk usage on gitserver \"gitserver-0\" is over 90% (95.10% used).",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			if tc.testSetup != nil {
				tc.testSetup()
			}

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
				conf.Mock(nil)

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
			for _, repoName := range tc.indexed {
				id := uint32(idMapping[api.RepoName(repoName)])
				if id == 0 {
					continue
				}
				err := db.ZoektRepos().UpdateIndexStatuses(ctx, zoekt.ReposMap{
					id: {
						Branches: []zoekt.RepositoryBranch{{Name: "main", Version: "d34db33f"}},
					},
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
				ObsvCtx: observation.TestContextTB(t),
				Store:   store,
				Now:     clock.Now,
			}

			mockDB := dbmocks.NewMockDBFrom(db)
			if tc.sourcerErr != nil {
				sourcer := NewFakeSourcer(tc.sourcerErr, NewFakeSource(extSvc, nil))
				syncer.Sourcer = sourcer

				noopRecorder := func(ctx context.Context, progress SyncProgress, final bool) error {
					return nil
				}
				err = syncer.SyncExternalService(ctx, extSvc.ID, time.Millisecond, noopRecorder)
				// In prod, SyncExternalService is kicked off by a worker queue. Any error
				// returned will be stored in the external_service_sync_jobs table, so we fake
				// that here.
				if err != nil {
					externalServices := dbmocks.NewMockExternalServiceStore()
					externalServices.GetLatestSyncErrorsFunc.SetDefaultReturn(
						[]*database.SyncError{
							{ServiceID: extSvc.ID, Message: err.Error()},
						},
						nil,
					)
					mockDB.ExternalServicesFunc.SetDefaultReturn(externalServices)
				}
			}

			if len(tc.repos) < 1 && tc.sourcerErr == nil {
				externalServices := dbmocks.NewMockExternalServiceStore()
				externalServices.GetLatestSyncErrorsFunc.SetDefaultReturn(
					[]*database.SyncError{},
					nil,
				)
				mockDB.ExternalServicesFunc.SetDefaultReturn(externalServices)
			}

			if tc.err == "" {
				tc.err = "<nil>"
			}

			res, err := FetchStatusMessages(ctx, mockDB, mockGitserverClient)
			assert.Equal(t, tc.err, fmt.Sprint(err))
			assert.Equal(t, tc.res, res)
		})
	}
}
