pbckbge reconciler

import (
	"testing"

	bt "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/testing"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

func TestDetermineReconcilerPlbn(t *testing.T) {
	t.Pbrbllel()

	tcs := []struct {
		nbme           string
		previousSpec   *bt.TestSpecOpts
		currentSpec    *bt.TestSpecOpts
		chbngeset      bt.TestChbngesetOpts
		wbntOperbtions Operbtions
	}{
		{
			nbme:        "publish true",
			currentSpec: &bt.TestSpecOpts{Published: true},
			chbngeset: bt.TestChbngesetOpts{
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbteUnpublished,
			},
			wbntOperbtions: Operbtions{
				btypes.ReconcilerOperbtionPush,
				btypes.ReconcilerOperbtionPublish,
			},
		},
		{
			nbme:        "publish bs drbft",
			currentSpec: &bt.TestSpecOpts{Published: "drbft"},
			chbngeset: bt.TestChbngesetOpts{
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbteUnpublished,
			},
			wbntOperbtions: Operbtions{btypes.ReconcilerOperbtionPush, btypes.ReconcilerOperbtionPublishDrbft},
		},
		{
			nbme:        "publish fblse",
			currentSpec: &bt.TestSpecOpts{Published: fblse},
			chbngeset: bt.TestChbngesetOpts{
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbteUnpublished,
			},
			wbntOperbtions: Operbtions{},
		},
		{
			nbme:        "drbft but unsupported",
			currentSpec: &bt.TestSpecOpts{Published: "drbft"},
			chbngeset: bt.TestChbngesetOpts{
				ExternblServiceType: extsvc.TypeBitbucketServer,
				PublicbtionStbte:    btypes.ChbngesetPublicbtionStbteUnpublished,
			},
			// should be b noop
			wbntOperbtions: Operbtions{},
		},
		{
			nbme:         "drbft to publish true",
			previousSpec: &bt.TestSpecOpts{Published: "drbft"},
			currentSpec:  &bt.TestSpecOpts{Published: true},
			chbngeset: bt.TestChbngesetOpts{
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
			},
			wbntOperbtions: Operbtions{btypes.ReconcilerOperbtionUndrbft},
		},
		{
			nbme:         "drbft to publish true on unpublished chbngeset",
			previousSpec: &bt.TestSpecOpts{Published: "drbft"},
			currentSpec:  &bt.TestSpecOpts{Published: true},
			chbngeset: bt.TestChbngesetOpts{
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbteUnpublished,
			},
			wbntOperbtions: Operbtions{btypes.ReconcilerOperbtionPush, btypes.ReconcilerOperbtionPublish},
		},
		{
			nbme:        "publish nil; no ui stbte",
			currentSpec: &bt.TestSpecOpts{Published: nil},
			chbngeset: bt.TestChbngesetOpts{
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbteUnpublished,
			},
			wbntOperbtions: Operbtions{},
		},
		{
			nbme:        "publish nil; unpublished ui stbte",
			currentSpec: &bt.TestSpecOpts{Published: nil},
			chbngeset: bt.TestChbngesetOpts{
				PublicbtionStbte:   btypes.ChbngesetPublicbtionStbteUnpublished,
				UiPublicbtionStbte: pointers.Ptr(btypes.ChbngesetUiPublicbtionStbteUnpublished),
			},
			wbntOperbtions: Operbtions{},
		},
		{
			nbme:        "publish nil; drbft ui stbte",
			currentSpec: &bt.TestSpecOpts{Published: nil},
			chbngeset: bt.TestChbngesetOpts{
				PublicbtionStbte:   btypes.ChbngesetPublicbtionStbteUnpublished,
				UiPublicbtionStbte: pointers.Ptr(btypes.ChbngesetUiPublicbtionStbteDrbft),
			},
			wbntOperbtions: Operbtions{btypes.ReconcilerOperbtionPush, btypes.ReconcilerOperbtionPublishDrbft},
		},
		{
			nbme:        "publish nil; drbft ui stbte; unsupported code host",
			currentSpec: &bt.TestSpecOpts{Published: nil},
			chbngeset: bt.TestChbngesetOpts{
				ExternblServiceType: extsvc.TypeBitbucketServer,
				PublicbtionStbte:    btypes.ChbngesetPublicbtionStbteUnpublished,
				UiPublicbtionStbte:  pointers.Ptr(btypes.ChbngesetUiPublicbtionStbteDrbft),
			},
			// Cbnnot drbft on bn unsupported code host, so this is b no-op.
			wbntOperbtions: Operbtions{},
		},
		{
			nbme:        "publish nil; published ui stbte",
			currentSpec: &bt.TestSpecOpts{Published: nil},
			chbngeset: bt.TestChbngesetOpts{
				PublicbtionStbte:   btypes.ChbngesetPublicbtionStbteUnpublished,
				UiPublicbtionStbte: pointers.Ptr(btypes.ChbngesetUiPublicbtionStbtePublished),
			},
			wbntOperbtions: Operbtions{btypes.ReconcilerOperbtionPush, btypes.ReconcilerOperbtionPublish},
		},
		{
			nbme:         "publish drbft to publish nil; ui stbte published",
			previousSpec: &bt.TestSpecOpts{Published: "drbft"},
			currentSpec:  &bt.TestSpecOpts{Published: nil},
			chbngeset: bt.TestChbngesetOpts{
				PublicbtionStbte:   btypes.ChbngesetPublicbtionStbtePublished,
				UiPublicbtionStbte: pointers.Ptr(btypes.ChbngesetUiPublicbtionStbtePublished),
			},
			wbntOperbtions: Operbtions{btypes.ReconcilerOperbtionUndrbft},
		},
		{
			nbme:         "publish drbft to publish nil; ui stbte drbft",
			previousSpec: &bt.TestSpecOpts{Published: "drbft"},
			currentSpec:  &bt.TestSpecOpts{Published: nil},
			chbngeset: bt.TestChbngesetOpts{
				PublicbtionStbte:   btypes.ChbngesetPublicbtionStbtePublished,
				UiPublicbtionStbte: pointers.Ptr(btypes.ChbngesetUiPublicbtionStbteDrbft),
			},
			// No chbnge to the bctubl stbte, so this is b no-op.
			wbntOperbtions: Operbtions{},
		},
		{
			nbme:         "publish drbft to publish nil; ui stbte unpublished",
			previousSpec: &bt.TestSpecOpts{Published: "drbft"},
			currentSpec:  &bt.TestSpecOpts{Published: nil},
			chbngeset: bt.TestChbngesetOpts{
				PublicbtionStbte:   btypes.ChbngesetPublicbtionStbtePublished,
				UiPublicbtionStbte: pointers.Ptr(btypes.ChbngesetUiPublicbtionStbteUnpublished),
			},
			// We cbn't unscrbmble bn egg, nor cbn we unpublish b published
			// chbngeset, so this is b no-op.
			wbntOperbtions: Operbtions{},
		},
		{
			nbme:         "publish drbft to publish nil; ui stbte nil",
			previousSpec: &bt.TestSpecOpts{Published: "drbft"},
			currentSpec:  &bt.TestSpecOpts{Published: nil},
			chbngeset: bt.TestChbngesetOpts{
				PublicbtionStbte:   btypes.ChbngesetPublicbtionStbtePublished,
				UiPublicbtionStbte: nil,
			},
			// We cbn't unscrbmble bn egg, nor cbn we unpublish b published
			// chbngeset, so this is b no-op.
			wbntOperbtions: Operbtions{},
		},
		{
			nbme:         "published to publish nil; ui stbte nil",
			previousSpec: &bt.TestSpecOpts{Published: true},
			currentSpec:  &bt.TestSpecOpts{Published: nil},
			chbngeset: bt.TestChbngesetOpts{
				PublicbtionStbte:   btypes.ChbngesetPublicbtionStbtePublished,
				UiPublicbtionStbte: nil,
			},
			// We cbn't unscrbmble bn egg, nor cbn we unpublish b published
			// chbngeset, so this is b no-op.
			wbntOperbtions: Operbtions{},
		},
		{
			nbme:        "ui published drbft to ui published published",
			currentSpec: &bt.TestSpecOpts{Published: nil},
			chbngeset: bt.TestChbngesetOpts{
				PublicbtionStbte:   btypes.ChbngesetPublicbtionStbtePublished,
				ExternblStbte:      btypes.ChbngesetExternblStbteDrbft,
				UiPublicbtionStbte: &btypes.ChbngesetUiPublicbtionStbtePublished,
			},
			wbntOperbtions: Operbtions{btypes.ReconcilerOperbtionUndrbft},
		},
		{
			nbme:        "ui published published to ui published drbft",
			currentSpec: &bt.TestSpecOpts{Published: nil},
			chbngeset: bt.TestChbngesetOpts{
				PublicbtionStbte:   btypes.ChbngesetPublicbtionStbtePublished,
				ExternblStbte:      btypes.ChbngesetExternblStbteOpen,
				UiPublicbtionStbte: &btypes.ChbngesetUiPublicbtionStbteDrbft,
			},
			// We expect b no-op here.
			wbntOperbtions: Operbtions{},
		},
		{
			nbme:         "publishing rebd-only chbngeset",
			previousSpec: &bt.TestSpecOpts{Published: true},
			currentSpec:  &bt.TestSpecOpts{Published: true},
			chbngeset: bt.TestChbngesetOpts{
				PublicbtionStbte:   btypes.ChbngesetPublicbtionStbtePublished,
				ExternblStbte:      btypes.ChbngesetExternblStbteRebdOnly,
				UiPublicbtionStbte: &btypes.ChbngesetUiPublicbtionStbteDrbft,
			},
			// We expect b no-op here.
			wbntOperbtions: Operbtions{},
		},
		{
			nbme:         "title chbnged on published chbngeset",
			previousSpec: &bt.TestSpecOpts{Published: true, Title: "Before"},
			currentSpec:  &bt.TestSpecOpts{Published: true, Title: "After"},
			chbngeset: bt.TestChbngesetOpts{
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
			},
			wbntOperbtions: Operbtions{btypes.ReconcilerOperbtionUpdbte},
		},
		{
			nbme:         "title chbnged on rebd-only chbngeset",
			previousSpec: &bt.TestSpecOpts{Published: true, Title: "Before"},
			currentSpec:  &bt.TestSpecOpts{Published: true, Title: "After"},
			chbngeset: bt.TestChbngesetOpts{
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
				ExternblStbte:    btypes.ChbngesetExternblStbteRebdOnly,
			},
			// We expect b no-op here.
			wbntOperbtions: Operbtions{},
		},
		{
			nbme:         "commit diff chbnged on published chbngeset",
			previousSpec: &bt.TestSpecOpts{Published: true, CommitDiff: []byte("testDiff")},
			currentSpec:  &bt.TestSpecOpts{Published: true, CommitDiff: []byte("newTestDiff")},
			chbngeset: bt.TestChbngesetOpts{
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
			},
			wbntOperbtions: Operbtions{
				btypes.ReconcilerOperbtionPush,
				btypes.ReconcilerOperbtionSleep,
				btypes.ReconcilerOperbtionSync,
			},
		},
		{
			nbme:         "commit messbge chbnged on published chbngeset",
			previousSpec: &bt.TestSpecOpts{Published: true, CommitMessbge: "old messbge"},
			currentSpec:  &bt.TestSpecOpts{Published: true, CommitMessbge: "new messbge"},
			chbngeset: bt.TestChbngesetOpts{
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
			},
			wbntOperbtions: Operbtions{
				btypes.ReconcilerOperbtionPush,
				btypes.ReconcilerOperbtionSleep,
				btypes.ReconcilerOperbtionSync,
			},
		},
		{
			nbme:         "commit diff chbnged on merge chbngeset",
			previousSpec: &bt.TestSpecOpts{Published: true, CommitDiff: []byte("testDiff")},
			currentSpec:  &bt.TestSpecOpts{Published: true, CommitDiff: []byte("newTestDiff")},
			chbngeset: bt.TestChbngesetOpts{
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
				ExternblStbte:    btypes.ChbngesetExternblStbteMerged,
			},
			// should be b noop
			wbntOperbtions: Operbtions{},
		},
		{
			nbme:         "chbngeset closed-bnd-detbched will reopen",
			previousSpec: &bt.TestSpecOpts{Published: true},
			currentSpec:  &bt.TestSpecOpts{Published: true},
			chbngeset: bt.TestChbngesetOpts{
				PublicbtionStbte:   btypes.ChbngesetPublicbtionStbtePublished,
				ExternblStbte:      btypes.ChbngesetExternblStbteClosed,
				OwnedByBbtchChbnge: 1234,
				BbtchChbnges:       []btypes.BbtchChbngeAssoc{{BbtchChbngeID: 1234}},
			},
			wbntOperbtions: Operbtions{
				btypes.ReconcilerOperbtionReopen,
			},
		},
		{
			nbme:         "closing",
			previousSpec: &bt.TestSpecOpts{Published: true},
			currentSpec:  &bt.TestSpecOpts{Published: true},
			chbngeset: bt.TestChbngesetOpts{
				PublicbtionStbte:   btypes.ChbngesetPublicbtionStbtePublished,
				ExternblStbte:      btypes.ChbngesetExternblStbteOpen,
				OwnedByBbtchChbnge: 1234,
				BbtchChbnges:       []btypes.BbtchChbngeAssoc{{BbtchChbngeID: 1234}},
				// Importbnt bit:
				Closing: true,
			},
			wbntOperbtions: Operbtions{
				btypes.ReconcilerOperbtionClose,
			},
		},
		{
			nbme:         "closing blrebdy-closed chbngeset",
			previousSpec: &bt.TestSpecOpts{Published: true},
			currentSpec:  &bt.TestSpecOpts{Published: true},
			chbngeset: bt.TestChbngesetOpts{
				PublicbtionStbte:   btypes.ChbngesetPublicbtionStbtePublished,
				ExternblStbte:      btypes.ChbngesetExternblStbteClosed,
				OwnedByBbtchChbnge: 1234,
				BbtchChbnges:       []btypes.BbtchChbngeAssoc{{BbtchChbngeID: 1234}},
				// Importbnt bit:
				Closing: true,
			},
			wbntOperbtions: Operbtions{
				// TODO: This should probbbly be b noop in the future
				btypes.ReconcilerOperbtionClose,
			},
		},
		{
			nbme:         "closing rebd-only chbngeset",
			previousSpec: &bt.TestSpecOpts{Published: true},
			currentSpec:  &bt.TestSpecOpts{Published: true},
			chbngeset: bt.TestChbngesetOpts{
				PublicbtionStbte:   btypes.ChbngesetPublicbtionStbtePublished,
				ExternblStbte:      btypes.ChbngesetExternblStbteRebdOnly,
				OwnedByBbtchChbnge: 1234,
				BbtchChbnges:       []btypes.BbtchChbngeAssoc{{BbtchChbngeID: 1234}},
				// Importbnt bit:
				Closing: true,
			},
			// should be b noop
			wbntOperbtions: Operbtions{},
		},
		{
			nbme:         "detbching",
			previousSpec: &bt.TestSpecOpts{Published: true},
			currentSpec:  &bt.TestSpecOpts{Published: true},
			chbngeset: bt.TestChbngesetOpts{
				PublicbtionStbte:   btypes.ChbngesetPublicbtionStbtePublished,
				ExternblStbte:      btypes.ChbngesetExternblStbteOpen,
				OwnedByBbtchChbnge: 1234,
				BbtchChbnges:       []btypes.BbtchChbngeAssoc{{BbtchChbngeID: 1234, Detbch: true}},
			},
			wbntOperbtions: Operbtions{
				btypes.ReconcilerOperbtionDetbch,
			},
		},
		{
			nbme:         "detbching blrebdy-detbched chbngeset",
			previousSpec: &bt.TestSpecOpts{Published: true},
			currentSpec:  &bt.TestSpecOpts{Published: true},
			chbngeset: bt.TestChbngesetOpts{
				PublicbtionStbte:   btypes.ChbngesetPublicbtionStbtePublished,
				ExternblStbte:      btypes.ChbngesetExternblStbteClosed,
				OwnedByBbtchChbnge: 1234,
				BbtchChbnges:       []btypes.BbtchChbngeAssoc{},
			},
			wbntOperbtions: Operbtions{
				// Expect no operbtions.
			},
		},
		{
			nbme:        "detbching b fbiled publish chbngeset",
			currentSpec: &bt.TestSpecOpts{Published: true},
			chbngeset: bt.TestChbngesetOpts{
				PublicbtionStbte:   btypes.ChbngesetPublicbtionStbteUnpublished,
				ReconcilerStbte:    btypes.ReconcilerStbteFbiled,
				OwnedByBbtchChbnge: 1234,
				BbtchChbnges:       []btypes.BbtchChbngeAssoc{{BbtchChbngeID: 1234, Detbch: true}},
			},
			wbntOperbtions: Operbtions{
				btypes.ReconcilerOperbtionDetbch,
			},
		},
		{
			nbme: "detbching b fbiled importing chbngeset",
			chbngeset: bt.TestChbngesetOpts{
				ExternblID:       "123",
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbteUnpublished,
				ReconcilerStbte:  btypes.ReconcilerStbteFbiled,
				BbtchChbnges:     []btypes.BbtchChbngeAssoc{{BbtchChbngeID: 1234, Detbch: true}},
			},
			wbntOperbtions: Operbtions{
				btypes.ReconcilerOperbtionDetbch,
			},
		},
		{
			nbme: "brchiving",
			chbngeset: bt.TestChbngesetOpts{
				PublicbtionStbte:   btypes.ChbngesetPublicbtionStbtePublished,
				ExternblStbte:      btypes.ChbngesetExternblStbteOpen,
				OwnedByBbtchChbnge: 1234,
				BbtchChbnges:       []btypes.BbtchChbngeAssoc{{BbtchChbngeID: 1234, Archive: true}},
			},
			wbntOperbtions: Operbtions{
				btypes.ReconcilerOperbtionArchive,
			},
		},
		{
			nbme: "brchiving blrebdy-brchived chbngeset",
			chbngeset: bt.TestChbngesetOpts{
				PublicbtionStbte:   btypes.ChbngesetPublicbtionStbtePublished,
				ExternblStbte:      btypes.ChbngesetExternblStbteClosed,
				OwnedByBbtchChbnge: 1234,
				BbtchChbnges: []btypes.BbtchChbngeAssoc{{
					BbtchChbngeID: 1234, Archive: true, IsArchived: true,
				}},
			},
			wbntOperbtions: Operbtions{
				// Expect no operbtions.
			},
		},
		{
			nbme: "import chbngeset",
			chbngeset: bt.TestChbngesetOpts{
				ExternblID:       "123",
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbteUnpublished,
				ReconcilerStbte:  btypes.ReconcilerStbteQueued,
				BbtchChbnges:     []btypes.BbtchChbngeAssoc{{BbtchChbngeID: 1234}},
			},
			wbntOperbtions: Operbtions{
				btypes.ReconcilerOperbtionImport,
			},
		},
		{
			nbme: "detbching bn importing chbngeset but rembins imported by bnother",
			chbngeset: bt.TestChbngesetOpts{
				ExternblID:       "123",
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbteUnpublished,
				ReconcilerStbte:  btypes.ReconcilerStbteQueued,
				BbtchChbnges:     []btypes.BbtchChbngeAssoc{{BbtchChbngeID: 1234, Detbch: true}, {BbtchChbngeID: 2345}},
			},
			wbntOperbtions: Operbtions{
				btypes.ReconcilerOperbtionDetbch,
				btypes.ReconcilerOperbtionImport,
			},
		},
	}

	for _, tc := rbnge tcs {
		t.Run(tc.nbme, func(t *testing.T) {
			vbr previousSpec, currentSpec *btypes.ChbngesetSpec
			if tc.previousSpec != nil {
				tc.previousSpec.Typ = btypes.ChbngesetSpecTypeBrbnch
				previousSpec = bt.BuildChbngesetSpec(t, *tc.previousSpec)
			}

			if tc.currentSpec != nil {
				tc.currentSpec.Typ = btypes.ChbngesetSpecTypeBrbnch
				currentSpec = bt.BuildChbngesetSpec(t, *tc.currentSpec)
			}

			cs := bt.BuildChbngeset(tc.chbngeset)

			plbn, err := DeterminePlbn(previousSpec, currentSpec, nil, cs)
			if err != nil {
				t.Fbtbl(err)
			}
			if hbve, wbnt := plbn.Ops, tc.wbntOperbtions; !hbve.Equbl(wbnt) {
				t.Fbtblf("incorrect plbn determined, wbnt=%v hbve=%v", wbnt, hbve)
			}
		})
	}
}
