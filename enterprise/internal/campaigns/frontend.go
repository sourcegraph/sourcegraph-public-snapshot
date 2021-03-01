package campaigns

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/background"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/resolvers"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/webhooks"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/database/globalstatedb"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
)

// InitFrontend initializes the given enterpriseServices to include the required resolvers for campaigns
// and sets up webhook handlers for changeset events.
func InitFrontend(ctx context.Context, db dbutil.DB, outOfBandMigrationRunner *oobmigration.Runner, enterpriseServices *enterprise.Services) error {
	globalState, err := globalstatedb.Get(ctx)
	if err != nil {
		return err
	}

	cstore := store.New(db)

	enterpriseServices.CampaignsResolver = resolvers.New(cstore)
	enterpriseServices.GitHubWebhook = webhooks.NewGitHubWebhook(cstore)
	enterpriseServices.BitbucketServerWebhook = webhooks.NewBitbucketServerWebhook(
		cstore,
		"sourcegraph-"+globalState.SiteID,
	)
	enterpriseServices.GitLabWebhook = webhooks.NewGitLabWebhook(cstore)

	return background.RegisterMigrations(cstore, outOfBandMigrationRunner)
}
