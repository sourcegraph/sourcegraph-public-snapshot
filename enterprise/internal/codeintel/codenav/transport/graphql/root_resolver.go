package graphql

import (
	"context"
	"strings"

	"github.com/opentracing/opentracing-go/log"

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

func NewRootResolver(observationCtx *observation.Context, svc CodeNavService, autoindexingSvc AutoIndexingService, uploadSvc UploadsService, gitserverClient gitserver.Client, siteAdminChecker sharedresolvers.SiteAdminChecker, repoStore database.RepoStore, locationResolverFactory *sharedresolvers.CachedLocationResolverFactory, prefetcherFactory *sharedresolvers.PrefetcherFactory, maxIndexSearch, hunkCacheSize int) (resolverstubs.CodeNavServiceResolver, error) {
	hunkCache, err := codenav.NewHunkCache(hunkCacheSize)
	if err != nil {
		return nil, err
	}

	return &rootResolver{
		svc:                            svc,
		autoindexingSvc:                autoindexingSvc,
		uploadSvc:                      uploadSvc,
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
	ctx, errTracer, endObservation := r.operations.gitBlobLsifData.WithErrors(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repoID", int(args.Repo.ID)),
		log.String("commit", string(args.Commit)),
		log.String("path", args.Path),
		log.Bool("exactPath", args.ExactPath),
		log.String("toolName", args.ToolName),
	}})
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

	reqState := codenav.NewRequestState(
		uploads,
		r.repoStore,
		authz.DefaultSubRepoPermsChecker,
		r.gitserverClient,
		args.Repo,
		string(args.Commit),
		args.Path,
		r.maximumIndexesPerMonikerSearch,
		r.hunkCache,
	)

	return newGitBlobLSIFDataResolver(
		r.svc,
		r.uploadSvc,
		r.gitserverClient,
		r.siteAdminChecker,
		r.repoStore,
		r.prefetcherFactory.Create(),
		r.locationResolverFactory.Create(),
		reqState,
		errTracer,
		r.operations,
	), nil
}

// gitBlobLSIFDataResolver is the main interface to bundle-related operations exposed to the GraphQL API. This
// resolver concerns itself with GraphQL/API-specific behaviors (auth, validation, marshaling, etc.).
// All code intel-specific behavior is delegated to the underlying resolver instance, which is defined
// in the parent package.
type gitBlobLSIFDataResolver struct {
	codeNavSvc       CodeNavService
	uploadsSvc       UploadsService
	gitserverClient  gitserver.Client
	siteAdminChecker sharedresolvers.SiteAdminChecker
	repoStore        database.RepoStore
	prefetcher       *sharedresolvers.Prefetcher

	requestState     codenav.RequestState
	locationResolver *sharedresolvers.CachedLocationResolver
	errTracer        *observation.ErrCollector

	operations *operations
}

// NewQueryResolver creates a new QueryResolver with the given resolver that defines all code intel-specific
// behavior. A cached location resolver instance is also given to the query resolver, which should be used
// to resolve all location-related values.
func newGitBlobLSIFDataResolver(
	codeNavSvc CodeNavService,
	uploadsSvc UploadsService,
	gitserverClient gitserver.Client,
	siteAdminChecker sharedresolvers.SiteAdminChecker,
	repoStore database.RepoStore,
	prefetcher *sharedresolvers.Prefetcher,
	locationResolver *sharedresolvers.CachedLocationResolver,
	requestState codenav.RequestState,
	errTracer *observation.ErrCollector,
	operations *operations,
) resolverstubs.GitBlobLSIFDataResolver {
	return &gitBlobLSIFDataResolver{
		codeNavSvc:       codeNavSvc,
		uploadsSvc:       uploadsSvc,
		gitserverClient:  gitserverClient,
		siteAdminChecker: siteAdminChecker,
		repoStore:        repoStore,
		prefetcher:       prefetcher,
		requestState:     requestState,
		locationResolver: locationResolver,
		errTracer:        errTracer,
		operations:       operations,
	}
}

func (r *gitBlobLSIFDataResolver) ToGitTreeLSIFData() (resolverstubs.GitTreeLSIFDataResolver, bool) {
	return r, true
}

func (r *gitBlobLSIFDataResolver) ToGitBlobLSIFData() (resolverstubs.GitBlobLSIFDataResolver, bool) {
	return r, true
}
