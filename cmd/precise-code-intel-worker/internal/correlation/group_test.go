package correlation

import (
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-worker/internal/correlation/datastructures"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-worker/internal/correlation/lsif"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bloomfilter"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/types"
)

func TestConvert(t *testing.T) {
	state := &State{
		LSIFVersion: "0.4.3",
		DocumentData: map[string]lsif.Document{
			"d01": {URI: "foo.go", Contains: datastructures.IDSet{"r01": {}, "r02": {}, "r03": {}}},
			"d02": {URI: "bar.go", Contains: datastructures.IDSet{"r04": {}, "r05": {}, "r06": {}}},
			"d03": {URI: "baz.go", Contains: datastructures.IDSet{"r07": {}, "r08": {}, "r09": {}}},
		},
		RangeData: map[string]lsif.Range{
			"r01": {StartLine: 1, StartCharacter: 2, EndLine: 3, EndCharacter: 4, DefinitionResultID: "x01", MonikerIDs: datastructures.IDSet{"m01": {}, "m02": {}}},
			"r02": {StartLine: 2, StartCharacter: 3, EndLine: 4, EndCharacter: 5, ReferenceResultID: "x06", MonikerIDs: datastructures.IDSet{"m03": {}, "m04": {}}},
			"r03": {StartLine: 3, StartCharacter: 4, EndLine: 5, EndCharacter: 6, DefinitionResultID: "x02"},
			"r04": {StartLine: 4, StartCharacter: 5, EndLine: 6, EndCharacter: 7, ReferenceResultID: "x07"},
			"r05": {StartLine: 5, StartCharacter: 6, EndLine: 7, EndCharacter: 8, DefinitionResultID: "x03"},
			"r06": {StartLine: 6, StartCharacter: 7, EndLine: 8, EndCharacter: 9, HoverResultID: "x08"},
			"r07": {StartLine: 7, StartCharacter: 8, EndLine: 9, EndCharacter: 0, DefinitionResultID: "x04"},
			"r08": {StartLine: 8, StartCharacter: 9, EndLine: 0, EndCharacter: 1, HoverResultID: "x09"},
			"r09": {StartLine: 9, StartCharacter: 0, EndLine: 1, EndCharacter: 2, DefinitionResultID: "x05"},
		},
		DefinitionData: map[string]datastructures.DefaultIDSetMap{
			"x01": {"d01": {"r03": {}}, "d02": {"r04": {}}, "d03": {"r07": {}}},
			"x02": {"d01": {"r02": {}}, "d02": {"r05": {}}, "d03": {"r08": {}}},
			"x03": {"d01": {"r01": {}}, "d02": {"r06": {}}, "d03": {"r09": {}}},
			"x04": {"d01": {"r03": {}}, "d02": {"r05": {}}, "d03": {"r07": {}}},
			"x05": {"d01": {"r02": {}}, "d02": {"r06": {}}, "d03": {"r08": {}}},
		},
		ReferenceData: map[string]datastructures.DefaultIDSetMap{
			"x06": {"d01": {"r03": {}}, "d03": {"r07": {}, "r09": {}}},
			"x07": {"d01": {"r02": {}}, "d03": {"r07": {}, "r09": {}}},
		},
		HoverData: map[string]string{
			"x08": "foo",
			"x09": "bar",
		},
		MonikerData: map[string]lsif.Moniker{
			"m01": {Kind: "import", Scheme: "scheme A", Identifier: "ident A", PackageInformationID: "p01"},
			"m02": {Kind: "import", Scheme: "scheme B", Identifier: "ident B"},
			"m03": {Kind: "export", Scheme: "scheme C", Identifier: "ident C", PackageInformationID: "p02"},
			"m04": {Kind: "export", Scheme: "scheme D", Identifier: "ident D"},
		},
		PackageInformationData: map[string]lsif.PackageInformation{
			"p01": {Name: "pkg A", Version: "0.1.0"},
			"p02": {Name: "pkg B", Version: "1.2.3"},
		},
		ImportedMonikers: datastructures.IDSet{"m01": {}},
		ExportedMonikers: datastructures.IDSet{"m03": {}},
	}

	actualBundleData, err := groupBundleData(state, 42)
	if err != nil {
		t.Fatalf("unexpected error converting correlation state to types: %s", err)
	}
	// Ensure arrays have deterministic order so we can compare with a canned expected object structure
	normalizeGroupedBundleData(actualBundleData)

	expectedFilter, err := bloomfilter.CreateFilter([]string{"ident A"})
	if err != nil {
		t.Fatalf("unexpected error creating bloom filter: %s", err)
	}

	expectedBundleData := &GroupedBundleData{
		LSIFVersion:     "0.4.3",
		NumResultChunks: 1,
		Documents: map[string]types.DocumentData{
			"foo.go": {
				Ranges: map[types.ID]types.RangeData{
					"r01": {StartLine: 1, StartCharacter: 2, EndLine: 3, EndCharacter: 4, DefinitionResultID: "x01", MonikerIDs: []types.ID{"m01", "m02"}},
					"r02": {StartLine: 2, StartCharacter: 3, EndLine: 4, EndCharacter: 5, ReferenceResultID: "x06", MonikerIDs: []types.ID{"m03", "m04"}},
					"r03": {StartLine: 3, StartCharacter: 4, EndLine: 5, EndCharacter: 6, DefinitionResultID: "x02"},
				},
				HoverResults: map[types.ID]string{},
				Monikers: map[types.ID]types.MonikerData{
					"m01": {Kind: "import", Scheme: "scheme A", Identifier: "ident A", PackageInformationID: "p01"},
					"m02": {Kind: "import", Scheme: "scheme B", Identifier: "ident B"},
					"m03": {Kind: "export", Scheme: "scheme C", Identifier: "ident C", PackageInformationID: "p02"},
					"m04": {Kind: "export", Scheme: "scheme D", Identifier: "ident D"},
				},
				PackageInformation: map[types.ID]types.PackageInformationData{
					"p01": {Name: "pkg A", Version: "0.1.0"},
					"p02": {Name: "pkg B", Version: "1.2.3"},
				},
			},
			"bar.go": {
				Ranges: map[types.ID]types.RangeData{
					"r04": {StartLine: 4, StartCharacter: 5, EndLine: 6, EndCharacter: 7, ReferenceResultID: "x07"},
					"r05": {StartLine: 5, StartCharacter: 6, EndLine: 7, EndCharacter: 8, DefinitionResultID: "x03"},
					"r06": {StartLine: 6, StartCharacter: 7, EndLine: 8, EndCharacter: 9, HoverResultID: "x08"},
				},
				HoverResults:       map[types.ID]string{"x08": "foo"},
				Monikers:           map[types.ID]types.MonikerData{},
				PackageInformation: map[types.ID]types.PackageInformationData{},
			},
			"baz.go": {
				Ranges: map[types.ID]types.RangeData{
					"r07": {StartLine: 7, StartCharacter: 8, EndLine: 9, EndCharacter: 0, DefinitionResultID: "x04"},
					"r08": {StartLine: 8, StartCharacter: 9, EndLine: 0, EndCharacter: 1, HoverResultID: "x09"},
					"r09": {StartLine: 9, StartCharacter: 0, EndLine: 1, EndCharacter: 2, DefinitionResultID: "x05"},
				},
				HoverResults:       map[types.ID]string{"x09": "bar"},
				Monikers:           map[types.ID]types.MonikerData{},
				PackageInformation: map[types.ID]types.PackageInformationData{},
			},
		},
		ResultChunks: map[int]types.ResultChunkData{
			0: {
				DocumentPaths: map[types.ID]string{
					"d01": "foo.go",
					"d02": "bar.go",
					"d03": "baz.go",
				},
				DocumentIDRangeIDs: map[types.ID][]types.DocumentIDRangeID{
					"x01": {
						{DocumentID: "d01", RangeID: "r03"},
						{DocumentID: "d02", RangeID: "r04"},
						{DocumentID: "d03", RangeID: "r07"},
					},
					"x02": {
						{DocumentID: "d01", RangeID: "r02"},
						{DocumentID: "d02", RangeID: "r05"},
						{DocumentID: "d03", RangeID: "r08"},
					},
					"x03": {
						{DocumentID: "d01", RangeID: "r01"},
						{DocumentID: "d02", RangeID: "r06"},
						{DocumentID: "d03", RangeID: "r09"},
					},
					"x04": {
						{DocumentID: "d01", RangeID: "r03"},
						{DocumentID: "d02", RangeID: "r05"},
						{DocumentID: "d03", RangeID: "r07"},
					},
					"x05": {
						{DocumentID: "d01", RangeID: "r02"},
						{DocumentID: "d02", RangeID: "r06"},
						{DocumentID: "d03", RangeID: "r08"},
					},
					"x06": {
						{DocumentID: "d01", RangeID: "r03"},
						{DocumentID: "d03", RangeID: "r07"},
						{DocumentID: "d03", RangeID: "r09"},
					},
					"x07": {
						{DocumentID: "d01", RangeID: "r02"},
						{DocumentID: "d03", RangeID: "r07"},
						{DocumentID: "d03", RangeID: "r09"},
					},
				},
			},
		},
		Definitions: []types.MonikerLocations{
			{
				Scheme:     "scheme A",
				Identifier: "ident A",
				Locations: []types.Location{
					{URI: "bar.go", StartLine: 4, StartCharacter: 5, EndLine: 6, EndCharacter: 7},
					{URI: "baz.go", StartLine: 7, StartCharacter: 8, EndLine: 9, EndCharacter: 0},
					{URI: "foo.go", StartLine: 3, StartCharacter: 4, EndLine: 5, EndCharacter: 6},
				},
			},
			{
				Scheme:     "scheme B",
				Identifier: "ident B",
				Locations: []types.Location{
					{URI: "bar.go", StartLine: 4, StartCharacter: 5, EndLine: 6, EndCharacter: 7},
					{URI: "baz.go", StartLine: 7, StartCharacter: 8, EndLine: 9, EndCharacter: 0},
					{URI: "foo.go", StartLine: 3, StartCharacter: 4, EndLine: 5, EndCharacter: 6},
				},
			},
		},
		References: []types.MonikerLocations{
			{
				Scheme:     "scheme C",
				Identifier: "ident C",
				Locations: []types.Location{
					{URI: "baz.go", StartLine: 7, StartCharacter: 8, EndLine: 9, EndCharacter: 0},
					{URI: "baz.go", StartLine: 9, StartCharacter: 0, EndLine: 1, EndCharacter: 2},
					{URI: "foo.go", StartLine: 3, StartCharacter: 4, EndLine: 5, EndCharacter: 6},
				},
			},
			{
				Scheme:     "scheme D",
				Identifier: "ident D",
				Locations: []types.Location{
					{URI: "baz.go", StartLine: 7, StartCharacter: 8, EndLine: 9, EndCharacter: 0},
					{URI: "baz.go", StartLine: 9, StartCharacter: 0, EndLine: 1, EndCharacter: 2},
					{URI: "foo.go", StartLine: 3, StartCharacter: 4, EndLine: 5, EndCharacter: 6},
				},
			},
		},
		Packages: []types.Package{
			{DumpID: 42, Scheme: "scheme C", Name: "pkg B", Version: "1.2.3"},
		},
		PackageReferences: []types.PackageReference{
			{DumpID: 42, Scheme: "scheme A", Name: "pkg A", Version: "0.1.0", Filter: expectedFilter},
		},
	}

	if diff := cmp.Diff(expectedBundleData, actualBundleData); diff != "" {
		t.Errorf("unexpected bundle data (-want +got):\n%s", diff)
	}
}

