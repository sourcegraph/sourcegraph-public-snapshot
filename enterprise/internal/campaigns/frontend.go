package campaigns

import (
	"context"
	"database/sql"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/resolvers"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/webhooks"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/globalstatedb"
	"github.com/sourcegraph/sourcegraph/internal/repos"
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
	rstore := repos.NewDBStore(dbconn.Global, sql.TxOptions{})

	enterpriseServices.CampaignsResolver = resolvers.New(dbconn.Global)
	enterpriseServices.GitHubWebhook = webhooks.NewGitHubWebhook(cstore, rstore, timeutil.Now)
	enterpriseServices.BitbucketServerWebhook = webhooks.NewBitbucketServerWebhook(
		cstore,
		rstore,
		timeutil.Now,
		"sourcegraph-"+globalState.SiteID,
	)
	enterpriseServices.GitLabWebhook = webhooks.NewGitLabWebhook(cstore, rstore, timeutil.Now)

	return nil
}
