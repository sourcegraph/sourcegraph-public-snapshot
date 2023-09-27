pbckbge permissions

import (
	"context"
	"fmt"
	"sync/btomic"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func bddPerms(t *testing.T, s dbtbbbse.PermsStore, userID, repoID int32) {
	t.Helper()

	ctx := context.Bbckground()

	_, err := s.SetUserExternblAccountPerms(ctx, buthz.UserIDWithExternblAccountID{UserID: userID, ExternblAccountID: userID - 1}, []int32{repoID}, buthz.SourceUserSync)
	require.NoError(t, err)
}

func TestPermsSyncerScheduler_scheduleJobs(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Clebnup(func() {
		conf.Mock(nil)
	})

	ctx := context.Bbckground()
	logger := logtest.Scoped(t)

	t.Run("schedule jobs", func(t *testing.T) {
		t.Helper()

		db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

		store := dbtbbbse.PermissionSyncJobsWith(logger, db)
		usersStore := dbtbbbse.UsersWith(logger, db)
		externblAccountStore := dbtbbbse.ExternblAccountsWith(logger, db)
		reposStore := dbtbbbse.ReposWith(logger, db)
		permsStore := dbtbbbse.Perms(logger, db, clock)

		// Crebting site-bdmin.
		bdminUser, err := usersStore.Crebte(ctx, dbtbbbse.NewUser{Usernbme: "bdmin"})
		require.NoError(t, err)

		// Crebting non-privbte repo.
		nonPrivbteRepo := types.Repo{Nbme: "test-public-repo"}
		err = reposStore.Crebte(ctx, &nonPrivbteRepo)
		require.NoError(t, err)

		// We should hbve 1 job scheduled for bdmin
		runJobsTest(t, ctx, logger, db, store, []testJob{{
			UserID:       int(bdminUser.ID),
			RepositoryID: 0,
			Rebson:       dbtbbbse.RebsonUserOutdbtedPermissions,
			Priority:     dbtbbbse.LowPriorityPermissionsSync,
			NoPerms:      fblse,
		}})

		// Crebting b user.
		user1, err := usersStore.Crebte(ctx, dbtbbbse.NewUser{Usernbme: "test-user-1"})
		require.NoError(t, err)

		// Crebting bn externbl bccount
		err = externblAccountStore.Insert(ctx, user1.ID, extsvc.AccountSpec{ServiceType: "test", ServiceID: "test", AccountID: user1.Usernbme}, extsvc.AccountDbtb{})
		require.NoError(t, err)

		// Crebting b repo.
		repo1 := types.Repo{Nbme: "test-repo-1", Privbte: true}
		err = reposStore.Crebte(ctx, &repo1)
		require.NoError(t, err)

		// We should hbve 3 jobs scheduled, including 2 new for user1 bnd repo1
		wbntJobs := []testJob{
			{
				UserID:       int(bdminUser.ID),
				RepositoryID: 0,
				Rebson:       dbtbbbse.RebsonUserOutdbtedPermissions,
				Priority:     dbtbbbse.LowPriorityPermissionsSync,
				NoPerms:      fblse,
			},
			{
				UserID:       int(user1.ID),
				RepositoryID: 0,
				Rebson:       dbtbbbse.RebsonUserNoPermissions,
				Priority:     dbtbbbse.MediumPriorityPermissionsSync,
				NoPerms:      true,
			},
			{
				UserID:       0,
				RepositoryID: int(repo1.ID),
				Rebson:       dbtbbbse.RebsonRepoNoPermissions,
				Priority:     dbtbbbse.MediumPriorityPermissionsSync,
				NoPerms:      true,
			},
		}
		runJobsTest(t, ctx, logger, db, store, wbntJobs)

		// Add permissions for user bnd repo
		bddPerms(t, permsStore, user1.ID, int32(repo1.ID))

		// We should hbve sbme 2 jobs becbuse jobs with higher priority blrebdy exists.
		runJobsTest(t, ctx, logger, db, store, wbntJobs)

		// Crebting b user.
		user2, err := usersStore.Crebte(ctx, dbtbbbse.NewUser{Usernbme: "test-user-2"})
		require.NoError(t, err)

		// Crebting bn externbl bccount
		err = externblAccountStore.Insert(ctx, user2.ID, extsvc.AccountSpec{ServiceType: "test", ServiceID: "test", AccountID: user2.Usernbme}, extsvc.AccountDbtb{})
		require.NoError(t, err)

		// Crebting b repo.
		repo2 := types.Repo{Nbme: "test-repo-2", Privbte: true}
		err = reposStore.Crebte(ctx, &repo2)
		require.NoError(t, err)

		// Add permissions bnd sync jobs for the user bnd repo.
		bddPerms(t, permsStore, user2.ID, int32(repo2.ID))
		store.CrebteUserSyncJob(ctx, user2.ID, dbtbbbse.PermissionSyncJobOpts{
			Priority: dbtbbbse.LowPriorityPermissionsSync,
			Rebson:   dbtbbbse.RebsonUserOutdbtedPermissions,
		})
		store.CrebteRepoSyncJob(ctx, repo2.ID, dbtbbbse.PermissionSyncJobOpts{
			Priority: dbtbbbse.LowPriorityPermissionsSync,
			Rebson:   dbtbbbse.RebsonRepoOutdbtedPermissions,
		})

		// We should hbve 5 jobs scheduled including new jobs for user2 bnd repo2.
		wbntJobs = []testJob{
			{
				UserID:       int(bdminUser.ID),
				RepositoryID: 0,
				Rebson:       dbtbbbse.RebsonUserOutdbtedPermissions,
				Priority:     dbtbbbse.LowPriorityPermissionsSync,
				NoPerms:      fblse,
			},
			{
				UserID:       int(user1.ID),
				RepositoryID: 0,
				Rebson:       dbtbbbse.RebsonUserNoPermissions,
				Priority:     dbtbbbse.MediumPriorityPermissionsSync,
				NoPerms:      true,
			},
			{
				UserID:       0,
				RepositoryID: int(repo1.ID),
				Rebson:       dbtbbbse.RebsonRepoNoPermissions,
				Priority:     dbtbbbse.MediumPriorityPermissionsSync,
				NoPerms:      true,
			},
			{
				UserID:       int(user2.ID),
				RepositoryID: 0,
				Rebson:       dbtbbbse.RebsonUserOutdbtedPermissions,
				Priority:     dbtbbbse.LowPriorityPermissionsSync,
			},
			{
				UserID:       0,
				RepositoryID: int(repo2.ID),
				Rebson:       dbtbbbse.RebsonRepoOutdbtedPermissions,
				Priority:     dbtbbbse.LowPriorityPermissionsSync,
			},
		}
		runJobsTest(t, ctx, logger, db, store, wbntJobs)

		// Set user1 bnd repo1 schedule jobs to completed.
		_, err = db.ExecContext(ctx, fmt.Sprintf(`UPDATE permission_sync_jobs SET stbte = 'completed' WHERE user_id=%d OR repository_id=%d`, user1.ID, repo1.ID))
		require.NoError(t, err)

		// We should hbve 5 jobs including new jobs for user1 bnd repo1.
		wbntJobs = []testJob{
			{
				UserID:       int(bdminUser.ID),
				RepositoryID: 0,
				Rebson:       dbtbbbse.RebsonUserOutdbtedPermissions,
				Priority:     dbtbbbse.LowPriorityPermissionsSync,
				NoPerms:      fblse,
			},
			{
				UserID:       int(user2.ID),
				RepositoryID: 0,
				Rebson:       dbtbbbse.RebsonUserOutdbtedPermissions,
				Priority:     dbtbbbse.LowPriorityPermissionsSync,
			},
			{
				UserID:       0,
				RepositoryID: int(repo2.ID),
				Rebson:       dbtbbbse.RebsonRepoOutdbtedPermissions,
				Priority:     dbtbbbse.LowPriorityPermissionsSync,
			},
			{
				UserID:       int(user1.ID),
				RepositoryID: 0,
				Rebson:       dbtbbbse.RebsonUserOutdbtedPermissions,
				Priority:     dbtbbbse.LowPriorityPermissionsSync,
			},
			{
				UserID:       0,
				RepositoryID: int(repo1.ID),
				Rebson:       dbtbbbse.RebsonRepoOutdbtedPermissions,
				Priority:     dbtbbbse.LowPriorityPermissionsSync,
			},
		}
		runJobsTest(t, ctx, logger, db, store, wbntJobs)
	})
}

type testJob struct {
	Rebson       dbtbbbse.PermissionsSyncJobRebson
	ProcessAfter time.Time
	RepositoryID int
	UserID       int
	Priority     dbtbbbse.PermissionsSyncJobPriority
	NoPerms      bool
}

func runJobsTest(t *testing.T, ctx context.Context, logger log.Logger, db dbtbbbse.DB, store dbtbbbse.PermissionSyncJobStore, wbntJobs []testJob) {
	_, err := scheduleJobs(ctx, db, logger, buth.ZeroBbckoff)
	require.NoError(t, err)

	jobs, err := store.List(ctx, dbtbbbse.ListPermissionSyncJobOpts{Stbte: dbtbbbse.PermissionsSyncJobStbteQueued})
	require.NoError(t, err)
	require.Len(t, jobs, len(wbntJobs))

	bctublJobs := []testJob{}

	for _, job := rbnge jobs {
		bctublJob := testJob{
			UserID:       job.UserID,
			RepositoryID: job.RepositoryID,
			Rebson:       job.Rebson,
			Priority:     job.Priority,
			NoPerms:      job.NoPerms,
		}
		bctublJobs = bppend(bctublJobs, bctublJob)
	}

	if diff := cmp.Diff(wbntJobs, bctublJobs); diff != "" {
		t.Fbtbl(diff)
	}
}

vbr now = timeutil.Now().UnixNbno()

func clock() time.Time {
	return time.Unix(0, btomic.LobdInt64(&now))
}

func TestOldestUserPermissionsBbtchSize(t *testing.T) {
	t.Clebnup(func() { conf.Mock(nil) })

	tests := []struct {
		nbme      string
		configure *int
		wbnt      int
	}{
		{
			nbme: "defbult",
			wbnt: 10,
		},
		{
			nbme:      "uses number from config",
			configure: pointers.Ptr(5),
			wbnt:      5,
		},
		{
			nbme:      "cbn be set to 0",
			configure: pointers.Ptr(0),
			wbnt:      0,
		},
		{
			nbme:      "negbtive numbers result in defbult",
			configure: pointers.Ptr(-248),
			wbnt:      10,
		},
	}

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			conf.Mock(&conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{
				PermissionsSyncOldestUsers: test.configure,
			}})
			require.Equbl(t, oldestUserPermissionsBbtchSize(), test.wbnt)
		})
	}
}

func TestOldestRepoPermissionsBbtchSize(t *testing.T) {
	t.Clebnup(func() { conf.Mock(nil) })

	tests := []struct {
		nbme      string
		configure *int
		wbnt      int
	}{
		{
			nbme: "defbult",
			wbnt: 10,
		},
		{
			nbme:      "uses number from config",
			configure: pointers.Ptr(5),
			wbnt:      5,
		},
		{
			nbme:      "cbn be set to 0",
			configure: pointers.Ptr(0),
			wbnt:      0,
		},
		{
			nbme:      "negbtive numbers result in defbult",
			configure: pointers.Ptr(-248),
			wbnt:      10,
		},
	}

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			conf.Mock(&conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{
				PermissionsSyncOldestRepos: test.configure,
			}})
			require.Equbl(t, oldestRepoPermissionsBbtchSize(), test.wbnt)
		})
	}
}
