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

func (r *rootResolver) GetPreciseContext(ctx context.Context, input *resolverstubs.GetPreciseContextInput) (resolverstubs.PreciseContextOutputResolver, error) {
	if err := validateGetPreciseContextInput(input); err != nil {
		return nil, err
	}

	// DEBUG
	if input.Input.ActiveFileSelectionRange != nil {
		fmt.Println("I gots start selection", input.Input.ActiveFileSelectionRange.StartLine, input.Input.ActiveFileSelectionRange.StartCharacter)
		fmt.Println("I gots end selection", input.Input.ActiveFileSelectionRange.EndLine, input.Input.ActiveFileSelectionRange.EndCharacter)
	}

	context, err := r.svc.GetPreciseContext(ctx, input)
	if err != nil {
		return nil, err
	}

	resolvers := make([]resolverstubs.PreciseContextResolver, 0, len(context))
	for _, c := range context {
		resolvers = append(resolvers, &preciseContextResolver{
			scipSymbolName:        c.ScipSymbolName,
			fuzzySymbolName:       c.FuzzySymbolName,
			scipDescriptorSuffix:  c.ScipDescriptorSuffix,
			fuzzyDescriptorSuffix: c.FuzzyDescriptorSuffix,
			repositoryName:        c.RepositoryName,
			text:                  c.Text,
			filepath:              c.FilePath,
		})
	}
	return &preciseContextOutputResolver{context: resolvers}, nil
}

type preciseContextOutputResolver struct {
	context []resolverstubs.PreciseContextResolver
}

func (r *preciseContextOutputResolver) Context() []resolverstubs.PreciseContextResolver {
	return r.context
}

type preciseContextResolver struct {
	scipSymbolName        string
	fuzzySymbolName       string
	repositoryName        string
	scipDescriptorSuffix  string
	fuzzyDescriptorSuffix string
	text                  string
	filepath              string
}

func (r *preciseContextResolver) ScipSymbolName() string  { return r.scipSymbolName }
func (r *preciseContextResolver) FuzzySymbolName() string { return r.fuzzySymbolName }
func (r *preciseContextResolver) ScipDescriptorSuffix() string {
	return r.scipDescriptorSuffix
}
func (r *preciseContextResolver) FuzzyDescriptorSuffix() string {
	return r.fuzzyDescriptorSuffix
}
func (r *preciseContextResolver) RepositoryName() string { return r.repositoryName }
func (r *preciseContextResolver) Text() string           { return r.text }
func (r *preciseContextResolver) FilePath() string       { return r.filepath }
