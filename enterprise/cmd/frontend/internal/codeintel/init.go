package codeintel

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	logger "github.com/sourcegraph/log"
	"go.opentelemetry.io/otel"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	codeintelresolvers "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers"
	codeintelgqlresolvers "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers/graphql"
	autoindexinggraphql "github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/transport/graphql"
	codenavgraphql "github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/transport/graphql"
	policiesgraphql "github.com/sourcegraph/sourcegraph/internal/codeintel/policies/transport/graphql"
	uploadgraphql "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/transport/graphql"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	executorgraphql "github.com/sourcegraph/sourcegraph/internal/services/executors/transport/graphql"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

func Init(ctx context.Context, db database.DB, config *Config, enterpriseServices *enterprise.Services, services *FrontendServices) error {
	oc := func(name string) *observation.Context {
		return &observation.Context{
			Logger:     logger.Scoped(name+".transport.graphql", "codeintel "+name+" graphql transport"),
			Tracer:     &trace.Tracer{TracerProvider: otel.GetTracerProvider()},
			Registerer: prometheus.DefaultRegisterer,
		}
	}

	executorResolver := executorgraphql.New(db)
	codenavRootResolver := codenavgraphql.NewRootResolver(services.CodenavService, services.AutoIndexingService, services.UploadsService, services.PoliciesService, services.gitserverClient, config.MaximumIndexesPerMonikerSearch, config.HunkCacheSize, oc("codenav"))
	policyRootResolver := policiesgraphql.NewRootResolver(services.PoliciesService, oc("policies"))
	autoindexingRootResolver := autoindexinggraphql.NewRootResolver(services.AutoIndexingService, services.UploadsService, services.PoliciesService, oc("autoindexing"))
	uploadRootResolver := uploadgraphql.NewRootResolver(services.UploadsService, services.AutoIndexingService, services.PoliciesService, oc("upload"))
	resolvers := codeintelresolvers.NewResolver(codenavRootResolver, executorResolver, policyRootResolver, autoindexingRootResolver, uploadRootResolver)
	enterpriseServices.CodeIntelResolver = codeintelgqlresolvers.NewResolver(resolvers)
	enterpriseServices.NewCodeIntelUploadHandler = services.NewUploadHandler
	return nil
}
