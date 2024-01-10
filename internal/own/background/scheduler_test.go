package background

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/derision-test/glock"
	"github.com/keegancsmith/sqlf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/own/types"
)

func verifyCount(t *testing.T, ctx context.Context, db database.DB, signaName string, expected int) {
	store := basestore.NewWithHandle(db.Handle())
	// Check that correct rows were added to own_background_jobs

	count, _, err := basestore.ScanFirstInt(store.Query(ctx, sqlf.Sprintf("SELECT COUNT(*) FROM own_background_jobs WHERE job_type = (select id from own_signal_configurations where name = %s)", signaName)))
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, expected, count)
}

func TestOwnRepoIndexSchedulerJob_JobsAutoIndex(t *testing.T) {
	obsCtx := observation.TestContextTB(t)
	logger := obsCtx.Logger
	db := database.NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()

	insertRepo(t, db, 500, "great-repo-1", true)
	insertRepo(t, db, 501, "great-repo-2", true)
	insertRepo(t, db, 502, "great-repo-3", true)
	insertRepo(t, db, 503, "great-repo-4", false)

	wantJobCountByName := map[string]int{
		types.SignalRecentContributors: 3,
		types.Analytics:                0, // Turned off by default
	}

	for _, jobType := range QueuePerRepoIndexJobs {
		t.Run(jobType.Name, func(t *testing.T) {
			job := newOwnRepoIndexSchedulerJob(db, jobType, logger)
			require.NoError(t, job.Handle(ctx))
			verifyCount(t, ctx, db, job.jobType.Name, wantJobCountByName[job.jobType.Name])
		})
	}

}

func TestOwnRepoIndexSchedulerJob_AnalyticsEnabled(t *testing.T) {
	obsCtx := observation.TestContextTB(t)
	logger := obsCtx.Logger
	db := database.NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()

	insertRepo(t, db, 500, "great-repo-1", true)
	insertRepo(t, db, 501, "great-repo-2", true)
	insertRepo(t, db, 502, "great-repo-3", true)
	insertRepo(t, db, 503, "great-repo-4", false)

	err := db.OwnSignalConfigurations().UpdateConfiguration(ctx, database.UpdateSignalConfigurationArgs{
		Name:    types.Analytics,
		Enabled: true,
	})
	require.NoError(t, err)
	var jobType IndexJobType
	for _, jobType = range QueuePerRepoIndexJobs {
		if jobType.Name == types.Analytics {
			break
		}
	}
	require.True(t, jobType.Name == types.Analytics)
	job := newOwnRepoIndexSchedulerJob(db, jobType, logger)
	require.NoError(t, job.Handle(ctx))
	verifyCount(t, ctx, db, job.jobType.Name, 3)
}

