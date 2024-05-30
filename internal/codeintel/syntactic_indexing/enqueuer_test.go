package syntactic_indexing

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/reposcheduler"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/syntactic_indexing/internal/testutils"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/syntactic_indexing/jobstore"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestSyntacticIndexingStoreEnqueue(t *testing.T) {
	/*
		The purpose of this test is to verify that methods InsertIndexes and IsQueued
		correctly interact with each other, and that the records inserted using those methods
		are valid from the point of view of the DB worker interface
	*/
	observationCtx := observation.TestContextTB(t)
	sqlDB := dbtest.NewDB(t)
	db := database.NewDB(observationCtx.Logger, sqlDB)
	ctx := context.Background()

	jobStore, err := jobstore.NewStoreWithDB(observationCtx, sqlDB)
	require.NoError(t, err, "unexpected error creating dbworker stores")

	repoSchedulingStore := reposcheduler.NewSyntacticStore(observationCtx, db)

	repoStore := db.Repos()

	gsClient := gitserver.NewMockClient()
	enqueuer := NewIndexEnqueuer(observationCtx, jobStore, repoSchedulingStore, repoStore, gsClient)

	tacosRepoId, tacosRepoName, tacosCommit := api.RepoID(1), "tangy/tacos", testutils.MakeCommit(1)
	empanadasRepoId, empanadasRepoName, empanadasCommit := api.RepoID(2), "salty/empanadas", testutils.MakeCommit(2)
	mangosRepoId, mangosRepoName := api.RepoID(3), "juicy/mangos"

	testutils.InsertRepo(t, db, tacosRepoId, tacosRepoName)
	testutils.InsertRepo(t, db, empanadasRepoId, empanadasRepoName)
	testutils.InsertRepo(t, db, mangosRepoId, mangosRepoName)

	gsClient.ResolveRevisionFunc.SetDefaultHook(func(ctx context.Context, repo api.RepoName, rev string, options gitserver.ResolveRevisionOptions) (api.CommitID, error) {
		isTacos := repo == api.RepoName(tacosRepoName) && rev == string(tacosCommit)
		isEmpanadas := repo == api.RepoName(empanadasRepoName) && rev == string(empanadasCommit)

		if isTacos || isEmpanadas {
			return api.CommitID(rev), nil
		} else {
			return api.CommitID("what"), errors.New(fmt.Sprintf("Unexpected repo (`%s`) and revision (`%s`) requested from gitserver: ", repo, rev))
		}
	})

	isQueued, err := jobStore.IsQueued(ctx, tacosRepoId, tacosCommit)
	require.False(t, isQueued)
	require.NoError(t, err)

	// Happy path
	scheduled, err := enqueuer.QueueIndexingJobs(ctx, tacosRepoId, tacosCommit, EnqueueOptions{})

	require.NoError(t, err)
	require.Equal(t, 1, len(scheduled))
	require.Equal(t, scheduled[0].Commit, tacosCommit)
	require.Equal(t, scheduled[0].RepositoryID, tacosRepoId)
	require.Equal(t, scheduled[0].State, jobstore.Queued)
	require.Equal(t, scheduled[0].RepositoryName, tacosRepoName)

	// scheduling the same (repo, revision) twice doesn't return an error,
	// but also doesn't insert a new job
	result := unwrap(enqueuer.QueueIndexingJobs(ctx, tacosRepoId, tacosCommit, EnqueueOptions{}))(t)
	require.Empty(t, result)

	// force: true in EnqueueOptions allows scheduling the same (repo, revision) twice
	reinserted := unwrap(enqueuer.QueueIndexingJobs(ctx, tacosRepoId, tacosCommit, EnqueueOptions{force: true}))(t)
	require.Equal(t, 1, len(reinserted))
	require.NotEqual(t, reinserted[0].ID, scheduled[0].ID) // ensure it's actually a new job
	require.Equal(t, reinserted[0].Commit, tacosCommit)
	require.Equal(t, reinserted[0].RepositoryID, tacosRepoId)
	require.Equal(t, reinserted[0].State, jobstore.Queued)
	require.Equal(t, reinserted[0].RepositoryName, tacosRepoName)

}
