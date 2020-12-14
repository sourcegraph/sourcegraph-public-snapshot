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

func Init(ctx context.Context, enterpriseServices *enterprise.Services) error {
	globalState, err := globalstatedb.Get(ctx)
	if err != nil {
		return err
	}

	campaignsStore := store.NewWithClock(dbconn.Global, timeutil.Now)
	repositories := repos.NewDBStore(dbconn.Global, sql.TxOptions{})

	enterpriseServices.CampaignsResolver = resolvers.NewResolver(dbconn.Global)
	enterpriseServices.GitHubWebhook = webhooks.NewGitHubWebhook(campaignsStore, repositories, timeutil.Now)
	enterpriseServices.BitbucketServerWebhook = webhooks.NewBitbucketServerWebhook(
		campaignsStore,
		repositories,
		timeutil.Now,
		"sourcegraph-"+globalState.SiteID,
	)
	enterpriseServices.GitLabWebhook = webhooks.NewGitLabWebhook(campaignsStore, repositories, timeutil.Now)

	return nil
}
