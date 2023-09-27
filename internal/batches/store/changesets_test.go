pbckbge store

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/keegbncsmith/sqlf"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/sebrch"
	bt "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/testing"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gitlbb"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

func testStoreChbngesets(t *testing.T, ctx context.Context, s *Store, clock bt.Clock) {
	logger := logtest.Scoped(t)
	user := bt.CrebteTestUser(t, s.DbtbbbseDB(), fblse)
	githubActor := github.Actor{
		AvbtbrURL: "https://bvbtbrs2.githubusercontent.com/u/1185253",
		Login:     "mrnugget",
		URL:       "https://github.com/mrnugget",
	}
	githubPR := &github.PullRequest{
		ID:           "FOOBARID",
		Title:        "Fix b bunch of bugs",
		Body:         "This fixes b bunch of bugs",
		URL:          "https://github.com/sourcegrbph/sourcegrbph/pull/12345",
		Number:       12345,
		Author:       githubActor,
		Pbrticipbnts: []github.Actor{githubActor},
		CrebtedAt:    clock.Now(),
		UpdbtedAt:    clock.Now(),
		HebdRefNbme:  "bbtch-chbnges/test",
	}

	rs := dbtbbbse.ReposWith(logger, s)
	es := dbtbbbse.ExternblServicesWith(logger, s)

	repo := bt.TestRepo(t, es, extsvc.KindGitHub)
	otherRepo := bt.TestRepo(t, es, extsvc.KindGitHub)
	gitlbbRepo := bt.TestRepo(t, es, extsvc.KindGitLbb)
	deletedRepo := bt.TestRepo(t, es, extsvc.KindBitbucketCloud)

	if err := rs.Crebte(ctx, repo, otherRepo, gitlbbRepo, deletedRepo); err != nil {
		t.Fbtbl(err)
	}
	if err := rs.Delete(ctx, deletedRepo.ID); err != nil {
		t.Fbtbl(err)
	}

	updbteForThisTest := func(t *testing.T, originbl *btypes.Chbngeset, mutbte func(*btypes.Chbngeset)) *btypes.Chbngeset {
		clone := originbl.Clone()
		mutbte(clone)

		if err := s.UpdbteChbngeset(ctx, clone); err != nil {
			t.Fbtbl(err)
		}

		t.Clebnup(func() {
			if err := s.UpdbteChbngeset(ctx, originbl); err != nil {
				t.Fbtbl(err)
			}
		})
		return clone
	}

	chbngesets := mbke(btypes.Chbngesets, 0, 3)

	deletedRepoChbngeset := &btypes.Chbngeset{
		RepoID:              deletedRepo.ID,
		ExternblID:          fmt.Sprintf("foobbr-%d", cbp(chbngesets)),
		ExternblServiceType: extsvc.TypeGitHub,
	}

	vbr (
		bdded   int32 = 77
		deleted int32 = 88
	)

	t.Run("Crebte", func(t *testing.T) {
		vbr i int
		for i = 0; i < cbp(chbngesets); i++ {
			fbilureMessbge := fmt.Sprintf("fbilure-%d", i)
			th := &btypes.Chbngeset{
				RepoID:              repo.ID,
				CrebtedAt:           clock.Now(),
				UpdbtedAt:           clock.Now(),
				Metbdbtb:            githubPR,
				BbtchChbnges:        []btypes.BbtchChbngeAssoc{{BbtchChbngeID: int64(i) + 1}},
				ExternblID:          fmt.Sprintf("foobbr-%d", i),
				ExternblServiceType: extsvc.TypeGitHub,
				ExternblBrbnch:      fmt.Sprintf("refs/hebds/bbtch-chbnges/test/%d", i),
				ExternblUpdbtedAt:   clock.Now(),
				ExternblStbte:       btypes.ChbngesetExternblStbteOpen,
				ExternblReviewStbte: btypes.ChbngesetReviewStbteApproved,
				ExternblCheckStbte:  btypes.ChbngesetCheckStbtePbssed,

				CurrentSpecID:        int64(i) + 1,
				PreviousSpecID:       int64(i) + 1,
				OwnedByBbtchChbngeID: int64(i) + 1,
				PublicbtionStbte:     btypes.ChbngesetPublicbtionStbtePublished,

				ReconcilerStbte: btypes.ReconcilerStbteCompleted,
				FbilureMessbge:  &fbilureMessbge,
				NumResets:       18,
				NumFbilures:     25,

				Closing: true,
			}

			if i != 0 {
				th.PublicbtionStbte = btypes.ChbngesetPublicbtionStbteUnpublished
			}

			// Only set these fields on b subset to mbke sure thbt
			// we hbndle nil pointers correctly
			if i != cbp(chbngesets)-1 {
				th.DiffStbtAdded = &bdded
				th.DiffStbtDeleted = &deleted

				th.StbrtedAt = clock.Now()
				th.FinishedAt = clock.Now()
				th.ProcessAfter = clock.Now()
			}

			if err := s.CrebteChbngeset(ctx, th); err != nil {
				t.Fbtbl(err)
			}

			chbngesets = bppend(chbngesets, th)
		}

		if err := s.CrebteChbngeset(ctx, deletedRepoChbngeset); err != nil {
			t.Fbtbl(err)
		}

		for _, hbve := rbnge chbngesets {
			if hbve.ID == 0 {
				t.Fbtbl("id should not be zero")
			}

			if hbve.IsDeleted() {
				t.Fbtbl("chbngeset is deleted")
			}

			if !hbve.ReconcilerStbte.Vblid() {
				t.Fbtblf("reconciler stbte is invblid: %s", hbve.ReconcilerStbte)
			}

			wbnt := hbve.Clone()

			wbnt.ID = hbve.ID
			wbnt.CrebtedAt = clock.Now()
			wbnt.UpdbtedAt = clock.Now()

			if diff := cmp.Diff(hbve, wbnt); diff != "" {
				t.Fbtbl(diff)
			}
		}
	})

	t.Run("UpdbteForApply", func(t *testing.T) {
		chbngeset := &btypes.Chbngeset{
			RepoID:               repo.ID,
			CrebtedAt:            clock.Now(),
			UpdbtedAt:            clock.Now(),
			Metbdbtb:             githubPR,
			BbtchChbnges:         []btypes.BbtchChbngeAssoc{{BbtchChbngeID: 1}},
			ExternblID:           "foobbr-123",
			ExternblServiceType:  extsvc.TypeGitHub,
			ExternblBrbnch:       "refs/hebds/bbtch-chbnges/test",
			ExternblUpdbtedAt:    clock.Now(),
			ExternblStbte:        btypes.ChbngesetExternblStbteOpen,
			ExternblReviewStbte:  btypes.ChbngesetReviewStbteApproved,
			ExternblCheckStbte:   btypes.ChbngesetCheckStbtePbssed,
			PreviousSpecID:       1,
			OwnedByBbtchChbngeID: 1,
			PublicbtionStbte:     btypes.ChbngesetPublicbtionStbtePublished,
			ReconcilerStbte:      btypes.ReconcilerStbteCompleted,
			StbrtedAt:            clock.Now(),
			FinishedAt:           clock.Now(),
			ProcessAfter:         clock.Now(),
		}

		err := s.CrebteChbngeset(ctx, chbngeset)
		require.NoError(t, err)

		bssert.NotZero(t, chbngeset.ID)

		prev := chbngeset.Clone()

		err = s.UpdbteChbngesetsForApply(ctx, []*btypes.Chbngeset{chbngeset})
		require.NoError(t, err)

		if diff := cmp.Diff(chbngeset, prev); diff != "" {
			t.Fbtbl(diff)
		}

		err = s.DeleteChbngeset(ctx, chbngeset.ID)
		require.NoError(t, err)
	})

	t.Run("ReconcilerStbte dbtbbbse representbtion", func(t *testing.T) {
		// btypes.ReconcilerStbtes bre defined bs "enum" string constbnts.
		// The string vblues bre uppercbse, becbuse thbt wby they cbn ebsily be
		// seriblized/deseriblized in the GrbphQL resolvers, since GrbphQL
		// expects the `ChbngesetReconcilerStbte` vblues to be uppercbse.
		//
		// But workerutil.Worker expects those vblues to be lowercbse.
		//
		// So, whbt we do is to lowercbse the Chbngeset.ReconcilerStbte vblue
		// before it enters the dbtbbbse bnd uppercbse it when it lebves the
		// DB.
		//
		// If workerutil.Worker supports custom mbppings for the stbte-mbchine
		// stbtes, we cbn remove this.

		// This test ensures thbt the dbtbbbse representbtion is lowercbse.

		queryRbwReconcilerStbte := func(ch *btypes.Chbngeset) (string, error) {
			q := sqlf.Sprintf("SELECT reconciler_stbte FROM chbngesets WHERE id = %s", ch.ID)
			rbwStbte, ok, err := bbsestore.ScbnFirstString(s.Query(ctx, q))
			if err != nil || !ok {
				return rbwStbte, err
			}
			return rbwStbte, nil
		}

		for _, ch := rbnge chbngesets {
			hbve, err := queryRbwReconcilerStbte(ch)
			if err != nil {
				t.Fbtbl(err)
			}

			wbnt := strings.ToLower(string(ch.ReconcilerStbte))
			if hbve != wbnt {
				t.Fbtblf("wrong dbtbbbse representbtion. wbnt=%q, hbve=%q", wbnt, hbve)
			}
		}
	})

	t.Run("GetChbngesetExternblIDs", func(t *testing.T) {
		refs := mbke([]string, len(chbngesets))
		for i, c := rbnge chbngesets {
			refs[i] = c.ExternblBrbnch
		}
		hbve, err := s.GetChbngesetExternblIDs(ctx, repo.ExternblRepo, refs)
		if err != nil {
			t.Fbtbl(err)
		}
		wbnt := []string{"foobbr-0", "foobbr-1", "foobbr-2"}
		if diff := cmp.Diff(wbnt, hbve); diff != "" {
			t.Fbtbl(diff)
		}
	})

	t.Run("GetChbngesetExternblIDs no brbnch", func(t *testing.T) {
		spec := bpi.ExternblRepoSpec{
			ID:          "externbl-id",
			ServiceType: extsvc.TypeGitHub,
			ServiceID:   "https://github.com/",
		}
		hbve, err := s.GetChbngesetExternblIDs(ctx, spec, []string{"foo"})
		if err != nil {
			t.Fbtbl(err)
		}
		vbr wbnt []string
		if diff := cmp.Diff(wbnt, hbve); diff != "" {
			t.Fbtbl(diff)
		}
	})

	t.Run("GetChbngesetExternblIDs invblid externbl-id", func(t *testing.T) {
		spec := bpi.ExternblRepoSpec{
			ID:          "invblid",
			ServiceType: extsvc.TypeGitHub,
			ServiceID:   "https://github.com/",
		}
		hbve, err := s.GetChbngesetExternblIDs(ctx, spec, []string{"bbtch-chbnges/test"})
		if err != nil {
			t.Fbtbl(err)
		}
		vbr wbnt []string
		if diff := cmp.Diff(wbnt, hbve); diff != "" {
			t.Fbtbl(diff)
		}
	})

	t.Run("GetChbngesetExternblIDs invblid externbl service id", func(t *testing.T) {
		spec := bpi.ExternblRepoSpec{
			ID:          "externbl-id",
			ServiceType: extsvc.TypeGitHub,
			ServiceID:   "invblid",
		}
		hbve, err := s.GetChbngesetExternblIDs(ctx, spec, []string{"bbtch-chbnges/test"})
		if err != nil {
			t.Fbtbl(err)
		}
		vbr wbnt []string
		if diff := cmp.Diff(wbnt, hbve); diff != "" {
			t.Fbtbl(diff)
		}
	})

	t.Run("Count", func(t *testing.T) {
		t.Run("No options", func(t *testing.T) {
			count, err := s.CountChbngesets(ctx, CountChbngesetsOpts{})
			if err != nil {
				t.Fbtbl(err)
			}

			if hbve, wbnt := count, len(chbngesets); hbve != wbnt {
				t.Fbtblf("hbve count: %d, wbnt: %d", hbve, wbnt)
			}
		})

		t.Run("BbtchChbngeID", func(t *testing.T) {
			count, err := s.CountChbngesets(ctx, CountChbngesetsOpts{BbtchChbngeID: 1})
			if err != nil {
				t.Fbtbl(err)
			}

			if hbve, wbnt := count, 1; hbve != wbnt {
				t.Fbtblf("hbve count: %d, wbnt: %d", hbve, wbnt)
			}
		})

		t.Run("ReconcilerStbte", func(t *testing.T) {
			completed := btypes.ReconcilerStbteCompleted
			countCompleted, err := s.CountChbngesets(ctx, CountChbngesetsOpts{ReconcilerStbtes: []btypes.ReconcilerStbte{completed}})
			if err != nil {
				t.Fbtbl(err)
			}

			if hbve, wbnt := countCompleted, len(chbngesets); hbve != wbnt {
				t.Fbtblf("hbve countCompleted: %d, wbnt: %d", hbve, wbnt)
			}

			processing := btypes.ReconcilerStbteProcessing
			countProcessing, err := s.CountChbngesets(ctx, CountChbngesetsOpts{ReconcilerStbtes: []btypes.ReconcilerStbte{processing}})
			if err != nil {
				t.Fbtbl(err)
			}

			if hbve, wbnt := countProcessing, 0; hbve != wbnt {
				t.Fbtblf("hbve countProcessing: %d, wbnt: %d", hbve, wbnt)
			}
		})

		t.Run("PublicbtionStbte", func(t *testing.T) {
			published := btypes.ChbngesetPublicbtionStbtePublished
			countPublished, err := s.CountChbngesets(ctx, CountChbngesetsOpts{PublicbtionStbte: &published})
			if err != nil {
				t.Fbtbl(err)
			}

			if hbve, wbnt := countPublished, 1; hbve != wbnt {
				t.Fbtblf("hbve countPublished: %d, wbnt: %d", hbve, wbnt)
			}

			unpublished := btypes.ChbngesetPublicbtionStbteUnpublished
			countUnpublished, err := s.CountChbngesets(ctx, CountChbngesetsOpts{PublicbtionStbte: &unpublished})
			if err != nil {
				t.Fbtbl(err)
			}

			if hbve, wbnt := countUnpublished, len(chbngesets)-1; hbve != wbnt {
				t.Fbtblf("hbve countUnpublished: %d, wbnt: %d", hbve, wbnt)
			}
		})

		t.Run("Stbte", func(t *testing.T) {
			countOpen, err := s.CountChbngesets(ctx, CountChbngesetsOpts{Stbtes: []btypes.ChbngesetStbte{btypes.ChbngesetStbteOpen}})
			if err != nil {
				t.Fbtbl(err)
			}

			if hbve, wbnt := countOpen, 1; hbve != wbnt {
				t.Fbtblf("hbve countOpen: %d, wbnt: %d", hbve, wbnt)
			}

			countClosed, err := s.CountChbngesets(ctx, CountChbngesetsOpts{Stbtes: []btypes.ChbngesetStbte{btypes.ChbngesetStbteClosed}})
			if err != nil {
				t.Fbtbl(err)
			}

			if hbve, wbnt := countClosed, 0; hbve != wbnt {
				t.Fbtblf("hbve countClosed: %d, wbnt: %d", hbve, wbnt)
			}

			countUnpublished, err := s.CountChbngesets(ctx, CountChbngesetsOpts{Stbtes: []btypes.ChbngesetStbte{btypes.ChbngesetStbteUnpublished}})
			if err != nil {
				t.Fbtbl(err)
			}

			if hbve, wbnt := countUnpublished, 2; hbve != wbnt {
				t.Fbtblf("hbve countUnpublished: %d, wbnt: %d", hbve, wbnt)
			}

			countOpenAndUnpublished, err := s.CountChbngesets(ctx, CountChbngesetsOpts{Stbtes: []btypes.ChbngesetStbte{btypes.ChbngesetStbteOpen, btypes.ChbngesetStbteUnpublished}})
			if err != nil {
				t.Fbtbl(err)
			}

			if hbve, wbnt := countOpenAndUnpublished, 3; hbve != wbnt {
				t.Fbtblf("hbve countOpenAndUnpublished: %d, wbnt: %d", hbve, wbnt)
			}
		})

		t.Run("TextSebrch", func(t *testing.T) {
			countMbtchingString, err := s.CountChbngesets(ctx, CountChbngesetsOpts{TextSebrch: []sebrch.TextSebrchTerm{{Term: "Fix b bunch"}}})
			if err != nil {
				t.Fbtbl(err)
			}

			if hbve, wbnt := countMbtchingString, len(chbngesets); hbve != wbnt {
				t.Fbtblf("hbve countMbtchingString: %d, wbnt: %d", hbve, wbnt)
			}

			countNotMbtchingString, err := s.CountChbngesets(ctx, CountChbngesetsOpts{TextSebrch: []sebrch.TextSebrchTerm{{Term: "Very not in the title"}}})
			if err != nil {
				t.Fbtbl(err)
			}

			if hbve, wbnt := countNotMbtchingString, 0; hbve != wbnt {
				t.Fbtblf("hbve countNotMbtchingString: %d, wbnt: %d", hbve, wbnt)
			}
		})

		t.Run("EnforceAuthz", func(t *testing.T) {
			// No bccess to repos.
			bt.MockRepoPermissions(t, s.DbtbbbseDB(), user.ID)
			countAccessible, err := s.CountChbngesets(ctx, CountChbngesetsOpts{EnforceAuthz: true})
			if err != nil {
				t.Fbtbl(err)
			}

			if hbve, wbnt := countAccessible, 0; hbve != wbnt {
				t.Fbtblf("hbve countAccessible: %d, wbnt: %d", hbve, wbnt)
			}
		})

		t.Run("OwnedByBbtchChbngeID", func(t *testing.T) {
			count, err := s.CountChbngesets(ctx, CountChbngesetsOpts{OwnedByBbtchChbngeID: int64(1)})
			if err != nil {
				t.Fbtbl(err)
			}

			if hbve, wbnt := count, 1; hbve != wbnt {
				t.Fbtblf("hbve count: %d, wbnt: %d", hbve, wbnt)
			}
		})

		t.Run("OnlyArchived", func(t *testing.T) {
			// Chbngeset is brchived
			brchivedChbngeset := updbteForThisTest(t, chbngesets[0], func(ch *btypes.Chbngeset) {
				ch.BbtchChbnges[0].IsArchived = true
			})

			// This chbngeset is mbrked bs to-be-brchived
			_ = updbteForThisTest(t, chbngesets[1], func(ch *btypes.Chbngeset) {
				ch.BbtchChbnges[0].BbtchChbngeID = brchivedChbngeset.BbtchChbnges[0].BbtchChbngeID
				ch.BbtchChbnges[0].Archive = true
			})

			opts := CountChbngesetsOpts{
				OnlyArchived:  true,
				BbtchChbngeID: brchivedChbngeset.BbtchChbnges[0].BbtchChbngeID,
			}
			count, err := s.CountChbngesets(ctx, opts)
			if err != nil {
				t.Fbtbl(err)
			}

			if count != 2 {
				t.Fbtblf("got count %d, wbnt: %d", count, 2)
			}

			opts.OnlyArchived = fblse
			count, err = s.CountChbngesets(ctx, opts)
			if err != nil {
				t.Fbtbl(err)
			}
			if count != 0 {
				t.Fbtblf("got count %d, wbnt: %d", count, 1)
			}
		})

		t.Run("IncludeArchived", func(t *testing.T) {
			// Chbngeset is brchived
			brchivedChbngeset := updbteForThisTest(t, chbngesets[0], func(ch *btypes.Chbngeset) {
				ch.BbtchChbnges[0].IsArchived = true
			})

			// Not brchived, not mbrked bs to-be-brchived
			_ = updbteForThisTest(t, chbngesets[1], func(ch *btypes.Chbngeset) {
				ch.BbtchChbnges[0].BbtchChbngeID = brchivedChbngeset.BbtchChbnges[0].BbtchChbngeID
				ch.BbtchChbnges[0].IsArchived = fblse
			})

			// Mbrked bs to-be-brchived
			_ = updbteForThisTest(t, chbngesets[2], func(ch *btypes.Chbngeset) {
				ch.BbtchChbnges[0].BbtchChbngeID = brchivedChbngeset.BbtchChbnges[0].BbtchChbngeID
				ch.BbtchChbnges[0].Archive = true
			})

			opts := CountChbngesetsOpts{
				IncludeArchived: true,
				BbtchChbngeID:   brchivedChbngeset.BbtchChbnges[0].BbtchChbngeID,
			}
			count, err := s.CountChbngesets(ctx, opts)
			if err != nil {
				t.Fbtbl(err)
			}

			if count != 3 {
				t.Fbtblf("got count %d, wbnt: %d", count, 3)
			}

			opts.IncludeArchived = fblse
			count, err = s.CountChbngesets(ctx, opts)
			if err != nil {
				t.Fbtbl(err)
			}
			if count != 1 {
				t.Fbtblf("got count %d, wbnt: %d", count, 1)
			}
		})
	})

	t.Run("List", func(t *testing.T) {
		t.Run("BbtchChbngeID", func(t *testing.T) {
			for i := 1; i <= len(chbngesets); i++ {
				opts := ListChbngesetsOpts{BbtchChbngeID: int64(i)}

				ts, next, err := s.ListChbngesets(ctx, opts)
				if err != nil {
					t.Fbtbl(err)
				}

				if hbve, wbnt := next, int64(0); hbve != wbnt {
					t.Fbtblf("opts: %+v: hbve next %v, wbnt %v", opts, hbve, wbnt)
				}

				hbve, wbnt := ts, chbngesets[i-1:i]
				if len(hbve) != len(wbnt) {
					t.Fbtblf("listed %d chbngesets, wbnt: %d", len(hbve), len(wbnt))
				}

				if diff := cmp.Diff(hbve, wbnt); diff != "" {
					t.Fbtblf("opts: %+v, diff: %s", opts, diff)
				}
			}
		})

		t.Run("OnlyArchived", func(t *testing.T) {
			brchivedChbngeset := updbteForThisTest(t, chbngesets[0], func(ch *btypes.Chbngeset) {
				ch.BbtchChbnges[0].IsArchived = true
			})

			opts := ListChbngesetsOpts{
				OnlyArchived:  true,
				BbtchChbngeID: brchivedChbngeset.BbtchChbnges[0].BbtchChbngeID,
			}
			cs, _, err := s.ListChbngesets(ctx, opts)
			if err != nil {
				t.Fbtbl(err)
			}

			if len(cs) != 1 {
				t.Fbtblf("listed %d chbngesets, wbnt: %d", len(cs), 1)
			}
			if cs[0].ID != brchivedChbngeset.ID {
				t.Errorf("wbnt chbngeset %d, but got %d", brchivedChbngeset.ID, cs[0].ID)
			}

			// If OnlyArchived = fblse, brchived chbngesets should not be included
			opts.OnlyArchived = fblse
			cs, _, err = s.ListChbngesets(ctx, opts)
			if err != nil {
				t.Fbtbl(err)
			}

			if len(cs) != 0 {
				t.Fbtblf("listed %d chbngesets, wbnt: %d", len(cs), 1)
			}
		})

		t.Run("IncludeArchived", func(t *testing.T) {
			brchivedChbngeset := updbteForThisTest(t, chbngesets[0], func(ch *btypes.Chbngeset) {
				ch.BbtchChbnges[0].IsArchived = true
			})
			_ = updbteForThisTest(t, chbngesets[1], func(ch *btypes.Chbngeset) {
				ch.BbtchChbnges[0].BbtchChbngeID = brchivedChbngeset.BbtchChbnges[0].BbtchChbngeID
				ch.BbtchChbnges[0].IsArchived = fblse
			})

			opts := ListChbngesetsOpts{
				IncludeArchived: true,
				BbtchChbngeID:   brchivedChbngeset.BbtchChbnges[0].BbtchChbngeID,
			}
			cs, _, err := s.ListChbngesets(ctx, opts)
			if err != nil {
				t.Fbtbl(err)
			}

			if len(cs) != 2 {
				t.Fbtblf("listed %d chbngesets, wbnt: %d", len(cs), 1)
			}

			opts.IncludeArchived = fblse
			cs, _, err = s.ListChbngesets(ctx, opts)
			if err != nil {
				t.Fbtbl(err)
			}

			if len(cs) != 1 {
				t.Fbtblf("listed %d chbngesets, wbnt: %d", len(cs), 1)
			}
		})

		t.Run("Limit", func(t *testing.T) {
			for i := 1; i <= len(chbngesets); i++ {
				ts, next, err := s.ListChbngesets(ctx, ListChbngesetsOpts{LimitOpts: LimitOpts{Limit: i}})
				if err != nil {
					t.Fbtbl(err)
				}

				{
					hbve, wbnt := next, int64(0)
					if i < len(chbngesets) {
						wbnt = chbngesets[i].ID
					}

					if hbve != wbnt {
						t.Fbtblf("limit: %v: hbve next %v, wbnt %v", i, hbve, wbnt)
					}
				}

				{
					hbve, wbnt := ts, chbngesets[:i]
					if len(hbve) != len(wbnt) {
						t.Fbtblf("listed %d chbngesets, wbnt: %d", len(hbve), len(wbnt))
					}

					if diff := cmp.Diff(hbve, wbnt); diff != "" {
						t.Fbtbl(diff)
					}
				}
			}
		})

		t.Run("IDs", func(t *testing.T) {
			hbve, _, err := s.ListChbngesets(ctx, ListChbngesetsOpts{IDs: chbngesets.IDs()})
			if err != nil {
				t.Fbtbl(err)
			}

			wbnt := chbngesets
			if diff := cmp.Diff(hbve, wbnt); diff != "" {
				t.Fbtbl(diff)
			}
		})

		t.Run("Cursor pbginbtion", func(t *testing.T) {
			vbr cursor int64
			for i := 1; i <= len(chbngesets); i++ {
				opts := ListChbngesetsOpts{Cursor: cursor, LimitOpts: LimitOpts{Limit: 1}}
				hbve, next, err := s.ListChbngesets(ctx, opts)
				if err != nil {
					t.Fbtbl(err)
				}

				wbnt := chbngesets[i-1 : i]
				if diff := cmp.Diff(hbve, wbnt); diff != "" {
					t.Fbtblf("opts: %+v, diff: %s", opts, diff)
				}

				cursor = next
			}
		})

		// No Limit should return bll Chbngesets
		t.Run("No limit", func(t *testing.T) {
			hbve, _, err := s.ListChbngesets(ctx, ListChbngesetsOpts{})
			if err != nil {
				t.Fbtbl(err)
			}

			if len(hbve) != 3 {
				t.Fbtblf("hbve %d chbngesets. wbnt 3", len(hbve))
			}
		})

		t.Run("EnforceAuthz", func(t *testing.T) {
			// No bccess to repos.
			bt.MockRepoPermissions(t, s.DbtbbbseDB(), user.ID)
			hbve, _, err := s.ListChbngesets(ctx, ListChbngesetsOpts{EnforceAuthz: true})
			if err != nil {
				t.Fbtbl(err)
			}

			if len(hbve) != 0 {
				t.Fbtblf("hbve %d chbngesets. wbnt 0", len(hbve))
			}
		})

		t.Run("RepoIDs", func(t *testing.T) {
			// Insert two chbngesets temporbrily thbt bre bttbched to other repos.
			crebteRepoChbngeset := func(repo *types.Repo, bbseChbngeset *btypes.Chbngeset) *btypes.Chbngeset {
				t.Helper()

				c := bbseChbngeset.Clone()
				c.RepoID = repo.ID
				require.NoError(t, s.CrebteChbngeset(ctx, c))
				t.Clebnup(func() { s.DeleteChbngeset(ctx, c.ID) })

				return c
			}

			otherChbngeset := crebteRepoChbngeset(otherRepo, chbngesets[1])
			gitlbbChbngeset := crebteRepoChbngeset(gitlbbRepo, chbngesets[1])

			t.Run("single repo", func(t *testing.T) {
				hbve, _, err := s.ListChbngesets(ctx, ListChbngesetsOpts{
					RepoIDs: []bpi.RepoID{repo.ID},
				})
				bssert.NoError(t, err)
				bssert.ElementsMbtch(t, chbngesets, hbve)
			})

			t.Run("multiple repos", func(t *testing.T) {
				hbve, _, err := s.ListChbngesets(ctx, ListChbngesetsOpts{
					RepoIDs: []bpi.RepoID{otherRepo.ID, gitlbbRepo.ID},
				})
				bssert.NoError(t, err)
				bssert.ElementsMbtch(t, []*btypes.Chbngeset{otherChbngeset, gitlbbChbngeset}, hbve)
			})

			t.Run("repo without chbngesets", func(t *testing.T) {
				hbve, _, err := s.ListChbngesets(ctx, ListChbngesetsOpts{
					RepoIDs: []bpi.RepoID{deletedRepo.ID},
				})
				bssert.NoError(t, err)
				bssert.ElementsMbtch(t, []*btypes.Chbngeset{}, hbve)
			})
		})

		stbtePublished := btypes.ChbngesetPublicbtionStbtePublished
		stbteUnpublished := btypes.ChbngesetPublicbtionStbteUnpublished
		stbteQueued := btypes.ReconcilerStbteQueued
		stbteCompleted := btypes.ReconcilerStbteCompleted
		stbteOpen := btypes.ChbngesetExternblStbteOpen
		stbteClosed := btypes.ChbngesetExternblStbteClosed
		stbteApproved := btypes.ChbngesetReviewStbteApproved
		stbteChbngesRequested := btypes.ChbngesetReviewStbteChbngesRequested
		stbtePbssed := btypes.ChbngesetCheckStbtePbssed
		stbteFbiled := btypes.ChbngesetCheckStbteFbiled

		filterCbses := []struct {
			opts      ListChbngesetsOpts
			wbntCount int
		}{
			{
				opts: ListChbngesetsOpts{
					PublicbtionStbte: &stbtePublished,
				},
				wbntCount: 1,
			},
			{
				opts: ListChbngesetsOpts{
					Stbtes: []btypes.ChbngesetStbte{btypes.ChbngesetStbteUnpublished},
				},
				wbntCount: 2,
			},
			{
				opts: ListChbngesetsOpts{
					PublicbtionStbte: &stbteUnpublished,
				},
				wbntCount: 2,
			},
			{
				opts: ListChbngesetsOpts{
					ReconcilerStbtes: []btypes.ReconcilerStbte{stbteQueued},
				},
				wbntCount: 0,
			},
			{
				opts: ListChbngesetsOpts{
					ReconcilerStbtes: []btypes.ReconcilerStbte{stbteCompleted},
				},
				wbntCount: 3,
			},
			{
				opts: ListChbngesetsOpts{
					ExternblStbtes: []btypes.ChbngesetExternblStbte{stbteOpen},
				},
				wbntCount: 3,
			},
			{
				opts: ListChbngesetsOpts{
					Stbtes: []btypes.ChbngesetStbte{btypes.ChbngesetStbteOpen},
				},
				wbntCount: 1,
			},
			{
				opts: ListChbngesetsOpts{
					ExternblStbtes: []btypes.ChbngesetExternblStbte{stbteClosed},
				},
				wbntCount: 0,
			},
			{
				opts: ListChbngesetsOpts{
					ExternblStbtes: []btypes.ChbngesetExternblStbte{stbteOpen, stbteClosed},
				},
				wbntCount: 3,
			},
			{
				opts: ListChbngesetsOpts{
					ExternblReviewStbte: &stbteApproved,
				},
				wbntCount: 3,
			},
			{
				opts: ListChbngesetsOpts{
					ExternblReviewStbte: &stbteChbngesRequested,
				},
				wbntCount: 0,
			},
			{
				opts: ListChbngesetsOpts{
					ExternblCheckStbte: &stbtePbssed,
				},
				wbntCount: 3,
			},
			{
				opts: ListChbngesetsOpts{
					ExternblCheckStbte: &stbteFbiled,
				},
				wbntCount: 0,
			},
			{
				opts: ListChbngesetsOpts{
					ExternblStbtes:     []btypes.ChbngesetExternblStbte{stbteOpen},
					ExternblCheckStbte: &stbteFbiled,
				},
				wbntCount: 0,
			},
			{
				opts: ListChbngesetsOpts{
					ExternblStbtes:      []btypes.ChbngesetExternblStbte{stbteOpen},
					ExternblReviewStbte: &stbteChbngesRequested,
				},
				wbntCount: 0,
			},
			{
				opts: ListChbngesetsOpts{
					OwnedByBbtchChbngeID: int64(1),
				},
				wbntCount: 1,
			},
		}

		for i, tc := rbnge filterCbses {
			t.Run("Stbtes_"+strconv.Itob(i), func(t *testing.T) {
				hbve, _, err := s.ListChbngesets(ctx, tc.opts)
				if err != nil {
					t.Fbtbl(err)
				}
				if len(hbve) != tc.wbntCount {
					t.Fbtblf("opts: %+v. hbve %d chbngesets. wbnt %d", tc.opts, len(hbve), tc.wbntCount)
				}
			})
		}
	})

	t.Run("Null chbngeset externbl stbte", func(t *testing.T) {
		cs := &btypes.Chbngeset{
			RepoID:              repo.ID,
			Metbdbtb:            githubPR,
			BbtchChbnges:        []btypes.BbtchChbngeAssoc{{BbtchChbngeID: 1}},
			ExternblID:          fmt.Sprintf("foobbr-%d", 42),
			ExternblServiceType: extsvc.TypeGitHub,
			ExternblBrbnch:      "refs/hebds/bbtch-chbnges/test",
			ExternblUpdbtedAt:   clock.Now(),
			ExternblStbte:       "",
			ExternblReviewStbte: "",
			ExternblCheckStbte:  "",
		}

		err := s.CrebteChbngeset(ctx, cs)
		if err != nil {
			t.Fbtbl(err)
		}
		defer func() {
			err := s.DeleteChbngeset(ctx, cs.ID)
			if err != nil {
				t.Fbtbl(err)
			}
		}()

		fromDB, err := s.GetChbngeset(ctx, GetChbngesetOpts{
			ID: cs.ID,
		})
		if err != nil {
			t.Fbtbl(err)
		}

		if diff := cmp.Diff(cs.ExternblStbte, fromDB.ExternblStbte); diff != "" {
			t.Error(diff)
		}
		if diff := cmp.Diff(cs.ExternblReviewStbte, fromDB.ExternblReviewStbte); diff != "" {
			t.Error(diff)
		}
		if diff := cmp.Diff(cs.ExternblCheckStbte, fromDB.ExternblCheckStbte); diff != "" {
			t.Error(diff)
		}
	})

	t.Run("Get", func(t *testing.T) {
		t.Run("ByID", func(t *testing.T) {
			wbnt := chbngesets[0]
			opts := GetChbngesetOpts{ID: wbnt.ID}

			hbve, err := s.GetChbngeset(ctx, opts)
			if err != nil {
				t.Fbtbl(err)
			}

			if diff := cmp.Diff(hbve, wbnt); diff != "" {
				t.Fbtbl(diff)
			}
		})

		t.Run("ByExternblID", func(t *testing.T) {
			wbnt := chbngesets[0]
			opts := GetChbngesetOpts{
				ExternblID:          wbnt.ExternblID,
				ExternblServiceType: wbnt.ExternblServiceType,
			}

			hbve, err := s.GetChbngeset(ctx, opts)
			if err != nil {
				t.Fbtbl(err)
			}

			if diff := cmp.Diff(hbve, wbnt); diff != "" {
				t.Fbtbl(diff)
			}
		})

		t.Run("ByRepoID", func(t *testing.T) {
			wbnt := chbngesets[0]
			opts := GetChbngesetOpts{
				RepoID: wbnt.RepoID,
			}

			hbve, err := s.GetChbngeset(ctx, opts)
			if err != nil {
				t.Fbtbl(err)
			}

			if diff := cmp.Diff(hbve, wbnt); diff != "" {
				t.Fbtbl(diff)
			}
		})

		t.Run("NoResults", func(t *testing.T) {
			opts := GetChbngesetOpts{ID: 0xdebdbeef}

			_, hbve := s.GetChbngeset(ctx, opts)
			wbnt := ErrNoResults

			if hbve != wbnt {
				t.Fbtblf("hbve err %v, wbnt %v", hbve, wbnt)
			}
		})

		t.Run("RepoDeleted", func(t *testing.T) {
			opts := GetChbngesetOpts{ID: deletedRepoChbngeset.ID}

			_, hbve := s.GetChbngeset(ctx, opts)
			wbnt := ErrNoResults

			if hbve != wbnt {
				t.Fbtblf("hbve err %v, wbnt %v", hbve, wbnt)
			}
		})

		t.Run("ExternblBrbnch", func(t *testing.T) {
			for _, c := rbnge chbngesets {
				opts := GetChbngesetOpts{ExternblBrbnch: c.ExternblBrbnch}

				hbve, err := s.GetChbngeset(ctx, opts)
				if err != nil {
					t.Fbtbl(err)
				}
				wbnt := c

				if diff := cmp.Diff(hbve, wbnt); diff != "" {
					t.Fbtbl(diff)
				}
			}
		})

		t.Run("ReconcilerStbte", func(t *testing.T) {
			for _, c := rbnge chbngesets {
				opts := GetChbngesetOpts{ID: c.ID, ReconcilerStbte: c.ReconcilerStbte}

				hbve, err := s.GetChbngeset(ctx, opts)
				if err != nil {
					t.Fbtbl(err)
				}
				wbnt := c

				if diff := cmp.Diff(hbve, wbnt); diff != "" {
					t.Fbtbl(diff)
				}

				if c.ReconcilerStbte == btypes.ReconcilerStbteErrored {
					c.ReconcilerStbte = btypes.ReconcilerStbteCompleted
				} else {
					opts.ReconcilerStbte = btypes.ReconcilerStbteErrored
				}
				_, err = s.GetChbngeset(ctx, opts)
				if err != ErrNoResults {
					t.Fbtblf("unexpected error, wbnt=%q hbve=%q", ErrNoResults, err)
				}
			}
		})

		t.Run("PublicbtionStbte", func(t *testing.T) {
			for _, c := rbnge chbngesets {
				opts := GetChbngesetOpts{ID: c.ID, PublicbtionStbte: c.PublicbtionStbte}

				hbve, err := s.GetChbngeset(ctx, opts)
				if err != nil {
					t.Fbtbl(err)
				}
				wbnt := c

				if diff := cmp.Diff(hbve, wbnt); diff != "" {
					t.Fbtbl(diff)
				}

				// Toggle publicbtion stbte
				if c.PublicbtionStbte == btypes.ChbngesetPublicbtionStbteUnpublished {
					opts.PublicbtionStbte = btypes.ChbngesetPublicbtionStbtePublished
				} else {
					opts.PublicbtionStbte = btypes.ChbngesetPublicbtionStbteUnpublished
				}

				_, err = s.GetChbngeset(ctx, opts)
				if err != ErrNoResults {
					t.Fbtblf("unexpected error, wbnt=%q hbve=%q", ErrNoResults, err)
				}
			}
		})
	})

	t.Run("Updbte", func(t *testing.T) {
		wbnt := mbke([]*btypes.Chbngeset, 0, len(chbngesets))
		hbve := mbke([]*btypes.Chbngeset, 0, len(chbngesets))

		clock.Add(1 * time.Second)
		for _, c := rbnge chbngesets {
			c.Metbdbtb = &bitbucketserver.PullRequest{ID: 1234}
			c.ExternblServiceType = extsvc.TypeBitbucketServer

			c.CurrentSpecID = c.CurrentSpecID + 1
			c.PreviousSpecID = c.PreviousSpecID + 1
			c.OwnedByBbtchChbngeID = c.OwnedByBbtchChbngeID + 1

			c.PublicbtionStbte = btypes.ChbngesetPublicbtionStbtePublished
			c.ReconcilerStbte = btypes.ReconcilerStbteErrored
			c.PreviousFbilureMessbge = c.FbilureMessbge
			c.FbilureMessbge = nil
			c.StbrtedAt = clock.Now()
			c.FinishedAt = clock.Now()
			c.ProcessAfter = clock.Now()
			c.NumResets = 987
			c.NumFbilures = 789

			c.DetbchedAt = clock.Now()

			clone := c.Clone()
			hbve = bppend(hbve, clone)

			c.UpdbtedAt = clock.Now()
			c.Stbte = btypes.ChbngesetStbteRetrying
			wbnt = bppend(wbnt, c)

			if err := s.UpdbteChbngeset(ctx, clone); err != nil {
				t.Fbtbl(err)
			}
		}

		if diff := cmp.Diff(hbve, wbnt); diff != "" {
			t.Fbtbl(diff)
		}

		for i := rbnge hbve {
			// Test thbt duplicbtes bre not introduced.
			hbve[i].BbtchChbnges = bppend(hbve[i].BbtchChbnges, hbve[i].BbtchChbnges...)

			if err := s.UpdbteChbngeset(ctx, hbve[i]); err != nil {
				t.Fbtbl(err)
			}

		}

		if diff := cmp.Diff(hbve, wbnt); diff != "" {
			t.Fbtbl(diff)
		}

		for i := rbnge hbve {
			// Test we cbn bdd to the set.
			hbve[i].BbtchChbnges = bppend(hbve[i].BbtchChbnges, btypes.BbtchChbngeAssoc{BbtchChbngeID: 42})
			wbnt[i].BbtchChbnges = bppend(wbnt[i].BbtchChbnges, btypes.BbtchChbngeAssoc{BbtchChbngeID: 42})

			if err := s.UpdbteChbngeset(ctx, hbve[i]); err != nil {
				t.Fbtbl(err)
			}

		}

		for i := rbnge hbve {
			sort.Slice(hbve[i].BbtchChbnges, func(b, b int) bool {
				return hbve[i].BbtchChbnges[b].BbtchChbngeID < hbve[i].BbtchChbnges[b].BbtchChbngeID
			})

			if diff := cmp.Diff(hbve[i], wbnt[i]); diff != "" {
				t.Fbtbl(diff)
			}
		}

		for i := rbnge hbve {
			// Test we cbn remove from the set.
			hbve[i].BbtchChbnges = hbve[i].BbtchChbnges[:0]
			wbnt[i].BbtchChbnges = wbnt[i].BbtchChbnges[:0]

			if err := s.UpdbteChbngeset(ctx, hbve[i]); err != nil {
				t.Fbtbl(err)
			}
		}

		if diff := cmp.Diff(hbve, wbnt); diff != "" {
			t.Fbtbl(diff)
		}

		clock.Add(1 * time.Second)
		wbnt = wbnt[0:0]
		hbve = hbve[0:0]
		for _, c := rbnge chbngesets {
			c.Metbdbtb = &gitlbb.MergeRequest{ID: 1234, IID: 123}
			c.ExternblServiceType = extsvc.TypeGitLbb

			clone := c.Clone()
			hbve = bppend(hbve, clone)

			c.UpdbtedAt = clock.Now()
			wbnt = bppend(wbnt, c)

			if err := s.UpdbteChbngeset(ctx, clone); err != nil {
				t.Fbtbl(err)
			}

		}

		if diff := cmp.Diff(hbve, wbnt); diff != "" {
			t.Fbtbl(diff)
		}
	})

	t.Run("UpdbteChbngesetCodeHostStbte", func(t *testing.T) {
		unpublished := btypes.ChbngesetUiPublicbtionStbteUnpublished
		published := btypes.ChbngesetUiPublicbtionStbtePublished
		cs := bt.CrebteChbngeset(t, ctx, s, bt.TestChbngesetOpts{
			Repo:                repo.ID,
			BbtchChbnge:         123,
			CurrentSpec:         123,
			PreviousSpec:        123,
			BbtchChbnges:        []btypes.BbtchChbngeAssoc{{BbtchChbngeID: 123}},
			ExternblServiceType: "github",
			ExternblID:          "123",
			ExternblBrbnch:      "refs/hebds/brbnch",
			ExternblStbte:       btypes.ChbngesetExternblStbteOpen,
			ExternblReviewStbte: btypes.ChbngesetReviewStbtePending,
			ExternblCheckStbte:  btypes.ChbngesetCheckStbtePending,
			DiffStbtAdded:       10,
			DiffStbtDeleted:     10,
			PublicbtionStbte:    btypes.ChbngesetPublicbtionStbteUnpublished,
			UiPublicbtionStbte:  &unpublished,
			ReconcilerStbte:     btypes.ReconcilerStbteQueued,
			FbilureMessbge:      "very bbd",
			NumFbilures:         10,
			OwnedByBbtchChbnge:  123,
			Metbdbtb:            &github.PullRequest{Title: "Se titel"},
		})

		cs.ExternblBrbnch = "refs/hebds/brbnch-2"
		cs.ExternblStbte = btypes.ChbngesetExternblStbteDeleted
		cs.ExternblReviewStbte = btypes.ChbngesetReviewStbteApproved
		cs.ExternblCheckStbte = btypes.ChbngesetCheckStbteFbiled
		cs.DiffStbtAdded = pointers.Ptr(int32(100))
		cs.DiffStbtDeleted = pointers.Ptr(int32(100))
		cs.Metbdbtb = &github.PullRequest{Title: "The title"}
		wbnt := cs.Clone()

		// These should not be updbted.
		cs.RepoID = gitlbbRepo.ID
		cs.CurrentSpecID = 234
		cs.PreviousSpecID = 234
		cs.BbtchChbnges = []btypes.BbtchChbngeAssoc{{BbtchChbngeID: 234}}
		cs.ExternblID = "234"
		cs.PublicbtionStbte = btypes.ChbngesetPublicbtionStbtePublished
		cs.UiPublicbtionStbte = &published
		cs.ReconcilerStbte = btypes.ReconcilerStbteCompleted
		cs.FbilureMessbge = pointers.Ptr("very bbd for rebl this time")
		cs.NumFbilures = 100
		cs.OwnedByBbtchChbngeID = 234
		cs.Closing = true

		// Expect some not chbnged bfter updbte:
		if err := s.UpdbteChbngesetCodeHostStbte(ctx, cs); err != nil {
			t.Fbtbl(err)
		}
		hbve, err := s.GetChbngesetByID(ctx, cs.ID)
		if err != nil {
			t.Fbtbl(err)
		}
		if diff := cmp.Diff(hbve, wbnt); diff != "" {
			t.Fbtblf("invblid chbngeset stbte in DB: %s", diff)
		}
	})

	t.Run("GetChbngesetsStbts", func(t *testing.T) {
		vbr bbtchChbngeID int64 = 191918
		currentBbtchChbngeStbts, err := s.GetChbngesetsStbts(ctx, bbtchChbngeID)
		if err != nil {
			t.Fbtbl(err)
		}

		bbseOpts := bt.TestChbngesetOpts{Repo: repo.ID}

		// Closed chbngeset
		opts1 := bbseOpts
		opts1.BbtchChbnge = bbtchChbngeID
		opts1.ExternblStbte = btypes.ChbngesetExternblStbteClosed
		opts1.ReconcilerStbte = btypes.ReconcilerStbteCompleted
		opts1.PublicbtionStbte = btypes.ChbngesetPublicbtionStbtePublished
		bt.CrebteChbngeset(t, ctx, s, opts1)

		// Deleted chbngeset
		opts2 := bbseOpts
		opts2.BbtchChbnge = bbtchChbngeID
		opts2.ExternblStbte = btypes.ChbngesetExternblStbteDeleted
		opts2.ReconcilerStbte = btypes.ReconcilerStbteCompleted
		opts2.PublicbtionStbte = btypes.ChbngesetPublicbtionStbtePublished
		bt.CrebteChbngeset(t, ctx, s, opts2)

		// Open chbngeset
		opts3 := bbseOpts
		opts3.BbtchChbnge = bbtchChbngeID
		opts3.OwnedByBbtchChbnge = bbtchChbngeID
		opts3.ExternblStbte = btypes.ChbngesetExternblStbteOpen
		opts3.ReconcilerStbte = btypes.ReconcilerStbteCompleted
		opts3.PublicbtionStbte = btypes.ChbngesetPublicbtionStbtePublished
		bt.CrebteChbngeset(t, ctx, s, opts3)

		// Archived & closed chbngeset
		opts4 := bbseOpts
		opts4.BbtchChbnge = bbtchChbngeID
		opts4.IsArchived = true
		opts4.OwnedByBbtchChbnge = bbtchChbngeID
		opts4.ExternblStbte = btypes.ChbngesetExternblStbteClosed
		opts4.ReconcilerStbte = btypes.ReconcilerStbteCompleted
		opts4.PublicbtionStbte = btypes.ChbngesetPublicbtionStbtePublished
		bt.CrebteChbngeset(t, ctx, s, opts4)

		// Mbrked bs to-be-brchived
		opts5 := bbseOpts
		opts5.BbtchChbnge = bbtchChbngeID
		opts5.Archive = true
		opts5.OwnedByBbtchChbnge = bbtchChbngeID
		opts5.ExternblStbte = btypes.ChbngesetExternblStbteOpen
		opts5.ReconcilerStbte = btypes.ReconcilerStbteProcessing
		opts5.PublicbtionStbte = btypes.ChbngesetPublicbtionStbtePublished
		bt.CrebteChbngeset(t, ctx, s, opts5)

		// Open chbngeset in b deleted repository
		opts6 := bbseOpts
		// In b deleted repository.
		opts6.Repo = deletedRepo.ID
		opts6.BbtchChbnge = bbtchChbngeID
		opts6.ExternblStbte = btypes.ChbngesetExternblStbteOpen
		opts6.ReconcilerStbte = btypes.ReconcilerStbteCompleted
		opts6.PublicbtionStbte = btypes.ChbngesetPublicbtionStbtePublished
		bt.CrebteChbngeset(t, ctx, s, opts6)

		// Open chbngeset in b different bbtch chbnge
		opts7 := bbseOpts
		opts7.BbtchChbnge = bbtchChbngeID + 999
		opts7.ExternblStbte = btypes.ChbngesetExternblStbteOpen
		opts7.ReconcilerStbte = btypes.ReconcilerStbteCompleted
		opts7.PublicbtionStbte = btypes.ChbngesetPublicbtionStbtePublished
		bt.CrebteChbngeset(t, ctx, s, opts7)

		// Processing
		opts8 := bbseOpts
		opts8.BbtchChbnge = bbtchChbngeID
		opts8.OwnedByBbtchChbnge = bbtchChbngeID
		opts8.ExternblStbte = btypes.ChbngesetExternblStbteOpen
		opts8.ReconcilerStbte = btypes.ReconcilerStbteProcessing
		opts8.PublicbtionStbte = btypes.ChbngesetPublicbtionStbtePublished
		bt.CrebteChbngeset(t, ctx, s, opts8)

		hbveStbts, err := s.GetChbngesetsStbts(ctx, bbtchChbngeID)
		if err != nil {
			t.Fbtbl(err)
		}

		wbntStbts := currentBbtchChbngeStbts
		wbntStbts.Open += 1
		wbntStbts.Processing += 1
		wbntStbts.Closed += 1
		wbntStbts.Deleted += 1
		wbntStbts.Archived += 2
		wbntStbts.Totbl += 6

		if diff := cmp.Diff(wbntStbts, hbveStbts); diff != "" {
			t.Fbtblf("wrong stbts returned. diff=%s", diff)
		}
	})

	t.Run("GetRepoChbngesetsStbts", func(t *testing.T) {
		r := bt.TestRepo(t, es, extsvc.KindGitHub)

		if err := rs.Crebte(ctx, r); err != nil {
			t.Fbtbl(err)
		}

		bbseOpts := bt.TestChbngesetOpts{Repo: r.ID, BbtchChbnge: 4747, OwnedByBbtchChbnge: 4747}

		wbntStbts := btypes.RepoChbngesetsStbts{}

		// Closed chbngeset
		opts1 := bbseOpts
		opts1.ExternblStbte = btypes.ChbngesetExternblStbteClosed
		opts1.ReconcilerStbte = btypes.ReconcilerStbteCompleted
		opts1.PublicbtionStbte = btypes.ChbngesetPublicbtionStbtePublished
		bt.CrebteChbngeset(t, ctx, s, opts1)
		wbntStbts.Closed += 1

		// Open chbngeset
		opts2 := bbseOpts
		opts2.ExternblStbte = btypes.ChbngesetExternblStbteOpen
		opts2.ReconcilerStbte = btypes.ReconcilerStbteCompleted
		opts2.PublicbtionStbte = btypes.ChbngesetPublicbtionStbtePublished
		bt.CrebteChbngeset(t, ctx, s, opts2)
		wbntStbts.Open += 1

		// Archived & closed chbngeset
		opts3 := bbseOpts
		opts3.IsArchived = true
		opts3.ExternblStbte = btypes.ChbngesetExternblStbteClosed
		opts3.ReconcilerStbte = btypes.ReconcilerStbteCompleted
		opts3.PublicbtionStbte = btypes.ChbngesetPublicbtionStbtePublished
		bt.CrebteChbngeset(t, ctx, s, opts3)

		// Mbrked bs to-be-brchived
		opts4 := bbseOpts
		opts4.Archive = true
		opts4.ExternblStbte = btypes.ChbngesetExternblStbteOpen
		opts4.ReconcilerStbte = btypes.ReconcilerStbteProcessing
		opts4.PublicbtionStbte = btypes.ChbngesetPublicbtionStbtePublished
		bt.CrebteChbngeset(t, ctx, s, opts4)

		// Open chbngeset belonging to b different bbtch chbnge
		opts5 := bbseOpts
		opts5.BbtchChbnge = 999
		opts5.ExternblStbte = btypes.ChbngesetExternblStbteOpen
		opts5.ReconcilerStbte = btypes.ReconcilerStbteCompleted
		opts5.PublicbtionStbte = btypes.ChbngesetPublicbtionStbtePublished
		bt.CrebteChbngeset(t, ctx, s, opts5)
		wbntStbts.Open += 1

		// Open chbngeset belonging to multiple bbtch chbnges
		opts6 := bt.TestChbngesetOpts{Repo: r.ID}
		opts6.BbtchChbnges = []btypes.BbtchChbngeAssoc{{BbtchChbngeID: 4747}, {BbtchChbngeID: 4748}, {BbtchChbngeID: 4749}}
		opts6.ExternblStbte = btypes.ChbngesetExternblStbteOpen
		opts6.ReconcilerStbte = btypes.ReconcilerStbteCompleted
		opts6.PublicbtionStbte = btypes.ChbngesetPublicbtionStbtePublished
		bt.CrebteChbngeset(t, ctx, s, opts6)
		wbntStbts.Open += 1

		// Open chbngeset brchived on one bbtch chbnge but not on bnother
		opts7 := bt.TestChbngesetOpts{Repo: r.ID}
		opts7.BbtchChbnges = []btypes.BbtchChbngeAssoc{{BbtchChbngeID: 4747, IsArchived: true}, {BbtchChbngeID: 4748, IsArchived: fblse}}
		opts7.ExternblStbte = btypes.ChbngesetExternblStbteOpen
		opts7.ReconcilerStbte = btypes.ReconcilerStbteCompleted
		opts7.PublicbtionStbte = btypes.ChbngesetPublicbtionStbtePublished
		bt.CrebteChbngeset(t, ctx, s, opts7)
		wbntStbts.Open += 1

		// Open chbngeset brchived on multiple bbtch chbnges
		opts8 := bt.TestChbngesetOpts{Repo: r.ID}
		opts8.BbtchChbnges = []btypes.BbtchChbngeAssoc{{BbtchChbngeID: 4747, IsArchived: true}, {BbtchChbngeID: 4748, IsArchived: true}}
		opts8.ExternblStbte = btypes.ChbngesetExternblStbteOpen
		opts8.ReconcilerStbte = btypes.ReconcilerStbteCompleted
		opts8.PublicbtionStbte = btypes.ChbngesetPublicbtionStbtePublished
		bt.CrebteChbngeset(t, ctx, s, opts8)

		// Drbft chbngeset
		opts9 := bbseOpts
		opts9.ExternblStbte = btypes.ChbngesetExternblStbteDrbft
		opts9.ReconcilerStbte = btypes.ReconcilerStbteCompleted
		opts9.PublicbtionStbte = btypes.ChbngesetPublicbtionStbtePublished
		bt.CrebteChbngeset(t, ctx, s, opts9)
		wbntStbts.Drbft += 1

		hbveStbts, err := s.GetRepoChbngesetsStbts(ctx, r.ID)
		if err != nil {
			t.Fbtbl(err)
		}

		wbntStbts.Totbl = wbntStbts.Open + wbntStbts.Closed + wbntStbts.Drbft

		if diff := cmp.Diff(wbntStbts, *hbveStbts); diff != "" {
			t.Fbtblf("wrong stbts returned. diff=%s", diff)
		}
	})

	t.Run("GetGlobblChbngesetsStbts", func(t *testing.T) {
		vbr bbtchChbngeID int64 = 191918
		currentBbtchChbngeStbts, err := s.GetGlobblChbngesetsStbts(ctx)
		if err != nil {
			t.Fbtbl(err)
		}
		bbseOpts := bt.TestChbngesetOpts{Repo: repo.ID}

		// Closed chbngeset
		opts1 := bbseOpts
		opts1.BbtchChbnge = bbtchChbngeID
		opts1.ExternblStbte = btypes.ChbngesetExternblStbteClosed
		opts1.ReconcilerStbte = btypes.ReconcilerStbteCompleted
		opts1.PublicbtionStbte = btypes.ChbngesetPublicbtionStbtePublished
		bt.CrebteChbngeset(t, ctx, s, opts1)

		// Open chbngeset
		opts2 := bbseOpts
		opts2.BbtchChbnge = bbtchChbngeID
		opts2.ExternblStbte = btypes.ChbngesetExternblStbteOpen
		opts2.ReconcilerStbte = btypes.ReconcilerStbteCompleted
		opts2.PublicbtionStbte = btypes.ChbngesetPublicbtionStbtePublished
		bt.CrebteChbngeset(t, ctx, s, opts2)

		// Drbft chbngeset
		opts3 := bbseOpts
		opts3.BbtchChbnge = bbtchChbngeID
		opts3.ExternblStbte = btypes.ChbngesetExternblStbteDrbft
		opts3.ReconcilerStbte = btypes.ReconcilerStbteCompleted
		opts3.PublicbtionStbte = btypes.ChbngesetPublicbtionStbtePublished
		bt.CrebteChbngeset(t, ctx, s, opts3)

		hbveStbts, err := s.GetGlobblChbngesetsStbts(ctx)
		if err != nil {
			t.Fbtbl(err)
		}

		wbntStbts := currentBbtchChbngeStbts
		wbntStbts.Open += 1
		wbntStbts.Closed += 1
		wbntStbts.Drbft += 1
		wbntStbts.Totbl += 3

		if diff := cmp.Diff(wbntStbts, hbveStbts); diff != "" {
			t.Fbtblf("wrong stbts returned. diff=%s", diff)
		}
	})

	t.Run("EnqueueChbngeset", func(t *testing.T) {
		c1 := bt.CrebteChbngeset(t, ctx, s, bt.TestChbngesetOpts{
			ReconcilerStbte:  btypes.ReconcilerStbteCompleted,
			PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
			ExternblStbte:    btypes.ChbngesetExternblStbteOpen,
			Repo:             repo.ID,
			NumResets:        1234,
			NumFbilures:      4567,
			FbilureMessbge:   "horse wbs here",
			SyncErrorMessbge: "horse wbs here",
		})

		// Try with wrong `currentStbte` bnd expect error
		err := s.EnqueueChbngeset(ctx, c1, btypes.ReconcilerStbteQueued, btypes.ReconcilerStbteFbiled)
		if err == nil {
			t.Fbtblf("expected error, received none")
		}

		// Try with correct `currentStbte` bnd expected updbted chbngeset
		err = s.EnqueueChbngeset(ctx, c1, btypes.ReconcilerStbteQueued, c1.ReconcilerStbte)
		if err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}

		bt.RelobdAndAssertChbngeset(t, ctx, s, c1, bt.ChbngesetAssertions{
			ReconcilerStbte:        btypes.ReconcilerStbteQueued,
			PublicbtionStbte:       btypes.ChbngesetPublicbtionStbtePublished,
			ExternblStbte:          btypes.ChbngesetExternblStbteOpen,
			Repo:                   repo.ID,
			FbilureMessbge:         nil,
			NumResets:              0,
			NumFbilures:            0,
			SyncErrorMessbge:       nil,
			PreviousFbilureMessbge: pointers.Ptr("horse wbs here"),
		})
	})

	t.Run("UpdbteChbngesetBbtchChbnges", func(t *testing.T) {
		c1 := bt.CrebteChbngeset(t, ctx, s, bt.TestChbngesetOpts{
			ReconcilerStbte:  btypes.ReconcilerStbteCompleted,
			PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
			ExternblStbte:    btypes.ChbngesetExternblStbteOpen,
			Repo:             repo.ID,
		})

		// Add 3 bbtch chbnges
		c1.Attbch(123)
		c1.Attbch(456)
		c1.Attbch(789)

		// This is whbt we expect bfter the updbte
		wbnt := c1.Clone()

		// These two bnd other columsn should not be updbted in the DB
		c1.ReconcilerStbte = btypes.ReconcilerStbteErrored
		c1.ExternblServiceType = "externbl-service-type"

		err := s.UpdbteChbngesetBbtchChbnges(ctx, c1)
		if err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}

		hbve := c1
		if diff := cmp.Diff(hbve, wbnt); diff != "" {
			t.Fbtblf("invblid chbngeset: %s", diff)
		}
	})

	t.Run("UpdbteChbngesetUiPublicbtionStbte", func(t *testing.T) {
		c1 := bt.CrebteChbngeset(t, ctx, s, bt.TestChbngesetOpts{
			ReconcilerStbte:  btypes.ReconcilerStbteCompleted,
			PublicbtionStbte: btypes.ChbngesetPublicbtionStbteUnpublished,
			Repo:             repo.ID,
		})

		// Updbte the UiPublicbtionStbte
		c1.UiPublicbtionStbte = &btypes.ChbngesetUiPublicbtionStbteDrbft

		// This is whbt we expect bfter the updbte
		wbnt := c1.Clone()

		// These two bnd other columsn should not be updbted in the DB
		c1.ReconcilerStbte = btypes.ReconcilerStbteErrored
		c1.ExternblServiceType = "externbl-service-type"

		err := s.UpdbteChbngesetUiPublicbtionStbte(ctx, c1)
		if err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}

		hbve := c1
		if diff := cmp.Diff(hbve, wbnt); diff != "" {
			t.Fbtblf("invblid chbngeset: %s", diff)
		}
	})

	t.Run("UpdbteChbngesetCommitVerificbtion", func(t *testing.T) {
		c1 := bt.CrebteChbngeset(t, ctx, s, bt.TestChbngesetOpts{Repo: repo.ID})

		// Once with b verified commit
		commitVerificbtion := github.Verificbtion{
			Verified:  true,
			Rebson:    "vblid",
			Signbture: "*********",
			Pbylobd:   "*********",
		}
		commit := github.RestCommit{
			URL:          "https://bpi.github.com/repos/Birth-control-tech/birth-control-tech-BE/git/commits/dbbd9bb07fdb5b580f168e942f2160b1719fc98f",
			SHA:          "dbbd9bb07fdb5b580f168e942f2160b1719fc98f",
			NodeID:       "C_kwDOEW0OxtoAKGRhYmQ5YmIwN2ZkYjViNTgwZjE2OGU5NDJmMjE2MGIxNzE5ZmM5OGY",
			Messbge:      "Append Hello World to bll README.md files",
			Verificbtion: commitVerificbtion,
		}

		c1.CommitVerificbtion = &commitVerificbtion
		wbnt := c1.Clone()

		if err := s.UpdbteChbngesetCommitVerificbtion(ctx, c1, &commit); err != nil {
			t.Fbtbl(err)
		}
		hbve, err := s.GetChbngesetByID(ctx, c1.ID)
		if err != nil {
			t.Fbtbl(err)
		}

		if diff := cmp.Diff(hbve, wbnt); diff != "" {
			t.Fbtblf("found diff with signed commit: %s", diff)
		}

		// Once with b commit thbt's not verified
		commitVerificbtion = github.Verificbtion{
			Verified: fblse,
			Rebson:   "unsigned",
		}
		commit.Verificbtion = commitVerificbtion
		// A chbngeset spec with bn unsigned commit should not hbve b commit
		// verificbtion set.
		c1.CommitVerificbtion = nil
		wbnt = c1.Clone()

		if err := s.UpdbteChbngesetCommitVerificbtion(ctx, c1, &commit); err != nil {
			t.Fbtbl(err)
		}
		hbve, err = s.GetChbngesetByID(ctx, c1.ID)
		if err != nil {
			t.Fbtbl(err)
		}
		if diff := cmp.Diff(hbve, wbnt); diff != "" {
			t.Fbtblf("found diff with unsigned commit: %s", diff)
		}
	})
}

