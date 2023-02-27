package graphql

import (
	"github.com/sourcegraph/go-lsp"

	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
)

type hoverResolver struct {
	text     string
	lspRange lsp.Range
}

func NewHoverResolver(text string, lspRange lsp.Range) resolverstubs.HoverResolver {
	return &hoverResolver{
		text:     text,
		lspRange: lspRange,
	}
}

func (r *hoverResolver) Markdown() resolverstubs.Markdown   { return resolverstubs.Markdown(r.text) }
func (r *hoverResolver) Range() resolverstubs.RangeResolver { return NewRangeResolver(r.lspRange) }
