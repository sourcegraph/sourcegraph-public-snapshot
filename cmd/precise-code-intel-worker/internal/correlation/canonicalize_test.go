package correlation

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-worker/internal/correlation/datastructures"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-worker/internal/correlation/lsif"
)

func TestCanonicalizeDocuments(t *testing.T) {
	state := &State{
		DocumentData: map[string]lsif.DocumentData{
			"d01": {URI: "main.go", Contains: datastructures.IDSet{"r01": {}}},
			"d02": {URI: "foo.go", Contains: datastructures.IDSet{"r02": {}}},
			"d03": {URI: "bar.go", Contains: datastructures.IDSet{"r03": {}}},
			"d04": {URI: "main.go", Contains: datastructures.IDSet{"r04": {}}},
		},
		DefinitionData: map[string]datastructures.DefaultIDSetMap{
			"x01": {"d01": datastructures.IDSet{"r05": {}}},
			"x02": {"d02": datastructures.IDSet{"r06": {}}, "d04": datastructures.IDSet{"r07": {}}},
		},
		ReferenceData: map[string]datastructures.DefaultIDSetMap{
			"x03": {"d01": datastructures.IDSet{"r08": {}}},
			"x04": {"d03": datastructures.IDSet{"r09": {}}, "d04": datastructures.IDSet{"r10": {}}},
		},
	}
	canonicalizeDocuments(state)

	expectedState := &State{
		DocumentData: map[string]lsif.DocumentData{
			"d01": {URI: "main.go", Contains: datastructures.IDSet{"r01": {}, "r04": {}}},
			"d02": {URI: "foo.go", Contains: datastructures.IDSet{"r02": {}}},
			"d03": {URI: "bar.go", Contains: datastructures.IDSet{"r03": {}}},
		},
		DefinitionData: map[string]datastructures.DefaultIDSetMap{
			"x01": {"d01": datastructures.IDSet{"r05": {}}},
			"x02": {"d02": datastructures.IDSet{"r06": {}}, "d01": datastructures.IDSet{"r07": {}}},
		},
		ReferenceData: map[string]datastructures.DefaultIDSetMap{
			"x03": {"d01": datastructures.IDSet{"r08": {}}},
			"x04": {"d03": datastructures.IDSet{"r09": {}}, "d01": datastructures.IDSet{"r10": {}}},
		},
	}

	if diff := cmp.Diff(expectedState, state); diff != "" {
		t.Errorf("unexpected state (-want +got):\n%s", diff)
	}
}

func TestCanonicalizeReferenceResults(t *testing.T) {
	linkedReferenceResults := datastructures.DisjointIDSet{}
	linkedReferenceResults.Union("x01", "x03")

	state := &State{
		RangeData: map[string]lsif.RangeData{
			"r01": {ReferenceResultID: "x02"},
			"r02": {ReferenceResultID: "x03"},
		},
		ResultSetData: map[string]lsif.ResultSetData{
			"s03": {ReferenceResultID: "x03"},
			"s04": {ReferenceResultID: "x04"},
		},
		ReferenceData: map[string]datastructures.DefaultIDSetMap{
			"x01": {"d01": {"r05": {}}},
			"x02": {"d02": {"r06": {}}, "d04": {"r07": {}}},
			"x03": {"d01": {"r08": {}}, "d03": {"r09": {}}},
			"x04": {"d04": {"r10": {}}},
		},
		LinkedReferenceResults: linkedReferenceResults,
	}
	canonicalizeReferenceResults(state)

	expectedState := &State{
		RangeData: map[string]lsif.RangeData{
			"r01": {ReferenceResultID: "x02"},
			"r02": {ReferenceResultID: "x01"},
		},
		ResultSetData: map[string]lsif.ResultSetData{
			"s03": {ReferenceResultID: "x01"},
			"s04": {ReferenceResultID: "x04"},
		},
		ReferenceData: map[string]datastructures.DefaultIDSetMap{
			"x01": {"d01": {"r05": {}, "r08": {}}, "d03": {"r09": {}}},
			"x02": {"d02": {"r06": {}}, "d04": {"r07": {}}},
			"x04": {"d04": {"r10": {}}},
		},

		LinkedReferenceResults: linkedReferenceResults,
	}

	if diff := cmp.Diff(expectedState, state); diff != "" {
		t.Errorf("unexpected state (-want +got):\n%s", diff)
	}
}

