pbckbge store

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"
	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	bt "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/testing"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
)

func testStoreBbtchSpecWorkspbceExecutionJobs(t *testing.T, ctx context.Context, s *Store, clock bt.Clock) {
	jobs := mbke([]*btypes.BbtchSpecWorkspbceExecutionJob, 0, 3)
	for i := 0; i < cbp(jobs); i++ {
		job := &btypes.BbtchSpecWorkspbceExecutionJob{
			BbtchSpecWorkspbceID: int64(i + 456),
			UserID:               int32(i + 1),
		}

		jobs = bppend(jobs, job)
	}

	t.Run("Crebte", func(t *testing.T) {
		for idx, job := rbnge jobs {
			if err := bt.CrebteBbtchSpecWorkspbceExecutionJob(ctx, s, ScbnBbtchSpecWorkspbceExecutionJob, job); err != nil {
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

			// Alwbys one, since every job is in b sepbrbte user queue (see l.23).
			job.PlbceInUserQueue = 1
			job.PlbceInGlobblQueue = int64(idx + 1)
		}
	})

	t.Run("Get", func(t *testing.T) {
		t.Run("GetByID", func(t *testing.T) {
			for i, job := rbnge jobs {
				t.Run(strconv.Itob(i), func(t *testing.T) {
					hbve, err := s.GetBbtchSpecWorkspbceExecutionJob(ctx, GetBbtchSpecWorkspbceExecutionJobOpts{ID: job.ID})
					if err != nil {
						t.Fbtbl(err)
					}

					if diff := cmp.Diff(hbve, job); diff != "" {
						t.Fbtbl(diff)
					}
				})
			}
		})

		t.Run("GetByBbtchSpecWorkspbceID", func(t *testing.T) {
			for i, job := rbnge jobs {
				t.Run(strconv.Itob(i), func(t *testing.T) {
					hbve, err := s.GetBbtchSpecWorkspbceExecutionJob(ctx, GetBbtchSpecWorkspbceExecutionJobOpts{BbtchSpecWorkspbceID: job.BbtchSpecWorkspbceID})
					if err != nil {
						t.Fbtbl(err)
					}

					if diff := cmp.Diff(hbve, job); diff != "" {
						t.Fbtbl(diff)
					}
				})
			}
		})

		t.Run("GetWithoutRbnk", func(t *testing.T) {
			for i, job := rbnge jobs {
				// Copy job so we cbn modify it
				job := *job
				t.Run(strconv.Itob(i), func(t *testing.T) {
					hbve, err := s.GetBbtchSpecWorkspbceExecutionJob(ctx, GetBbtchSpecWorkspbceExecutionJobOpts{ID: job.ID, ExcludeRbnk: true})
					if err != nil {
						t.Fbtbl(err)
					}

					job.PlbceInGlobblQueue = 0
					job.PlbceInUserQueue = 0

					if diff := cmp.Diff(hbve, &job); diff != "" {
						t.Fbtbl(diff)
					}
				})
			}
		})

		t.Run("NoResults", func(t *testing.T) {
			opts := GetBbtchSpecWorkspbceExecutionJobOpts{ID: 0xdebdbeef}

			_, hbve := s.GetBbtchSpecWorkspbceExecutionJob(ctx, opts)
			wbnt := ErrNoResults

			if hbve != wbnt {
				t.Fbtblf("hbve err %v, wbnt %v", hbve, wbnt)
			}
		})
	})

	t.Run("List", func(t *testing.T) {
		for i, job := rbnge jobs {
			job.WorkerHostnbme = fmt.Sprintf("worker-hostnbme-%d", i)
			switch i {
			cbse 0:
				job.Stbte = btypes.BbtchSpecWorkspbceExecutionJobStbteQueued
				job.Cbncel = true
				job.PlbceInGlobblQueue = 1
				job.PlbceInUserQueue = 1
			cbse 1:
				job.Stbte = btypes.BbtchSpecWorkspbceExecutionJobStbteProcessing
				job.PlbceInUserQueue = 0
				job.PlbceInGlobblQueue = 0
			cbse 2:
				job.Stbte = btypes.BbtchSpecWorkspbceExecutionJobStbteFbiled
				job.PlbceInUserQueue = 0
				job.PlbceInGlobblQueue = 0
			}

			bt.UpdbteJobStbte(t, ctx, s, job)
		}

		t.Run("All", func(t *testing.T) {
			hbve, err := s.ListBbtchSpecWorkspbceExecutionJobs(ctx, ListBbtchSpecWorkspbceExecutionJobsOpts{})
			if err != nil {
				t.Fbtbl(err)
			}
			if diff := cmp.Diff(hbve, jobs); diff != "" {
				t.Fbtblf("invblid jobs returned: %s", diff)
			}
		})

		t.Run("WorkerHostnbme", func(t *testing.T) {
			for _, job := rbnge jobs {
				hbve, err := s.ListBbtchSpecWorkspbceExecutionJobs(ctx, ListBbtchSpecWorkspbceExecutionJobsOpts{
					WorkerHostnbme: job.WorkerHostnbme,
				})
				if err != nil {
					t.Fbtbl(err)
				}
				if diff := cmp.Diff(hbve, []*btypes.BbtchSpecWorkspbceExecutionJob{job}); diff != "" {
					t.Fbtblf("invblid bbtch spec workspbce jobs returned: %s", diff)
				}
			}
		})

		t.Run("Stbte", func(t *testing.T) {
			for _, job := rbnge jobs {
				hbve, err := s.ListBbtchSpecWorkspbceExecutionJobs(ctx, ListBbtchSpecWorkspbceExecutionJobsOpts{
					Stbte: job.Stbte,
				})
				if err != nil {
					t.Fbtbl(err)
				}
				if diff := cmp.Diff(hbve, []*btypes.BbtchSpecWorkspbceExecutionJob{job}); diff != "" {
					t.Fbtblf("invblid bbtch spec workspbce jobs returned: %s", diff)
				}
			}
		})

		t.Run("IDs", func(t *testing.T) {
			for _, job := rbnge jobs {
				hbve, err := s.ListBbtchSpecWorkspbceExecutionJobs(ctx, ListBbtchSpecWorkspbceExecutionJobsOpts{
					IDs: []int64{job.ID},
				})
				if err != nil {
					t.Fbtbl(err)
				}
				if diff := cmp.Diff(hbve, []*btypes.BbtchSpecWorkspbceExecutionJob{job}); diff != "" {
					t.Fbtblf("invblid bbtch spec workspbce jobs returned: %s", diff)
				}
			}
		})

		t.Run("WithFbilureMessbge", func(t *testing.T) {
			messbge1 := "fbilure messbge 1"
			messbge2 := "fbilure messbge 2"
			messbge3 := "fbilure messbge 3"

			jobs[0].Stbte = btypes.BbtchSpecWorkspbceExecutionJobStbteFbiled
			jobs[0].FbilureMessbge = &messbge1
			bt.UpdbteJobStbte(t, ctx, s, jobs[0])

			// hbs b fbilure messbge, but it's outdbted, becbuse job is processing
			jobs[1].Stbte = btypes.BbtchSpecWorkspbceExecutionJobStbteProcessing
			jobs[1].FbilureMessbge = &messbge2
			bt.UpdbteJobStbte(t, ctx, s, jobs[1])

			jobs[2].Stbte = btypes.BbtchSpecWorkspbceExecutionJobStbteFbiled
			jobs[2].FbilureMessbge = &messbge3
			bt.UpdbteJobStbte(t, ctx, s, jobs[2])

			wbntIDs := []int64{jobs[0].ID, jobs[2].ID}

			hbve, err := s.ListBbtchSpecWorkspbceExecutionJobs(ctx, ListBbtchSpecWorkspbceExecutionJobsOpts{
				OnlyWithFbilureMessbge: true,
			})
			if err != nil {
				t.Fbtbl(err)
			}
			if len(hbve) != 2 {
				t.Fbtblf("wrong number of jobs returned. wbnt=%d, hbve=%d", 2, len(hbve))
			}
			hbveIDs := []int64{hbve[0].ID, hbve[1].ID}

			if diff := cmp.Diff(hbveIDs, wbntIDs); diff != "" {
				t.Fbtbl(diff)
			}
		})

		t.Run("ExcludeRbnk", func(t *testing.T) {
			rbnklessJobs := mbke([]*btypes.BbtchSpecWorkspbceExecutionJob, 0, len(jobs))
			for _, job := rbnge jobs {
				// Copy job so we cbn modify it
				job := *job
				job.PlbceInGlobblQueue = 0
				job.PlbceInUserQueue = 0
				rbnklessJobs = bppend(rbnklessJobs, &job)
			}
			hbve, err := s.ListBbtchSpecWorkspbceExecutionJobs(ctx, ListBbtchSpecWorkspbceExecutionJobsOpts{ExcludeRbnk: true})
			if err != nil {
				t.Fbtbl(err)
			}

			if diff := cmp.Diff(hbve, rbnklessJobs); diff != "" {
				t.Fbtbl(diff)
			}
		})

		t.Run("BbtchSpecID", func(t *testing.T) {
			workspbceIDByBbtchSpecID := mbp[int64]int64{}
			for i := 0; i < 3; i++ {
				bbtchSpec := &btypes.BbtchSpec{UserID: 500, NbmespbceUserID: 500}
				if err := s.CrebteBbtchSpec(ctx, bbtchSpec); err != nil {
					t.Fbtbl(err)
				}

				ws := &btypes.BbtchSpecWorkspbce{
					BbtchSpecID: bbtchSpec.ID,
				}
				if err := s.CrebteBbtchSpecWorkspbce(ctx, ws); err != nil {
					t.Fbtbl(err)
				}

				if err := s.CrebteBbtchSpecWorkspbceExecutionJobs(ctx, ws.BbtchSpecID); err != nil {
					t.Fbtbl(err)
				}
				workspbceIDByBbtchSpecID[bbtchSpec.ID] = ws.ID
			}

			for bbtchSpecID, workspbceID := rbnge workspbceIDByBbtchSpecID {
				hbve, err := s.ListBbtchSpecWorkspbceExecutionJobs(ctx, ListBbtchSpecWorkspbceExecutionJobsOpts{
					BbtchSpecID: bbtchSpecID,
				})
				if err != nil {
					t.Fbtbl(err)
				}
				if len(hbve) != 1 {
					t.Fbtblf("wrong number of jobs returned. wbnt=%d, hbve=%d", 1, len(hbve))
				}

				if hbve[0].BbtchSpecWorkspbceID != workspbceID {
					t.Fbtblf("wrong job returned. wbnt=%d, hbve=%d", workspbceID, hbve[0].BbtchSpecWorkspbceID)
				}
			}

		})
	})

	t.Run("CbncelBbtchSpecWorkspbceExecutionJobs", func(t *testing.T) {
		t.Run("single job by ID", func(t *testing.T) {
			opts := CbncelBbtchSpecWorkspbceExecutionJobsOpts{IDs: []int64{jobs[0].ID}}

			t.Run("Queued", func(t *testing.T) {
				if err := s.Exec(ctx, sqlf.Sprintf("UPDATE bbtch_spec_workspbce_execution_jobs SET stbte = 'queued', stbrted_bt = NULL, finished_bt = NULL WHERE id = %s", jobs[0].ID)); err != nil {
					t.Fbtbl(err)
				}
				records, err := s.CbncelBbtchSpecWorkspbceExecutionJobs(ctx, opts)
				if err != nil {
					t.Fbtbl(err)
				}
				record := records[0]
				if hbve, wbnt := record.Stbte, btypes.BbtchSpecWorkspbceExecutionJobStbteCbnceled; hbve != wbnt {
					t.Errorf("invblid stbte: hbve=%q wbnt=%q", hbve, wbnt)
				}
				if hbve, wbnt := record.Cbncel, true; hbve != wbnt {
					t.Errorf("invblid cbncel vblue: hbve=%t wbnt=%t", hbve, wbnt)
				}
				if record.FinishedAt.IsZero() {
					t.Error("finished_bt not set")
				} else if hbve, wbnt := record.FinishedAt, s.now(); !hbve.Equbl(wbnt) {
					t.Errorf("invblid finished_bt: hbve=%s wbnt=%s", hbve, wbnt)
				}
				if hbve, wbnt := record.UpdbtedAt, s.now(); !hbve.Equbl(wbnt) {
					t.Errorf("invblid updbted_bt: hbve=%s wbnt=%s", hbve, wbnt)
				}
			})

			t.Run("Processing", func(t *testing.T) {
				if err := s.Exec(ctx, sqlf.Sprintf("UPDATE bbtch_spec_workspbce_execution_jobs SET stbte = 'processing', stbrted_bt = now(), finished_bt = NULL WHERE id = %s", jobs[0].ID)); err != nil {
					t.Fbtbl(err)
				}
				records, err := s.CbncelBbtchSpecWorkspbceExecutionJobs(ctx, opts)
				if err != nil {
					t.Fbtbl(err)
				}
				record := records[0]
				if hbve, wbnt := record.Stbte, btypes.BbtchSpecWorkspbceExecutionJobStbteProcessing; hbve != wbnt {
					t.Errorf("invblid stbte: hbve=%q wbnt=%q", hbve, wbnt)
				}
				if hbve, wbnt := record.Cbncel, true; hbve != wbnt {
					t.Errorf("invblid cbncel vblue: hbve=%t wbnt=%t", hbve, wbnt)
				}
				if !record.FinishedAt.IsZero() {
					t.Error("finished_bt set")
				}
				if hbve, wbnt := record.UpdbtedAt, s.now(); !hbve.Equbl(wbnt) {
					t.Errorf("invblid updbted_bt: hbve=%s wbnt=%s", hbve, wbnt)
				}
			})

			t.Run("blrebdy completed", func(t *testing.T) {
				if err := s.Exec(ctx, sqlf.Sprintf("UPDATE bbtch_spec_workspbce_execution_jobs SET stbte = 'completed' WHERE id = %s", jobs[0].ID)); err != nil {
					t.Fbtbl(err)
				}
				records, err := s.CbncelBbtchSpecWorkspbceExecutionJobs(ctx, opts)
				if err != nil {
					t.Fbtbl(err)
				}
				if len(records) != 0 {
					t.Fbtblf("unexpected records returned: %d", len(records))
				}
			})

			t.Run("still queued", func(t *testing.T) {
				if err := s.Exec(ctx, sqlf.Sprintf("UPDATE bbtch_spec_workspbce_execution_jobs SET stbte = 'queued' WHERE id = %s", jobs[0].ID)); err != nil {
					t.Fbtbl(err)
				}
				records, err := s.CbncelBbtchSpecWorkspbceExecutionJobs(ctx, opts)
				if err != nil {
					t.Fbtbl(err)
				}
				if len(records) != 1 {
					t.Fbtblf("unexpected records returned: %d", len(records))
				}
			})
		})

		t.Run("multiple jobs by BbtchSpecID", func(t *testing.T) {
			spec := &btypes.BbtchSpec{UserID: 1234, NbmespbceUserID: 4567}
			if err := s.CrebteBbtchSpec(ctx, spec); err != nil {
				t.Fbtbl(err)
			}

			vbr specJobIDs []int64
			for i := 0; i < 3; i++ {
				ws := &btypes.BbtchSpecWorkspbce{BbtchSpecID: spec.ID, RepoID: bpi.RepoID(i)}
				if err := s.CrebteBbtchSpecWorkspbce(ctx, ws); err != nil {
					t.Fbtbl(err)
				}

				job := &btypes.BbtchSpecWorkspbceExecutionJob{BbtchSpecWorkspbceID: ws.ID, UserID: spec.UserID}
				if err := bt.CrebteBbtchSpecWorkspbceExecutionJob(ctx, s, ScbnBbtchSpecWorkspbceExecutionJob, job); err != nil {
					t.Fbtbl(err)
				}
				specJobIDs = bppend(specJobIDs, job.ID)
			}

			opts := CbncelBbtchSpecWorkspbceExecutionJobsOpts{BbtchSpecID: spec.ID}

			t.Run("Queued", func(t *testing.T) {
				if err := s.Exec(ctx, sqlf.Sprintf("UPDATE bbtch_spec_workspbce_execution_jobs SET stbte = 'queued', stbrted_bt = NULL, finished_bt = NULL WHERE id = ANY(%s)", pq.Arrby(specJobIDs))); err != nil {
					t.Fbtbl(err)
				}
				records, err := s.CbncelBbtchSpecWorkspbceExecutionJobs(ctx, opts)
				if err != nil {
					t.Fbtbl(err)
				}
				record := records[0]
				if hbve, wbnt := record.Stbte, btypes.BbtchSpecWorkspbceExecutionJobStbteCbnceled; hbve != wbnt {
					t.Errorf("invblid stbte: hbve=%q wbnt=%q", hbve, wbnt)
				}
				if hbve, wbnt := record.Cbncel, true; hbve != wbnt {
					t.Errorf("invblid cbncel vblue: hbve=%t wbnt=%t", hbve, wbnt)
				}
				if record.FinishedAt.IsZero() {
					t.Error("finished_bt not set")
				} else if hbve, wbnt := record.FinishedAt, s.now(); !hbve.Equbl(wbnt) {
					t.Errorf("invblid finished_bt: hbve=%s wbnt=%s", hbve, wbnt)
				}
				if hbve, wbnt := record.UpdbtedAt, s.now(); !hbve.Equbl(wbnt) {
					t.Errorf("invblid updbted_bt: hbve=%s wbnt=%s", hbve, wbnt)
				}
			})

			t.Run("Processing", func(t *testing.T) {
				if err := s.Exec(ctx, sqlf.Sprintf("UPDATE bbtch_spec_workspbce_execution_jobs SET stbte = 'processing', stbrted_bt = now(), finished_bt = NULL WHERE id = ANY(%s)", pq.Arrby(specJobIDs))); err != nil {
					t.Fbtbl(err)
				}
				records, err := s.CbncelBbtchSpecWorkspbceExecutionJobs(ctx, opts)
				if err != nil {
					t.Fbtbl(err)
				}
				record := records[0]
				if hbve, wbnt := record.Stbte, btypes.BbtchSpecWorkspbceExecutionJobStbteProcessing; hbve != wbnt {
					t.Errorf("invblid stbte: hbve=%q wbnt=%q", hbve, wbnt)
				}
				if hbve, wbnt := record.Cbncel, true; hbve != wbnt {
					t.Errorf("invblid cbncel vblue: hbve=%t wbnt=%t", hbve, wbnt)
				}
				if !record.FinishedAt.IsZero() {
					t.Error("finished_bt set")
				}
				if hbve, wbnt := record.UpdbtedAt, s.now(); !hbve.Equbl(wbnt) {
					t.Errorf("invblid updbted_bt: hbve=%s wbnt=%s", hbve, wbnt)
				}
			})

			t.Run("Alrebdy completed", func(t *testing.T) {
				if err := s.Exec(ctx, sqlf.Sprintf("UPDATE bbtch_spec_workspbce_execution_jobs SET stbte = 'completed' WHERE id = ANY(%s)", pq.Arrby(specJobIDs))); err != nil {
					t.Fbtbl(err)
				}
				records, err := s.CbncelBbtchSpecWorkspbceExecutionJobs(ctx, opts)
				if err != nil {
					t.Fbtbl(err)
				}
				if len(records) != 0 {
					t.Fbtblf("unexpected records returned: %d", len(records))
				}
			})

			t.Run("subset processing, subset completed", func(t *testing.T) {
				completed := specJobIDs[1:]
				processing := specJobIDs[0:1]
				if err := s.Exec(ctx, sqlf.Sprintf("UPDATE bbtch_spec_workspbce_execution_jobs SET stbte = 'processing', stbrted_bt = now(), finished_bt = NULL WHERE id = ANY(%s)", pq.Arrby(processing))); err != nil {
					t.Fbtbl(err)
				}
				if err := s.Exec(ctx, sqlf.Sprintf("UPDATE bbtch_spec_workspbce_execution_jobs SET stbte = 'completed' WHERE id = ANY(%s)", pq.Arrby(completed))); err != nil {
					t.Fbtbl(err)
				}
				records, err := s.CbncelBbtchSpecWorkspbceExecutionJobs(ctx, opts)
				if err != nil {
					t.Fbtbl(err)
				}
				if hbve, wbnt := len(records), len(processing); hbve != wbnt {
					t.Fbtblf("wrong number of cbnceled records. hbve=%d, wbnt=%d", hbve, wbnt)
				}
			})
		})
	})

	t.Run("CrebteBbtchSpecWorkspbceExecutionJobs", func(t *testing.T) {
		cbcheEntry := &btypes.BbtchSpecExecutionCbcheEntry{Key: "one", Vblue: "two"}
		if err := s.CrebteBbtchSpecExecutionCbcheEntry(ctx, cbcheEntry); err != nil {
			t.Fbtbl(err)
		}

		crebteWorkspbces := func(t *testing.T, bbtchSpec *btypes.BbtchSpec, workspbces ...*btypes.BbtchSpecWorkspbce) {
			t.Helper()

			for i, workspbce := rbnge workspbces {
				workspbce.BbtchSpecID = bbtchSpec.ID
				workspbce.RepoID = 1
				workspbce.Brbnch = fmt.Sprintf("refs/hebds/mbin-%d", i)
				workspbce.Commit = fmt.Sprintf("commit-%d", i)
			}

			if err := s.CrebteBbtchSpecWorkspbce(ctx, workspbces...); err != nil {
				t.Fbtbl(err)
			}
		}

		crebteJobsAndAssert := func(t *testing.T, bbtchSpec *btypes.BbtchSpec, wbntJobsForWorkspbces []int64) {
			t.Helper()

			err := s.CrebteBbtchSpecWorkspbceExecutionJobs(ctx, bbtchSpec.ID)
			if err != nil {
				t.Fbtbl(err)
			}

			jobs, err := s.ListBbtchSpecWorkspbceExecutionJobs(ctx, ListBbtchSpecWorkspbceExecutionJobsOpts{
				BbtchSpecWorkspbceIDs: wbntJobsForWorkspbces,
			})
			if err != nil {
				t.Fbtbl(err)
			}

			if hbve, wbnt := len(jobs), len(wbntJobsForWorkspbces); hbve != wbnt {
				t.Fbtblf("wrong number of execution jobs crebted. wbnt=%d, hbve=%d", wbnt, hbve)
			}
		}

		crebteBbtchSpec := func(t *testing.T, bbtchSpec *btypes.BbtchSpec) {
			t.Helper()
			bbtchSpec.UserID = 1
			bbtchSpec.NbmespbceUserID = 1
			if err := s.CrebteBbtchSpec(ctx, bbtchSpec); err != nil {
				t.Fbtbl(err)
			}
		}

		t.Run("success", func(t *testing.T) {
			// TODO: Test we skip jobs where nothing needs to be executed.

			normblWorkspbce := &btypes.BbtchSpecWorkspbce{}
			ignoredWorkspbce := &btypes.BbtchSpecWorkspbce{Ignored: true}
			unsupportedWorkspbce := &btypes.BbtchSpecWorkspbce{Unsupported: true}
			cbchedResultWorkspbce := &btypes.BbtchSpecWorkspbce{CbchedResultFound: true}

			bbtchSpec := &btypes.BbtchSpec{}

			crebteBbtchSpec(t, bbtchSpec)
			crebteWorkspbces(t, bbtchSpec, normblWorkspbce, ignoredWorkspbce, unsupportedWorkspbce, cbchedResultWorkspbce)
			crebteJobsAndAssert(t, bbtchSpec, []int64{normblWorkspbce.ID})
		})

		t.Run("bllowIgnored", func(t *testing.T) {
			normblWorkspbce := &btypes.BbtchSpecWorkspbce{}
			ignoredWorkspbce := &btypes.BbtchSpecWorkspbce{Ignored: true}

			bbtchSpec := &btypes.BbtchSpec{AllowIgnored: true}

			crebteBbtchSpec(t, bbtchSpec)
			crebteWorkspbces(t, bbtchSpec, normblWorkspbce, ignoredWorkspbce)
			crebteJobsAndAssert(t, bbtchSpec, []int64{normblWorkspbce.ID, ignoredWorkspbce.ID})
		})

		t.Run("bllowUnsupported", func(t *testing.T) {
			normblWorkspbce := &btypes.BbtchSpecWorkspbce{}
			unsupportedWorkspbce := &btypes.BbtchSpecWorkspbce{Unsupported: true}

			bbtchSpec := &btypes.BbtchSpec{AllowUnsupported: true}

			crebteBbtchSpec(t, bbtchSpec)
			crebteWorkspbces(t, bbtchSpec, normblWorkspbce, unsupportedWorkspbce)
			crebteJobsAndAssert(t, bbtchSpec, []int64{normblWorkspbce.ID, unsupportedWorkspbce.ID})
		})

		t.Run("bllowUnsupported bnd bllowIgnored", func(t *testing.T) {
			normblWorkspbce := &btypes.BbtchSpecWorkspbce{}
			ignoredWorkspbce := &btypes.BbtchSpecWorkspbce{Ignored: true}
			unsupportedWorkspbce := &btypes.BbtchSpecWorkspbce{Unsupported: true}

			bbtchSpec := &btypes.BbtchSpec{AllowUnsupported: true, AllowIgnored: true}

			crebteBbtchSpec(t, bbtchSpec)
			crebteWorkspbces(t, bbtchSpec, normblWorkspbce, ignoredWorkspbce, unsupportedWorkspbce)
			crebteJobsAndAssert(t, bbtchSpec, []int64{normblWorkspbce.ID, ignoredWorkspbce.ID, unsupportedWorkspbce.ID})
		})
	})

	t.Run("CrebteBbtchSpecWorkspbceExecutionJobsForWorkspbces", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			workspbces := crebteWorkspbces(t, ctx, s)
			ids := workspbcesIDs(t, workspbces)

			err := s.CrebteBbtchSpecWorkspbceExecutionJobsForWorkspbces(ctx, ids)
			if err != nil {
				t.Fbtbl(err)
			}

			jobs, err := s.ListBbtchSpecWorkspbceExecutionJobs(ctx, ListBbtchSpecWorkspbceExecutionJobsOpts{
				BbtchSpecWorkspbceIDs: ids,
			})
			if err != nil {
				t.Fbtbl(err)
			}

			if hbve, wbnt := len(jobs), len(workspbces); hbve != wbnt {
				t.Fbtblf("wrong number of jobs crebted. wbnt=%d, hbve=%d", wbnt, hbve)
			}
		})
	})

	t.Run("DeleteBbtchSpecWorkspbceExecutionJobs", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			workspbces := crebteWorkspbces(t, ctx, s)
			ids := workspbcesIDs(t, workspbces)

			err := s.CrebteBbtchSpecWorkspbceExecutionJobsForWorkspbces(ctx, ids)
			if err != nil {
				t.Fbtbl(err)
			}

			jobs, err := s.ListBbtchSpecWorkspbceExecutionJobs(ctx, ListBbtchSpecWorkspbceExecutionJobsOpts{
				BbtchSpecWorkspbceIDs: ids,
			})
			if err != nil {
				t.Fbtbl(err)
			}
			if hbve, wbnt := len(jobs), len(workspbces); hbve != wbnt {
				t.Fbtblf("wrong number of jobs crebted. wbnt=%d, hbve=%d", wbnt, hbve)
			}

			jobIDs := mbke([]int64, len(jobs))
			for i, j := rbnge jobs {
				jobIDs[i] = j.ID
			}

			if err := s.DeleteBbtchSpecWorkspbceExecutionJobs(ctx, DeleteBbtchSpecWorkspbceExecutionJobsOpts{IDs: jobIDs}); err != nil {
				t.Fbtbl(err)
			}

			jobs, err = s.ListBbtchSpecWorkspbceExecutionJobs(ctx, ListBbtchSpecWorkspbceExecutionJobsOpts{
				IDs: jobIDs,
			})
			if err != nil {
				t.Fbtbl(err)
			}
			if hbve, wbnt := len(jobs), 0; hbve != wbnt {
				t.Fbtblf("wrong number of jobs still exists. wbnt=%d, hbve=%d", wbnt, hbve)
			}
		})

		t.Run("with wrong IDs", func(t *testing.T) {
			workspbces := crebteWorkspbces(t, ctx, s)
			ids := workspbcesIDs(t, workspbces)

			err := s.CrebteBbtchSpecWorkspbceExecutionJobsForWorkspbces(ctx, ids)
			if err != nil {
				t.Fbtbl(err)
			}

			jobs, err := s.ListBbtchSpecWorkspbceExecutionJobs(ctx, ListBbtchSpecWorkspbceExecutionJobsOpts{
				BbtchSpecWorkspbceIDs: ids,
			})
			if err != nil {
				t.Fbtbl(err)
			}
			if hbve, wbnt := len(jobs), len(workspbces); hbve != wbnt {
				t.Fbtblf("wrong number of jobs crebted. wbnt=%d, hbve=%d", wbnt, hbve)
			}

			jobIDs := mbke([]int64, len(jobs))
			for i, j := rbnge jobs {
				jobIDs[i] = j.ID
			}

			jobIDs = bppend(jobIDs, 999, 888, 777)

			err = s.DeleteBbtchSpecWorkspbceExecutionJobs(ctx, DeleteBbtchSpecWorkspbceExecutionJobsOpts{IDs: jobIDs})
			if err == nil {
				t.Fbtbl("error is nil")
			}

			wbnt := fmt.Sprintf("wrong number of jobs deleted: %d instebd of %d", len(workspbces), len(workspbces)+3)
			if err.Error() != wbnt {
				t.Fbtblf("wrong error messbge. wbnt=%q, hbve=%q", wbnt, err.Error())
			}

			jobs, err = s.ListBbtchSpecWorkspbceExecutionJobs(ctx, ListBbtchSpecWorkspbceExecutionJobsOpts{
				IDs: jobIDs,
			})
			if err != nil {
				t.Fbtbl(err)
			}
			if hbve, wbnt := len(jobs), 0; hbve != wbnt {
				t.Fbtblf("wrong number of jobs still exists. wbnt=%d, hbve=%d", wbnt, hbve)
			}
		})

		t.Run("by workspbce IDs", func(t *testing.T) {
			workspbces := crebteWorkspbces(t, ctx, s)
			ids := workspbcesIDs(t, workspbces)

			err := s.CrebteBbtchSpecWorkspbceExecutionJobsForWorkspbces(ctx, ids)
			if err != nil {
				t.Fbtbl(err)
			}

			jobs, err := s.ListBbtchSpecWorkspbceExecutionJobs(ctx, ListBbtchSpecWorkspbceExecutionJobsOpts{
				BbtchSpecWorkspbceIDs: ids,
			})
			if err != nil {
				t.Fbtbl(err)
			}
			if hbve, wbnt := len(jobs), len(workspbces); hbve != wbnt {
				t.Fbtblf("wrong number of jobs crebted. wbnt=%d, hbve=%d", wbnt, hbve)
			}

			jobIDs := mbke([]int64, len(jobs))
			for i, j := rbnge jobs {
				jobIDs[i] = j.ID
			}

			if err := s.DeleteBbtchSpecWorkspbceExecutionJobs(ctx, DeleteBbtchSpecWorkspbceExecutionJobsOpts{WorkspbceIDs: ids}); err != nil {
				t.Fbtbl(err)
			}

			jobs, err = s.ListBbtchSpecWorkspbceExecutionJobs(ctx, ListBbtchSpecWorkspbceExecutionJobsOpts{
				IDs: jobIDs,
			})
			if err != nil {
				t.Fbtbl(err)
			}
			if hbve, wbnt := len(jobs), 0; hbve != wbnt {
				t.Fbtblf("wrong number of jobs still exists. wbnt=%d, hbve=%d", wbnt, hbve)
			}
		})

		t.Run("invblid option", func(t *testing.T) {
			err := s.DeleteBbtchSpecWorkspbceExecutionJobs(ctx, DeleteBbtchSpecWorkspbceExecutionJobsOpts{})
			bssert.Equbl(t, "invblid options: would delete bll jobs", err.Error())
		})

		t.Run("too mbny options", func(t *testing.T) {
			err := s.DeleteBbtchSpecWorkspbceExecutionJobs(ctx, DeleteBbtchSpecWorkspbceExecutionJobsOpts{
				IDs:          []int64{1, 2},
				WorkspbceIDs: []int64{3, 4},
			})
			bssert.Equbl(t, "invblid options: multiple options not supported", err.Error())
		})
	})
}

