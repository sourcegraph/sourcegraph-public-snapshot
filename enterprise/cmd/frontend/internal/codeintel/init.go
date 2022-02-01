package codeintel

import (
	"context"
	"net/http"

	"github.com/cockroachdb/errors"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	codeintelresolvers "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers"
	codeintelgqlresolvers "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers/graphql"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/policies"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/honey"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func Init(ctx context.Context, db database.DB, conf conftypes.UnifiedWatchable, enterpriseServices *enterprise.Services, observationContext *observation.Context, services *Services) error {
	if err := config.Validate(); err != nil {
		return err
	}

	resolverObservationContext := &observation.Context{
		Logger:     observationContext.Logger,
		Tracer:     observationContext.Tracer,
		Registerer: observationContext.Registerer,
		Sentry:     services.hub,
		HoneyDataset: &honey.Dataset{
			Name:       "codeintel-graphql",
			SampleRate: 4,
		},
	}

	resolver, err := newResolver(ctx, db, resolverObservationContext, services)
	if err != nil {
		return err
	}

	enterpriseServices.CodeIntelResolver = resolver
	enterpriseServices.NewCodeIntelUploadHandler = newUploadHandler(services)
	return nil
}

func newResolver(ctx context.Context, db database.DB, observationContext *observation.Context, services *Services) (gql.CodeIntelResolver, error) {
	policyMatcher := policies.NewMatcher(
		services.gitserverClient,
		policies.NoopExtractor,
		false,
		false,
	)

	hunkCache, err := codeintelresolvers.NewHunkCache(config.HunkCacheSize)
	if err != nil {
		return nil, errors.Errorf("failed to initialize hunk cache: %s", err)
	}

	innerResolver := codeintelresolvers.NewResolver(
		services.dbStore,
		services.lsifStore,
		services.gitserverClient,
		policyMatcher,
		services.indexEnqueuer,
		hunkCache,
		observationContext,
		db,
	)

	return codeintelgqlresolvers.NewResolver(db, innerResolver, &observation.Context{Sentry: observationContext.Sentry}), nil
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
