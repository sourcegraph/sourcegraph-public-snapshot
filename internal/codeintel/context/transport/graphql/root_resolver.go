package graphql

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/types"

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

func (r *rootResolver) GetPreciseContext(ctx context.Context, input *resolverstubs.GetPreciseContextInput) (resolverstubs.PreciseContextOutputResolver, error) {
	if err := validateGetPreciseContextInput(input); err != nil {
		return nil, err
	}

	// DEBUG
	// if input.Input.ActiveFileSelectionRange != nil {
	// 	fmt.Println("I gots start selection", input.Input.ActiveFileSelectionRange.StartLine, input.Input.ActiveFileSelectionRange.StartCharacter)
	// 	fmt.Println("I gots end selection", input.Input.ActiveFileSelectionRange.EndLine, input.Input.ActiveFileSelectionRange.EndCharacter)
	// }

	context, traceLogs, err := r.svc.GetPreciseContext(ctx, input)
	if err != nil {
		return nil, err
	}

	resolvers := make([]resolverstubs.PreciseContextResolver, 0, len(context))
	for _, c := range context {
		resolvers = append(resolvers, &preciseContextResolver{
			symbol:            &preciseSymbolReferenceResolver{c.Symbol},
			repositoryName:    c.RepositoryName,
			definitionSnippet: c.DefinitionSnippet,
			filepath:          c.Filepath,
		})
	}

	return &preciseContextOutputResolver{
		context:   resolvers,
		traceLogs: traceLogs,
	}, nil
}

type preciseContextOutputResolver struct {
	context   []resolverstubs.PreciseContextResolver
	traceLogs string
}

func (r *preciseContextOutputResolver) Context() []resolverstubs.PreciseContextResolver {
	return r.context
}

func (r *preciseContextOutputResolver) TraceLogs() string {
	return r.traceLogs
}

type preciseContextResolver struct {
	symbol            *preciseSymbolReferenceResolver
	definitionSnippet string
	repositoryName    string
	filepath          string
}

type preciseSymbolReferenceResolver struct {
	ref types.PreciseSymbolReference
}

func (r *preciseContextResolver) Symbol() resolvers.PreciseSymbolReferenceResolver { return r.symbol }
func (r *preciseContextResolver) DefinitionSnippet() string                        { return r.definitionSnippet }
func (r *preciseContextResolver) RepositoryName() string                           { return r.repositoryName }
func (r *preciseContextResolver) Filepath() string                                 { return r.filepath }
func (r *preciseContextResolver) CanonicalLocationURL() string                     { return "UNIMPLEMENTED" } // TODO
func (r *preciseSymbolReferenceResolver) ScipName() string                         { return r.ref.ScipName }
func (r *preciseSymbolReferenceResolver) ScipDescriptorSuffix() string             { return r.ref.DescriptorSuffix }
func (r *preciseSymbolReferenceResolver) FuzzyName() *string                       { return r.ref.FuzzyName }
