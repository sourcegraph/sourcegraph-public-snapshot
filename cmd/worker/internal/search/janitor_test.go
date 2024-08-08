package search

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/keegancsmith/sqlf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/service"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/store"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/store/storetest"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/types"
)

func TestJanitor(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx, db, exhaustiveStore, svc, bucket := setupTest(t)

	cases := []struct {
		name                string
		cascade             storetest.StateCascade
		wantLogUploaded     bool
		wantJobState        types.JobState
		wantColIsAggregated bool
	}{
		{
			name: "all jobs completed",
			cascade: storetest.StateCascade{
				SearchJob: types.JobStateCompleted,
				RepoJobs: []types.JobState{
					types.JobStateCompleted,
				},
				RepoRevJobs: []types.JobState{
					types.JobStateCompleted,
					types.JobStateCompleted,
				},
			},
			wantLogUploaded:     true,
			wantJobState:        types.JobStateCompleted,
			wantColIsAggregated: true,
		},
		{
			name: "1 job failed",
			cascade: storetest.StateCascade{
				SearchJob: types.JobStateCompleted,
				RepoJobs: []types.JobState{
					types.JobStateCompleted,
				},
				RepoRevJobs: []types.JobState{
					types.JobStateFailed, // failed is terminal
					types.JobStateCompleted,
				},
			},
			wantLogUploaded:     true,
			wantJobState:        types.JobStateFailed,
			wantColIsAggregated: true,
		},
		{
			name: "Still processing, don't update job state",
			cascade: storetest.StateCascade{
				SearchJob: types.JobStateCompleted,
				RepoJobs: []types.JobState{
					types.JobStateCompleted,
				},
				RepoRevJobs: []types.JobState{
					types.JobStateErrored, // errored is not terminal
				},
			},
			wantLogUploaded:     false,
			wantJobState:        types.JobStateCompleted,
			wantColIsAggregated: false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			searchJobID := storetest.CreateJobCascade(t, ctx, exhaustiveStore, tt.cascade)

			// Use context.Background() to test if the janitor sets the user context
			// correctly
			err := runJanitor(context.Background(), db, svc)
			require.NoError(t, err)

			j, err := exhaustiveStore.GetExhaustiveSearchJob(ctx, searchJobID)
			require.NoError(t, err)
			assert.Equal(t, tt.wantJobState, j.State)

			logs, ok := bucket[fmt.Sprintf("log-%d.csv", searchJobID)]
			require.Equal(t, tt.wantLogUploaded, ok)
			if tt.wantLogUploaded {
				require.Equal(t, len(strings.Split(logs, "\n")), len(tt.cascade.RepoRevJobs)+2) // 2 = 1 header + 1 final newline
			}

			// Ensure that the repo jobs have been deleted.
			wantCount := len(tt.cascade.RepoJobs)
			if tt.wantLogUploaded {
				wantCount = 0
			}
			require.Equal(t, wantCount, countRepoJobs(t, db, searchJobID))

			// Ensure that the job is marked as aggregated
			require.Equal(t, tt.wantColIsAggregated, j.IsAggregated)
		})
	}
}

func countRepoJobs(t *testing.T, db database.DB, searchJobID int64) int {
	t.Helper()

	q := sqlf.Sprintf("SELECT COUNT(*) FROM exhaustive_search_repo_jobs WHERE search_job_id = %s", searchJobID)
	var count int
	err := db.QueryRowContext(context.Background(), q.Query(sqlf.PostgresBindVar), q.Args()...).Scan(&count)
	require.NoError(t, err)
	return count
}

func setupTest(t *testing.T) (context.Context, database.DB, *store.Store, *service.Service, map[string]string) {
	t.Helper()

	observationCtx := observation.TestContextTB(t)
	logger := observationCtx.Logger

	mockUploadStore, bucket := newMockUploadStore(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	stor := store.New(db, observation.TestContextTB(t))
	svc := service.New(observationCtx, stor, mockUploadStore, service.NewSearcherFake())

	bs := basestore.NewWithHandle(db.Handle())
	userID, err := storetest.CreateUser(bs, "user1")
	require.NoError(t, err)

	_, err = storetest.CreateRepo(db, "repo1")
	require.NoError(t, err)

	ctx := actor.WithActor(context.Background(), actor.FromUser(userID))

	return ctx, db, stor, svc, bucket
}
