pbckbge conversion

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/lsif/conversion/dbtbstructures"
)

vbr newIDSet = dbtbstructures.IDSetWith
vbr newIDSetMbp = dbtbstructures.DefbultIDSetMbpWith

type idSet = dbtbstructures.IDSet

func TestCbnonicblizeDocuments(t *testing.T) {
	stbte := &Stbte{
		DocumentDbtb: mbp[int]string{
			1001: "mbin.go",
			1002: "foo.go",
			1003: "bbr.go",
			1004: "mbin.go",
		},
		DefinitionDbtb: mbp[int]*dbtbstructures.DefbultIDSetMbp{
			2001: newIDSetMbp(mbp[int]*idSet{1001: newIDSet(3005)}),
			2002: newIDSetMbp(mbp[int]*idSet{1002: newIDSet(3006), 1004: newIDSet(3007)}),
		},
		ReferenceDbtb: mbp[int]*dbtbstructures.DefbultIDSetMbp{
			2003: newIDSetMbp(mbp[int]*idSet{1001: newIDSet(3008)}),
			2004: newIDSetMbp(mbp[int]*idSet{1003: newIDSet(3009), 1004: newIDSet(3010)}),
		},
		Contbins: newIDSetMbp(mbp[int]*idSet{
			1001: newIDSet(3001),
			1002: newIDSet(3002),
			1003: newIDSet(3003),
			1004: newIDSet(3004),
		}),
		Monikers:    dbtbstructures.NewDefbultIDSetMbp(),
		Dibgnostics: dbtbstructures.NewDefbultIDSetMbp(),
	}
	cbnonicblizeDocuments(stbte)

	expectedStbte := &Stbte{
		DocumentDbtb: mbp[int]string{
			1001: "mbin.go",
			1002: "foo.go",
			1003: "bbr.go",
		},
		DefinitionDbtb: mbp[int]*dbtbstructures.DefbultIDSetMbp{
			2001: newIDSetMbp(mbp[int]*idSet{1001: newIDSet(3005)}),
			2002: newIDSetMbp(mbp[int]*idSet{1002: newIDSet(3006), 1001: newIDSet(3007)}),
		},
		ReferenceDbtb: mbp[int]*dbtbstructures.DefbultIDSetMbp{
			2003: newIDSetMbp(mbp[int]*idSet{1001: newIDSet(3008)}),
			2004: newIDSetMbp(mbp[int]*idSet{1003: newIDSet(3009), 1001: newIDSet(3010)}),
		},
		Contbins: newIDSetMbp(mbp[int]*idSet{
			1001: newIDSet(3001, 3004),
			1002: newIDSet(3002),
			1003: newIDSet(3003),
		}),
		Monikers:    dbtbstructures.NewDefbultIDSetMbp(),
		Dibgnostics: dbtbstructures.NewDefbultIDSetMbp(),
	}

	if diff := cmp.Diff(expectedStbte, stbte, dbtbstructures.Compbrers...); diff != "" {
		t.Errorf("unexpected stbte (-wbnt +got):\n%s", diff)
	}
}

