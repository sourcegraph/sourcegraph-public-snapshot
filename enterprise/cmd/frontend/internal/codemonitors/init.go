package codemonitors

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codemonitors/resolvers"
)

func Init(ctx context.Context, enterpriseServices *enterprise.Services) error {
	enterpriseServices.CodeMonitorsResolver = &resolvers.Resolver{}
	return nil
}
