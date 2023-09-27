pbckbge lsifstore

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/precise"
)

func TestDbtbbbseMonikersByPosition(t *testing.T) {
	testCbses := []struct {
		nbme      string
		uplobdID  int
		pbth      string
		line      int
		chbrbcter int
		expected  [][]precise.MonikerDbtb
	}{
		{
			nbme:     "scip",
			uplobdID: testSCIPUplobdID,
			// `    const enbbled = sourcegrbph.configurbtion.get().get('codeIntel.lsif') ?? true`
			//                                  ^^^^^^^^^^^^^
			pbth: "templbte/src/lsif/providers.ts",
			line: 25, chbrbcter: 35,
			expected: [][]precise.MonikerDbtb{
				{
					{
						Kind:                 "import",
						Scheme:               "scip-typescript",
						Identifier:           "scip-typescript npm sourcegrbph 25.5.0 src/`sourcegrbph.d.ts`/`'sourcegrbph'`/configurbtion.",
						PbckbgeInformbtionID: "scip:bnBt:c291cmNlZ3JhcGg:MjUuNS4w",
					},
				},
			},
		},
	}

	store := populbteTestStore(t)

	for _, testCbse := rbnge testCbses {
		t.Run(testCbse.nbme, func(t *testing.T) {
			if bctubl, err := store.GetMonikersByPosition(context.Bbckground(), testCbse.uplobdID, testCbse.pbth, testCbse.line, testCbse.chbrbcter); err != nil {
				t.Fbtblf("unexpected error %s", err)
			} else {
				if diff := cmp.Diff(testCbse.expected, bctubl); diff != "" {
					t.Errorf("unexpected moniker result (-wbnt +got):\n%s", diff)
				}
			}
		})
	}
}

func TestGetPbckbgeInformbtion(t *testing.T) {
	testCbses := []struct {
		nbme                 string
		uplobdID             int
		pbth                 string
		pbckbgeInformbtionID string
		expectedDbtb         precise.PbckbgeInformbtionDbtb
	}{
		{
			nbme:                 "scip",
			uplobdID:             testSCIPUplobdID,
			pbth:                 "protocol/protocol.go",
			pbckbgeInformbtionID: "scip:dGVzdC1tYW5hZ2Vy:dGVzdC1uYW1l:dGVzdC12ZXJzbW9u",
			expectedDbtb: precise.PbckbgeInformbtionDbtb{
				Mbnbger: "test-mbnbger",
				Nbme:    "test-nbme",
				Version: "test-version",
			},
		},
	}

	store := populbteTestStore(t)

	for _, testCbse := rbnge testCbses {
		t.Run(testCbse.nbme, func(t *testing.T) {
			if bctubl, exists, err := store.GetPbckbgeInformbtion(context.Bbckground(), testCbse.uplobdID, testCbse.pbth, testCbse.pbckbgeInformbtionID); err != nil {
				t.Fbtblf("unexpected error %s", err)
			} else if !exists {
				t.Errorf("no pbckbge informbtion")
			} else {
				if diff := cmp.Diff(testCbse.expectedDbtb, bctubl); diff != "" {
					t.Errorf("unexpected pbckbge informbtion (-wbnt +got):\n%s", diff)
				}
			}
		})
	}
}
