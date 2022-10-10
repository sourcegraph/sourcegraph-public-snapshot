package codeintel

import (
	"context"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/log"
	"go.opentelemetry.io/otel"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/internal/codeintel"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing"
	autoindexinggraphql "github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/transport/graphql"
	codenavgraphql "github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/transport/graphql"
	policiesgraphql "github.com/sourcegraph/sourcegraph/internal/codeintel/policies/transport/graphql"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/lsifuploadstore"
	uploadgraphql "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/transport/graphql"
	uploadshttp "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/transport/http"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	executorgraphql "github.com/sourcegraph/sourcegraph/internal/services/executors/transport/graphql"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

func Init(
	ctx context.Context,
	db database.DB,
	siteConfig conftypes.WatchableSiteConfig,
	config *Config,
	enterpriseServices *enterprise.Services,
) (
	*autoindexing.Service,
	http.Handler,
	error,
) {
	logger := log.Scoped("codeintel", "codeintel services")

	services, err := codeintel.GetServices(codeintel.Databases{
		DB:          db,
		CodeIntelDB: mustInitializeCodeIntelDB(logger),
	})
	if err != nil {
		return nil, nil, err
	}

	// Initialize blob stores
	observationContext := &observation.Context{
		Logger:     logger,
		Tracer:     &trace.Tracer{TracerProvider: otel.GetTracerProvider()},
		Registerer: prometheus.DefaultRegisterer,
	}
	uploadStore, err := lsifuploadstore.New(context.Background(), config.LSIFUploadStoreConfig, observationContext)
	if err != nil {
		return nil, nil, err
	}

	gitserverClient := gitserver.New(db, &observation.TestContext)
	newUploadHandler := func(withCodeHostAuth bool) http.Handler {
		return uploadshttp.GetHandler(services.UploadsService, db, uploadStore, withCodeHostAuth)
	}

	autoindexingRootResolver := autoindexinggraphql.NewRootResolver(
		services.AutoIndexingService,
		services.UploadsService,
		services.PoliciesService,
		scopedContext("autoindexing"),
	)

	codenavRootResolver := codenavgraphql.NewRootResolver(
		services.CodenavService,
		services.AutoIndexingService,
		services.UploadsService,
		services.PoliciesService,
		gitserverClient,
		config.MaximumIndexesPerMonikerSearch,
		config.HunkCacheSize,
		scopedContext("codenav"),
	)

	executorResolver := executorgraphql.New(db)

	policyRootResolver := policiesgraphql.NewRootResolver(
		services.PoliciesService,
		scopedContext("policies"),
	)

	uploadRootResolver := uploadgraphql.NewRootResolver(
		services.UploadsService,
		services.AutoIndexingService,
		services.PoliciesService,
		scopedContext("upload"),
	)

	enterpriseServices.CodeIntelResolver = newResolver(
		autoindexingRootResolver,
		codenavRootResolver,
		executorResolver,
		policyRootResolver,
		uploadRootResolver,
	)
	enterpriseServices.NewCodeIntelUploadHandler = newUploadHandler
	return services.AutoIndexingService, newUploadHandler(false), nil
}

func mustInitializeCodeIntelDB(logger log.Logger) stores.CodeIntelDB {
	dsn := conf.GetServiceConnectionValueAndRestartOnChange(func(serviceConnections conftypes.ServiceConnections) string {
		return serviceConnections.CodeIntelPostgresDSN
	})

	db, err := connections.EnsureNewCodeIntelDB(dsn, "frontend", &observation.TestContext)
	if err != nil {
		logger.Fatal("Failed to connect to codeintel database", log.Error(err))
	}

	return stores.NewCodeIntelDB(db)
}

func scopedContext(name string) *observation.Context {
	return &observation.Context{
		Logger:     log.Scoped(name+".transport.graphql", "codeintel "+name+" graphql transport"),
		Tracer:     &trace.Tracer{TracerProvider: otel.GetTracerProvider()},
		Registerer: prometheus.DefaultRegisterer,
	}
}
