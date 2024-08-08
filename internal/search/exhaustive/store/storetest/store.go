package storetest

import (
	"context"
	"fmt"
	"testing"

	"github.com/keegancsmith/sqlf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/store"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/types"
	internaltypes "github.com/sourcegraph/sourcegraph/internal/types"
)

// CreateJobCascade creates a cascade of jobs (1 search job -> n repo jobs -> m
// repo rev jobs) with states as defined in StateCascade.
//
// This is a fairly large test helper, because don't want to start the worker
// routines, but instead we want to create a snapshot of the state of the jobs
// at a given point in time.
func CreateJobCascade(
	t *testing.T,
	ctx context.Context,
	stor *store.Store,
	casc StateCascade,
) (searchJobID int64) {
	t.Helper()

	searchJob := types.ExhaustiveSearchJob{
		InitiatorID: actor.FromContext(ctx).UID,
		Query:       "repo:job1",
		WorkerJob:   types.WorkerJob{State: casc.SearchJob},
	}

	repoJobs := make([]types.ExhaustiveSearchRepoJob, len(casc.RepoJobs))
	for i, r := range casc.RepoJobs {
		repoJobs[i] = types.ExhaustiveSearchRepoJob{
			WorkerJob: types.WorkerJob{State: r},
			RepoID:    1, // same repo for all tests
			RefSpec:   "HEAD",
		}
	}

	repoRevJobs := make([]types.ExhaustiveSearchRepoRevisionJob, len(casc.RepoRevJobs))
	for i, rr := range casc.RepoRevJobs {
		repoRevJobs[i] = types.ExhaustiveSearchRepoRevisionJob{
			WorkerJob: types.WorkerJob{State: rr, FailureMessage: fmt.Sprintf("repoRevJob-%d", i)},
			Revision:  "HEAD",
		}
	}

	jobID, err := stor.CreateExhaustiveSearchJob(ctx, searchJob)
	require.NoError(t, err)
	assert.NotZero(t, jobID)

	err = stor.Exec(ctx, sqlf.Sprintf("UPDATE exhaustive_search_jobs SET state = %s WHERE id = %s", casc.SearchJob, jobID))
	require.NoError(t, err)

	for i, r := range repoJobs {
		r.SearchJobID = jobID
		repoJobID, err := stor.CreateExhaustiveSearchRepoJob(ctx, r)
		require.NoError(t, err)
		assert.NotZero(t, repoJobID)

		err = stor.Exec(ctx, sqlf.Sprintf("UPDATE exhaustive_search_repo_jobs SET state = %s WHERE id = %s", casc.RepoJobs[i], repoJobID))
		require.NoError(t, err)

		for j, rr := range repoRevJobs {
			rr.SearchRepoJobID = repoJobID
			repoRevJobID, err := stor.CreateExhaustiveSearchRepoRevisionJob(ctx, rr)
			require.NoError(t, err)
			assert.NotZero(t, repoRevJobID)
			require.NoError(t, err)

			err = stor.Exec(ctx, sqlf.Sprintf("UPDATE exhaustive_search_repo_revision_jobs SET state = %s, failure_message = %s WHERE id = %s", casc.RepoRevJobs[j], repoRevJobs[j].FailureMessage, repoRevJobID))
			require.NoError(t, err)
		}
	}

	return jobID
}

type StateCascade struct {
	SearchJob   types.JobState
	RepoJobs    []types.JobState
	RepoRevJobs []types.JobState
}

func CreateRepo(db database.DB, name string) (api.RepoID, error) {
	repoStore := db.Repos()
	repo := internaltypes.Repo{Name: api.RepoName(name)}
	err := repoStore.Create(context.Background(), &repo)
	return repo.ID, err
}

func CreateUser(store *basestore.Store, username string) (int32, error) {
	admin := username == "admin"
	q := sqlf.Sprintf(`INSERT INTO users(username, site_admin) VALUES(%s, %s) RETURNING id`, username, admin)
	return basestore.ScanAny[int32](store.QueryRow(context.Background(), q))
}
