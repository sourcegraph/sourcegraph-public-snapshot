package codemonitors

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codemonitors/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

func Init(ctx context.Context, db dbutil.DB, enterpriseServices *enterprise.Services) error {
	enterpriseServices.CodeMonitorsResolver = resolvers.NewResolver(db)
	return nil
}
