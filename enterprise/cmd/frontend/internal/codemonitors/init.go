package codemonitors

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codemonitors/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
)

func Init(ctx context.Context, enterpriseServices *enterprise.Services) error {
	enterpriseServices.CodeMonitorsResolver = resolvers.NewResolver(dbconn.Global)
	return nil
}
