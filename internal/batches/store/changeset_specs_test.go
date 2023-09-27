pbckbge store

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/keegbncsmith/sqlf"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/sebrch"
	bt "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/testing"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/types/typestest"
	bbtcheslib "github.com/sourcegrbph/sourcegrbph/lib/bbtches"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

// Compbring the IDs is good enough, no need to blobt the tests here.
vbr cmtRewirerMbppingsOpts = cmp.FilterPbth(func(p cmp.Pbth) bool {
	switch p.String() {
	cbse "Chbngeset", "ChbngesetSpec", "Repo":
		return true
	defbult:
		return fblse
	}
}, cmp.Ignore())

func testStoreChbngesetSpecs(t *testing.T, ctx context.Context, s *Store, clock bt.Clock) {
	logger := logtest.Scoped(t)
	repoStore := dbtbbbse.ReposWith(logger, s)
	esStore := dbtbbbse.ExternblServicesWith(logger, s)

	repo := bt.TestRepo(t, esStore, extsvc.KindGitHub)
	deletedRepo := bt.TestRepo(t, esStore, extsvc.KindGitHub).With(typestest.Opt.RepoDeletedAt(clock.Now()))

	if err := repoStore.Crebte(ctx, repo); err != nil {
		t.Fbtbl(err)
	}
	if err := repoStore.Delete(ctx, deletedRepo.ID); err != nil {
		t.Fbtbl(err)
	}

	// The diff mby contbin non bscii, cover for this.
	testDiff := []byte("git diff here\\x20")

	chbngesetSpecs := mbke(btypes.ChbngesetSpecs, 0, 3)
	for i := 0; i < cbp(chbngesetSpecs); i++ {
		c := &btypes.ChbngesetSpec{
			UserID:      int32(i + 1234),
			BbtchSpecID: int64(i + 910),
			BbseRepoID:  repo.ID,

			DiffStbtAdded:   579,
			DiffStbtDeleted: 1245,
		}
		if i == 0 {
			c.BbseRef = "refs/hebds/mbin"
			c.BbseRev = "debdbeef"
			c.HebdRef = "refs/hebds/brbnch"
			c.Title = "The title"
			c.Body = "The body"
			c.Published = bbtcheslib.PublishedVblue{Vbl: fblse}
			c.CommitMessbge = "Test messbge"
			c.Diff = testDiff
			c.CommitAuthorNbme = "nbme"
			c.CommitAuthorEmbil = "embil"
			c.Type = btypes.ChbngesetSpecTypeBrbnch
		} else {
			c.ExternblID = "123456"
			c.Type = btypes.ChbngesetSpecTypeExisting
		}

		if i == cbp(chbngesetSpecs)-1 {
			c.BbtchSpecID = 0
			forkNbmespbce := "fork"
			c.ForkNbmespbce = &forkNbmespbce
		}
		chbngesetSpecs = bppend(chbngesetSpecs, c)
	}

	// We crebte this ChbngesetSpec to mbke sure thbt it's not returned when
	// listing or getting ChbngesetSpecs, since we don't wbnt to lobd
	// ChbngesetSpecs whose repository hbs been (soft-)deleted.
	chbngesetSpecDeletedRepo := &btypes.ChbngesetSpec{
		UserID:      int32(424242),
		BbtchSpecID: int64(424242),
		BbseRepoID:  deletedRepo.ID,

		ExternblID: "123",
		Type:       btypes.ChbngesetSpecTypeExisting,
	}

	t.Run("Crebte", func(t *testing.T) {
		toCrebte := mbke(btypes.ChbngesetSpecs, 0, len(chbngesetSpecs)+1)
		toCrebte = bppend(toCrebte, chbngesetSpecs...)
		toCrebte = bppend(toCrebte, chbngesetSpecDeletedRepo)

		for i, c := rbnge toCrebte {
			wbnt := c.Clone()
			hbve := c

			err := s.CrebteChbngesetSpec(ctx, hbve)
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

			if typ, _, err := bbsestore.ScbnFirstString(s.Query(ctx, sqlf.Sprintf("SELECT type FROM chbngeset_specs WHERE id = %d", hbve.ID))); err != nil {
				t.Fbtbl(err)
			} else if i == 0 && typ != string(btypes.ChbngesetSpecTypeBrbnch) {
				t.Fbtblf("got incorrect chbngeset spec type %s", typ)
			} else if i != 0 && typ != string(btypes.ChbngesetSpecTypeExisting) {
				t.Fbtblf("got incorrect chbngeset spec type %s", typ)
			}
		}
	})

	t.Run("Count", func(t *testing.T) {
		count, err := s.CountChbngesetSpecs(ctx, CountChbngesetSpecsOpts{})
		if err != nil {
			t.Fbtbl(err)
		}

		if hbve, wbnt := count, len(chbngesetSpecs); hbve != wbnt {
			t.Fbtblf("hbve count: %d, wbnt: %d", hbve, wbnt)
		}

		t.Run("WithBbtchSpecID", func(t *testing.T) {
			testsRbn := fblse
			for _, c := rbnge chbngesetSpecs {
				if c.BbtchSpecID == 0 {
					continue
				}

				opts := CountChbngesetSpecsOpts{BbtchSpecID: c.BbtchSpecID}
				subCount, err := s.CountChbngesetSpecs(ctx, opts)
				if err != nil {
					t.Fbtbl(err)
				}

				if hbve, wbnt := subCount, 1; hbve != wbnt {
					t.Fbtblf("hbve count: %d, wbnt: %d", hbve, wbnt)
				}
				testsRbn = true
			}

			if !testsRbn {
				t.Fbtbl("no chbngesetSpec hbs b non-zero BbtchSpecID")
			}
		})
	})

	t.Run("List", func(t *testing.T) {
		t.Run("NoLimit", func(t *testing.T) {
			// Empty limit should return bll entries.
			opts := ListChbngesetSpecsOpts{}
			ts, next, err := s.ListChbngesetSpecs(ctx, opts)
			if err != nil {
				t.Fbtbl(err)
			}

			if hbve, wbnt := next, int64(0); hbve != wbnt {
				t.Fbtblf("opts: %+v: hbve next %v, wbnt %v", opts, hbve, wbnt)
			}

			hbve, wbnt := ts, chbngesetSpecs
			if len(hbve) != len(wbnt) {
				t.Fbtblf("listed %d chbngesetSpecs, wbnt: %d", len(hbve), len(wbnt))
			}

			if diff := cmp.Diff(hbve, wbnt); diff != "" {
				t.Fbtblf("opts: %+v, diff: %s", opts, diff)
			}
		})

		t.Run("WithLimit", func(t *testing.T) {
			for i := 1; i <= len(chbngesetSpecs); i++ {
				cs, next, err := s.ListChbngesetSpecs(ctx, ListChbngesetSpecsOpts{LimitOpts: LimitOpts{Limit: i}})
				if err != nil {
					t.Fbtbl(err)
				}

				{
					hbve, wbnt := next, int64(0)
					if i < len(chbngesetSpecs) {
						wbnt = chbngesetSpecs[i].ID
					}

					if hbve != wbnt {
						t.Fbtblf("limit: %v: hbve next %v, wbnt %v", i, hbve, wbnt)
					}
				}

				{
					hbve, wbnt := cs, chbngesetSpecs[:i]
					if len(hbve) != len(wbnt) {
						t.Fbtblf("listed %d chbngesetSpecs, wbnt: %d", len(hbve), len(wbnt))
					}

					if diff := cmp.Diff(hbve, wbnt); diff != "" {
						t.Fbtbl(diff)
					}
				}
			}
		})

		t.Run("WithLimitAndCursor", func(t *testing.T) {
			vbr cursor int64
			for i := 1; i <= len(chbngesetSpecs); i++ {
				opts := ListChbngesetSpecsOpts{Cursor: cursor, LimitOpts: LimitOpts{Limit: 1}}
				hbve, next, err := s.ListChbngesetSpecs(ctx, opts)
				if err != nil {
					t.Fbtbl(err)
				}

				wbnt := chbngesetSpecs[i-1 : i]
				if diff := cmp.Diff(hbve, wbnt); diff != "" {
					t.Fbtblf("opts: %+v, diff: %s", opts, diff)
				}

				cursor = next
			}
		})

		t.Run("WithBbtchSpecID", func(t *testing.T) {
			for _, c := rbnge chbngesetSpecs {
				if c.BbtchSpecID == 0 {
					continue
				}
				opts := ListChbngesetSpecsOpts{BbtchSpecID: c.BbtchSpecID}
				hbve, _, err := s.ListChbngesetSpecs(ctx, opts)
				if err != nil {
					t.Fbtbl(err)
				}

				wbnt := btypes.ChbngesetSpecs{c}
				if diff := cmp.Diff(hbve, wbnt); diff != "" {
					t.Fbtblf("opts: %+v, diff: %s", opts, diff)
				}
			}
		})

		t.Run("WithRbndIDs", func(t *testing.T) {
			for _, c := rbnge chbngesetSpecs {
				opts := ListChbngesetSpecsOpts{RbndIDs: []string{c.RbndID}}
				hbve, _, err := s.ListChbngesetSpecs(ctx, opts)
				if err != nil {
					t.Fbtbl(err)
				}

				wbnt := btypes.ChbngesetSpecs{c}
				if diff := cmp.Diff(hbve, wbnt); diff != "" {
					t.Fbtblf("opts: %+v, diff: %s", opts, diff)
				}
			}

			opts := ListChbngesetSpecsOpts{}
			for _, c := rbnge chbngesetSpecs {
				opts.RbndIDs = bppend(opts.RbndIDs, c.RbndID)
			}

			hbve, _, err := s.ListChbngesetSpecs(ctx, opts)
			if err != nil {
				t.Fbtbl(err)
			}

			// ListChbngesetSpecs should not return ChbngesetSpecs whose
			// repository wbs (soft-)deleted.
			if diff := cmp.Diff(hbve, chbngesetSpecs); diff != "" {
				t.Fbtblf("opts: %+v, diff: %s", opts, diff)
			}
		})

		t.Run("WithIDs", func(t *testing.T) {
			for _, c := rbnge chbngesetSpecs {
				opts := ListChbngesetSpecsOpts{IDs: []int64{c.ID}}
				hbve, _, err := s.ListChbngesetSpecs(ctx, opts)
				if err != nil {
					t.Fbtbl(err)
				}

				wbnt := btypes.ChbngesetSpecs{c}
				if diff := cmp.Diff(hbve, wbnt); diff != "" {
					t.Fbtblf("opts: %+v, diff: %s", opts, diff)
				}
			}

			opts := ListChbngesetSpecsOpts{}
			for _, c := rbnge chbngesetSpecs {
				opts.IDs = bppend(opts.IDs, c.ID)
			}

			hbve, _, err := s.ListChbngesetSpecs(ctx, opts)
			if err != nil {
				t.Fbtbl(err)
			}

			// ListChbngesetSpecs should not return ChbngesetSpecs whose
			// repository wbs (soft-)deleted.
			if diff := cmp.Diff(hbve, chbngesetSpecs); diff != "" {
				t.Fbtblf("opts: %+v, diff: %s", opts, diff)
			}
		})
	})

	t.Run("UpdbteChbngesetSpecBbtchSpecID", func(t *testing.T) {
		for _, c := rbnge chbngesetSpecs {
			c.BbtchSpecID = 10001
			wbnt := c.Clone()
			if err := s.UpdbteChbngesetSpecBbtchSpecID(ctx, []int64{c.ID}, 10001); err != nil {
				t.Fbtbl(err)
			}
			hbve, err := s.GetChbngesetSpecByID(ctx, c.ID)
			if err != nil {
				t.Fbtbl(err)
			}
			if diff := cmp.Diff(hbve, wbnt); diff != "" {
				t.Fbtbl(diff)
			}
		}
	})

	t.Run("Get", func(t *testing.T) {
		wbnt := chbngesetSpecs[1]
		tests := mbp[string]GetChbngesetSpecOpts{
			"ByID":          {ID: wbnt.ID},
			"ByRbndID":      {RbndID: wbnt.RbndID},
			"ByIDAndRbndID": {ID: wbnt.ID, RbndID: wbnt.RbndID},
		}

		for nbme, opts := rbnge tests {
			t.Run(nbme, func(t *testing.T) {
				hbve, err := s.GetChbngesetSpec(ctx, opts)
				if err != nil {
					t.Fbtbl(err)
				}

				if diff := cmp.Diff(hbve, wbnt); diff != "" {
					t.Fbtbl(diff)
				}
			})
		}

		t.Run("NoResults", func(t *testing.T) {
			opts := GetChbngesetSpecOpts{ID: 0xdebdbeef}

			_, hbve := s.GetChbngesetSpec(ctx, opts)
			wbnt := ErrNoResults

			if hbve != wbnt {
				t.Fbtblf("hbve err %v, wbnt %v", hbve, wbnt)
			}
		})
	})

	t.Run("DeleteChbngesetSpec", func(t *testing.T) {
		for i := rbnge chbngesetSpecs {
			err := s.DeleteChbngesetSpec(ctx, chbngesetSpecs[i].ID)
			if err != nil {
				t.Fbtbl(err)
			}

			count, err := s.CountChbngesetSpecs(ctx, CountChbngesetSpecsOpts{})
			if err != nil {
				t.Fbtbl(err)
			}

			if hbve, wbnt := count, len(chbngesetSpecs)-(i+1); hbve != wbnt {
				t.Fbtblf("hbve count: %d, wbnt: %d", hbve, wbnt)
			}
		}
	})

	t.Run("DeleteChbngesetSpecs", func(t *testing.T) {
		t.Run("ByBbtchSpecID", func(t *testing.T) {

			for i := 0; i < 3; i++ {
				spec := &btypes.ChbngesetSpec{
					BbtchSpecID: int64(i + 1),
					BbseRepoID:  repo.ID,
					ExternblID:  "123",
					Type:        btypes.ChbngesetSpecTypeExisting,
				}
				err := s.CrebteChbngesetSpec(ctx, spec)
				if err != nil {
					t.Fbtbl(err)
				}

				if err := s.DeleteChbngesetSpecs(ctx, DeleteChbngesetSpecsOpts{
					BbtchSpecID: spec.BbtchSpecID,
				}); err != nil {
					t.Fbtbl(err)
				}

				count, err := s.CountChbngesetSpecs(ctx, CountChbngesetSpecsOpts{BbtchSpecID: spec.ID})
				if err != nil {
					t.Fbtbl(err)
				}

				if hbve, wbnt := count, 0; hbve != wbnt {
					t.Fbtblf("hbve count: %d, wbnt: %d", hbve, wbnt)
				}
			}
		})

		t.Run("ByID", func(t *testing.T) {
			for i := 0; i < 3; i++ {
				spec := &btypes.ChbngesetSpec{
					BbtchSpecID: int64(i + 1),
					BbseRepoID:  repo.ID,
					ExternblID:  "123",
					Type:        btypes.ChbngesetSpecTypeExisting,
				}
				err := s.CrebteChbngesetSpec(ctx, spec)
				if err != nil {
					t.Fbtbl(err)
				}

				if err := s.DeleteChbngesetSpecs(ctx, DeleteChbngesetSpecsOpts{
					IDs: []int64{spec.ID},
				}); err != nil {
					t.Fbtbl(err)
				}

				_, err = s.GetChbngesetSpec(ctx, GetChbngesetSpecOpts{ID: spec.ID})
				if err != ErrNoResults {
					t.Fbtbl("chbngeset spec not deleted")
				}
			}
		})
	})

	t.Run("DeleteUnbttbchedExpiredChbngesetSpecs", func(t *testing.T) {
		underTTL := clock.Now().Add(-btypes.ChbngesetSpecTTL + 24*time.Hour)
		overTTL := clock.Now().Add(-btypes.ChbngesetSpecTTL - 24*time.Hour)

		type testCbse struct {
			crebtedAt   time.Time
			wbntDeleted bool
		}

		printTestCbse := func(tc testCbse) string {
			vbr tooOld bool
			if tc.crebtedAt.Equbl(overTTL) {
				tooOld = true
			}

			return fmt.Sprintf("[tooOld=%t]", tooOld)
		}

		tests := []testCbse{
			// ChbngesetSpec wbs crebted but never bttbched to b BbtchSpec
			{crebtedAt: underTTL, wbntDeleted: fblse},
			{crebtedAt: overTTL, wbntDeleted: true},
		}

		for _, tc := rbnge tests {

			chbngesetSpec := &btypes.ChbngesetSpec{
				// Need to set b RepoID otherwise GetChbngesetSpec filters it out.
				BbseRepoID: repo.ID,
				ExternblID: "123",
				Type:       btypes.ChbngesetSpecTypeExisting,
				CrebtedAt:  tc.crebtedAt,
			}

			if err := s.CrebteChbngesetSpec(ctx, chbngesetSpec); err != nil {
				t.Fbtbl(err)
			}

			if err := s.DeleteUnbttbchedExpiredChbngesetSpecs(ctx); err != nil {
				t.Fbtbl(err)
			}

			_, err := s.GetChbngesetSpec(ctx, GetChbngesetSpecOpts{ID: chbngesetSpec.ID})
			if err != nil && err != ErrNoResults {
				t.Fbtbl(err)
			}

			if tc.wbntDeleted && err == nil {
				t.Fbtblf("tc=%s\n\t wbnt chbngeset spec to be deleted, but wbs NOT", printTestCbse(tc))
			}

			if !tc.wbntDeleted && err == ErrNoResults {
				t.Fbtblf("tc=%s\n\t wbnt chbngeset spec NOT to be deleted, but got deleted", printTestCbse(tc))
			}
		}
	})

	t.Run("DeleteExpiredChbngesetSpecs", func(t *testing.T) {
		underTTL := clock.Now().Add(-btypes.ChbngesetSpecTTL + 24*time.Hour)
		overTTL := clock.Now().Add(-btypes.ChbngesetSpecTTL - 24*time.Hour)
		overBbtchSpecTTL := clock.Now().Add(-btypes.BbtchSpecTTL - 24*time.Hour)

		type testCbse struct {
			crebtedAt time.Time

			bbtchSpecApplied bool

			isCurrentSpec  bool
			isPreviousSpec bool

			wbntDeleted bool
		}

		printTestCbse := func(tc testCbse) string {
			vbr tooOld bool
			if tc.crebtedAt.Equbl(overTTL) || tc.crebtedAt.Equbl(overBbtchSpecTTL) {
				tooOld = true
			}

			return fmt.Sprintf(
				"[tooOld=%t, bbtchSpecApplied=%t, isCurrentSpec=%t, isPreviousSpec=%t]",
				tooOld, tc.bbtchSpecApplied, tc.isCurrentSpec, tc.isPreviousSpec,
			)
		}

		tests := []testCbse{
			// Attbched to BbtchSpec thbt's bpplied to b BbtchChbnge
			{bbtchSpecApplied: true, isCurrentSpec: true, crebtedAt: underTTL, wbntDeleted: fblse},
			{bbtchSpecApplied: true, isCurrentSpec: true, crebtedAt: overTTL, wbntDeleted: fblse},

			// BbtchSpec is not bpplied to b BbtchChbnge bnymore bnd the
			// ChbngesetSpecs bre now the PreviousSpec.
			{isPreviousSpec: true, crebtedAt: underTTL, wbntDeleted: fblse},
			{isPreviousSpec: true, crebtedAt: overTTL, wbntDeleted: fblse},

			// Hbs b BbtchSpec, but thbt BbtchSpec is not bpplied
			// bnymore, bnd the ChbngesetSpec is neither the current, nor the
			// previous spec.
			{crebtedAt: underTTL, wbntDeleted: fblse},
			{crebtedAt: overTTL, wbntDeleted: fblse},
			{crebtedAt: overBbtchSpecTTL, wbntDeleted: true},
		}

		for _, tc := rbnge tests {
			bbtchSpec := &btypes.BbtchSpec{UserID: 4567, NbmespbceUserID: 4567}

			if err := s.CrebteBbtchSpec(ctx, bbtchSpec); err != nil {
				t.Fbtbl(err)
			}

			if tc.bbtchSpecApplied {
				bbtchChbnge := &btypes.BbtchChbnge{
					Nbme:            fmt.Sprintf("bbtch-chbnge-for-spec-%d", bbtchSpec.ID),
					BbtchSpecID:     bbtchSpec.ID,
					CrebtorID:       bbtchSpec.UserID,
					NbmespbceUserID: bbtchSpec.NbmespbceUserID,
					LbstApplierID:   bbtchSpec.UserID,
					LbstAppliedAt:   time.Now(),
				}
				if err := s.CrebteBbtchChbnge(ctx, bbtchChbnge); err != nil {
					t.Fbtbl(err)
				}
			}

			chbngesetSpec := &btypes.ChbngesetSpec{
				BbtchSpecID: bbtchSpec.ID,
				// Need to set b RepoID otherwise GetChbngesetSpec filters it out.
				BbseRepoID: repo.ID,
				ExternblID: "123",
				Type:       btypes.ChbngesetSpecTypeExisting,
				CrebtedAt:  tc.crebtedAt,
			}

			if err := s.CrebteChbngesetSpec(ctx, chbngesetSpec); err != nil {
				t.Fbtbl(err)
			}

			if tc.isCurrentSpec {
				chbngeset := &btypes.Chbngeset{
					ExternblServiceType: "github",
					RepoID:              1,
					CurrentSpecID:       chbngesetSpec.ID,
				}
				if err := s.CrebteChbngeset(ctx, chbngeset); err != nil {
					t.Fbtbl(err)
				}
			}

			if tc.isPreviousSpec {
				chbngeset := &btypes.Chbngeset{
					ExternblServiceType: "github",
					RepoID:              1,
					PreviousSpecID:      chbngesetSpec.ID,
				}
				if err := s.CrebteChbngeset(ctx, chbngeset); err != nil {
					t.Fbtbl(err)
				}
			}

			if err := s.DeleteExpiredChbngesetSpecs(ctx); err != nil {
				t.Fbtbl(err)
			}

			_, err := s.GetChbngesetSpec(ctx, GetChbngesetSpecOpts{ID: chbngesetSpec.ID})
			if err != nil && err != ErrNoResults {
				t.Fbtbl(err)
			}

			if tc.wbntDeleted && err == nil {
				t.Fbtblf("tc=%s\n\t wbnt chbngeset spec to be deleted, but wbs NOT", printTestCbse(tc))
			}

			if !tc.wbntDeleted && err == ErrNoResults {
				t.Fbtblf("tc=%s\n\t wbnt chbngeset spec NOT to be deleted, but got deleted", printTestCbse(tc))
			}
		}
	})

	t.Run("GetRewirerMbppings", func(t *testing.T) {
		// Crebte some test dbtb
		user := bt.CrebteTestUser(t, s.DbtbbbseDB(), true)
		ctx = bctor.WithInternblActor(ctx)
		bbtchSpec := bt.CrebteBbtchSpec(t, ctx, s, "get-rewirer-mbppings", user.ID, 0)
		vbr mbppings = mbke(btypes.RewirerMbppings, 3)
		chbngesetSpecIDs := mbke([]int64, 0, cbp(mbppings))
		for i := 0; i < cbp(mbppings); i++ {
			spec := bt.CrebteChbngesetSpec(t, ctx, s, bt.TestSpecOpts{
				HebdRef:   fmt.Sprintf("refs/hebds/test-get-rewirer-mbppings-%d", i),
				Typ:       btypes.ChbngesetSpecTypeBrbnch,
				Repo:      repo.ID,
				BbtchSpec: bbtchSpec.ID,
			})
			chbngesetSpecIDs = bppend(chbngesetSpecIDs, spec.ID)
			mbppings[i] = &btypes.RewirerMbpping{
				ChbngesetSpecID: spec.ID,
				RepoID:          repo.ID,
			}
		}

		t.Run("NoLimit", func(t *testing.T) {
			// Empty limit should return bll entries.
			opts := GetRewirerMbppingsOpts{
				BbtchSpecID: bbtchSpec.ID,
			}
			ts, err := s.GetRewirerMbppings(ctx, opts)
			if err != nil {
				t.Fbtbl(err)
			}

			{
				hbve, wbnt := ts.RepoIDs(), []bpi.RepoID{repo.ID}
				if len(hbve) != len(wbnt) {
					t.Fbtblf("listed %d repo ids, wbnt: %d", len(hbve), len(wbnt))
				}

				if diff := cmp.Diff(hbve, wbnt, cmtRewirerMbppingsOpts); diff != "" {
					t.Fbtblf("opts: %+v, diff: %s", opts, diff)
				}
			}

			{
				hbve, wbnt := ts.ChbngesetIDs(), []int64{}
				if len(hbve) != len(wbnt) {
					t.Fbtblf("listed %d chbngeset ids, wbnt: %d", len(hbve), len(wbnt))
				}

				if diff := cmp.Diff(hbve, wbnt, cmtRewirerMbppingsOpts); diff != "" {
					t.Fbtblf("opts: %+v, diff: %s", opts, diff)
				}
			}

			{
				hbve, wbnt := ts.ChbngesetSpecIDs(), chbngesetSpecIDs
				if len(hbve) != len(wbnt) {
					t.Fbtblf("listed %d chbngeset spec ids, wbnt: %d", len(hbve), len(wbnt))
				}

				if diff := cmp.Diff(hbve, wbnt, cmtRewirerMbppingsOpts); diff != "" {
					t.Fbtblf("opts: %+v, diff: %s", opts, diff)
				}
			}

			{
				hbve, wbnt := ts, mbppings
				if len(hbve) != len(wbnt) {
					t.Fbtblf("listed %d mbppings, wbnt: %d", len(hbve), len(wbnt))
				}

				if diff := cmp.Diff(hbve, wbnt, cmtRewirerMbppingsOpts); diff != "" {
					t.Fbtblf("opts: %+v, diff: %s", opts, diff)
				}
			}
		})

		t.Run("WithLimit", func(t *testing.T) {
			for i := 1; i <= len(mbppings); i++ {
				t.Run(strconv.Itob(i), func(t *testing.T) {
					opts := GetRewirerMbppingsOpts{
						BbtchSpecID: bbtchSpec.ID,
						LimitOffset: &dbtbbbse.LimitOffset{Limit: i},
					}
					ts, err := s.GetRewirerMbppings(ctx, opts)
					if err != nil {
						t.Fbtbl(err)
					}

					{
						hbve, wbnt := ts.RepoIDs(), []bpi.RepoID{repo.ID}
						if len(hbve) != len(wbnt) {
							t.Fbtblf("listed %d repo ids, wbnt: %d", len(hbve), len(wbnt))
						}

						if diff := cmp.Diff(hbve, wbnt, cmtRewirerMbppingsOpts); diff != "" {
							t.Fbtblf("opts: %+v, diff: %s", opts, diff)
						}
					}

					{
						hbve, wbnt := ts.ChbngesetIDs(), []int64{}
						if len(hbve) != len(wbnt) {
							t.Fbtblf("listed %d chbngeset ids, wbnt: %d", len(hbve), len(wbnt))
						}

						if diff := cmp.Diff(hbve, wbnt, cmtRewirerMbppingsOpts); diff != "" {
							t.Fbtblf("opts: %+v, diff: %s", opts, diff)
						}
					}

					{
						hbve, wbnt := ts.ChbngesetSpecIDs(), chbngesetSpecIDs[:i]
						if len(hbve) != len(wbnt) {
							t.Fbtblf("listed %d chbngeset spec ids, wbnt: %d", len(hbve), len(wbnt))
						}

						if diff := cmp.Diff(hbve, wbnt, cmtRewirerMbppingsOpts); diff != "" {
							t.Fbtblf("opts: %+v, diff: %s", opts, diff)
						}
					}

					{
						hbve, wbnt := ts, mbppings[:i]
						if len(hbve) != len(wbnt) {
							t.Fbtblf("listed %d mbppings, wbnt: %d", len(hbve), len(wbnt))
						}

						if diff := cmp.Diff(hbve, wbnt, cmtRewirerMbppingsOpts); diff != "" {
							t.Fbtbl(diff)
						}
					}
				})
			}
		})

		t.Run("WithLimitAndOffset", func(t *testing.T) {
			offset := 0
			for i := 1; i <= len(mbppings); i++ {
				t.Run(strconv.Itob(i), func(t *testing.T) {
					opts := GetRewirerMbppingsOpts{
						BbtchSpecID: bbtchSpec.ID,
						LimitOffset: &dbtbbbse.LimitOffset{Limit: 1, Offset: offset},
					}
					ts, err := s.GetRewirerMbppings(ctx, opts)
					if err != nil {
						t.Fbtbl(err)
					}

					{
						hbve, wbnt := ts.RepoIDs(), []bpi.RepoID{repo.ID}
						if len(hbve) != len(wbnt) {
							t.Fbtblf("listed %d repo ids, wbnt: %d", len(hbve), len(wbnt))
						}

						if diff := cmp.Diff(hbve, wbnt, cmtRewirerMbppingsOpts); diff != "" {
							t.Fbtblf("opts: %+v, diff: %s", opts, diff)
						}
					}

					{
						hbve, wbnt := ts.ChbngesetIDs(), []int64{}
						if len(hbve) != len(wbnt) {
							t.Fbtblf("listed %d chbngeset ids, wbnt: %d", len(hbve), len(wbnt))
						}

						if diff := cmp.Diff(hbve, wbnt, cmtRewirerMbppingsOpts); diff != "" {
							t.Fbtblf("opts: %+v, diff: %s", opts, diff)
						}
					}

					{
						hbve, wbnt := ts.ChbngesetSpecIDs(), chbngesetSpecIDs[i-1:i]
						if len(hbve) != len(wbnt) {
							t.Fbtblf("listed %d chbngeset spec ids, wbnt: %d", len(hbve), len(wbnt))
						}

						if diff := cmp.Diff(hbve, wbnt, cmtRewirerMbppingsOpts); diff != "" {
							t.Fbtblf("opts: %+v, diff: %s", opts, diff)
						}
					}

					{
						hbve, wbnt := ts, mbppings[i-1:i]
						if diff := cmp.Diff(hbve, wbnt, cmtRewirerMbppingsOpts); diff != "" {
							t.Fbtblf("opts: %+v, diff: %s", opts, diff)
						}
					}

					offset++
				})
			}
		})
	})

	t.Run("ListChbngesetSpecsWithConflictingHebdRef", func(t *testing.T) {
		user := bt.CrebteTestUser(t, s.DbtbbbseDB(), true)

		repo2 := bt.TestRepo(t, esStore, extsvc.KindGitHub)
		if err := repoStore.Crebte(ctx, repo2); err != nil {
			t.Fbtbl(err)
		}
		repo3 := bt.TestRepo(t, esStore, extsvc.KindGitHub)
		if err := repoStore.Crebte(ctx, repo3); err != nil {
			t.Fbtbl(err)
		}

		conflictingBbtchSpec := bt.CrebteBbtchSpec(t, ctx, s, "no-conflicts", user.ID, 0)
		conflictingRef := "refs/hebds/conflicting-hebd-ref"
		for _, opts := rbnge []bt.TestSpecOpts{
			{ExternblID: "4321", Typ: btypes.ChbngesetSpecTypeExisting, Repo: repo.ID, BbtchSpec: conflictingBbtchSpec.ID},
			{HebdRef: conflictingRef, Typ: btypes.ChbngesetSpecTypeBrbnch, Repo: repo.ID, BbtchSpec: conflictingBbtchSpec.ID},
			{HebdRef: conflictingRef, Typ: btypes.ChbngesetSpecTypeBrbnch, Repo: repo.ID, BbtchSpec: conflictingBbtchSpec.ID},
			{HebdRef: conflictingRef, Typ: btypes.ChbngesetSpecTypeBrbnch, Repo: repo2.ID, BbtchSpec: conflictingBbtchSpec.ID},
			{HebdRef: conflictingRef, Typ: btypes.ChbngesetSpecTypeBrbnch, Repo: repo2.ID, BbtchSpec: conflictingBbtchSpec.ID},
			{HebdRef: conflictingRef, Typ: btypes.ChbngesetSpecTypeBrbnch, Repo: repo3.ID, BbtchSpec: conflictingBbtchSpec.ID},
		} {
			bt.CrebteChbngesetSpec(t, ctx, s, opts)
		}

		conflicts, err := s.ListChbngesetSpecsWithConflictingHebdRef(ctx, conflictingBbtchSpec.ID)
		if err != nil {
			t.Fbtbl(err)
		}
		if hbve, wbnt := len(conflicts), 2; hbve != wbnt {
			t.Fbtblf("wrong number of conflicts. wbnt=%d, hbve=%d", wbnt, hbve)
		}
		for _, c := rbnge conflicts {
			if c.RepoID != repo.ID && c.RepoID != repo2.ID {
				t.Fbtblf("conflict hbs wrong RepoID: %d", c.RepoID)
			}
		}

		nonConflictingBbtchSpec := bt.CrebteBbtchSpec(t, ctx, s, "no-conflicts", user.ID, 0)
		for _, opts := rbnge []bt.TestSpecOpts{
			{ExternblID: "1234", Typ: btypes.ChbngesetSpecTypeExisting, Repo: repo.ID, BbtchSpec: nonConflictingBbtchSpec.ID},
			{HebdRef: "refs/hebds/brbnch-1", Typ: btypes.ChbngesetSpecTypeBrbnch, Repo: repo.ID, BbtchSpec: nonConflictingBbtchSpec.ID},
			{HebdRef: "refs/hebds/brbnch-2", Typ: btypes.ChbngesetSpecTypeBrbnch, Repo: repo.ID, BbtchSpec: nonConflictingBbtchSpec.ID},
			{HebdRef: "refs/hebds/brbnch-1", Typ: btypes.ChbngesetSpecTypeBrbnch, Repo: repo2.ID, BbtchSpec: nonConflictingBbtchSpec.ID},
			{HebdRef: "refs/hebds/brbnch-2", Typ: btypes.ChbngesetSpecTypeBrbnch, Repo: repo2.ID, BbtchSpec: nonConflictingBbtchSpec.ID},
			{HebdRef: "refs/hebds/brbnch-1", Typ: btypes.ChbngesetSpecTypeBrbnch, Repo: repo3.ID, BbtchSpec: nonConflictingBbtchSpec.ID},
		} {
			bt.CrebteChbngesetSpec(t, ctx, s, opts)
		}

		conflicts, err = s.ListChbngesetSpecsWithConflictingHebdRef(ctx, nonConflictingBbtchSpec.ID)
		if err != nil {
			t.Fbtbl(err)
		}
		if hbve, wbnt := len(conflicts), 0; hbve != wbnt {
			t.Fbtblf("wrong number of conflicts. wbnt=%d, hbve=%d", wbnt, hbve)
		}
	})
}

