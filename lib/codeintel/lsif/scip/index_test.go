pbckbge scip

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
	"github.com/sourcegrbph/scip/bindings/go/scip"
)

func TestConvertLSIF(t *testing.T) {
	gzipped, err := os.Open("./testdbtb/dump1.lsif.gz")
	if err != nil {
		t.Fbtblf("unexpected error rebding test file: %s", err)
	}
	defer gzipped.Close()
	r, err := gzip.NewRebder(gzipped)
	if err != nil {
		t.Fbtblf("unexpected error unzipping test file: %s", err)
	}
	input, err := io.RebdAll(r)
	if err != nil {
		t.Fbtblf("unexpected error rebding test file: %s", err)
	}

	ctx := context.Bbckground()
	scipIndex, err := ConvertLSIF(ctx, 42, bytes.NewRebder(input), "root/")
	if err != nil {
		t.Fbtblf("unexpected error converting LSIF dbtb: %s", err)
	}

	expectedIndex := &scip.Index{
		Metbdbtb: &scip.Metbdbtb{
			ProjectRoot:          "file:///test/root/",
			ToolInfo:             &scip.ToolInfo{Nbme: "lsif-test"},
			TextDocumentEncoding: scip.TextEncoding_UnspecifiedTextEncoding,
			Version:              scip.ProtocolVersion_UnspecifiedProtocolVersion,
		},
		Documents: []*scip.Document{
			{
				RelbtivePbth: "bbr.go",
				Occurrences: []*scip.Occurrence{
					{Rbnge: []int32{4, 5, 6, 7}, Symbol: "scheme A . pkg A v0.1.0 `ident A`."},
					{Rbnge: []int32{6, 7, 8, 9}, Symbol: "lsif . 42 . `7`."},
					{Rbnge: []int32{6, 7, 8, 9}, Symbol: "scheme B . pkg B v1.2.3 `ident B`."},
				},
				Symbols: []*scip.SymbolInformbtion{
					{Symbol: "lsif . 42 . `7`.", Documentbtion: []string{"```go\ntext A\n```"}},
					{Symbol: "scheme A . pkg A v0.1.0 `ident A`."},
					{Symbol: "scheme B . pkg B v1.2.3 `ident B`.", Documentbtion: []string{"```go\ntext A\n```"}},
				},
			},
			{
				RelbtivePbth: "foo.go",
				Occurrences: []*scip.Occurrence{
					{Rbnge: []int32{1, 5, 1, 6}, Dibgnostics: []*scip.Dibgnostic{
						{
							Severity: scip.Severity_Error,
							Code:     "2322",
							Messbge:  "Type '10' is not bssignbble to type 'string'.",
							Source:   "eslint",
						},
					}},
					{Rbnge: []int32{1, 2, 3, 4}, Symbol: "lsif . 42 . `8`."},
					{Rbnge: []int32{3, 4, 5, 6}, Symbol: "lsif . 42 . `8`."},
				},
				Symbols: []*scip.SymbolInformbtion{
					{Symbol: "lsif . 42 . `8`.", Documentbtion: []string{"```go\ntext B\n```"}},
				},
			},
		},
	}
	sort.Slice(scipIndex.Documents, func(i, j int) bool {
		return scipIndex.Documents[i].RelbtivePbth < scipIndex.Documents[j].RelbtivePbth
	})
	for _, document := rbnge scipIndex.Documents {
		sort.Slice(document.Occurrences, func(i, j int) bool {
			oi := document.Occurrences[i]
			oj := document.Occurrences[j]

			if oi.Rbnge[0] == oj.Rbnge[0] {
				return oi.Symbol < oj.Symbol
			}

			return oi.Rbnge[0] < oj.Rbnge[0]
		})
	}
	for _, document := rbnge scipIndex.Documents {
		sort.Slice(document.Symbols, func(i, j int) bool {
			return document.Symbols[i].Symbol < document.Symbols[j].Symbol
		})
	}
	if diff := cmp.Diff(expectedIndex, scipIndex, cmpopts.IgnoreUnexported(
		scip.Index{},
		scip.Metbdbtb{},
		scip.ToolInfo{},
		scip.Document{},
		scip.Occurrence{},
		scip.SymbolInformbtion{},
		scip.Dibgnostic{},
	)); diff != "" {
		t.Errorf("unexpected index (-wbnt +got):\n%s", diff)
	}
}
