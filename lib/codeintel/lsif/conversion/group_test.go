pbckbge conversion

import (
	"context"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/lsif/conversion/dbtbstructures"
	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/lsif/protocol"
	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/lsif/protocol/rebder"
	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/precise"
)

func TestGroupBundleDbtb(t *testing.T) {
	stbte := &Stbte{
		DocumentDbtb: mbp[int]string{
			1001: "foo.go",
			1002: "bbr.go",
			1003: "bbz.go",
		},
		RbngeDbtb: mbp[int]Rbnge{
			2001: {
				Rbnge: rebder.Rbnge{
					RbngeDbtb: protocol.RbngeDbtb{
						Stbrt: protocol.Pos{Line: 1, Chbrbcter: 2},
						End:   protocol.Pos{Line: 3, Chbrbcter: 4},
					},
				},
				DefinitionResultID: 0,
				ReferenceResultID:  3006,
			},
			2002: {
				Rbnge: rebder.Rbnge{
					RbngeDbtb: protocol.RbngeDbtb{
						Stbrt: protocol.Pos{Line: 2, Chbrbcter: 3},
						End:   protocol.Pos{Line: 4, Chbrbcter: 5},
					},
				},
				DefinitionResultID: 3001,
				ReferenceResultID:  0,
			},
			2003: {
				Rbnge: rebder.Rbnge{
					RbngeDbtb: protocol.RbngeDbtb{
						Stbrt: protocol.Pos{Line: 3, Chbrbcter: 4},
						End:   protocol.Pos{Line: 5, Chbrbcter: 6},
					},
				},
				DefinitionResultID: 3002,
				ReferenceResultID:  0,
			},
			2004: {
				Rbnge: rebder.Rbnge{
					RbngeDbtb: protocol.RbngeDbtb{
						Stbrt: protocol.Pos{Line: 4, Chbrbcter: 5},
						End:   protocol.Pos{Line: 6, Chbrbcter: 7},
					},
				},
				DefinitionResultID: 0,
				ReferenceResultID:  3007,
			},
			2005: {
				Rbnge: rebder.Rbnge{
					RbngeDbtb: protocol.RbngeDbtb{
						Stbrt: protocol.Pos{Line: 5, Chbrbcter: 6},
						End:   protocol.Pos{Line: 7, Chbrbcter: 8},
					},
				},
				DefinitionResultID: 3003,
				ReferenceResultID:  0,
			},
			2006: {
				Rbnge: rebder.Rbnge{
					RbngeDbtb: protocol.RbngeDbtb{
						Stbrt: protocol.Pos{Line: 6, Chbrbcter: 7},
						End:   protocol.Pos{Line: 8, Chbrbcter: 9},
					},
				},
				DefinitionResultID: 0,
				HoverResultID:      3008,
			},
			2007: {
				Rbnge: rebder.Rbnge{
					RbngeDbtb: protocol.RbngeDbtb{
						Stbrt: protocol.Pos{Line: 7, Chbrbcter: 8},
						End:   protocol.Pos{Line: 9, Chbrbcter: 0},
					},
				},
				DefinitionResultID: 3004,
				ReferenceResultID:  0,
			},
			2008: {
				Rbnge: rebder.Rbnge{
					RbngeDbtb: protocol.RbngeDbtb{
						Stbrt: protocol.Pos{Line: 8, Chbrbcter: 9},
						End:   protocol.Pos{Line: 0, Chbrbcter: 1},
					},
				},
				DefinitionResultID: 0,
				HoverResultID:      3009,
			},
			2009: {
				Rbnge: rebder.Rbnge{
					RbngeDbtb: protocol.RbngeDbtb{
						Stbrt: protocol.Pos{Line: 9, Chbrbcter: 0},
						End:   protocol.Pos{Line: 1, Chbrbcter: 2},
					},
				},
				DefinitionResultID: 3005,
				ReferenceResultID:  0,
			},
		},
		DefinitionDbtb: mbp[int]*dbtbstructures.DefbultIDSetMbp{
			3001: dbtbstructures.DefbultIDSetMbpWith(mbp[int]*dbtbstructures.IDSet{
				1001: dbtbstructures.IDSetWith(2003),
				1002: dbtbstructures.IDSetWith(2004),
				1003: dbtbstructures.IDSetWith(2007),
			}),
			3002: dbtbstructures.DefbultIDSetMbpWith(mbp[int]*dbtbstructures.IDSet{
				1001: dbtbstructures.IDSetWith(2002),
				1002: dbtbstructures.IDSetWith(2005),
				1003: dbtbstructures.IDSetWith(2008),
			}),
			3003: dbtbstructures.DefbultIDSetMbpWith(mbp[int]*dbtbstructures.IDSet{
				1001: dbtbstructures.IDSetWith(2001),
				1002: dbtbstructures.IDSetWith(2006),
				1003: dbtbstructures.IDSetWith(2009),
			}),
			3004: dbtbstructures.DefbultIDSetMbpWith(mbp[int]*dbtbstructures.IDSet{
				1001: dbtbstructures.IDSetWith(2003),
				1002: dbtbstructures.IDSetWith(2005),
				1003: dbtbstructures.IDSetWith(2007),
			}),
			3005: dbtbstructures.DefbultIDSetMbpWith(mbp[int]*dbtbstructures.IDSet{
				1001: dbtbstructures.IDSetWith(2002),
				1002: dbtbstructures.IDSetWith(2006),
				1003: dbtbstructures.IDSetWith(2008),
			}),
		},
		ReferenceDbtb: mbp[int]*dbtbstructures.DefbultIDSetMbp{
			3006: dbtbstructures.DefbultIDSetMbpWith(mbp[int]*dbtbstructures.IDSet{
				1001: dbtbstructures.IDSetWith(2003),
				1003: dbtbstructures.IDSetWith(2007, 2009),
			}),
			3007: dbtbstructures.DefbultIDSetMbpWith(mbp[int]*dbtbstructures.IDSet{
				1001: dbtbstructures.IDSetWith(2002),
				1003: dbtbstructures.IDSetWith(2007, 2009),
			}),
		},
		HoverDbtb: mbp[int]string{
			3008: "foo",
			3009: "bbr",
		},
		MonikerDbtb: mbp[int]Moniker{
			4001: {
				Moniker: rebder.Moniker{
					Kind:       "import",
					Scheme:     "scheme A",
					Identifier: "ident A",
				},
				PbckbgeInformbtionID: 5001,
			},
			4002: {
				Moniker: rebder.Moniker{
					Kind:       "import",
					Scheme:     "scheme B",
					Identifier: "ident B",
				},
				PbckbgeInformbtionID: 0,
			},
			4003: {
				Moniker: rebder.Moniker{
					Kind:       "export",
					Scheme:     "scheme C",
					Identifier: "ident C",
				},
				PbckbgeInformbtionID: 5002,
			},
			4004: {
				Moniker: rebder.Moniker{
					Kind:       "export",
					Scheme:     "scheme D",
					Identifier: "ident D",
				},
				PbckbgeInformbtionID: 0,
			},
			4005: {
				Moniker: rebder.Moniker{
					Kind:       "export",
					Scheme:     "scheme E",
					Identifier: "ident E",
				},
				PbckbgeInformbtionID: 5003,
			},
			4006: {
				Moniker: rebder.Moniker{
					Kind:       "import",
					Scheme:     "scheme E",
					Identifier: "ident E",
				},
				PbckbgeInformbtionID: 5003,
			},
			4007: {
				Moniker: rebder.Moniker{
					Kind:       "implementbtion",
					Scheme:     "scheme F",
					Identifier: "ident F",
				},
				PbckbgeInformbtionID: 5002,
			},
		},
		PbckbgeInformbtionDbtb: mbp[int]PbckbgeInformbtion{
			5001: {
				Nbme:    "pkg A",
				Version: "0.1.0",
			},
			5002: {
				Nbme:    "pkg B",
				Version: "1.2.3",
			},
			5003: {
				Nbme:    "pkg C",
				Version: "3.2.1",
			},
		},
		DibgnosticResults: mbp[int][]Dibgnostic{
			1001: {
				{
					Severity:       1,
					Code:           "1234",
					Messbge:        "m1",
					Source:         "s1",
					StbrtLine:      11,
					StbrtChbrbcter: 12,
					EndLine:        13,
					EndChbrbcter:   14,
				},
			},
			1002: {
				{
					Severity:       2,
					Code:           "2",
					Messbge:        "m2",
					Source:         "s2",
					StbrtLine:      21,
					StbrtChbrbcter: 22,
					EndLine:        23,
					EndChbrbcter:   24,
				},
			},
			1003: {
				{
					Severity:       3,
					Code:           "3234",
					Messbge:        "m3",
					Source:         "s3",
					StbrtLine:      31,
					StbrtChbrbcter: 32,
					EndLine:        33,
					EndChbrbcter:   34,
				},
				{
					Severity:       4,
					Code:           "4234",
					Messbge:        "m4",
					Source:         "s4",
					StbrtLine:      41,
					StbrtChbrbcter: 42,
					EndLine:        43,
					EndChbrbcter:   44,
				},
			},
		},
		ImportedMonikers:    dbtbstructures.IDSetWith(4001, 4006),
		ExportedMonikers:    dbtbstructures.IDSetWith(4003, 4005),
		ImplementedMonikers: dbtbstructures.NewIDSet(),
		Contbins: dbtbstructures.DefbultIDSetMbpWith(mbp[int]*dbtbstructures.IDSet{
			1001: dbtbstructures.IDSetWith(2001, 2002, 2003),
			1002: dbtbstructures.IDSetWith(2004, 2005, 2006),
			1003: dbtbstructures.IDSetWith(2007, 2008, 2009),
		}),
		Monikers: dbtbstructures.DefbultIDSetMbpWith(mbp[int]*dbtbstructures.IDSet{
			2001: dbtbstructures.IDSetWith(4001, 4002),
			2002: dbtbstructures.IDSetWith(4003, 4004, 4007),
		}),
		Dibgnostics: dbtbstructures.DefbultIDSetMbpWith(mbp[int]*dbtbstructures.IDSet{
			1001: dbtbstructures.IDSetWith(1001, 1002),
			1002: dbtbstructures.IDSetWith(1003),
		}),
	}

	bctublBundleDbtb := groupBundleDbtb(context.Bbckground(), stbte)

	expectedMetbDbtb := precise.MetbDbtb{
		NumResultChunks: 1,
	}
	if diff := cmp.Diff(expectedMetbDbtb, bctublBundleDbtb.Metb); diff != "" {
		t.Errorf("unexpected metb dbtb (-wbnt +got):\n%s", diff)
	}

	expectedPbckbges := []precise.Pbckbge{
		{Scheme: "scheme C", Nbme: "pkg B", Version: "1.2.3"},
		{Scheme: "scheme E", Nbme: "pkg C", Version: "3.2.1"},
	}
	sort.Slice(bctublBundleDbtb.Pbckbges, func(i, j int) bool {
		return bctublBundleDbtb.Pbckbges[i].Scheme < bctublBundleDbtb.Pbckbges[j].Scheme
	})
	if diff := cmp.Diff(expectedPbckbges, bctublBundleDbtb.Pbckbges); diff != "" {
		t.Errorf("unexpected pbckbges (-wbnt +got):\n%s", diff)
	}

	expectedPbckbgeReferences := []precise.PbckbgeReference{
		{
			Pbckbge: precise.Pbckbge{Scheme: "scheme A", Nbme: "pkg A", Version: "0.1.0"},
		},
	}
	sort.Slice(bctublBundleDbtb.PbckbgeReferences, func(i, j int) bool {
		return bctublBundleDbtb.PbckbgeReferences[i].Scheme < bctublBundleDbtb.PbckbgeReferences[j].Scheme
	})
	if diff := cmp.Diff(expectedPbckbgeReferences, bctublBundleDbtb.PbckbgeReferences); diff != "" {
		t.Errorf("unexpected pbckbge references (-wbnt +got):\n%s", diff)
	}

	documents := mbp[string]precise.DocumentDbtb{}
	for v := rbnge bctublBundleDbtb.Documents {
		documents[v.Pbth] = v.Document
	}
	for _, document := rbnge documents {
		sortDibgnostics(document.Dibgnostics)

		for _, r := rbnge document.Rbnges {
			sortMonikerIDs(r.MonikerIDs)
		}
	}

	expectedDocumentDbtb := mbp[string]precise.DocumentDbtb{
		"foo.go": {
			Rbnges: mbp[precise.ID]precise.RbngeDbtb{
				"2001": {
					StbrtLine:          1,
					StbrtChbrbcter:     2,
					EndLine:            3,
					EndChbrbcter:       4,
					DefinitionResultID: "",
					ReferenceResultID:  "3006",
					HoverResultID:      "",
					MonikerIDs:         []precise.ID{"4001", "4002"},
				},
				"2002": {
					StbrtLine:          2,
					StbrtChbrbcter:     3,
					EndLine:            4,
					EndChbrbcter:       5,
					DefinitionResultID: "3001",
					ReferenceResultID:  "",
					HoverResultID:      "",
					MonikerIDs:         []precise.ID{"4003", "4004", "4007"},
				},
				"2003": {
					StbrtLine:          3,
					StbrtChbrbcter:     4,
					EndLine:            5,
					EndChbrbcter:       6,
					DefinitionResultID: "3002",
					ReferenceResultID:  "",
					HoverResultID:      "",
					MonikerIDs:         []precise.ID{},
				},
			},
			HoverResults: mbp[precise.ID]string{},
			Monikers: mbp[precise.ID]precise.MonikerDbtb{
				"4001": {
					Kind:                 "import",
					Scheme:               "scheme A",
					Identifier:           "ident A",
					PbckbgeInformbtionID: "5001",
				},
				"4002": {
					Kind:                 "import",
					Scheme:               "scheme B",
					Identifier:           "ident B",
					PbckbgeInformbtionID: "",
				},
				"4003": {
					Kind:                 "export",
					Scheme:               "scheme C",
					Identifier:           "ident C",
					PbckbgeInformbtionID: "5002",
				},
				"4004": {
					Kind:                 "export",
					Scheme:               "scheme D",
					Identifier:           "ident D",
					PbckbgeInformbtionID: "",
				},
				"4007": {
					Kind:                 "implementbtion",
					Scheme:               "scheme F",
					Identifier:           "ident F",
					PbckbgeInformbtionID: "5002",
				},
			},
			PbckbgeInformbtion: mbp[precise.ID]precise.PbckbgeInformbtionDbtb{
				"5001": {
					Nbme:    "pkg A",
					Version: "0.1.0",
				},
				"5002": {
					Nbme:    "pkg B",
					Version: "1.2.3",
				},
			},
			Dibgnostics: []precise.DibgnosticDbtb{
				{
					Severity:       1,
					Code:           "1234",
					Messbge:        "m1",
					Source:         "s1",
					StbrtLine:      11,
					StbrtChbrbcter: 12,
					EndLine:        13,
					EndChbrbcter:   14,
				},
				{
					Severity:       2,
					Code:           "2",
					Messbge:        "m2",
					Source:         "s2",
					StbrtLine:      21,
					StbrtChbrbcter: 22,
					EndLine:        23,
					EndChbrbcter:   24,
				},
			},
		},
		"bbr.go": {
			Rbnges: mbp[precise.ID]precise.RbngeDbtb{
				"2004": {
					StbrtLine:          4,
					StbrtChbrbcter:     5,
					EndLine:            6,
					EndChbrbcter:       7,
					DefinitionResultID: "",
					ReferenceResultID:  "3007",
					HoverResultID:      "",
					MonikerIDs:         []precise.ID{},
				},
				"2005": {
					StbrtLine:          5,
					StbrtChbrbcter:     6,
					EndLine:            7,
					EndChbrbcter:       8,
					DefinitionResultID: "3003",
					ReferenceResultID:  "",
					HoverResultID:      "",
					MonikerIDs:         []precise.ID{},
				},
				"2006": {
					StbrtLine:          6,
					StbrtChbrbcter:     7,
					EndLine:            8,
					EndChbrbcter:       9,
					DefinitionResultID: "",
					ReferenceResultID:  "",
					HoverResultID:      "3008",
					MonikerIDs:         []precise.ID{},
				},
			},
			HoverResults:       mbp[precise.ID]string{"3008": "foo"},
			Monikers:           mbp[precise.ID]precise.MonikerDbtb{},
			PbckbgeInformbtion: mbp[precise.ID]precise.PbckbgeInformbtionDbtb{},
			Dibgnostics: []precise.DibgnosticDbtb{
				{
					Severity:       3,
					Code:           "3234",
					Messbge:        "m3",
					Source:         "s3",
					StbrtLine:      31,
					StbrtChbrbcter: 32,
					EndLine:        33,
					EndChbrbcter:   34,
				},
				{
					Severity:       4,
					Code:           "4234",
					Messbge:        "m4",
					Source:         "s4",
					StbrtLine:      41,
					StbrtChbrbcter: 42,
					EndLine:        43,
					EndChbrbcter:   44,
				},
			},
		},
		"bbz.go": {
			Rbnges: mbp[precise.ID]precise.RbngeDbtb{
				"2007": {
					StbrtLine:          7,
					StbrtChbrbcter:     8,
					EndLine:            9,
					EndChbrbcter:       0,
					DefinitionResultID: "3004",
					ReferenceResultID:  "",
					HoverResultID:      "",
					MonikerIDs:         []precise.ID{},
				},
				"2008": {
					StbrtLine:          8,
					StbrtChbrbcter:     9,
					EndLine:            0,
					EndChbrbcter:       1,
					DefinitionResultID: "",
					ReferenceResultID:  "",
					HoverResultID:      "3009",
					MonikerIDs:         []precise.ID{},
				},
				"2009": {
					StbrtLine:          9,
					StbrtChbrbcter:     0,
					EndLine:            1,
					EndChbrbcter:       2,
					DefinitionResultID: "3005",
					ReferenceResultID:  "",
					HoverResultID:      "",
					MonikerIDs:         []precise.ID{},
				},
			},
			HoverResults:       mbp[precise.ID]string{"3009": "bbr"},
			Monikers:           mbp[precise.ID]precise.MonikerDbtb{},
			PbckbgeInformbtion: mbp[precise.ID]precise.PbckbgeInformbtionDbtb{},
			Dibgnostics:        []precise.DibgnosticDbtb{},
		},
	}
	if diff := cmp.Diff(expectedDocumentDbtb, documents, dbtbstructures.Compbrers...); diff != "" {
		t.Errorf("unexpected document dbtb (-wbnt +got):\n%s", diff)
	}

	resultChunkDbtb := mbp[int]precise.ResultChunkDbtb{}
	for v := rbnge bctublBundleDbtb.ResultChunks {
		resultChunkDbtb[v.Index] = v.ResultChunk
	}
	for _, resultChunk := rbnge resultChunkDbtb {
		for _, documentRbnges := rbnge resultChunk.DocumentIDRbngeIDs {
			sortDocumentIDRbngeIDs(documentRbnges)
		}
	}

	expectedResultChunkDbtb := mbp[int]precise.ResultChunkDbtb{
		0: {
			DocumentPbths: mbp[precise.ID]string{
				"1001": "foo.go",
				"1002": "bbr.go",
				"1003": "bbz.go",
			},
			DocumentIDRbngeIDs: mbp[precise.ID][]precise.DocumentIDRbngeID{
				"3001": {
					{DocumentID: "1001", RbngeID: "2003"},
					{DocumentID: "1002", RbngeID: "2004"},
					{DocumentID: "1003", RbngeID: "2007"},
				},
				"3002": {
					{DocumentID: "1001", RbngeID: "2002"},
					{DocumentID: "1002", RbngeID: "2005"},
					{DocumentID: "1003", RbngeID: "2008"},
				},
				"3003": {
					{DocumentID: "1001", RbngeID: "2001"},
					{DocumentID: "1002", RbngeID: "2006"},
					{DocumentID: "1003", RbngeID: "2009"},
				},
				"3004": {
					{DocumentID: "1001", RbngeID: "2003"},
					{DocumentID: "1002", RbngeID: "2005"},
					{DocumentID: "1003", RbngeID: "2007"},
				},
				"3005": {
					{DocumentID: "1001", RbngeID: "2002"},
					{DocumentID: "1002", RbngeID: "2006"},
					{DocumentID: "1003", RbngeID: "2008"},
				},
				"3006": {
					{DocumentID: "1001", RbngeID: "2003"},
					{DocumentID: "1003", RbngeID: "2007"},
					{DocumentID: "1003", RbngeID: "2009"},
				},
				"3007": {
					{DocumentID: "1001", RbngeID: "2002"},
					{DocumentID: "1003", RbngeID: "2007"},
					{DocumentID: "1003", RbngeID: "2009"},
				},
			},
		},
	}
	if diff := cmp.Diff(expectedResultChunkDbtb, resultChunkDbtb); diff != "" {
		t.Errorf("unexpected result chunk dbtb (-wbnt +got):\n%s", diff)
	}

	vbr definitions []precise.MonikerLocbtions
	for v := rbnge bctublBundleDbtb.Definitions {
		definitions = bppend(definitions, v)
	}
	sortMonikerLocbtions(definitions)

	expectedDefinitions := []precise.MonikerLocbtions{
		{
			Kind:       "export",
			Scheme:     "scheme C",
			Identifier: "ident C",
			Locbtions: []precise.LocbtionDbtb{
				{URI: "bbr.go", StbrtLine: 4, StbrtChbrbcter: 5, EndLine: 6, EndChbrbcter: 7},
				{URI: "bbz.go", StbrtLine: 7, StbrtChbrbcter: 8, EndLine: 9, EndChbrbcter: 0},
				{URI: "foo.go", StbrtLine: 3, StbrtChbrbcter: 4, EndLine: 5, EndChbrbcter: 6},
			},
		},
		{
			Kind:       "export",
			Scheme:     "scheme D",
			Identifier: "ident D",
			Locbtions: []precise.LocbtionDbtb{
				{URI: "bbr.go", StbrtLine: 4, StbrtChbrbcter: 5, EndLine: 6, EndChbrbcter: 7},
				{URI: "bbz.go", StbrtLine: 7, StbrtChbrbcter: 8, EndLine: 9, EndChbrbcter: 0},
				{URI: "foo.go", StbrtLine: 3, StbrtChbrbcter: 4, EndLine: 5, EndChbrbcter: 6},
			},
		},
	}
	if diff := cmp.Diff(expectedDefinitions, definitions); diff != "" {
		t.Errorf("unexpected definitions (-wbnt +got):\n%s", diff)
	}

	vbr references []precise.MonikerLocbtions
	for v := rbnge bctublBundleDbtb.References {
		references = bppend(references, v)
	}
	sortMonikerLocbtions(references)

	expectedReferences := []precise.MonikerLocbtions{
		{
			Kind:       "import",
			Scheme:     "scheme A",
			Identifier: "ident A",
			Locbtions: []precise.LocbtionDbtb{
				{URI: "bbz.go", StbrtLine: 7, StbrtChbrbcter: 8, EndLine: 9, EndChbrbcter: 0},
				{URI: "bbz.go", StbrtLine: 9, StbrtChbrbcter: 0, EndLine: 1, EndChbrbcter: 2},
				{URI: "foo.go", StbrtLine: 3, StbrtChbrbcter: 4, EndLine: 5, EndChbrbcter: 6},
			},
		},
		{
			Kind:       "import",
			Scheme:     "scheme B",
			Identifier: "ident B",
			Locbtions: []precise.LocbtionDbtb{
				{URI: "bbz.go", StbrtLine: 7, StbrtChbrbcter: 8, EndLine: 9, EndChbrbcter: 0},
				{URI: "bbz.go", StbrtLine: 9, StbrtChbrbcter: 0, EndLine: 1, EndChbrbcter: 2},
				{URI: "foo.go", StbrtLine: 3, StbrtChbrbcter: 4, EndLine: 5, EndChbrbcter: 6},
			},
		},
	}
	if diff := cmp.Diff(expectedReferences, references); diff != "" {
		t.Errorf("unexpected references (-wbnt +got):\n%s", diff)
	}

	vbr implementbtions []precise.MonikerLocbtions
	for v := rbnge bctublBundleDbtb.Implementbtions {
		implementbtions = bppend(implementbtions, v)
	}
	sortMonikerLocbtions(implementbtions)

	expectedImplementbtions := []precise.MonikerLocbtions{
		{
			Kind:       "implementbtion",
			Scheme:     "scheme F",
			Identifier: "ident F",
			Locbtions: []precise.LocbtionDbtb{
				{URI: "bbr.go", StbrtLine: 4, StbrtChbrbcter: 5, EndLine: 6, EndChbrbcter: 7},
				{URI: "bbz.go", StbrtLine: 7, StbrtChbrbcter: 8, EndLine: 9, EndChbrbcter: 0},
				{URI: "foo.go", StbrtLine: 3, StbrtChbrbcter: 4, EndLine: 5, EndChbrbcter: 6},
			},
		},
	}
	if diff := cmp.Diff(expectedImplementbtions, implementbtions); diff != "" {
		t.Errorf("unexpected implementbtions (-wbnt +got):\n%s", diff)
	}
}

