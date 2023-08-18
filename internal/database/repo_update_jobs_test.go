package database

import (
	"context"
	"testing"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func TestRepoUpdateJobs_Create(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()
	store := RepoUpdateJobStoreWith(db)

	// Zero jobs expected when none are inserted.
	repoUpdateJobs, err := store.List(ctx, ListRepoUpdateJobOpts{})
	require.NoError(t, err)
	assert.Empty(t, repoUpdateJobs)

	// Creating repos.
	err = db.Repos().Create(ctx, &types.Repo{ID: api.RepoID(1), Name: "repo1"})
	require.NoError(t, err)
	err = db.Repos().Create(ctx, &types.Repo{ID: api.RepoID(2), Name: "repo2"})
	require.NoError(t, err)

	// Queued job should be successfully created.
	createdJob, ok, err := store.Create(ctx, CreateRepoUpdateJobOpts{RepoName: "repo1", Priority: types.HighPriorityRepoUpdate, FetchRevision: "HEAD"})
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, 1, createdJob.ID)
	assert.Equal(t, types.HighPriorityRepoUpdate, createdJob.Priority)
	assert.Equal(t, "queued", createdJob.State)
	assert.Equal(t, "HEAD", createdJob.FetchRevision)

	listAndAssert := func(wantJob types.RepoUpdateJob, listOpts ListRepoUpdateJobOpts) {
		repoUpdateJobs, err = store.List(ctx, listOpts)
		require.NoError(t, err)
		assert.Len(t, repoUpdateJobs, 1)
		gotJob := repoUpdateJobs[0]
		assert.Equal(t, wantJob.RepoID, gotJob.RepoID)
		assert.Equal(t, wantJob.Priority, gotJob.Priority)
		assert.Equal(t, wantJob.OverwriteClone, gotJob.OverwriteClone)
	}

	// Created job should be listed.
	listAndAssert(createdJob, ListRepoUpdateJobOpts{ID: createdJob.ID, States: []string{createdJob.State, "errored", "failed"}})

	// Second queued job for the same Repo ID should not be created.
	_, ok, err = store.Create(ctx, CreateRepoUpdateJobOpts{RepoName: "repo1", Priority: types.HighPriorityRepoUpdate})
	require.NoError(t, err)
	assert.False(t, ok)

	// Second queued job for a different repo should be created successfully.
	createdJob, ok, err = store.Create(ctx, CreateRepoUpdateJobOpts{RepoName: "repo2", Priority: types.LowPriorityRepoUpdate, Clone: true, OverwriteClone: true})
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, 3, createdJob.ID)
	assert.Equal(t, types.LowPriorityRepoUpdate, createdJob.Priority)
	assert.Equal(t, "queued", createdJob.State)
	assert.True(t, createdJob.Clone)
	assert.True(t, createdJob.OverwriteClone)

	// Created job should be listed.
	listAndAssert(createdJob, ListRepoUpdateJobOpts{ID: createdJob.ID, States: []string{createdJob.State}})

	// Both job should be listed if we don't specify the ID.
	repoUpdateJobs, err = store.List(ctx, ListRepoUpdateJobOpts{})
	require.NoError(t, err)
	assert.Len(t, repoUpdateJobs, 2)
}

