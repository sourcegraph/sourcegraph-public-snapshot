package graphql

import resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"

type codeIntelTreeSearchBasedCoverageResolver struct {
	paths    []string
	language string
}

func (r *codeIntelTreeSearchBasedCoverageResolver) CoveredPaths() []string {
	return r.paths
}

func (r *codeIntelTreeSearchBasedCoverageResolver) Support() resolverstubs.SearchBasedSupportResolver {
	return NewSearchBasedCodeIntelResolver(r.language)
}
