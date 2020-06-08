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
	input, err := ioutil.ReadFile("../../testdata/dump1.lsif")
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
		DocumentData: map[string]lsif.Document{
			"02": {
				URI:         "foo.go",
				Contains:    datastructures.IDSet{"04": {}, "05": {}, "06": {}},
				Diagnostics: datastructures.IDSet{"49": {}},
			},
			"03": {
				URI:         "bar.go",
				Contains:    datastructures.IDSet{"07": {}, "08": {}, "09": {}},
				Diagnostics: datastructures.IDSet{},
			},
		},
		RangeData: map[string]lsif.Range{
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
		ResultSetData: map[string]lsif.ResultSet{
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
		MonikerData: map[string]lsif.Moniker{
			"18": {Kind: "import", Scheme: "scheme A", Identifier: "ident A", PackageInformationID: "22"},
			"19": {Kind: "export", Scheme: "scheme B", Identifier: "ident B", PackageInformationID: "23"},
			"20": {Kind: "import", Scheme: "scheme C", Identifier: "ident C", PackageInformationID: ""},
			"21": {Kind: "export", Scheme: "scheme D", Identifier: "ident D", PackageInformationID: ""},
		},
		PackageInformationData: map[string]lsif.PackageInformation{
			"22": {Name: "pkg A", Version: "v0.1.0"},
			"23": {Name: "pkg B", Version: "v1.2.3"},
		},
		Diagnostics: map[string]lsif.DiagnosticResult{
			"49": {
				Result: []lsif.Diagnostic{
					{
						Severity:       1,
						Code:           "2322",
						Message:        "Type '10' is not assignable to type 'string'.",
						Source:         "eslint",
						StartLine:      1,
						StartCharacter: 5,
						EndLine:        1,
						EndCharacter:   6,
					},
				},
			},
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

func TestCorrelateMetaDataRoot(t *testing.T) {
	input, err := ioutil.ReadFile("../../testdata/dump2.lsif")
	if err != nil {
		t.Fatalf("unexpected error reading test file: %s", err)
	}

	state, err := correlateFromReader(bytes.NewReader(input), "root/")
	if err != nil {
		t.Fatalf("unexpected error correlating input: %s", err)
	}

	expectedState := &State{
		LSIFVersion: "0.4.3",
		ProjectRoot: "file:///test/root/",
		DocumentData: map[string]lsif.Document{
			"02": {
				URI:         "foo.go",
				Contains:    datastructures.IDSet{},
				Diagnostics: datastructures.IDSet{},
			},
		},
		RangeData:              map[string]lsif.Range{},
		ResultSetData:          map[string]lsif.ResultSet{},
		DefinitionData:         map[string]datastructures.DefaultIDSetMap{},
		ReferenceData:          map[string]datastructures.DefaultIDSetMap{},
		HoverData:              map[string]string{},
		MonikerData:            map[string]lsif.Moniker{},
		PackageInformationData: map[string]lsif.PackageInformation{},
		Diagnostics:            map[string]lsif.DiagnosticResult{},
		NextData:               map[string]string{},
		ImportedMonikers:       datastructures.IDSet{},
		ExportedMonikers:       datastructures.IDSet{},
		LinkedMonikers:         datastructures.DisjointIDSet{},
		LinkedReferenceResults: datastructures.DisjointIDSet{},
	}

	if diff := cmp.Diff(expectedState, state); diff != "" {
		t.Errorf("unexpected state (-want +got):\n%s", diff)
	}
}

func TestCorrelateMetaDataRootX(t *testing.T) {
	input, err := ioutil.ReadFile("../../testdata/dump3.lsif")
	if err != nil {
		t.Fatalf("unexpected error reading test file: %s", err)
	}

	state, err := correlateFromReader(bytes.NewReader(input), "")
	if err != nil {
		t.Fatalf("unexpected error correlating input: %s", err)
	}

	expectedState := &State{
		LSIFVersion: "0.4.3",
		ProjectRoot: "file:///__w/sourcegraph/sourcegraph/shared/",
		DocumentData: map[string]lsif.Document{
			"02": {
				URI:         "../node_modules/@types/history/index.d.ts",
				Contains:    datastructures.IDSet{},
				Diagnostics: datastructures.IDSet{},
			},
		},
		RangeData:              map[string]lsif.Range{},
		ResultSetData:          map[string]lsif.ResultSet{},
		DefinitionData:         map[string]datastructures.DefaultIDSetMap{},
		ReferenceData:          map[string]datastructures.DefaultIDSetMap{},
		HoverData:              map[string]string{},
		MonikerData:            map[string]lsif.Moniker{},
		PackageInformationData: map[string]lsif.PackageInformation{},
		Diagnostics:            map[string]lsif.DiagnosticResult{},
		NextData:               map[string]string{},
		ImportedMonikers:       datastructures.IDSet{},
		ExportedMonikers:       datastructures.IDSet{},
		LinkedMonikers:         datastructures.DisjointIDSet{},
		LinkedReferenceResults: datastructures.DisjointIDSet{},
	}

	if diff := cmp.Diff(expectedState, state); diff != "" {
		t.Errorf("unexpected state (-want +got):\n%s", diff)
	}
}
