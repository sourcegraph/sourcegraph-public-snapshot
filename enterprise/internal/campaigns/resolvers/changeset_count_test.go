package resolvers

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	ee "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/resolvers/apitest"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
)

func TestChangesetCountsOverTime(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := backend.WithAuthzBypass(context.Background())
	dbtesting.SetupGlobalTestDB(t)

	userID := insertTestUser(t, dbconn.Global, "changeset-count-resolver", true)

	now := time.Now().UTC().Truncate(time.Microsecond)
	clock := func() time.Time {
		return now.UTC().Truncate(time.Microsecond)
	}
	store := ee.NewStoreWithClock(dbconn.Global, clock)
	rstore := repos.NewDBStore(dbconn.Global, sql.TxOptions{})

	repo := newGitHubTestRepo("github.com/sourcegraph/sourcegraph", 1)
	if err := rstore.UpsertRepos(ctx, repo); err != nil {
		t.Fatal(err)
	}

	campaign := &campaigns.Campaign{
		Name:            "my-unique-name",
		NamespaceUserID: userID,
		AuthorID:        userID,
	}
	if err := store.CreateCampaign(ctx, campaign); err != nil {
		t.Fatal(err)
	}

	changeset1 := createChangeset(t, ctx, store, testChangesetOpts{
		repo: repo.ID,
		// We don't need a spec because we don't query for fields that would
		// require it
		currentSpec:         0,
		externalServiceType: "github",
		// Unpublished changesets should not be considered.
		publicationState: campaigns.ChangesetPublicationStateUnpublished,
		ownedByCampaign:  campaign.ID,
		campaign:         campaign.ID,
	})
	changeset2 := createChangeset(t, ctx, store, testChangesetOpts{
		repo: repo.ID,
		// We don't need a spec because we don't query for fields that would
		// require it
		currentSpec:         0,
		externalServiceType: "github",
		externalState:       campaigns.ChangesetExternalStateOpen,
		publicationState:    campaigns.ChangesetPublicationStatePublished,
		ownedByCampaign:     campaign.ID,
		campaign:            campaign.ID,
		metadata: &github.PullRequest{
			CreatedAt: now.Add(-1 * 24 * time.Hour),
			TimelineItems: []github.TimelineItem{
				{Type: "MergedEvent", Item: &github.LabelEvent{
					CreatedAt: now,
				}},
			},
		},
	})

	addChangeset(t, ctx, store, campaign, changeset1.ID)
	addChangeset(t, ctx, store, campaign, changeset2.ID)

	s, err := graphqlbackend.NewSchema(&Resolver{store: store}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	campaignAPIID := string(campaigns.MarshalCampaignID(campaign.ID))
	input := map[string]interface{}{"campaign": campaignAPIID}
	var response struct{ Node apitest.Campaign }
	apitest.MustExec(actor.WithActor(context.Background(), actor.FromUser(userID)), t, s, input, &response, queryChangesetCountsConnection)

	wantCounts := []apitest.ChangesetCounts{
		{Date: marshalDateTime(t, now.Add(-6*24*time.Hour))},
		{Date: marshalDateTime(t, now.Add(-5*24*time.Hour))},
		{Date: marshalDateTime(t, now.Add(-4*24*time.Hour))},
		{Date: marshalDateTime(t, now.Add(-3*24*time.Hour))},
		{Date: marshalDateTime(t, now.Add(-2*24*time.Hour))},
		{Date: marshalDateTime(t, now.Add(-1*24*time.Hour)), Total: 1, Open: 1, OpenPending: 1},
		{Date: marshalDateTime(t, now), Total: 1, Merged: 1},
	}

	if diff := cmp.Diff(wantCounts, response.Node.ChangesetCountsOverTime); diff != "" {
		t.Fatalf("wrong changesets response (-want +got):\n%s", diff)
	}
}

const queryChangesetCountsConnection = `
query($campaign: ID!) {
  node(id: $campaign) {
    ... on Campaign {
            changesetCountsOverTime {
				date
				total
				merged
				closed
				open
				openApproved
				openChangesRequested
				openPending
			}
          }
        }
      }
`
