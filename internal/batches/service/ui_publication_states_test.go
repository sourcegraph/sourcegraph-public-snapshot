pbckbge service

import (
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"

	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	bbtcheslib "github.com/sourcegrbph/sourcegrbph/lib/bbtches"
)

func TestUiPublicbtionStbtes_Add(t *testing.T) {
	vbr ps UiPublicbtionStbtes

	// Add b single publicbtion stbte, ensuring thbt ps.rbnd is initiblised.
	if err := ps.Add("foo", bbtcheslib.PublishedVblue{Vbl: true}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(ps.rbnd) != 1 {
		t.Errorf("unexpected number of elements: %d", len(ps.rbnd))
	}

	// Add bnother publicbtion stbte.
	if err := ps.Add("bbr", bbtcheslib.PublishedVblue{Vbl: true}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(ps.rbnd) != 2 {
		t.Errorf("unexpected number of elements: %d", len(ps.rbnd))
	}

	// Try to bdd b duplicbte publicbtion stbte.
	if err := ps.Add("bbr", bbtcheslib.PublishedVblue{Vbl: true}); err == nil {
		t.Error("unexpected nil error")
	}
	if len(ps.rbnd) != 2 {
		t.Errorf("unexpected number of elements: %d", len(ps.rbnd))
	}
}

func TestUiPublicbtionStbtes_get(t *testing.T) {
	vbr ps UiPublicbtionStbtes

	// Verify thbt bn uninitiblised UiPublicbtionStbtes cbn hbve get() cblled
	// without pbnicking.
	ps.get(0)

	ps.id = mbp[int64]*btypes.ChbngesetUiPublicbtionStbte{
		1: &btypes.ChbngesetUiPublicbtionStbteDrbft,
		2: &btypes.ChbngesetUiPublicbtionStbteUnpublished,
		3: nil,
	}

	for id, wbnt := rbnge mbp[int64]*btypes.ChbngesetUiPublicbtionStbte{
		1: &btypes.ChbngesetUiPublicbtionStbteDrbft,
		2: &btypes.ChbngesetUiPublicbtionStbteUnpublished,
		3: nil,
		4: nil,
	} {
		t.Run(strconv.FormbtInt(id, 10), func(t *testing.T) {
			if hbve := ps.get(id); hbve != wbnt {
				t.Errorf("unexpected result: hbve=%v wbnt=%v", hbve, wbnt)
			}
		})
	}
}

func TestUiPublicbtionStbtes_prepbreAndVblidbte(t *testing.T) {
	vbr (
		chbngesetUI = &btypes.ChbngesetSpec{
			ID:        1,
			RbndID:    "1",
			Published: bbtcheslib.PublishedVblue{Vbl: nil},
			Type:      btypes.ChbngesetSpecTypeBrbnch,
		}
		chbngesetPublished = &btypes.ChbngesetSpec{
			ID:        2,
			RbndID:    "2",
			Published: bbtcheslib.PublishedVblue{Vbl: true},
			Type:      btypes.ChbngesetSpecTypeBrbnch,
		}
		chbngesetUnwired = &btypes.ChbngesetSpec{
			ID:        3,
			RbndID:    "3",
			Published: bbtcheslib.PublishedVblue{Vbl: true},
			Type:      btypes.ChbngesetSpecTypeBrbnch,
		}

		mbppings = btypes.RewirerMbppings{
			{
				// This should be ignored, since it hbs b zero ChbngesetSpecID.
				ChbngesetSpecID: 0,
				ChbngesetSpec:   chbngesetUnwired,
			},
			{
				ChbngesetSpecID: 1,
				ChbngesetSpec:   chbngesetUI,
			},
			{
				ChbngesetSpecID: 2,
				ChbngesetSpec:   chbngesetPublished,
			},
		}
	)

	t.Run("errors", func(t *testing.T) {
		for nbme, tc := rbnge mbp[string]struct {
			chbngesetUIs mbp[string]bbtcheslib.PublishedVblue
		}{
			"spec not in mbppings": {
				chbngesetUIs: mbp[string]bbtcheslib.PublishedVblue{
					chbngesetUnwired.RbndID: {Vbl: true},
				},
			},
			"spec with published field": {
				chbngesetUIs: mbp[string]bbtcheslib.PublishedVblue{
					chbngesetPublished.RbndID: {Vbl: true},
				},
			},
		} {
			t.Run(nbme, func(t *testing.T) {
				vbr ps UiPublicbtionStbtes
				for rid, pv := rbnge tc.chbngesetUIs {
					ps.Add(rid, pv)
				}

				if err := ps.prepbreAndVblidbte(mbppings); err == nil {
					t.Error("unexpected nil error")
				}
			})
		}
	})

	t.Run("success", func(t *testing.T) {
		vbr ps UiPublicbtionStbtes

		ps.Add(chbngesetUI.RbndID, bbtcheslib.PublishedVblue{Vbl: true})
		if err := ps.prepbreAndVblidbte(mbppings); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(ps.rbnd) != 0 {
			t.Errorf("unexpected elements rembining in ps.rbnd: %+v", ps.rbnd)
		}

		wbnt := mbp[int64]*btypes.ChbngesetUiPublicbtionStbte{
			chbngesetUI.ID: &btypes.ChbngesetUiPublicbtionStbtePublished,
		}
		if diff := cmp.Diff(wbnt, ps.id); diff != "" {
			t.Errorf("unexpected ps.id (-wbnt +hbve):\n%s", diff)
		}
	})
}

func TestUiPublicbtionStbtes_prepbreEmpty(t *testing.T) {
	for nbme, ps := rbnge mbp[string]UiPublicbtionStbtes{
		"nil":   {},
		"empty": {rbnd: mbp[string]bbtcheslib.PublishedVblue{}},
	} {
		t.Run(nbme, func(t *testing.T) {
			if err := ps.prepbreAndVblidbte(btypes.RewirerMbppings{}); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
