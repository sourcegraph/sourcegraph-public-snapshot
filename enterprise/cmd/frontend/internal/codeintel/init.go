package codeintel

import (
	"context"
	"net/http"

	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	logger "github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	codeintelresolvers "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers"
	codeintelgqlresolvers "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers/graphql"
	codenavgraphql "github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/transport/graphql"
	policies "github.com/sourcegraph/sourcegraph/internal/codeintel/policies/enterprise"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/honey"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	executorgraphql "github.com/sourcegraph/sourcegraph/internal/services/executors/transport/graphql"
	"github.com/sourcegraph/sourcegraph/internal/symbols"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

func Init(ctx context.Context, db database.DB, config *Config, enterpriseServices *enterprise.Services, services *Services) error {
	policyMatcher := policies.NewMatcher(services.gitserverClient, policies.NoopExtractor, false, false)

	codenavCtx := &observation.Context{
		Logger:     logger.Scoped("codenav.transport.graphql", "codeintel symbols graphql transport"),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}
	codenavResolver := codenavgraphql.New(services.CodeNavSvc, services.gitserverClient, config.MaximumIndexesPerMonikerSearch, config.HunkCacheSize, codenavCtx)
	executorResolver := executorgraphql.New(db)

	innerResolver := codeintelresolvers.NewResolver(
		services.dbStore,
		services.lsifStore,
		services.gitserverClient,
		policyMatcher,
		services.indexEnqueuer,
		symbols.DefaultClient,
		codenavResolver,
		executorResolver,
	)

	observationCtx := &observation.Context{Logger: nil, Tracer: &trace.Tracer{}, Registerer: nil, HoneyDataset: &honey.Dataset{}}

	enterpriseServices.CodeIntelResolver = codeintelgqlresolvers.NewResolver(db, services.gitserverClient, innerResolver, observationCtx)
	enterpriseServices.NewCodeIntelUploadHandler = newUploadHandler(services)

	return nil
}

func newUploadHandler(services *Services) func(internal bool) http.Handler {
	uploadHandler := func(internal bool) http.Handler {
		if internal {
			return services.InternalUploadHandler
		}

		return services.ExternalUploadHandler
	}

	return uploadHandler
}
