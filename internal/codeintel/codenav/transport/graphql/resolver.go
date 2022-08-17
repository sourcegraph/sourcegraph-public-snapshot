package graphql

import (
	"context"
	"time"

	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type Resolver interface {
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
