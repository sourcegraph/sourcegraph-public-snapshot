package dependencies

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/transport/graphql"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// Init initializes the given enterpriseServices to include the required
// resolvers for Dependencies.
func Init(ctx context.Context, db database.DB, _ conftypes.UnifiedWatchable, enterpriseServices *enterprise.Services, observationContext *observation.Context) error {
	// Register enterprise services.
	enterpriseServices.DependenciesResolver = graphql.New(db, observationContext.Logger)
	return nil
}
