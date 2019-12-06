package resolvers

import (
	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

type hoverResolver struct {
	text     string
	lspRange lsp.Range
}

var _ graphqlbackend.HoverResolver = &hoverResolver{}

func (r *hoverResolver) Markdown() graphqlbackend.MarkdownResolver {
	return graphqlbackend.NewMarkdownResolver(r.text)
}

func (r *hoverResolver) Range() graphqlbackend.RangeResolver {
	return graphqlbackend.NewRangeResolver(r.lspRange)
}