//
//

func normalizeGroupedBundleData(groupedBundleData *GroupedBundleData) {
	for _, document := range groupedBundleData.Documents {
		for _, r := range document.Ranges {
			sortMonikerIDs(r.MonikerIDs)
		}
	}

	for _, resultChunk := range groupedBundleData.ResultChunks {
		for _, documentRanges := range resultChunk.DocumentIDRangeIDs {
			sortDocumentIDRangeIDs(documentRanges)
		}
	}

	sortMonikerLocations(groupedBundleData.Definitions)
	sortMonikerLocations(groupedBundleData.References)
}

func sortMonikerIDs(s []types.ID) {
	sort.Slice(s, func(i, j int) bool {
		return strings.Compare(string(s[i]), string(s[j])) < 0
	})
}

func sortDocumentIDRangeIDs(s []types.DocumentIDRangeID) {
	sort.Slice(s, func(i, j int) bool {
		if cmp := strings.Compare(string(s[i].DocumentID), string(s[j].DocumentID)); cmp != 0 {
			return cmp < 0
		} else {
			return strings.Compare(string(s[i].RangeID), string(s[j].RangeID)) < 0
		}
	})
}

func sortMonikerLocations(monikerLocations []types.MonikerLocations) {
	sort.Slice(monikerLocations, func(i, j int) bool {
		if cmp := strings.Compare(monikerLocations[i].Scheme, monikerLocations[j].Scheme); cmp != 0 {
			return cmp < 0
		} else if cmp := strings.Compare(monikerLocations[i].Identifier, monikerLocations[j].Identifier); cmp != 0 {
			return cmp < 0
		}
		return false
	})

	for _, ml := range monikerLocations {
		sortLocations(ml.Locations)
	}
}

func sortLocations(locations []types.Location) {
	sort.Slice(locations, func(i, j int) bool {
		if cmp := strings.Compare(locations[i].URI, locations[j].URI); cmp != 0 {
			return cmp < 0
		}

		return locations[i].StartLine < locations[j].StartLine
	})
}
