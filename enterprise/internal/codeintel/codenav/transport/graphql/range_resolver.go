package graphql

import (
	"github.com/sourcegraph/go-lsp"

	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
)

type rangeResolver struct{ lspRange lsp.Range }

func NewRangeResolver(lspRange lsp.Range) resolverstubs.RangeResolver {
	return &rangeResolver{
		lspRange: lspRange,
	}
}

func (r *rangeResolver) Start() resolverstubs.PositionResolver { return r.start() }
func (r *rangeResolver) End() resolverstubs.PositionResolver   { return r.end() }

func (r *rangeResolver) start() *positionResolver { return &positionResolver{r.lspRange.Start} }
func (r *rangeResolver) end() *positionResolver   { return &positionResolver{r.lspRange.End} }

func (r *rangeResolver) urlFragment() string {
	if r.lspRange.Start == r.lspRange.End {
		return r.start().urlFragment(false)
	}
	hasCharacter := r.lspRange.Start.Character != 0 || r.lspRange.End.Character != 0
	return r.start().urlFragment(hasCharacter) + "-" + r.end().urlFragment(hasCharacter)
}
