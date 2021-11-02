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
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
)

func Init(ctx context.Context, db dbutil.DB, outOfBandMigrationRunner *oobmigration.Runner, enterpriseServices *enterprise.Services, observationContext *observation.Context) error {
	if err := initServices(ctx, db); err != nil {
		return err
	}

	if err := registerMigrations(ctx, db, outOfBandMigrationRunner); err != nil {
		return err
	}

	resolver, err := newResolver(ctx, db, observationContext)
	if err != nil {
		return err
	}

	uploadHandler, err := newUploadHandler(ctx, db)
	if err != nil {
		return err
	}

	enterpriseServices.CodeIntelResolver = resolver
	enterpriseServices.NewCodeIntelUploadHandler = uploadHandler
	return nil
}

func newResolver(ctx context.Context, db dbutil.DB, observationContext *observation.Context) (gql.CodeIntelResolver, error) {
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
	)

	return codeintelgqlresolvers.NewResolver(db, innerResolver), nil
}

func newUploadHandler(ctx context.Context, db dbutil.DB) (func(internal bool) http.Handler, error) {
	internalHandler, err := NewCodeIntelUploadHandler(ctx, db, true)
	if err != nil {
		return nil, err
	}

	externalHandler, err := NewCodeIntelUploadHandler(ctx, db, false)
	if err != nil {
		return nil, err
	}

	uploadHandler := func(internal bool) http.Handler {
		if internal {
			return internalHandler
		}

		return externalHandler
	}

	return uploadHandler, nil
}
