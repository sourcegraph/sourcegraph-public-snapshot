package codemonitors

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codemonitors/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func Init(ctx context.Context, db database.DB, _ conftypes.UnifiedWatchable, enterpriseServices *enterprise.Services, observationContext *observation.Context) error {
	enterpriseServices.CodeMonitorsResolver = resolvers.NewResolver(db)
	return nil
}