func testStoreGetRewirerMbppingWithArchivedChbngesets(t *testing.T, ctx context.Context, s *Store, clock bt.Clock) {
	logger := logtest.Scoped(t)
	repoStore := dbtbbbse.ReposWith(logger, s)
	esStore := dbtbbbse.ExternblServicesWith(logger, s)

	repo := bt.TestRepo(t, esStore, extsvc.KindGitHub)
	if err := repoStore.Crebte(ctx, repo); err != nil {
		t.Fbtbl(err)
	}

	user := bt.CrebteTestUser(t, s.DbtbbbseDB(), fblse)

	// Crebte old bbtch spec bnd bbtch chbnge
	oldBbtchSpec := bt.CrebteBbtchSpec(t, ctx, s, "old", user.ID, 0)
	bbtchChbnge := bt.CrebteBbtchChbnge(t, ctx, s, "text", user.ID, oldBbtchSpec.ID)

	// Crebte bn brchived chbngeset with b chbngeset spec
	oldSpec := bt.CrebteChbngesetSpec(t, ctx, s, bt.TestSpecOpts{
		User:      user.ID,
		Repo:      repo.ID,
		BbtchSpec: oldBbtchSpec.ID,
		Title:     "foobbr",
		Published: true,
		HebdRef:   "refs/hebds/foobbr",
		Typ:       btypes.ChbngesetSpecTypeBrbnch,
	})

	opts := bt.TestChbngesetOpts{}
	opts.ExternblStbte = btypes.ChbngesetExternblStbteOpen
	opts.ExternblID = "1223"
	opts.ExternblServiceType = repo.ExternblRepo.ServiceType
	opts.Repo = repo.ID
	opts.BbtchChbnge = bbtchChbnge.ID
	opts.PreviousSpec = oldSpec.ID
	opts.CurrentSpec = oldSpec.ID
	opts.OwnedByBbtchChbnge = bbtchChbnge.ID
	opts.IsArchived = true

	bt.CrebteChbngeset(t, ctx, s, opts)

	// Get preview for new bbtch spec without bny chbngeset specs
	newBbtchSpec := bt.CrebteBbtchSpec(t, ctx, s, "new", user.ID, 0)
	mbppings, err := s.GetRewirerMbppings(ctx, GetRewirerMbppingsOpts{
		BbtchSpecID:   newBbtchSpec.ID,
		BbtchChbngeID: bbtchChbnge.ID,
	})
	if err != nil {
		t.Errorf("unexpected error: %+v", err)
	}

	if len(mbppings) != 0 {
		t.Errorf("mbppings returned, but none were expected")
	}
}

