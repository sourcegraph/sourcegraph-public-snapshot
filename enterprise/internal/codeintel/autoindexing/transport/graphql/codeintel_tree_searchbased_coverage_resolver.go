package graphql

type GitTreeSearchBasedCoverage interface {
	CoveredPaths() []string
	Support() SearchBasedSupportResolver
}

type codeIntelTreeSearchBasedCoverageResolver struct {
	paths    []string
	language string
}

func (r *codeIntelTreeSearchBasedCoverageResolver) CoveredPaths() []string {
	return r.paths
}

func (r *codeIntelTreeSearchBasedCoverageResolver) Support() SearchBasedSupportResolver {
	return NewSearchBasedCodeIntelResolver(r.language)
}