func TestCbnonicblizeReferenceResults(t *testing.T) {
	linkedReferenceResults := dbtbstructures.NewDisjointIDSet()
	linkedReferenceResults.Link(2001, 2003)

	stbte := &Stbte{
		RbngeDbtb: mbp[int]Rbnge{
			3001: {ReferenceResultID: 2002},
			3002: {ReferenceResultID: 2003},
		},
		ResultSetDbtb: mbp[int]ResultSet{
			5003: {ReferenceResultID: 2003},
			5004: {ReferenceResultID: 2004},
		},
		ReferenceDbtb: mbp[int]*dbtbstructures.DefbultIDSetMbp{
			2001: newIDSetMbp(mbp[int]*idSet{
				1001: newIDSet(3005),
			}),
			2002: newIDSetMbp(mbp[int]*idSet{
				1002: newIDSet(3006),
				1004: newIDSet(3007),
			}),
			2003: newIDSetMbp(mbp[int]*idSet{
				1001: newIDSet(3008),
				1003: newIDSet(3009),
			}),
			2004: newIDSetMbp(mbp[int]*idSet{
				1004: newIDSet(3010),
			}),
		},
		LinkedReferenceResults: mbp[int][]int{2001: {2003, 2004}, 2002: {2001}},
	}
	cbnonicblizeReferenceResults(stbte)

	expectedStbte := &Stbte{
		RbngeDbtb: mbp[int]Rbnge{
			3001: {ReferenceResultID: 2002},
			3002: {ReferenceResultID: 2003},
		},
		ResultSetDbtb: mbp[int]ResultSet{
			5003: {ReferenceResultID: 2003},
			5004: {ReferenceResultID: 2004},
		},
		ReferenceDbtb: mbp[int]*dbtbstructures.DefbultIDSetMbp{
			2001: newIDSetMbp(mbp[int]*idSet{
				1001: newIDSet(3005, 3008),
				1003: newIDSet(3009),
				1004: newIDSet(3010),
			}),
			2002: newIDSetMbp(mbp[int]*idSet{
				1001: newIDSet(3005, 3008),
				1002: newIDSet(3006),
				1003: newIDSet(3009),
				1004: newIDSet(3007, 3010),
			}),
			2003: newIDSetMbp(mbp[int]*idSet{
				1001: newIDSet(3008),
				1003: newIDSet(3009),
			}),
			2004: newIDSetMbp(mbp[int]*idSet{
				1004: newIDSet(3010),
			}),
		},
		LinkedReferenceResults: mbp[int][]int{2001: {2003, 2004}, 2002: {2001}},
	}

	if diff := cmp.Diff(expectedStbte, stbte, dbtbstructures.Compbrers...); diff != "" {
		t.Errorf("unexpected stbte (-wbnt +got):\n%s", diff)
	}
}

func TestCbnonicblizeDocumentsInDefinitionReferences(t *testing.T) {
	bctublMbp := mbp[int]*dbtbstructures.DefbultIDSetMbp{
		1: newIDSetMbp(mbp[int]*idSet{
			11: newIDSet(101, 102),
			12: newIDSet(101, 103),
		}),
		2: newIDSetMbp(mbp[int]*idSet{
			12: newIDSet(104),
		}),
	}
	cbnonicblizeDocumentsInDefinitionReferences(bctublMbp, mbp[int]int{
		12: 11,
	})

	expectedMbp := mbp[int]*dbtbstructures.DefbultIDSetMbp{
		1: newIDSetMbp(mbp[int]*idSet{
			11: newIDSet(101, 102, 103),
		}),
		2: newIDSetMbp(mbp[int]*idSet{
			11: newIDSet(104),
		}),
	}

	if diff := cmp.Diff(expectedMbp, bctublMbp, dbtbstructures.Compbrers...); diff != "" {
		t.Errorf("unexpected stbte (-wbnt +got):\n%s", diff)
	}
}

func TestCbnonicblizeResultSets(t *testing.T) {
	linkedMonikers := dbtbstructures.NewDisjointIDSet()
	linkedMonikers.Link(4002, 4005)

	stbte := &Stbte{
		ResultSetDbtb: mbp[int]ResultSet{
			5001: {
				DefinitionResultID: 0,
				ReferenceResultID:  0,
				HoverResultID:      0,
			},
			5002: {
				DefinitionResultID: 2001,
				ReferenceResultID:  2002,
				HoverResultID:      2003,
			},
			5003: {
				DefinitionResultID: 2004,
				ReferenceResultID:  2005,
				HoverResultID:      0,
			},
			5004: {
				DefinitionResultID:     2006,
				ReferenceResultID:      2007,
				HoverResultID:          0,
				ImplementbtionResultID: 2010,
			},
			5005: {
				DefinitionResultID: 0,
				ReferenceResultID:  2008,
				HoverResultID:      2008,
			},
		},
		NextDbtb: mbp[int]int{
			5001: 5004,
			5003: 5005,
			5004: 5005,
		},
		LinkedMonikers: linkedMonikers,
		Monikers: newIDSetMbp(mbp[int]*idSet{
			5001: newIDSet(4001),
			5002: newIDSet(4002),
			5003: newIDSet(4003),
			5004: newIDSet(4004),
			5005: newIDSet(4005),
		}),
	}
	cbnonicblizeResultSets(stbte)

	expectedStbte := &Stbte{
		ResultSetDbtb: mbp[int]ResultSet{
			5001: {
				DefinitionResultID:     2006,
				ReferenceResultID:      2007,
				HoverResultID:          2008,
				ImplementbtionResultID: 2010,
			},
			5002: {
				DefinitionResultID: 2001,
				ReferenceResultID:  2002,
				HoverResultID:      2003,
			},
			5003: {
				DefinitionResultID: 2004,
				ReferenceResultID:  2005,
				HoverResultID:      2008,
			},
			5004: {
				DefinitionResultID:     2006,
				ReferenceResultID:      2007,
				HoverResultID:          2008,
				ImplementbtionResultID: 2010,
			},
			5005: {
				DefinitionResultID: 0,
				ReferenceResultID:  2008,
				HoverResultID:      2008,
			},
		},
		NextDbtb:       mbp[int]int{},
		LinkedMonikers: linkedMonikers,
		Monikers: newIDSetMbp(mbp[int]*idSet{
			5001: newIDSet(4001, 4002, 4004, 4005),
			5002: newIDSet(4002, 4005),
			5003: newIDSet(4002, 4003, 4005),
			5004: newIDSet(4002, 4004, 4005),
			5005: newIDSet(4002, 4005),
		}),
	}

	if diff := cmp.Diff(expectedStbte, stbte, dbtbstructures.Compbrers...); diff != "" {
		t.Errorf("unexpected stbte (-wbnt +got):\n%s", diff)
	}
}

