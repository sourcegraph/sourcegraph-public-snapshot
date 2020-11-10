package campaigns

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/db"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/globalstatedb"
)

func Init(ctx context.Context, enterpriseServices *enterprise.Services) error {
	globalState, err := globalstatedb.Get(ctx)
	if err != nil {
		return err
	}

	campaignsStore := campaigns.NewStoreWithClock(dbconn.Global, msResolutionClock)
	externalServices := db.NewExternalServicesStoreWithDB(dbconn.Global)

	enterpriseServices.CampaignsResolver = resolvers.NewResolver(dbconn.Global)
	enterpriseServices.GitHubWebhook = campaigns.NewGitHubWebhook(campaignsStore, externalServices, msResolutionClock)
	enterpriseServices.BitbucketServerWebhook = campaigns.NewBitbucketServerWebhook(
		campaignsStore,
		externalServices,
		msResolutionClock,
		"sourcegraph-"+globalState.SiteID,
	)
	enterpriseServices.GitLabWebhook = campaigns.NewGitLabWebhook(campaignsStore, externalServices, msResolutionClock)

	return nil
}

var msResolutionClock = func() time.Time { return time.Now().UTC().Truncate(time.Microsecond) }
