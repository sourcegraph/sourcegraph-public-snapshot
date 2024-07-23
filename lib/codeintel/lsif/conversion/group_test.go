package conversion

import (
	"context"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/conversion/datastructures"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/protocol"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/protocol/reader"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

func TestGroupBundleData(t *testing.T) {
	state := &State{
		DocumentData: map[int]string{
			1001: "foo.go",
			1002: "bar.go",
			1003: "baz.go",
		},
		RangeData: map[int]Range{
			2001: {
				Range: reader.Range{
					RangeData: protocol.RangeData{
						Start: protocol.Pos{Line: 1, Character: 2},
						End:   protocol.Pos{Line: 3, Character: 4},
					},
				},
				DefinitionResultID: 0,
				ReferenceResultID:  3006,
			},
			2002: {
				Range: reader.Range{
					RangeData: protocol.RangeData{
						Start: protocol.Pos{Line: 2, Character: 3},
						End:   protocol.Pos{Line: 4, Character: 5},
					},
				},
				DefinitionResultID: 3001,
				ReferenceResultID:  0,
			},
			2003: {
				Range: reader.Range{
					RangeData: protocol.RangeData{
						Start: protocol.Pos{Line: 3, Character: 4},
						End:   protocol.Pos{Line: 5, Character: 6},
					},
				},
				DefinitionResultID: 3002,
				ReferenceResultID:  0,
			},
			2004: {
				Range: reader.Range{
					RangeData: protocol.RangeData{
						Start: protocol.Pos{Line: 4, Character: 5},
						End:   protocol.Pos{Line: 6, Character: 7},
					},
				},
				DefinitionResultID: 0,
				ReferenceResultID:  3007,
			},
			2005: {
				Range: reader.Range{
					RangeData: protocol.RangeData{
						Start: protocol.Pos{Line: 5, Character: 6},
						End:   protocol.Pos{Line: 7, Character: 8},
					},
				},
				DefinitionResultID: 3003,
				ReferenceResultID:  0,
			},
			2006: {
				Range: reader.Range{
					RangeData: protocol.RangeData{
						Start: protocol.Pos{Line: 6, Character: 7},
						End:   protocol.Pos{Line: 8, Character: 9},
					},
				},
				DefinitionResultID: 0,
				HoverResultID:      3008,
			},
			2007: {
				Range: reader.Range{
					RangeData: protocol.RangeData{
						Start: protocol.Pos{Line: 7, Character: 8},
						End:   protocol.Pos{Line: 9, Character: 0},
					},
				},
				DefinitionResultID: 3004,
				ReferenceResultID:  0,
			},
			2008: {
				Range: reader.Range{
					RangeData: protocol.RangeData{
						Start: protocol.Pos{Line: 8, Character: 9},
						End:   protocol.Pos{Line: 0, Character: 1},
					},
				},
				DefinitionResultID: 0,
				HoverResultID:      3009,
			},
			2009: {
				Range: reader.Range{
					RangeData: protocol.RangeData{
						Start: protocol.Pos{Line: 9, Character: 0},
						End:   protocol.Pos{Line: 1, Character: 2},
					},
				},
				DefinitionResultID: 3005,
				ReferenceResultID:  0,
			},
		},
		DefinitionData: map[int]*datastructures.DefaultIDSetMap{
			3001: datastructures.DefaultIDSetMapWith(map[int]*datastructures.IDSet{
				1001: datastructures.IDSetWith(2003),
				1002: datastructures.IDSetWith(2004),
				1003: datastructures.IDSetWith(2007),
			}),
			3002: datastructures.DefaultIDSetMapWith(map[int]*datastructures.IDSet{
				1001: datastructures.IDSetWith(2002),
				1002: datastructures.IDSetWith(2005),
				1003: datastructures.IDSetWith(2008),
			}),
			3003: datastructures.DefaultIDSetMapWith(map[int]*datastructures.IDSet{
				1001: datastructures.IDSetWith(2001),
				1002: datastructures.IDSetWith(2006),
				1003: datastructures.IDSetWith(2009),
			}),
			3004: datastructures.DefaultIDSetMapWith(map[int]*datastructures.IDSet{
				1001: datastructures.IDSetWith(2003),
				1002: datastructures.IDSetWith(2005),
				1003: datastructures.IDSetWith(2007),
			}),
			3005: datastructures.DefaultIDSetMapWith(map[int]*datastructures.IDSet{
				1001: datastructures.IDSetWith(2002),
				1002: datastructures.IDSetWith(2006),
				1003: datastructures.IDSetWith(2008),
			}),
		},
		ReferenceData: map[int]*datastructures.DefaultIDSetMap{
			3006: datastructures.DefaultIDSetMapWith(map[int]*datastructures.IDSet{
				1001: datastructures.IDSetWith(2003),
				1003: datastructures.IDSetWith(2007, 2009),
			}),
			3007: datastructures.DefaultIDSetMapWith(map[int]*datastructures.IDSet{
				1001: datastructures.IDSetWith(2002),
				1003: datastructures.IDSetWith(2007, 2009),
			}),
		},
		HoverData: map[int]string{
			3008: "foo",
			3009: "bar",
		},
		MonikerData: map[int]Moniker{
			4001: {
				Moniker: reader.Moniker{
					Kind:       "import",
					Scheme:     "scheme A",
					Identifier: "ident A",
				},
				PackageInformationID: 5001,
			},
			4002: {
				Moniker: reader.Moniker{
					Kind:       "import",
					Scheme:     "scheme B",
					Identifier: "ident B",
				},
				PackageInformationID: 0,
			},
			4003: {
				Moniker: reader.Moniker{
					Kind:       "export",
					Scheme:     "scheme C",
					Identifier: "ident C",
				},
				PackageInformationID: 5002,
			},
			4004: {
				Moniker: reader.Moniker{
					Kind:       "export",
					Scheme:     "scheme D",
					Identifier: "ident D",
				},
				PackageInformationID: 0,
			},
			4005: {
				Moniker: reader.Moniker{
					Kind:       "export",
					Scheme:     "scheme E",
					Identifier: "ident E",
				},
				PackageInformationID: 5003,
			},
			4006: {
				Moniker: reader.Moniker{
					Kind:       "import",
					Scheme:     "scheme E",
					Identifier: "ident E",
				},
				PackageInformationID: 5003,
			},
			4007: {
				Moniker: reader.Moniker{
					Kind:       "implementation",
					Scheme:     "scheme F",
					Identifier: "ident F",
				},
				PackageInformationID: 5002,
			},
		},
		PackageInformationData: map[int]PackageInformation{
			5001: {
				Name:    "pkg A",
				Version: "0.1.0",
			},
			5002: {
				Name:    "pkg B",
				Version: "1.2.3",
			},
			5003: {
				Name:    "pkg C",
				Version: "3.2.1",
			},
		},
		DiagnosticResults: map[int][]Diagnostic{
			1001: {
				{
					Severity:       1,
					Code:           "1234",
					Message:        "m1",
					Source:         "s1",
					StartLine:      11,
					StartCharacter: 12,
					EndLine:        13,
					EndCharacter:   14,
				},
			},
			1002: {
				{
					Severity:       2,
					Code:           "2",
					Message:        "m2",
					Source:         "s2",
					StartLine:      21,
					StartCharacter: 22,
					EndLine:        23,
					EndCharacter:   24,
				},
			},
			1003: {
				{
					Severity:       3,
					Code:           "3234",
					Message:        "m3",
					Source:         "s3",
					StartLine:      31,
					StartCharacter: 32,
					EndLine:        33,
					EndCharacter:   34,
				},
				{
					Severity:       4,
					Code:           "4234",
					Message:        "m4",
					Source:         "s4",
					StartLine:      41,
					StartCharacter: 42,
					EndLine:        43,
					EndCharacter:   44,
				},
			},
		},
		ImportedMonikers:    datastructures.IDSetWith(4001, 4006),
		ExportedMonikers:    datastructures.IDSetWith(4003, 4005),
		ImplementedMonikers: datastructures.NewIDSet(),
		Contains: datastructures.DefaultIDSetMapWith(map[int]*datastructures.IDSet{
			1001: datastructures.IDSetWith(2001, 2002, 2003),
			1002: datastructures.IDSetWith(2004, 2005, 2006),
			1003: datastructures.IDSetWith(2007, 2008, 2009),
		}),
		Monikers: datastructures.DefaultIDSetMapWith(map[int]*datastructures.IDSet{
			2001: datastructures.IDSetWith(4001, 4002),
			2002: datastructures.IDSetWith(4003, 4004, 4007),
		}),
		Diagnostics: datastructures.DefaultIDSetMapWith(map[int]*datastructures.IDSet{
			1001: datastructures.IDSetWith(1001, 1002),
			1002: datastructures.IDSetWith(1003),
		}),
	}

	actualBundleData := groupBundleData(context.Background(), state)

	expectedMetaData := precise.MetaData{
		NumResultChunks: 1,
	}
	if diff := cmp.Diff(expectedMetaData, actualBundleData.Meta); diff != "" {
		t.Errorf("unexpected meta data (-want +got):\n%s", diff)
	}

	expectedPackages := []precise.Package{
		{Scheme: "scheme C", Name: "pkg B", Version: "1.2.3"},
		{Scheme: "scheme E", Name: "pkg C", Version: "3.2.1"},
	}
	sort.Slice(actualBundleData.Packages, func(i, j int) bool {
		return actualBundleData.Packages[i].Scheme < actualBundleData.Packages[j].Scheme
	})
	if diff := cmp.Diff(expectedPackages, actualBundleData.Packages); diff != "" {
		t.Errorf("unexpected packages (-want +got):\n%s", diff)
	}

	expectedPackageReferences := []precise.PackageReference{
		{
			Package: precise.Package{Scheme: "scheme A", Name: "pkg A", Version: "0.1.0"},
		},
	}
	sort.Slice(actualBundleData.PackageReferences, func(i, j int) bool {
		return actualBundleData.PackageReferences[i].Scheme < actualBundleData.PackageReferences[j].Scheme
	})
	if diff := cmp.Diff(expectedPackageReferences, actualBundleData.PackageReferences); diff != "" {
		t.Errorf("unexpected package references (-want +got):\n%s", diff)
	}

	documents := map[string]precise.DocumentData{}
	for v := range actualBundleData.Documents {
		documents[v.Path] = v.Document
	}
	for _, document := range documents {
		sortDiagnostics(document.Diagnostics)

		for _, r := range document.Ranges {
			sortMonikerIDs(r.MonikerIDs)
		}
	}

	expectedDocumentData := map[string]precise.DocumentData{
		"foo.go": {
			Ranges: map[precise.ID]precise.RangeData{
				"2001": {
					StartLine:          1,
					StartCharacter:     2,
					EndLine:            3,
					EndCharacter:       4,
					DefinitionResultID: "",
					ReferenceResultID:  "3006",
					HoverResultID:      "",
					MonikerIDs:         []precise.ID{"4001", "4002"},
				},
				"2002": {
					StartLine:          2,
					StartCharacter:     3,
					EndLine:            4,
					EndCharacter:       5,
					DefinitionResultID: "3001",
					ReferenceResultID:  "",
					HoverResultID:      "",
					MonikerIDs:         []precise.ID{"4003", "4004", "4007"},
				},
				"2003": {
					StartLine:          3,
					StartCharacter:     4,
					EndLine:            5,
					EndCharacter:       6,
					DefinitionResultID: "3002",
					ReferenceResultID:  "",
					HoverResultID:      "",
					MonikerIDs:         []precise.ID{},
				},
			},
			HoverResults: map[precise.ID]string{},
			Monikers: map[precise.ID]precise.MonikerData{
				"4001": {
					Kind:                 "import",
					Scheme:               "scheme A",
					Identifier:           "ident A",
					PackageInformationID: "5001",
				},
				"4002": {
					Kind:                 "import",
					Scheme:               "scheme B",
					Identifier:           "ident B",
					PackageInformationID: "",
				},
				"4003": {
					Kind:                 "export",
					Scheme:               "scheme C",
					Identifier:           "ident C",
					PackageInformationID: "5002",
				},
				"4004": {
					Kind:                 "export",
					Scheme:               "scheme D",
					Identifier:           "ident D",
					PackageInformationID: "",
				},
				"4007": {
					Kind:                 "implementation",
					Scheme:               "scheme F",
					Identifier:           "ident F",
					PackageInformationID: "5002",
				},
			},
			PackageInformation: map[precise.ID]precise.PackageInformationData{
				"5001": {
					Name:    "pkg A",
					Version: "0.1.0",
				},
				"5002": {
					Name:    "pkg B",
					Version: "1.2.3",
				},
			},
			Diagnostics: []precise.DiagnosticData{
				{
					Severity:       1,
					Code:           "1234",
					Message:        "m1",
					Source:         "s1",
					StartLine:      11,
					StartCharacter: 12,
					EndLine:        13,
					EndCharacter:   14,
				},
				{
					Severity:       2,
					Code:           "2",
					Message:        "m2",
					Source:         "s2",
					StartLine:      21,
					StartCharacter: 22,
					EndLine:        23,
					EndCharacter:   24,
				},
			},
		},
		"bar.go": {
			Ranges: map[precise.ID]precise.RangeData{
				"2004": {
					StartLine:          4,
					StartCharacter:     5,
					EndLine:            6,
					EndCharacter:       7,
					DefinitionResultID: "",
					ReferenceResultID:  "3007",
					HoverResultID:      "",
					MonikerIDs:         []precise.ID{},
				},
				"2005": {
					StartLine:          5,
					StartCharacter:     6,
					EndLine:            7,
					EndCharacter:       8,
					DefinitionResultID: "3003",
					ReferenceResultID:  "",
					HoverResultID:      "",
					MonikerIDs:         []precise.ID{},
				},
				"2006": {
					StartLine:          6,
					StartCharacter:     7,
					EndLine:            8,
					EndCharacter:       9,
					DefinitionResultID: "",
					ReferenceResultID:  "",
					HoverResultID:      "3008",
					MonikerIDs:         []precise.ID{},
				},
			},
			HoverResults:       map[precise.ID]string{"3008": "foo"},
			Monikers:           map[precise.ID]precise.MonikerData{},
			PackageInformation: map[precise.ID]precise.PackageInformationData{},
			Diagnostics: []precise.DiagnosticData{
				{
					Severity:       3,
					Code:           "3234",
					Message:        "m3",
					Source:         "s3",
					StartLine:      31,
					StartCharacter: 32,
					EndLine:        33,
					EndCharacter:   34,
				},
				{
					Severity:       4,
					Code:           "4234",
					Message:        "m4",
					Source:         "s4",
					StartLine:      41,
					StartCharacter: 42,
					EndLine:        43,
					EndCharacter:   44,
				},
			},
		},
		"baz.go": {
			Ranges: map[precise.ID]precise.RangeData{
				"2007": {
					StartLine:          7,
					StartCharacter:     8,
					EndLine:            9,
					EndCharacter:       0,
					DefinitionResultID: "3004",
					ReferenceResultID:  "",
					HoverResultID:      "",
					MonikerIDs:         []precise.ID{},
				},
				"2008": {
					StartLine:          8,
					StartCharacter:     9,
					EndLine:            0,
					EndCharacter:       1,
					DefinitionResultID: "",
					ReferenceResultID:  "",
					HoverResultID:      "3009",
					MonikerIDs:         []precise.ID{},
				},
				"2009": {
					StartLine:          9,
					StartCharacter:     0,
					EndLine:            1,
					EndCharacter:       2,
					DefinitionResultID: "3005",
					ReferenceResultID:  "",
					HoverResultID:      "",
					MonikerIDs:         []precise.ID{},
				},
			},
			HoverResults:       map[precise.ID]string{"3009": "bar"},
			Monikers:           map[precise.ID]precise.MonikerData{},
			PackageInformation: map[precise.ID]precise.PackageInformationData{},
			Diagnostics:        []precise.DiagnosticData{},
		},
	}
	if diff := cmp.Diff(expectedDocumentData, documents, datastructures.Comparers...); diff != "" {
		t.Errorf("unexpected document data (-want +got):\n%s", diff)
	}

	resultChunkData := map[int]precise.ResultChunkData{}
	for v := range actualBundleData.ResultChunks {
		resultChunkData[v.Index] = v.ResultChunk
	}
	for _, resultChunk := range resultChunkData {
		for _, documentRanges := range resultChunk.DocumentIDRangeIDs {
			sortDocumentIDRangeIDs(documentRanges)
		}
	}

	expectedResultChunkData := map[int]precise.ResultChunkData{
		0: {
			DocumentPaths: map[precise.ID]string{
				"1001": "foo.go",
				"1002": "bar.go",
				"1003": "baz.go",
			},
			DocumentIDRangeIDs: map[precise.ID][]precise.DocumentIDRangeID{
				"3001": {
					{DocumentID: "1001", RangeID: "2003"},
					{DocumentID: "1002", RangeID: "2004"},
					{DocumentID: "1003", RangeID: "2007"},
				},
				"3002": {
					{DocumentID: "1001", RangeID: "2002"},
					{DocumentID: "1002", RangeID: "2005"},
					{DocumentID: "1003", RangeID: "2008"},
				},
				"3003": {
					{DocumentID: "1001", RangeID: "2001"},
					{DocumentID: "1002", RangeID: "2006"},
					{DocumentID: "1003", RangeID: "2009"},
				},
				"3004": {
					{DocumentID: "1001", RangeID: "2003"},
					{DocumentID: "1002", RangeID: "2005"},
					{DocumentID: "1003", RangeID: "2007"},
				},
				"3005": {
					{DocumentID: "1001", RangeID: "2002"},
					{DocumentID: "1002", RangeID: "2006"},
					{DocumentID: "1003", RangeID: "2008"},
				},
				"3006": {
					{DocumentID: "1001", RangeID: "2003"},
					{DocumentID: "1003", RangeID: "2007"},
					{DocumentID: "1003", RangeID: "2009"},
				},
				"3007": {
					{DocumentID: "1001", RangeID: "2002"},
					{DocumentID: "1003", RangeID: "2007"},
					{DocumentID: "1003", RangeID: "2009"},
				},
			},
		},
	}
	if diff := cmp.Diff(expectedResultChunkData, resultChunkData); diff != "" {
		t.Errorf("unexpected result chunk data (-want +got):\n%s", diff)
	}

	var definitions []precise.MonikerLocations
	for v := range actualBundleData.Definitions {
		definitions = append(definitions, v)
	}
	sortMonikerLocations(definitions)

	expectedDefinitions := []precise.MonikerLocations{
		{
			Kind:       "export",
			Scheme:     "scheme C",
			Identifier: "ident C",
			Locations: []precise.LocationData{
				{DocumentPath: "bar.go", StartLine: 4, StartCharacter: 5, EndLine: 6, EndCharacter: 7},
				{DocumentPath: "baz.go", StartLine: 7, StartCharacter: 8, EndLine: 9, EndCharacter: 0},
				{DocumentPath: "foo.go", StartLine: 3, StartCharacter: 4, EndLine: 5, EndCharacter: 6},
			},
		},
		{
			Kind:       "export",
			Scheme:     "scheme D",
			Identifier: "ident D",
			Locations: []precise.LocationData{
				{DocumentPath: "bar.go", StartLine: 4, StartCharacter: 5, EndLine: 6, EndCharacter: 7},
				{DocumentPath: "baz.go", StartLine: 7, StartCharacter: 8, EndLine: 9, EndCharacter: 0},
				{DocumentPath: "foo.go", StartLine: 3, StartCharacter: 4, EndLine: 5, EndCharacter: 6},
			},
		},
	}
	if diff := cmp.Diff(expectedDefinitions, definitions); diff != "" {
		t.Errorf("unexpected definitions (-want +got):\n%s", diff)
	}

	var references []precise.MonikerLocations
	for v := range actualBundleData.References {
		references = append(references, v)
	}
	sortMonikerLocations(references)

	expectedReferences := []precise.MonikerLocations{
		{
			Kind:       "import",
			Scheme:     "scheme A",
			Identifier: "ident A",
			Locations: []precise.LocationData{
				{DocumentPath: "baz.go", StartLine: 7, StartCharacter: 8, EndLine: 9, EndCharacter: 0},
				{DocumentPath: "baz.go", StartLine: 9, StartCharacter: 0, EndLine: 1, EndCharacter: 2},
				{DocumentPath: "foo.go", StartLine: 3, StartCharacter: 4, EndLine: 5, EndCharacter: 6},
			},
		},
		{
			Kind:       "import",
			Scheme:     "scheme B",
			Identifier: "ident B",
			Locations: []precise.LocationData{
				{DocumentPath: "baz.go", StartLine: 7, StartCharacter: 8, EndLine: 9, EndCharacter: 0},
				{DocumentPath: "baz.go", StartLine: 9, StartCharacter: 0, EndLine: 1, EndCharacter: 2},
				{DocumentPath: "foo.go", StartLine: 3, StartCharacter: 4, EndLine: 5, EndCharacter: 6},
			},
		},
	}
	if diff := cmp.Diff(expectedReferences, references); diff != "" {
		t.Errorf("unexpected references (-want +got):\n%s", diff)
	}

	var implementations []precise.MonikerLocations
	for v := range actualBundleData.Implementations {
		implementations = append(implementations, v)
	}
	sortMonikerLocations(implementations)

	expectedImplementations := []precise.MonikerLocations{
		{
			Kind:       "implementation",
			Scheme:     "scheme F",
			Identifier: "ident F",
			Locations: []precise.LocationData{
				{DocumentPath: "bar.go", StartLine: 4, StartCharacter: 5, EndLine: 6, EndCharacter: 7},
				{DocumentPath: "baz.go", StartLine: 7, StartCharacter: 8, EndLine: 9, EndCharacter: 0},
				{DocumentPath: "foo.go", StartLine: 3, StartCharacter: 4, EndLine: 5, EndCharacter: 6},
			},
		},
	}
	if diff := cmp.Diff(expectedImplementations, implementations); diff != "" {
		t.Errorf("unexpected implementations (-want +got):\n%s", diff)
	}
}

//
//

func sortMonikerIDs(s []precise.ID) {
	sort.Slice(s, func(i, j int) bool {
		return strings.Compare(string(s[i]), string(s[j])) < 0
	})
}

func sortDiagnostics(s []precise.DiagnosticData) {
	sort.Slice(s, func(i, j int) bool {
		return strings.Compare(s[i].Message, s[j].Message) < 0
	})
}

func sortDocumentIDRangeIDs(s []precise.DocumentIDRangeID) {
	sort.Slice(s, func(i, j int) bool {
		if compareResult := strings.Compare(string(s[i].DocumentID), string(s[j].DocumentID)); compareResult != 0 {
			return compareResult < 0
		} else {
			return strings.Compare(string(s[i].RangeID), string(s[j].RangeID)) < 0
		}
	})
}

func sortMonikerLocations(monikerLocations []precise.MonikerLocations) {
	sort.Slice(monikerLocations, func(i, j int) bool {
		if compareResult := strings.Compare(monikerLocations[i].Scheme, monikerLocations[j].Scheme); compareResult != 0 {
			return compareResult < 0
		} else if compareResult := strings.Compare(monikerLocations[i].Identifier, monikerLocations[j].Identifier); compareResult != 0 {
			return compareResult < 0
		}
		return false
	})

	for _, ml := range monikerLocations {
		sortLocations(ml.Locations)
	}
}

func sortLocations(locations []precise.LocationData) {
	sort.Slice(locations, func(i, j int) bool {
		if compareResult := strings.Compare(locations[i].DocumentPath, locations[j].DocumentPath); compareResult != 0 {
			return compareResult < 0
		}

		return locations[i].StartLine < locations[j].StartLine
	})
}
