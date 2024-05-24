package syntactic_indexing

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/reposcheduler"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/syntactic_indexing/internal/testutils"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/syntactic_indexing/jobstore"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/stretchr/testify/require"
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

	tacosRepoId, tacosRepoName, tacosCommit := 1, "tangy/tacos", testutils.MakeCommit(1)
	empanadasRepoId, empanadasRepoName, empanadasCommit := 2, "salty/empanadas", testutils.MakeCommit(2)
	mangosRepoId, mangosRepoName, mangosCommit := 3, "juicy/mangos", testutils.MakeCommit(3)

	testutils.InsertRepo(t, db, tacosRepoId, tacosRepoName)
	testutils.InsertRepo(t, db, empanadasRepoId, empanadasRepoName)
	testutils.InsertRepo(t, db, mangosRepoId, mangosRepoName)

	gsClient.ResolveRevisionFunc.SetDefaultHook(func(ctx context.Context, repo api.RepoName, rev string, options gitserver.ResolveRevisionOptions) (api.CommitID, error) {
		isTacos := repo == api.RepoName(tacosRepoName) && rev == tacosCommit
		isEmpanadas := repo == api.RepoName(empanadasRepoName) && rev == empanadasCommit

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
	scheduled, err := enqueuer.QueueIndexes(ctx, tacosRepoId, tacosCommit, EnqueueOptions{})

	require.NoError(t, err)
	require.True(t, len(scheduled) == 1)
	require.Equal(t, scheduled[0].Commit, tacosCommit)
	require.Equal(t, scheduled[0].RepositoryID, tacosRepoId)
	require.Equal(t, scheduled[0].State, jobstore.Queued)
	require.Equal(t, scheduled[0].RepositoryName, tacosRepoName)

	// scheduling the same (repo, revision) twice doesn't return an error,
	// but also doesn't insert a new job
	result := unwrap(enqueuer.QueueIndexes(ctx, tacosRepoId, tacosCommit, EnqueueOptions{}))(t)
	require.Empty(t, result)

	// cannot schedule for non existent repositories
	_, err = enqueuer.QueueIndexes(ctx, 250, tacosCommit, EnqueueOptions{})
	require.Error(t, err)

	// cannot schedule for non existent revisions
	_, err = enqueuer.QueueIndexes(ctx, mangosRepoId, mangosCommit, EnqueueOptions{})
	require.Error(t, err)

}
