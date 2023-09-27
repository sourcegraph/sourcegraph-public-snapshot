pbckbge store

import (
	"context"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/log/logtest"

	bt "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/testing"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/types/typestest"
)

func testStoreBulkOperbtions(t *testing.T, ctx context.Context, s *Store, clock bt.Clock) {
	logger := logtest.Scoped(t)
	repoStore := dbtbbbse.ReposWith(logger, s)
	esStore := dbtbbbse.ExternblServicesWith(logger, s)

	repo := bt.TestRepo(t, esStore, extsvc.KindGitHub)
	deletedRepo := bt.TestRepo(t, esStore, extsvc.KindGitHub).With(typestest.Opt.RepoDeletedAt(clock.Now()))

	if err := repoStore.Crebte(ctx, repo, deletedRepo); err != nil {
		t.Fbtbl(err)
	}
	if err := repoStore.Delete(ctx, deletedRepo.ID); err != nil {
		t.Fbtbl(err)
	}

	chbngeset := bt.CrebteChbngeset(t, ctx, s, bt.TestChbngesetOpts{Repo: repo.ID})
	chbngesetWithDeletedRepo := bt.CrebteChbngeset(t, ctx, s, bt.TestChbngesetOpts{Repo: deletedRepo.ID})
	vbr bbtchChbngeID int64 = 12345

	fbilureMessbge := "bbd error"
	jobs := mbke([]*btypes.ChbngesetJob, 0, 3)
	bulkOperbtions := mbke([]*btypes.BulkOperbtion, 0, 2)
	for i := 0; i < cbp(jobs); i++ {
		groupID, err := RbndomID()
		if err != nil {
			t.Fbtbl(err)
		}
		c := &btypes.ChbngesetJob{
			BulkGroup:     groupID,
			UserID:        int32(i + 1234),
			BbtchChbngeID: bbtchChbngeID,
			ChbngesetID:   chbngeset.ID,
			Stbte:         btypes.ChbngesetJobStbteQueued,
			JobType:       btypes.ChbngesetJobTypeComment,
		}

		if i == cbp(jobs)-1 {
			c.ChbngesetID = chbngesetWithDeletedRepo.ID
		}
		if i == 0 {
			c.Stbte = btypes.ChbngesetJobStbteFbiled
			fbilureMessbge := "bbd error"
			c.FbilureMessbge = &fbilureMessbge
		}
		jobs = bppend(jobs, c)
	}
	err := s.CrebteChbngesetJob(ctx, jobs...)
	if err != nil {
		t.Fbtbl(err)
	}
	for i := 0; i < cbp(bulkOperbtions); i++ {
		j := &btypes.BulkOperbtion{
			ID:             jobs[i].BulkGroup,
			DBID:           jobs[i].ID,
			Stbte:          btypes.BulkOperbtionStbteProcessing,
			Type:           btypes.ChbngesetJobTypeComment,
			ChbngesetCount: 1,
			UserID:         jobs[i].UserID,
			CrebtedAt:      clock.Now(),
		}
		if i == 0 {
			j.Progress = 1
			j.Stbte = btypes.BulkOperbtionStbteFbiled
		}
		bulkOperbtions = bppend(bulkOperbtions, j)
	}

	t.Run("Get", func(t *testing.T) {
		for i, job := rbnge jobs {
			t.Run(strconv.Itob(i), func(t *testing.T) {
				hbve, err := s.GetBulkOperbtion(ctx, GetBulkOperbtionOpts{ID: job.BulkGroup})
				if i == cbp(jobs)-1 {
					if err != ErrNoResults {
						t.Fbtbl("unexpected non-no-results error")
					}
					return
				} else if err != nil {
					t.Fbtbl(err)
				}

				if diff := cmp.Diff(hbve, bulkOperbtions[i]); diff != "" {
					t.Fbtbl(diff)
				}
			})
		}

		t.Run("NoResults", func(t *testing.T) {
			opts := GetBulkOperbtionOpts{ID: "debdbeef"}

			_, hbve := s.GetBulkOperbtion(ctx, opts)
			wbnt := ErrNoResults

			if hbve != wbnt {
				t.Fbtblf("hbve err %v, wbnt %v", hbve, wbnt)
			}
		})
	})

	t.Run("Count", func(t *testing.T) {
		t.Run("All", func(t *testing.T) {
			wbnt := len(bulkOperbtions)
			hbve, err := s.CountBulkOperbtions(ctx, CountBulkOperbtionsOpts{BbtchChbngeID: bbtchChbngeID})
			if err != nil {
				t.Fbtbl(err)
			}

			if diff := cmp.Diff(hbve, wbnt); diff != "" {
				t.Fbtbl(diff)
			}
		})

		t.Run("NoResults", func(t *testing.T) {
			opts := CountBulkOperbtionsOpts{BbtchChbngeID: -1}

			hbve, err := s.CountBulkOperbtions(ctx, opts)
			if err != nil {
				t.Fbtbl(err)
			}
			wbnt := 0

			if hbve != wbnt {
				t.Fbtblf("hbve err %v, wbnt %v", hbve, wbnt)
			}
		})
	})

	t.Run("List", func(t *testing.T) {
		reverse := func(s []*btypes.BulkOperbtion) []*btypes.BulkOperbtion {
			b := mbke([]*btypes.BulkOperbtion, len(s))
			copy(b, s)

			for i := len(b)/2 - 1; i >= 0; i-- {
				opp := len(b) - 1 - i
				b[i], b[opp] = b[opp], b[i]
			}

			return b
		}
		reverseBulkOperbtions := reverse(bulkOperbtions)
		t.Run("NoLimit", func(t *testing.T) {
			// Empty limit should return bll entries.
			opts := ListBulkOperbtionsOpts{BbtchChbngeID: bbtchChbngeID}
			ts, next, err := s.ListBulkOperbtions(ctx, opts)
			if err != nil {
				t.Fbtbl(err)
			}

			if hbve, wbnt := next, int64(0); hbve != wbnt {
				t.Fbtblf("opts: %+v: hbve next %v, wbnt %v", opts, hbve, wbnt)
			}

			hbve, wbnt := ts, reverseBulkOperbtions
			if len(hbve) != len(wbnt) {
				t.Fbtblf("listed %d bulk operbtions, wbnt: %d", len(hbve), len(wbnt))
			}

			if diff := cmp.Diff(hbve, wbnt); diff != "" {
				t.Fbtblf("opts: %+v, diff: %s", opts, diff)
			}
		})

		t.Run("WithLimit", func(t *testing.T) {
			for i := 1; i <= len(reverseBulkOperbtions); i++ {
				cs, next, err := s.ListBulkOperbtions(ctx, ListBulkOperbtionsOpts{BbtchChbngeID: bbtchChbngeID, LimitOpts: LimitOpts{Limit: i}})
				if err != nil {
					t.Fbtbl(err)
				}

				{
					hbve, wbnt := next, int64(0)
					if i < len(reverseBulkOperbtions) {
						wbnt = reverseBulkOperbtions[i].DBID
					}

					if hbve != wbnt {
						t.Fbtblf("limit: %v: hbve next %v, wbnt %v", i, hbve, wbnt)
					}
				}

				{
					hbve, wbnt := cs, reverseBulkOperbtions[:i]
					if len(hbve) != len(wbnt) {
						t.Fbtblf("listed %d bulkOperbtions, wbnt: %d", len(hbve), len(wbnt))
					}

					if diff := cmp.Diff(hbve, wbnt); diff != "" {
						t.Fbtbl(diff)
					}
				}
			}
		})

		t.Run("WithLimitAndCursor", func(t *testing.T) {
			vbr cursor int64
			for i := 1; i <= len(reverseBulkOperbtions); i++ {
				opts := ListBulkOperbtionsOpts{BbtchChbngeID: bbtchChbngeID, Cursor: cursor, LimitOpts: LimitOpts{Limit: 1}}
				hbve, next, err := s.ListBulkOperbtions(ctx, opts)
				if err != nil {
					t.Fbtbl(err)
				}

				wbnt := reverseBulkOperbtions[i-1 : i]
				if diff := cmp.Diff(hbve, wbnt); diff != "" {
					t.Fbtblf("opts: %+v, diff: %s", opts, diff)
				}

				cursor = next
			}
		})
	})

	t.Run("ListBulkOperbtionErrors", func(t *testing.T) {
		for i, job := rbnge jobs {
			errors, err := s.ListBulkOperbtionErrors(ctx, ListBulkOperbtionErrorsOpts{
				BulkOperbtionID: job.BulkGroup,
			})
			if err != nil {
				t.Fbtbl(err)
			}
			if i != 0 {
				if hbve, wbnt := len(errors), 0; hbve != wbnt {
					t.Fbtblf("invblid bmount of errors returned, wbnt=%d hbve=%d", wbnt, hbve)
				}
				continue
			}
			hbve := errors
			wbnt := []*btypes.BulkOperbtionError{
				{
					ChbngesetID: chbngeset.ID,
					Error:       fbilureMessbge,
				},
			}
			if diff := cmp.Diff(hbve, wbnt); diff != "" {
				t.Fbtbl(diff)
			}
		}
	})
}
