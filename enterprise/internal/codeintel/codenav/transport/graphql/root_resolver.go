package graphql

import (
	"context"
	"strings"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/codenav"
	sharedresolvers "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type rootResolver struct {
	svc                            CodeNavService
	autoindexingSvc                AutoIndexingService
	uploadSvc                      UploadsService
	policiesSvc                    PolicyService
	gitserverClient                gitserver.Client
	siteAdminChecker               sharedresolvers.SiteAdminChecker
	repoStore                      database.RepoStore
	prefetcherFactory              *sharedresolvers.PrefetcherFactory
	locationResolverFactory        *sharedresolvers.CachedLocationResolverFactory
	maximumIndexesPerMonikerSearch int
	hunkCache                      codenav.HunkCache

	// Metrics
	operations *operations
}

func NewRootResolver(observationCtx *observation.Context, svc CodeNavService, autoindexingSvc AutoIndexingService, uploadSvc UploadsService, policiesSvc PolicyService, gitserverClient gitserver.Client, siteAdminChecker sharedresolvers.SiteAdminChecker, repoStore database.RepoStore, locationResolverFactory *sharedresolvers.CachedLocationResolverFactory, prefetcherFactory *sharedresolvers.PrefetcherFactory, maxIndexSearch, hunkCacheSize int) (resolverstubs.CodeNavServiceResolver, error) {
	hunkCache, err := codenav.NewHunkCache(hunkCacheSize)
	if err != nil {
		return nil, err
	}

	return &rootResolver{
		svc:                            svc,
		autoindexingSvc:                autoindexingSvc,
		uploadSvc:                      uploadSvc,
		policiesSvc:                    policiesSvc,
		gitserverClient:                gitserverClient,
		siteAdminChecker:               siteAdminChecker,
		repoStore:                      repoStore,
		prefetcherFactory:              prefetcherFactory,
		locationResolverFactory:        locationResolverFactory,
		operations:                     newOperations(observationCtx),
		hunkCache:                      hunkCache,
		maximumIndexesPerMonikerSearch: maxIndexSearch,
	}, nil
}

// 🚨 SECURITY: dbstore layer handles authz for query resolution
func (r *rootResolver) GitBlobLSIFData(ctx context.Context, args *resolverstubs.GitBlobLSIFDataArgs) (_ resolverstubs.GitBlobLSIFDataResolver, err error) {
	ctx, errTracer, endObservation := r.operations.gitBlobLsifData.WithErrors(ctx, &err, observation.Args{})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	uploads, err := r.svc.GetClosestDumpsForBlob(ctx, int(args.Repo.ID), string(args.Commit), args.Path, args.ExactPath, args.ToolName)
	if err != nil || len(uploads) == 0 {
		return nil, err
	}

	if len(uploads) == 0 {
		// If we're on sourcegraph.com and it's a rust package repo, index it on-demand
		if envvar.SourcegraphDotComMode() && strings.HasPrefix(string(args.Repo.Name), "crates/") {
			err = r.autoindexingSvc.QueueRepoRev(ctx, int(args.Repo.ID), string(args.Commit))
		}

		return nil, err
	}

	reqState := codenav.NewRequestState(uploads, r.repoStore, authz.DefaultSubRepoPermsChecker, r.gitserverClient, args.Repo, string(args.Commit), args.Path, r.maximumIndexesPerMonikerSearch, r.hunkCache)

	return NewGitBlobLSIFDataResolver(r.svc, r.uploadSvc, r.policiesSvc, r.gitserverClient, r.siteAdminChecker, r.repoStore, r.prefetcherFactory.Create(), r.locationResolverFactory.Create(), reqState, errTracer, r.operations), nil
}
