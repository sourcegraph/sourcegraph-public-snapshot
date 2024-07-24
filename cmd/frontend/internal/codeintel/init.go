package codeintel

import (
	"context"
	"net/http"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/codeintel"
	autoindexinggraphql "github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/transport/graphql"
	codenavgraphql "github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/transport/graphql"
	policiesgraphql "github.com/sourcegraph/sourcegraph/internal/codeintel/policies/transport/graphql"
	rankinggraphql "github.com/sourcegraph/sourcegraph/internal/codeintel/ranking/transport/graphql"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/lsifuploadstore"
	sharedresolvers "github.com/sourcegraph/sourcegraph/internal/codeintel/shared/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/resolvers/gitresolvers"
	uploadgraphql "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/transport/graphql"
	uploadshttp "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/transport/http"
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

	repoStore := db.Repos()
	siteAdminChecker := sharedresolvers.NewSiteAdminChecker(db)
	locationResolverFactory := gitresolvers.NewCachedLocationResolverFactory(repoStore, codeIntelServices.GitserverClient)
	uploadLoaderFactory := uploadgraphql.NewUploadLoaderFactory(codeIntelServices.UploadsService)
	autoIndexJobLoaderFactory := uploadgraphql.NewAutoIndexJobLoaderFactory(codeIntelServices.UploadsService)
	preciseIndexResolverFactory := uploadgraphql.NewPreciseIndexResolverFactory(
		codeIntelServices.UploadsService,
		codeIntelServices.PoliciesService,
		codeIntelServices.GitserverClient,
		siteAdminChecker,
		repoStore,
	)

	autoindexingRootResolver := autoindexinggraphql.NewRootResolver(
		scopedContext("autoindexing"),
		codeIntelServices.AutoIndexingService,
		siteAdminChecker,
		uploadLoaderFactory,
		autoIndexJobLoaderFactory,
		locationResolverFactory,
		preciseIndexResolverFactory,
	)

	codenavRootResolver := codenavgraphql.NewRootResolver(
		scopedContext("codenav"),
		codeIntelServices.CodenavService,
		codeIntelServices.AutoIndexingService,
		codeIntelServices.GitserverClient,
		siteAdminChecker,
		repoStore,
		uploadLoaderFactory,
		autoIndexJobLoaderFactory,
		preciseIndexResolverFactory,
		locationResolverFactory,
		ConfigInst.MaximumIndexesPerMonikerSearch,
	)

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
		siteAdminChecker,
		uploadLoaderFactory,
		autoIndexJobLoaderFactory,
		locationResolverFactory,
		preciseIndexResolverFactory,
	)

	rankingRootResolver := rankinggraphql.NewRootResolver(
		scopedContext("ranking"),
		codeIntelServices.RankingService,
		siteAdminChecker,
	)

	enterpriseServices.CodeIntelResolver = graphqlbackend.NewCodeIntelResolver(resolvers.NewCodeIntelResolver(
		autoindexingRootResolver,
		codenavRootResolver,
		policyRootResolver,
		uploadRootResolver,
		rankingRootResolver,
	))
	enterpriseServices.NewCodeIntelUploadHandler = newUploadHandler
	enterpriseServices.RankingService = codeIntelServices.RankingService
	return nil
}

func scopedContext(name string) *observation.Context {
	return observation.NewContext(log.Scoped(name + ".transport.graphql"))
}
