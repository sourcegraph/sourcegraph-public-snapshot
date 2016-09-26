package golang

import (
	"reflect"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/lsp"
)

func TestParseLoaderError(t *testing.T) {
	tests := []struct {
		errStr       string
		wantFilename string
		wantDiags    *lsp.Diagnostic
	}{
		{
			"/a/b/c:1:2: foo\nbar",
			"/a/b/c",
			&lsp.Diagnostic{
				Range: lsp.Range{
					// LSP is 0-indexed, so use line 0 char 1 (not 1 and 2)
					Start: lsp.Position{Line: 0, Character: 1},
					End:   lsp.Position{Line: 0, Character: 1},
				},
				Severity: lsp.Error,
				Source:   "go",
				Message:  "foo\nbar",
			},
		},
	}
	for _, test := range tests {
		filename, diag, err := parseLoaderError(test.errStr)
		if err != nil {
			t.Errorf("%q: %s", test.errStr, err)
			continue
		}
		if filename != test.wantFilename {
			t.Errorf("%q: got filename %q, want %q", test.errStr, filename, test.wantFilename)
		}
		if !reflect.DeepEqual(diag, test.wantDiags) {
			t.Errorf("%q: got diagnostic %+v, want %+v", test.errStr, diag, test.wantDiags)
		}
	}
}
