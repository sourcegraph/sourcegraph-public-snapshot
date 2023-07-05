package graphql

import (
	"context"
	"fmt"

	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type rootResolver struct {
	svc ContextService
}

func NewRootResolver(observationCtx *observation.Context, svc ContextService) resolverstubs.ContextServiceResolver {
	return &rootResolver{
		svc: svc,
	}
}

func (r *rootResolver) GetPreciseContext(ctx context.Context, input *resolverstubs.GetPreciseContextInput) (resolverstubs.PreciseContextResolver, error) {
	fmt.Println("GetSCIPSymbolArgs: ", input)
	p, err := r.svc.GetPreciseContext(ctx, input)
	if err != nil {
		return nil, err
	}

	resolvers := make([]resolverstubs.PreciseContextDataResolver, 0, len(p))
	for _, c := range p {
		resolvers = append(resolvers, &preciseDataResolver{
			symbol:            c.SymbolName,
			syntectDescriptor: c.SyntectDescriptor,
			repository:        c.Repository,
			symbolRole:        c.SymbolRole,
			confidence:        c.Confidence,
			text:              c.Text,
			filepath:          c.FilePath,
		})
	}
	return &preciseContextResolver{context: resolvers}, nil
}

type preciseContextResolver struct {
	context []resolverstubs.PreciseContextDataResolver
}

func (r *preciseContextResolver) Context() []resolverstubs.PreciseContextDataResolver {
	return r.context
}

type preciseDataResolver struct {
	symbol            string
	syntectDescriptor string
	repository        string
	symbolRole        int32
	confidence        string
	text              string
	filepath          string
}

func (r *preciseDataResolver) Symbol() string            { return r.symbol }
func (r *preciseDataResolver) SyntectDescriptor() string { return r.syntectDescriptor }
func (r *preciseDataResolver) Repository() string        { return r.repository }
func (r *preciseDataResolver) SymbolRole() int32         { return r.symbolRole }
func (r *preciseDataResolver) Confidence() string        { return r.confidence }
func (r *preciseDataResolver) Text() string              { return r.text }
func (r *preciseDataResolver) FilePath() string          { return r.filepath }
