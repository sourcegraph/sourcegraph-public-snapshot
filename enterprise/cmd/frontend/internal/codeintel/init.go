package codeintel

import (
	"context"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	logger "github.com/sourcegraph/log"
	"go.opentelemetry.io/otel"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing"
	autoindexinggraphql "github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/transport/graphql"
	codenavgraphql "github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/transport/graphql"
	policiesgraphql "github.com/sourcegraph/sourcegraph/internal/codeintel/policies/transport/graphql"
	uploadgraphql "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/transport/graphql"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	executorgraphql "github.com/sourcegraph/sourcegraph/internal/services/executors/transport/graphql"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

func Init(
	ctx context.Context,
	config *Config,
	siteConfig conftypes.WatchableSiteConfig,
	db database.DB,
	enterpriseServices *enterprise.Services,
) (
	*autoindexing.Service,
	http.Handler,
	error,
) {
	services, err := NewServices(ctx, config, siteConfig, db)
	if err != nil {
		return nil, nil, err
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
		services.gitserverClient,
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
	enterpriseServices.NewCodeIntelUploadHandler = services.NewUploadHandler
	return services.AutoIndexingService, services.NewUploadHandler(false), nil
}

func scopedContext(name string) *observation.Context {
	return &observation.Context{
		Logger:     logger.Scoped(name+".transport.graphql", "codeintel "+name+" graphql transport"),
		Tracer:     &trace.Tracer{TracerProvider: otel.GetTracerProvider()},
		Registerer: prometheus.DefaultRegisterer,
	}
}