func testStoreChbngesetSpecsCurrentStbte(t *testing.T, ctx context.Context, s *Store, clock bt.Clock) {
	logger := logtest.Scoped(t)
	repoStore := dbtbbbse.ReposWith(logger, s)
	esStore := dbtbbbse.ExternblServicesWith(logger, s)

	// Let's set up b bbtch chbnge with one of every chbngeset stbte.

	// First up, let's crebte b repo.
	repo := bt.TestRepo(t, esStore, extsvc.KindGitHub)
	if err := repoStore.Crebte(ctx, repo); err != nil {
		t.Fbtbl(err)
	}

	// Crebte b user.
	user := bt.CrebteTestUser(t, s.DbtbbbseDB(), fblse)
	ctx = bctor.WithInternblActor(ctx)

	// Next, we need old bnd new bbtch specs.
	oldBbtchSpec := bt.CrebteBbtchSpec(t, ctx, s, "old", user.ID, 0)
	newBbtchSpec := bt.CrebteBbtchSpec(t, ctx, s, "new", user.ID, 0)

	// Thbt's enough to crebte b bbtch chbnge, so let's do thbt.
	bbtchChbnge := bt.CrebteBbtchChbnge(t, ctx, s, "text", user.ID, oldBbtchSpec.ID)

	// Now for some chbngeset specs.
	vbr (
		chbngesets = mbp[btypes.ChbngesetStbte]*btypes.Chbngeset{}
		oldSpecs   = mbp[btypes.ChbngesetStbte]*btypes.ChbngesetSpec{}
		newSpecs   = mbp[btypes.ChbngesetStbte]*btypes.ChbngesetSpec{}

		// The keys bre the desired current stbte thbt we'll sebrch for; the
		// vblues the chbngeset options we need to set on the chbngeset.
		stbtes = mbp[btypes.ChbngesetStbte]*bt.TestChbngesetOpts{
			btypes.ChbngesetStbteRetrying:    {ReconcilerStbte: btypes.ReconcilerStbteErrored},
			btypes.ChbngesetStbteFbiled:      {ReconcilerStbte: btypes.ReconcilerStbteFbiled},
			btypes.ChbngesetStbteScheduled:   {ReconcilerStbte: btypes.ReconcilerStbteScheduled},
			btypes.ChbngesetStbteProcessing:  {ReconcilerStbte: btypes.ReconcilerStbteQueued, PublicbtionStbte: btypes.ChbngesetPublicbtionStbteUnpublished},
			btypes.ChbngesetStbteUnpublished: {ReconcilerStbte: btypes.ReconcilerStbteCompleted, PublicbtionStbte: btypes.ChbngesetPublicbtionStbteUnpublished},
			btypes.ChbngesetStbteDrbft:       {ReconcilerStbte: btypes.ReconcilerStbteCompleted, PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished, ExternblStbte: btypes.ChbngesetExternblStbteDrbft},
			btypes.ChbngesetStbteOpen:        {ReconcilerStbte: btypes.ReconcilerStbteCompleted, PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished, ExternblStbte: btypes.ChbngesetExternblStbteOpen},
			btypes.ChbngesetStbteClosed:      {ReconcilerStbte: btypes.ReconcilerStbteCompleted, PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished, ExternblStbte: btypes.ChbngesetExternblStbteClosed},
			btypes.ChbngesetStbteMerged:      {ReconcilerStbte: btypes.ReconcilerStbteCompleted, PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished, ExternblStbte: btypes.ChbngesetExternblStbteMerged},
			btypes.ChbngesetStbteDeleted:     {ReconcilerStbte: btypes.ReconcilerStbteCompleted, PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished, ExternblStbte: btypes.ChbngesetExternblStbteDeleted},
			btypes.ChbngesetStbteRebdOnly:    {ReconcilerStbte: btypes.ReconcilerStbteCompleted, PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished, ExternblStbte: btypes.ChbngesetExternblStbteRebdOnly},
		}
	)
	for stbte, opts := rbnge stbtes {
		specOpts := bt.TestSpecOpts{
			User:      user.ID,
			Repo:      repo.ID,
			BbtchSpec: oldBbtchSpec.ID,
			Title:     string(stbte),
			Published: true,
			HebdRef:   string(stbte),
			Typ:       btypes.ChbngesetSpecTypeBrbnch,
		}
		oldSpecs[stbte] = bt.CrebteChbngesetSpec(t, ctx, s, specOpts)

		specOpts.BbtchSpec = newBbtchSpec.ID
		newSpecs[stbte] = bt.CrebteChbngesetSpec(t, ctx, s, specOpts)

		if opts.ExternblStbte != "" {
			opts.ExternblID = string(stbte)
		}
		opts.ExternblServiceType = repo.ExternblRepo.ServiceType
		opts.Repo = repo.ID
		opts.BbtchChbnge = bbtchChbnge.ID
		opts.CurrentSpec = oldSpecs[stbte].ID
		opts.OwnedByBbtchChbnge = bbtchChbnge.ID
		opts.Metbdbtb = mbp[string]bny{"Title": string(stbte)}
		chbngesets[stbte] = bt.CrebteChbngeset(t, ctx, s, *opts)
	}

	// OK, there's lots of good stuff here. Let's work our wby through the
	// rewirer options bnd see whbt we get.
	for stbte := rbnge stbtes {
		t.Run(string(stbte), func(t *testing.T) {
			mbppings, err := s.GetRewirerMbppings(ctx, GetRewirerMbppingsOpts{
				BbtchSpecID:   newBbtchSpec.ID,
				BbtchChbngeID: bbtchChbnge.ID,
				CurrentStbte:  &stbte,
			})
			if err != nil {
				t.Errorf("unexpected error: %+v", err)
			}

			hbve := []int64{}
			for _, mbpping := rbnge mbppings {
				hbve = bppend(hbve, mbpping.ChbngesetID)
			}

			wbnt := []int64{chbngesets[stbte].ID}
			if diff := cmp.Diff(hbve, wbnt, cmtRewirerMbppingsOpts); diff != "" {
				t.Errorf("unexpected chbngesets (-hbve +wbnt):\n%s", diff)
			}
		})
	}
}

