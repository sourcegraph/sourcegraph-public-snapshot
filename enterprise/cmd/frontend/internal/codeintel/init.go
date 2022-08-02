package codeintel

import (
	"context"
	"net/http"

	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	logger "github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	codeintelresolvers "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers"
	codeintelgqlresolvers "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers/graphql"
	codenavgraphql "github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/transport/graphql"
	policies "github.com/sourcegraph/sourcegraph/internal/codeintel/policies/enterprise"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/honey"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/symbols"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

func Init(ctx context.Context, db database.DB, config *Config, enterpriseServices *enterprise.Services, observationContext *observation.Context, services *Services) error {
	resolverObservationContext := &observation.Context{
		Logger:     observationContext.Logger,
		Tracer:     observationContext.Tracer,
		Registerer: observationContext.Registerer,
		HoneyDataset: &honey.Dataset{
			Name:       "codeintel-graphql",
			SampleRate: 4,
		},
	}

	enterpriseServices.CodeIntelResolver = newResolver(db, config, resolverObservationContext, services)
	enterpriseServices.NewCodeIntelUploadHandler = newUploadHandler(services)
	return nil
}

func newResolver(db database.DB, config *Config, observationContext *observation.Context, services *Services) gql.CodeIntelResolver {
	policyMatcher := policies.NewMatcher(services.gitserverClient, policies.NoopExtractor, false, false)

	codenavCtx := &observation.Context{
		Logger:     logger.Scoped("codenav.transport.graphql", "codeintel symbols graphql transport"),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}
	codenavResolver := codenavgraphql.New(services.CodeNavSvc, config.HunkCacheSize, codenavCtx)

	innerResolver := codeintelresolvers.NewResolver(
		services.dbStore,
		services.lsifStore,
		services.gitserverClient,
		policyMatcher,
		services.indexEnqueuer,
		symbols.DefaultClient,
		config.MaximumIndexesPerMonikerSearch,
		observationContext,
		db,
		codenavResolver,
	)

	obsCtx := &observation.Context{
		Logger:       nil,
		Tracer:       &trace.Tracer{},
		Registerer:   nil,
		HoneyDataset: &honey.Dataset{},
	}

	return codeintelgqlresolvers.NewResolver(db, services.gitserverClient, innerResolver, obsCtx)
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
