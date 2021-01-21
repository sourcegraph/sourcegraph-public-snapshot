package insights

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/resolvers"
)

// Init initializes the given enterpriseServices to include the required resolvers for insights.
func Init(ctx context.Context, enterpriseServices *enterprise.Services) error {
	enterpriseServices.InsightsResolver = resolvers.New()
	return nil
}
