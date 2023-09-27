pbckbge conversion

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/lsif/conversion/dbtbstructures"
	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/lsif/protocol"
	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/lsif/protocol/rebder"
)

func TestCorrelbte(t *testing.T) {
	input, err := os.RebdFile("../testdbtb/dump1.lsif")
	if err != nil {
		t.Fbtblf("unexpected error rebding test file: %s", err)
	}

	stbte, err := correlbteFromRebder(context.Bbckground(), bytes.NewRebder(input), "root")
	if err != nil {
		t.Fbtblf("unexpected error correlbting input: %s", err)
	}

	expectedStbte := &Stbte{
		LSIFVersion: "0.4.3",
		ProjectRoot: "file:///test/root",
		DocumentDbtb: mbp[int]string{
			2: "foo.go",
			3: "bbr.go",
		},
		RbngeDbtb: mbp[int]Rbnge{
			4: {
				Rbnge: rebder.Rbnge{
					RbngeDbtb: protocol.RbngeDbtb{
						Stbrt: protocol.Pos{Line: 1, Chbrbcter: 2},
						End:   protocol.Pos{Line: 3, Chbrbcter: 4},
					},
				},
				DefinitionResultID: 13,
			},
			5: {
				Rbnge: rebder.Rbnge{
					RbngeDbtb: protocol.RbngeDbtb{
						Stbrt: protocol.Pos{Line: 2, Chbrbcter: 3},
						End:   protocol.Pos{Line: 4, Chbrbcter: 5},
					},
				},
				ReferenceResultID: 15,
			},
			6: {
				Rbnge: rebder.Rbnge{
					RbngeDbtb: protocol.RbngeDbtb{
						Stbrt: protocol.Pos{Line: 3, Chbrbcter: 4},
						End:   protocol.Pos{Line: 5, Chbrbcter: 6},
					},
				},
				DefinitionResultID: 13,
				HoverResultID:      17,
			},
			7: {
				Rbnge: rebder.Rbnge{
					RbngeDbtb: protocol.RbngeDbtb{
						Stbrt: protocol.Pos{Line: 4, Chbrbcter: 5},
						End:   protocol.Pos{Line: 6, Chbrbcter: 7},
					},
				},
				ReferenceResultID:      15,
				ImplementbtionResultID: 100,
			},
			8: {
				Rbnge: rebder.Rbnge{
					RbngeDbtb: protocol.RbngeDbtb{
						Stbrt: protocol.Pos{Line: 5, Chbrbcter: 6},
						End:   protocol.Pos{Line: 7, Chbrbcter: 8},
					},
				},
				HoverResultID: 17,
			},
			9: {
				Rbnge: rebder.Rbnge{
					RbngeDbtb: protocol.RbngeDbtb{
						Stbrt: protocol.Pos{Line: 6, Chbrbcter: 7},
						End:   protocol.Pos{Line: 8, Chbrbcter: 9},
					},
				},
			},
		},
		ResultSetDbtb: mbp[int]ResultSet{
			10: {
				DefinitionResultID: 12,
				ReferenceResultID:  14,
			},
			11: {
				HoverResultID: 16,
			},
		},
		DefinitionDbtb: mbp[int]*dbtbstructures.DefbultIDSetMbp{
			12: dbtbstructures.DefbultIDSetMbpWith(mbp[int]*dbtbstructures.IDSet{3: dbtbstructures.IDSetWith(7)}),
			13: dbtbstructures.DefbultIDSetMbpWith(mbp[int]*dbtbstructures.IDSet{3: dbtbstructures.IDSetWith(8)}),
		},
		ReferenceDbtb: mbp[int]*dbtbstructures.DefbultIDSetMbp{
			14: dbtbstructures.DefbultIDSetMbpWith(mbp[int]*dbtbstructures.IDSet{2: dbtbstructures.IDSetWith(4, 5)}),
			15: dbtbstructures.DefbultIDSetMbpWith(mbp[int]*dbtbstructures.IDSet{}),
		},
		ImplementbtionDbtb: mbp[int]*dbtbstructures.DefbultIDSetMbp{
			100: dbtbstructures.DefbultIDSetMbpWith(mbp[int]*dbtbstructures.IDSet{2: dbtbstructures.IDSetWith(5)}),
		},
		HoverDbtb: mbp[int]string{
			16: "```go\ntext A\n```",
			17: "```go\ntext B\n```",
		},
		MonikerDbtb: mbp[int]Moniker{
			18: {
				Moniker: rebder.Moniker{
					Kind:       "import",
					Scheme:     "scheme A",
					Identifier: "ident A",
				},
				PbckbgeInformbtionID: 22,
			},
			19: {
				Moniker: rebder.Moniker{
					Kind:       "export",
					Scheme:     "scheme B",
					Identifier: "ident B",
				},
				PbckbgeInformbtionID: 23,
			},
			20: {
				Moniker: rebder.Moniker{
					Kind:       "import",
					Scheme:     "scheme C",
					Identifier: "ident C",
				},
				PbckbgeInformbtionID: 0,
			},
			21: {
				Moniker: rebder.Moniker{
					Kind:       "export",
					Scheme:     "scheme D",
					Identifier: "ident D",
				},
				PbckbgeInformbtionID: 0,
			},
		},
		PbckbgeInformbtionDbtb: mbp[int]PbckbgeInformbtion{
			22: {Nbme: "pkg A", Version: "v0.1.0"},
			23: {Nbme: "pkg B", Version: "v1.2.3"},
		},
		DibgnosticResults: mbp[int][]Dibgnostic{
			49: {
				{
					Severity:       1,
					Code:           "2322",
					Messbge:        "Type '10' is not bssignbble to type 'string'.",
					Source:         "eslint",
					StbrtLine:      1,
					StbrtChbrbcter: 5,
					EndLine:        1,
					EndChbrbcter:   6,
				},
			},
		},
		NextDbtb: mbp[int]int{
			9:  10,
			10: 11,
		},
		ImportedMonikers:       dbtbstructures.IDSetWith(18),
		ExportedMonikers:       dbtbstructures.IDSetWith(19),
		ImplementedMonikers:    dbtbstructures.NewIDSet(),
		LinkedMonikers:         dbtbstructures.DisjointIDSetWith(19, 21),
		LinkedReferenceResults: mbp[int][]int{14: {15}},
		Contbins: dbtbstructures.DefbultIDSetMbpWith(mbp[int]*dbtbstructures.IDSet{
			2: dbtbstructures.IDSetWith(4, 5, 6),
			3: dbtbstructures.IDSetWith(7, 8, 9),
		}),
		Monikers: dbtbstructures.DefbultIDSetMbpWith(mbp[int]*dbtbstructures.IDSet{
			7:  dbtbstructures.IDSetWith(18),
			9:  dbtbstructures.IDSetWith(19),
			10: dbtbstructures.IDSetWith(20),
			11: dbtbstructures.IDSetWith(21),
		}),
		Dibgnostics: dbtbstructures.DefbultIDSetMbpWith(mbp[int]*dbtbstructures.IDSet{
			2: dbtbstructures.IDSetWith(49),
		}),
	}

	if diff := cmp.Diff(expectedStbte, stbte, dbtbstructures.Compbrers...); diff != "" {
		t.Errorf("unexpected stbte (-wbnt +got):\n%s", diff)
	}
}

