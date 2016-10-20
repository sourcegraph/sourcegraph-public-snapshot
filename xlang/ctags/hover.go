package ctags

import (
	"fmt"

	"sourcegraph.com/sourcegraph/sourcegraph/xlang/ctags/parser"

	"github.com/sourcegraph/sourcegraph-go/pkg/lsp"
	"golang.org/x/net/context"
)

var ErrNotFound = fmt.Errorf("definition not found")

func (h *Handler) handleHover(ctx context.Context, params lsp.TextDocumentPositionParams) (*lsp.Hover, error) {
	tags, err := h.getTags(ctx)
	if err != nil {
		return nil, err
	}

	word, wordStart, err := wordAtPosition(ctx, h.fs, params)
	if err != nil {
		return nil, err
	}

	tag := compareTags(word, tags)
	if tag == nil {
		return nil, ErrNotFound
	}

	start := lsp.Position{Line: params.Position.Line, Character: wordStart}
	hoverInfo := &lsp.Hover{
		Contents: []lsp.MarkedString{
			lsp.MarkedString{
				Language: "",
				Value:    "Type: " + tag.Kind,
			},
		},
		Range: lsp.Range{
			Start: start,
			End:   lsp.Position{Line: start.Line, Character: start.Character + len(word)},
		},
	}
	return hoverInfo, nil
}

func compareTags(word string, tags []parser.Tag) *parser.Tag {
	for _, tag := range tags {
		if tag.Name == word {
			return &tag
		}
	}
	return nil
}
