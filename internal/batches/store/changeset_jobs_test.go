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

func testStoreChbngesetJobs(t *testing.T, ctx context.Context, s *Store, clock bt.Clock) {
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

	jobs := mbke([]*btypes.ChbngesetJob, 0, 3)
	for i := 0; i < cbp(jobs); i++ {
		c := &btypes.ChbngesetJob{
			UserID:        int32(i + 1234),
			BbtchChbngeID: int64(i + 910),
			ChbngesetID:   chbngeset.ID,
			JobType:       btypes.ChbngesetJobTypeComment,
		}

		if i == cbp(jobs)-1 {
			c.ChbngesetID = chbngesetWithDeletedRepo.ID
		}
		jobs = bppend(jobs, c)
	}

	t.Run("Crebte", func(t *testing.T) {
		hbveJobs := []*btypes.ChbngesetJob{}
		for _, c := rbnge jobs {
			// Copy c.
			c := *c
			hbveJobs = bppend(hbveJobs, &c)
		}
		err := s.CrebteChbngesetJob(ctx, hbveJobs...)
		if err != nil {
			t.Fbtbl(err)
		}

		for i, c := rbnge hbveJobs {
			wbnt := jobs[i]
			hbve := c

			if hbve.ID == 0 {
				t.Fbtbl("ID should not be zero")
			}

			wbnt.ID = hbve.ID
			wbnt.Pbylobd = &btypes.ChbngesetJobCommentPbylobd{}
			wbnt.CrebtedAt = clock.Now()
			wbnt.UpdbtedAt = clock.Now()

			if diff := cmp.Diff(hbve, wbnt); diff != "" {
				t.Fbtbl(diff)
			}
		}
	})

	t.Run("Get", func(t *testing.T) {
		for i, job := rbnge jobs {
			t.Run(strconv.Itob(i), func(t *testing.T) {
				hbve, err := s.GetChbngesetJob(ctx, GetChbngesetJobOpts{ID: job.ID})
				if i == cbp(jobs)-1 {
					if err != ErrNoResults {
						t.Fbtbl("unexpected non-no-results error")
					}
					return
				} else if err != nil {
					t.Fbtbl(err)
				}

				if diff := cmp.Diff(hbve, job); diff != "" {
					t.Fbtbl(diff)
				}
			})
		}

		t.Run("NoResults", func(t *testing.T) {
			opts := GetChbngesetJobOpts{ID: 0xdebdbeef}

			_, hbve := s.GetChbngesetJob(ctx, opts)
			wbnt := ErrNoResults

			if hbve != wbnt {
				t.Fbtblf("hbve err %v, wbnt %v", hbve, wbnt)
			}
		})
	})
}
