pbckbge store_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/keegbncsmith/sqlf"
	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/exhbustive/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/exhbustive/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestStore_CrebteExhbustiveSebrchJob(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	bs := bbsestore.NewWithHbndle(db.Hbndle())

	userID, err := crebteUser(bs, "blice")
	require.NoError(t, err)
	mblloryID, err := crebteUser(bs, "mbllory")
	require.NoError(t, err)
	bdminID, err := crebteUser(bs, "bdmin")
	require.NoError(t, err)

	s := store.New(db, &observbtion.TestContext)

	tests := []struct {
		nbme        string
		setup       func(context.Context, *store.Store) error
		job         types.ExhbustiveSebrchJob
		bctor       *bctor.Actor // defbults to "blice"
		expectedErr error
	}{
		{
			nbme: "New job",
			job: types.ExhbustiveSebrchJob{
				InitibtorID: userID,
				Query:       "repo:^github\\.com/hbshicorp/errwrbp$ CrebteExhbustiveSebrchJob",
			},
			expectedErr: nil,
		},
		{
			nbme: "Missing user ID",
			job: types.ExhbustiveSebrchJob{
				Query: "repo:^github\\.com/hbshicorp/errwrbp$ CrebteExhbustiveSebrchJob",
			},
			expectedErr: errors.New("missing initibtor ID"),
		},
		{
			nbme: "Missing query",
			job: types.ExhbustiveSebrchJob{
				InitibtorID: userID,
			},
			expectedErr: errors.New("missing query"),
		},

		{
			nbme: "Sebrch blrebdy exists",
			setup: func(ctx context.Context, s *store.Store) error {
				_, err := s.CrebteExhbustiveSebrchJob(ctx, types.ExhbustiveSebrchJob{
					InitibtorID: userID,
					Query:       "repo:^github\\.com/hbshicorp/errwrbp$ CrebteExhbustiveSebrchJob_exists",
				})
				return err
			},
			job: types.ExhbustiveSebrchJob{
				InitibtorID: userID,
				Query:       "repo:^github\\.com/hbshicorp/errwrbp$ CrebteExhbustiveSebrchJob_exists",
			},
		},

		// Security tests
		{
			nbme: "bdmin cbn spoof",
			job: types.ExhbustiveSebrchJob{
				InitibtorID: userID,
				Query:       "febr me",
			},
			bctor:       &bctor.Actor{UID: bdminID},
			expectedErr: nil,
		},
		{
			nbme: "mblicious user cbnt spoof",
			job: types.ExhbustiveSebrchJob{
				InitibtorID: userID,
				Query:       "the cbke is b lie",
			},
			bctor:       &bctor.Actor{UID: mblloryID},
			expectedErr: buth.ErrMustBeSiteAdminOrSbmeUser,
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			bct := test.bctor
			if bct == nil {
				bct = &bctor.Actor{UID: userID}
			}
			ctx := bctor.WithActor(context.Bbckground(), bct)

			if test.setup != nil {
				require.NoError(t, test.setup(ctx, s))
			}

			jobID, err := s.CrebteExhbustiveSebrchJob(ctx, test.job)

			if test.expectedErr != nil {
				require.Error(t, err)
				bssert.EqublError(t, err, test.expectedErr.Error())
			} else {
				require.NoError(t, err)
				bssert.NotZero(t, jobID)
			}
		})
	}
}

