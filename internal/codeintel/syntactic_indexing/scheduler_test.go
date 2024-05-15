package syntactic_indexing

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/syntactic_indexing/internal/testutils"
	// policystore "github.com/sourcegraph/sourcegraph/internal/codeintel/policies/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/stretchr/testify/require"
)

func TestSyntacticIndexingScheduler(t *testing.T) {
	/*
		The purpose of this test is to verify that methods InsertIndexes and IsQueued
		correctly interact with each other, and that the records inserted using those methods
		are valid from the point of view of the DB worker interface
	*/
	observationCtx := observation.TestContextTB(t)
	sqlDB := dbtest.NewDB(t)

	config := &SchedulerConfig{
		PolicyBatchSize: 100,
	}

	scheduler, err := NewSyntacticJobScheduler(observationCtx, sqlDB, config)
	require.NoError(t, err)

	db := database.NewDB(observationCtx.Logger, sqlDB)

	ctx := context.Background()

	tacosRepoId, tacosRepoName, _ := 1, "github.com/tangy/tacos", testutils.MakeCommit(1)
	empanadasRepoId, empanadasRepoName, _ := 2, "github.com/salty/empanadas", testutils.MakeCommit(2)
	mangosRepoId, mangosRepoName, _ := 3, "gitlab.com/juicy/mangos", testutils.MakeCommit(3)

	testutils.InsertRepo(t, db, tacosRepoId, tacosRepoName)
	testutils.InsertRepo(t, db, empanadasRepoId, empanadasRepoName)
	testutils.InsertRepo(t, db, mangosRepoId, mangosRepoName)

	setupRepoPolicies(t, ctx, db)

	fmt.Println("what?")

	err = scheduler.Schedule(observationCtx, ctx, time.Now())

	require.Error(t, err)

	// gsClient.ResolveRevisionFunc.SetDefaultHook(func(ctx context.Context, repo api.RepoName, rev string, options gitserver.ResolveRevisionOptions) (api.CommitID, error) {
	// 	isTacos := repo == api.RepoName(tacosRepoName) && rev == tacosCommit
	// 	isEmpanadas := repo == api.RepoName(empanadasRepoName) && rev == empanadasCommit

	// 	if isTacos || isEmpanadas {
	// 		return api.CommitID(rev), nil
	// 	} else {
	// 		return api.CommitID("what"), errors.New(fmt.Sprintf("Unexpected repo (`%s`) and revision (`%s`) requested from gitserver: ", repo, rev))
	// 	}
	// })

	// isQueued, err := jobStore.IsQueued(ctx, tacosRepoId, tacosCommit)
	// require.False(t, isQueued)
	// require.NoError(t, err)

	// // Happy path
	// scheduled, err := enqueuer.QueueIndexes(ctx, tacosRepoId, tacosCommit, EnqueueOptions{})

	// require.NoError(t, err)
	// require.True(t, len(scheduled) == 1)
	// require.Equal(t, scheduled[0].Commit, tacosCommit)
	// require.Equal(t, scheduled[0].RepositoryID, tacosRepoId)
	// require.Equal(t, scheduled[0].State, jobstore.Queued)
	// require.Equal(t, scheduled[0].RepositoryName, tacosRepoName)

	// // cannot schedule same repo+revision twice
	// result, err := enqueuer.QueueIndexes(ctx, tacosRepoId, tacosCommit, EnqueueOptions{})
	// require.Empty(t, result)
	// require.NoError(t, err)

	// // cannot schedule for non existent repositories
	// _, err = enqueuer.QueueIndexes(ctx, 250, tacosCommit, EnqueueOptions{})
	// require.Error(t, err)

	// // cannot schedule for non existent revisions
	// _, err = enqueuer.QueueIndexes(ctx, mangosRepoId, mangosCommit, EnqueueOptions{})
	// require.Error(t, err)

}

func setupRepoPolicies(t *testing.T, ctx context.Context, db database.DB) {

	// store :=
	query := `
		INSERT INTO lsif_configuration_policies (
			id,
			repository_id,
			name,
			type,
			pattern,
			repository_patterns,
			retention_enabled,
			retention_duration_hours,
			retain_intermediate_commits,
			syntactic_indexing_enabled,
			indexing_enabled,
			index_commit_max_age_hours,
			index_intermediate_commits
		) VALUES
			--                        							              ↙ retention_enabled
			--                        							              |    ↙ retention_duration_hours
			--                        							              |    |    ↙ retain_intermediate_commits
			--                        							              |    |    |     ↙ syntactic_indexing_enabled
			--                        							              |    |    |     |      ↙ indexing_enabled
			--                        							              |    |    |     |      |      ↙ index_commit_max_age_hours
			--                        							              |    |    |     |      |      |     ↙ index_intermediate_commits
			(1000, 2,    'policy  1 abc', 'GIT_TREE', '', null,              false, 0, false, true,  false,  0, false),

			-- This policy is here to specifically disable syntactic indexing for repo with ID=3

			(1003, 3,    'policy  3 bcd', 'GIT_TREE', '', null,              false, 0, false, false, false, 0,  false),

			-- This policy is to enable syntactic indexing for all repositories starting with 'github.com'

			(1100, NULL, 'policy 10 def', 'GIT_TREE', '', '{github.com/*}',  false, 0, false, true,  false, 0,  false)
	`
	if _, err := db.ExecContext(ctx, query); err != nil {
		t.Fatalf("unexpected error while inserting configuration policies: %s", err)
	}

	// for policyID, patterns := range map[int][]string{
	// 106: {"gitlab.com/*"},
	// 107: {"gitlab.com/*1"},
	// 108: {"gitlab.com/*2"},
	// 109: {"github.com/*"},
	// 110: {"github.com/*"},
	// } {
	// if err := store.UpdateReposMatchingPatterns(ctx, patterns, policyID, nil); err != nil {
	// 	t.Fatalf("unexpected error while updating repositories matching patterns: %s", err)
	// }
	// }

}
