package graphql

import (
	"context"
	"time"

	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type Resolver interface {
	// Symbols client
	GetSupportedByCtags(ctx context.Context, filepath string, repoName api.RepoName) (bool, string, error)

	// Language support
	GetLanguagesRequestedBy(ctx context.Context, userID int) (_ []string, err error)
	SetRequestLanguageSupport(ctx context.Context, userID int, language string) (err error)

	// Factory for GitBlobLSIFDataResolver
	GitBlobLSIFDataResolverFactory(ctx context.Context, repo *types.Repo, commit, path, toolName string, exactPath bool) (_ GitBlobLSIFDataResolver, err error)
}

type resolver struct {
	svc                            Service
	gitserver                      GitserverClient
	maximumIndexesPerMonikerSearch int
	hunkCacheSize                  int

	// Metrics
	operations *operations
}

func New(svc Service, gitserver GitserverClient, maxIndexSearch, hunkCacheSize int, observationContext *observation.Context) Resolver {
	return &resolver{
		svc:                            svc,
		gitserver:                      gitserver,
		operations:                     newOperations(observationContext),
		hunkCacheSize:                  hunkCacheSize,
		maximumIndexesPerMonikerSearch: maxIndexSearch,
	}
}

func (r *resolver) GetSupportedByCtags(ctx context.Context, filepath string, repoName api.RepoName) (_ bool, _ string, err error) {
	ctx, _, endObservation := r.operations.getSupportedByCtags.With(ctx, &err, observation.Args{
		LogFields: []log.Field{log.String("repoName", string(repoName))},
	})
	defer endObservation(1, observation.Args{})

	return r.svc.GetSupportedByCtags(ctx, filepath, repoName)
}

func (r *resolver) SetRequestLanguageSupport(ctx context.Context, userID int, language string) (err error) {
	ctx, _, endObservation := r.operations.setRequestLanguageSupport.With(ctx, &err, observation.Args{
		LogFields: []log.Field{log.Int("userID", userID), log.String("language", language)},
	})
	defer endObservation(1, observation.Args{})

	return r.svc.SetRequestLanguageSupport(ctx, userID, language)
}

func (r *resolver) GetLanguagesRequestedBy(ctx context.Context, userID int) (_ []string, err error) {
	ctx, _, endObservation := r.operations.getLanguagesRequestedBy.With(ctx, &err, observation.Args{
		LogFields: []log.Field{log.Int("userID", userID)},
	})
	defer endObservation(1, observation.Args{})

	return r.svc.GetLanguagesRequestedBy(ctx, userID)
}

const slowQueryResolverRequestThreshold = time.Second

func (r *resolver) GitBlobLSIFDataResolverFactory(ctx context.Context, repo *types.Repo, commit, path, toolName string, exactPath bool) (_ GitBlobLSIFDataResolver, err error) {
	ctx, _, endObservation := observeResolver(ctx, &err, r.operations.getGitBlobLSIFDataResolver, slowQueryResolverRequestThreshold, observation.Args{
		LogFields: []log.Field{
			log.Int("repositoryID", int(repo.ID)),
			log.String("commit", commit),
			log.String("path", path),
			log.Bool("exactPath", exactPath),
			log.String("indexer", toolName),
		},
	})
	defer endObservation()

	uploads, err := r.svc.GetClosestDumpsForBlob(ctx, int(repo.ID), commit, path, exactPath, toolName)
	if err != nil || len(uploads) == 0 {
		return nil, err
	}

	reqState := codenav.NewRequestState(uploads, authz.DefaultSubRepoPermsChecker, r.gitserver, repo, commit, path, r.maximumIndexesPerMonikerSearch, r.hunkCacheSize)
	gbr := NewGitBlobLSIFDataResolver(r.svc, int(repo.ID), commit, path, r.operations, reqState)

	return gbr, nil
}
