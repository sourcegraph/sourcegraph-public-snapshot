package graphql

import (
	"context"

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
	if err := validateGetPreciseContextInput(input); err != nil {
		return nil, err
	}

	context, err := r.svc.GetPreciseContext(ctx, input)
	if err != nil {
		return nil, err
	}

	resolvers := make([]resolverstubs.PreciseContextOutputResolver, 0, len(context))
	for _, c := range context {
		resolvers = append(resolvers, &preciseDataResolver{
			scipSymbolName:  c.ScipSymbolName,
			fuzzySymbolName: c.FuzzySymbolName,
			repositoryName:  c.RepositoryName,
			symbolRole:      c.SymbolRole,
			confidence:      c.Confidence,
			text:            c.Text,
			filepath:        c.FilePath,
		})
	}
	return &preciseContextResolver{context: resolvers}, nil
}

type preciseContextResolver struct {
	context []resolverstubs.PreciseContextOutputResolver
}

func (r *preciseContextResolver) Context() []resolverstubs.PreciseContextOutputResolver {
	return r.context
}

type preciseDataResolver struct {
	scipSymbolName  string
	fuzzySymbolName string
	repositoryName  string
	symbolRole      int32
	confidence      string
	text            string
	filepath        string
}

func (r *preciseDataResolver) ScipSymbolName() string  { return r.scipSymbolName }
func (r *preciseDataResolver) FuzzySymbolName() string { return r.fuzzySymbolName }
func (r *preciseDataResolver) RepositoryName() string  { return r.repositoryName }
func (r *preciseDataResolver) SymbolRole() int32       { return r.symbolRole }
func (r *preciseDataResolver) Confidence() string      { return r.confidence }
func (r *preciseDataResolver) Text() string            { return r.text }
func (r *preciseDataResolver) FilePath() string        { return r.filepath }
