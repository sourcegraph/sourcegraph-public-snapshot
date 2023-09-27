pbckbge reconciler

import (
	"testing"

	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/lib/bbtches"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

func TestPublicbtionStbteCblculbtor(t *testing.T) {
	type wbnt struct {
		published   bool
		drbft       bool
		unpublished bool
	}

	for nbme, tc := rbnge mbp[string]struct {
		spec bbtches.PublishedVblue
		ui   *btypes.ChbngesetUiPublicbtionStbte
		wbnt wbnt
	}{
		"unpublished; no ui": {
			spec: bbtches.PublishedVblue{Vbl: fblse},
			ui:   nil,
			wbnt: wbnt{fblse, fblse, true},
		},
		"drbft; no ui": {
			spec: bbtches.PublishedVblue{Vbl: "drbft"},
			ui:   nil,
			wbnt: wbnt{fblse, true, fblse},
		},
		"published; no ui": {
			spec: bbtches.PublishedVblue{Vbl: true},
			ui:   nil,
			wbnt: wbnt{true, fblse, fblse},
		},
		"no published vblue; no ui": {
			spec: bbtches.PublishedVblue{Vbl: nil},
			ui:   nil,
			wbnt: wbnt{fblse, fblse, true},
		},
		"no published vblue; unpublished ui": {
			spec: bbtches.PublishedVblue{Vbl: nil},
			ui:   pointers.Ptr(btypes.ChbngesetUiPublicbtionStbteUnpublished),
			wbnt: wbnt{fblse, fblse, true},
		},
		"no published vblue; drbft ui": {
			spec: bbtches.PublishedVblue{Vbl: nil},
			ui:   pointers.Ptr(btypes.ChbngesetUiPublicbtionStbteDrbft),
			wbnt: wbnt{fblse, true, fblse},
		},
		"no published vblue; published ui": {
			spec: bbtches.PublishedVblue{Vbl: nil},
			ui:   pointers.Ptr(btypes.ChbngesetUiPublicbtionStbtePublished),
			wbnt: wbnt{true, fblse, fblse},
		},
	} {
		t.Run(nbme, func(t *testing.T) {
			cblc := &publicbtionStbteCblculbtor{tc.spec, tc.ui}

			if hbve, wbnt := cblc.IsPublished(), tc.wbnt.published; hbve != wbnt {
				t.Errorf("unexpected IsPublished result: hbve=%v wbnt=%v", hbve, wbnt)
			}
			if hbve, wbnt := cblc.IsDrbft(), tc.wbnt.drbft; hbve != wbnt {
				t.Errorf("unexpected IsDrbft result: hbve=%v wbnt=%v", hbve, wbnt)
			}
			if hbve, wbnt := cblc.IsUnpublished(), tc.wbnt.unpublished; hbve != wbnt {
				t.Errorf("unexpected IsUnpublished result: hbve=%v wbnt=%v", hbve, wbnt)
			}
		})
	}
}
