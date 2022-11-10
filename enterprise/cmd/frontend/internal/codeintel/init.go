package codeintel

import (
	"context"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/log"
	"go.opentelemetry.io/otel"

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
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

func init() {
	ConfigInst.Load()
}

func Init(
	ctx context.Context,
	db database.DB,
	codeIntelServices codeintel.Services,
	conf conftypes.UnifiedWatchable,
	enterpriseServices *enterprise.Services,
	observationContext *observation.Context,
) error {
	if err := ConfigInst.Validate(); err != nil {
		return err
	}

	uploadStore, err := lsifuploadstore.New(context.Background(), ConfigInst.LSIFUploadStoreConfig, observationContext)
	if err != nil {
		return err
	}

	gitserverClient := gitserver.New(db, &observation.TestContext)
	newUploadHandler := func(withCodeHostAuth bool) http.Handler {
		return uploadshttp.GetHandler(codeIntelServices.UploadsService, db, uploadStore, withCodeHostAuth)
	}

	autoindexingRootResolver := autoindexinggraphql.NewRootResolver(
		codeIntelServices.AutoIndexingService,
		codeIntelServices.UploadsService,
		codeIntelServices.PoliciesService,
		scopedContext("autoindexing"),
	)

	codenavRootResolver := codenavgraphql.NewRootResolver(
		codeIntelServices.CodenavService,
		codeIntelServices.AutoIndexingService,
		codeIntelServices.UploadsService,
		codeIntelServices.PoliciesService,
		gitserverClient,
		ConfigInst.MaximumIndexesPerMonikerSearch,
		ConfigInst.HunkCacheSize,
		scopedContext("codenav"),
	)

	policyRootResolver := policiesgraphql.NewRootResolver(
		codeIntelServices.PoliciesService,
		scopedContext("policies"),
	)

	uploadRootResolver := uploadgraphql.NewRootResolver(
		codeIntelServices.UploadsService,
		codeIntelServices.AutoIndexingService,
		codeIntelServices.PoliciesService,
		scopedContext("upload"),
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
	return &observation.Context{
		Logger:     log.Scoped(name+".transport.graphql", "codeintel "+name+" graphql transport"),
		Tracer:     &trace.Tracer{TracerProvider: otel.GetTracerProvider()},
		Registerer: prometheus.DefaultRegisterer,
	}
}