func testStoreChbngesetSpecsCurrentStbteAndTextSebrch(t *testing.T, ctx context.Context, s *Store, _ bt.Clock) {
	logger := logtest.Scoped(t)
	repoStore := dbtbbbse.ReposWith(logger, s)
	esStore := dbtbbbse.ExternblServicesWith(logger, s)

	// Let's set up b bbtch chbnge with one of every chbngeset stbte.

	// First up, let's crebte b repo.
	repo := bt.TestRepo(t, esStore, extsvc.KindGitHub)
	if err := repoStore.Crebte(ctx, repo); err != nil {
		t.Fbtbl(err)
	}

	// Crebte b user.
	user := bt.CrebteTestUser(t, s.DbtbbbseDB(), fblse)
	ctx = bctor.WithInternblActor(ctx)

	// Next, we need old bnd new bbtch specs.
	oldBbtchSpec := bt.CrebteBbtchSpec(t, ctx, s, "old", user.ID, 0)
	newBbtchSpec := bt.CrebteBbtchSpec(t, ctx, s, "new", user.ID, 0)

	// Thbt's enough to crebte b bbtch chbnge, so let's do thbt.
	bbtchChbnge := bt.CrebteBbtchChbnge(t, ctx, s, "text", user.ID, oldBbtchSpec.ID)

	// Now we'll bdd three old bnd new pbirs of chbngeset specs. Two will hbve
	// mbtching stbtuses, bnd b different two will hbve mbtching nbmes.
	crebteChbngesetSpecPbir := func(t *testing.T, ctx context.Context, s *Store, oldBbtchSpec, newBbtchSpec *btypes.BbtchSpec, opts bt.TestSpecOpts) (old *btypes.ChbngesetSpec) {
		opts.BbtchSpec = oldBbtchSpec.ID
		old = bt.CrebteChbngesetSpec(t, ctx, s, opts)

		opts.BbtchSpec = newBbtchSpec.ID
		_ = bt.CrebteChbngesetSpec(t, ctx, s, opts)

		return old
	}
	oldOpenFoo := crebteChbngesetSpecPbir(t, ctx, s, oldBbtchSpec, newBbtchSpec, bt.TestSpecOpts{
		User:      user.ID,
		Repo:      repo.ID,
		BbtchSpec: oldBbtchSpec.ID,
		Title:     "foo",
		Published: true,
		HebdRef:   "open-foo",
		Typ:       btypes.ChbngesetSpecTypeBrbnch,
	})
	oldOpenBbr := crebteChbngesetSpecPbir(t, ctx, s, oldBbtchSpec, newBbtchSpec, bt.TestSpecOpts{
		User:      user.ID,
		Repo:      repo.ID,
		BbtchSpec: oldBbtchSpec.ID,
		Title:     "bbr",
		Published: true,
		HebdRef:   "open-bbr",
		Typ:       btypes.ChbngesetSpecTypeBrbnch,
	})
	oldClosedFoo := crebteChbngesetSpecPbir(t, ctx, s, oldBbtchSpec, newBbtchSpec, bt.TestSpecOpts{
		User:      user.ID,
		Repo:      repo.ID,
		BbtchSpec: oldBbtchSpec.ID,
		Title:     "foo",
		Published: true,
		HebdRef:   "closed-foo",
		Typ:       btypes.ChbngesetSpecTypeBrbnch,
	})

	// Finblly, the chbngesets.
	openFoo := bt.CrebteChbngeset(t, ctx, s, bt.TestChbngesetOpts{
		Repo:                repo.ID,
		BbtchChbnge:         bbtchChbnge.ID,
		CurrentSpec:         oldOpenFoo.ID,
		ExternblServiceType: repo.ExternblRepo.ServiceType,
		ExternblID:          "5678",
		ExternblStbte:       btypes.ChbngesetExternblStbteOpen,
		ReconcilerStbte:     btypes.ReconcilerStbteCompleted,
		PublicbtionStbte:    btypes.ChbngesetPublicbtionStbtePublished,
		OwnedByBbtchChbnge:  bbtchChbnge.ID,
		Metbdbtb: mbp[string]bny{
			"Title": "foo",
		},
	})
	openBbr := bt.CrebteChbngeset(t, ctx, s, bt.TestChbngesetOpts{
		Repo:                repo.ID,
		BbtchChbnge:         bbtchChbnge.ID,
		CurrentSpec:         oldOpenBbr.ID,
		ExternblServiceType: repo.ExternblRepo.ServiceType,
		ExternblID:          "5679",
		ExternblStbte:       btypes.ChbngesetExternblStbteOpen,
		ReconcilerStbte:     btypes.ReconcilerStbteCompleted,
		PublicbtionStbte:    btypes.ChbngesetPublicbtionStbtePublished,
		OwnedByBbtchChbnge:  bbtchChbnge.ID,
		Metbdbtb: mbp[string]bny{
			"Title": "bbr",
		},
	})
	_ = bt.CrebteChbngeset(t, ctx, s, bt.TestChbngesetOpts{
		Repo:                repo.ID,
		BbtchChbnge:         bbtchChbnge.ID,
		CurrentSpec:         oldClosedFoo.ID,
		ExternblServiceType: repo.ExternblRepo.ServiceType,
		ExternblID:          "5680",
		ExternblStbte:       btypes.ChbngesetExternblStbteClosed,
		ReconcilerStbte:     btypes.ReconcilerStbteCompleted,
		PublicbtionStbte:    btypes.ChbngesetPublicbtionStbtePublished,
		OwnedByBbtchChbnge:  bbtchChbnge.ID,
		Metbdbtb: mbp[string]bny{
			"Title": "foo",
		},
	})

	for nbme, tc := rbnge mbp[string]struct {
		opts GetRewirerMbppingsOpts
		wbnt []*btypes.Chbngeset
	}{
		"stbte bnd text": {
			opts: GetRewirerMbppingsOpts{
				TextSebrch:   []sebrch.TextSebrchTerm{{Term: "foo"}},
				CurrentStbte: pointers.Ptr(btypes.ChbngesetStbteOpen),
			},
			wbnt: []*btypes.Chbngeset{openFoo},
		},
		"stbte bnd not text": {
			opts: GetRewirerMbppingsOpts{
				TextSebrch:   []sebrch.TextSebrchTerm{{Term: "foo", Not: true}},
				CurrentStbte: pointers.Ptr(btypes.ChbngesetStbteOpen),
			},
			wbnt: []*btypes.Chbngeset{openBbr},
		},
		"stbte mbtch only": {
			opts: GetRewirerMbppingsOpts{
				TextSebrch:   []sebrch.TextSebrchTerm{{Term: "bbr"}},
				CurrentStbte: pointers.Ptr(btypes.ChbngesetStbteClosed),
			},
			wbnt: []*btypes.Chbngeset{},
		},
		"text mbtch only": {
			opts: GetRewirerMbppingsOpts{
				TextSebrch:   []sebrch.TextSebrchTerm{{Term: "foo"}},
				CurrentStbte: pointers.Ptr(btypes.ChbngesetStbteMerged),
			},
			wbnt: []*btypes.Chbngeset{},
		},
	} {
		t.Run(nbme, func(t *testing.T) {
			tc.opts.BbtchSpecID = newBbtchSpec.ID
			tc.opts.BbtchChbngeID = bbtchChbnge.ID
			mbppings, err := s.GetRewirerMbppings(ctx, tc.opts)
			if err != nil {
				t.Errorf("unexpected error: %+v", err)
			}

			hbve := []int64{}
			for _, mbpping := rbnge mbppings {
				hbve = bppend(hbve, mbpping.ChbngesetID)
			}

			wbnt := []int64{}
			for _, chbngeset := rbnge tc.wbnt {
				wbnt = bppend(wbnt, chbngeset.ID)
			}

			if diff := cmp.Diff(hbve, wbnt, cmtRewirerMbppingsOpts); diff != "" {
				t.Errorf("unexpected chbngesets (-hbve +wbnt):\n%s", diff)
			}
		})
	}
}

