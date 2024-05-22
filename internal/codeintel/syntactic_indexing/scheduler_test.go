package syntactic_indexing

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/policies"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/reposcheduler"
	codeintelshared "github.com/sourcegraph/sourcegraph/internal/codeintel/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/syntactic_indexing/internal/testutils"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/syntactic_indexing/jobstore"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/stretchr/testify/require"
)

func TestSyntacticIndexingScheduler(t *testing.T) {
	observationCtx := observation.TestContextTB(t)

	// Bootstrap scheduler
	sqlDB := dbtest.NewDB(t)
	db := database.NewDB(observationCtx.Logger, sqlDB)
	config := &SchedulerConfig{
		PolicyBatchSize:     100,
		RepositoryBatchSize: 2500,
	}
	gitserverClient := gitserver.NewMockClient()
	scheduler, jobStore := bootstrapScheduler(t, observationCtx, sqlDB, gitserverClient, config)

	ctx := context.Background()

	// Setup repositories
	tacosRepoId, tacosRepoName, tacosCommit := 1, "github.com/tangy/tacos", testutils.MakeCommit(1)
	empanadasRepoId, empanadasRepoName, empanadasCommit := 2, "github.com/salty/empanadas", testutils.MakeCommit(2)
	mangosRepoId, mangosRepoName, _ := 3, "gitlab.com/juicy/mangos", testutils.MakeCommit(3)
	testutils.InsertRepo(t, db, tacosRepoId, tacosRepoName)
	testutils.InsertRepo(t, db, empanadasRepoId, empanadasRepoName)
	testutils.InsertRepo(t, db, mangosRepoId, mangosRepoName)

	setupRepoPolicies(t, ctx, db)

	gitserverClient.ResolveRevisionFunc.SetDefaultHook(func(ctx context.Context, repo api.RepoName, rev string, options gitserver.ResolveRevisionOptions) (api.CommitID, error) {
		isTacos := repo == api.RepoName(tacosRepoName) && rev == tacosCommit
		isEmpanadas := repo == api.RepoName(empanadasRepoName) && rev == empanadasCommit

		if isTacos || isEmpanadas {
			return api.CommitID(rev), nil
		} else {
			return api.CommitID("what"), errors.New(fmt.Sprintf("Unexpected repo (`%s`) and revision (`%s`) requested from gitserver: ", repo, rev))
		}
	})

	gitserverClient.ListRefsFunc.SetDefaultHook(func(ctx context.Context, repoName api.RepoName, opts gitserver.ListRefsOpts) ([]gitdomain.Ref, error) {

		fmt.Println(repoName)

		if string(repoName) == empanadasRepoName {
			return []gitdomain.Ref{
				{
					Name:     "refs/head/main",
					Type:     gitdomain.RefTypeBranch,
					CommitID: api.CommitID(empanadasCommit),
					IsHead:   true,
				},
			}, nil
		} else {
			return nil, errors.New(fmt.Sprintf("Unexpected repo (`%s`) requested from gitserver's ListRef", repoName))
		}

	})

	err := scheduler.Schedule(observationCtx, ctx, time.Now())
	require.NoError(t, err)

	require.Equal(t, 2, unwrap(jobStore.DBWorkerStore().QueuedCount(ctx, false))(t))
}

func unwrap[T any](v T, err error) func(*testing.T) T {
	return func(t *testing.T) T {
		require.NoError(t, err)
		return v
	}
}

func bootstrapScheduler(t *testing.T, observationCtx *observation.Context,
	sqlDB *sql.DB, gitserverClient gitserver.Client,
	config *SchedulerConfig) (SyntacticJobScheduler, jobstore.SyntacticIndexingJobStore) {

	db := database.NewDB(observationCtx.Logger, sqlDB)
	codeIntelDB := codeintelshared.NewCodeIntelDB(observationCtx.Logger, sqlDB)

	uploadsSvc := uploads.NewService(observationCtx, db, codeIntelDB, gitserverClient.Scoped("uploads"))
	policiesSvc := policies.NewService(observationCtx, db, uploadsSvc, gitserverClient.Scoped("policies"))

	schedulerConfig.Load()

	matcher := policies.NewMatcher(
		gitserverClient,
		policies.IndexingExtractor,
		true,
		true,
	)

	repoSchedulingStore := reposcheduler.NewSyntacticStore(observationCtx, db)
	repoSchedulingSvc := reposcheduler.NewService(repoSchedulingStore)

	jobStore, err := jobstore.NewStoreWithDB(observationCtx, sqlDB)
	require.NoError(t, err)

	repoStore := db.Repos()

	enqueuer := NewIndexEnqueuer(observationCtx, jobStore, repoSchedulingStore, repoStore, gitserverClient)

	scheduler, err := NewSyntaticJobScheduler(repoSchedulingSvc, *matcher, *policiesSvc, repoStore, enqueuer, *config)

	require.NoError(t, err)

	return scheduler, jobStore
}

func setupRepoPolicies(t *testing.T, ctx context.Context, db database.DB) {

	if _, err := db.ExecContext(context.Background(), `TRUNCATE lsif_configuration_policies`); err != nil {
		t.Fatalf("unexpected error while inserting configuration policies: %s", err)
	}

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
}
