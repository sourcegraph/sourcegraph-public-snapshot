package codeintel

import (
	"context"
	"net/http"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel"
	autoindexinggraphql "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing/transport/graphql"
	codenavgraphql "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/codenav/transport/graphql"
	policiesgraphql "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/policies/transport/graphql"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/gitserver"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/lsifuploadstore"
	uploadgraphql "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/transport/graphql"
	uploadshttp "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/transport/http"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func init() {
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

	gitserverClient := gitserver.New(&observation.TestContext, db)
	newUploadHandler := func(withCodeHostAuth bool) http.Handler {
		return uploadshttp.GetHandler(codeIntelServices.UploadsService, db, uploadStore, withCodeHostAuth)
	}

	autoindexingRootResolver := autoindexinggraphql.NewRootResolver(
		scopedContext("autoindexing"),
		codeIntelServices.AutoIndexingService,
		codeIntelServices.UploadsService,
		codeIntelServices.PoliciesService,
	)

	codenavRootResolver, err := codenavgraphql.NewRootResolver(
		scopedContext("codenav"),
		codeIntelServices.CodenavService,
		codeIntelServices.AutoIndexingService,
		codeIntelServices.UploadsService,
		codeIntelServices.PoliciesService,
		gitserverClient,
		ConfigInst.MaximumIndexesPerMonikerSearch,
		ConfigInst.HunkCacheSize,
	)
	if err != nil {
		return err
	}

	policyRootResolver := policiesgraphql.NewRootResolver(
		scopedContext("policies"),
		codeIntelServices.PoliciesService,
	)

	uploadRootResolver := uploadgraphql.NewRootResolver(
		scopedContext("upload"),
		codeIntelServices.UploadsService,
		codeIntelServices.AutoIndexingService,
		codeIntelServices.PoliciesService,
	)

	enterpriseServices.CodeIntelResolver = newResolver(
		autoindexingRootResolver,
		codenavRootResolver,
		policyRootResolver,
		uploadRootResolver,
	)
	enterpriseServices.NewCodeIntelUploadHandler = newUploadHandler
	enterpriseServices.RankingService = codeIntelServices.RankingService
	return nil
}

func scopedContext(name string) *observation.Context {
	return observation.NewContext(log.Scoped(name+".transport.graphql", "codeintel "+name+" graphql transport"))
}