func TestOwnRepoIndexSchedulerJob_JobsAreExcluded(t *testing.T) {
	obsCtx := observation.TestContextTB(t)
	logger := obsCtx.Logger
	db := database.NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()
	store := basestore.NewWithHandle(db.Handle())

	jobType := IndexJobType{
		Name:          "recent-contributors",
		IndexInterval: time.Hour * 24,
	}

	config, err := loadConfig(ctx, jobType, db.OwnSignalConfigurations())
	require.NoError(t, err)

	err = db.OwnSignalConfigurations().UpdateConfiguration(ctx, database.UpdateSignalConfigurationArgs{
		Name:    config.Name,
		Enabled: true,
	})
	require.NoError(t, err)

	clock := glock.NewMockClockAt(time.Now())

	halfInterval := clock.Now().UTC().Add(-1 * jobType.IndexInterval / 2)
	doubleInterval := clock.Now().UTC().Add(-1 * jobType.IndexInterval * 2)
	states := []string{"queued", "processing", "errored", "failed", "completed"}

	insertRepo(t, db, 500, "great-repo-1", true)
	insertRepo(t, db, 501, "great-repo-2", true)
	insertRepo(t, db, 502, "great-repo-3", true)

	doTest := func(t *testing.T, expectedRepos []int) {
		t.Helper()
		defer clearJobs(t, db)
		job := newOwnRepoIndexSchedulerJob(db, jobType, logger)
		job.clock = clock
		err := job.Handle(ctx)
		if err != nil {
			t.Fatal(err)
		}

		got, err := basestore.ScanInts(store.Query(ctx, sqlf.Sprintf("select repo_id from own_background_jobs where job_type = %s order by repo_id", config.ID)))
		require.NoError(t, err)
		assert.ElementsMatch(t, expectedRepos, got)
	}

	for _, state := range states {
		t.Run(state+" half interval", func(t *testing.T) {
			insertJob(t, db, 500, config, state, halfInterval)
			insertJob(t, db, 501, config, state, halfInterval)
			insertJob(t, db, 502, config, state, halfInterval)
			doTest(t, []int{500, 501, 502}) // expecting only non-existing repos to be inserted
		})

		t.Run(state+" double interval", func(t *testing.T) {
			insertJob(t, db, 500, config, state, doubleInterval)
			expected := []int{500, 501, 502}
			if state == "completed" || state == "failed" {
				// only for completed / failed records do we retry, but only after 1 full interval,
				expected = []int{500, 500, 501, 502}
			}
			doTest(t, expected)
		})
	}
	t.Run("config exclusions are not included", func(t *testing.T) {
		err := db.OwnSignalConfigurations().UpdateConfiguration(ctx, database.UpdateSignalConfigurationArgs{Name: types.SignalRecentContributors, ExcludedRepoPatterns: []string{"great-repo-1", "great-repo-2"}})
		require.NoError(t, err)
		doTest(t, []int{502})
	})
}

func insertRepo(t *testing.T, db database.DB, id int, name string, cloned bool) {
	if name == "" {
		name = fmt.Sprintf("n-%d", id)
	}

	deletedAt := sqlf.Sprintf("NULL")
	if strings.HasPrefix(name, "DELETED-") {
		deletedAt = sqlf.Sprintf("%s", time.Unix(1587396557, 0).UTC())
	}
	insertRepoQuery := sqlf.Sprintf(
		`INSERT INTO repo (id, name, deleted_at, private) VALUES (%s, %s, %s, %s) ON CONFLICT (id) DO NOTHING`,
		id,
		name,
		deletedAt,
		false,
	)
	if _, err := db.ExecContext(context.Background(), insertRepoQuery.Query(sqlf.PostgresBindVar), insertRepoQuery.Args()...); err != nil {
		t.Fatalf("unexpected error while upserting repository: %s", err)
	}

	status := "cloned"
	if strings.HasPrefix(name, "DELETED-") || !cloned {
		status = "not_cloned"
	}
	updateGitserverRepoQuery := sqlf.Sprintf(
		`UPDATE gitserver_repos SET clone_status = %s WHERE repo_id = %s`,
		status,
		id,
	)
	if _, err := db.ExecContext(context.Background(), updateGitserverRepoQuery.Query(sqlf.PostgresBindVar), updateGitserverRepoQuery.Args()...); err != nil {
		t.Fatalf("unexpected error while upserting gitserver repository: %s", err)
	}
}

func insertJob(t *testing.T, db database.DB, repoId int, config database.SignalConfiguration, state string, finishedAt time.Time) {
	q := sqlf.Sprintf("insert into own_background_jobs (repo_id, job_type, state, finished_at) values (%s, %s, %s, %s);", repoId, config.ID, state, finishedAt)
	if finishedAt.IsZero() {
		q = sqlf.Sprintf("insert into own_background_jobs (repo_id, job_type, state) values (%s, %s, %s);", repoId, config.ID, state)
	}
	if _, err := db.ExecContext(context.Background(), q.Query(sqlf.PostgresBindVar), q.Args()...); err != nil {
		t.Fatal(err)
	}
}

func clearJobs(t *testing.T, db database.DB) {
	if _, err := db.ExecContext(context.Background(), "truncate own_background_jobs;"); err != nil {
		t.Fatal(err)
	}
}
