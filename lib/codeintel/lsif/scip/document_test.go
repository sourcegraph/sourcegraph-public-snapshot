package scip

import (
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/sourcegraph/scip/bindings/go/scip"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

func TestConvertLSIFDocument(t *testing.T) {
	r0 := precise.RangeData{
		StartLine:              50,
		StartCharacter:         25,
		EndLine:                50,
		EndCharacter:           35,
		DefinitionResultID:     "d0",
		ReferenceResultID:      "",
		ImplementationResultID: "",
		HoverResultID:          "h0",
		MonikerIDs:             []precise.ID{"m0"},
	}
	r1 := precise.RangeData{
		StartLine:              51,
		StartCharacter:         25,
		EndLine:                51,
		EndCharacter:           35,
		DefinitionResultID:     "d0",
		ReferenceResultID:      "",
		ImplementationResultID: "",
		HoverResultID:          "h0",
		MonikerIDs:             []precise.ID{"m0"},
	}
	r2 := precise.RangeData{
		StartLine:              52,
		StartCharacter:         25,
		EndLine:                52,
		EndCharacter:           35,
		DefinitionResultID:     "d0",
		ReferenceResultID:      "",
		ImplementationResultID: "",
		HoverResultID:          "h0",
		MonikerIDs:             []precise.ID{"m0"},
	}
	m0 := precise.MonikerData{
		Kind:                 "import",
		Scheme:               "node",
		Identifier:           "padItUp",
		PackageInformationID: "p0",
	}
	p0 := precise.PackageInformationData{
		Manager: "npm",
		Name:    "left-pad",
		Version: "0.1.0",
	}
	document := precise.DocumentData{
		Ranges: map[precise.ID]precise.RangeData{
			"r0": r0,
			"r1": r1,
			"r2": r2,
		},
		HoverResults: map[precise.ID]string{
			"h0": "hello world",
		},
		Monikers: map[precise.ID]precise.MonikerData{
			"m0": m0,
		},
		PackageInformation: map[precise.ID]precise.PackageInformationData{
			"p0": p0,
		},
		Diagnostics: []precise.DiagnosticData{},
	}

	targetRangeFetcher := func(resultID precise.ID) []precise.ID {
		if resultID == "d0" {
			return []precise.ID{"r0"}
		}

		return nil
	}
	scipDocument := ConvertLSIFDocument(42, targetRangeFetcher, "lsif-sml", "src/main.ml", document)

	if expected := "src/main.ml"; scipDocument.RelativePath != expected {
		t.Fatalf("unexpected path. want=%s have=%s", expected, scipDocument.RelativePath)
	}
	if expected := "SML"; scipDocument.Language != expected {
		t.Fatalf("unexpected language. want=%s have=%s", expected, scipDocument.Language)
	}

	expectedOccurrences := []*scip.Occurrence{
		{Range: []int32{50, 25, 50, 35}, Symbol: "lsif . 42 . `r0`.", SymbolRoles: int32(scip.SymbolRole_Definition)},
		{Range: []int32{50, 25, 50, 35}, Symbol: "node npm left-pad 0.1.0 `padItUp`.", SymbolRoles: int32(scip.SymbolRole_Definition)},
		{Range: []int32{51, 25, 51, 35}, Symbol: "lsif . 42 . `r0`."},
		{Range: []int32{51, 25, 51, 35}, Symbol: "node npm left-pad 0.1.0 `padItUp`."},
		{Range: []int32{52, 25, 52, 35}, Symbol: "lsif . 42 . `r0`."},
		{Range: []int32{52, 25, 52, 35}, Symbol: "node npm left-pad 0.1.0 `padItUp`."},
	}
	sort.Slice(scipDocument.Occurrences, func(i, j int) bool {
		oi := scipDocument.Occurrences[i]
		oj := scipDocument.Occurrences[j]

		if oi.Range[0] == oj.Range[0] {
			return oi.Symbol[0] < oj.Symbol[0]
		}

		return oi.Range[0] < oj.Range[0]
	})
	if diff := cmp.Diff(expectedOccurrences, scipDocument.Occurrences, cmpopts.IgnoreUnexported(scip.Occurrence{})); diff != "" {
		t.Errorf("unexpected occurrences (-want +got):\n%s", diff)
	}

	expectedSymbols := []*scip.SymbolInformation{
		{Symbol: "lsif . 42 . `r0`.", Documentation: []string{"hello world"}, Relationships: nil},
		{Symbol: "node npm left-pad 0.1.0 `padItUp`.", Documentation: []string{"hello world"}, Relationships: nil},
	}
	sort.Slice(scipDocument.Symbols, func(i, j int) bool {
		return scipDocument.Symbols[i].Symbol < scipDocument.Symbols[j].Symbol
	})
	if diff := cmp.Diff(expectedSymbols, scipDocument.Symbols, cmpopts.IgnoreUnexported(scip.SymbolInformation{})); diff != "" {
		t.Errorf("unexpected symbols (-want +got):\n%s", diff)
	}
}
