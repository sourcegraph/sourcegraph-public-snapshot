package codeintel

import (
	"context"
	"net/http"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel"
	autoindexinggraphql "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing/transport/graphql"
	codenavgraphql "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/codenav/transport/graphql"
	policiesgraphql "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/policies/transport/graphql"
	sentinelgraphql "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/sentinel/transport/graphql"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/lsifuploadstore"
	sharedresolvers "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/resolvers"
	uploadgraphql "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/transport/graphql"
	uploadshttp "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/transport/http"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/cloneurls"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func LoadConfig() {
	ConfigInst.Load()
}

func Init(
	ctx context.Context,
	observationCtx *observation.Context,
	db database.DB,
	codeIntelServices codeintel.Services,
	conf conftypes.UnifiedWatchable,
	enterpriseServices *enterprise.Services,
) error {
	if err := ConfigInst.Validate(); err != nil {
		return err
	}

	uploadStore, err := lsifuploadstore.New(context.Background(), observationCtx, ConfigInst.LSIFUploadStoreConfig)
	if err != nil {
		return err
	}

	newUploadHandler := func(withCodeHostAuth bool) http.Handler {
		return uploadshttp.GetHandler(codeIntelServices.UploadsService, db, codeIntelServices.GitserverClient, uploadStore, withCodeHostAuth)
	}

	cloneURLToRepoName := func(ctx context.Context, submoduleURL string) (api.RepoName, error) {
		return cloneurls.RepoSourceCloneURLToRepoName(ctx, db, submoduleURL)
	}
	repoStore := db.Repos()
	siteAdminChecker := sharedresolvers.NewSiteAdminChecker(db)
	locationResolverFactory := sharedresolvers.NewCachedLocationResolverFactory(cloneURLToRepoName, repoStore, codeIntelServices.GitserverClient)
	prefetcherFactory := sharedresolvers.NewPrefetcherFactory(codeIntelServices.UploadsService)

	autoindexingRootResolver := autoindexinggraphql.NewRootResolver(
		scopedContext("autoindexing"),
		codeIntelServices.AutoIndexingService,
		codeIntelServices.UploadsService,
		codeIntelServices.PoliciesService,
		codeIntelServices.GitserverClient,
		siteAdminChecker,
		repoStore,
		prefetcherFactory,
		locationResolverFactory,
	)

	codenavRootResolver, err := codenavgraphql.NewRootResolver(
		scopedContext("codenav"),
		codeIntelServices.CodenavService,
		codeIntelServices.AutoIndexingService,
		codeIntelServices.UploadsService,
		codeIntelServices.PoliciesService,
		codeIntelServices.GitserverClient,
		siteAdminChecker,
		repoStore,
		locationResolverFactory,
		prefetcherFactory,
		ConfigInst.MaximumIndexesPerMonikerSearch,
		ConfigInst.HunkCacheSize,
	)
	if err != nil {
		return err
	}

	policyRootResolver := policiesgraphql.NewRootResolver(
		scopedContext("policies"),
		codeIntelServices.PoliciesService,
		repoStore,
		siteAdminChecker,
	)

	uploadRootResolver := uploadgraphql.NewRootResolver(
		scopedContext("upload"),
		codeIntelServices.UploadsService,
		codeIntelServices.AutoIndexingService,
		codeIntelServices.PoliciesService,
		codeIntelServices.GitserverClient,
		siteAdminChecker,
		repoStore,
		prefetcherFactory,
		locationResolverFactory,
	)

	sentinelRootResolver := sentinelgraphql.NewRootResolver(
		scopedContext("sentinel"),
		codeIntelServices.SentinelService,
		codeIntelServices.UploadsService,
		codeIntelServices.PoliciesService,
		codeIntelServices.GitserverClient,
		siteAdminChecker,
		repoStore,
		prefetcherFactory,
		locationResolverFactory,
	)

	enterpriseServices.CodeIntelResolver = graphqlbackend.NewCodeIntelResolver(resolvers.NewCodeIntelResolver(
		autoindexingRootResolver,
		codenavRootResolver,
		policyRootResolver,
		uploadRootResolver,
		sentinelRootResolver,
	))
	enterpriseServices.NewCodeIntelUploadHandler = newUploadHandler
	enterpriseServices.RankingService = codeIntelServices.RankingService
	return nil
}

func scopedContext(name string) *observation.Context {
	return observation.NewContext(log.Scoped(name+".transport.graphql", "codeintel "+name+" graphql transport"))
}
