package graphql

import (
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
)

type codeIntelTreePreciseCoverageResolver struct {
	confidence preciseSupportInferenceConfidence
	indexer    types.CodeIntelIndexerResolver
}

func (r *codeIntelTreePreciseCoverageResolver) Support() resolverstubs.PreciseSupportResolver {
	return NewPreciseCodeIntelSupportResolverFromIndexers([]resolverstubs.CodeIntelIndexerResolver{r.indexer})
}

func (r *codeIntelTreePreciseCoverageResolver) Confidence() string {
	return string(r.confidence)
}
