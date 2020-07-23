package correlation

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-worker/internal/correlation/datastructures"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-worker/internal/correlation/lsif"
)

func TestCanonicalizeDocuments(t *testing.T) {
	state := &State{
		DocumentData: map[int]lsif.Document{
			1001: {
				URI:         "main.go",
				Contains:    datastructures.IDSetWith(3001),
				Diagnostics: datastructures.NewIDSet(),
			},
			1002: {
				URI:         "foo.go",
				Contains:    datastructures.IDSetWith(3002),
				Diagnostics: datastructures.NewIDSet(),
			},
			1003: {
				URI:         "bar.go",
				Contains:    datastructures.IDSetWith(3003),
				Diagnostics: datastructures.NewIDSet(),
			},
			1004: {
				URI:         "main.go",
				Contains:    datastructures.IDSetWith(3004),
				Diagnostics: datastructures.NewIDSet(),
			},
		},
		DefinitionData: map[int]datastructures.DefaultIDSetMap{
			2001: {
				1001: datastructures.IDSetWith(3005),
			},
			2002: {
				1002: datastructures.IDSetWith(3006),
				1004: datastructures.IDSetWith(3007),
			},
		},
		ReferenceData: map[int]datastructures.DefaultIDSetMap{
			2003: {
				1001: datastructures.IDSetWith(3008),
			},
			2004: {
				1003: datastructures.IDSetWith(3009),
				1004: datastructures.IDSetWith(3010),
			},
		},
	}
	canonicalizeDocuments(state)

	expectedState := &State{
		DocumentData: map[int]lsif.Document{
			1001: {
				URI:         "main.go",
				Contains:    datastructures.IDSetWith(3001, 3004),
				Diagnostics: datastructures.NewIDSet(),
			},
			1002: {
				URI:         "foo.go",
				Contains:    datastructures.IDSetWith(3002),
				Diagnostics: datastructures.NewIDSet(),
			},
			1003: {
				URI:         "bar.go",
				Contains:    datastructures.IDSetWith(3003),
				Diagnostics: datastructures.NewIDSet(),
			},
		},
		DefinitionData: map[int]datastructures.DefaultIDSetMap{
			2001: {
				1001: datastructures.IDSetWith(3005),
			},
			2002: {
				1002: datastructures.IDSetWith(3006),
				1001: datastructures.IDSetWith(3007),
			},
		},
		ReferenceData: map[int]datastructures.DefaultIDSetMap{
			2003: {
				1001: datastructures.IDSetWith(3008),
			},
			2004: {
				1003: datastructures.IDSetWith(3009),
				1001: datastructures.IDSetWith(3010),
			},
		},
	}

	if diff := cmp.Diff(expectedState, state, datastructures.IDSetComparer); diff != "" {
		t.Errorf("unexpected state (-want +got):\n%s", diff)
	}
}

func TestCanonicalizeReferenceResults(t *testing.T) {
	linkedReferenceResults := datastructures.DisjointIDSet{}
	linkedReferenceResults.Union(2001, 2003)

	state := &State{
		RangeData: map[int]lsif.Range{
			3001: {ReferenceResultID: 2002},
			3002: {ReferenceResultID: 2003},
		},
		ResultSetData: map[int]lsif.ResultSet{
			5003: {ReferenceResultID: 2003},
			5004: {ReferenceResultID: 2004},
		},
		ReferenceData: map[int]datastructures.DefaultIDSetMap{
			2001: {
				1001: datastructures.IDSetWith(3005),
			},
			2002: {
				1002: datastructures.IDSetWith(3006),
				1004: datastructures.IDSetWith(3007),
			},
			2003: {
				1001: datastructures.IDSetWith(3008),
				1003: datastructures.IDSetWith(3009),
			},
			2004: {
				1004: datastructures.IDSetWith(3010),
			},
		},
		LinkedReferenceResults: linkedReferenceResults,
	}
	canonicalizeReferenceResults(state)

	expectedState := &State{
		RangeData: map[int]lsif.Range{
			3001: {ReferenceResultID: 2002},
			3002: {ReferenceResultID: 2001},
		},
		ResultSetData: map[int]lsif.ResultSet{
			5003: {ReferenceResultID: 2001},
			5004: {ReferenceResultID: 2004},
		},
		ReferenceData: map[int]datastructures.DefaultIDSetMap{
			2001: {
				1001: datastructures.IDSetWith(3005, 3008),
				1003: datastructures.IDSetWith(3009),
			},
			2002: {
				1002: datastructures.IDSetWith(3006),
				1004: datastructures.IDSetWith(3007),
			},
			2004: {
				1004: datastructures.IDSetWith(3010),
			},
		},

		LinkedReferenceResults: linkedReferenceResults,
	}

	if diff := cmp.Diff(expectedState, state, datastructures.IDSetComparer); diff != "" {
		t.Errorf("unexpected state (-want +got):\n%s", diff)
	}
}

