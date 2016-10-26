package ctags

import (
	"testing"
	"time"

	"context"

	"github.com/sourcegraph/ctxvfs"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
)

var rubyFile = []byte(`# a comment
def foo(name)
	"Hello, " + name
end
 foo("milton")
`)
var defLoc = lsp.Location{
	URI: "file:////hello.rb",
	Range: lsp.Range{
		Start: lsp.Position{Line: 1, Character: 4},
		End:   lsp.Position{Line: 1, Character: 7},
	},
}

func setupHandler() (*Handler, context.Context) {
	ctx := context.Background()
	fs := ctxvfs.Map(map[string][]byte{"hello.rb": rubyFile})
	h := Handler{}
	h.fs = fs
	h.mode = "ruby"
	return &h, ctx
}

func TestSymbols(t *testing.T) {
	h, ctx := setupHandler()

	params := lsp.WorkspaceSymbolParams{
		Query: "foo",
		Limit: 7,
	}
	syms, err := h.handleSymbol(ctx, params)
	if err != nil {
		t.Fatal(err)
	}

	assert := func(b bool) {
		if !b {
			t.Errorf("expected true, but got false")
		}
	}

	assert(len(syms) == 1)
	sym := syms[0]
	assert(sym.Name == "foo")
	assert(sym.Kind == lsp.SKMethod)
	assert(sym.Location == defLoc)
}

func TestDefinition(t *testing.T) {
	h, ctx := setupHandler()
	type test struct {
		col    int
		result []lsp.Location
	}
	tests := []test{
		test{-100, []lsp.Location{}},
		test{4, []lsp.Location{defLoc}},
		test{8, []lsp.Location{}},
	}
	time.Sleep(100 * time.Millisecond)
	for _, test := range tests {
		locs := defAtPoint(t, test.col, h, ctx)
		if len(locs) > 1 {
			t.Error("too many results")
		}
		if len(test.result) != len(locs) {
			t.Errorf("expected to get %d locations, got %d", len(test.result), len(locs))
			t.FailNow()
		}
		if len(test.result) == 0 && len(locs) == 0 {
			continue
		}
		if locs[0] != test.result[0] {
			t.Errorf("expected to get location %v, got %v", test.result, locs[0])
		}
	}
}

func defAtPoint(t *testing.T, col int, h *Handler, ctx context.Context) []lsp.Location {
	params := lsp.TextDocumentPositionParams{
		Position: lsp.Position{Line: 4, Character: col},
		TextDocument: lsp.TextDocumentIdentifier{
			URI: "hello.rb",
		},
	}
	locs, _ := h.handleDefinition(ctx, params)
	return locs
}

func TestWordAtPoint(t *testing.T) {
	line := "foo bar baz quix"
	assert := func(b bool) {
		if !b {
			t.Errorf("expected true but got false")
		}
	}
	var word = func(line string, col int) string {
		word, _ := wordAtPoint(line, col)
		return word
	}
	assert(word(line, 4) == "")
	assert(word(line, 5) == "bar")
	assert(word(line, 7) == "bar")
	assert(word(line, 8) == "")
}

func TestReferences(t *testing.T) {
	h, ctx := setupHandler()
	params := lsp.ReferenceParams{
		TextDocumentPositionParams: lsp.TextDocumentPositionParams{
			Position: lsp.Position{
				Line:      1,
				Character: 5,
			},
			TextDocument: lsp.TextDocumentIdentifier{
				URI: "file:///hello.rb",
			},
		},
	}
	refs, err := h.handleReferences(ctx, params)
	if err != nil {
		t.Fatal(err)
	}
	expectedRefs := []lsp.Location{
		lsp.Location{
			URI: "file:///hello.rb",
			Range: lsp.Range{
				Start: lsp.Position{
					Line:      1,
					Character: 4,
				},
				End: lsp.Position{
					Line:      1,
					Character: 7,
				},
			},
		},
		lsp.Location{
			URI: "file:///hello.rb",
			Range: lsp.Range{
				Start: lsp.Position{
					Line:      4,
					Character: 1,
				},
				End: lsp.Position{
					Line:      4,
					Character: 4,
				},
			},
		},
	}
	for i, e := range expectedRefs {
		if refs[i] != e {
			t.Errorf("expected %v, but got %v", e, refs[i])
		}
	}
}

func TestHover(t *testing.T) {
	h, ctx := setupHandler()
	params := lsp.TextDocumentPositionParams{
		Position: lsp.Position{
			Line:      1,
			Character: 5,
		},
		TextDocument: lsp.TextDocumentIdentifier{
			URI: "file:///hello.rb",
		},
	}
	hover, err := h.handleHover(ctx, params)
	if err != nil {
		t.Fatal(err)
	}
	assert := func(b bool) {
		if !b {
			t.Errorf("expected true but got false")
		}
	}

	assert(len(hover.Contents) == 1)
	assert(hover.Contents[0] == lsp.MarkedString{Language: "ruby", Value: "def foo(name)"})
	assert(hover.Range == defLoc.Range)
}
