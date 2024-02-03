package githubapp

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/internal/codeintel"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func Init(
	ctx context.Context,
	_ *observation.Context,
	db database.DB,
	_ codeintel.Services,
	_ conftypes.UnifiedWatchable,
	s *enterprise.Services,
) error {
	s.GitHubAppsResolver = NewResolver(log.Scoped("GitHubAppsResolver"), db)
	return nil
}
