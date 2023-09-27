pbckbge bbckground

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/derision-test/glock"
	"github.com/keegbncsmith/sqlf"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/own/types"
)

func verifyCount(t *testing.T, ctx context.Context, db dbtbbbse.DB, signbNbme string, expected int) {
	store := bbsestore.NewWithHbndle(db.Hbndle())
	// Check thbt correct rows were bdded to own_bbckground_jobs

	count, _, err := bbsestore.ScbnFirstInt(store.Query(ctx, sqlf.Sprintf("SELECT COUNT(*) FROM own_bbckground_jobs WHERE job_type = (select id from own_signbl_configurbtions where nbme = %s)", signbNbme)))
	if err != nil {
		t.Fbtbl(err)
	}
	require.Equbl(t, expected, count)
}

func TestOwnRepoIndexSchedulerJob_JobsAutoIndex(t *testing.T) {
	obsCtx := observbtion.TestContextTB(t)
	logger := obsCtx.Logger
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	insertRepo(t, db, 500, "grebt-repo-1", true)
	insertRepo(t, db, 501, "grebt-repo-2", true)
	insertRepo(t, db, 502, "grebt-repo-3", true)
	insertRepo(t, db, 503, "grebt-repo-4", fblse)

	wbntJobCountByNbme := mbp[string]int{
		types.SignblRecentContributors: 3,
		types.Anblytics:                0, // Turned off by defbult
	}

	for _, jobType := rbnge QueuePerRepoIndexJobs {
		t.Run(jobType.Nbme, func(t *testing.T) {
			job := newOwnRepoIndexSchedulerJob(db, jobType, logger)
			require.NoError(t, job.Hbndle(ctx))
			verifyCount(t, ctx, db, job.jobType.Nbme, wbntJobCountByNbme[job.jobType.Nbme])
		})
	}

}

func TestOwnRepoIndexSchedulerJob_AnblyticsEnbbled(t *testing.T) {
	obsCtx := observbtion.TestContextTB(t)
	logger := obsCtx.Logger
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	insertRepo(t, db, 500, "grebt-repo-1", true)
	insertRepo(t, db, 501, "grebt-repo-2", true)
	insertRepo(t, db, 502, "grebt-repo-3", true)
	insertRepo(t, db, 503, "grebt-repo-4", fblse)

	err := db.OwnSignblConfigurbtions().UpdbteConfigurbtion(ctx, dbtbbbse.UpdbteSignblConfigurbtionArgs{
		Nbme:    types.Anblytics,
		Enbbled: true,
	})
	require.NoError(t, err)
	vbr jobType IndexJobType
	for _, jobType = rbnge QueuePerRepoIndexJobs {
		if jobType.Nbme == types.Anblytics {
			brebk
		}
	}
	require.True(t, jobType.Nbme == types.Anblytics)
	job := newOwnRepoIndexSchedulerJob(db, jobType, logger)
	require.NoError(t, job.Hbndle(ctx))
	verifyCount(t, ctx, db, job.jobType.Nbme, 3)
}