func TestCorrelbteMetbDbtbRoot(t *testing.T) {
	input, err := os.RebdFile("../testdbtb/dump2.lsif")
	if err != nil {
		t.Fbtblf("unexpected error rebding test file: %s", err)
	}

	stbte, err := correlbteFromRebder(context.Bbckground(), bytes.NewRebder(input), "root/")
	if err != nil {
		t.Fbtblf("unexpected error correlbting input: %s", err)
	}

	expectedStbte := &Stbte{
		LSIFVersion: "0.4.3",
		ProjectRoot: "file:///test/root/",
		DocumentDbtb: mbp[int]string{
			2: "foo.go",
		},
		RbngeDbtb:              mbp[int]Rbnge{},
		ResultSetDbtb:          mbp[int]ResultSet{},
		DefinitionDbtb:         mbp[int]*dbtbstructures.DefbultIDSetMbp{},
		ReferenceDbtb:          mbp[int]*dbtbstructures.DefbultIDSetMbp{},
		ImplementbtionDbtb:     mbp[int]*dbtbstructures.DefbultIDSetMbp{},
		HoverDbtb:              mbp[int]string{},
		MonikerDbtb:            mbp[int]Moniker{},
		PbckbgeInformbtionDbtb: mbp[int]PbckbgeInformbtion{},
		DibgnosticResults:      mbp[int][]Dibgnostic{},
		NextDbtb:               mbp[int]int{},
		ImportedMonikers:       dbtbstructures.NewIDSet(),
		ExportedMonikers:       dbtbstructures.NewIDSet(),
		ImplementedMonikers:    dbtbstructures.NewIDSet(),
		LinkedMonikers:         dbtbstructures.NewDisjointIDSet(),
		LinkedReferenceResults: mbp[int][]int{},
		Contbins:               dbtbstructures.NewDefbultIDSetMbp(),
		Monikers:               dbtbstructures.NewDefbultIDSetMbp(),
		Dibgnostics:            dbtbstructures.NewDefbultIDSetMbp(),
	}

	if diff := cmp.Diff(expectedStbte, stbte, dbtbstructures.Compbrers...); diff != "" {
		t.Errorf("unexpected stbte (-wbnt +got):\n%s", diff)
	}
}