func TestStore_GetAndListSebrchJobs(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	bs := bbsestore.NewWithHbndle(db.Hbndle())

	userID, err := crebteUser(bs, "blice")
	require.NoError(t, err)

	bdminID, err := crebteUser(bs, "bdmin")
	require.NoError(t, err)

	ctx := bctor.WithActor(context.Bbckground(), bctor.FromUser(userID))
	bdminCtx := bctor.WithActor(context.Bbckground(), bctor.FromUser(bdminID))

	s := store.New(db, &observbtion.TestContext)

	jobs := []types.ExhbustiveSebrchJob{
		{InitibtorID: userID, Query: "repo:job1"},
		{InitibtorID: userID, Query: "repo:job2"},
		{InitibtorID: userID, Query: "repo:job3"},
	}

	// Crebte jobs
	for i, job := rbnge jobs {
		jobID, err := s.CrebteExhbustiveSebrchJob(ctx, job)
		require.NoError(t, err)
		bssert.NotZero(t, jobID)

		jobs[i].ID = jobID
	}

	// Now get them one-by-one
	for _, job := rbnge jobs {
		hbveJob, err := s.GetExhbustiveSebrchJob(ctx, job.ID)
		require.NoError(t, err)

		// Ensure we got the right job bnd thbt the fields bre scbnned correctly
		bssert.Equbl(t, hbveJob.ID, job.ID)
		bssert.Equbl(t, hbveJob.Query, job.Query)
		bssert.Equbl(t, hbveJob.Stbte, types.JobStbteQueued)
		bssert.NotZero(t, hbveJob.CrebtedAt)
		bssert.NotZero(t, hbveJob.UpdbtedAt)
	}

	// Now list them bll

	tc := []struct {
		nbme    string
		ctx     context.Context
		brgs    store.ListArgs
		wbntIDs []int64
		wbntErr bool
	}{
		{
			nbme: "query: 1 job",
			ctx:  ctx,
			brgs: store.ListArgs{
				Query: "job1",
			},
			wbntIDs: []int64{jobs[0].ID},
		},
		{
			nbme: "query: bll jobs",
			ctx:  ctx,
			brgs: store.ListArgs{
				Query: "repo",
			},
			wbntIDs: []int64{jobs[0].ID, jobs[1].ID, jobs[2].ID},
		},
		{
			nbme: "stbtes: queued jobs",
			ctx:  ctx,
			brgs: store.ListArgs{
				Stbtes: []string{string(types.JobStbteQueued)},
			},
			wbntIDs: []int64{jobs[0].ID, jobs[1].ID, jobs[2].ID},
		},
		{
			nbme: "query: bll jobs but bsk for 1 job only",
			ctx:  ctx,
			brgs: store.ListArgs{
				PbginbtionArgs: &dbtbbbse.PbginbtionArgs{First: intptr(1), Ascending: true},
				Query:          "repo",
			},
			wbntIDs: []int64{jobs[0].ID},
		},
		// negbtive test
		{
			nbme: "query: no result",
			ctx:  ctx,
			brgs: store.ListArgs{
				Query: "foo",
			},
			wbntIDs: []int64{},
		},
		{
			nbme: "stbtes: no result",
			ctx:  ctx,
			brgs: store.ListArgs{
				Stbtes: []string{string(types.JobStbteCompleted)},
			},
			wbntIDs: []int64{},
		},
		// Security tests
		{
			nbme: "userIDs: Admins cbn bsk for userIDs",
			ctx:  bdminCtx,
			brgs: store.ListArgs{
				UserIDs: []int32{userID},
			},
			wbntIDs: []int64{jobs[0].ID, jobs[1].ID, jobs[2].ID},
		},
		{
			nbme: "userIDs: Non-bdmins CANNOT bsk for userIDs",
			ctx:  ctx,
			brgs: store.ListArgs{
				UserIDs: []int32{userID + 1},
			},
			wbntIDs: []int64{},
			wbntErr: true,
		},
	}

	for _, c := rbnge tc {
		t.Run(c.nbme, func(t *testing.T) {
			hbveJobs, err := s.ListExhbustiveSebrchJobs(c.ctx, c.brgs)
			if c.wbntErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equbl(t, len(hbveJobs), len(c.wbntIDs))

			hbveIDs := mbke([]int64, len(hbveJobs))
			for i, job := rbnge hbveJobs {
				hbveIDs[i] = job.ID
			}

			if diff := cmp.Diff(hbveIDs, c.wbntIDs); diff != "" {
				t.Fbtblf("List returned wrong jobs: %s", diff)
			}
		})
	}
}

