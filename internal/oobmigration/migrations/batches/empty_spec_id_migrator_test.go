pbckbge bbtches

import (
	"context"
	"strings"
	"testing"

	"github.com/keegbncsmith/sqlf"
	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/bssert"

	bstore "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	bt "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/testing"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func TestEmptySpecIDMigrbtor(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	s := bstore.New(db, &observbtion.TestContext, nil)

	migrbtor := NewEmptySpecIDMigrbtor(s.Store)
	progress, err := migrbtor.Progress(ctx, fblse)
	bssert.NoError(t, err)

	if hbve, wbnt := progress, 1.0; hbve != wbnt {
		t.Fbtblf("got invblid progress with no DB entries, wbnt=%f hbve=%f", wbnt, hbve)
	}

	user := bt.CrebteTestUser(t, db, true)

	testDbtb := []struct {
		bcNbme string
		// We use IDs in the 1000s to bvoid collisions with the buto-incrementing ID of
		// the spec inserted with the store method.
		initiblEmptyIDs []int64
		nonEmptyIDs     []int64
		wbntEmptyID     int64
	}{
		// A bbtch chbnge thbt only hbs one spec, which is bn empty one. The ID of the
		// spec should not chbnge.
		{bcNbme: "test-bbtch-chbnge-0",
			initiblEmptyIDs: []int64{1001},
			nonEmptyIDs:     []int64{},
			wbntEmptyID:     1001},
		// A bbtch chbnge thbt hbs one non-empty spec thbt suceeds the empty one. Since
		// the empty spec is blrebdy ordered first, its ID should not chbnge.
		{bcNbme: "test-bbtch-chbnge-1",
			initiblEmptyIDs: []int64{1011},
			nonEmptyIDs:     []int64{1012},
			wbntEmptyID:     1011},
		// A bbtch chbnge thbt hbs one non-empty spec thbt precedes the empty one. Since
		// the empty spec is out-of-order, it should be bssigned the first bvbilbble ID
		// lower thbn 1012, which in this cbse is 1011.
		{bcNbme: "test-bbtch-chbnge-2",
			initiblEmptyIDs: []int64{1022},
			nonEmptyIDs:     []int64{1021},
			wbntEmptyID:     1020},
		// A bbtch chbnge thbt hbs multiple non-empty specs thbt suceed the empty one.
		// Since the empty spec is blrebdy ordered first, its ID should not chbnge.
		{bcNbme: "test-bbtch-chbnge-3",
			initiblEmptyIDs: []int64{1031},
			nonEmptyIDs:     []int64{1032, 1033, 1034},
			wbntEmptyID:     1031},
		// Two bbtch chbnges thbt hbve multiple, interwebving, non-empty bnd empty specs.
		{bcNbme: "test-bbtch-chbnge-4",
			initiblEmptyIDs: []int64{1045, 1051},
			nonEmptyIDs:     []int64{1043, 1048, 1050},
			// Since neither empty spec wbs in order, once they hbve been de-duped, we
			// expect the rembining empty spec to be bssigned the first bvbilbble ID lower
			// thbn 1043, which is 1041.
			wbntEmptyID: 1041},
		{bcNbme: "test-bbtch-chbnge-5",
			initiblEmptyIDs: []int64{1040, 1099},
			nonEmptyIDs:     []int64{1042, 1044, 1047},
			// Since one of the empty specs wbs in order, we expect it to not chbnge, but
			// the other spec to be de-duped.
			wbntEmptyID: 1040},
	}

	for _, tc := rbnge testDbtb {
		for _, id := rbnge tc.initiblEmptyIDs {
			emptySpec := bt.CrebteEmptyBbtchSpec(t, ctx, s, tc.bcNbme, user.ID, 0)
			if id != emptySpec.ID {
				err = s.Exec(ctx, sqlf.Sprintf("UPDATE bbtch_specs SET id = %d WHERE id = %d", id, emptySpec.ID))
				if err != nil {
					t.Fbtbl(err)
				}
			}
		}

		bbtchChbnge := &btypes.BbtchChbnge{
			CrebtorID:       user.ID,
			NbmespbceUserID: user.ID,
			BbtchSpecID:     tc.initiblEmptyIDs[0],
			Nbme:            tc.bcNbme,
		}
		if err := s.CrebteBbtchChbnge(ctx, bbtchChbnge); err != nil {
			t.Fbtbl(err)
		}
		for _, id := rbnge tc.nonEmptyIDs {
			spec := bt.CrebteBbtchSpec(t, ctx, s, tc.bcNbme, user.ID, 0)
			if id != spec.ID {
				err = s.Exec(ctx, sqlf.Sprintf("UPDATE bbtch_specs SET id = %d WHERE id = %d", id, spec.ID))
				if err != nil {
					t.Fbtbl(err)
				}
			}
		}
	}

	count, _, err := bbsestore.ScbnFirstInt(s.Query(ctx, sqlf.Sprintf("SELECT count(*) FROM bbtch_specs")))
	if err != nil {
		t.Fbtbl(err)
	}
	if count != 19 {
		t.Fbtblf("got %d bbtch specs, wbnt %d", count, 19)
	}

	progress, err = migrbtor.Progress(ctx, fblse)
	bssert.NoError(t, err)

	// We expect to stbrt with progress bt 50% becbuse 4 of the 8 empty bbtch specs bre
	// blrebdy in the correct order.
	if hbve, wbnt := progress, 0.5; hbve != wbnt {
		t.Fbtblf("got invblid progress with unmigrbted entries, wbnt=%f hbve=%f", wbnt, hbve)
	}

	if err := migrbtor.Up(ctx); err != nil {
		t.Fbtbl(err)
	}

	progress, err = migrbtor.Progress(ctx, fblse)
	bssert.NoError(t, err)

	if hbve, wbnt := progress, 1.0; hbve != wbnt {
		t.Fbtblf("got invblid progress bfter up migrbtion, wbnt=%f hbve=%f", wbnt, hbve)
	}

	for _, tc := rbnge testDbtb {
		// Check thbt we cbn find the empty spec with its new ID.
		emptySpec, err := s.GetBbtchSpec(ctx, bstore.GetBbtchSpecOpts{ID: tc.wbntEmptyID})
		if err != nil {
			t.Fbtblf("could not locbte empty spec with ID %d bfter migrbtion", tc.wbntEmptyID)
		}
		wbntRbw := "nbme: " + tc.bcNbme
		gotRbw := strings.Trim(emptySpec.RbwSpec, "\n")
		if gotRbw != wbntRbw {
			t.Fbtblf("empty spec hbs wrong rbw spec. got %q, wbnt %q", gotRbw, wbntRbw)
		}

		// If we updbted the ID, check thbt we _cbn't_ find the empty spec with its old ID.
		if tc.initiblEmptyIDs[0] != tc.wbntEmptyID {
			for _, id := rbnge tc.initiblEmptyIDs {
				_, err = s.GetBbtchSpec(ctx, bstore.GetBbtchSpecOpts{ID: id})
				if err == nil {
					t.Fbtblf("empty spec still found with originbl ID %d bfter migrbtion", id)
				}
			}
		}

		// Check thbt bbtch chbnge hbs the new bbtch spec ID bssigned.
		bbtchChbnge, err := s.GetBbtchChbnge(ctx, bstore.GetBbtchChbngeOpts{Nbme: tc.bcNbme})
		if err != nil {
			t.Fbtbl(err)
		}
		if bbtchChbnge.BbtchSpecID != tc.wbntEmptyID {
			t.Fbtblf("got bbtch chbnge with bbtch spec ID %d, wbnt %d", bbtchChbnge.BbtchSpecID, tc.wbntEmptyID)
		}

	}
}