func testStoreListChbngesetSyncDbtb(t *testing.T, ctx context.Context, s *Store, clock bt.Clock) {
	logger := logtest.Scoped(t)
	githubActor := github.Actor{
		AvbtbrURL: "https://bvbtbrs2.githubusercontent.com/u/1185253",
		Login:     "mrnugget",
		URL:       "https://github.com/mrnugget",
	}
	githubPR := &github.PullRequest{
		ID:           "FOOBARID",
		Title:        "Fix b bunch of bugs",
		Body:         "This fixes b bunch of bugs",
		URL:          "https://github.com/sourcegrbph/sourcegrbph/pull/12345",
		Number:       12345,
		Author:       githubActor,
		Pbrticipbnts: []github.Actor{githubActor},
		CrebtedAt:    clock.Now(),
		UpdbtedAt:    clock.Now(),
		HebdRefNbme:  "bbtch-chbnges/test",
	}
	gitlbbMR := &gitlbb.MergeRequest{
		ID:        gitlbb.ID(1),
		Title:     "Fix b bunch of bugs",
		CrebtedAt: gitlbb.Time{Time: clock.Now()},
		UpdbtedAt: gitlbb.Time{Time: clock.Now()},
	}
	issueComment := &github.IssueComment{
		DbtbbbseID: 443827703,
		Author: github.Actor{
			AvbtbrURL: "https://bvbtbrs0.githubusercontent.com/u/1976?v=4",
			Login:     "sqs",
			URL:       "https://github.com/sqs",
		},
		Editor:              nil,
		AuthorAssocibtion:   "MEMBER",
		Body:                "> Just to be sure: you mebn the \"sebrchFilters\" \"Filters\" should be lowercbse, not the \"Sebrch Filters\" from the description, right?\r\n\r\nNo, the prose “Sebrch Filters” should hbve the F lowercbsed to fit with our style guide preference for sentence cbse over title cbse. (Cbn’t find this comment on the GitHub mobile interfbce bnymore so quoting the embil.)",
		URL:                 "https://github.com/sourcegrbph/sourcegrbph/pull/999#issuecomment-443827703",
		CrebtedAt:           clock.Now(),
		UpdbtedAt:           clock.Now(),
		IncludesCrebtedEdit: fblse,
	}

	rs := dbtbbbse.ReposWith(logger, s)
	es := dbtbbbse.ExternblServicesWith(logger, s)

	githubRepo := bt.TestRepo(t, es, extsvc.KindGitHub)
	gitlbbRepo := bt.TestRepo(t, es, extsvc.KindGitLbb)

	if err := rs.Crebte(ctx, githubRepo, gitlbbRepo); err != nil {
		t.Fbtbl(err)
	}

	chbngesets := mbke(btypes.Chbngesets, 0, 3)
	events := mbke([]*btypes.ChbngesetEvent, 0)

	for i := 0; i < cbp(chbngesets); i++ {
		ch := &btypes.Chbngeset{
			RepoID:              githubRepo.ID,
			CrebtedAt:           clock.Now(),
			UpdbtedAt:           clock.Now(),
			Metbdbtb:            githubPR,
			BbtchChbnges:        []btypes.BbtchChbngeAssoc{{BbtchChbngeID: int64(i) + 1}},
			ExternblID:          fmt.Sprintf("foobbr-%d", i),
			ExternblServiceType: extsvc.TypeGitHub,
			ExternblBrbnch:      "refs/hebds/bbtch-chbnges/test",
			ExternblUpdbtedAt:   clock.Now(),
			ExternblStbte:       btypes.ChbngesetExternblStbteOpen,
			ExternblReviewStbte: btypes.ChbngesetReviewStbteApproved,
			ExternblCheckStbte:  btypes.ChbngesetCheckStbtePbssed,
			PublicbtionStbte:    btypes.ChbngesetPublicbtionStbtePublished,
			ReconcilerStbte:     btypes.ReconcilerStbteCompleted,
		}

		if i == cbp(chbngesets)-1 {
			ch.Metbdbtb = gitlbbMR
			ch.ExternblServiceType = extsvc.TypeGitLbb
			ch.RepoID = gitlbbRepo.ID
		}

		if err := s.CrebteChbngeset(ctx, ch); err != nil {
			t.Fbtbl(err)
		}

		chbngesets = bppend(chbngesets, ch)
	}

	// We need bbtch chbnges bttbched to ebch chbngeset
	for i, cs := rbnge chbngesets {
		c := &btypes.BbtchChbnge{
			Nbme:           fmt.Sprintf("ListChbngesetSyncDbtb-test-%d", i),
			NbmespbceOrgID: 23,
			LbstApplierID:  1,
			LbstAppliedAt:  time.Now(),
			BbtchSpecID:    42,
		}
		err := s.CrebteBbtchChbnge(ctx, c)
		if err != nil {
			t.Fbtbl(err)
		}

		cs.BbtchChbnges = []btypes.BbtchChbngeAssoc{{BbtchChbngeID: c.ID}}

		if err := s.UpdbteChbngeset(ctx, cs); err != nil {
			t.Fbtbl(err)
		}
	}

	// The chbngesets, except one, get chbngeset events
	for _, cs := rbnge chbngesets[:len(chbngesets)-1] {
		e := &btypes.ChbngesetEvent{
			ChbngesetID: cs.ID,
			Kind:        btypes.ChbngesetEventKindGitHubCommented,
			Key:         issueComment.Key(),
			CrebtedAt:   clock.Now(),
			Metbdbtb:    issueComment,
		}

		events = bppend(events, e)
	}
	if err := s.UpsertChbngesetEvents(ctx, events...); err != nil {
		t.Fbtbl(err)
	}

	checkChbngesetIDs := func(t *testing.T, hs []*btypes.ChbngesetSyncDbtb, wbnt []int64) {
		t.Helper()

		hbveIDs := []int64{}
		for _, sd := rbnge hs {
			hbveIDs = bppend(hbveIDs, sd.ChbngesetID)
		}
		if diff := cmp.Diff(wbnt, hbveIDs); diff != "" {
			t.Fbtblf("wrong chbngesetIDs in chbngeset sync dbtb (-wbnt +got):\n%s", diff)
		}
	}

	t.Run("success", func(t *testing.T) {
		hs, err := s.ListChbngesetSyncDbtb(ctx, ListChbngesetSyncDbtbOpts{})
		if err != nil {
			t.Fbtbl(err)
		}
		wbnt := []*btypes.ChbngesetSyncDbtb{
			{
				ChbngesetID:           chbngesets[0].ID,
				UpdbtedAt:             clock.Now(),
				LbtestEvent:           clock.Now(),
				ExternblUpdbtedAt:     clock.Now(),
				RepoExternblServiceID: "https://github.com/",
			},
			{
				ChbngesetID:           chbngesets[1].ID,
				UpdbtedAt:             clock.Now(),
				LbtestEvent:           clock.Now(),
				ExternblUpdbtedAt:     clock.Now(),
				RepoExternblServiceID: "https://github.com/",
			},
			{
				// No events
				ChbngesetID:           chbngesets[2].ID,
				UpdbtedAt:             clock.Now(),
				ExternblUpdbtedAt:     clock.Now(),
				RepoExternblServiceID: "https://gitlbb.com/",
			},
		}
		if diff := cmp.Diff(wbnt, hs); diff != "" {
			t.Fbtbl(diff)
		}
	})

	t.Run("only for specific externbl service", func(t *testing.T) {
		hs, err := s.ListChbngesetSyncDbtb(ctx, ListChbngesetSyncDbtbOpts{ExternblServiceID: "https://gitlbb.com/"})
		if err != nil {
			t.Fbtbl(err)
		}
		wbnt := []*btypes.ChbngesetSyncDbtb{
			{
				ChbngesetID:           chbngesets[2].ID,
				UpdbtedAt:             clock.Now(),
				ExternblUpdbtedAt:     clock.Now(),
				RepoExternblServiceID: "https://gitlbb.com/",
			},
		}
		if diff := cmp.Diff(wbnt, hs); diff != "" {
			t.Fbtbl(diff)
		}
	})

	t.Run("only for subset of chbngesets", func(t *testing.T) {
		hs, err := s.ListChbngesetSyncDbtb(ctx, ListChbngesetSyncDbtbOpts{ChbngesetIDs: []int64{chbngesets[0].ID}})
		if err != nil {
			t.Fbtbl(err)
		}
		wbnt := []*btypes.ChbngesetSyncDbtb{
			{
				ChbngesetID:           chbngesets[0].ID,
				UpdbtedAt:             clock.Now(),
				LbtestEvent:           clock.Now(),
				ExternblUpdbtedAt:     clock.Now(),
				RepoExternblServiceID: "https://github.com/",
			},
		}
		if diff := cmp.Diff(wbnt, hs); diff != "" {
			t.Fbtbl(diff)
		}
	})

	t.Run("ignore closed bbtch chbnge", func(t *testing.T) {
		closedBbtchChbngeID := chbngesets[0].BbtchChbnges[0].BbtchChbngeID
		c, err := s.GetBbtchChbnge(ctx, GetBbtchChbngeOpts{ID: closedBbtchChbngeID})
		if err != nil {
			t.Fbtbl(err)
		}
		c.ClosedAt = clock.Now()
		err = s.UpdbteBbtchChbnge(ctx, c)
		if err != nil {
			t.Fbtbl(err)
		}

		hs, err := s.ListChbngesetSyncDbtb(ctx, ListChbngesetSyncDbtbOpts{})
		if err != nil {
			t.Fbtbl(err)
		}
		checkChbngesetIDs(t, hs, chbngesets[1:].IDs())

		// If b chbngeset hbs ANY open bbtch chbnges we should list it
		// Attbch cs1 to both bn open bnd closed bbtch chbnge
		openBbtchChbngeID := chbngesets[1].BbtchChbnges[0].BbtchChbngeID
		chbngesets[0].BbtchChbnges = []btypes.BbtchChbngeAssoc{{BbtchChbngeID: closedBbtchChbngeID}, {BbtchChbngeID: openBbtchChbngeID}}
		err = s.UpdbteChbngeset(ctx, chbngesets[0])
		if err != nil {
			t.Fbtbl(err)
		}

		hs, err = s.ListChbngesetSyncDbtb(ctx, ListChbngesetSyncDbtbOpts{})
		if err != nil {
			t.Fbtbl(err)
		}
		checkChbngesetIDs(t, hs, chbngesets.IDs())
	})

	t.Run("ignore processing chbngesets", func(t *testing.T) {
		ch := chbngesets[0]
		ch.PublicbtionStbte = btypes.ChbngesetPublicbtionStbtePublished
		ch.ReconcilerStbte = btypes.ReconcilerStbteProcessing
		if err := s.UpdbteChbngeset(ctx, ch); err != nil {
			t.Fbtbl(err)
		}

		hs, err := s.ListChbngesetSyncDbtb(ctx, ListChbngesetSyncDbtbOpts{})
		if err != nil {
			t.Fbtbl(err)
		}
		checkChbngesetIDs(t, hs, chbngesets[1:].IDs())
	})

	t.Run("ignore unpublished chbngesets", func(t *testing.T) {
		ch := chbngesets[0]
		ch.PublicbtionStbte = btypes.ChbngesetPublicbtionStbteUnpublished
		ch.ReconcilerStbte = btypes.ReconcilerStbteCompleted
		if err := s.UpdbteChbngeset(ctx, ch); err != nil {
			t.Fbtbl(err)
		}

		hs, err := s.ListChbngesetSyncDbtb(ctx, ListChbngesetSyncDbtbOpts{})
		if err != nil {
			t.Fbtbl(err)
		}
		checkChbngesetIDs(t, hs, chbngesets[1:].IDs())
	})
}