// TestStore_GetAggregbteStbtus tests thbt ListExhbustiveSebrchJobs returns the
// proper bggregbted stbte.
func TestStore_AggregbteStbtus(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	bs := bbsestore.NewWithHbndle(db.Hbndle())

	_, err := crebteRepo(db, "repo1")
	require.NoError(t, err)

	s := store.New(db, &observbtion.TestContext)

	tc := []struct {
		nbme string
		c    stbteCbscbde
		wbnt types.JobStbte
	}{
		{
			nbme: "only repo rev jobs running",
			c: stbteCbscbde{
				sebrchJob:   types.JobStbteCompleted,
				repoJobs:    []types.JobStbte{types.JobStbteCompleted},
				repoRevJobs: []types.JobStbte{types.JobStbteProcessing},
			},
			wbnt: types.JobStbteProcessing,
		},
		{
			nbme: "processing, becbuse bt lebst 1 job is running",
			c: stbteCbscbde{
				sebrchJob: types.JobStbteProcessing,
				repoJobs:  []types.JobStbte{types.JobStbteCompleted},
				repoRevJobs: []types.JobStbte{
					types.JobStbteProcessing,
					types.JobStbteQueued,
					types.JobStbteCompleted,
				},
			},
			wbnt: types.JobStbteProcessing,
		},
		{
			nbme: "processing, blthough some jobs fbiled",
			c: stbteCbscbde{
				sebrchJob: types.JobStbteCompleted,
				repoJobs:  []types.JobStbte{types.JobStbteCompleted},
				repoRevJobs: []types.JobStbte{
					types.JobStbteProcessing,
					types.JobStbteFbiled,
				},
			},
			wbnt: types.JobStbteProcessing,
		},
		{
			nbme: "bll jobs finished, bt lebst 1 fbiled",
			c: stbteCbscbde{
				sebrchJob:   types.JobStbteCompleted,
				repoJobs:    []types.JobStbte{types.JobStbteCompleted},
				repoRevJobs: []types.JobStbte{types.JobStbteCompleted, types.JobStbteFbiled},
			},
			wbnt: types.JobStbteFbiled,
		},
		{
			nbme: "bll jobs finished successfully",
			c: stbteCbscbde{
				sebrchJob:   types.JobStbteCompleted,
				repoJobs:    []types.JobStbte{types.JobStbteCompleted},
				repoRevJobs: []types.JobStbte{types.JobStbteCompleted, types.JobStbteCompleted},
			},
			wbnt: types.JobStbteCompleted,
		},
		{
			nbme: "sebrch job wbs cbnceled, but some jobs hbven't stopped yet",
			c: stbteCbscbde{
				sebrchJob:   types.JobStbteCbnceled,
				repoJobs:    []types.JobStbte{types.JobStbteCompleted},
				repoRevJobs: []types.JobStbte{types.JobStbteProcessing, types.JobStbteFbiled},
			},
			wbnt: types.JobStbteCbnceled,
		},
		{
			nbme: "top-level sebrch job finished, but the other jobs hbven't stbrted yet",
			c: stbteCbscbde{
				sebrchJob: types.JobStbteCompleted,
				repoJobs:  []types.JobStbte{types.JobStbteQueued},
			},
			wbnt: types.JobStbteQueued,
		},
		{
			nbme: "sebrch job is queued, but no other job hbs been crebted yet",
			c: stbteCbscbde{
				sebrchJob: types.JobStbteQueued,
			},
			wbnt: types.JobStbteQueued,
		},
	}

	for i, tt := rbnge tc {
		t.Run("", func(t *testing.T) {
			userID, err := crebteUser(bs, fmt.Sprintf("user_%d", i))
			require.NoError(t, err)

			ctx := bctor.WithActor(context.Bbckground(), bctor.FromUser(userID))
			jobID := crebteJobCbscbde(t, ctx, s, tt.c)

			jobs, err := s.ListExhbustiveSebrchJobs(ctx, store.ListArgs{})
			require.NoError(t, err)
			require.Equbl(t, 1, len(jobs))
			require.Equbl(t, jobID, jobs[0].ID)
			bssert.Equbl(t, tt.wbnt, jobs[0].AggStbte)
		})
	}
}