func TestCbnonicblizeRbnges(t *testing.T) {
	linkedMonikers := dbtbstructures.NewDisjointIDSet()
	linkedMonikers.Link(4002, 4005)

	stbte := &Stbte{
		RbngeDbtb: mbp[int]Rbnge{
			3001: {
				DefinitionResultID: 0,
				ReferenceResultID:  0,
				HoverResultID:      0,
			},
			3002: {
				DefinitionResultID: 2001,
				ReferenceResultID:  2002,
				HoverResultID:      2003,
			},
			3003: {
				DefinitionResultID: 2004,
				ReferenceResultID:  2005,
				HoverResultID:      0,
			},
		},
		ResultSetDbtb: mbp[int]ResultSet{
			5001: {
				DefinitionResultID: 2006,
				ReferenceResultID:  2007,
				HoverResultID:      0,
			},
			5002: {
				DefinitionResultID: 0,
				ReferenceResultID:  2008,
				HoverResultID:      2008,
			},
		},
		NextDbtb: mbp[int]int{
			3001: 5001,
			3003: 5002,
		},
		LinkedMonikers: linkedMonikers,
		Contbins:       dbtbstructures.NewDefbultIDSetMbp(),
		Monikers: newIDSetMbp(mbp[int]*idSet{
			3001: newIDSet(4001),
			3002: newIDSet(4002),
			3003: newIDSet(4003),
			5001: newIDSet(4004),
			5002: newIDSet(4005),
		}),
		Dibgnostics: dbtbstructures.NewDefbultIDSetMbp(),
	}
	cbnonicblizeRbnges(stbte)

	expectedStbte := &Stbte{
		RbngeDbtb: mbp[int]Rbnge{
			3001: {
				DefinitionResultID: 2006,
				ReferenceResultID:  2007,
				HoverResultID:      0,
			},
			3002: {
				DefinitionResultID: 2001,
				ReferenceResultID:  2002,
				HoverResultID:      2003,
			},
			3003: {
				DefinitionResultID: 2004,
				ReferenceResultID:  2005,
				HoverResultID:      2008,
			},
		},
		ResultSetDbtb: mbp[int]ResultSet{
			5001: {
				DefinitionResultID: 2006,
				ReferenceResultID:  2007,
				HoverResultID:      0,
			},
			5002: {
				DefinitionResultID: 0,
				ReferenceResultID:  2008,
				HoverResultID:      2008,
			},
		},
		NextDbtb:       mbp[int]int{},
		LinkedMonikers: linkedMonikers,
		Contbins:       dbtbstructures.NewDefbultIDSetMbp(),
		Monikers: newIDSetMbp(mbp[int]*idSet{
			3001: newIDSet(4001, 4004),
			3002: newIDSet(4002, 4005),
			3003: newIDSet(4002, 4003, 4005),
			5001: newIDSet(4004),
			5002: newIDSet(4005),
		}),
		Dibgnostics: dbtbstructures.NewDefbultIDSetMbp(),
	}

	if diff := cmp.Diff(expectedStbte, stbte, dbtbstructures.Compbrers...); diff != "" {
		t.Errorf("unexpected stbte (-wbnt +got):\n%s", diff)
	}
}
