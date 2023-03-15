package scip

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"os"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/sourcegraph/scip/bindings/go/scip"
)

func TestConvertLSIF(t *testing.T) {
	gzipped, err := os.Open("./testdata/dump1.lsif.gz")
	if err != nil {
		t.Fatalf("unexpected error reading test file: %s", err)
	}
	defer gzipped.Close()
	r, err := gzip.NewReader(gzipped)
	if err != nil {
		t.Fatalf("unexpected error unzipping test file: %s", err)
	}
	input, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("unexpected error reading test file: %s", err)
	}

	ctx := context.Background()
	scipIndex, err := ConvertLSIF(ctx, 42, bytes.NewReader(input), "root/")
	if err != nil {
		t.Fatalf("unexpected error converting LSIF data: %s", err)
	}

	expectedIndex := &scip.Index{
		Metadata: &scip.Metadata{
			ProjectRoot:          "file:///test/root/",
			ToolInfo:             &scip.ToolInfo{Name: "lsif-test"},
			TextDocumentEncoding: scip.TextEncoding_UnspecifiedTextEncoding,
			Version:              scip.ProtocolVersion_UnspecifiedProtocolVersion,
		},
		Documents: []*scip.Document{
			{
				RelativePath: "bar.go",
				Occurrences: []*scip.Occurrence{
					{Range: []int32{4, 5, 6, 7}, Symbol: "scheme A . pkg A v0.1.0 `ident A`."},
					{Range: []int32{6, 7, 8, 9}, Symbol: "lsif . 42 . `7`."},
					{Range: []int32{6, 7, 8, 9}, Symbol: "scheme B . pkg B v1.2.3 `ident B`."},
				},
				Symbols: []*scip.SymbolInformation{
					{Symbol: "lsif . 42 . `7`.", Documentation: []string{"```go\ntext A\n```"}},
					{Symbol: "scheme A . pkg A v0.1.0 `ident A`."},
					{Symbol: "scheme B . pkg B v1.2.3 `ident B`.", Documentation: []string{"```go\ntext A\n```"}},
				},
			},
			{
				RelativePath: "foo.go",
				Occurrences: []*scip.Occurrence{
					{Range: []int32{1, 5, 1, 6}, Diagnostics: []*scip.Diagnostic{
						{
							Severity: scip.Severity_Error,
							Code:     "2322",
							Message:  "Type '10' is not assignable to type 'string'.",
							Source:   "eslint",
						},
					}},
					{Range: []int32{1, 2, 3, 4}, Symbol: "lsif . 42 . `8`."},
					{Range: []int32{3, 4, 5, 6}, Symbol: "lsif . 42 . `8`."},
				},
				Symbols: []*scip.SymbolInformation{
					{Symbol: "lsif . 42 . `8`.", Documentation: []string{"```go\ntext B\n```"}},
				},
			},
		},
	}
	sort.Slice(scipIndex.Documents, func(i, j int) bool {
		return scipIndex.Documents[i].RelativePath < scipIndex.Documents[j].RelativePath
	})
	for _, document := range scipIndex.Documents {
		sort.Slice(document.Occurrences, func(i, j int) bool {
			oi := document.Occurrences[i]
			oj := document.Occurrences[j]

			if oi.Range[0] == oj.Range[0] {
				return oi.Symbol < oj.Symbol
			}

			return oi.Range[0] < oj.Range[0]
		})
	}
	for _, document := range scipIndex.Documents {
		sort.Slice(document.Symbols, func(i, j int) bool {
			return document.Symbols[i].Symbol < document.Symbols[j].Symbol
		})
	}
	if diff := cmp.Diff(expectedIndex, scipIndex, cmpopts.IgnoreUnexported(
		scip.Index{},
		scip.Metadata{},
		scip.ToolInfo{},
		scip.Document{},
		scip.Occurrence{},
		scip.SymbolInformation{},
		scip.Diagnostic{},
	)); diff != "" {
		t.Errorf("unexpected index (-want +got):\n%s", diff)
	}
}
