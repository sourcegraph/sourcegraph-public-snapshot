package conversion

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/conversion/datastructures"
)

var newIDSet = datastructures.IDSetWith
var newIDSetMap = datastructures.DefaultIDSetMapWith

type idSet = datastructures.IDSet

func TestCanonicalizeDocuments(t *testing.T) {
	state := &State{
		DocumentData: map[int]string{
			1001: "main.go",
			1002: "foo.go",
			1003: "bar.go",
			1004: "main.go",
		},
		DefinitionData: map[int]*datastructures.DefaultIDSetMap{
			2001: newIDSetMap(map[int]*idSet{1001: newIDSet(3005)}),
			2002: newIDSetMap(map[int]*idSet{1002: newIDSet(3006), 1004: newIDSet(3007)}),
		},
		ReferenceData: map[int]*datastructures.DefaultIDSetMap{
			2003: newIDSetMap(map[int]*idSet{1001: newIDSet(3008)}),
			2004: newIDSetMap(map[int]*idSet{1003: newIDSet(3009), 1004: newIDSet(3010)}),
		},
		Contains: newIDSetMap(map[int]*idSet{
			1001: newIDSet(3001),
			1002: newIDSet(3002),
			1003: newIDSet(3003),
			1004: newIDSet(3004),
		}),
		Monikers:    datastructures.NewDefaultIDSetMap(),
		Diagnostics: datastructures.NewDefaultIDSetMap(),
	}
	canonicalizeDocuments(state)

	expectedState := &State{
		DocumentData: map[int]string{
			1001: "main.go",
			1002: "foo.go",
			1003: "bar.go",
		},
		DefinitionData: map[int]*datastructures.DefaultIDSetMap{
			2001: newIDSetMap(map[int]*idSet{1001: newIDSet(3005)}),
			2002: newIDSetMap(map[int]*idSet{1002: newIDSet(3006), 1001: newIDSet(3007)}),
		},
		ReferenceData: map[int]*datastructures.DefaultIDSetMap{
			2003: newIDSetMap(map[int]*idSet{1001: newIDSet(3008)}),
			2004: newIDSetMap(map[int]*idSet{1003: newIDSet(3009), 1001: newIDSet(3010)}),
		},
		Contains: newIDSetMap(map[int]*idSet{
			1001: newIDSet(3001, 3004),
			1002: newIDSet(3002),
			1003: newIDSet(3003),
		}),
		Monikers:    datastructures.NewDefaultIDSetMap(),
		Diagnostics: datastructures.NewDefaultIDSetMap(),
	}

	if diff := cmp.Diff(expectedState, state, datastructures.Comparers...); diff != "" {
		t.Errorf("unexpected state (-want +got):\n%s", diff)
	}
}

func TestCanonicalizeReferenceResults(t *testing.T) {
	linkedReferenceResults := datastructures.NewDisjointIDSet()
	linkedReferenceResults.Link(2001, 2003)

	state := &State{
		RangeData: map[int]Range{
			3001: {ReferenceResultID: 2002},
			3002: {ReferenceResultID: 2003},
		},
		ResultSetData: map[int]ResultSet{
			5003: {ReferenceResultID: 2003},
			5004: {ReferenceResultID: 2004},
		},
		ReferenceData: map[int]*datastructures.DefaultIDSetMap{
			2001: newIDSetMap(map[int]*idSet{
				1001: newIDSet(3005),
			}),
			2002: newIDSetMap(map[int]*idSet{
				1002: newIDSet(3006),
				1004: newIDSet(3007),
			}),
			2003: newIDSetMap(map[int]*idSet{
				1001: newIDSet(3008),
				1003: newIDSet(3009),
			}),
			2004: newIDSetMap(map[int]*idSet{
				1004: newIDSet(3010),
			}),
		},
		LinkedReferenceResults: map[int][]int{2001: {2003, 2004}, 2002: {2001}},
	}
	canonicalizeReferenceResults(state)

	expectedState := &State{
		RangeData: map[int]Range{
			3001: {ReferenceResultID: 2002},
			3002: {ReferenceResultID: 2003},
		},
		ResultSetData: map[int]ResultSet{
			5003: {ReferenceResultID: 2003},
			5004: {ReferenceResultID: 2004},
		},
		ReferenceData: map[int]*datastructures.DefaultIDSetMap{
			2001: newIDSetMap(map[int]*idSet{
				1001: newIDSet(3005, 3008),
				1003: newIDSet(3009),
				1004: newIDSet(3010),
			}),
			2002: newIDSetMap(map[int]*idSet{
				1001: newIDSet(3005, 3008),
				1002: newIDSet(3006),
				1003: newIDSet(3009),
				1004: newIDSet(3007, 3010),
			}),
			2003: newIDSetMap(map[int]*idSet{
				1001: newIDSet(3008),
				1003: newIDSet(3009),
			}),
			2004: newIDSetMap(map[int]*idSet{
				1004: newIDSet(3010),
			}),
		},
		LinkedReferenceResults: map[int][]int{2001: {2003, 2004}, 2002: {2001}},
	}

	if diff := cmp.Diff(expectedState, state, datastructures.Comparers...); diff != "" {
		t.Errorf("unexpected state (-want +got):\n%s", diff)
	}
}

