package resolvers

import (
	"github.com/sourcegraph/go-lsp"
	bundles "github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/client"
)

func convertRange(r bundles.Range) lsp.Range {
	return lsp.Range{
		Start: lsp.Position{Line: r.Start.Line, Character: r.Start.Character},
		End:   lsp.Position{Line: r.End.Line, Character: r.End.Character},
	}
}
