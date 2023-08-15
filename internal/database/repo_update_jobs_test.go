package database

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestRepoUpdateJobs(t *testing.T) {
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

	// Creating a repo.
	err = db.Repos().Create(ctx, &types.Repo{ID: api.RepoID(1), Name: "repo1"})
	require.NoError(t, err)

	// Queued job should be successfully created.
	createdJob, ok, err := store.Create(ctx, RepoUpdateJobOpts{RepoID: 1, Priority: types.HighPriorityRepoUpdate})
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, 1, createdJob.ID)
	assert.Equal(t, types.HighPriorityRepoUpdate, createdJob.Priority)
	assert.Equal(t, "queued", createdJob.State)

	wantJob := createdJob
	// Created job should be listed.
	repoUpdateJobs, err = store.List(ctx, ListRepoUpdateJobOpts{ID: createdJob.ID, States: []string{createdJob.State, "errored", "failed"}})
	require.NoError(t, err)
	assert.Len(t, repoUpdateJobs, 1)
	gotJob := repoUpdateJobs[0]
	assert.Equal(t, wantJob.RepoID, gotJob.RepoID)
	assert.Equal(t, wantJob.Priority, gotJob.Priority)

	// Second queued job for the same Repo ID should not be created.
	_, ok, err = store.Create(ctx, RepoUpdateJobOpts{RepoID: 1, Priority: types.HighPriorityRepoUpdate})
	require.NoError(t, err)
	assert.False(t, ok)
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
	haveJob, _, err := store.Create(ctx, RepoUpdateJobOpts{RepoID: 1, Priority: types.HighPriorityRepoUpdate})
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