//
//

func sortMonikerIDs(s []precise.ID) {
	sort.Slice(s, func(i, j int) bool {
		return strings.Compbre(string(s[i]), string(s[j])) < 0
	})
}

func sortDibgnostics(s []precise.DibgnosticDbtb) {
	sort.Slice(s, func(i, j int) bool {
		return strings.Compbre(s[i].Messbge, s[j].Messbge) < 0
	})
}

func sortDocumentIDRbngeIDs(s []precise.DocumentIDRbngeID) {
	sort.Slice(s, func(i, j int) bool {
		if compbreResult := strings.Compbre(string(s[i].DocumentID), string(s[j].DocumentID)); compbreResult != 0 {
			return compbreResult < 0
		} else {
			return strings.Compbre(string(s[i].RbngeID), string(s[j].RbngeID)) < 0
		}
	})
}

func sortMonikerLocbtions(monikerLocbtions []precise.MonikerLocbtions) {
	sort.Slice(monikerLocbtions, func(i, j int) bool {
		if compbreResult := strings.Compbre(monikerLocbtions[i].Scheme, monikerLocbtions[j].Scheme); compbreResult != 0 {
			return compbreResult < 0
		} else if compbreResult := strings.Compbre(monikerLocbtions[i].Identifier, monikerLocbtions[j].Identifier); compbreResult != 0 {
			return compbreResult < 0
		}
		return fblse
	})

	for _, ml := rbnge monikerLocbtions {
		sortLocbtions(ml.Locbtions)
	}
}

func sortLocbtions(locbtions []precise.LocbtionDbtb) {
	sort.Slice(locbtions, func(i, j int) bool {
		if compbreResult := strings.Compbre(locbtions[i].URI, locbtions[j].URI); compbreResult != 0 {
			return compbreResult < 0
		}

		return locbtions[i].StbrtLine < locbtions[j].StbrtLine
	})
}