func TestRepoUpdateJobs_List(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()
	store := RepoUpdateJobStoreWith(db)

	// Creating 3 repos.
	err := db.Repos().Create(ctx, &types.Repo{ID: api.RepoID(1), Name: "repo1"})
	require.NoError(t, err)
	err = db.Repos().Create(ctx, &types.Repo{ID: api.RepoID(2), Name: "repo2"})
	require.NoError(t, err)
	err = db.Repos().Create(ctx, &types.Repo{ID: api.RepoID(3), Name: "repo3"})
	require.NoError(t, err)

	// Creating 2 queued jobs and 1 finished job.
	queuedJob1, _, err := store.Create(ctx, CreateRepoUpdateJobOpts{RepoName: "repo1", Priority: types.HighPriorityRepoUpdate})
	require.NoError(t, err)
	queuedJob2, _, err := store.Create(ctx, CreateRepoUpdateJobOpts{RepoName: "repo2", Priority: types.HighPriorityRepoUpdate})
	require.NoError(t, err)

	completedJob, _, err := store.Create(ctx, CreateRepoUpdateJobOpts{RepoName: "repo3", Priority: types.LowPriorityRepoUpdate})
	require.NoError(t, err)
	err = store.Handle().Exec(ctx, sqlf.Sprintf("UPDATE repo_update_jobs SET state = 'completed' WHERE id = 3"))
	require.NoError(t, err)
	completedJob.State = "completed"

	tests := map[string]struct {
		listOpts ListRepoUpdateJobOpts
		wantJobs []types.RepoUpdateJob
	}{
		"list by job ID": {
			listOpts: ListRepoUpdateJobOpts{ID: 1},
			wantJobs: []types.RepoUpdateJob{queuedJob1},
		},
		"list by repo ID": {
			listOpts: ListRepoUpdateJobOpts{RepoID: 1},
			wantJobs: []types.RepoUpdateJob{queuedJob1},
		},
		"list by repo name": {
			listOpts: ListRepoUpdateJobOpts{RepoName: "repo1"},
			wantJobs: []types.RepoUpdateJob{queuedJob1},
		},
		"list by state": {
			listOpts: ListRepoUpdateJobOpts{States: []string{"completed"}},
			wantJobs: []types.RepoUpdateJob{completedJob},
		},
		"list by repo name and state": {
			listOpts: ListRepoUpdateJobOpts{RepoName: "repo2", States: []string{"completed", "queued"}},
			wantJobs: []types.RepoUpdateJob{queuedJob2},
		},
		"list by repo name and ID, ID takes precedence": {
			listOpts: ListRepoUpdateJobOpts{RepoName: "repo2", RepoID: 1},
			wantJobs: []types.RepoUpdateJob{queuedJob1},
		},
		"list first queued": {
			listOpts: ListRepoUpdateJobOpts{
				States: []string{"queued"},
				PaginationArgs: &PaginationArgs{
					First:     pointers.Ptr(1),
					OrderBy:   OrderBy{{Field: "repo_update_jobs.queued_at"}},
					Ascending: true,
				},
			},
			wantJobs: []types.RepoUpdateJob{queuedJob1},
		},
		"order by queued ASC": {
			listOpts: ListRepoUpdateJobOpts{
				States: []string{"queued", "completed"},
				PaginationArgs: &PaginationArgs{
					First:     pointers.Ptr(3),
					OrderBy:   OrderBy{{Field: "repo_update_jobs.queued_at"}},
					Ascending: true,
				},
			},
			wantJobs: []types.RepoUpdateJob{queuedJob1, queuedJob2, completedJob},
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			gotJobs, err := store.List(ctx, test.listOpts)
			require.NoError(t, err)
			wantJobs := test.wantJobs
			assert.Equal(t, len(wantJobs), len(gotJobs))
			for i := 0; i < len(wantJobs); i++ {
				assert.Equal(t, wantJobs[i], gotJobs[i])
			}
		})
	}
}

func TestRepoUpdateJobs_SaveUpdateJobResults(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()
	store := RepoUpdateJobStoreWith(db)

	// No error should be returned when updating a non-existent job.
	err := store.SaveUpdateJobResults(ctx, 1, SaveUpdateJobResultsOpts{LastFetched: time.Time{}})
	require.NoError(t, err)

	// Creating a repo.
	err = db.Repos().Create(ctx, &types.Repo{ID: api.RepoID(1), Name: "repo1"})
	require.NoError(t, err)

	// Queued job should be successfully created.
	haveJob, _, err := store.Create(ctx, CreateRepoUpdateJobOpts{RepoName: "repo1", Priority: types.HighPriorityRepoUpdate})
	require.NoError(t, err)
	assert.Zero(t, haveJob.LastFetched)
	assert.Zero(t, haveJob.LastChanged)
	assert.Zero(t, haveJob.UpdateIntervalSeconds)

	now := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	err = store.SaveUpdateJobResults(ctx, haveJob.ID, SaveUpdateJobResultsOpts{LastFetched: time.Time{}, LastChanged: now, UpdateIntervalSeconds: 42})
	require.NoError(t, err)

	// Updated job should be listed.
	repoUpdateJobs, err := store.List(ctx, ListRepoUpdateJobOpts{})
	require.NoError(t, err)
	assert.Len(t, repoUpdateJobs, 1)
	gotJob := repoUpdateJobs[0]
	assert.Zero(t, gotJob.LastFetched)
	assert.Equal(t, now, gotJob.LastChanged)
	assert.Equal(t, 42, gotJob.UpdateIntervalSeconds)
}