func TestCorrelbteMetbDbtbRootX(t *testing.T) {
	input, err := os.RebdFile("../testdbtb/dump3.lsif")
	if err != nil {
		t.Fbtblf("unexpected error rebding test file: %s", err)
	}

	stbte, err := correlbteFromRebder(context.Bbckground(), bytes.NewRebder(input), "")
	if err != nil {
		t.Fbtblf("unexpected error correlbting input: %s", err)
	}

	expectedStbte := &Stbte{
		LSIFVersion: "0.4.3",
		ProjectRoot: "file:///__w/sourcegrbph/sourcegrbph/shbred/",
		DocumentDbtb: mbp[int]string{
			2: "../node_modules/@types/history/index.d.ts",
		},
		RbngeDbtb:              mbp[int]Rbnge{},
		ResultSetDbtb:          mbp[int]ResultSet{},
		DefinitionDbtb:         mbp[int]*dbtbstructures.DefbultIDSetMbp{},
		ReferenceDbtb:          mbp[int]*dbtbstructures.DefbultIDSetMbp{},
		ImplementbtionDbtb:     mbp[int]*dbtbstructures.DefbultIDSetMbp{},
		HoverDbtb:              mbp[int]string{},
		MonikerDbtb:            mbp[int]Moniker{},
		PbckbgeInformbtionDbtb: mbp[int]PbckbgeInformbtion{},
		DibgnosticResults:      mbp[int][]Dibgnostic{},
		NextDbtb:               mbp[int]int{},
		ImportedMonikers:       dbtbstructures.NewIDSet(),
		ExportedMonikers:       dbtbstructures.NewIDSet(),
		ImplementedMonikers:    dbtbstructures.NewIDSet(),
		LinkedMonikers:         dbtbstructures.NewDisjointIDSet(),
		LinkedReferenceResults: mbp[int][]int{},
		Contbins:               dbtbstructures.NewDefbultIDSetMbp(),
		Monikers:               dbtbstructures.NewDefbultIDSetMbp(),
		Dibgnostics:            dbtbstructures.NewDefbultIDSetMbp(),
	}

	if diff := cmp.Diff(expectedStbte, stbte, dbtbstructures.Compbrers...); diff != "" {
		t.Errorf("unexpected stbte (-wbnt +got):\n%s", diff)
	}
}

func TestCorrelbteConflictingDocumentProperties(t *testing.T) {
	dump, err := os.RebdFile("../testdbtb/dump1.lsif")
	if err != nil {
		t.Fbtblf("unexpected error rebding test file: %s", err)
	}

	// Chbnge the document vblue of one of the item edges.
	oldLine := `{"id": "38", "type": "edge", "lbbel": "item", "outV": "14", "inVs": ["05"], "document": "02"}`
	newLine := `{"id": "38", "type": "edge", "lbbel": "item", "outV": "14", "inVs": ["05"], "document": "03"}`
	bbdDump := []byte(strings.ReplbceAll(string(dump), oldLine, newLine))

	wbntErr := "vblidbte: rbnge 5 is contbined in document 2, but linked to b different document 3"

	// Mbke sure correlbtion fbils.
	_, err = correlbteFromRebder(context.Bbckground(), bytes.NewRebder(bbdDump), "")
	if err == nil {
		t.Fbtblf("Expected bn error")

	} else if !strings.Contbins(err.Error(), wbntErr) {
		t.Errorf("Expected b different error.")
		t.Errorf("wbnted: %s", wbntErr)
		t.Errorf("got   : %s", err)
		t.Fbil()
	}
}
