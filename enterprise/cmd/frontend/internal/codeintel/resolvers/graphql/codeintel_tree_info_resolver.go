package graphql

import (
	"context"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type codeIntelTreeInfoResolver struct {
	resolver  resolvers.Resolver
	commit    string
	path      string
	files     []string
	repo      *types.Repo
	errTracer *observation.ErrCollector
}

func NewCodeIntelTreeInfoResolver(
	resolver resolvers.Resolver,
	repo *types.Repo,
	commit, path string,
	files []string,
	errTracer *observation.ErrCollector,
) gql.GitTreeCodeIntelSupportResolver {
	return &codeIntelTreeInfoResolver{
		resolver:  resolver,
		repo:      repo,
		commit:    commit,
		path:      path,
		files:     files,
		errTracer: errTracer,
	}
}

func (r *codeIntelTreeInfoResolver) SearchBasedSupport(ctx context.Context) ([]gql.GitTreeSearchBasedCoverage, error) {
	langMapping := make(map[string][]string)

	for _, file := range r.files {
		ok, lang, err := r.resolver.SupportedByCtags(ctx, file, r.repo.Name)
		if err != nil {
			return nil, err
		}
		if ok {
			langMapping[lang] = append(langMapping[lang], file)
		}
	}

	resolvers := make([]gql.GitTreeSearchBasedCoverage, 0, len(langMapping))

	for lang, files := range langMapping {
		resolvers = append(resolvers, &codeIntelTreeSearchBasedCoverageResolver{
			paths:    files,
			language: lang,
		})
	}

	return resolvers, nil
}

func (r *codeIntelTreeInfoResolver) PreciseSupport(ctx context.Context) ([]gql.GitTreePreciseCoverage, error) {
	return nil, nil
}

type codeIntelTreePreciseCoverageResolver struct{}

func (r *codeIntelTreePreciseCoverageResolver) CoveredPaths() []string {
	return nil
}

func (r *codeIntelTreePreciseCoverageResolver) Support() gql.PreciseSupportResolver {
	return nil
}

func (r *codeIntelTreePreciseCoverageResolver) Confidence() string {
	return ""
}

type codeIntelTreeSearchBasedCoverageResolver struct {
	paths    []string
	language string
}

func (r *codeIntelTreeSearchBasedCoverageResolver) CoveredPaths() []string {
	return r.paths
}

func (r *codeIntelTreeSearchBasedCoverageResolver) Support() gql.SearchBasedSupportResolver {
	return NewSearchBasedCodeIntelResolver(&r.language)
}