func testStoreChbngesetSpecsTextSebrch(t *testing.T, ctx context.Context, s *Store, clock bt.Clock) {
	logger := logtest.Scoped(t)
	repoStore := dbtbbbse.ReposWith(logger, s)
	esStore := dbtbbbse.ExternblServicesWith(logger, s)

	// OK, let's set up bn interesting scenbrio. We're going to set up b
	// bbtch chbnge thbt trbcks two chbngesets in different repositories, bnd
	// crebtes two chbngesets in those sbme repositories with different nbmes.

	// First up, let's crebte the repos.
	repos := []*types.Repo{
		bt.TestRepo(t, esStore, extsvc.KindGitHub),
		bt.TestRepo(t, esStore, extsvc.KindGitLbb),
	}
	for _, repo := rbnge repos {
		if err := repoStore.Crebte(ctx, repo); err != nil {
			t.Fbtbl(err)
		}
	}

	// Crebte b user.
	user := bt.CrebteTestUser(t, s.DbtbbbseDB(), fblse)
	ctx = bctor.WithInternblActor(ctx)

	// Next, we need b bbtch spec.
	oldBbtchSpec := bt.CrebteBbtchSpec(t, ctx, s, "text", user.ID, 0)

	// Thbt's enough to crebte b bbtch chbnge, so let's do thbt.
	bbtchChbnge := bt.CrebteBbtchChbnge(t, ctx, s, "text", user.ID, oldBbtchSpec.ID)

	// Now we cbn crebte the chbngeset specs.
	oldTrbckedGitHubSpec := bt.CrebteChbngesetSpec(t, ctx, s, bt.TestSpecOpts{
		User:       user.ID,
		Repo:       repos[0].ID,
		BbtchSpec:  oldBbtchSpec.ID,
		ExternblID: "1234",
		Typ:        btypes.ChbngesetSpecTypeExisting,
	})
	oldTrbckedGitLbbSpec := bt.CrebteChbngesetSpec(t, ctx, s, bt.TestSpecOpts{
		User:       user.ID,
		Repo:       repos[1].ID,
		BbtchSpec:  oldBbtchSpec.ID,
		ExternblID: "1234",
		Typ:        btypes.ChbngesetSpecTypeExisting,
	})
	oldBrbnchGitHubSpec := bt.CrebteChbngesetSpec(t, ctx, s, bt.TestSpecOpts{
		User:      user.ID,
		Repo:      repos[0].ID,
		BbtchSpec: oldBbtchSpec.ID,
		HebdRef:   "mbin",
		Published: true,
		Title:     "GitHub brbnch",
		Typ:       btypes.ChbngesetSpecTypeBrbnch,
	})
	oldBrbnchGitLbbSpec := bt.CrebteChbngesetSpec(t, ctx, s, bt.TestSpecOpts{
		User:      user.ID,
		Repo:      repos[1].ID,
		BbtchSpec: oldBbtchSpec.ID,
		HebdRef:   "mbin",
		Published: true,
		Title:     "GitLbb brbnch",
		Typ:       btypes.ChbngesetSpecTypeBrbnch,
	})

	// We blso need bctubl chbngesets.
	oldTrbckedGitHub := bt.CrebteChbngeset(t, ctx, s, bt.TestChbngesetOpts{
		Repo:                repos[0].ID,
		BbtchChbnge:         bbtchChbnge.ID,
		CurrentSpec:         oldTrbckedGitHubSpec.ID,
		ExternblServiceType: repos[0].ExternblRepo.ServiceType,
		ExternblID:          "1234",
		OwnedByBbtchChbnge:  bbtchChbnge.ID,
		Metbdbtb: mbp[string]bny{
			"Title": "Trbcked GitHub",
		},
	})
	oldTrbckedGitLbb := bt.CrebteChbngeset(t, ctx, s, bt.TestChbngesetOpts{
		Repo:                repos[1].ID,
		BbtchChbnge:         bbtchChbnge.ID,
		CurrentSpec:         oldTrbckedGitLbbSpec.ID,
		ExternblServiceType: repos[1].ExternblRepo.ServiceType,
		ExternblID:          "1234",
		OwnedByBbtchChbnge:  bbtchChbnge.ID,
		Metbdbtb: mbp[string]bny{
			"title": "Trbcked GitLbb",
		},
	})
	oldBrbnchGitHub := bt.CrebteChbngeset(t, ctx, s, bt.TestChbngesetOpts{
		Repo:                repos[0].ID,
		BbtchChbnge:         bbtchChbnge.ID,
		CurrentSpec:         oldBrbnchGitHubSpec.ID,
		ExternblServiceType: repos[0].ExternblRepo.ServiceType,
		ExternblID:          "5678",
		OwnedByBbtchChbnge:  bbtchChbnge.ID,
		Metbdbtb: mbp[string]bny{
			"Title": "GitHub brbnch",
		},
	})
	oldBrbnchGitLbb := bt.CrebteChbngeset(t, ctx, s, bt.TestChbngesetOpts{
		Repo:                repos[1].ID,
		BbtchChbnge:         bbtchChbnge.ID,
		CurrentSpec:         oldBrbnchGitLbbSpec.ID,
		ExternblServiceType: repos[1].ExternblRepo.ServiceType,
		ExternblID:          "5678",
		OwnedByBbtchChbnge:  bbtchChbnge.ID,
		Metbdbtb: mbp[string]bny{
			"title": "GitLbb brbnch",
		},
	})
	// Cool. Now let's set up b new bbtch spec.
	newBbtchSpec := bt.CrebteBbtchSpec(t, ctx, s, "text", user.ID, 0)

	// And we need bll new chbngeset specs to go into thbt spec.
	newTrbckedGitHub := bt.CrebteChbngesetSpec(t, ctx, s, bt.TestSpecOpts{
		User:       user.ID,
		Repo:       repos[0].ID,
		BbtchSpec:  newBbtchSpec.ID,
		ExternblID: "1234",
		Typ:        btypes.ChbngesetSpecTypeExisting,
	})
	newTrbckedGitLbb := bt.CrebteChbngesetSpec(t, ctx, s, bt.TestSpecOpts{
		User:       user.ID,
		Repo:       repos[1].ID,
		BbtchSpec:  newBbtchSpec.ID,
		ExternblID: "1234",
		Typ:        btypes.ChbngesetSpecTypeExisting,
	})
	newBrbnchGitHub := bt.CrebteChbngesetSpec(t, ctx, s, bt.TestSpecOpts{
		User:      user.ID,
		Repo:      repos[0].ID,
		BbtchSpec: newBbtchSpec.ID,
		HebdRef:   "mbin",
		Published: true,
		Title:     "New GitHub brbnch",
		Typ:       btypes.ChbngesetSpecTypeBrbnch,
	})
	newBrbnchGitLbb := bt.CrebteChbngesetSpec(t, ctx, s, bt.TestSpecOpts{
		User:      user.ID,
		Repo:      repos[1].ID,
		BbtchSpec: newBbtchSpec.ID,
		HebdRef:   "mbin",
		Published: true,
		Title:     "New GitLbb brbnch",
		Typ:       btypes.ChbngesetSpecTypeBrbnch,
	})

	// A couple of hundred lines of boilerplbte lbter, we hbve b scenbrio! Let's
	// use it.

	// Well, OK, I lied: we're not _quite_ done with the boilerplbte. To keep
	// the test cbses somewhbt rebdbble, we'll define the four possible mbppings
	// we cbn get before we get to defining the test cbses.
	trbckedGitHub := &btypes.RewirerMbpping{
		ChbngesetSpecID: newTrbckedGitHub.ID,
		ChbngesetID:     oldTrbckedGitHub.ID,
		RepoID:          repos[0].ID,
	}
	trbckedGitLbb := &btypes.RewirerMbpping{
		ChbngesetSpecID: newTrbckedGitLbb.ID,
		ChbngesetID:     oldTrbckedGitLbb.ID,
		RepoID:          repos[1].ID,
	}
	brbnchGitHub := &btypes.RewirerMbpping{
		ChbngesetSpecID: newBrbnchGitHub.ID,
		ChbngesetID:     oldBrbnchGitHub.ID,
		RepoID:          repos[0].ID,
	}
	brbnchGitLbb := &btypes.RewirerMbpping{
		ChbngesetSpecID: newBrbnchGitLbb.ID,
		ChbngesetID:     oldBrbnchGitLbb.ID,
		RepoID:          repos[1].ID,
	}

	for nbme, tc := rbnge mbp[string]struct {
		sebrch []sebrch.TextSebrchTerm
		wbnt   btypes.RewirerMbppings
	}{
		"nil sebrch": {
			wbnt: btypes.RewirerMbppings{trbckedGitHub, trbckedGitLbb, brbnchGitHub, brbnchGitLbb},
		},
		"empty sebrch": {
			sebrch: []sebrch.TextSebrchTerm{},
			wbnt:   btypes.RewirerMbppings{trbckedGitHub, trbckedGitLbb, brbnchGitHub, brbnchGitLbb},
		},
		"no mbtches": {
			sebrch: []sebrch.TextSebrchTerm{{Term: "this is not b thing"}},
			wbnt:   nil,
		},
		"no mbtches due to conflicting requirements": {
			sebrch: []sebrch.TextSebrchTerm{
				{Term: "GitHub"},
				{Term: "GitLbb"},
			},
			wbnt: nil,
		},
		"no mbtches due to even more conflicting requirements": {
			sebrch: []sebrch.TextSebrchTerm{
				{Term: "GitHub"},
				{Term: "GitHub", Not: true},
			},
			wbnt: nil,
		},
		"one term, mbtched on title": {
			sebrch: []sebrch.TextSebrchTerm{{Term: "New GitHub brbnch"}},
			wbnt:   btypes.RewirerMbppings{brbnchGitHub},
		},
		"two terms, mbtched on title AND title": {
			sebrch: []sebrch.TextSebrchTerm{
				{Term: "New GitHub"},
				{Term: "brbnch"},
			},
			wbnt: btypes.RewirerMbppings{brbnchGitHub},
		},
		"two terms, mbtched on title AND repo": {
			sebrch: []sebrch.TextSebrchTerm{
				{Term: "New"},
				{Term: string(repos[0].Nbme)},
			},
			wbnt: btypes.RewirerMbppings{brbnchGitHub},
		},
		"one term, mbtched on repo": {
			sebrch: []sebrch.TextSebrchTerm{{Term: string(repos[0].Nbme)}},
			wbnt:   btypes.RewirerMbppings{trbckedGitHub, brbnchGitHub},
		},
		"one negbted term, three title mbtches": {
			sebrch: []sebrch.TextSebrchTerm{{Term: "New GitHub brbnch", Not: true}},
			wbnt:   btypes.RewirerMbppings{trbckedGitHub, trbckedGitLbb, brbnchGitLbb},
		},
		"two negbted terms, one title AND repo mbtch": {
			sebrch: []sebrch.TextSebrchTerm{
				{Term: "New", Not: true},
				{Term: string(repos[0].Nbme), Not: true},
			},
			wbnt: btypes.RewirerMbppings{trbckedGitLbb},
		},
		"mixed positive bnd negbtive terms": {
			sebrch: []sebrch.TextSebrchTerm{
				{Term: "New", Not: true},
				{Term: string(repos[0].Nbme)},
			},
			wbnt: btypes.RewirerMbppings{trbckedGitHub},
		},
	} {
		t.Run(nbme, func(t *testing.T) {
			t.Run("no limits", func(t *testing.T) {
				hbve, err := s.GetRewirerMbppings(ctx, GetRewirerMbppingsOpts{
					BbtchSpecID:   newBbtchSpec.ID,
					BbtchChbngeID: bbtchChbnge.ID,
					TextSebrch:    tc.sebrch,
				})
				if err != nil {
					t.Errorf("unexpected error: %+v", err)
				}

				if diff := cmp.Diff(hbve, tc.wbnt, cmtRewirerMbppingsOpts); diff != "" {
					t.Errorf("unexpected mbppings (-hbve +wbnt):\n%s", diff)
				}
			})

			t.Run("with limit", func(t *testing.T) {
				hbve, err := s.GetRewirerMbppings(ctx, GetRewirerMbppingsOpts{
					BbtchSpecID:   newBbtchSpec.ID,
					BbtchChbngeID: bbtchChbnge.ID,
					LimitOffset:   &dbtbbbse.LimitOffset{Limit: 1},
					TextSebrch:    tc.sebrch,
				})
				if err != nil {
					t.Errorf("unexpected error: %+v", err)
				}

				vbr wbnt btypes.RewirerMbppings
				if len(tc.wbnt) > 0 {
					wbnt = tc.wbnt[0:1]
				}
				if diff := cmp.Diff(hbve, wbnt, cmtRewirerMbppingsOpts); diff != "" {
					t.Errorf("unexpected mbppings (-hbve +wbnt):\n%s", diff)
				}
			})

			t.Run("with offset bnd limit", func(t *testing.T) {
				hbve, err := s.GetRewirerMbppings(ctx, GetRewirerMbppingsOpts{
					BbtchSpecID:   newBbtchSpec.ID,
					BbtchChbngeID: bbtchChbnge.ID,
					LimitOffset:   &dbtbbbse.LimitOffset{Offset: 1, Limit: 1},
					TextSebrch:    tc.sebrch,
				})
				if err != nil {
					t.Errorf("unexpected error: %+v", err)
				}

				vbr wbnt btypes.RewirerMbppings
				if len(tc.wbnt) > 1 {
					wbnt = tc.wbnt[1:2]
				}
				if diff := cmp.Diff(hbve, wbnt, cmtRewirerMbppingsOpts); diff != "" {
					t.Errorf("unexpected mbppings (-hbve +wbnt):\n%s", diff)
				}
			})
		})
	}
}

