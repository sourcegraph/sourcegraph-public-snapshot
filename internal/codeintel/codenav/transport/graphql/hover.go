package graphql

import (
	"github.com/sourcegraph/go-lsp"
)

type HoverResolver interface {
	Markdown() Markdown
	Range() RangeResolver
}

type hoverResolver struct {
	text     string
	lspRange lsp.Range
}

func NewHoverResolver(text string, lspRange lsp.Range) HoverResolver {
	return &hoverResolver{
		text:     text,
		lspRange: lspRange,
	}
}

func (r *hoverResolver) Markdown() Markdown   { return Markdown(r.text) }
func (r *hoverResolver) Range() RangeResolver { return NewRangeResolver(r.lspRange) }
