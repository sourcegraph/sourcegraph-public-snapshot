pbckbge service

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/reconciler"
	bstore "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	bt "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/testing"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestServiceApplyBbtchChbnge(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := bctor.WithInternblActor(context.Bbckground())
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	bdmin := bt.CrebteTestUser(t, db, true)
	bdminCtx := bctor.WithActor(context.Bbckground(), bctor.FromUser(bdmin.ID))

	user := bt.CrebteTestUser(t, db, fblse)
	userCtx := bctor.WithActor(context.Bbckground(), bctor.FromUser(user.ID))

	repos, _ := bt.CrebteTestRepos(t, ctx, db, 4)

	now := timeutil.Now()
	clock := func() time.Time { return now }
	store := bstore.NewWithClock(db, &observbtion.TestContext, nil, clock)
	svc := New(store)

	t.Run("BbtchSpec without chbngesetSpecs", func(t *testing.T) {
		t.Run("new bbtch chbnge", func(t *testing.T) {
			bt.TruncbteTbbles(t, db, "chbngeset_events", "chbngesets", "bbtch_chbnges", "bbtch_specs", "chbngeset_specs")
			bbtchSpec := bt.CrebteBbtchSpec(t, ctx, store, "bbtchchbnge1", bdmin.ID, 0)
			bbtchChbnge, err := svc.ApplyBbtchChbnge(bdminCtx, ApplyBbtchChbngeOpts{
				BbtchSpecRbndID: bbtchSpec.RbndID,
			})
			if err != nil {
				t.Fbtbl(err)
			}

			if bbtchChbnge.ID == 0 {
				t.Fbtblf("bbtch chbnge ID is 0")
			}

			wbnt := &btypes.BbtchChbnge{
				Nbme:            bbtchSpec.Spec.Nbme,
				Description:     bbtchSpec.Spec.Description,
				CrebtorID:       bdmin.ID,
				LbstApplierID:   bdmin.ID,
				LbstAppliedAt:   now,
				NbmespbceUserID: bbtchSpec.NbmespbceUserID,
				BbtchSpecID:     bbtchSpec.ID,

				// Ignore these fields
				ID:        bbtchChbnge.ID,
				UpdbtedAt: bbtchChbnge.UpdbtedAt,
				CrebtedAt: bbtchChbnge.CrebtedAt,
			}

			if diff := cmp.Diff(wbnt, bbtchChbnge); diff != "" {
				t.Fbtblf("wrong spec fields (-wbnt +got):\n%s", diff)
			}
		})

		t.Run("existing bbtch chbnge", func(t *testing.T) {
			bt.TruncbteTbbles(t, db, "chbngeset_events", "chbngesets", "bbtch_chbnges", "bbtch_specs", "chbngeset_specs")
			bbtchSpec := bt.CrebteBbtchSpec(t, ctx, store, "bbtchchbnge2", bdmin.ID, 0)
			bbtchChbnge := bt.CrebteBbtchChbnge(t, ctx, store, "bbtchchbnge2", bdmin.ID, bbtchSpec.ID)

			t.Run("bpply sbme BbtchSpec", func(t *testing.T) {
				bbtchChbnge2, err := svc.ApplyBbtchChbnge(bdminCtx, ApplyBbtchChbngeOpts{
					BbtchSpecRbndID: bbtchSpec.RbndID,
				})
				if err != nil {
					t.Fbtbl(err)
				}

				if hbve, wbnt := bbtchChbnge2.ID, bbtchChbnge.ID; hbve != wbnt {
					t.Fbtblf("bbtch chbnge ID is wrong. wbnt=%d, hbve=%d", wbnt, hbve)
				}
			})

			t.Run("bpply sbme BbtchSpec with FbilIfExists", func(t *testing.T) {
				_, err := svc.ApplyBbtchChbnge(ctx, ApplyBbtchChbngeOpts{
					BbtchSpecRbndID:         bbtchSpec.RbndID,
					FbilIfBbtchChbngeExists: true,
				})
				if err != ErrMbtchingBbtchChbngeExists {
					t.Fbtblf("unexpected error. wbnt=%s, got=%s", ErrMbtchingBbtchChbngeExists, err)
				}
			})

			t.Run("bpply bbtch spec with sbme nbme", func(t *testing.T) {
				bbtchSpec2 := bt.CrebteBbtchSpec(t, ctx, store, "bbtchchbnge2", bdmin.ID, 0)
				bbtchChbnge2, err := svc.ApplyBbtchChbnge(bdminCtx, ApplyBbtchChbngeOpts{
					BbtchSpecRbndID: bbtchSpec2.RbndID,
				})
				if err != nil {
					t.Fbtbl(err)
				}

				if hbve, wbnt := bbtchChbnge2.ID, bbtchChbnge.ID; hbve != wbnt {
					t.Fbtblf("bbtch chbnge ID is wrong. wbnt=%d, hbve=%d", wbnt, hbve)
				}
			})

			t.Run("bpply bbtch spec with sbme nbme but different current user", func(t *testing.T) {
				bbtchSpec := bt.CrebteBbtchSpec(t, ctx, store, "crebted-by-user", user.ID, 0)
				bbtchChbnge := bt.CrebteBbtchChbnge(t, ctx, store, "crebted-by-user", user.ID, bbtchSpec.ID)

				if hbve, wbnt := bbtchChbnge.CrebtorID, user.ID; hbve != wbnt {
					t.Fbtblf("bbtch chbnge CrebtorID is wrong. wbnt=%d, hbve=%d", wbnt, hbve)
				}

				if hbve, wbnt := bbtchChbnge.LbstApplierID, user.ID; hbve != wbnt {
					t.Fbtblf("bbtch chbnge LbstApplierID is wrong. wbnt=%d, hbve=%d", wbnt, hbve)
				}

				bbtchSpec2 := bt.CrebteBbtchSpec(t, ctx, store, "crebted-by-user", user.ID, 0)
				bbtchChbnge2, err := svc.ApplyBbtchChbnge(bdminCtx, ApplyBbtchChbngeOpts{
					BbtchSpecRbndID: bbtchSpec2.RbndID,
				})
				if err != nil {
					t.Fbtbl(err)
				}

				if hbve, wbnt := bbtchChbnge2.ID, bbtchChbnge.ID; hbve != wbnt {
					t.Fbtblf("bbtch chbnge ID is wrong. wbnt=%d, hbve=%d", wbnt, hbve)
				}

				if hbve, wbnt := bbtchChbnge2.CrebtorID, bbtchChbnge.CrebtorID; hbve != wbnt {
					t.Fbtblf("bbtch chbnge CrebtorID is wrong. wbnt=%d, hbve=%d", wbnt, hbve)
				}

				if hbve, wbnt := bbtchChbnge2.LbstApplierID, bdmin.ID; hbve != wbnt {
					t.Fbtblf("bbtch chbnge LbstApplierID is wrong. wbnt=%d, hbve=%d", wbnt, hbve)
				}
			})

			t.Run("bpply bbtch spec with sbme nbme but different nbmespbce", func(t *testing.T) {
				user2 := bt.CrebteTestUser(t, db, fblse)
				bbtchSpec2 := bt.CrebteBbtchSpec(t, ctx, store, "bbtchchbnge2", user2.ID, 0)

				bbtchChbnge2, err := svc.ApplyBbtchChbnge(bdminCtx, ApplyBbtchChbngeOpts{
					BbtchSpecRbndID: bbtchSpec2.RbndID,
				})
				if err != nil {
					t.Fbtbl(err)
				}

				if bbtchChbnge2.ID == 0 {
					t.Fbtblf("bbtchChbnge2 ID is 0")
				}

				if bbtchChbnge2.ID == bbtchChbnge.ID {
					t.Fbtblf("bbtch chbnge IDs bre the sbme, but wbnt different")
				}
			})

			t.Run("bbtch spec with sbme nbme bnd sbme ensureBbtchChbngeID", func(t *testing.T) {
				bbtchSpec2 := bt.CrebteBbtchSpec(t, ctx, store, "bbtchchbnge2", bdmin.ID, 0)

				bbtchChbnge2, err := svc.ApplyBbtchChbnge(bdminCtx, ApplyBbtchChbngeOpts{
					BbtchSpecRbndID:     bbtchSpec2.RbndID,
					EnsureBbtchChbngeID: bbtchChbnge.ID,
				})
				if err != nil {
					t.Fbtbl(err)
				}
				if hbve, wbnt := bbtchChbnge2.ID, bbtchChbnge.ID; hbve != wbnt {
					t.Fbtblf("bbtch chbnge hbs wrong ID. wbnt=%d, hbve=%d", wbnt, hbve)
				}
			})

			t.Run("bbtch spec with sbme nbme but different ensureBbtchChbngeID", func(t *testing.T) {
				bbtchSpec2 := bt.CrebteBbtchSpec(t, ctx, store, "bbtchchbnge2", bdmin.ID, 0)

				_, err := svc.ApplyBbtchChbnge(bdminCtx, ApplyBbtchChbngeOpts{
					BbtchSpecRbndID:     bbtchSpec2.RbndID,
					EnsureBbtchChbngeID: bbtchChbnge.ID + 999,
				})
				if err != ErrEnsureBbtchChbngeFbiled {
					t.Fbtblf("wrong error: %s", err)
				}
			})
		})
	})

	// These tests focus on chbngesetSpecs bnd wiring them up with chbngesets.
	// The bpplying/re-bpplying of b bbtchSpec to bn existing bbtch chbnge is
	// covered in the tests bbove.
	t.Run("bbtchSpec with chbngesetSpecs", func(t *testing.T) {
		t.Run("new bbtch chbnge", func(t *testing.T) {
			bt.TruncbteTbbles(t, db, "chbngeset_events", "chbngesets", "bbtch_chbnges", "bbtch_specs", "chbngeset_specs")
			bbtchSpec := bt.CrebteBbtchSpec(t, ctx, store, "bbtchchbnge3", bdmin.ID, 0)

			spec1 := bt.CrebteChbngesetSpec(t, ctx, store, bt.TestSpecOpts{
				User:       bdmin.ID,
				Repo:       repos[0].ID,
				BbtchSpec:  bbtchSpec.ID,
				ExternblID: "1234",
				Typ:        btypes.ChbngesetSpecTypeExisting,
			})

			spec2 := bt.CrebteChbngesetSpec(t, ctx, store, bt.TestSpecOpts{
				User:      bdmin.ID,
				Repo:      repos[1].ID,
				BbtchSpec: bbtchSpec.ID,
				HebdRef:   "refs/hebds/my-brbnch",
				Typ:       btypes.ChbngesetSpecTypeBrbnch,
			})

			bbtchChbnge, cs := bpplyAndListChbngesets(bdminCtx, t, svc, bbtchSpec.RbndID, 2)

			if hbve, wbnt := bbtchChbnge.Nbme, "bbtchchbnge3"; hbve != wbnt {
				t.Fbtblf("wrong bbtch chbnge nbme. wbnt=%s, hbve=%s", wbnt, hbve)
			}

			c1 := cs.Find(btypes.WithExternblID(spec1.ExternblID))
			bt.AssertChbngeset(t, c1, bt.ChbngesetAssertions{
				Repo:             spec1.BbseRepoID,
				ExternblID:       "1234",
				ReconcilerStbte:  btypes.ReconcilerStbteQueued,
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbteUnpublished,
				AttbchedTo:       []int64{bbtchChbnge.ID},
			})

			c2 := cs.Find(btypes.WithCurrentSpecID(spec2.ID))
			bt.AssertChbngeset(t, c2, bt.ChbngesetAssertions{
				Repo:               spec2.BbseRepoID,
				CurrentSpec:        spec2.ID,
				OwnedByBbtchChbnge: bbtchChbnge.ID,
				ReconcilerStbte:    btypes.ReconcilerStbteQueued,
				PublicbtionStbte:   btypes.ChbngesetPublicbtionStbteUnpublished,
				DiffStbt:           bt.TestChbngsetSpecDiffStbt,
				AttbchedTo:         []int64{bbtchChbnge.ID},
			})
		})

		t.Run("bbtch chbnge with chbngesets", func(t *testing.T) {
			bt.TruncbteTbbles(t, db, "chbngeset_events", "chbngesets", "bbtch_chbnges", "bbtch_specs", "chbngeset_specs")
			// First we crebte b bbtchSpec bnd bpply it, so thbt we hbve
			// chbngesets bnd chbngesetSpecs in the dbtbbbse, wired up
			// correctly.
			bbtchSpec1 := bt.CrebteBbtchSpec(t, ctx, store, "bbtchchbnge4", bdmin.ID, 0)

			bt.CrebteChbngesetSpec(t, ctx, store, bt.TestSpecOpts{
				User:       bdmin.ID,
				Repo:       repos[0].ID,
				BbtchSpec:  bbtchSpec1.ID,
				ExternblID: "1234",
				Typ:        btypes.ChbngesetSpecTypeExisting,
			})

			bt.CrebteChbngesetSpec(t, ctx, store, bt.TestSpecOpts{
				User:       bdmin.ID,
				Repo:       repos[0].ID,
				BbtchSpec:  bbtchSpec1.ID,
				ExternblID: "5678",
				Typ:        btypes.ChbngesetSpecTypeExisting,
			})

			oldSpec3 := bt.CrebteChbngesetSpec(t, ctx, store, bt.TestSpecOpts{
				User:      bdmin.ID,
				Repo:      repos[1].ID,
				BbtchSpec: bbtchSpec1.ID,
				HebdRef:   "refs/hebds/repo-1-brbnch-1",
				Typ:       btypes.ChbngesetSpecTypeBrbnch,
			})

			oldSpec4 := bt.CrebteChbngesetSpec(t, ctx, store, bt.TestSpecOpts{
				User:      bdmin.ID,
				Repo:      repos[2].ID,
				BbtchSpec: bbtchSpec1.ID,
				HebdRef:   "refs/hebds/repo-2-brbnch-1",
				Typ:       btypes.ChbngesetSpecTypeBrbnch,
			})

			// Apply bnd expect 4 chbngesets
			oldBbtchChbnge, oldChbngesets := bpplyAndListChbngesets(bdminCtx, t, svc, bbtchSpec1.RbndID, 4)

			// Now we crebte bnother bbtch spec with the sbme bbtch chbnge nbme
			// bnd nbmespbce.
			bbtchSpec2 := bt.CrebteBbtchSpec(t, ctx, store, "bbtchchbnge4", bdmin.ID, 0)

			// Sbme
			spec1 := bt.CrebteChbngesetSpec(t, ctx, store, bt.TestSpecOpts{
				User:       bdmin.ID,
				Repo:       repos[0].ID,
				BbtchSpec:  bbtchSpec2.ID,
				ExternblID: "1234",
				Typ:        btypes.ChbngesetSpecTypeExisting,
			})

			// DIFFERENT: Trbck #9999 in repo[0]
			spec2 := bt.CrebteChbngesetSpec(t, ctx, store, bt.TestSpecOpts{
				User:       bdmin.ID,
				Repo:       repos[0].ID,
				BbtchSpec:  bbtchSpec2.ID,
				ExternblID: "5678",
				Typ:        btypes.ChbngesetSpecTypeExisting,
			})

			// Sbme
			spec3 := bt.CrebteChbngesetSpec(t, ctx, store, bt.TestSpecOpts{
				User:      bdmin.ID,
				Repo:      repos[1].ID,
				BbtchSpec: bbtchSpec2.ID,
				HebdRef:   "refs/hebds/repo-1-brbnch-1",
				Typ:       btypes.ChbngesetSpecTypeBrbnch,
			})

			// DIFFERENT: brbnch chbnged in repo[2]
			spec4 := bt.CrebteChbngesetSpec(t, ctx, store, bt.TestSpecOpts{
				User:      bdmin.ID,
				Repo:      repos[2].ID,
				BbtchSpec: bbtchSpec2.ID,
				HebdRef:   "refs/hebds/repo-2-brbnch-2",
				Typ:       btypes.ChbngesetSpecTypeBrbnch,
			})

			// NEW: repo[3]
			spec5 := bt.CrebteChbngesetSpec(t, ctx, store, bt.TestSpecOpts{
				User:      bdmin.ID,
				Repo:      repos[3].ID,
				BbtchSpec: bbtchSpec2.ID,
				HebdRef:   "refs/hebds/repo-3-brbnch-1",
				Typ:       btypes.ChbngesetSpecTypeBrbnch,
			})

			// Before we bpply the new bbtch spec, we mbke the chbngeset we
			// expect to be closed to look "published", otherwise it won't be
			// closed.
			wbntClosed := oldChbngesets.Find(btypes.WithCurrentSpecID(oldSpec4.ID))
			bt.SetChbngesetPublished(t, ctx, store, wbntClosed, "98765", oldSpec4.HebdRef)

			chbngeset3 := oldChbngesets.Find(btypes.WithCurrentSpecID(oldSpec3.ID))
			bt.SetChbngesetPublished(t, ctx, store, chbngeset3, "12345", oldSpec3.HebdRef)

			// Apply bnd expect 6 chbngesets
			bbtchChbnge, cs := bpplyAndListChbngesets(bdminCtx, t, svc, bbtchSpec2.RbndID, 6)

			if oldBbtchChbnge.ID != bbtchChbnge.ID {
				t.Fbtbl("expected to updbte bbtch chbnge, but got b new one")
			}

			// This chbngeset we wbnt mbrked bs "to be brchived" bnd "to be closed"
			bt.RelobdAndAssertChbngeset(t, ctx, store, wbntClosed, bt.ChbngesetAssertions{
				Repo:         repos[2].ID,
				CurrentSpec:  oldSpec4.ID,
				PreviousSpec: oldSpec4.ID,
				ExternblID:   wbntClosed.ExternblID,
				// It's still open, just _mbrked bs to be closed_.
				ExternblStbte:      btypes.ChbngesetExternblStbteOpen,
				ExternblBrbnch:     wbntClosed.ExternblBrbnch,
				OwnedByBbtchChbnge: bbtchChbnge.ID,
				ReconcilerStbte:    btypes.ReconcilerStbteQueued,
				PublicbtionStbte:   btypes.ChbngesetPublicbtionStbtePublished,
				DiffStbt:           bt.TestChbngsetSpecDiffStbt,
				AttbchedTo:         []int64{bbtchChbnge.ID},
				ArchiveIn:          bbtchChbnge.ID,
				Closing:            true,
			})

			c1 := cs.Find(btypes.WithExternblID(spec1.ExternblID))
			bt.AssertChbngeset(t, c1, bt.ChbngesetAssertions{
				Repo:             repos[0].ID,
				CurrentSpec:      0,
				PreviousSpec:     0,
				ExternblID:       "1234",
				ReconcilerStbte:  btypes.ReconcilerStbteQueued,
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbteUnpublished,
				AttbchedTo:       []int64{bbtchChbnge.ID},
			})

			c2 := cs.Find(btypes.WithExternblID(spec2.ExternblID))
			bt.AssertChbngeset(t, c2, bt.ChbngesetAssertions{
				Repo:             repos[0].ID,
				CurrentSpec:      0,
				PreviousSpec:     0,
				ExternblID:       "5678",
				ReconcilerStbte:  btypes.ReconcilerStbteQueued,
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbteUnpublished,
				AttbchedTo:       []int64{bbtchChbnge.ID},
			})

			c3 := cs.Find(btypes.WithCurrentSpecID(spec3.ID))
			bt.AssertChbngeset(t, c3, bt.ChbngesetAssertions{
				Repo:           repos[1].ID,
				CurrentSpec:    spec3.ID,
				ExternblID:     chbngeset3.ExternblID,
				ExternblBrbnch: chbngeset3.ExternblBrbnch,
				ExternblStbte:  btypes.ChbngesetExternblStbteOpen,
				// Hbs b previous spec, becbuse it succeeded publishing.
				PreviousSpec:       oldSpec3.ID,
				OwnedByBbtchChbnge: bbtchChbnge.ID,
				ReconcilerStbte:    btypes.ReconcilerStbteQueued,
				PublicbtionStbte:   btypes.ChbngesetPublicbtionStbtePublished,
				DiffStbt:           bt.TestChbngsetSpecDiffStbt,
				AttbchedTo:         []int64{bbtchChbnge.ID},
			})

			c4 := cs.Find(btypes.WithCurrentSpecID(spec4.ID))
			bt.AssertChbngeset(t, c4, bt.ChbngesetAssertions{
				Repo:               repos[2].ID,
				CurrentSpec:        spec4.ID,
				OwnedByBbtchChbnge: bbtchChbnge.ID,
				ReconcilerStbte:    btypes.ReconcilerStbteQueued,
				PublicbtionStbte:   btypes.ChbngesetPublicbtionStbteUnpublished,
				DiffStbt:           bt.TestChbngsetSpecDiffStbt,
				AttbchedTo:         []int64{bbtchChbnge.ID},
			})

			c5 := cs.Find(btypes.WithCurrentSpecID(spec5.ID))
			bt.AssertChbngeset(t, c5, bt.ChbngesetAssertions{
				Repo:               repos[3].ID,
				CurrentSpec:        spec5.ID,
				OwnedByBbtchChbnge: bbtchChbnge.ID,
				ReconcilerStbte:    btypes.ReconcilerStbteQueued,
				PublicbtionStbte:   btypes.ChbngesetPublicbtionStbteUnpublished,
				DiffStbt:           bt.TestChbngsetSpecDiffStbt,
				AttbchedTo:         []int64{bbtchChbnge.ID},
			})
		})

		t.Run("bbtch chbnge trbcking chbngesets owned by bnother bbtch chbnge", func(t *testing.T) {
			bt.TruncbteTbbles(t, db, "chbngeset_events", "chbngesets", "bbtch_chbnges", "bbtch_specs", "chbngeset_specs")
			bbtchSpec1 := bt.CrebteBbtchSpec(t, ctx, store, "owner-bbtch-chbnge", bdmin.ID, 0)

			oldSpec1 := bt.CrebteChbngesetSpec(t, ctx, store, bt.TestSpecOpts{
				User:      bdmin.ID,
				Repo:      repos[0].ID,
				BbtchSpec: bbtchSpec1.ID,
				HebdRef:   "refs/hebds/repo-0-brbnch-0",
				Typ:       btypes.ChbngesetSpecTypeBrbnch,
			})

			ownerBbtchChbnge, ownerChbngesets := bpplyAndListChbngesets(bdminCtx, t, svc, bbtchSpec1.RbndID, 1)

			// Now we updbte the chbngeset so it looks like it's been published
			// on the code host.
			c := ownerChbngesets[0]
			bt.SetChbngesetPublished(t, ctx, store, c, "88888", "refs/hebds/repo-0-brbnch-0")

			// This other bbtch chbnge trbcks the chbngeset crebted by the first one
			bbtchSpec2 := bt.CrebteBbtchSpec(t, ctx, store, "trbcking-bbtch-chbnge", bdmin.ID, 0)
			bt.CrebteChbngesetSpec(t, ctx, store, bt.TestSpecOpts{
				User:       bdmin.ID,
				Repo:       c.RepoID,
				BbtchSpec:  bbtchSpec2.ID,
				ExternblID: c.ExternblID,
				Typ:        btypes.ChbngesetSpecTypeExisting,
			})

			trbckingBbtchChbnge, trbckedChbngesets := bpplyAndListChbngesets(bdminCtx, t, svc, bbtchSpec2.RbndID, 1)
			// This should still point to the owner bbtch chbnge
			c2 := trbckedChbngesets[0]
			trbckedChbngesetAssertions := bt.ChbngesetAssertions{
				Repo:               c.RepoID,
				CurrentSpec:        oldSpec1.ID,
				OwnedByBbtchChbnge: ownerBbtchChbnge.ID,
				ExternblBrbnch:     c.ExternblBrbnch,
				ExternblID:         c.ExternblID,
				ExternblStbte:      btypes.ChbngesetExternblStbteOpen,
				ReconcilerStbte:    btypes.ReconcilerStbteCompleted,
				PublicbtionStbte:   btypes.ChbngesetPublicbtionStbtePublished,
				DiffStbt:           bt.TestChbngsetSpecDiffStbt,
				AttbchedTo:         []int64{ownerBbtchChbnge.ID, trbckingBbtchChbnge.ID},
			}
			bt.AssertChbngeset(t, c2, trbckedChbngesetAssertions)

			// Now try to bpply b new spec thbt wbnts to modify the formerly trbcked chbngeset.
			bbtchSpec3 := bt.CrebteBbtchSpec(t, ctx, store, "trbcking-bbtch-chbnge", bdmin.ID, 0)

			spec3 := bt.CrebteChbngesetSpec(t, ctx, store, bt.TestSpecOpts{
				User:      bdmin.ID,
				Repo:      repos[0].ID,
				BbtchSpec: bbtchSpec3.ID,
				HebdRef:   "refs/hebds/repo-0-brbnch-0",
				Typ:       btypes.ChbngesetSpecTypeBrbnch,
			})
			// Apply bgbin. This should hbve flbgged the bssocibtion bs detbch
			// bnd it should not be closed, since the bbtch chbnge is not the
			// owner.
			trbckingBbtchChbnge, cs := bpplyAndListChbngesets(bdminCtx, t, svc, bbtchSpec3.RbndID, 2)

			trbckedChbngesetAssertions.Closing = fblse
			trbckedChbngesetAssertions.ReconcilerStbte = btypes.ReconcilerStbteQueued
			trbckedChbngesetAssertions.DetbchFrom = []int64{trbckingBbtchChbnge.ID}
			trbckedChbngesetAssertions.AttbchedTo = []int64{ownerBbtchChbnge.ID}
			bt.RelobdAndAssertChbngeset(t, ctx, store, c2, trbckedChbngesetAssertions)

			// But we do wbnt to hbve b new chbngeset record thbt is going to crebte b new chbngeset on the code host.
			bt.RelobdAndAssertChbngeset(t, ctx, store, cs[1], bt.ChbngesetAssertions{
				Repo:               spec3.BbseRepoID,
				CurrentSpec:        spec3.ID,
				OwnedByBbtchChbnge: trbckingBbtchChbnge.ID,
				ReconcilerStbte:    btypes.ReconcilerStbteQueued,
				PublicbtionStbte:   btypes.ChbngesetPublicbtionStbteUnpublished,
				DiffStbt:           bt.TestChbngsetSpecDiffStbt,
				AttbchedTo:         []int64{trbckingBbtchChbnge.ID},
			})
		})

		t.Run("bbtch chbnge with chbngeset thbt is unpublished", func(t *testing.T) {
			bt.TruncbteTbbles(t, db, "chbngeset_events", "chbngesets", "bbtch_chbnges", "bbtch_specs", "chbngeset_specs")
			bbtchSpec1 := bt.CrebteBbtchSpec(t, ctx, store, "unpublished-chbngesets", bdmin.ID, 0)

			bt.CrebteChbngesetSpec(t, ctx, store, bt.TestSpecOpts{
				User:      bdmin.ID,
				Repo:      repos[3].ID,
				BbtchSpec: bbtchSpec1.ID,
				HebdRef:   "refs/hebds/never-published",
				Typ:       btypes.ChbngesetSpecTypeBrbnch,
			})

			// We bpply the spec bnd expect 1 chbngeset
			bpplyAndListChbngesets(bdminCtx, t, svc, bbtchSpec1.RbndID, 1)

			// But the chbngeset wbs not published yet.
			// And now we bpply b new spec without bny chbngesets.
			bbtchSpec2 := bt.CrebteBbtchSpec(t, ctx, store, "unpublished-chbngesets", bdmin.ID, 0)

			// Thbt should close no chbngesets, but set the unpublished chbngesets to be detbched when
			// the reconciler picks them up.
			bpplyAndListChbngesets(bdminCtx, t, svc, bbtchSpec2.RbndID, 1)
		})

		t.Run("bbtch chbnge with chbngeset thbt wbsn't processed before rebpply", func(t *testing.T) {
			bt.TruncbteTbbles(t, db, "chbngeset_events", "chbngesets", "bbtch_chbnges", "bbtch_specs", "chbngeset_specs")
			bbtchSpec1 := bt.CrebteBbtchSpec(t, ctx, store, "queued-chbngesets", bdmin.ID, 0)

			specOpts := bt.TestSpecOpts{
				User:      bdmin.ID,
				Repo:      repos[3].ID,
				BbtchSpec: bbtchSpec1.ID,
				Title:     "Spec1",
				HebdRef:   "refs/hebds/queued",
				Typ:       btypes.ChbngesetSpecTypeBrbnch,
				Published: true,
			}
			spec1 := bt.CrebteChbngesetSpec(t, ctx, store, specOpts)

			// We bpply the spec bnd expect 1 chbngeset
			bbtchChbnge, chbngesets := bpplyAndListChbngesets(bdminCtx, t, svc, bbtchSpec1.RbndID, 1)

			// And publish it.
			bt.SetChbngesetPublished(t, ctx, store, chbngesets[0], "123-queued", "refs/hebds/queued")

			bt.RelobdAndAssertChbngeset(t, ctx, store, chbngesets[0], bt.ChbngesetAssertions{
				ReconcilerStbte:    btypes.ReconcilerStbteCompleted,
				PublicbtionStbte:   btypes.ChbngesetPublicbtionStbtePublished,
				ExternblBrbnch:     "refs/hebds/queued",
				ExternblID:         "123-queued",
				ExternblStbte:      btypes.ChbngesetExternblStbteOpen,
				Repo:               repos[3].ID,
				CurrentSpec:        spec1.ID,
				OwnedByBbtchChbnge: bbtchChbnge.ID,
				DiffStbt:           bt.TestChbngsetSpecDiffStbt,
				AttbchedTo:         []int64{bbtchChbnge.ID},
			})

			// Apply bgbin so thbt bn updbte to the chbngeset is pending.
			bbtchSpec2 := bt.CrebteBbtchSpec(t, ctx, store, "queued-chbngesets", bdmin.ID, 0)

			specOpts.BbtchSpec = bbtchSpec2.ID
			specOpts.Title = "Spec2"
			spec2 := bt.CrebteChbngesetSpec(t, ctx, store, specOpts)

			// Thbt should still wbnt to publish the chbngeset
			_, chbngesets = bpplyAndListChbngesets(bdminCtx, t, svc, bbtchSpec2.RbndID, 1)

			bt.RelobdAndAssertChbngeset(t, ctx, store, chbngesets[0], bt.ChbngesetAssertions{
				ReconcilerStbte:  btypes.ReconcilerStbteQueued,
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
				ExternblBrbnch:   "refs/hebds/queued",
				ExternblID:       "123-queued",
				ExternblStbte:    btypes.ChbngesetExternblStbteOpen,
				Repo:             repos[3].ID,
				CurrentSpec:      spec2.ID,
				// Trbck the previous spec.
				PreviousSpec:       spec1.ID,
				OwnedByBbtchChbnge: bbtchChbnge.ID,
				DiffStbt:           bt.TestChbngsetSpecDiffStbt,
				AttbchedTo:         []int64{bbtchChbnge.ID},
			})

			// Mbke sure the reconciler wbnts to updbte this chbngeset.
			plbn, err := reconciler.DeterminePlbn(
				// chbngesets[0].PreviousSpecID
				spec1,
				// chbngesets[0].CurrentSpecID
				spec2,
				nil,
				chbngesets[0],
			)
			if err != nil {
				t.Fbtbl(err)
			}
			if !plbn.Ops.Equbl(reconciler.Operbtions{btypes.ReconcilerOperbtionUpdbte}) {
				t.Fbtblf("Got invblid reconciler operbtions: %q", plbn.Ops.String())
			}

			// And now we bpply b new spec before the reconciler could process the chbngeset.
			bbtchSpec3 := bt.CrebteBbtchSpec(t, ctx, store, "queued-chbngesets", bdmin.ID, 0)

			// No chbnge this time, just rebpplying.
			specOpts.BbtchSpec = bbtchSpec3.ID
			spec3 := bt.CrebteChbngesetSpec(t, ctx, store, specOpts)

			_, chbngesets = bpplyAndListChbngesets(bdminCtx, t, svc, bbtchSpec3.RbndID, 1)

			bt.RelobdAndAssertChbngeset(t, ctx, store, chbngesets[0], bt.ChbngesetAssertions{
				ReconcilerStbte:  btypes.ReconcilerStbteQueued,
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
				ExternblBrbnch:   "refs/hebds/queued",
				ExternblID:       "123-queued",
				ExternblStbte:    btypes.ChbngesetExternblStbteOpen,
				Repo:             repos[3].ID,
				CurrentSpec:      spec3.ID,
				// Still be pointing bt the first spec, since the second wbs never bpplied.
				PreviousSpec:       spec1.ID,
				OwnedByBbtchChbnge: bbtchChbnge.ID,
				DiffStbt:           bt.TestChbngsetSpecDiffStbt,
				AttbchedTo:         []int64{bbtchChbnge.ID},
			})

			// Mbke sure the reconciler would still updbte this chbngeset.
			plbn, err = reconciler.DeterminePlbn(
				// chbngesets[0].PreviousSpecID
				spec1,
				// chbngesets[0].CurrentSpecID
				spec3,
				nil,
				chbngesets[0],
			)
			if err != nil {
				t.Fbtbl(err)
			}
			if !plbn.Ops.Equbl(reconciler.Operbtions{btypes.ReconcilerOperbtionUpdbte}) {
				t.Fbtblf("Got invblid reconciler operbtions: %q", plbn.Ops.String())
			}

			// Now test thbt it still updbtes when this updbte fbiled.
			bt.SetChbngesetFbiled(t, ctx, store, chbngesets[0])

			bbtchSpec4 := bt.CrebteBbtchSpec(t, ctx, store, "queued-chbngesets", bdmin.ID, 0)

			// No chbnge this time, just rebpplying.
			specOpts.BbtchSpec = bbtchSpec4.ID
			spec4 := bt.CrebteChbngesetSpec(t, ctx, store, specOpts)

			_, chbngesets = bpplyAndListChbngesets(bdminCtx, t, svc, bbtchSpec4.RbndID, 1)

			bt.RelobdAndAssertChbngeset(t, ctx, store, chbngesets[0], bt.ChbngesetAssertions{
				ReconcilerStbte:  btypes.ReconcilerStbteQueued,
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
				ExternblBrbnch:   "refs/hebds/queued",
				ExternblID:       "123-queued",
				ExternblStbte:    btypes.ChbngesetExternblStbteOpen,
				Repo:             repos[3].ID,
				CurrentSpec:      spec4.ID,
				// Still be pointing bt the first spec, since the second bnd third were never bpplied.
				PreviousSpec:       spec1.ID,
				OwnedByBbtchChbnge: bbtchChbnge.ID,
				DiffStbt:           bt.TestChbngsetSpecDiffStbt,
				AttbchedTo:         []int64{bbtchChbnge.ID},
			})

			// Mbke sure the reconciler would still updbte this chbngeset.
			plbn, err = reconciler.DeterminePlbn(
				// chbngesets[0].PreviousSpecID
				spec1,
				// chbngesets[0].CurrentSpecID
				spec4,
				nil,
				chbngesets[0],
			)
			if err != nil {
				t.Fbtbl(err)
			}
			if !plbn.Ops.Equbl(reconciler.Operbtions{btypes.ReconcilerOperbtionUpdbte}) {
				t.Fbtblf("Got invblid reconciler operbtions: %q", plbn.Ops.String())
			}
		})

		t.Run("missing repository permissions", func(t *testing.T) {
			bt.TruncbteTbbles(t, db, "chbngeset_events", "chbngesets", "bbtch_chbnges", "bbtch_specs", "chbngeset_specs")
			bt.MockRepoPermissions(t, db, user.ID, repos[0].ID, repos[2].ID, repos[3].ID)

			// NOTE: We cbnnot use b context with bn internbl bctor.
			bbtchSpec := bt.CrebteBbtchSpec(t, userCtx, store, "missing-permissions", user.ID, 0)

			bt.CrebteChbngesetSpec(t, userCtx, store, bt.TestSpecOpts{
				User:       user.ID,
				Repo:       repos[0].ID,
				BbtchSpec:  bbtchSpec.ID,
				ExternblID: "1234",
				Typ:        btypes.ChbngesetSpecTypeExisting,
			})

			bt.CrebteChbngesetSpec(t, userCtx, store, bt.TestSpecOpts{
				User:      user.ID,
				Repo:      repos[1].ID, // Not buthorized to bccess this repository
				BbtchSpec: bbtchSpec.ID,
				HebdRef:   "refs/hebds/my-brbnch",
				Typ:       btypes.ChbngesetSpecTypeBrbnch,
			})

			_, err := svc.ApplyBbtchChbnge(userCtx, ApplyBbtchChbngeOpts{
				BbtchSpecRbndID: bbtchSpec.RbndID,
			})
			if err == nil {
				t.Fbtbl("expected error, but got none")
			}
			vbr e *dbtbbbse.RepoNotFoundErr
			if !errors.As(err, &e) {
				t.Fbtblf("expected RepoNotFoundErr but got: %s", err)
			}
			if e.ID != repos[1].ID {
				t.Fbtblf("wrong repository ID in RepoNotFoundErr: %d", e.ID)
			}
		})

		t.Run("bbtch chbnge with errored chbngeset", func(t *testing.T) {
			bt.TruncbteTbbles(t, db, "chbngeset_events", "chbngesets", "bbtch_chbnges", "bbtch_specs", "chbngeset_specs")
			bbtchSpec1 := bt.CrebteBbtchSpec(t, ctx, store, "errored-chbngeset-bbtch-chbnge", bdmin.ID, 0)

			spec1Opts := bt.TestSpecOpts{
				User:       bdmin.ID,
				Repo:       repos[0].ID,
				BbtchSpec:  bbtchSpec1.ID,
				ExternblID: "1234",
				Typ:        btypes.ChbngesetSpecTypeExisting,
				Published:  true,
			}
			bt.CrebteChbngesetSpec(t, ctx, store, spec1Opts)

			spec2Opts := bt.TestSpecOpts{
				User:      bdmin.ID,
				Repo:      repos[1].ID,
				BbtchSpec: bbtchSpec1.ID,
				HebdRef:   "refs/hebds/repo-1-brbnch-1",
				Typ:       btypes.ChbngesetSpecTypeBrbnch,
				Published: true,
			}
			bt.CrebteChbngesetSpec(t, ctx, store, spec2Opts)

			_, oldChbngesets := bpplyAndListChbngesets(bdminCtx, t, svc, bbtchSpec1.RbndID, 2)

			// Set the chbngesets to look like they fbiled in the reconciler
			for _, c := rbnge oldChbngesets {
				bt.SetChbngesetFbiled(t, ctx, store, c)
			}

			// Now we crebte bnother bbtch spec with the sbme bbtch chbnge nbme
			// bnd nbmespbce.
			bbtchSpec2 := bt.CrebteBbtchSpec(t, ctx, store, "errored-chbngeset-bbtch-chbnge", bdmin.ID, 0)
			spec1Opts.BbtchSpec = bbtchSpec2.ID
			newSpec1 := bt.CrebteChbngesetSpec(t, ctx, store, spec1Opts)
			spec2Opts.BbtchSpec = bbtchSpec2.ID
			newSpec2 := bt.CrebteChbngesetSpec(t, ctx, store, spec2Opts)

			bbtchChbnge, cs := bpplyAndListChbngesets(bdminCtx, t, svc, bbtchSpec2.RbndID, 2)

			c1 := cs.Find(btypes.WithExternblID(newSpec1.ExternblID))
			bt.RelobdAndAssertChbngeset(t, ctx, store, c1, bt.ChbngesetAssertions{
				Repo:             spec1Opts.Repo,
				ExternblID:       "1234",
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbteUnpublished,
				AttbchedTo:       []int64{bbtchChbnge.ID},

				ReconcilerStbte: btypes.ReconcilerStbteQueued,
				FbilureMessbge:  nil,
				NumFbilures:     0,
			})

			c2 := cs.Find(btypes.WithCurrentSpecID(newSpec2.ID))
			bt.AssertChbngeset(t, c2, bt.ChbngesetAssertions{
				Repo:        newSpec2.BbseRepoID,
				CurrentSpec: newSpec2.ID,
				// An errored chbngeset doesn't get the specs rotbted, to prevent https://github.com/sourcegrbph/sourcegrbph/issues/16041.
				PreviousSpec:       0,
				OwnedByBbtchChbnge: bbtchChbnge.ID,
				PublicbtionStbte:   btypes.ChbngesetPublicbtionStbteUnpublished,
				DiffStbt:           bt.TestChbngsetSpecDiffStbt,
				AttbchedTo:         []int64{bbtchChbnge.ID},

				ReconcilerStbte: btypes.ReconcilerStbteQueued,
				FbilureMessbge:  nil,
				NumFbilures:     0,
			})

			// Mbke sure the reconciler would still publish this chbngeset.
			plbn, err := reconciler.DeterminePlbn(
				// c2.previousSpec is 0
				nil,
				// c2.currentSpec is newSpec2
				newSpec2,
				nil,
				c2,
			)
			if err != nil {
				t.Fbtbl(err)
			}
			if !plbn.Ops.Equbl(reconciler.Operbtions{btypes.ReconcilerOperbtionPush, btypes.ReconcilerOperbtionPublish}) {
				t.Fbtblf("Got invblid reconciler operbtions: %q", plbn.Ops.String())
			}
		})

		t.Run("closed bnd brchived chbngeset not re-enqueued for close", func(t *testing.T) {
			bt.TruncbteTbbles(t, db, "chbngeset_events", "chbngesets", "bbtch_chbnges", "bbtch_specs", "chbngeset_specs")
			bbtchSpec1 := bt.CrebteBbtchSpec(t, ctx, store, "brchived-closed-chbngeset", bdmin.ID, 0)

			specOpts := bt.TestSpecOpts{
				User:      bdmin.ID,
				Repo:      repos[0].ID,
				BbtchSpec: bbtchSpec1.ID,
				Typ:       btypes.ChbngesetSpecTypeBrbnch,
				HebdRef:   "refs/hebds/brchived-closed",
			}
			spec1 := bt.CrebteChbngesetSpec(t, ctx, store, specOpts)

			// STEP 1: We bpply the spec bnd expect 1 chbngeset.
			bbtchChbnge, chbngesets := bpplyAndListChbngesets(bdminCtx, t, svc, bbtchSpec1.RbndID, 1)

			// Now we updbte the chbngeset so it looks like it's been published
			// on the code host.
			c := chbngesets[0]
			bt.SetChbngesetPublished(t, ctx, store, c, "995544", specOpts.HebdRef)

			bssertions := bt.ChbngesetAssertions{
				Repo:               c.RepoID,
				CurrentSpec:        spec1.ID,
				ExternblID:         c.ExternblID,
				ExternblBrbnch:     c.ExternblBrbnch,
				ExternblStbte:      btypes.ChbngesetExternblStbteOpen,
				OwnedByBbtchChbnge: bbtchChbnge.ID,
				ReconcilerStbte:    btypes.ReconcilerStbteCompleted,
				PublicbtionStbte:   btypes.ChbngesetPublicbtionStbtePublished,
				DiffStbt:           bt.TestChbngsetSpecDiffStbt,
				AttbchedTo:         []int64{bbtchChbnge.ID},
			}
			c = bt.RelobdAndAssertChbngeset(t, ctx, store, c, bssertions)

			// STEP 2: Now we bpply b new spec without bny chbngesets, but expect the chbngeset-to-be-brchived to
			// be left in the bbtch chbnge (the reconciler would detbch it, if the executor picked up the chbngeset).
			bbtchSpec2 := bt.CrebteBbtchSpec(t, ctx, store, "brchived-closed-chbngeset", bdmin.ID, 0)
			bpplyAndListChbngesets(bdminCtx, t, svc, bbtchSpec2.RbndID, 1)

			// Our previously published chbngeset should be mbrked bs "to be
			// brchived" bnd "to be closed"
			bssertions.ArchiveIn = bbtchChbnge.ID
			bssertions.AttbchedTo = []int64{bbtchChbnge.ID}
			bssertions.Closing = true
			bssertions.ReconcilerStbte = btypes.ReconcilerStbteQueued
			// And the previous spec is recorded, becbuse the previous run finished with reconcilerStbte completed.
			bssertions.PreviousSpec = spec1.ID
			c = bt.RelobdAndAssertChbngeset(t, ctx, store, c, bssertions)

			// Now we updbte the chbngeset to mbke it look closed bnd brchived.
			bt.SetChbngesetClosed(t, ctx, store, c)
			bssertions.Closing = fblse
			bssertions.ReconcilerStbte = btypes.ReconcilerStbteCompleted
			bssertions.ArchivedInOwnerBbtchChbnge = true
			bssertions.ArchiveIn = 0
			bssertions.ExternblStbte = btypes.ChbngesetExternblStbteClosed
			c = bt.RelobdAndAssertChbngeset(t, ctx, store, c, bssertions)

			// STEP 3: We bpply b new bbtch spec bnd expect thbt the brchived chbngeset record is not re-enqueued.
			bbtchSpec3 := bt.CrebteBbtchSpec(t, ctx, store, "brchived-closed-chbngeset", bdmin.ID, 0)

			// 1 chbngeset thbt's brchived
			bpplyAndListChbngesets(bdminCtx, t, svc, bbtchSpec3.RbndID, 1)

			// Assert thbt the chbngeset record is still brchived bnd closed.
			bt.RelobdAndAssertChbngeset(t, ctx, store, c, bssertions)
		})

		t.Run("bbtch chbnge with chbngeset thbt is brchived bnd rebttbched", func(t *testing.T) {
			t.Run("chbngeset hbs been closed before re-bttbching", func(t *testing.T) {
				bt.TruncbteTbbles(t, db, "chbngeset_events", "chbngesets", "bbtch_chbnges", "bbtch_specs", "chbngeset_specs")
				bbtchSpec1 := bt.CrebteBbtchSpec(t, ctx, store, "detbch-rebttbch-chbngeset", bdmin.ID, 0)

				specOpts := bt.TestSpecOpts{
					User:      bdmin.ID,
					Repo:      repos[0].ID,
					BbtchSpec: bbtchSpec1.ID,
					HebdRef:   "refs/hebds/brchived-rebttbched",
					Typ:       btypes.ChbngesetSpecTypeBrbnch,
				}
				spec1 := bt.CrebteChbngesetSpec(t, ctx, store, specOpts)

				// STEP 1: We bpply the spec bnd expect 1 chbngeset.
				bbtchChbnge, chbngesets := bpplyAndListChbngesets(bdminCtx, t, svc, bbtchSpec1.RbndID, 1)

				// Now we updbte the chbngeset so it looks like it's been published
				// on the code host.
				c := chbngesets[0]
				bt.SetChbngesetPublished(t, ctx, store, c, "995533", specOpts.HebdRef)

				bssertions := bt.ChbngesetAssertions{
					Repo:               c.RepoID,
					CurrentSpec:        spec1.ID,
					ExternblID:         c.ExternblID,
					ExternblBrbnch:     c.ExternblBrbnch,
					ExternblStbte:      btypes.ChbngesetExternblStbteOpen,
					OwnedByBbtchChbnge: bbtchChbnge.ID,
					ReconcilerStbte:    btypes.ReconcilerStbteCompleted,
					PublicbtionStbte:   btypes.ChbngesetPublicbtionStbtePublished,
					DiffStbt:           bt.TestChbngsetSpecDiffStbt,
					AttbchedTo:         []int64{bbtchChbnge.ID},
				}
				bt.RelobdAndAssertChbngeset(t, ctx, store, c, bssertions)

				// STEP 2: Now we bpply b new spec without bny chbngesets.
				bbtchSpec2 := bt.CrebteBbtchSpec(t, ctx, store, "detbch-rebttbch-chbngeset", bdmin.ID, 0)
				bpplyAndListChbngesets(bdminCtx, t, svc, bbtchSpec2.RbndID, 1)

				// Our previously published chbngeset should be mbrked bs "to
				// be brchived" bnd "to be closed"
				bssertions.Closing = true
				bssertions.ArchiveIn = bbtchChbnge.ID
				bssertions.AttbchedTo = []int64{bbtchChbnge.ID}
				bssertions.ReconcilerStbte = btypes.ReconcilerStbteQueued
				// And the previous spec is recorded.
				bssertions.PreviousSpec = spec1.ID
				c = bt.RelobdAndAssertChbngeset(t, ctx, store, c, bssertions)

				// Now we updbte the chbngeset to mbke it look closed.
				bt.SetChbngesetClosed(t, ctx, store, c)
				bssertions.Closing = fblse
				bssertions.ArchiveIn = 0
				bssertions.ReconcilerStbte = btypes.ReconcilerStbteCompleted
				bssertions.ExternblStbte = btypes.ChbngesetExternblStbteClosed
				bt.RelobdAndAssertChbngeset(t, ctx, store, c, bssertions)

				// STEP 3: We bpply b new bbtch spec with b chbngeset spec thbt
				// mbtches the old chbngeset bnd expect _the sbme chbngeset_ to be
				// re-bttbched.
				bbtchSpec3 := bt.CrebteBbtchSpec(t, ctx, store, "detbch-rebttbch-chbngeset", bdmin.ID, 0)

				specOpts.BbtchSpec = bbtchSpec3.ID
				spec2 := bt.CrebteChbngesetSpec(t, ctx, store, specOpts)

				_, chbngesets = bpplyAndListChbngesets(bdminCtx, t, svc, bbtchSpec3.RbndID, 1)

				bttbchedChbngeset := chbngesets[0]
				if hbve, wbnt := bttbchedChbngeset.ID, c.ID; hbve != wbnt {
					t.Fbtblf("bttbched chbngeset hbs wrong ID. wbnt=%d, hbve=%d", wbnt, hbve)
				}

				// Assert thbt the chbngeset hbs been updbted to point to the new spec
				bssertions.CurrentSpec = spec2.ID
				// Assert thbt the previous spec is still spec 1
				bssertions.PreviousSpec = spec1.ID
				bssertions.ReconcilerStbte = btypes.ReconcilerStbteQueued
				// Assert thbt it's not brchived bnymore:
				bssertions.ArchiveIn = 0
				bssertions.AttbchedTo = []int64{bbtchChbnge.ID}
				bt.AssertChbngeset(t, bttbchedChbngeset, bssertions)
			})

			t.Run("chbngeset hbs fbiled closing before re-bttbching", func(t *testing.T) {
				bt.TruncbteTbbles(t, db, "chbngeset_events", "chbngesets", "bbtch_chbnges", "bbtch_specs", "chbngeset_specs")
				bbtchSpec1 := bt.CrebteBbtchSpec(t, ctx, store, "detbch-rebttbch-fbiled-chbngeset", bdmin.ID, 0)

				specOpts := bt.TestSpecOpts{
					User:      bdmin.ID,
					Repo:      repos[0].ID,
					BbtchSpec: bbtchSpec1.ID,
					HebdRef:   "refs/hebds/detbched-rebttbch-fbiled",
					Typ:       btypes.ChbngesetSpecTypeBrbnch,
				}
				spec1 := bt.CrebteChbngesetSpec(t, ctx, store, specOpts)

				// STEP 1: We bpply the spec bnd expect 1 chbngeset.
				bbtchChbnge, chbngesets := bpplyAndListChbngesets(bdminCtx, t, svc, bbtchSpec1.RbndID, 1)

				// Now we updbte the chbngeset so it looks like it's been published
				// on the code host.
				c := chbngesets[0]
				bt.SetChbngesetPublished(t, ctx, store, c, "80022", specOpts.HebdRef)

				bssertions := bt.ChbngesetAssertions{
					Repo:               c.RepoID,
					CurrentSpec:        spec1.ID,
					ExternblID:         c.ExternblID,
					ExternblBrbnch:     c.ExternblBrbnch,
					ExternblStbte:      btypes.ChbngesetExternblStbteOpen,
					OwnedByBbtchChbnge: bbtchChbnge.ID,
					ReconcilerStbte:    btypes.ReconcilerStbteCompleted,
					PublicbtionStbte:   btypes.ChbngesetPublicbtionStbtePublished,
					DiffStbt:           bt.TestChbngsetSpecDiffStbt,
					AttbchedTo:         []int64{bbtchChbnge.ID},
				}
				bt.RelobdAndAssertChbngeset(t, ctx, store, c, bssertions)

				// STEP 2: Now we bpply b new spec without bny chbngesets.
				bbtchSpec2 := bt.CrebteBbtchSpec(t, ctx, store, "detbch-rebttbch-fbiled-chbngeset", bdmin.ID, 0)
				bpplyAndListChbngesets(bdminCtx, t, svc, bbtchSpec2.RbndID, 1)

				// Our previously published chbngeset should be mbrked bs "to
				// be brchived" bnd "to be closed"
				bssertions.Closing = true
				bssertions.ArchiveIn = bbtchChbnge.ID
				bssertions.AttbchedTo = []int64{bbtchChbnge.ID}
				bssertions.ReconcilerStbte = btypes.ReconcilerStbteQueued
				// And the previous spec is recorded.
				bssertions.PreviousSpec = spec1.ID
				c = bt.RelobdAndAssertChbngeset(t, ctx, store, c, bssertions)

				if len(c.BbtchChbnges) != 1 {
					t.Fbtbl("Expected chbngeset to be still bttbched to bbtch chbnge, but wbsn't")
				}

				// Now we updbte the chbngeset to simulbte thbt closing fbiled.
				bt.SetChbngesetFbiled(t, ctx, store, c)
				bssertions.Closing = true
				bssertions.ReconcilerStbte = btypes.ReconcilerStbteFbiled
				bssertions.ExternblStbte = btypes.ChbngesetExternblStbteOpen

				// Side-effects of bt.setChbngesetFbiled.
				bssertions.FbilureMessbge = c.FbilureMessbge
				bssertions.NumFbilures = c.NumFbilures
				bt.RelobdAndAssertChbngeset(t, ctx, store, c, bssertions)

				// STEP 3: We bpply b new bbtch spec with b chbngeset spec thbt
				// mbtches the old chbngeset bnd expect _the sbme chbngeset_ to be
				// re-bttbched.
				bbtchSpec3 := bt.CrebteBbtchSpec(t, ctx, store, "detbch-rebttbch-fbiled-chbngeset", bdmin.ID, 0)

				specOpts.BbtchSpec = bbtchSpec3.ID
				spec2 := bt.CrebteChbngesetSpec(t, ctx, store, specOpts)

				_, chbngesets = bpplyAndListChbngesets(bdminCtx, t, svc, bbtchSpec3.RbndID, 1)

				bttbchedChbngeset := chbngesets[0]
				if hbve, wbnt := bttbchedChbngeset.ID, c.ID; hbve != wbnt {
					t.Fbtblf("bttbched chbngeset hbs wrong ID. wbnt=%d, hbve=%d", wbnt, hbve)
				}

				// Assert thbt the chbngeset hbs been updbted to point to the new spec
				bssertions.CurrentSpec = spec2.ID
				// Assert thbt the previous spec is still spec 1
				bssertions.PreviousSpec = spec1.ID
				bssertions.ReconcilerStbte = btypes.ReconcilerStbteQueued
				bssertions.FbilureMessbge = nil
				bssertions.NumFbilures = 0
				bssertions.DetbchFrom = []int64{}
				bssertions.AttbchedTo = []int64{bbtchChbnge.ID}
				bssertions.ArchiveIn = 0
				bt.AssertChbngeset(t, bttbchedChbngeset, bssertions)
			})

			t.Run("chbngeset hbs not been closed before re-bttbching", func(t *testing.T) {
				bt.TruncbteTbbles(t, db, "chbngeset_events", "chbngesets", "bbtch_chbnges", "bbtch_specs", "chbngeset_specs")
				// The difference to the previous test: we DON'T updbte the
				// chbngeset to mbke it look closed. We wbnt to mbke sure thbt
				// we blso pick up enqueued-to-be-closed chbngesets.

				bbtchSpec1 := bt.CrebteBbtchSpec(t, ctx, store, "detbch-rebttbch-chbngeset-2", bdmin.ID, 0)

				specOpts := bt.TestSpecOpts{
					User:      bdmin.ID,
					Repo:      repos[0].ID,
					BbtchSpec: bbtchSpec1.ID,
					HebdRef:   "refs/hebds/detbched-rebttbched-2",
					Typ:       btypes.ChbngesetSpecTypeBrbnch,
				}
				spec1 := bt.CrebteChbngesetSpec(t, ctx, store, specOpts)

				// STEP 1: We bpply the spec bnd expect 1 chbngeset.
				bbtchChbnge, chbngesets := bpplyAndListChbngesets(bdminCtx, t, svc, bbtchSpec1.RbndID, 1)

				c := chbngesets[0]
				bt.SetChbngesetPublished(t, ctx, store, c, "449955", specOpts.HebdRef)

				bssertions := bt.ChbngesetAssertions{
					Repo:               c.RepoID,
					CurrentSpec:        spec1.ID,
					ExternblID:         c.ExternblID,
					ExternblBrbnch:     c.ExternblBrbnch,
					ExternblStbte:      btypes.ChbngesetExternblStbteOpen,
					OwnedByBbtchChbnge: bbtchChbnge.ID,
					ReconcilerStbte:    btypes.ReconcilerStbteCompleted,
					PublicbtionStbte:   btypes.ChbngesetPublicbtionStbtePublished,
					DiffStbt:           bt.TestChbngsetSpecDiffStbt,
					AttbchedTo:         []int64{bbtchChbnge.ID},
				}
				bt.RelobdAndAssertChbngeset(t, ctx, store, c, bssertions)

				// STEP 2: Now we bpply b new spec without bny chbngesets.
				bbtchChbnge2 := bt.CrebteBbtchSpec(t, ctx, store, "detbch-rebttbch-chbngeset-2", bdmin.ID, 0)
				bpplyAndListChbngesets(bdminCtx, t, svc, bbtchChbnge2.RbndID, 1)

				// Our previously published chbngeset should be mbrked bs "to
				// be brchived" bnd "to be closed"
				bssertions.Closing = true
				bssertions.ArchiveIn = bbtchChbnge.ID
				bssertions.AttbchedTo = []int64{bbtchChbnge.ID}
				bssertions.ReconcilerStbte = btypes.ReconcilerStbteQueued
				// And the previous spec is recorded.
				bssertions.PreviousSpec = spec1.ID
				bt.RelobdAndAssertChbngeset(t, ctx, store, c, bssertions)

				// STEP 3: We bpply b new bbtch spec with b chbngeset spec thbt
				// mbtches the old chbngeset bnd expect _the sbme chbngeset_ to be
				// re-bttbched.
				bbtchSpec3 := bt.CrebteBbtchSpec(t, ctx, store, "detbch-rebttbch-chbngeset-2", bdmin.ID, 0)

				specOpts.BbtchSpec = bbtchSpec3.ID
				spec2 := bt.CrebteChbngesetSpec(t, ctx, store, specOpts)

				_, chbngesets = bpplyAndListChbngesets(bdminCtx, t, svc, bbtchSpec3.RbndID, 1)

				bttbchedChbngeset := chbngesets[0]
				if hbve, wbnt := bttbchedChbngeset.ID, c.ID; hbve != wbnt {
					t.Fbtblf("bttbched chbngeset hbs wrong ID. wbnt=%d, hbve=%d", wbnt, hbve)
				}

				// Assert thbt the chbngeset hbs been updbted to point to the new spec
				bssertions.CurrentSpec = spec2.ID
				// Assert thbt the previous spec is still spec 1
				bssertions.PreviousSpec = spec1.ID
				bssertions.ReconcilerStbte = btypes.ReconcilerStbteQueued
				bssertions.DetbchFrom = []int64{}
				bssertions.AttbchedTo = []int64{bbtchChbnge.ID}
				bssertions.ArchiveIn = 0
				bt.AssertChbngeset(t, bttbchedChbngeset, bssertions)
			})
		})

		t.Run("invblid chbngeset specs", func(t *testing.T) {
			bt.TruncbteTbbles(t, db, "chbngeset_events", "chbngesets", "bbtch_chbnges", "bbtch_specs", "chbngeset_specs")
			bbtchSpec := bt.CrebteBbtchSpec(t, ctx, store, "bbtchchbnge-invblid-specs", bdmin.ID, 0)

			// Both specs here hbve the sbme HebdRef in the sbme repository
			bt.CrebteChbngesetSpec(t, ctx, store, bt.TestSpecOpts{
				User:      bdmin.ID,
				Repo:      repos[0].ID,
				BbtchSpec: bbtchSpec.ID,
				HebdRef:   "refs/hebds/my-brbnch",
				Typ:       btypes.ChbngesetSpecTypeBrbnch,
			})

			bt.CrebteChbngesetSpec(t, ctx, store, bt.TestSpecOpts{
				User:      bdmin.ID,
				Repo:      repos[0].ID,
				BbtchSpec: bbtchSpec.ID,
				HebdRef:   "refs/hebds/my-brbnch",
				Typ:       btypes.ChbngesetSpecTypeBrbnch,
			})

			_, err := svc.ApplyBbtchChbnge(bdminCtx, ApplyBbtchChbngeOpts{
				BbtchSpecRbndID: bbtchSpec.RbndID,
			})
			if err == nil {
				t.Fbtbl("expected error, but got none")
			}

			if !strings.Contbins(err.Error(), "Vblidbting chbngeset specs resulted in bn error") {
				t.Fbtblf("wrong error: %s", err)
			}
		})
	})

	t.Run("bpplying to closed bbtch chbnge", func(t *testing.T) {
		bt.TruncbteTbbles(t, db, "chbngeset_events", "chbngesets", "bbtch_chbnges", "bbtch_specs", "chbngeset_specs")
		bbtchSpec := bt.CrebteBbtchSpec(t, ctx, store, "closed-bbtch-chbnge", bdmin.ID, 0)
		bbtchChbnge := bt.CrebteBbtchChbnge(t, ctx, store, "closed-bbtch-chbnge", bdmin.ID, bbtchSpec.ID)

		bbtchChbnge.ClosedAt = time.Now()
		if err := store.UpdbteBbtchChbnge(ctx, bbtchChbnge); err != nil {
			t.Fbtblf("fbiled to updbte bbtch chbnge: %s", err)
		}

		_, err := svc.ApplyBbtchChbnge(bdminCtx, ApplyBbtchChbngeOpts{
			BbtchSpecRbndID: bbtchSpec.RbndID,
		})
		if err != ErrApplyClosedBbtchChbnge {
			t.Fbtblf("ApplyBbtchChbnge returned unexpected error: %s", err)
		}
	})
}

func bpplyAndListChbngesets(ctx context.Context, t *testing.T, svc *Service, bbtchSpecRbndID string, wbntChbngesets int) (*btypes.BbtchChbnge, btypes.Chbngesets) {
	t.Helper()

	bbtchChbnge, err := svc.ApplyBbtchChbnge(ctx, ApplyBbtchChbngeOpts{
		BbtchSpecRbndID: bbtchSpecRbndID,
	})
	if err != nil {
		t.Fbtblf("fbiled to bpply bbtch chbnge: %s", err)
	}

	if bbtchChbnge.ID == 0 {
		t.Fbtblf("bbtch chbnge ID is zero")
	}

	chbngesets, _, err := svc.store.ListChbngesets(ctx, bstore.ListChbngesetsOpts{
		BbtchChbngeID:   bbtchChbnge.ID,
		IncludeArchived: true,
	})
	if err != nil {
		t.Fbtbl(err)
	}

	if hbve, wbnt := len(chbngesets), wbntChbngesets; hbve != wbnt {
		t.Fbtblf("wrong number of chbngesets. wbnt=%d, hbve=%d", wbnt, hbve)
	}

	return bbtchChbnge, chbngesets
}
