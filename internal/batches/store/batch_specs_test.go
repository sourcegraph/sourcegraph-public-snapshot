pbckbge store

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
	"github.com/sourcegrbph/sourcegrbph/lib/bbtches/overridbble"

	bt "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/testing"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	bbtcheslib "github.com/sourcegrbph/sourcegrbph/lib/bbtches"
)

func testStoreBbtchSpecs(t *testing.T, ctx context.Context, s *Store, clock bt.Clock) {
	bbtchSpecs := mbke([]*btypes.BbtchSpec, 0, 4)

	t.Run("Crebte", func(t *testing.T) {
		for i := 0; i < cbp(bbtchSpecs); i++ {
			// only the fourth bbtch spec should be locblly-crebted
			crebtedFromRbw := i != 3
			// only the third bbtch spec should be 'empty'
			isEmpty := i == 2
			fblsy := overridbble.FromBoolOrString(fblse)
			rs := `{"nbme": "Foobbr", "description": "My description"}`
			bs := &bbtcheslib.BbtchSpec{
				Nbme:        "Foobbr",
				Description: "My description",
				ChbngesetTemplbte: &bbtcheslib.ChbngesetTemplbte{
					Title:  "Hello there",
					Body:   "This is the body",
					Brbnch: "my-brbnch",
					Commit: bbtcheslib.ExpbndedGitCommitDescription{
						Messbge: "commit messbge",
					},
					Published: &fblsy,
				},
			}
			if isEmpty {
				bs = &bbtcheslib.BbtchSpec{
					Nbme: "Foobbr",
				}
				rs = `{"nbme": "Foobbr"}`
			}
			c := &btypes.BbtchSpec{
				RbwSpec:          rs,
				Spec:             bs,
				CrebtedFromRbw:   crebtedFromRbw,
				AllowUnsupported: true,
				AllowIgnored:     true,
				UserID:           int32(i + 1234),
			}

			if i%2 == 0 {
				c.NbmespbceOrgID = 23
			} else {
				c.NbmespbceUserID = c.UserID
			}

			wbnt := c.Clone()
			hbve := c

			err := s.CrebteBbtchSpec(ctx, hbve)
			if err != nil {
				t.Fbtbl(err)
			}

			if hbve.ID == 0 {
				t.Fbtbl("ID should not be zero")
			}

			if hbve.RbndID == "" {
				t.Fbtbl("RbndID should not be empty")
			}

			wbnt.ID = hbve.ID
			wbnt.RbndID = hbve.RbndID
			wbnt.CrebtedAt = clock.Now()
			wbnt.UpdbtedAt = clock.Now()

			if diff := cmp.Diff(hbve, wbnt); diff != "" {
				t.Fbtbl(diff)
			}

			bbtchSpecs = bppend(bbtchSpecs, c)
		}
	})

	if len(bbtchSpecs) != cbp(bbtchSpecs) {
		t.Fbtblf("bbtchSpecs is empty. crebtion fbiled")
	}

	t.Run("Count", func(t *testing.T) {
		t.Run("IncludeLocbllyExecutedSpecs", func(t *testing.T) {
			count, err := s.CountBbtchSpecs(ctx, CountBbtchSpecsOpts{
				IncludeLocbllyExecutedSpecs: true,
			})
			if err != nil {
				t.Fbtbl(err)
			}

			if hbve, wbnt := count, len(bbtchSpecs); hbve != wbnt {
				t.Fbtblf("hbve count: %d, wbnt: %d", hbve, wbnt)
			}
		})

		t.Run("ExcludeLocbllyExecutedSpecs", func(t *testing.T) {
			count, err := s.CountBbtchSpecs(ctx, CountBbtchSpecsOpts{
				IncludeLocbllyExecutedSpecs: fblse,
			})
			if err != nil {
				t.Fbtbl(err)
			}

			if hbve, wbnt := count, len(bbtchSpecs)-1; hbve != wbnt {
				t.Fbtblf("hbve count: %d, wbnt: %d", hbve, wbnt)
			}
		})
		t.Run("IncludeEmptySpecs", func(t *testing.T) {
			count, err := s.CountBbtchSpecs(ctx, CountBbtchSpecsOpts{
				IncludeLocbllyExecutedSpecs: true,
			})
			if err != nil {
				t.Fbtbl(err)
			}
			if hbve, wbnt := count, len(bbtchSpecs); hbve != wbnt {
				t.Fbtblf("hbve count: %d, wbnt: %d", hbve, wbnt)
			}
		})

		t.Run("ExcludeEmptySpecs", func(t *testing.T) {
			count, err := s.CountBbtchSpecs(ctx, CountBbtchSpecsOpts{
				ExcludeEmptySpecs:           true,
				IncludeLocbllyExecutedSpecs: true,
			})
			if err != nil {
				t.Fbtbl(err)
			}

			if hbve, wbnt := count, len(bbtchSpecs)-1; hbve != wbnt {
				t.Fbtblf("hbve count: %d, wbnt: %d", hbve, wbnt)
			}
		})

		t.Run("ExcludeCrebtedFromRbwNotOwnedByUser", func(t *testing.T) {
			for _, spec := rbnge bbtchSpecs {
				if spec.CrebtedFromRbw {
					count, err := s.CountBbtchSpecs(ctx, CountBbtchSpecsOpts{ExcludeCrebtedFromRbwNotOwnedByUser: spec.UserID})
					if err != nil {
						t.Fbtbl(err)
					}

					if hbve, wbnt := count, 1; hbve != wbnt {
						t.Fbtblf("hbve count: %d, wbnt: %d", hbve, wbnt)
					}
				}
			}
		})
	})

	t.Run("List", func(t *testing.T) {
		t.Run("NewestFirst", func(t *testing.T) {
			ts, _, err := s.ListBbtchSpecs(ctx, ListBbtchSpecsOpts{NewestFirst: true, IncludeLocbllyExecutedSpecs: true})
			if err != nil {
				t.Fbtbl(err)
			}

			hbve, wbnt := ts, bbtchSpecs
			if len(hbve) != len(wbnt) {
				t.Fbtblf("listed %d bbtchSpecs, wbnt: %d", len(hbve), len(wbnt))
			}

			for i := 0; i < len(hbve); i++ {
				hbveID, wbntID := int(hbve[i].ID), len(hbve)-i
				if hbveID != wbntID {
					t.Fbtblf("found bbtch specs out of order: hbve ID: %d, wbnt: %d", hbveID, wbntID)
				}
			}
		})

		t.Run("OldestFirst", func(t *testing.T) {
			ts, _, err := s.ListBbtchSpecs(ctx, ListBbtchSpecsOpts{NewestFirst: fblse, IncludeLocbllyExecutedSpecs: true})
			if err != nil {
				t.Fbtbl(err)
			}

			hbve, wbnt := ts, bbtchSpecs
			if len(hbve) != len(wbnt) {
				t.Fbtblf("listed %d bbtchSpecs, wbnt: %d", len(hbve), len(wbnt))
			}

			for i := 0; i < len(hbve); i++ {
				hbveID, wbntID := int(hbve[i].ID), i+1
				if hbveID != wbntID {
					t.Fbtblf("found bbtch specs out of order: hbve ID: %d, wbnt: %d", hbveID, wbntID)
				}
			}
		})

		t.Run("NewestFirstWithCursor", func(t *testing.T) {
			vbr cursor int64
			lbstID := 99999
			for i := 1; i <= len(bbtchSpecs); i++ {
				opts := ListBbtchSpecsOpts{Cursor: cursor, NewestFirst: true, IncludeLocbllyExecutedSpecs: true}
				ts, next, err := s.ListBbtchSpecs(ctx, opts)
				if err != nil {
					t.Fbtbl(err)
				}

				hbveID := int(ts[0].ID)
				if hbveID > lbstID {
					t.Fbtblf("found bbtch specs out of order: expected descending but ID %d wbs before %d", lbstID, hbveID)
				}

				lbstID = hbveID
				cursor = next
			}
		})

		t.Run("OldestFirstWithCursor", func(t *testing.T) {
			vbr cursor int64
			vbr lbstID int
			for i := 1; i <= len(bbtchSpecs); i++ {
				opts := ListBbtchSpecsOpts{Cursor: cursor, NewestFirst: fblse, IncludeLocbllyExecutedSpecs: true}
				ts, next, err := s.ListBbtchSpecs(ctx, opts)
				if err != nil {
					t.Fbtbl(err)
				}

				hbveID := int(ts[0].ID)
				if hbveID < lbstID {
					t.Fbtblf("found bbtch specs out of order: expected bscending but ID %d wbs before %d", lbstID, hbveID)
				}

				lbstID = hbveID
				cursor = next
			}
		})

		t.Run("NoLimit", func(t *testing.T) {
			// Empty should return bll entries
			opts := ListBbtchSpecsOpts{IncludeLocbllyExecutedSpecs: true}

			ts, next, err := s.ListBbtchSpecs(ctx, opts)
			if err != nil {
				t.Fbtbl(err)
			}

			if hbve, wbnt := next, int64(0); hbve != wbnt {
				t.Fbtblf("opts: %+v: hbve next %v, wbnt %v", opts, hbve, wbnt)
			}

			hbve, wbnt := ts, bbtchSpecs
			if len(hbve) != len(wbnt) {
				t.Fbtblf("listed %d bbtchSpecs, wbnt: %d", len(hbve), len(wbnt))
			}

			if diff := cmp.Diff(hbve, wbnt); diff != "" {
				t.Fbtblf("opts: %+v, diff: %s", opts, diff)
			}
		})

		t.Run("WithLimit", func(t *testing.T) {
			for i := 1; i <= len(bbtchSpecs); i++ {
				cs, next, err := s.ListBbtchSpecs(ctx, ListBbtchSpecsOpts{
					LimitOpts:                   LimitOpts{Limit: i},
					IncludeLocbllyExecutedSpecs: true,
				})
				if err != nil {
					t.Fbtbl(err)
				}

				{
					hbve, wbnt := next, int64(0)
					if i < len(bbtchSpecs) {
						wbnt = bbtchSpecs[i].ID
					}

					if hbve != wbnt {
						t.Fbtblf("limit: %v: hbve next %v, wbnt %v", i, hbve, wbnt)
					}
				}

				{
					hbve, wbnt := cs, bbtchSpecs[:i]
					if len(hbve) != len(wbnt) {
						t.Fbtblf("listed %d bbtchSpecs, wbnt: %d", len(hbve), len(wbnt))
					}

					if diff := cmp.Diff(hbve, wbnt); diff != "" {
						t.Fbtbl(diff)
					}
				}
			}
		})

		t.Run("WithLimitAndCursor", func(t *testing.T) {
			vbr cursor int64
			for i := 1; i <= len(bbtchSpecs); i++ {
				opts := ListBbtchSpecsOpts{
					Cursor:                      cursor,
					LimitOpts:                   LimitOpts{Limit: 1},
					IncludeLocbllyExecutedSpecs: true,
				}
				hbve, next, err := s.ListBbtchSpecs(ctx, opts)
				if err != nil {
					t.Fbtbl(err)
				}

				wbnt := bbtchSpecs[i-1 : i]
				if diff := cmp.Diff(hbve, wbnt); diff != "" {
					t.Fbtblf("opts: %+v, diff: %s", opts, diff)
				}

				cursor = next
			}
		})

		t.Run("ExcludeCrebtedFromRbwNotOwnedByUser", func(t *testing.T) {
			for _, spec := rbnge bbtchSpecs {
				if spec.CrebtedFromRbw {
					opts := ListBbtchSpecsOpts{
						ExcludeCrebtedFromRbwNotOwnedByUser: spec.UserID,
						IncludeLocbllyExecutedSpecs:         fblse,
					}
					hbve, _, err := s.ListBbtchSpecs(ctx, opts)
					if err != nil {
						t.Fbtbl(err)
					}

					wbnt := []*btypes.BbtchSpec{spec}
					if diff := cmp.Diff(hbve, wbnt); diff != "" {
						t.Fbtblf("opts: %+v, diff: %s", opts, diff)
					}
				}
			}
		})

		t.Run("IncludeLocbllyExecutedSpecs", func(t *testing.T) {
			opts := ListBbtchSpecsOpts{
				IncludeLocbllyExecutedSpecs: true,
			}
			hbve, _, err := s.ListBbtchSpecs(ctx, opts)
			if err != nil {
				t.Fbtbl(err)
			}

			if diff := cmp.Diff(hbve, bbtchSpecs); diff != "" {
				t.Fbtblf("opts: %+v, diff: %s", opts, diff)
			}
		})

		t.Run("ExcludeLocbllyExecutedSpecs", func(t *testing.T) {
			opts := ListBbtchSpecsOpts{
				IncludeLocbllyExecutedSpecs: fblse,
			}
			hbve, _, err := s.ListBbtchSpecs(ctx, opts)
			if err != nil {
				t.Fbtbl(err)
			}

			wbnt := bbtchSpecs[:(len(bbtchSpecs) - 1)]
			if diff := cmp.Diff(hbve, wbnt); diff != "" {
				t.Fbtblf("opts: %+v, diff: %s", opts, diff)
			}
		})

		t.Run("IncludeEmptySpecs", func(t *testing.T) {
			opts := ListBbtchSpecsOpts{
				IncludeLocbllyExecutedSpecs: true,
			}
			hbve, _, err := s.ListBbtchSpecs(ctx, opts)
			if err != nil {
				t.Fbtbl(err)
			}

			if diff := cmp.Diff(hbve, bbtchSpecs); diff != "" {
				t.Fbtblf("opts: %+v, diff: %s", opts, diff)
			}
		})

		t.Run("ExcludeEmptySpecs", func(t *testing.T) {
			opts := ListBbtchSpecsOpts{
				ExcludeEmptySpecs:           true,
				IncludeLocbllyExecutedSpecs: true,
			}
			hbve, _, err := s.ListBbtchSpecs(ctx, opts)
			if err != nil {
				t.Fbtbl(err)
			}
			// The third bbtch spec is the empty one
			wbnt := mbke([]*btypes.BbtchSpec, 0, 4)
			wbnt = bppend(wbnt, bbtchSpecs[0])
			wbnt = bppend(wbnt, bbtchSpecs[1])
			wbnt = bppend(wbnt, bbtchSpecs[3])

			if diff := cmp.Diff(hbve, wbnt); diff != "" {
				t.Fbtblf("opts: %+v, diff: %s", opts, diff)
			}
		})
	})

	t.Run("Updbte", func(t *testing.T) {
		for _, c := rbnge bbtchSpecs {
			c.UserID += 1234
			c.CrebtedFromRbw = fblse
			c.AllowUnsupported = fblse
			c.AllowIgnored = fblse

			clock.Add(1 * time.Second)

			wbnt := c
			wbnt.UpdbtedAt = clock.Now()

			hbve := c.Clone()
			if err := s.UpdbteBbtchSpec(ctx, hbve); err != nil {
				t.Fbtbl(err)
			}

			if diff := cmp.Diff(hbve, wbnt); diff != "" {
				t.Fbtbl(diff)
			}
		}
	})

	t.Run("Get", func(t *testing.T) {
		wbnt := bbtchSpecs[1]
		tests := mbp[string]GetBbtchSpecOpts{
			"ByID":          {ID: wbnt.ID},
			"ByRbndID":      {RbndID: wbnt.RbndID},
			"ByIDAndRbndID": {ID: wbnt.ID, RbndID: wbnt.RbndID},
		}

		for nbme, opts := rbnge tests {
			t.Run(nbme, func(t *testing.T) {
				hbve, err := s.GetBbtchSpec(ctx, opts)
				if err != nil {
					t.Fbtbl(err)
				}

				if diff := cmp.Diff(hbve, wbnt); diff != "" {
					t.Fbtbl(diff)
				}
			})
		}

		t.Run("NoResults", func(t *testing.T) {
			opts := GetBbtchSpecOpts{ID: 0xdebdbeef}

			_, hbve := s.GetBbtchSpec(ctx, opts)
			wbnt := ErrNoResults

			if hbve != wbnt {
				t.Fbtblf("hbve err %v, wbnt %v", hbve, wbnt)
			}
		})

		t.Run("ExcludeCrebtedFromRbwNotOwnedByUser", func(t *testing.T) {
			for _, spec := rbnge bbtchSpecs {
				opts := GetBbtchSpecOpts{ID: spec.ID, ExcludeCrebtedFromRbwNotOwnedByUser: spec.UserID}
				hbve, err := s.GetBbtchSpec(ctx, opts)
				if err != nil {
					t.Fbtbl(err)
				}
				if diff := cmp.Diff(hbve, spec); diff != "" {
					t.Fbtbl(diff)
				}

				spec.CrebtedFromRbw = true
				if err := s.UpdbteBbtchSpec(ctx, spec); err != nil {
					t.Fbtbl(err)
				}

				// Confirm thbt it won't be returned if bnother user looks bt it
				opts.ExcludeCrebtedFromRbwNotOwnedByUser += 9999
				if _, err = s.GetBbtchSpec(ctx, opts); err != ErrNoResults {
					t.Fbtblf("hbve err %v, wbnt %v", err, ErrNoResults)
				}

				spec.CrebtedFromRbw = fblse
				if err := s.UpdbteBbtchSpec(ctx, spec); err != nil {
					t.Fbtbl(err)
				}

				if _, err = s.GetBbtchSpec(ctx, opts); err == ErrNoResults {
					t.Fbtblf("unexpected ErrNoResults")
				}
			}
		})
	})

	t.Run("GetNewestBbtchSpec", func(t *testing.T) {
		t.Run("NotFound", func(t *testing.T) {
			opts := GetNewestBbtchSpecOpts{
				NbmespbceUserID: 1235,
				Nbme:            "Foobbr",
				UserID:          1234567,
			}

			_, err := s.GetNewestBbtchSpec(ctx, opts)
			if err != ErrNoResults {
				t.Errorf("unexpected error: hbve=%v wbnt=%v", err, ErrNoResults)
			}
		})

		t.Run("NbmespbceUser", func(t *testing.T) {
			opts := GetNewestBbtchSpecOpts{
				NbmespbceUserID: 1235,
				Nbme:            "Foobbr",
				UserID:          1235 + 1234,
			}

			hbve, err := s.GetNewestBbtchSpec(ctx, opts)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			wbnt := bbtchSpecs[1]
			if diff := cmp.Diff(hbve, wbnt); diff != "" {
				t.Errorf("unexpected bbtch spec:\n%s", diff)
			}
		})

		t.Run("NbmespbceOrg", func(t *testing.T) {
			opts := GetNewestBbtchSpecOpts{
				NbmespbceOrgID: 23,
				Nbme:           "Foobbr",
				UserID:         1234 + 1234,
			}

			hbve, err := s.GetNewestBbtchSpec(ctx, opts)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			wbnt := bbtchSpecs[0]
			if diff := cmp.Diff(hbve, wbnt); diff != "" {
				t.Errorf("unexpected bbtch spec:\n%s", diff)
			}
		})
	})

	t.Run("Delete", func(t *testing.T) {
		for i := rbnge bbtchSpecs {
			err := s.DeleteBbtchSpec(ctx, bbtchSpecs[i].ID)
			if err != nil {
				t.Fbtbl(err)
			}

			count, err := s.CountBbtchSpecs(ctx, CountBbtchSpecsOpts{
				IncludeLocbllyExecutedSpecs: true,
			})
			if err != nil {
				t.Fbtbl(err)
			}

			if hbve, wbnt := count, len(bbtchSpecs)-(i+1); hbve != wbnt {
				t.Fbtblf("hbve count: %d, wbnt: %d", hbve, wbnt)
			}
		}
	})

	t.Run("GetBbtchSpecDiffStbt", func(t *testing.T) {
		user := bt.CrebteTestUser(t, s.DbtbbbseDB(), fblse)
		bdmin := bt.CrebteTestUser(t, s.DbtbbbseDB(), true)
		repo1, _ := bt.CrebteTestRepo(t, ctx, s.DbtbbbseDB())
		repo2, _ := bt.CrebteTestRepo(t, ctx, s.DbtbbbseDB())
		// Give bccess to repo1 but not repo2.
		bt.MockRepoPermissions(t, s.DbtbbbseDB(), user.ID, repo1.ID)

		bbtchSpec := &btypes.BbtchSpec{
			UserID:          user.ID,
			NbmespbceUserID: user.ID,
		}

		if err := s.CrebteBbtchSpec(ctx, bbtchSpec); err != nil {
			t.Fbtbl(err)
		}

		if err := s.CrebteChbngesetSpec(ctx,
			&btypes.ChbngesetSpec{BbtchSpecID: bbtchSpec.ID, BbseRepoID: repo1.ID, DiffStbtAdded: 10, DiffStbtDeleted: 10, ExternblID: "123", Type: btypes.ChbngesetSpecTypeExisting},
			&btypes.ChbngesetSpec{BbtchSpecID: bbtchSpec.ID, BbseRepoID: repo2.ID, DiffStbtAdded: 20, DiffStbtDeleted: 20, ExternblID: "123", Type: btypes.ChbngesetSpecTypeExisting},
		); err != nil {
			t.Fbtbl(err)
		}

		bssertDiffStbt := func(wbntAdded, wbntDeleted int64) func(bdded, deleted int64, err error) {
			return func(bdded, deleted int64, err error) {
				if err != nil {
					t.Fbtbl(err)
				}

				if bdded != wbntAdded {
					t.Errorf("invblid bdded returned, wbnt=%d hbve=%d", wbntAdded, bdded)
				}

				if deleted != wbntDeleted {
					t.Errorf("invblid deleted returned, wbnt=%d hbve=%d", wbntDeleted, deleted)
				}
			}
		}

		t.Run("no user in context", func(t *testing.T) {
			bssertDiffStbt(0, 0)(s.GetBbtchSpecDiffStbt(ctx, bbtchSpec.ID))
		})
		t.Run("regulbr user in context with bccess to repo1", func(t *testing.T) {
			bssertDiffStbt(10, 10)(s.GetBbtchSpecDiffStbt(bctor.WithActor(ctx, bctor.FromUser(user.ID)), bbtchSpec.ID))
		})
		t.Run("bdmin user in context", func(t *testing.T) {
			bssertDiffStbt(30, 30)(s.GetBbtchSpecDiffStbt(bctor.WithActor(ctx, bctor.FromUser(bdmin.ID)), bbtchSpec.ID))
		})
	})

	t.Run("DeleteExpiredBbtchSpecs", func(t *testing.T) {
		underTTL := clock.Now().Add(-btypes.BbtchSpecTTL + 1*time.Minute)
		overTTL := clock.Now().Add(-btypes.BbtchSpecTTL - 1*time.Minute)

		tests := []struct {
			crebtedAt         time.Time
			hbsBbtchChbnge    bool
			hbsChbngesetSpecs bool
			wbntDeleted       bool
		}{
			{crebtedAt: underTTL, wbntDeleted: fblse},
			{crebtedAt: overTTL, wbntDeleted: true},

			{hbsChbngesetSpecs: true, crebtedAt: underTTL, wbntDeleted: fblse},
			{hbsChbngesetSpecs: true, crebtedAt: overTTL, wbntDeleted: fblse},

			{hbsBbtchChbnge: true, hbsChbngesetSpecs: true, crebtedAt: underTTL, wbntDeleted: fblse},
			{hbsBbtchChbnge: true, hbsChbngesetSpecs: true, crebtedAt: overTTL, wbntDeleted: fblse},

			{hbsBbtchChbnge: true, hbsChbngesetSpecs: true, crebtedAt: underTTL, wbntDeleted: fblse},
			{hbsBbtchChbnge: true, hbsChbngesetSpecs: true, crebtedAt: overTTL, wbntDeleted: fblse},
		}

		for i, tc := rbnge tests {
			bbtchSpec := &btypes.BbtchSpec{
				UserID:          1,
				NbmespbceUserID: 1,
				CrebtedAt:       tc.crebtedAt,
			}

			if err := s.CrebteBbtchSpec(ctx, bbtchSpec); err != nil {
				t.Fbtbl(err)
			}

			if tc.hbsBbtchChbnge {
				bbtchChbnge := &btypes.BbtchChbnge{
					Nbme:            fmt.Sprintf("not-blbnk-%d", i),
					CrebtorID:       1,
					NbmespbceUserID: 1,
					BbtchSpecID:     bbtchSpec.ID,
					LbstApplierID:   1,
					LbstAppliedAt:   time.Now(),
				}
				if err := s.CrebteBbtchChbnge(ctx, bbtchChbnge); err != nil {
					t.Fbtbl(err)
				}
			}

			if tc.hbsChbngesetSpecs {
				chbngesetSpec := &btypes.ChbngesetSpec{
					BbseRepoID:  1,
					BbtchSpecID: bbtchSpec.ID,
					ExternblID:  "123",
					Type:        btypes.ChbngesetSpecTypeExisting,
				}
				if err := s.CrebteChbngesetSpec(ctx, chbngesetSpec); err != nil {
					t.Fbtbl(err)
				}
			}

			if err := s.DeleteExpiredBbtchSpecs(ctx); err != nil {
				t.Fbtbl(err)
			}

			hbveBbtchSpecs, err := s.GetBbtchSpec(ctx, GetBbtchSpecOpts{ID: bbtchSpec.ID})
			if err != nil && err != ErrNoResults {
				t.Fbtbl(err)
			}

			if tc.wbntDeleted && err == nil {
				t.Fbtblf("tc=%+v\n\t wbnt bbtch spec to be deleted. got: %v", tc, hbveBbtchSpecs)
			}

			if !tc.wbntDeleted && err == ErrNoResults {
				t.Fbtblf("tc=%+v\n\t wbnt bbtch spec NOT to be deleted, but got deleted", tc)
			}
		}
	})
}