func TestCanonicalizeDocumentsInDefinitionReferences(t *testing.T) {
	actualMap := map[int]*datastructures.DefaultIDSetMap{
		1: newIDSetMap(map[int]*idSet{
			11: newIDSet(101, 102),
			12: newIDSet(101, 103),
		}),
		2: newIDSetMap(map[int]*idSet{
			12: newIDSet(104),
		}),
	}
	canonicalizeDocumentsInDefinitionReferences(actualMap, map[int]int{
		12: 11,
	})

	expectedMap := map[int]*datastructures.DefaultIDSetMap{
		1: newIDSetMap(map[int]*idSet{
			11: newIDSet(101, 102, 103),
		}),
		2: newIDSetMap(map[int]*idSet{
			11: newIDSet(104),
		}),
	}

	if diff := cmp.Diff(expectedMap, actualMap, datastructures.Comparers...); diff != "" {
		t.Errorf("unexpected state (-want +got):\n%s", diff)
	}
}

func TestCanonicalizeResultSets(t *testing.T) {
	linkedMonikers := datastructures.NewDisjointIDSet()
	linkedMonikers.Link(4002, 4005)

	state := &State{
		ResultSetData: map[int]ResultSet{
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
				ImplementationResultID: 2010,
			},
			5005: {
				DefinitionResultID: 0,
				ReferenceResultID:  2008,
				HoverResultID:      2008,
			},
		},
		NextData: map[int]int{
			5001: 5004,
			5003: 5005,
			5004: 5005,
		},
		LinkedMonikers: linkedMonikers,
		Monikers: newIDSetMap(map[int]*idSet{
			5001: newIDSet(4001),
			5002: newIDSet(4002),
			5003: newIDSet(4003),
			5004: newIDSet(4004),
			5005: newIDSet(4005),
		}),
	}
	canonicalizeResultSets(state)

	expectedState := &State{
		ResultSetData: map[int]ResultSet{
			5001: {
				DefinitionResultID:     2006,
				ReferenceResultID:      2007,
				HoverResultID:          2008,
				ImplementationResultID: 2010,
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
				ImplementationResultID: 2010,
			},
			5005: {
				DefinitionResultID: 0,
				ReferenceResultID:  2008,
				HoverResultID:      2008,
			},
		},
		NextData:       map[int]int{},
		LinkedMonikers: linkedMonikers,
		Monikers: newIDSetMap(map[int]*idSet{
			5001: newIDSet(4001, 4002, 4004, 4005),
			5002: newIDSet(4002, 4005),
			5003: newIDSet(4002, 4003, 4005),
			5004: newIDSet(4002, 4004, 4005),
			5005: newIDSet(4002, 4005),
		}),
	}

	if diff := cmp.Diff(expectedState, state, datastructures.Comparers...); diff != "" {
		t.Errorf("unexpected state (-want +got):\n%s", diff)
	}
}

func TestCanonicalizeRanges(t *testing.T) {
	linkedMonikers := datastructures.NewDisjointIDSet()
	linkedMonikers.Link(4002, 4005)

	state := &State{
		RangeData: map[int]Range{
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
		ResultSetData: map[int]ResultSet{
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
		NextData: map[int]int{
			3001: 5001,
			3003: 5002,
		},
		LinkedMonikers: linkedMonikers,
		Contains:       datastructures.NewDefaultIDSetMap(),
		Monikers: newIDSetMap(map[int]*idSet{
			3001: newIDSet(4001),
			3002: newIDSet(4002),
			3003: newIDSet(4003),
			5001: newIDSet(4004),
			5002: newIDSet(4005),
		}),
		Diagnostics: datastructures.NewDefaultIDSetMap(),
	}
	canonicalizeRanges(state)

	expectedState := &State{
		RangeData: map[int]Range{
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
		ResultSetData: map[int]ResultSet{
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
		NextData:       map[int]int{},
		LinkedMonikers: linkedMonikers,
		Contains:       datastructures.NewDefaultIDSetMap(),
		Monikers: newIDSetMap(map[int]*idSet{
			3001: newIDSet(4001, 4004),
			3002: newIDSet(4002, 4005),
			3003: newIDSet(4002, 4003, 4005),
			5001: newIDSet(4004),
			5002: newIDSet(4005),
		}),
		Diagnostics: datastructures.NewDefaultIDSetMap(),
	}

	if diff := cmp.Diff(expectedState, state, datastructures.Comparers...); diff != "" {
		t.Errorf("unexpected state (-want +got):\n%s", diff)
	}
}