func TestCanonicalizeResultSets(t *testing.T) {
	linkedMonikers := datastructures.DisjointIDSet{}
	linkedMonikers.Union(4002, 4005)

	state := &State{
		ResultSetData: map[int]lsif.ResultSet{
			5001: {
				DefinitionResultID: 0,
				ReferenceResultID:  0,
				HoverResultID:      0,
				MonikerIDs:         datastructures.IDSetWith(4001),
			},
			5002: {
				DefinitionResultID: 2001,
				ReferenceResultID:  2002,
				HoverResultID:      2003,
				MonikerIDs:         datastructures.IDSetWith(4002),
			},
			5003: {
				DefinitionResultID: 2004,
				ReferenceResultID:  2005,
				HoverResultID:      0,
				MonikerIDs:         datastructures.IDSetWith(4003),
			},
			5004: {
				DefinitionResultID: 2006,
				ReferenceResultID:  2007,
				HoverResultID:      0,
				MonikerIDs:         datastructures.IDSetWith(4004),
			},
			5005: {
				DefinitionResultID: 0,
				ReferenceResultID:  2008,
				HoverResultID:      2008,
				MonikerIDs:         datastructures.IDSetWith(4005),
			},
		},
		NextData: map[int]int{
			5001: 5004,
			5003: 5005,
			5004: 5005,
		},
		LinkedMonikers: linkedMonikers,
	}
	canonicalizeResultSets(state)

	expectedState := &State{
		ResultSetData: map[int]lsif.ResultSet{
			5001: {
				DefinitionResultID: 2006,
				ReferenceResultID:  2007,
				HoverResultID:      2008,
				MonikerIDs:         datastructures.IDSetWith(4001, 4002, 4004, 4005),
			},
			5002: {
				DefinitionResultID: 2001,
				ReferenceResultID:  2002,
				HoverResultID:      2003,
				MonikerIDs:         datastructures.IDSetWith(4002, 4005),
			},
			5003: {
				DefinitionResultID: 2004,
				ReferenceResultID:  2005,
				HoverResultID:      2008,
				MonikerIDs:         datastructures.IDSetWith(4002, 4003, 4005),
			},
			5004: {
				DefinitionResultID: 2006,
				ReferenceResultID:  2007,
				HoverResultID:      2008,
				MonikerIDs:         datastructures.IDSetWith(4002, 4004, 4005),
			},
			5005: {
				DefinitionResultID: 0,
				ReferenceResultID:  2008,
				HoverResultID:      2008,
				MonikerIDs:         datastructures.IDSetWith(4002, 4005),
			},
		},
		NextData:       map[int]int{},
		LinkedMonikers: linkedMonikers,
	}

	if diff := cmp.Diff(expectedState, state, datastructures.IDSetComparer); diff != "" {
		t.Errorf("unexpected state (-want +got):\n%s", diff)
	}
}

func TestCanonicalizeRanges(t *testing.T) {
	linkedMonikers := datastructures.DisjointIDSet{}
	linkedMonikers.Union(4002, 4005)

	state := &State{
		RangeData: map[int]lsif.Range{
			3001: {
				DefinitionResultID: 0,
				ReferenceResultID:  0,
				HoverResultID:      0,
				MonikerIDs:         datastructures.IDSetWith(4001),
			},
			3002: {
				DefinitionResultID: 2001,
				ReferenceResultID:  2002,
				HoverResultID:      2003,
				MonikerIDs:         datastructures.IDSetWith(4002),
			},
			3003: {
				DefinitionResultID: 2004,
				ReferenceResultID:  2005,
				HoverResultID:      0,
				MonikerIDs:         datastructures.IDSetWith(4003),
			},
		},
		ResultSetData: map[int]lsif.ResultSet{
			5001: {
				DefinitionResultID: 2006,
				ReferenceResultID:  2007,
				HoverResultID:      0,
				MonikerIDs:         datastructures.IDSetWith(4004),
			},
			5002: {
				DefinitionResultID: 0,
				ReferenceResultID:  2008,
				HoverResultID:      2008,
				MonikerIDs:         datastructures.IDSetWith(4005),
			},
		},
		NextData: map[int]int{
			3001: 5001,
			3003: 5002,
		},
		LinkedMonikers: linkedMonikers,
	}
	canonicalizeRanges(state)

	expectedState := &State{
		RangeData: map[int]lsif.Range{
			3001: {
				DefinitionResultID: 2006,
				ReferenceResultID:  2007,
				HoverResultID:      0,
				MonikerIDs:         datastructures.IDSetWith(4001, 4004),
			},
			3002: {
				DefinitionResultID: 2001,
				ReferenceResultID:  2002,
				HoverResultID:      2003,
				MonikerIDs:         datastructures.IDSetWith(4002, 4005),
			},
			3003: {
				DefinitionResultID: 2004,
				ReferenceResultID:  2005,
				HoverResultID:      2008,
				MonikerIDs:         datastructures.IDSetWith(4002, 4003, 4005),
			},
		},
		ResultSetData: map[int]lsif.ResultSet{
			5001: {
				DefinitionResultID: 2006,
				ReferenceResultID:  2007,
				HoverResultID:      0,
				MonikerIDs:         datastructures.IDSetWith(4004),
			},
			5002: {
				DefinitionResultID: 0,
				ReferenceResultID:  2008,
				HoverResultID:      2008,
				MonikerIDs:         datastructures.IDSetWith(4005),
			},
		},
		NextData:       map[int]int{},
		LinkedMonikers: linkedMonikers,
	}

	if diff := cmp.Diff(expectedState, state, datastructures.IDSetComparer); diff != "" {
		t.Errorf("unexpected state (-want +got):\n%s", diff)
	}
}