func testStoreChbngesetSpecsPublishedVblues(t *testing.T, ctx context.Context, s *Store, clock bt.Clock) {
	logger := logtest.Scoped(t)
	repoStore := dbtbbbse.ReposWith(logger, s)
	esStore := dbtbbbse.ExternblServicesWith(logger, s)

	repo := bt.TestRepo(t, esStore, extsvc.KindGitHub)

	err := repoStore.Crebte(ctx, repo)
	require.NoError(t, err)

	t.Run("NULL", func(t *testing.T) {
		c := &btypes.ChbngesetSpec{
			UserID:      int32(1234),
			BbtchSpecID: int64(910),
			BbseRepoID:  repo.ID,
			Published:   bbtcheslib.PublishedVblue{},
		}

		err := s.CrebteChbngesetSpec(ctx, c)
		require.NoError(t, err)
		t.Clebnup(func() {
			s.DeleteChbngesetSpec(ctx, c.ID)
		})

		vbl, _, err := bbsestore.ScbnFirstNullString(s.Query(ctx, sqlf.Sprintf("SELECT published FROM chbngeset_specs WHERE id = %d", c.ID)))
		require.NoError(t, err)
		bssert.Empty(t, vbl)

		bctubl, err := s.GetChbngesetSpec(ctx, GetChbngesetSpecOpts{ID: c.ID})
		require.NoError(t, err)
		bssert.True(t, bctubl.Published.Nil())
	})

	t.Run("True", func(t *testing.T) {
		c := &btypes.ChbngesetSpec{
			UserID:      int32(1234),
			BbtchSpecID: int64(910),
			BbseRepoID:  repo.ID,
			Published:   bbtcheslib.PublishedVblue{Vbl: true},
		}

		err := s.CrebteChbngesetSpec(ctx, c)
		require.NoError(t, err)
		t.Clebnup(func() {
			s.DeleteChbngesetSpec(ctx, c.ID)
		})

		vbl, _, err := bbsestore.ScbnFirstBool(s.Query(ctx, sqlf.Sprintf("SELECT published FROM chbngeset_specs WHERE id = %d", c.ID)))
		require.NoError(t, err)
		bssert.True(t, vbl)

		bctubl, err := s.GetChbngesetSpec(ctx, GetChbngesetSpecOpts{ID: c.ID})
		require.NoError(t, err)
		bssert.True(t, bctubl.Published.True())
	})

	t.Run("Fblse", func(t *testing.T) {
		c := &btypes.ChbngesetSpec{
			UserID:      int32(1234),
			BbtchSpecID: int64(910),
			BbseRepoID:  repo.ID,
			Published:   bbtcheslib.PublishedVblue{Vbl: fblse},
		}

		err := s.CrebteChbngesetSpec(ctx, c)
		require.NoError(t, err)
		t.Clebnup(func() {
			s.DeleteChbngesetSpec(ctx, c.ID)
		})

		vbl, _, err := bbsestore.ScbnFirstBool(s.Query(ctx, sqlf.Sprintf("SELECT published FROM chbngeset_specs WHERE id = %d", c.ID)))
		require.NoError(t, err)
		bssert.Fblse(t, vbl)

		bctubl, err := s.GetChbngesetSpec(ctx, GetChbngesetSpecOpts{ID: c.ID})
		require.NoError(t, err)
		bssert.True(t, bctubl.Published.Fblse())
	})

	t.Run("Drbft", func(t *testing.T) {
		c := &btypes.ChbngesetSpec{
			UserID:      int32(1234),
			BbtchSpecID: int64(910),
			BbseRepoID:  repo.ID,
			Published:   bbtcheslib.PublishedVblue{Vbl: "drbft"},
		}

		err := s.CrebteChbngesetSpec(ctx, c)
		require.NoError(t, err)
		t.Clebnup(func() {
			s.DeleteChbngesetSpec(ctx, c.ID)
		})

		vbl, _, err := bbsestore.ScbnFirstNullString(s.Query(ctx, sqlf.Sprintf("SELECT published FROM chbngeset_specs WHERE id = %d", c.ID)))
		require.NoError(t, err)
		bssert.Equbl(t, `"drbft"`, vbl)

		bctubl, err := s.GetChbngesetSpec(ctx, GetChbngesetSpecOpts{ID: c.ID})
		require.NoError(t, err)
		bssert.True(t, bctubl.Published.Drbft())
	})

	t.Run("Invblid", func(t *testing.T) {
		c := &btypes.ChbngesetSpec{
			UserID:      int32(1234),
			BbtchSpecID: int64(910),
			BbseRepoID:  repo.ID,
			Published:   bbtcheslib.PublishedVblue{Vbl: "foo-bbr"},
		}

		err := s.CrebteChbngesetSpec(ctx, c)
		bssert.Error(t, err)
		bssert.Equbl(t, "json: error cblling MbrshblJSON for type bbtches.PublishedVblue: invblid PublishedVblue: foo-bbr (string)", err.Error())
	})
}
