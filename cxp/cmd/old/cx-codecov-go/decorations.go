package main

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/sourcegraph/sourcegraph/cxp"
	"github.com/sourcegraph/sourcegraph/xlang/lspext"
	"github.com/sourcegraph/sourcegraph/xlang/uri"
)

// publishDecorationsForOpenDocuments publishes new decorations for all open documents. It should be
// called after a change (to configuration or the open files) that may require decorations to
// change.
func publishDecorationsForOpenDocuments(ctx context.Context, conn *jsonrpc2.Conn, clientCap *cxp.ClientCapabilities, settings extensionSettings, rootURI *uri.URI, openDocuments map[lsp.DocumentURI]struct{}) error {
	if !clientCap.Decorations.Dynamic {
		return nil
	}

	for doc := range openDocuments {
		decorations, err := getDecorations(ctx, settings, rootURI, doc)
		if err != nil {
			return errors.WithMessage(err, fmt.Sprintf("getting decorations for %s", doc))
		}
		_ = conn.Notify(ctx, "textDocument/publishDecorations", cxp.TextDocumentPublishDecorationsParams{
			TextDocument: lsp.TextDocumentIdentifier{URI: doc},
			Decorations:  decorations,
		})
	}
	return nil
}

func getDecorations(ctx context.Context, settings extensionSettings, root *uri.URI, doc lsp.DocumentURI) ([]lspext.TextDocumentDecoration, error) {
	if settings.LineDecorations != nil && !*settings.LineDecorations {
		return nil, nil // line decorations disabled
	}

	docURI, err := url.Parse(string(doc))
	if err != nil {
		return nil, err
	}
	coverage, err := getCoverageForFile(ctx, settings.Token, string(root.Repo()), root.Rev(), strings.TrimPrefix(docURI.Path, "/"))
	if err != nil {
		return nil, err
	}
	return codecovToDecorations(coverage), nil
}

func codecovToDecorations(coverageByLine map[int]lineCoverage) []lspext.TextDocumentDecoration {
	greenHue := 116
	redHue := 0
	yellowHue := 60
	saturated := 100
	hsla := func(hue int, saturation int, lightness int, alpha float64) string {
		return fmt.Sprintf("hsla(%d, %d%%, %d%%, %.2f)", hue, saturation, lightness, alpha)
	}

	decorations := []lspext.TextDocumentDecoration{}
	for line, c := range coverageByLine {
		var hue int
		switch {
		case c.skip:
			continue
		case c.isPartial():
			if c.partials == c.branches {
				// All branches were covered.
				hue = greenHue
			} else if c.partials == 0 {
				// No coverage.
				hue = redHue
			} else {
				// Some branches were not covered.
				hue = yellowHue
			}
		case c.hits > 0:
			hue = greenHue
		case c.hits == 0:
			hue = redHue
		}
		decorations = append(decorations, lspext.TextDocumentDecoration{
			// TODO: tune alpha down to ~0.125 after giving demos (higher is easier to see in demos)
			BackgroundColor: hsla(hue, saturated, 50, 0.2),
			Range: lsp.Range{
				Start: lsp.Position{Line: line - 1},
				End:   lsp.Position{Line: line - 1},
			},
			IsWholeLine: true,
		})
	}

	return decorations
}
