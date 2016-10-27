package ctags

import (
	"context"
	"strings"

	"github.com/sourcegraph/go-langserver/pkg/lsp"

	"sourcegraph.com/sourcegraph/sourcegraph/xlang/ctags/parser"
)

func handleHover(ctx context.Context, params lsp.TextDocumentPositionParams) (*lsp.Hover, error) {
	tags, err := getTags(ctx)
	if err != nil {
		return nil, err
	}

	word, wordStart, err := wordAtPosition(ctx, params)
	if err != nil {
		return nil, err
	}

	tag := compareTags(word, tags)
	if tag == nil {
		return nil, nil
	}

	start := lsp.Position{Line: params.Position.Line, Character: wordStart}
	var info string
	if tag.Signature != "" {
		info = tag.Kind + tag.Signature
	} else {
		info = strings.TrimSpace(tag.DefLinePrefix)
	}
	hoverInfo := &lsp.Hover{
		Contents: []lsp.MarkedString{
			lsp.MarkedString{
				Language: strings.ToLower(tag.Language),
				Value:    info,
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
