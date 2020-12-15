package campaigns

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/resolvers"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/webhooks"
	"github.com/sourcegraph/sourcegraph/internal/db"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/globalstatedb"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
)

// InitFrontend initializes the given enterpriseServices to include the required resolvers for campaigns
// and sets up webhook handlers for changeset events.
func InitFrontend(ctx context.Context, enterpriseServices *enterprise.Services) error {
	globalState, err := globalstatedb.Get(ctx)
	if err != nil {
		return err
	}

	cstore := store.NewWithClock(dbconn.Global, timeutil.Now)
	esStore := db.NewExternalServicesStoreWithDB(dbconn.Global)

	enterpriseServices.CampaignsResolver = resolvers.New(dbconn.Global)
	enterpriseServices.GitHubWebhook = webhooks.NewGitHubWebhook(cstore, esStore, timeutil.Now)
	enterpriseServices.BitbucketServerWebhook = webhooks.NewBitbucketServerWebhook(
		cstore,
		esStore,
		timeutil.Now,
		"sourcegraph-"+globalState.SiteID,
	)
	enterpriseServices.GitLabWebhook = webhooks.NewGitLabWebhook(cstore, esStore, timeutil.Now)

	return nil
}