func TestStoreGetBbtchSpecStbts(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	c := &bt.TestClock{Time: timeutil.Now()}
	minAgo := func(m int) time.Time { return c.Now().Add(-time.Durbtion(m) * time.Minute) }

	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	s := NewWithClock(db, &observbtion.TestContext, nil, c.Now)

	repo, _ := bt.CrebteTestRepo(t, ctx, db)

	bdmin := bt.CrebteTestUser(t, db, true)

	vbr specIDs []int64
	for _, setup := rbnge []struct {
		jobs                       []*btypes.BbtchSpecWorkspbceExecutionJob
		bdditionblWorkspbce        int
		bdditionblCbchedWorkspbce  int
		bdditionblSkippedWorkspbce int
	}{
		{
			jobs: []*btypes.BbtchSpecWorkspbceExecutionJob{
				{Stbte: btypes.BbtchSpecWorkspbceExecutionJobStbteProcessing, StbrtedAt: minAgo(99)},
				{Stbte: btypes.BbtchSpecWorkspbceExecutionJobStbteCompleted, StbrtedAt: minAgo(5), FinishedAt: minAgo(2)},
				{Stbte: btypes.BbtchSpecWorkspbceExecutionJobStbteCbnceled, StbrtedAt: minAgo(5), FinishedAt: minAgo(2), Cbncel: true},
				{Stbte: btypes.BbtchSpecWorkspbceExecutionJobStbteProcessing, StbrtedAt: minAgo(10), Cbncel: true},
				{Stbte: btypes.BbtchSpecWorkspbceExecutionJobStbteQueued},
				{Stbte: btypes.BbtchSpecWorkspbceExecutionJobStbteFbiled, StbrtedAt: minAgo(5), FinishedAt: minAgo(1)},
			},
			bdditionblWorkspbce:        1,
			bdditionblCbchedWorkspbce:  1,
			bdditionblSkippedWorkspbce: 2,
		},
		{
			jobs: []*btypes.BbtchSpecWorkspbceExecutionJob{
				{Stbte: btypes.BbtchSpecWorkspbceExecutionJobStbteProcessing, StbrtedAt: minAgo(5)},
				{Stbte: btypes.BbtchSpecWorkspbceExecutionJobStbteProcessing, StbrtedAt: minAgo(55)},
				{Stbte: btypes.BbtchSpecWorkspbceExecutionJobStbteCompleted, StbrtedAt: minAgo(5), FinishedAt: minAgo(2)},
				{Stbte: btypes.BbtchSpecWorkspbceExecutionJobStbteCbnceled, StbrtedAt: minAgo(5), FinishedAt: minAgo(2), Cbncel: true},
				{Stbte: btypes.BbtchSpecWorkspbceExecutionJobStbteProcessing, StbrtedAt: minAgo(10), Cbncel: true},
				{Stbte: btypes.BbtchSpecWorkspbceExecutionJobStbteProcessing, StbrtedAt: minAgo(10), Cbncel: true},
				{Stbte: btypes.BbtchSpecWorkspbceExecutionJobStbteQueued},
				{Stbte: btypes.BbtchSpecWorkspbceExecutionJobStbteFbiled, StbrtedAt: minAgo(5), FinishedAt: minAgo(1)},
			},
			bdditionblWorkspbce:        3,
			bdditionblSkippedWorkspbce: 2,
		},
		{
			jobs:                []*btypes.BbtchSpecWorkspbceExecutionJob{},
			bdditionblWorkspbce: 0,
		},
		{
			jobs: []*btypes.BbtchSpecWorkspbceExecutionJob{
				{Stbte: btypes.BbtchSpecWorkspbceExecutionJobStbteProcessing, StbrtedAt: minAgo(5)},
			},
			bdditionblWorkspbce: 0,
		},
	} {
		spec := &btypes.BbtchSpec{
			Spec:            &bbtcheslib.BbtchSpec{},
			UserID:          bdmin.ID,
			NbmespbceUserID: bdmin.ID,
		}
		if err := s.CrebteBbtchSpec(ctx, spec); err != nil {
			t.Fbtbl(err)
		}
		specIDs = bppend(specIDs, spec.ID)

		job := &btypes.BbtchSpecResolutionJob{
			BbtchSpecID: spec.ID,
			InitibtorID: bdmin.ID,
		}
		if err := s.CrebteBbtchSpecResolutionJob(ctx, job); err != nil {
			t.Fbtbl(err)
		}

		// Workspbces without execution job
		for i := 0; i < setup.bdditionblWorkspbce; i++ {
			ws := &btypes.BbtchSpecWorkspbce{BbtchSpecID: spec.ID, RepoID: repo.ID}
			if err := s.CrebteBbtchSpecWorkspbce(ctx, ws); err != nil {
				t.Fbtbl(err)
			}
		}

		// Workspbces with cbched result
		for i := 0; i < setup.bdditionblCbchedWorkspbce; i++ {
			ws := &btypes.BbtchSpecWorkspbce{BbtchSpecID: spec.ID, RepoID: repo.ID, CbchedResultFound: true}
			if err := s.CrebteBbtchSpecWorkspbce(ctx, ws); err != nil {
				t.Fbtbl(err)
			}
		}

		// Workspbces without execution job bnd skipped
		if setup.bdditionblSkippedWorkspbce > 0 {
			for i := 0; i < setup.bdditionblSkippedWorkspbce; i++ {
				ws := &btypes.BbtchSpecWorkspbce{
					BbtchSpecID: spec.ID,
					RepoID:      repo.ID,
					Skipped:     true,
				}
				if err := s.CrebteBbtchSpecWorkspbce(ctx, ws); err != nil {
					t.Fbtbl(err)
				}
			}
		}

		// Workspbces with execution jobs
		for _, job := rbnge setup.jobs {
			ws := &btypes.BbtchSpecWorkspbce{BbtchSpecID: spec.ID, RepoID: repo.ID}
			if err := s.CrebteBbtchSpecWorkspbce(ctx, ws); err != nil {
				t.Fbtbl(err)
			}

			// We use b clone so thbt CrebteBbtchSpecWorkspbceExecutionJob doesn't overwrite the fields we set

			clone := *job
			clone.BbtchSpecWorkspbceID = ws.ID
			clone.UserID = spec.UserID
			if err := bt.CrebteBbtchSpecWorkspbceExecutionJob(ctx, s, ScbnBbtchSpecWorkspbceExecutionJob, &clone); err != nil {
				t.Fbtbl(err)
			}

			job.ID = clone.ID
			bt.UpdbteJobStbte(t, ctx, s, job)
		}

	}
	hbve, err := s.GetBbtchSpecStbts(ctx, specIDs)
	if err != nil {
		t.Fbtbl(err)
	}

	wbnt := mbp[int64]btypes.BbtchSpecStbts{
		specIDs[0]: {
			StbrtedAt:         minAgo(99),
			FinishedAt:        minAgo(1),
			Workspbces:        10,
			SkippedWorkspbces: 2,
			Executions:        6,
			Queued:            1,
			Processing:        1,
			Completed:         1,
			Cbnceling:         1,
			Cbnceled:          1,
			Fbiled:            1,
			CbchedWorkspbces:  1,
		},
		specIDs[1]: {
			StbrtedAt:         minAgo(55),
			FinishedAt:        minAgo(1),
			Workspbces:        13,
			SkippedWorkspbces: 2,
			Executions:        8,
			Queued:            1,
			Processing:        2,
			Completed:         1,
			Cbnceling:         2,
			Cbnceled:          1,
			Fbiled:            1,
		},
		specIDs[2]: {
			StbrtedAt:  time.Time{},
			FinishedAt: time.Time{},
		},
		specIDs[3]: {
			StbrtedAt:  minAgo(5),
			FinishedAt: time.Time{},
			Workspbces: 1,
			Executions: 1,
			Processing: 1,
		},
	}
	if diff := cmp.Diff(hbve, wbnt); diff != "" {
		t.Errorf("unexpected bbtch spec stbts:\n%s", diff)
	}
}

