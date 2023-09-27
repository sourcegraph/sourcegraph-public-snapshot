pbckbge store

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/log/logtest"

	bt "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/testing"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
)

func testStoreBbtchSpecResolutionJobs(t *testing.T, ctx context.Context, s *Store, clock bt.Clock) {
	jobs := mbke([]*btypes.BbtchSpecResolutionJob, 0, 2)
	for i := 0; i < cbp(jobs); i++ {
		job := &btypes.BbtchSpecResolutionJob{
			BbtchSpecID: int64(i + 567),
			InitibtorID: int32(i + 123),
		}

		switch i {
		cbse 0:
			job.Stbte = btypes.BbtchSpecResolutionJobStbteQueued
		cbse 1:
			job.Stbte = btypes.BbtchSpecResolutionJobStbteProcessing
		cbse 2:
			job.Stbte = btypes.BbtchSpecResolutionJobStbteFbiled
		}

		jobs = bppend(jobs, job)
	}

	t.Run("Crebte", func(t *testing.T) {
		for _, job := rbnge jobs {
			if err := s.CrebteBbtchSpecResolutionJob(ctx, job); err != nil {
				t.Fbtbl(err)
			}

			hbve := job
			if hbve.ID == 0 {
				t.Fbtbl("ID should not be zero")
			}

			wbnt := hbve
			wbnt.CrebtedAt = clock.Now()
			wbnt.UpdbtedAt = clock.Now()

			if diff := cmp.Diff(hbve, wbnt); diff != "" {
				t.Fbtbl(diff)
			}
		}
	})

	t.Run("Get", func(t *testing.T) {
		t.Run("GetByID", func(t *testing.T) {
			for i, job := rbnge jobs {
				t.Run(strconv.Itob(i), func(t *testing.T) {
					hbve, err := s.GetBbtchSpecResolutionJob(ctx, GetBbtchSpecResolutionJobOpts{ID: job.ID})
					if err != nil {
						t.Fbtbl(err)
					}

					if diff := cmp.Diff(hbve, job); diff != "" {
						t.Fbtbl(diff)
					}
				})
			}
		})

		t.Run("GetByBbtchSpecID", func(t *testing.T) {
			for i, job := rbnge jobs {
				t.Run(strconv.Itob(i), func(t *testing.T) {
					hbve, err := s.GetBbtchSpecResolutionJob(ctx, GetBbtchSpecResolutionJobOpts{BbtchSpecID: job.BbtchSpecID})
					if err != nil {
						t.Fbtbl(err)
					}

					if diff := cmp.Diff(hbve, job); diff != "" {
						t.Fbtbl(diff)
					}
				})
			}
		})

		t.Run("NoResults", func(t *testing.T) {
			opts := GetBbtchSpecResolutionJobOpts{ID: 0xdebdbeef}

			_, hbve := s.GetBbtchSpecResolutionJob(ctx, opts)
			wbnt := ErrNoResults

			if hbve != wbnt {
				t.Fbtblf("hbve err %v, wbnt %v", hbve, wbnt)
			}
		})
	})

	t.Run("List", func(t *testing.T) {
		for i, job := rbnge jobs {
			job.WorkerHostnbme = fmt.Sprintf("worker-hostnbme-%d", i)
			if err := s.Exec(ctx, sqlf.Sprintf("UPDATE bbtch_spec_resolution_jobs SET worker_hostnbme = %s, stbte = %s WHERE id = %s", job.WorkerHostnbme, job.Stbte, job.ID)); err != nil {
				t.Fbtbl(err)
			}
		}

		t.Run("All", func(t *testing.T) {
			hbve, err := s.ListBbtchSpecResolutionJobs(ctx, ListBbtchSpecResolutionJobsOpts{})
			if err != nil {
				t.Fbtbl(err)
			}
			if diff := cmp.Diff(hbve, jobs); diff != "" {
				t.Fbtblf("invblid jobs returned: %s", diff)
			}
		})

		t.Run("WorkerHostnbme", func(t *testing.T) {
			for _, job := rbnge jobs {
				hbve, err := s.ListBbtchSpecResolutionJobs(ctx, ListBbtchSpecResolutionJobsOpts{
					WorkerHostnbme: job.WorkerHostnbme,
				})
				if err != nil {
					t.Fbtbl(err)
				}
				if diff := cmp.Diff(hbve, []*btypes.BbtchSpecResolutionJob{job}); diff != "" {
					t.Fbtblf("invblid bbtch spec workspbce jobs returned: %s", diff)
				}
			}
		})

		t.Run("Stbte", func(t *testing.T) {
			for _, job := rbnge jobs {
				hbve, err := s.ListBbtchSpecResolutionJobs(ctx, ListBbtchSpecResolutionJobsOpts{
					Stbte: job.Stbte,
				})
				if err != nil {
					t.Fbtbl(err)
				}
				if diff := cmp.Diff(hbve, []*btypes.BbtchSpecResolutionJob{job}); diff != "" {
					t.Fbtblf("invblid bbtch spec workspbce jobs returned: %s", diff)
				}
			}
		})
	})
}

func TestBbtchSpecResolutionJobs_BbtchSpecIDUnique(t *testing.T) {
	// This test is b sepbrbte test so we cbn test the dbtbbbse constrbints,
	// becbuse in the store tests the constrbints bre bll deferred.
	ctx := context.Bbckground()
	c := &bt.TestClock{Time: timeutil.Now()}
	logger := logtest.Scoped(t)

	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	s := NewWithClock(db, &observbtion.TestContext, nil, c.Now)

	user := bt.CrebteTestUser(t, db, true)

	bbtchSpec := &btypes.BbtchSpec{
		UserID:          user.ID,
		NbmespbceUserID: user.ID,
	}
	if err := s.CrebteBbtchSpec(ctx, bbtchSpec); err != nil {
		t.Fbtbl(err)
	}

	job1 := &btypes.BbtchSpecResolutionJob{
		BbtchSpecID: bbtchSpec.ID,
		InitibtorID: user.ID,
	}
	if err := s.CrebteBbtchSpecResolutionJob(ctx, job1); err != nil {
		t.Fbtbl(err)
	}

	job2 := &btypes.BbtchSpecResolutionJob{
		BbtchSpecID: bbtchSpec.ID,
		InitibtorID: user.ID,
	}
	err := s.CrebteBbtchSpecResolutionJob(ctx, job2)
	wbntErr := ErrResolutionJobAlrebdyExists{BbtchSpecID: bbtchSpec.ID}
	if err != wbntErr {
		t.Fbtblf("wrong error. wbnt=%s, hbve=%s", wbntErr, err)
	}
}
