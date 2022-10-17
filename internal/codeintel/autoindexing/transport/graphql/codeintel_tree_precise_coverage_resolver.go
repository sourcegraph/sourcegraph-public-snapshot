package graphql

import (
	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/types"
)

type GitTreePreciseCoverage interface {
	Support() PreciseSupportResolver
	Confidence() string
}

type codeIntelTreePreciseCoverageResolver struct {
	confidence preciseSupportInferenceConfidence
	indexer    types.CodeIntelIndexerResolver
}

func (r *codeIntelTreePreciseCoverageResolver) Support() PreciseSupportResolver {
	return NewPreciseCodeIntelSupportResolverFromIndexers([]types.CodeIntelIndexerResolver{r.indexer})
}

func (r *codeIntelTreePreciseCoverageResolver) Confidence() string {
	return string(r.confidence)
}
