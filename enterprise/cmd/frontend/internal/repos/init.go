package repos

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/repos/webhooks"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/repos/webhooks/resolvers"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// Init initializes the given enterpriseServices with the webhook handlers for handling GitHub push events.
func Init(
	_ context.Context,
	db database.DB,
	_ codeintel.Services,
	_ conftypes.UnifiedWatchable,
	enterpriseServices *enterprise.Services,
	_ *observation.Context,
) error {
	enterpriseServices.GitHubSyncWebhook = webhooks.NewGitHubWebhookHandler()
	enterpriseServices.WebhooksResolver = resolvers.NewWebhooksResolver(db)
	return nil
}