func TestCanonicalizeResultSets(t *testing.T) {
	linkedMonikers := datastructures.DisjointIDSet{}
	linkedMonikers.Union("m02", "m05")

	state := &State{
		ResultSetData: map[string]lsif.ResultSetData{
			"s01": {
				DefinitionResultID: "",
				ReferenceResultID:  "",
				HoverResultID:      "",
				MonikerIDs:         datastructures.IDSet{"m01": {}},
			},
			"s02": {
				DefinitionResultID: "x01",
				ReferenceResultID:  "x02",
				HoverResultID:      "x03",
				MonikerIDs:         datastructures.IDSet{"m02": {}},
			},
			"s03": {
				DefinitionResultID: "x04",
				ReferenceResultID:  "x05",
				HoverResultID:      "",
				MonikerIDs:         datastructures.IDSet{"m03": {}},
			},
			"s04": {
				DefinitionResultID: "x06",
				ReferenceResultID:  "x07",
				HoverResultID:      "",
				MonikerIDs:         datastructures.IDSet{"m04": {}},
			},
			"s05": {
				DefinitionResultID: "",
				ReferenceResultID:  "x08",
				HoverResultID:      "x08",
				MonikerIDs:         datastructures.IDSet{"m05": {}},
			},
		},
		NextData: map[string]string{
			"s01": "s04",
			"s03": "s05",
			"s04": "s05",
		},
		LinkedMonikers: linkedMonikers,
	}
	canonicalizeResultSets(state)

	expectedState := &State{
		ResultSetData: map[string]lsif.ResultSetData{
			"s01": {
				DefinitionResultID: "x06",
				ReferenceResultID:  "x07",
				HoverResultID:      "x08",
				MonikerIDs:         datastructures.IDSet{"m01": {}, "m02": {}, "m04": {}, "m05": {}},
			},
			"s02": {
				DefinitionResultID: "x01",
				ReferenceResultID:  "x02",
				HoverResultID:      "x03",
				MonikerIDs:         datastructures.IDSet{"m02": {}, "m05": {}},
			},
			"s03": {
				DefinitionResultID: "x04",
				ReferenceResultID:  "x05",
				HoverResultID:      "x08",
				MonikerIDs:         datastructures.IDSet{"m02": {}, "m03": {}, "m05": {}},
			},
			"s04": {
				DefinitionResultID: "x06",
				ReferenceResultID:  "x07",
				HoverResultID:      "x08",
				MonikerIDs:         datastructures.IDSet{"m02": {}, "m04": {}, "m05": {}},
			},
			"s05": {
				DefinitionResultID: "",
				ReferenceResultID:  "x08",
				HoverResultID:      "x08",
				MonikerIDs:         datastructures.IDSet{"m02": {}, "m05": {}},
			},
		},
		NextData:       map[string]string{},
		LinkedMonikers: linkedMonikers,
	}

	if diff := cmp.Diff(expectedState, state); diff != "" {
		t.Errorf("unexpected state (-want +got):\n%s", diff)
	}
}

func TestCanonicalizeRanges(t *testing.T) {
	linkedMonikers := datastructures.DisjointIDSet{}
	linkedMonikers.Union("m02", "m05")

	state := &State{
		RangeData: map[string]lsif.RangeData{
			"r01": {
				DefinitionResultID: "",
				ReferenceResultID:  "",
				HoverResultID:      "",
				MonikerIDs:         datastructures.IDSet{"m01": {}},
			},
			"r02": {
				DefinitionResultID: "x01",
				ReferenceResultID:  "x02",
				HoverResultID:      "x03",
				MonikerIDs:         datastructures.IDSet{"m02": {}},
			},
			"r03": {
				DefinitionResultID: "x04",
				ReferenceResultID:  "x05",
				HoverResultID:      "",
				MonikerIDs:         datastructures.IDSet{"m03": {}},
			},
		},
		ResultSetData: map[string]lsif.ResultSetData{
			"s01": {
				DefinitionResultID: "x06",
				ReferenceResultID:  "x07",
				HoverResultID:      "",
				MonikerIDs:         datastructures.IDSet{"m04": {}},
			},
			"s02": {
				DefinitionResultID: "",
				ReferenceResultID:  "x08",
				HoverResultID:      "x08",
				MonikerIDs:         datastructures.IDSet{"m05": {}},
			},
		},
		NextData: map[string]string{
			"r01": "s01",
			"r03": "s02",
		},
		LinkedMonikers: linkedMonikers,
	}
	canonicalizeRanges(state)

	expectedState := &State{
		RangeData: map[string]lsif.RangeData{
			"r01": {
				DefinitionResultID: "x06",
				ReferenceResultID:  "x07",
				HoverResultID:      "",
				MonikerIDs:         datastructures.IDSet{"m01": {}, "m04": {}},
			},
			"r02": {
				DefinitionResultID: "x01",
				ReferenceResultID:  "x02",
				HoverResultID:      "x03",
				MonikerIDs:         datastructures.IDSet{"m02": {}, "m05": {}},
			},
			"r03": {
				DefinitionResultID: "x04",
				ReferenceResultID:  "x05",
				HoverResultID:      "x08",
				MonikerIDs:         datastructures.IDSet{"m02": {}, "m03": {}, "m05": {}},
			},
		},
		ResultSetData: map[string]lsif.ResultSetData{
			"s01": {
				DefinitionResultID: "x06",
				ReferenceResultID:  "x07",
				HoverResultID:      "",
				MonikerIDs:         datastructures.IDSet{"m04": {}},
			},
			"s02": {
				DefinitionResultID: "",
				ReferenceResultID:  "x08",
				HoverResultID:      "x08",
				MonikerIDs:         datastructures.IDSet{"m05": {}},
			},
		},
		NextData:       map[string]string{},
		LinkedMonikers: linkedMonikers,
	}

	if diff := cmp.Diff(expectedState, state); diff != "" {
		t.Errorf("unexpected state (-want +got):\n%s", diff)
	}
}
