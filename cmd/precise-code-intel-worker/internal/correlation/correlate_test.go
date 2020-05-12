package correlation

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-worker/internal/correlation/datastructures"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-worker/internal/correlation/lsif"
)

func TestCorrelate(t *testing.T) {
	input, err := ioutil.ReadFile("../../testdata/dump.lsif")
	if err != nil {
		t.Fatalf("unexpected error reading test file: %s", err)
	}

	state, err := correlateFromReader(bytes.NewReader(input), "root")
	if err != nil {
		t.Fatalf("unexpected error correlating input: %s", err)
	}

	expectedState := &State{
		LSIFVersion: "0.4.3",
		ProjectRoot: "file:///test/root",
		DocumentData: map[string]lsif.DocumentData{
			"02": {URI: "foo.go", Contains: datastructures.IDSet{"04": {}, "05": {}, "06": {}}},
			"03": {URI: "bar.go", Contains: datastructures.IDSet{"07": {}, "08": {}, "09": {}}},
		},
		RangeData: map[string]lsif.RangeData{
			"04": {
				StartLine:          1,
				StartCharacter:     2,
				EndLine:            3,
				EndCharacter:       4,
				DefinitionResultID: "13",
				MonikerIDs:         datastructures.IDSet{},
			},
			"05": {
				StartLine:         2,
				StartCharacter:    3,
				EndLine:           4,
				EndCharacter:      5,
				ReferenceResultID: "15",
				MonikerIDs:        datastructures.IDSet{},
			},
			"06": {
				StartLine:          3,
				StartCharacter:     4,
				EndLine:            5,
				EndCharacter:       6,
				DefinitionResultID: "13",
				HoverResultID:      "17",
				MonikerIDs:         datastructures.IDSet{},
			},
			"07": {
				StartLine:         4,
				StartCharacter:    5,
				EndLine:           6,
				EndCharacter:      7,
				ReferenceResultID: "15",
				MonikerIDs:        datastructures.IDSet{"18": {}},
			},
			"08": {
				StartLine:      5,
				StartCharacter: 6,
				EndLine:        7,
				EndCharacter:   8,
				HoverResultID:  "17",
				MonikerIDs:     datastructures.IDSet{},
			},
			"09": {
				StartLine:      6,
				StartCharacter: 7,
				EndLine:        8,
				EndCharacter:   9,
				MonikerIDs:     datastructures.IDSet{"19": {}},
			},
		},
		ResultSetData: map[string]lsif.ResultSetData{
			"10": {
				DefinitionResultID: "12",
				ReferenceResultID:  "14",
				MonikerIDs:         datastructures.IDSet{"20": {}},
			},
			"11": {
				HoverResultID: "16",
				MonikerIDs:    datastructures.IDSet{"21": {}},
			},
		},
		DefinitionData: map[string]datastructures.DefaultIDSetMap{
			"12": {"03": {"07": {}}},
			"13": {"03": {"08": {}}},
		},
		ReferenceData: map[string]datastructures.DefaultIDSetMap{
			"14": {"02": {"04": {}, "05": {}}},
			"15": {},
		},
		HoverData: map[string]string{
			"16": "```go\ntext A\n```",
			"17": "```go\ntext B\n```",
		},
		MonikerData: map[string]lsif.MonikerData{
			"18": {Kind: "import", Scheme: "scheme A", Identifier: "ident A", PackageInformationID: "22"},
			"19": {Kind: "export", Scheme: "scheme B", Identifier: "ident B", PackageInformationID: "23"},
			"20": {Kind: "import", Scheme: "scheme C", Identifier: "ident C", PackageInformationID: ""},
			"21": {Kind: "export", Scheme: "scheme D", Identifier: "ident D", PackageInformationID: ""},
		},
		PackageInformationData: map[string]lsif.PackageInformationData{
			"22": {Name: "pkg A", Version: "v0.1.0"},
			"23": {Name: "pkg B", Version: "v1.2.3"},
		},
		NextData: map[string]string{
			"09": "10",
			"10": "11",
		},
		ImportedMonikers:       datastructures.IDSet{"18": {}},
		ExportedMonikers:       datastructures.IDSet{"19": {}},
		LinkedMonikers:         datastructures.DisjointIDSet{"19": {"21": {}}, "21": {"19": {}}},
		LinkedReferenceResults: datastructures.DisjointIDSet{"14": {"15": {}}, "15": {"14": {}}},
	}

	if diff := cmp.Diff(expectedState, state); diff != "" {
		t.Errorf("unexpected state (-want +got):\n%s", diff)
	}
}