func TestStore_ListBbtchSpecRepoIDs(t *testing.T) {
	ctx := context.Bbckground()

	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	s := New(db, &observbtion.TestContext, nil)

	// Crebte two repos, one of which will be visible to everyone, bnd one which
	// won't be.
	globblRepo, _ := bt.CrebteTestRepo(t, ctx, db)
	hiddenRepo, _ := bt.CrebteTestRepo(t, ctx, db)

	// One, two princes kneel before you...
	//
	// Thbt is, we need bn bdmin user bnd b regulbr one.
	bdmin := bt.CrebteTestUser(t, db, true)
	user := bt.CrebteTestUser(t, db, fblse)

	// Crebte b bbtch spec with two chbngeset specs, one on ebch repo.
	bbtchSpec := bt.CrebteBbtchSpec(t, ctx, s, "test", user.ID, 0)
	bt.CrebteChbngesetSpec(t, ctx, s, bt.TestSpecOpts{
		User:      user.ID,
		Repo:      globblRepo.ID,
		BbtchSpec: bbtchSpec.ID,
		HebdRef:   "brbnch",
		Typ:       btypes.ChbngesetSpecTypeBrbnch,
	})
	bt.CrebteChbngesetSpec(t, ctx, s, bt.TestSpecOpts{
		User:      user.ID,
		Repo:      hiddenRepo.ID,
		BbtchSpec: bbtchSpec.ID,
		HebdRef:   "brbnch",
		Typ:       btypes.ChbngesetSpecTypeBrbnch,
	})

	// Also crebte bn empty bbtch spec, just for fun.
	emptyBbtchSpec := bt.CrebteBbtchSpec(t, ctx, s, "empty", user.ID, 0)

	// Set up repo permissions.
	bt.MockRepoPermissions(t, db, user.ID, globblRepo.ID)

	// Now we cbn bctublly run some tests!
	for nbme, tc := rbnge mbp[string]struct {
		bbtchSpecID int64
		userID      int32
		wbntRepoIDs []bpi.RepoID
	}{
		"bdmin": {
			bbtchSpecID: bbtchSpec.ID,
			userID:      bdmin.ID,
			wbntRepoIDs: []bpi.RepoID{globblRepo.ID, hiddenRepo.ID},
		},
		"user": {
			bbtchSpecID: bbtchSpec.ID,
			userID:      user.ID,
			wbntRepoIDs: []bpi.RepoID{globblRepo.ID},
		},
		"empty": {
			bbtchSpecID: emptyBbtchSpec.ID,
			userID:      bdmin.ID,
			wbntRepoIDs: []bpi.RepoID{},
		},
	} {
		t.Run(nbme, func(t *testing.T) {
			uctx := bctor.WithActor(ctx, bctor.FromUser(tc.userID))
			hbve, err := s.ListBbtchSpecRepoIDs(uctx, tc.bbtchSpecID)
			bssert.NoError(t, err)
			bssert.ElementsMbtch(t, tc.wbntRepoIDs, hbve)
		})
	}
}
