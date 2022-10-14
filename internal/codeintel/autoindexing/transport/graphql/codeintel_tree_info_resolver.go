package graphql

import (
	"context"
	"strings"

	codeinteltypes "github.com/sourcegraph/sourcegraph/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
)

type GitTreeCodeIntelSupportResolver interface {
	SearchBasedSupport(context.Context) (*[]GitTreeSearchBasedCoverage, error)
	PreciseSupport(context.Context) (*[]GitTreePreciseCoverage, error)
}

type preciseSupportInferenceConfidence string

const (
	languageSupport           preciseSupportInferenceConfidence = "LANGUAGE_SUPPORTED"
	projectStructureSupported preciseSupportInferenceConfidence = "PROJECT_STRUCTURE_SUPPORTED"
	indexJobInfered           preciseSupportInferenceConfidence = "INDEX_JOB_INFERED"
)

type codeIntelTreeInfoResolver struct {
	autoindexSvc AutoIndexingService
	commit       string
	path         string
	files        []string
	repo         *types.Repo
	errTracer    *observation.ErrCollector
}

func NewCodeIntelTreeInfoResolver(autoindexSvc AutoIndexingService, repo *types.Repo, commit, path string, files []string, errTracer *observation.ErrCollector) GitTreeCodeIntelSupportResolver {
	return &codeIntelTreeInfoResolver{
		autoindexSvc: autoindexSvc,
		repo:         repo,
		commit:       commit,
		path:         path,
		files:        files,
		errTracer:    errTracer,
	}
}

func (r *codeIntelTreeInfoResolver) SearchBasedSupport(ctx context.Context) (*[]GitTreeSearchBasedCoverage, error) {
	langMapping := make(map[string][]string)
	for _, file := range r.files {
		ok, lang, err := r.autoindexSvc.GetSupportedByCtags(ctx, file, r.repo.Name)
		if err != nil {
			return nil, err
		}
		if ok {
			langMapping[lang] = append(langMapping[lang], file)
		}
	}

	resolvers := make([]GitTreeSearchBasedCoverage, 0, len(langMapping))

	for lang, files := range langMapping {
		resolvers = append(resolvers, &codeIntelTreeSearchBasedCoverageResolver{
			paths:    files,
			language: lang,
		})
	}

	return &resolvers, nil
}

func (r *codeIntelTreeInfoResolver) PreciseSupport(ctx context.Context) (*[]GitTreePreciseCoverage, error) {
	configurations, _, err := r.autoindexSvc.InferIndexConfiguration(ctx, int(r.repo.ID), r.commit, true)
	if err != nil {
		return nil, err
	}

	var resolvers []GitTreePreciseCoverage

	if configurations != nil {
		for _, job := range configurations.IndexJobs {
			if job.Root == r.path {
				resolvers = append(resolvers, &codeIntelTreePreciseCoverageResolver{
					confidence: indexJobInfered,
					// drop the tag if it exists
					indexer: codeinteltypes.NewCodeIntelIndexerResolverFrom(codeinteltypes.ImageToIndexer[strings.Split(job.Indexer, ":")[0]]),
				})
			}
		}
	}

	_, hints, err := r.autoindexSvc.InferIndexConfiguration(ctx, int(r.repo.ID), r.commit, true)
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
				indexer: codeinteltypes.NewCodeIntelIndexerResolverFrom(codeinteltypes.ImageToIndexer[hint.Indexer]),
			})
		}
	}

	return &resolvers, nil
}