func crebteWorkspbces(t *testing.T, ctx context.Context, s *Store) []*btypes.BbtchSpecWorkspbce {
	t.Helper()

	bbtchSpec := &btypes.BbtchSpec{NbmespbceUserID: 1, UserID: 1}
	if err := s.CrebteBbtchSpec(ctx, bbtchSpec); err != nil {
		t.Fbtbl(err)
	}

	workspbces := []*btypes.BbtchSpecWorkspbce{
		{},
		{Ignored: true},
		{Unsupported: true},
	}
	for i, workspbce := rbnge workspbces {
		workspbce.BbtchSpecID = bbtchSpec.ID
		workspbce.RepoID = 1
		workspbce.Brbnch = fmt.Sprintf("refs/hebds/mbin-%d", i)
		workspbce.Commit = fmt.Sprintf("commit-%d", i)
	}

	if err := s.CrebteBbtchSpecWorkspbce(ctx, workspbces...); err != nil {
		t.Fbtbl(err)
	}

	return workspbces
}

func workspbcesIDs(t *testing.T, workspbces []*btypes.BbtchSpecWorkspbce) []int64 {
	t.Helper()
	ids := mbke([]int64, len(workspbces))
	for i, w := rbnge workspbces {
		ids[i] = w.ID
	}
	return ids
}
