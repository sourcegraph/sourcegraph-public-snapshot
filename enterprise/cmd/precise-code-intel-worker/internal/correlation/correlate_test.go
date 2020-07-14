package correlation

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-worker/internal/correlation/datastructures"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-worker/internal/correlation/lsif"
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
		DocumentData: map[int]lsif.Document{
			2: {
				URI:         "foo.go",
				Contains:    datastructures.IDSetWith(4, 5, 6),
				Diagnostics: datastructures.IDSetWith(49),
			},
			3: {
				URI:         "bar.go",
				Contains:    datastructures.IDSetWith(7, 8, 9),
				Diagnostics: datastructures.NewIDSet(),
			},
		},
		RangeData: map[int]lsif.Range{
			4: {
				StartLine:          1,
				StartCharacter:     2,
				EndLine:            3,
				EndCharacter:       4,
				DefinitionResultID: 13,
				MonikerIDs:         datastructures.NewIDSet(),
			},
			5: {
				StartLine:         2,
				StartCharacter:    3,
				EndLine:           4,
				EndCharacter:      5,
				ReferenceResultID: 15,
				MonikerIDs:        datastructures.NewIDSet(),
			},
			6: {
				StartLine:          3,
				StartCharacter:     4,
				EndLine:            5,
				EndCharacter:       6,
				DefinitionResultID: 13,
				HoverResultID:      17,
				MonikerIDs:         datastructures.NewIDSet(),
			},
			7: {
				StartLine:         4,
				StartCharacter:    5,
				EndLine:           6,
				EndCharacter:      7,
				ReferenceResultID: 15,
				MonikerIDs:        datastructures.IDSetWith(18),
			},
			8: {
				StartLine:      5,
				StartCharacter: 6,
				EndLine:        7,
				EndCharacter:   8,
				HoverResultID:  17,
				MonikerIDs:     datastructures.NewIDSet(),
			},
			9: {
				StartLine:      6,
				StartCharacter: 7,
				EndLine:        8,
				EndCharacter:   9,
				MonikerIDs:     datastructures.IDSetWith(19),
			},
		},
		ResultSetData: map[int]lsif.ResultSet{
			10: {
				DefinitionResultID: 12,
				ReferenceResultID:  14,
				MonikerIDs:         datastructures.IDSetWith(20),
			},
			11: {
				HoverResultID: 16,
				MonikerIDs:    datastructures.IDSetWith(21),
			},
		},
		DefinitionData: map[int]datastructures.DefaultIDSetMap{
			12: {3: datastructures.IDSetWith(7)},
			13: {3: datastructures.IDSetWith(8)},
		},
		ReferenceData: map[int]datastructures.DefaultIDSetMap{
			14: {2: datastructures.IDSetWith(4, 5)},
			15: {},
		},
		HoverData: map[int]string{
			16: "```go\ntext A\n```",
			17: "```go\ntext B\n```",
		},
		MonikerData: map[int]lsif.Moniker{
			18: {
				Kind:                 "import",
				Scheme:               "scheme A",
				Identifier:           "ident A",
				PackageInformationID: 22,
			},
			19: {
				Kind:                 "export",
				Scheme:               "scheme B",
				Identifier:           "ident B",
				PackageInformationID: 23,
			},
			20: {
				Kind:                 "import",
				Scheme:               "scheme C",
				Identifier:           "ident C",
				PackageInformationID: 0,
			},
			21: {
				Kind:                 "export",
				Scheme:               "scheme D",
				Identifier:           "ident D",
				PackageInformationID: 0,
			},
		},
		PackageInformationData: map[int]lsif.PackageInformation{
			22: {Name: "pkg A", Version: "v0.1.0"},
			23: {Name: "pkg B", Version: "v1.2.3"},
		},
		Diagnostics: map[int]lsif.DiagnosticResult{
			49: {
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
		NextData: map[int]int{
			9:  10,
			10: 11,
		},
		ImportedMonikers: datastructures.IDSetWith(18),
		ExportedMonikers: datastructures.IDSetWith(19),
		LinkedMonikers: datastructures.DisjointIDSet{
			19: datastructures.IDSetWith(21),
			21: datastructures.IDSetWith(19),
		},
		LinkedReferenceResults: datastructures.DisjointIDSet{
			14: datastructures.IDSetWith(15),
			15: datastructures.IDSetWith(14),
		},
	}

	if diff := cmp.Diff(expectedState, state, datastructures.IDSetComparer); diff != "" {
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
		DocumentData: map[int]lsif.Document{
			2: {
				URI:         "foo.go",
				Contains:    datastructures.NewIDSet(),
				Diagnostics: datastructures.NewIDSet(),
			},
		},
		RangeData:              map[int]lsif.Range{},
		ResultSetData:          map[int]lsif.ResultSet{},
		DefinitionData:         map[int]datastructures.DefaultIDSetMap{},
		ReferenceData:          map[int]datastructures.DefaultIDSetMap{},
		HoverData:              map[int]string{},
		MonikerData:            map[int]lsif.Moniker{},
		PackageInformationData: map[int]lsif.PackageInformation{},
		Diagnostics:            map[int]lsif.DiagnosticResult{},
		NextData:               map[int]int{},
		ImportedMonikers:       datastructures.NewIDSet(),
		ExportedMonikers:       datastructures.NewIDSet(),
		LinkedMonikers:         datastructures.DisjointIDSet{},
		LinkedReferenceResults: datastructures.DisjointIDSet{},
	}

	if diff := cmp.Diff(expectedState, state, datastructures.IDSetComparer); diff != "" {
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
		DocumentData: map[int]lsif.Document{
			2: {
				URI:         "../node_modules/@types/history/index.d.ts",
				Contains:    datastructures.NewIDSet(),
				Diagnostics: datastructures.NewIDSet(),
			},
		},
		RangeData:              map[int]lsif.Range{},
		ResultSetData:          map[int]lsif.ResultSet{},
		DefinitionData:         map[int]datastructures.DefaultIDSetMap{},
		ReferenceData:          map[int]datastructures.DefaultIDSetMap{},
		HoverData:              map[int]string{},
		MonikerData:            map[int]lsif.Moniker{},
		PackageInformationData: map[int]lsif.PackageInformation{},
		Diagnostics:            map[int]lsif.DiagnosticResult{},
		NextData:               map[int]int{},
		ImportedMonikers:       datastructures.NewIDSet(),
		ExportedMonikers:       datastructures.NewIDSet(),
		LinkedMonikers:         datastructures.DisjointIDSet{},
		LinkedReferenceResults: datastructures.DisjointIDSet{},
	}

	if diff := cmp.Diff(expectedState, state, datastructures.IDSetComparer); diff != "" {
		t.Errorf("unexpected state (-want +got):\n%s", diff)
	}
}