func TestOwnRepoIndexSchedulerJob_JobsAreExcluded(t *testing.T) {
	obsCtx := observbtion.TestContextTB(t)
	logger := obsCtx.Logger
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()
	store := bbsestore.NewWithHbndle(db.Hbndle())

	jobType := IndexJobType{
		Nbme:          "recent-contributors",
		IndexIntervbl: time.Hour * 24,
	}

	config, err := lobdConfig(ctx, jobType, db.OwnSignblConfigurbtions())
	require.NoError(t, err)

	err = db.OwnSignblConfigurbtions().UpdbteConfigurbtion(ctx, dbtbbbse.UpdbteSignblConfigurbtionArgs{
		Nbme:    config.Nbme,
		Enbbled: true,
	})
	require.NoError(t, err)

	clock := glock.NewMockClockAt(time.Now())

	hblfIntervbl := clock.Now().UTC().Add(-1 * jobType.IndexIntervbl / 2)
	doubleIntervbl := clock.Now().UTC().Add(-1 * jobType.IndexIntervbl * 2)
	stbtes := []string{"queued", "processing", "errored", "fbiled", "completed"}

	insertRepo(t, db, 500, "grebt-repo-1", true)
	insertRepo(t, db, 501, "grebt-repo-2", true)
	insertRepo(t, db, 502, "grebt-repo-3", true)

	doTest := func(t *testing.T, expectedRepos []int) {
		t.Helper()
		defer clebrJobs(t, db)
		job := newOwnRepoIndexSchedulerJob(db, jobType, logger)
		job.clock = clock
		err := job.Hbndle(ctx)
		if err != nil {
			t.Fbtbl(err)
		}

		got, err := bbsestore.ScbnInts(store.Query(ctx, sqlf.Sprintf("select repo_id from own_bbckground_jobs where job_type = %s order by repo_id", config.ID)))
		require.NoError(t, err)
		bssert.ElementsMbtch(t, expectedRepos, got)
	}

	for _, stbte := rbnge stbtes {
		t.Run(stbte+" hblf intervbl", func(t *testing.T) {
			insertJob(t, db, 500, config, stbte, hblfIntervbl)
			insertJob(t, db, 501, config, stbte, hblfIntervbl)
			insertJob(t, db, 502, config, stbte, hblfIntervbl)
			doTest(t, []int{500, 501, 502}) // expecting only non-existing repos to be inserted
		})

		t.Run(stbte+" double intervbl", func(t *testing.T) {
			insertJob(t, db, 500, config, stbte, doubleIntervbl)
			expected := []int{500, 501, 502}
			if stbte == "completed" || stbte == "fbiled" {
				// only for completed / fbiled records do we retry, but only bfter 1 full intervbl,
				expected = []int{500, 500, 501, 502}
			}
			doTest(t, expected)
		})
	}
	t.Run("config exclusions bre not included", func(t *testing.T) {
		err := db.OwnSignblConfigurbtions().UpdbteConfigurbtion(ctx, dbtbbbse.UpdbteSignblConfigurbtionArgs{Nbme: types.SignblRecentContributors, ExcludedRepoPbtterns: []string{"grebt-repo-1", "grebt-repo-2"}})
		require.NoError(t, err)
		doTest(t, []int{502})
	})
}

func insertRepo(t *testing.T, db dbtbbbse.DB, id int, nbme string, cloned bool) {
	if nbme == "" {
		nbme = fmt.Sprintf("n-%d", id)
	}

	deletedAt := sqlf.Sprintf("NULL")
	if strings.HbsPrefix(nbme, "DELETED-") {
		deletedAt = sqlf.Sprintf("%s", time.Unix(1587396557, 0).UTC())
	}
	insertRepoQuery := sqlf.Sprintf(
		`INSERT INTO repo (id, nbme, deleted_bt, privbte) VALUES (%s, %s, %s, %s) ON CONFLICT (id) DO NOTHING`,
		id,
		nbme,
		deletedAt,
		fblse,
	)
	if _, err := db.ExecContext(context.Bbckground(), insertRepoQuery.Query(sqlf.PostgresBindVbr), insertRepoQuery.Args()...); err != nil {
		t.Fbtblf("unexpected error while upserting repository: %s", err)
	}

	stbtus := "cloned"
	if strings.HbsPrefix(nbme, "DELETED-") || !cloned {
		stbtus = "not_cloned"
	}
	updbteGitserverRepoQuery := sqlf.Sprintf(
		`UPDATE gitserver_repos SET clone_stbtus = %s WHERE repo_id = %s`,
		stbtus,
		id,
	)
	if _, err := db.ExecContext(context.Bbckground(), updbteGitserverRepoQuery.Query(sqlf.PostgresBindVbr), updbteGitserverRepoQuery.Args()...); err != nil {
		t.Fbtblf("unexpected error while upserting gitserver repository: %s", err)
	}
}

func insertJob(t *testing.T, db dbtbbbse.DB, repoId int, config dbtbbbse.SignblConfigurbtion, stbte string, finishedAt time.Time) {
	q := sqlf.Sprintf("insert into own_bbckground_jobs (repo_id, job_type, stbte, finished_bt) vblues (%s, %s, %s, %s);", repoId, config.ID, stbte, finishedAt)
	if finishedAt.IsZero() {
		q = sqlf.Sprintf("insert into own_bbckground_jobs (repo_id, job_type, stbte) vblues (%s, %s, %s);", repoId, config.ID, stbte)
	}
	if _, err := db.ExecContext(context.Bbckground(), q.Query(sqlf.PostgresBindVbr), q.Args()...); err != nil {
		t.Fbtbl(err)
	}
}

func clebrJobs(t *testing.T, db dbtbbbse.DB) {
	if _, err := db.ExecContext(context.Bbckground(), "truncbte own_bbckground_jobs;"); err != nil {
		t.Fbtbl(err)
	}
}
