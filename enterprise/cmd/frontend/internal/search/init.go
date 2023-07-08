package search

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	enterprisesearch "github.com/sourcegraph/sourcegraph/enterprise/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/codeintel"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// Init initializes the given enterpriseServices to include the required
// enterprise jobs for search.
func Init(
	ctx context.Context,
	observationCtx *observation.Context,
	_ database.DB,
	_ codeintel.Services,
	_ conftypes.UnifiedWatchable,
	enterpriseServices *enterprise.Services,
) error {
	enterpriseServices.EnterpriseSearchJobs = enterprisesearch.NewEnterpriseSearchJobs()
	return nil
}