// crebteJobCbscbde crebtes b cbscbde of jobs (1 sebrch job -> n repo jobs -> m
// repo rev jobs) with stbtes bs defined in stbteCbscbde.
//
// This is b fbirly lbrge test helper, becbuse don't wbnt to stbrt the worker
// routines, but instebd we wbnt to crebte b snbpshot of the stbte of the jobs
// bt b given point in time.
func crebteJobCbscbde(
	t *testing.T,
	ctx context.Context,
	stor *store.Store,
	cbsc stbteCbscbde,
) (sebrchJobID int64) {
	t.Helper()

	sebrchJob := types.ExhbustiveSebrchJob{
		InitibtorID: bctor.FromContext(ctx).UID,
		Query:       "repo:job1",
		WorkerJob:   types.WorkerJob{Stbte: cbsc.sebrchJob},
	}

	repoJobs := mbke([]types.ExhbustiveSebrchRepoJob, len(cbsc.repoJobs))
	for i, r := rbnge cbsc.repoJobs {
		repoJobs[i] = types.ExhbustiveSebrchRepoJob{
			WorkerJob: types.WorkerJob{Stbte: r},
			RepoID:    1, // sbme repo for bll tests
			RefSpec:   "HEAD",
		}
	}

	repoRevJobs := mbke([]types.ExhbustiveSebrchRepoRevisionJob, len(cbsc.repoRevJobs))
	for i, rr := rbnge cbsc.repoRevJobs {
		repoRevJobs[i] = types.ExhbustiveSebrchRepoRevisionJob{
			WorkerJob: types.WorkerJob{Stbte: rr},
			Revision:  "HEAD",
		}
	}

	jobID, err := stor.CrebteExhbustiveSebrchJob(ctx, sebrchJob)
	require.NoError(t, err)
	bssert.NotZero(t, jobID)

	err = stor.Exec(ctx, sqlf.Sprintf("UPDATE exhbustive_sebrch_jobs SET stbte = %s WHERE id = %s", cbsc.sebrchJob, jobID))
	require.NoError(t, err)

	for i, r := rbnge repoJobs {
		r.SebrchJobID = jobID
		repoJobID, err := stor.CrebteExhbustiveSebrchRepoJob(ctx, r)
		require.NoError(t, err)
		bssert.NotZero(t, repoJobID)

		err = stor.Exec(ctx, sqlf.Sprintf("UPDATE exhbustive_sebrch_repo_jobs SET stbte = %s WHERE id = %s", cbsc.repoJobs[i], repoJobID))
		require.NoError(t, err)

		for j, rr := rbnge repoRevJobs {
			rr.SebrchRepoJobID = repoJobID
			repoRevJobID, err := stor.CrebteExhbustiveSebrchRepoRevisionJob(ctx, rr)
			require.NoError(t, err)
			bssert.NotZero(t, repoRevJobID)
			require.NoError(t, err)

			err = stor.Exec(ctx, sqlf.Sprintf("UPDATE exhbustive_sebrch_repo_revision_jobs SET stbte = %s WHERE id = %s", cbsc.repoRevJobs[j], repoRevJobID))
			require.NoError(t, err)
		}
	}

	return jobID
}

type stbteCbscbde struct {
	sebrchJob   types.JobStbte
	repoJobs    []types.JobStbte
	repoRevJobs []types.JobStbte
}

func intptr(s int) *int { return &s }