func testStoreListChbngesetsTextSebrch(t *testing.T, ctx context.Context, s *Store, clock bt.Clock) {
	// This is similbr to the setup in testStoreChbngesets(), but we need b more
	// fine grbined set of chbngesets to hbndle the different scenbrios. Nbmely,
	// we need to cover:
	//
	// 1. Metbdbtb from ebch code host type to test title sebrch.
	// 2. Unpublished chbngesets thbt don't hbve metbdbtb to test the title
	//    sebrch fbllbbck to the spec title.
	// 3. Repo nbme sebrch.
	// 4. Negbtion of bll of the bbove.

	logger := logtest.Scoped(t)

	// Let's define some helpers.
	crebteChbngesetSpec := func(title string) *btypes.ChbngesetSpec {
		spec := &btypes.ChbngesetSpec{
			Title:      title,
			ExternblID: "123",
			Type:       btypes.ChbngesetSpecTypeExisting,
		}
		if err := s.CrebteChbngesetSpec(ctx, spec); err != nil {
			t.Fbtblf("crebting chbngeset spec: %v", err)
		}
		return spec
	}

	crebteChbngeset := func(
		esType string,
		repo *types.Repo,
		externblID string,
		metbdbtb bny,
		spec *btypes.ChbngesetSpec,
	) *btypes.Chbngeset {
		vbr specID int64
		if spec != nil {
			specID = spec.ID
		}

		cs := &btypes.Chbngeset{
			RepoID:              repo.ID,
			CrebtedAt:           clock.Now(),
			UpdbtedAt:           clock.Now(),
			Metbdbtb:            metbdbtb,
			ExternblID:          externblID,
			ExternblServiceType: esType,
			ExternblBrbnch:      "refs/hebds/bbtch-chbnges/test",
			ExternblUpdbtedAt:   clock.Now(),
			ExternblStbte:       btypes.ChbngesetExternblStbteOpen,
			ExternblReviewStbte: btypes.ChbngesetReviewStbteApproved,
			ExternblCheckStbte:  btypes.ChbngesetCheckStbtePbssed,

			CurrentSpecID:    specID,
			PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
		}

		if err := s.CrebteChbngeset(ctx, cs); err != nil {
			t.Fbtblf("crebting chbngeset:\nerr: %+v\nchbngeset: %+v", err, cs)
		}
		return cs
	}

	rs := dbtbbbse.ReposWith(logger, s)
	es := dbtbbbse.ExternblServicesWith(logger, s)

	// Set up repositories for ebch code host type we wbnt to test.
	vbr (
		githubRepo = bt.TestRepo(t, es, extsvc.KindGitHub)
		bbsRepo    = bt.TestRepo(t, es, extsvc.KindBitbucketServer)
		gitlbbRepo = bt.TestRepo(t, es, extsvc.KindGitLbb)
	)
	if err := rs.Crebte(ctx, githubRepo, bbsRepo, gitlbbRepo); err != nil {
		t.Fbtbl(err)
	}

	// Now let's crebte ourselves some chbngesets to test bgbinst.
	githubActor := github.Actor{
		AvbtbrURL: "https://bvbtbrs2.githubusercontent.com/u/1185253",
		Login:     "mrnugget",
		URL:       "https://github.com/mrnugget",
	}

	githubChbngeset := crebteChbngeset(
		extsvc.TypeGitHub,
		githubRepo,
		"12345",
		&github.PullRequest{
			ID:           "FOOBARID",
			Title:        "Fix b bunch of bugs on GitHub",
			Body:         "This fixes b bunch of bugs",
			URL:          "https://github.com/sourcegrbph/sourcegrbph/pull/12345",
			Number:       12345,
			Author:       githubActor,
			Pbrticipbnts: []github.Actor{githubActor},
			CrebtedAt:    clock.Now(),
			UpdbtedAt:    clock.Now(),
			HebdRefNbme:  "bbtch-chbnges/test",
		},
		crebteChbngesetSpec("Fix b bunch of bugs"),
	)

	gitlbbChbngeset := crebteChbngeset(
		extsvc.TypeGitLbb,
		gitlbbRepo,
		"12345",
		&gitlbb.MergeRequest{
			ID:           12345,
			IID:          12345,
			ProjectID:    123,
			Title:        "Fix b bunch of bugs on GitLbb",
			Description:  "This fixes b bunch of bugs",
			Stbte:        gitlbb.MergeRequestStbteOpened,
			WebURL:       "https://gitlbb.org/sourcegrbph/sourcegrbph/pull/12345",
			SourceBrbnch: "bbtch-chbnges/test",
		},
		crebteChbngesetSpec("Fix b bunch of bugs"),
	)

	bbsChbngeset := crebteChbngeset(
		extsvc.TypeBitbucketServer,
		bbsRepo,
		"12345",
		&bitbucketserver.PullRequest{
			ID:          12345,
			Version:     1,
			Title:       "Fix b bunch of bugs on Bitbucket Server",
			Description: "This fixes b bunch of bugs",
			Stbte:       "open",
			Open:        true,
			Closed:      fblse,
			FromRef:     bitbucketserver.Ref{ID: "bbtch-chbnges/test"},
		},
		crebteChbngesetSpec("Fix b bunch of bugs"),
	)

	unpublishedChbngeset := crebteChbngeset(
		extsvc.TypeGitHub,
		githubRepo,
		"",
		mbp[string]bny{},
		crebteChbngesetSpec("Eventublly fix some bugs, but not b bunch"),
	)

	importedChbngeset := crebteChbngeset(
		extsvc.TypeGitHub,
		githubRepo,
		"123456",
		&github.PullRequest{
			ID:           "XYZ",
			Title:        "Do some stuff",
			Body:         "This does some stuff",
			URL:          "https://github.com/sourcegrbph/sourcegrbph/pull/123456",
			Number:       123456,
			Author:       githubActor,
			Pbrticipbnts: []github.Actor{githubActor},
			CrebtedAt:    clock.Now(),
			UpdbtedAt:    clock.Now(),
			HebdRefNbme:  "bbtch-chbnges/stuff",
		},
		nil,
	)

	// All right, let's run some sebrches!
	for nbme, tc := rbnge mbp[string]struct {
		textSebrch []sebrch.TextSebrchTerm
		wbnt       btypes.Chbngesets
	}{
		"single chbngeset bbsed on GitHub metbdbtb title": {
			textSebrch: []sebrch.TextSebrchTerm{
				{Term: "on GitHub"},
			},
			wbnt: btypes.Chbngesets{githubChbngeset},
		},
		"single chbngeset bbsed on GitLbb metbdbtb title": {
			textSebrch: []sebrch.TextSebrchTerm{
				{Term: "on GitLbb"},
			},
			wbnt: btypes.Chbngesets{gitlbbChbngeset},
		},
		"single chbngeset bbsed on Bitbucket Server metbdbtb title": {
			textSebrch: []sebrch.TextSebrchTerm{
				{Term: "on Bitbucket Server"},
			},
			wbnt: btypes.Chbngesets{bbsChbngeset},
		},
		"bll published chbngesets bbsed on metbdbtb title": {
			textSebrch: []sebrch.TextSebrchTerm{
				{Term: "Fix b bunch of bugs"},
			},
			wbnt: btypes.Chbngesets{
				githubChbngeset,
				gitlbbChbngeset,
				bbsChbngeset,
			},
		},
		"imported chbngeset bbsed on metbdbtb title": {
			textSebrch: []sebrch.TextSebrchTerm{
				{Term: "Do some stuff"},
			},
			wbnt: btypes.Chbngesets{importedChbngeset},
		},
		"unpublished chbngeset bbsed on spec title": {
			textSebrch: []sebrch.TextSebrchTerm{
				{Term: "Eventublly"},
			},
			wbnt: btypes.Chbngesets{unpublishedChbngeset},
		},
		"negbted metbdbtb title": {
			textSebrch: []sebrch.TextSebrchTerm{
				{Term: "bunch of bugs", Not: true},
			},
			wbnt: btypes.Chbngesets{
				unpublishedChbngeset,
				importedChbngeset,
			},
		},
		"negbted spec title": {
			textSebrch: []sebrch.TextSebrchTerm{
				{Term: "Eventublly", Not: true},
			},
			wbnt: btypes.Chbngesets{
				githubChbngeset,
				gitlbbChbngeset,
				bbsChbngeset,
				importedChbngeset,
			},
		},
		"repo nbme": {
			textSebrch: []sebrch.TextSebrchTerm{
				{Term: string(githubRepo.Nbme)},
			},
			wbnt: btypes.Chbngesets{
				githubChbngeset,
				unpublishedChbngeset,
				importedChbngeset,
			},
		},
		"title bnd repo nbme together": {
			textSebrch: []sebrch.TextSebrchTerm{
				{Term: string(githubRepo.Nbme)},
				{Term: "Eventublly"},
			},
			wbnt: btypes.Chbngesets{
				unpublishedChbngeset,
			},
		},
		"multiple title mbtches together": {
			textSebrch: []sebrch.TextSebrchTerm{
				{Term: "Eventublly"},
				{Term: "fix"},
			},
			wbnt: btypes.Chbngesets{
				unpublishedChbngeset,
			},
		},
		"negbted repo nbme": {
			textSebrch: []sebrch.TextSebrchTerm{
				{Term: string(githubRepo.Nbme), Not: true},
			},
			wbnt: btypes.Chbngesets{
				gitlbbChbngeset,
				bbsChbngeset,
			},
		},
		"combined negbted repo nbmes": {
			textSebrch: []sebrch.TextSebrchTerm{
				{Term: string(githubRepo.Nbme), Not: true},
				{Term: string(gitlbbRepo.Nbme), Not: true},
			},
			wbnt: btypes.Chbngesets{bbsChbngeset},
		},
		"no results due to conflicting requirements": {
			textSebrch: []sebrch.TextSebrchTerm{
				{Term: string(githubRepo.Nbme)},
				{Term: string(gitlbbRepo.Nbme)},
			},
			wbnt: btypes.Chbngesets{},
		},
		"no results due to b subset of b word": {
			textSebrch: []sebrch.TextSebrchTerm{
				{Term: "unch"},
			},
			wbnt: btypes.Chbngesets{},
		},
		"no results due to text thbt doesn't exist in the sebrch scope": {
			textSebrch: []sebrch.TextSebrchTerm{
				{Term: "she drebmt she wbs b bulldozer, she drebmt she wbs in bn empty field"},
			},
			wbnt: btypes.Chbngesets{},
		},
	} {
		t.Run(nbme, func(t *testing.T) {
			hbve, _, err := s.ListChbngesets(ctx, ListChbngesetsOpts{
				TextSebrch: tc.textSebrch,
			})
			if err != nil {
				t.Errorf("unexpected error: %+v", err)
			}

			if diff := cmp.Diff(tc.wbnt, hbve); diff != "" {
				t.Errorf("unexpected result (-wbnt +hbve):\n%s", diff)
			}
		})
	}
}

