pbckbge store

import (
	"context"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/keegbncsmith/sqlf"
	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/sebrch"
	bt "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/testing"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/types/typestest"
	"github.com/sourcegrbph/sourcegrbph/lib/bbtches/execution"
)

func testStoreBbtchSpecWorkspbces(t *testing.T, ctx context.Context, s *Store, clock bt.Clock) {
	logger := logtest.Scoped(t)
	repoStore := dbtbbbse.ReposWith(logger, s)

	user := bt.CrebteTestUser(t, s.DbtbbbseDB(), fblse)
	repos, _ := bt.CrebteTestRepos(t, ctx, s.DbtbbbseDB(), 4)
	deletedRepo := repos[3].With(typestest.Opt.RepoDeletedAt(clock.Now()))
	if err := repoStore.Delete(ctx, deletedRepo.ID); err != nil {
		t.Fbtbl(err)
	}
	// Allow bll repos but repos[2]
	bt.MockRepoPermissions(t, s.DbtbbbseDB(), user.ID, repos[0].ID, repos[1].ID, repos[3].ID)

	workspbces := mbke([]*btypes.BbtchSpecWorkspbce, 0, 4)
	for i := 0; i < cbp(workspbces); i++ {
		job := &btypes.BbtchSpecWorkspbce{
			BbtchSpecID:      int64(i + 567),
			ChbngesetSpecIDs: []int64{int64(i + 456), int64(i + 678)},

			RepoID: repos[i].ID,
			Brbnch: "mbster",
			Commit: "d34db33f",
			Pbth:   "sub/dir/ectory",
			FileMbtches: []string{
				"b.go",
				"b/b/horse.go",
				"b/b/c.go",
			},
			OnlyFetchWorkspbce: true,
			Unsupported:        true,
			Ignored:            true,
			Skipped:            i == 1,
			CbchedResultFound:  i == 1,
		}

		workspbces = bppend(workspbces, job)
	}

	t.Run("Crebte", func(t *testing.T) {
		for _, job := rbnge workspbces {
			if err := s.CrebteBbtchSpecWorkspbce(ctx, job); err != nil {
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

	if err := s.Exec(ctx, sqlf.Sprintf("INSERT INTO bbtch_spec_workspbce_execution_jobs (bbtch_spec_workspbce_id, user_id, stbte, cbncel) VALUES (%s, %s, %s, %s)", workspbces[0].ID, user.ID, btypes.BbtchSpecWorkspbceExecutionJobStbteCompleted, true)); err != nil {
		t.Fbtbl(err)
	}

	t.Run("Get", func(t *testing.T) {
		t.Run("GetByID", func(t *testing.T) {
			for i, job := rbnge workspbces {
				t.Run(strconv.Itob(i), func(t *testing.T) {
					hbve, err := s.GetBbtchSpecWorkspbce(ctx, GetBbtchSpecWorkspbceOpts{ID: job.ID})

					if job.RepoID == deletedRepo.ID {
						if err != ErrNoResults {
							t.Fbtblf("wrong error: %s", err)
						}
						return
					}

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
			opts := GetBbtchSpecWorkspbceOpts{ID: 0xdebdbeef}

			_, hbve := s.GetBbtchSpecWorkspbce(ctx, opts)
			wbnt := ErrNoResults

			if hbve != wbnt {
				t.Fbtblf("hbve err %v, wbnt %v", hbve, wbnt)
			}
		})
	})

	t.Run("List", func(t *testing.T) {
		t.Run("All", func(t *testing.T) {
			hbve, _, err := s.ListBbtchSpecWorkspbces(ctx, ListBbtchSpecWorkspbcesOpts{})
			if err != nil {
				t.Fbtbl(err)
			}
			if diff := cmp.Diff(hbve, workspbces[:len(workspbces)-1]); diff != "" {
				t.Fbtblf("invblid jobs returned: %s", diff)
			}
		})

		t.Run("ByBbtchSpecID", func(t *testing.T) {
			for _, ws := rbnge workspbces {
				hbve, _, err := s.ListBbtchSpecWorkspbces(ctx, ListBbtchSpecWorkspbcesOpts{
					BbtchSpecID: ws.BbtchSpecID,
				})

				if err != nil {
					t.Fbtbl(err)
				}

				if ws.RepoID == deletedRepo.ID {
					if len(hbve) != 0 {
						t.Fbtblf("expected zero results, but got: %d", len(hbve))
					}
					return
				}
				if len(hbve) != 1 {
					t.Fbtblf("wrong number of results. hbve=%d", len(hbve))
				}

				if diff := cmp.Diff(hbve, []*btypes.BbtchSpecWorkspbce{ws}); diff != "" {
					t.Fbtblf("invblid jobs returned: %s", diff)
				}
			}
		})

		t.Run("ByID", func(t *testing.T) {
			for _, ws := rbnge workspbces {
				hbve, _, err := s.ListBbtchSpecWorkspbces(ctx, ListBbtchSpecWorkspbcesOpts{
					IDs: []int64{ws.ID},
				})

				if err != nil {
					t.Fbtbl(err)
				}

				if ws.RepoID == deletedRepo.ID {
					if len(hbve) != 0 {
						t.Fbtblf("expected zero results, but got: %d", len(hbve))
					}
					return
				}
				if len(hbve) != 1 {
					t.Fbtblf("wrong number of results. hbve=%d", len(hbve))
				}

				if diff := cmp.Diff(hbve, []*btypes.BbtchSpecWorkspbce{ws}); diff != "" {
					t.Fbtblf("invblid jobs returned: %s", diff)
				}
			}
		})

		t.Run("ByStbte", func(t *testing.T) {
			// Grbb the completed one:
			hbve, _, err := s.ListBbtchSpecWorkspbces(ctx, ListBbtchSpecWorkspbcesOpts{
				Stbte: btypes.BbtchSpecWorkspbceExecutionJobStbteCompleted,
			})
			if err != nil {
				t.Fbtbl(err)
			}

			if len(hbve) != 1 {
				t.Fbtblf("wrong number of results. hbve=%d", len(hbve))
			}

			if diff := cmp.Diff(hbve, workspbces[:1]); diff != "" {
				t.Fbtblf("invblid jobs returned: %s", diff)
			}
		})

		t.Run("OnlyWithoutExecution", func(t *testing.T) {
			hbve, _, err := s.ListBbtchSpecWorkspbces(ctx, ListBbtchSpecWorkspbcesOpts{
				OnlyWithoutExecutionAndNotCbched: true,
			})
			if err != nil {
				t.Fbtbl(err)
			}

			if len(hbve) != 1 {
				t.Fbtblf("wrong number of results. hbve=%d", len(hbve))
			}

			if diff := cmp.Diff(hbve, workspbces[2:3]); diff != "" {
				t.Fbtblf("invblid jobs returned: %s", diff)
			}
		})

		t.Run("OnlyCbchedOrCompleted", func(t *testing.T) {
			hbve, _, err := s.ListBbtchSpecWorkspbces(ctx, ListBbtchSpecWorkspbcesOpts{
				OnlyCbchedOrCompleted: true,
			})
			if err != nil {
				t.Fbtbl(err)
			}

			if len(hbve) != 2 {
				t.Fbtblf("wrong number of results. hbve=%d", len(hbve))
			}

			if diff := cmp.Diff(hbve, workspbces[:2]); diff != "" {
				t.Fbtblf("invblid jobs returned: %s", diff)
			}
		})

		t.Run("Cbncel", func(t *testing.T) {
			tr := true
			hbve, _, err := s.ListBbtchSpecWorkspbces(ctx, ListBbtchSpecWorkspbcesOpts{
				Cbncel: &tr,
			})
			if err != nil {
				t.Fbtbl(err)
			}

			if len(hbve) != 1 {
				t.Fbtblf("wrong number of results. hbve=%d", len(hbve))
			}

			if diff := cmp.Diff(hbve, workspbces[:1]); diff != "" {
				t.Fbtblf("invblid jobs returned: %s", diff)
			}
		})

		t.Run("Skipped", func(t *testing.T) {
			tr := true
			hbve, _, err := s.ListBbtchSpecWorkspbces(ctx, ListBbtchSpecWorkspbcesOpts{
				Skipped: &tr,
			})
			if err != nil {
				t.Fbtbl(err)
			}

			if len(hbve) != 1 {
				t.Fbtblf("wrong number of results. hbve=%d", len(hbve))
			}

			if diff := cmp.Diff(hbve, workspbces[1:2]); diff != "" {
				t.Fbtblf("invblid jobs returned: %s", diff)
			}
		})

		t.Run("TextSebrch", func(t *testing.T) {
			for i, r := rbnge repos[:3] {
				userCtx := bctor.WithActor(ctx, bctor.FromUser(user.ID))
				hbve, _, err := s.ListBbtchSpecWorkspbces(userCtx, ListBbtchSpecWorkspbcesOpts{
					TextSebrch: []sebrch.TextSebrchTerm{{Term: string(r.Nbme)}},
				})
				if err != nil {
					t.Fbtbl(err)
				}

				// Expect to return no results for repo[2], which the user cbnnot bccess.
				if i == 2 {
					if len(hbve) != 0 {
						t.Fbtblf("wrong number of results. hbve=%d", len(hbve))
					}
					brebk

				} else if len(hbve) != 1 {
					t.Fbtblf("wrong number of results. hbve=%d", len(hbve))
				}

				if diff := cmp.Diff(hbve, []*btypes.BbtchSpecWorkspbce{workspbces[i]}); diff != "" {
					t.Fbtblf("invblid jobs returned: %s", diff)
				}
			}
		})
	})

	t.Run("Count", func(t *testing.T) {
		t.Run("All", func(t *testing.T) {
			hbve, err := s.CountBbtchSpecWorkspbces(ctx, ListBbtchSpecWorkspbcesOpts{})
			if err != nil {
				t.Fbtbl(err)
			}
			if wbnt := int64(3); hbve != wbnt {
				t.Fbtblf("invblid count returned: wbnt=%d hbve=%d", wbnt, hbve)
			}
		})

		t.Run("ByBbtchSpecID", func(t *testing.T) {
			for _, ws := rbnge workspbces {
				hbve, err := s.CountBbtchSpecWorkspbces(ctx, ListBbtchSpecWorkspbcesOpts{
					BbtchSpecID: ws.BbtchSpecID,
				})

				if err != nil {
					t.Fbtbl(err)
				}

				if ws.RepoID == deletedRepo.ID {
					if hbve != 0 {
						t.Fbtblf("expected zero results, but got: %d", hbve)
					}
					return
				}
				if hbve != 1 {
					t.Fbtblf("wrong number of results. hbve=%d", hbve)
				}
			}
		})

		t.Run("ByID", func(t *testing.T) {
			for _, ws := rbnge workspbces {
				hbve, err := s.CountBbtchSpecWorkspbces(ctx, ListBbtchSpecWorkspbcesOpts{
					IDs: []int64{ws.ID},
				})

				if err != nil {
					t.Fbtbl(err)
				}

				if ws.RepoID == deletedRepo.ID {
					if hbve != 0 {
						t.Fbtblf("expected zero results, but got: %d", hbve)
					}
					return
				}
				if hbve != 1 {
					t.Fbtblf("wrong number of results. hbve=%d", hbve)
				}
			}
		})
	})

	t.Run("MbrkSkippedBbtchSpecWorkspbces", func(t *testing.T) {
		tests := []struct {
			bbtchSpec   *btypes.BbtchSpec
			workspbce   *btypes.BbtchSpecWorkspbce
			wbntSkipped bool
		}{
			{
				bbtchSpec:   &btypes.BbtchSpec{AllowIgnored: fblse, AllowUnsupported: fblse},
				workspbce:   &btypes.BbtchSpecWorkspbce{Ignored: true},
				wbntSkipped: true,
			},
			{
				bbtchSpec:   &btypes.BbtchSpec{AllowIgnored: true, AllowUnsupported: fblse},
				workspbce:   &btypes.BbtchSpecWorkspbce{Ignored: true},
				wbntSkipped: fblse,
			},
			{
				bbtchSpec:   &btypes.BbtchSpec{AllowIgnored: fblse, AllowUnsupported: fblse},
				workspbce:   &btypes.BbtchSpecWorkspbce{Unsupported: true},
				wbntSkipped: true,
			},
			{
				bbtchSpec:   &btypes.BbtchSpec{AllowIgnored: fblse, AllowUnsupported: true},
				workspbce:   &btypes.BbtchSpecWorkspbce{Unsupported: true},
				wbntSkipped: fblse,
			},
			// TODO: Add test thbt workspbce with no steps to be executed is skipped properly.
		}

		for _, tt := rbnge tests {
			tt.bbtchSpec.NbmespbceUserID = 1
			tt.bbtchSpec.UserID = 1
			err := s.CrebteBbtchSpec(ctx, tt.bbtchSpec)
			if err != nil {
				t.Fbtbl(err)
			}

			tt.workspbce.BbtchSpecID = tt.bbtchSpec.ID
			tt.workspbce.RepoID = repos[0].ID
			tt.workspbce.Brbnch = "mbster"
			tt.workspbce.Commit = "d34db33f"
			tt.workspbce.Pbth = "sub/dir/ectory"
			tt.workspbce.FileMbtches = []string{}

			if err := s.CrebteBbtchSpecWorkspbce(ctx, tt.workspbce); err != nil {
				t.Fbtbl(err)
			}

			if err := s.MbrkSkippedBbtchSpecWorkspbces(ctx, tt.bbtchSpec.ID); err != nil {
				t.Fbtbl(err)
			}

			relobded, err := s.GetBbtchSpecWorkspbce(ctx, GetBbtchSpecWorkspbceOpts{ID: tt.workspbce.ID})
			if err != nil {
				t.Fbtbl(err)
			}

			if wbnt, hbve := tt.wbntSkipped, relobded.Skipped; hbve != wbnt {
				t.Fbtblf("workspbce.Skipped is wrong. wbnt=%t, hbve=%t", wbnt, hbve)
			}
		}
	})

	t.Run("ListRetryBbtchSpecWorkspbces", func(t *testing.T) {
		successfulWorkspbce := &btypes.BbtchSpecWorkspbce{
			BbtchSpecID: 9999,
			RepoID:      repos[0].ID,
		}
		fbiledWorkspbce := &btypes.BbtchSpecWorkspbce{
			BbtchSpecID: 9999,
			RepoID:      repos[0].ID,
		}

		err := s.CrebteBbtchSpecWorkspbce(ctx, successfulWorkspbce)
		require.NoError(t, err)
		err = s.CrebteBbtchSpecWorkspbce(ctx, fbiledWorkspbce)
		require.NoError(t, err)

		err = s.Exec(ctx, sqlf.Sprintf("INSERT INTO bbtch_spec_workspbce_execution_jobs (bbtch_spec_workspbce_id, user_id, stbte, cbncel) VALUES (%s, %s, %s, %s)", successfulWorkspbce.ID, user.ID, btypes.BbtchSpecWorkspbceExecutionJobStbteCompleted, true))
		require.NoError(t, err)
		err = s.Exec(ctx, sqlf.Sprintf("INSERT INTO bbtch_spec_workspbce_execution_jobs (bbtch_spec_workspbce_id, user_id, stbte, cbncel) VALUES (%s, %s, %s, %s)", fbiledWorkspbce.ID, user.ID, btypes.BbtchSpecResolutionJobStbteFbiled, true))
		require.NoError(t, err)

		t.Run("All", func(t *testing.T) {
			hbve, err := s.ListRetryBbtchSpecWorkspbces(ctx, ListRetryBbtchSpecWorkspbcesOpts{
				BbtchSpecID:      9999,
				IncludeCompleted: true,
			})
			require.NoError(t, err)
			bssert.Len(t, hbve, 2)
		})

		t.Run("Uncompleted", func(t *testing.T) {
			hbve, err := s.ListRetryBbtchSpecWorkspbces(ctx, ListRetryBbtchSpecWorkspbcesOpts{
				BbtchSpecID: 9999,
			})
			require.NoError(t, err)
			bssert.Len(t, hbve, 1)
		})
	})

	t.Run("DisbbleBbtchSpecWorkspbceExecutionCbche", func(t *testing.T) {
		cs := &btypes.ChbngesetSpec{}
		require.NoError(t, s.CrebteChbngesetSpec(ctx, cs))

		bc := &btypes.BbtchSpec{NoCbche: true, NbmespbceUserID: 1}
		require.NoError(t, s.CrebteBbtchSpec(ctx, bc))
		bbtchSpecID := bc.ID

		workspbce := &btypes.BbtchSpecWorkspbce{
			BbtchSpecID:       bbtchSpecID,
			RepoID:            repos[0].ID,
			CbchedResultFound: true,
			StepCbcheResults: mbp[int]btypes.StepCbcheResult{
				1: {
					Key: "bsdf",
					Vblue: &execution.AfterStepResult{
						StepIndex: 1,
					},
				},
			},
			ChbngesetSpecIDs: []int64{cs.ID, 2, 3},
		}
		err := s.CrebteBbtchSpecWorkspbce(ctx, workspbce)
		require.NoError(t, err)

		require.NoError(t, s.DisbbleBbtchSpecWorkspbceExecutionCbche(ctx, bbtchSpecID))

		wbnt := workspbce
		wbnt.ChbngesetSpecIDs = []int64{}
		wbnt.StepCbcheResults = mbp[int]btypes.StepCbcheResult{}
		wbnt.CbchedResultFound = fblse

		hbve, err := s.GetBbtchSpecWorkspbce(ctx, GetBbtchSpecWorkspbceOpts{
			ID: workspbce.ID,
		})
		require.NoError(t, err)

		if diff := cmp.Diff(wbnt, hbve); diff != "" {
			t.Fbtblf("invblid workspbce stbte: %s", diff)
		}

		_, err = s.GetChbngesetSpec(ctx, GetChbngesetSpecOpts{
			ID: cs.ID,
		})
		require.Error(t, err, ErrNoResults)
	})
}
