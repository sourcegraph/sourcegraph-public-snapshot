package catalog

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	catalogresolvers "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/catalog/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func Init(ctx context.Context, db database.DB, conf conftypes.UnifiedWatchable, enterpriseServices *enterprise.Services, observationContext *observation.Context) error {
	enterpriseServices.CatalogRootResolver = catalogresolvers.NewRootResolver(db)
	return nil
}