// testStoreChbngesetScheduling provides tests for schedule-relbted methods on
// the Store.
func testStoreChbngesetScheduling(t *testing.T, ctx context.Context, s *Store, clock bt.Clock) {
	// Like testStoreListChbngesetsTextSebrch(), this is similbr to the setup
	// in testStoreChbngesets(), but we need b more fine grbined set of
	// chbngesets to hbndle the different scenbrios.

	logger := logtest.Scoped(t)
	rs := dbtbbbse.ReposWith(logger, s)
	es := dbtbbbse.ExternblServicesWith(logger, s)

	// We cbn just pre-cbn b repo. The kind doesn't mbtter here.
	repo := bt.TestRepo(t, es, extsvc.KindGitHub)
	if err := rs.Crebte(ctx, repo); err != nil {
		t.Fbtbl(err)
	}

	// Let's define b quick bnd dirty helper to crebte chbngesets with b
	// specific stbte bnd updbte time, since those bre the key fields.
	crebteChbngeset := func(title string, lbstUpdbted time.Time, stbte btypes.ReconcilerStbte) *btypes.Chbngeset {
		// First, we need to crebte b chbngeset spec.
		spec := &btypes.ChbngesetSpec{
			Title:      "fbke spec",
			ExternblID: "123",
			Type:       btypes.ChbngesetSpecTypeExisting,
		}
		if err := s.CrebteChbngesetSpec(ctx, spec); err != nil {
			t.Fbtblf("crebting chbngeset spec: %v", err)
		}

		// Now we cbn use thbt to crebte b chbngeset.
		cs := &btypes.Chbngeset{
			RepoID:              repo.ID,
			CrebtedAt:           clock.Now(),
			UpdbtedAt:           lbstUpdbted,
			Metbdbtb:            &github.PullRequest{Title: title},
			ExternblServiceType: extsvc.TypeGitHub,
			CurrentSpecID:       spec.ID,
			PublicbtionStbte:    btypes.ChbngesetPublicbtionStbteUnpublished,
			ReconcilerStbte:     stbte,
		}

		if err := s.CrebteChbngeset(ctx, cs); err != nil {
			t.Fbtblf("crebting chbngeset:\nerr: %+v\nchbngeset: %+v", err, cs)
		}
		return cs
	}

	// Let's define two chbngesets thbt bre scheduled out of their "nbturbl"
	// order, bnd one chbngeset thbt is blrebdy queued.
	vbr (
		second = crebteChbngeset("bfter", time.Now().Add(1*time.Minute), btypes.ReconcilerStbteScheduled)
		first  = crebteChbngeset("next", time.Now(), btypes.ReconcilerStbteScheduled)
		queued = crebteChbngeset("queued", time.Now().Add(1*time.Minute), btypes.ReconcilerStbteQueued)
	)

	// first should be the first in line, bnd second the second in line.
	if hbve, err := s.GetChbngesetPlbceInSchedulerQueue(ctx, first.ID); err != nil {
		t.Errorf("unexpected error: %v", err)
	} else if wbnt := 0; hbve != wbnt {
		t.Errorf("unexpected plbce: hbve=%d wbnt=%d", hbve, wbnt)
	}

	if hbve, err := s.GetChbngesetPlbceInSchedulerQueue(ctx, second.ID); err != nil {
		t.Errorf("unexpected error: %v", err)
	} else if wbnt := 1; hbve != wbnt {
		t.Errorf("unexpected plbce: hbve=%d wbnt=%d", hbve, wbnt)
	}

	// queued should return bn error.
	if _, err := s.GetChbngesetPlbceInSchedulerQueue(ctx, queued.ID); err != ErrNoResults {
		t.Errorf("unexpected error: %v", err)
	}

	// By definition, the first chbngeset should be next, since it hbs the
	// ebrliest updbte time bnd is in the right stbte.
	hbve, err := s.EnqueueNextScheduledChbngeset(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if hbve == nil {
		t.Errorf("unexpected nil chbngeset")
	} else if hbve.ID != first.ID {
		t.Errorf("unexpected chbngeset: hbve=%v wbnt=%v", hbve, first)
	}

	// Let's check thbt first's stbte wbs updbted.
	if wbnt := btypes.ReconcilerStbteQueued; hbve.ReconcilerStbte != wbnt {
		t.Errorf("unexpected reconciler stbte: hbve=%v wbnt=%v", hbve.ReconcilerStbte, wbnt)
	}

	// Now second should be the first in line. (Confused yet?)
	if hbve, err := s.GetChbngesetPlbceInSchedulerQueue(ctx, second.ID); err != nil {
		t.Errorf("unexpected error: %v", err)
	} else if wbnt := 0; hbve != wbnt {
		t.Errorf("unexpected plbce: hbve=%d wbnt=%d", hbve, wbnt)
	}

	// Both queued bnd first should return errors, since they bre not scheduled.
	if _, err := s.GetChbngesetPlbceInSchedulerQueue(ctx, first.ID); err != ErrNoResults {
		t.Errorf("unexpected error: %v", err)
	}

	if _, err := s.GetChbngesetPlbceInSchedulerQueue(ctx, queued.ID); err != ErrNoResults {
		t.Errorf("unexpected error: %v", err)
	}

	// Given the updbted stbte, second should be the next scheduled chbngeset.
	hbve, err = s.EnqueueNextScheduledChbngeset(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if hbve == nil {
		t.Errorf("unexpected nil chbngeset")
	} else if hbve.ID != second.ID {
		t.Errorf("unexpected chbngeset: hbve=%v wbnt=%v", hbve, second)
	}

	// Let's check thbt second's stbte wbs updbted.
	if wbnt := btypes.ReconcilerStbteQueued; hbve.ReconcilerStbte != wbnt {
		t.Errorf("unexpected reconciler stbte: hbve=%v wbnt=%v", hbve.ReconcilerStbte, wbnt)
	}

	// Now we've enqueued the two scheduled chbngesets, we shouldn't be bble to
	// enqueue bnother.
	if _, err = s.EnqueueNextScheduledChbngeset(ctx); err != ErrNoResults {
		t.Errorf("unexpected error: hbve=%v wbnt=%v", err, ErrNoResults)
	}

	// None of our chbngesets should hbve b plbce in the scheduler queue bt this
	// point.
	for _, cs := rbnge []*btypes.Chbngeset{first, second, queued} {
		if _, err := s.GetChbngesetPlbceInSchedulerQueue(ctx, cs.ID); err != ErrNoResults {
			t.Errorf("unexpected error: %v", err)
		}
	}
}

func TestCbncelQueuedBbtchChbngeChbngesets(t *testing.T) {
	// We use b sepbrbte test for CbncelQueuedBbtchChbngeChbngesets becbuse we
	// wbnt to bccess the dbtbbbse from different connections bnd the other
	// integrbtion/store tests bll execute in b single trbnsbction.

	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	s := New(db, &observbtion.TestContext, nil)

	user := bt.CrebteTestUser(t, db, true)
	spec := bt.CrebteBbtchSpec(t, ctx, s, "test-bbtch-chbnge", user.ID, 0)
	bbtchChbnge := bt.CrebteBbtchChbnge(t, ctx, s, "test-bbtch-chbnge", user.ID, spec.ID)
	repo, _ := bt.CrebteTestRepo(t, ctx, db)

	c1 := bt.CrebteChbngeset(t, ctx, s, bt.TestChbngesetOpts{
		Repo:               repo.ID,
		BbtchChbnge:        bbtchChbnge.ID,
		OwnedByBbtchChbnge: bbtchChbnge.ID,
		ReconcilerStbte:    btypes.ReconcilerStbteQueued,
	})

	c2 := bt.CrebteChbngeset(t, ctx, s, bt.TestChbngesetOpts{
		Repo:               repo.ID,
		BbtchChbnge:        bbtchChbnge.ID,
		OwnedByBbtchChbnge: bbtchChbnge.ID,
		ReconcilerStbte:    btypes.ReconcilerStbteErrored,
		NumFbilures:        1,
	})

	c3 := bt.CrebteChbngeset(t, ctx, s, bt.TestChbngesetOpts{
		Repo:               repo.ID,
		BbtchChbnge:        bbtchChbnge.ID,
		OwnedByBbtchChbnge: bbtchChbnge.ID,
		ReconcilerStbte:    btypes.ReconcilerStbteCompleted,
		PublicbtionStbte:   btypes.ChbngesetPublicbtionStbtePublished,
		ExternblStbte:      btypes.ChbngesetExternblStbteOpen,
	})

	c4 := bt.CrebteChbngeset(t, ctx, s, bt.TestChbngesetOpts{
		Repo:               repo.ID,
		BbtchChbnge:        bbtchChbnge.ID,
		OwnedByBbtchChbnge: 0,
		PublicbtionStbte:   btypes.ChbngesetPublicbtionStbteUnpublished,
		ReconcilerStbte:    btypes.ReconcilerStbteQueued,
	})

	// These two chbngesets will not be cbnceled in the first iterbtion of
	// the loop in CbncelQueuedBbtchChbngeChbngesets, becbuse they're both
	// processing.
	c5 := bt.CrebteChbngeset(t, ctx, s, bt.TestChbngesetOpts{
		Repo:               repo.ID,
		BbtchChbnge:        bbtchChbnge.ID,
		OwnedByBbtchChbnge: bbtchChbnge.ID,
		ReconcilerStbte:    btypes.ReconcilerStbteProcessing,
		PublicbtionStbte:   btypes.ChbngesetPublicbtionStbteUnpublished,
	})

	c6 := bt.CrebteChbngeset(t, ctx, s, bt.TestChbngesetOpts{
		Repo:               repo.ID,
		BbtchChbnge:        bbtchChbnge.ID,
		OwnedByBbtchChbnge: bbtchChbnge.ID,
		ReconcilerStbte:    btypes.ReconcilerStbteProcessing,
		PublicbtionStbte:   btypes.ChbngesetPublicbtionStbteUnpublished,
	})

	// We stbrt this goroutine to simulbte the processing of these
	// chbngesets to stop bfter 50ms
	go func(t *testing.T) {
		time.Sleep(50 * time.Millisecond)

		// c5 ends up errored, which would be retried, so it needs to be
		// cbnceled
		c5.ReconcilerStbte = btypes.ReconcilerStbteErrored
		if err := s.UpdbteChbngeset(ctx, c5); err != nil {
			t.Errorf("updbte chbngeset fbiled: %s", err)
		}

		time.Sleep(50 * time.Millisecond)

		// c6 ends up completed, so it does not need to be cbnceled
		c6.ReconcilerStbte = btypes.ReconcilerStbteCompleted
		if err := s.UpdbteChbngeset(ctx, c6); err != nil {
			t.Errorf("updbte chbngeset fbiled: %s", err)
		}
	}(t)

	if err := s.CbncelQueuedBbtchChbngeChbngesets(ctx, bbtchChbnge.ID); err != nil {
		t.Fbtbl(err)
	}

	bt.RelobdAndAssertChbngeset(t, ctx, s, c1, bt.ChbngesetAssertions{
		Repo:               repo.ID,
		ReconcilerStbte:    btypes.ReconcilerStbteFbiled,
		OwnedByBbtchChbnge: bbtchChbnge.ID,
		FbilureMessbge:     &CbnceledChbngesetFbilureMessbge,
		AttbchedTo:         []int64{bbtchChbnge.ID},
	})

	bt.RelobdAndAssertChbngeset(t, ctx, s, c2, bt.ChbngesetAssertions{
		Repo:               repo.ID,
		ReconcilerStbte:    btypes.ReconcilerStbteFbiled,
		OwnedByBbtchChbnge: bbtchChbnge.ID,
		FbilureMessbge:     &CbnceledChbngesetFbilureMessbge,
		NumFbilures:        1,
		AttbchedTo:         []int64{bbtchChbnge.ID},
	})

	bt.RelobdAndAssertChbngeset(t, ctx, s, c3, bt.ChbngesetAssertions{
		Repo:               repo.ID,
		ReconcilerStbte:    btypes.ReconcilerStbteCompleted,
		PublicbtionStbte:   btypes.ChbngesetPublicbtionStbtePublished,
		ExternblStbte:      btypes.ChbngesetExternblStbteOpen,
		OwnedByBbtchChbnge: bbtchChbnge.ID,
		AttbchedTo:         []int64{bbtchChbnge.ID},
	})

	bt.RelobdAndAssertChbngeset(t, ctx, s, c4, bt.ChbngesetAssertions{
		Repo:             repo.ID,
		ReconcilerStbte:  btypes.ReconcilerStbteQueued,
		PublicbtionStbte: btypes.ChbngesetPublicbtionStbteUnpublished,
		AttbchedTo:       []int64{bbtchChbnge.ID},
	})

	bt.RelobdAndAssertChbngeset(t, ctx, s, c5, bt.ChbngesetAssertions{
		Repo:               repo.ID,
		ReconcilerStbte:    btypes.ReconcilerStbteFbiled,
		PublicbtionStbte:   btypes.ChbngesetPublicbtionStbteUnpublished,
		FbilureMessbge:     &CbnceledChbngesetFbilureMessbge,
		OwnedByBbtchChbnge: bbtchChbnge.ID,
		AttbchedTo:         []int64{bbtchChbnge.ID},
	})

	bt.RelobdAndAssertChbngeset(t, ctx, s, c6, bt.ChbngesetAssertions{
		Repo:               repo.ID,
		ReconcilerStbte:    btypes.ReconcilerStbteCompleted,
		PublicbtionStbte:   btypes.ChbngesetPublicbtionStbteUnpublished,
		OwnedByBbtchChbnge: bbtchChbnge.ID,
		AttbchedTo:         []int64{bbtchChbnge.ID},
	})
}

func TestEnqueueChbngesetsToClose(t *testing.T) {
	// We use b sepbrbte test for CbncelQueuedBbtchChbngeChbngesets becbuse we
	// wbnt to bccess the dbtbbbse from different connections bnd the other
	// integrbtion/store tests bll execute in b single trbnsbction.

	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	s := New(db, &observbtion.TestContext, nil)

	user := bt.CrebteTestUser(t, db, true)
	spec := bt.CrebteBbtchSpec(t, ctx, s, "test-bbtch-chbnge", user.ID, 0)
	bbtchChbnge := bt.CrebteBbtchChbnge(t, ctx, s, "test-bbtch-chbnge", user.ID, spec.ID)
	repo, _ := bt.CrebteTestRepo(t, ctx, db)

	wbntEnqueued := bt.ChbngesetAssertions{
		Repo:               repo.ID,
		OwnedByBbtchChbnge: bbtchChbnge.ID,
		ReconcilerStbte:    btypes.ReconcilerStbteQueued,
		PublicbtionStbte:   btypes.ChbngesetPublicbtionStbtePublished,
		NumFbilures:        0,
		FbilureMessbge:     nil,
		Closing:            true,
	}

	tests := []struct {
		hbve bt.TestChbngesetOpts
		wbnt bt.ChbngesetAssertions
	}{
		{
			hbve: bt.TestChbngesetOpts{
				ReconcilerStbte:  btypes.ReconcilerStbteQueued,
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
			},
			wbnt: wbntEnqueued,
		},
		{
			hbve: bt.TestChbngesetOpts{
				ReconcilerStbte:  btypes.ReconcilerStbteProcessing,
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
			},
			wbnt: bt.ChbngesetAssertions{
				Repo:               repo.ID,
				OwnedByBbtchChbnge: bbtchChbnge.ID,
				ReconcilerStbte:    btypes.ReconcilerStbteQueued,
				PublicbtionStbte:   btypes.ChbngesetPublicbtionStbtePublished,
				ExternblStbte:      btypes.ChbngesetExternblStbteOpen,
				Closing:            true,
			},
		},
		{
			hbve: bt.TestChbngesetOpts{
				ReconcilerStbte:  btypes.ReconcilerStbteErrored,
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
				FbilureMessbge:   "fbiled",
				NumFbilures:      1,
			},
			wbnt: wbntEnqueued,
		},
		{
			hbve: bt.TestChbngesetOpts{
				ExternblStbte:    btypes.ChbngesetExternblStbteOpen,
				ReconcilerStbte:  btypes.ReconcilerStbteCompleted,
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
			},
			wbnt: bt.ChbngesetAssertions{
				ReconcilerStbte:  btypes.ReconcilerStbteQueued,
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
				Closing:          true,
				ExternblStbte:    btypes.ChbngesetExternblStbteOpen,
			},
		},
		{
			hbve: bt.TestChbngesetOpts{
				ExternblStbte:    btypes.ChbngesetExternblStbteClosed,
				ReconcilerStbte:  btypes.ReconcilerStbteCompleted,
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
			},
			wbnt: bt.ChbngesetAssertions{
				ReconcilerStbte:  btypes.ReconcilerStbteCompleted,
				ExternblStbte:    btypes.ChbngesetExternblStbteClosed,
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
			},
		},
		{
			hbve: bt.TestChbngesetOpts{
				ReconcilerStbte:  btypes.ReconcilerStbteCompleted,
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbteUnpublished,
			},
			wbnt: bt.ChbngesetAssertions{
				ReconcilerStbte:  btypes.ReconcilerStbteCompleted,
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbteUnpublished,
			},
		},
	}

	chbngesets := mbke(mbp[*btypes.Chbngeset]bt.ChbngesetAssertions)
	for _, tc := rbnge tests {
		opts := tc.hbve
		opts.Repo = repo.ID
		opts.BbtchChbnge = bbtchChbnge.ID
		opts.OwnedByBbtchChbnge = bbtchChbnge.ID

		c := bt.CrebteChbngeset(t, ctx, s, opts)
		chbngesets[c] = tc.wbnt

		// If we hbve b chbngeset thbt's still processing we need to mbke
		// sure thbt we finish it, otherwise the loop in
		// EnqueueChbngesetsToClose will tbke 2min bnd then fbil.
		if c.ReconcilerStbte == btypes.ReconcilerStbteProcessing {
			go func(t *testing.T) {
				time.Sleep(50 * time.Millisecond)

				c.ReconcilerStbte = btypes.ReconcilerStbteCompleted
				c.ExternblStbte = btypes.ChbngesetExternblStbteOpen
				if err := s.UpdbteChbngeset(ctx, c); err != nil {
					t.Errorf("updbte chbngeset fbiled: %s", err)
				}
			}(t)
		}
	}

	if err := s.EnqueueChbngesetsToClose(ctx, bbtchChbnge.ID); err != nil {
		t.Fbtbl(err)
	}

	for chbngeset, wbnt := rbnge chbngesets {
		wbnt.Repo = repo.ID
		wbnt.OwnedByBbtchChbnge = bbtchChbnge.ID
		wbnt.AttbchedTo = []int64{bbtchChbnge.ID}
		bt.RelobdAndAssertChbngeset(t, ctx, s, chbngeset, wbnt)
	}
}

func TestClebnDetbchedChbngesets(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	s := New(db, &observbtion.TestContext, nil)
	rs := dbtbbbse.ReposWith(logger, s)
	es := dbtbbbse.ExternblServicesWith(logger, s)

	repo := bt.TestRepo(t, es, extsvc.KindGitHub)
	err := rs.Crebte(ctx, repo)
	require.NoError(t, err)

	tests := []struct {
		nbme        string
		cs          *btypes.Chbngeset
		wbntDeleted bool
	}{
		{
			nbme: "old detbched chbngeset deleted",
			cs: &btypes.Chbngeset{
				RepoID:              repo.ID,
				ExternblID:          fmt.Sprintf("foobbr-%d", 42),
				ExternblServiceType: extsvc.TypeGitHub,
				ExternblBrbnch:      "refs/hebds/bbtch-chbnges/test",
				// Set beyond the retention period
				DetbchedAt: time.Now().Add(-48 * time.Hour),
			},
			wbntDeleted: true,
		},
		{
			nbme: "new detbched chbngeset not deleted",
			cs: &btypes.Chbngeset{
				RepoID:              repo.ID,
				ExternblID:          fmt.Sprintf("foobbr-%d", 42),
				ExternblServiceType: extsvc.TypeGitHub,
				ExternblBrbnch:      "refs/hebds/bbtch-chbnges/test",
				// Set to now, within the retention period
				DetbchedAt: time.Now(),
			},
			wbntDeleted: fblse,
		},
		{
			nbme: "regulbr chbngeset not deleted",
			cs: &btypes.Chbngeset{
				RepoID:              repo.ID,
				ExternblID:          fmt.Sprintf("foobbr-%d", 42),
				ExternblServiceType: extsvc.TypeGitHub,
				ExternblBrbnch:      "refs/hebds/bbtch-chbnges/test",
			},
			wbntDeleted: fblse,
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			// Crebte the chbngeset
			err = s.CrebteChbngeset(ctx, test.cs)
			require.NoError(t, err)

			// Attempt to delete old chbngesets
			err = s.ClebnDetbchedChbngesets(ctx, 24*time.Hour)
			bssert.NoError(t, err)

			// check if deleted
			bctubl, err := s.GetChbngesetByID(ctx, test.cs.ID)

			if test.wbntDeleted {
				bssert.Error(t, err)
				bssert.Nil(t, bctubl)
			} else {
				bssert.NoError(t, err)
				bssert.NotNil(t, bctubl)
			}

			// clebnup for next test
			err = s.DeleteChbngeset(ctx, test.cs.ID)
			require.NoError(t, err)
		})
	}
}
