package graphql

import (
	"context"
	"strings"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
)

type preciseSupportInferenceConfidence string

const (
	languageSupport           preciseSupportInferenceConfidence = "LANGUAGE_SUPPORTED"
	projectStructureSupported preciseSupportInferenceConfidence = "PROJECT_STRUCTURE_SUPPORTED"
	indexJobInfered           preciseSupportInferenceConfidence = "INDEX_JOB_INFERED"
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

func (r *codeIntelTreeInfoResolver) SearchBasedSupport(ctx context.Context) (*[]gql.GitTreeSearchBasedCoverage, error) {
	langMapping := make(map[string][]string)
	codeNavResolver := r.resolver.CodeNavResolver()
	for _, file := range r.files {
		ok, lang, err := codeNavResolver.GetSupportedByCtags(ctx, file, r.repo.Name)
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

	return &resolvers, nil
}

func (r *codeIntelTreeInfoResolver) PreciseSupport(ctx context.Context) (*[]gql.GitTreePreciseCoverage, error) {
	autoIndexingResolver := r.resolver.AutoIndexingResolver()
	configurations, ok, err := autoIndexingResolver.InferedIndexConfiguration(ctx, int(r.repo.ID), r.commit)
	if err != nil {
		return nil, err
	}

	var resolvers []gql.GitTreePreciseCoverage

	if ok {
		for _, job := range configurations.IndexJobs {
			if job.Root == r.path {
				resolvers = append(resolvers, &codeIntelTreePreciseCoverageResolver{
					confidence: indexJobInfered,
					// drop the tag if it exists
					indexer: imageToIndexer[strings.Split(job.Indexer, ":")[0]],
				})
			}
		}
	}

	hints, err := autoIndexingResolver.InferedIndexConfigurationHints(ctx, int(r.repo.ID), r.commit)
	if err != nil {
		return nil, err
	}

	for _, hint := range hints {
		if hint.Root == r.path {
			var confidence preciseSupportInferenceConfidence
			switch hint.HintConfidence {
			case config.HintConfidenceLanguageSupport:
				confidence = languageSupport
			case config.HintConfidenceProjectStructureSupported:
				confidence = projectStructureSupported
			default:
				continue
			}
			resolvers = append(resolvers, &codeIntelTreePreciseCoverageResolver{
				confidence: confidence,
				// expected that job hints don't include a tag in the indexer name
				indexer: imageToIndexer[hint.Indexer],
			})
		}
	}

	return &resolvers, nil
}

type codeIntelTreePreciseCoverageResolver struct {
	confidence preciseSupportInferenceConfidence
	indexer    gql.CodeIntelIndexerResolver
}

func (r *codeIntelTreePreciseCoverageResolver) Support() gql.PreciseSupportResolver {
	return NewPreciseCodeIntelSupportResolverFromIndexers([]gql.CodeIntelIndexerResolver{r.indexer})
}

func (r *codeIntelTreePreciseCoverageResolver) Confidence() string {
	return string(r.confidence)
}

type codeIntelTreeSearchBasedCoverageResolver struct {
	paths    []string
	language string
}

func (r *codeIntelTreeSearchBasedCoverageResolver) CoveredPaths() []string {
	return r.paths
}

func (r *codeIntelTreeSearchBasedCoverageResolver) Support() gql.SearchBasedSupportResolver {
	return NewSearchBasedCodeIntelResolver(r.language)
}
