package resolvers

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	ee "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/resolvers/apitest"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
)

func TestChangesetConnectionResolver(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := backend.WithAuthzBypass(context.Background())
	dbtesting.SetupGlobalTestDB(t)

	userID := insertTestUser(t, dbconn.Global, "changeset-connection-resolver", true)

	store := ee.NewStore(dbconn.Global)
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
		publicationState:    campaigns.ChangesetPublicationStateUnpublished,
		ownedByCampaign:     campaign.ID,
		campaign:            campaign.ID,
	})

	changeset2 := createChangeset(t, ctx, store, testChangesetOpts{
		repo:                repo.ID,
		externalServiceType: "github",
		externalID:          "12345",
		externalBranch:      "open-pr",
		externalState:       campaigns.ChangesetExternalStateOpen,
		publicationState:    campaigns.ChangesetPublicationStatePublished,
		ownedByCampaign:     campaign.ID,
		campaign:            campaign.ID,
	})

	changeset3 := createChangeset(t, ctx, store, testChangesetOpts{
		repo:                repo.ID,
		externalServiceType: "github",
		externalID:          "56789",
		externalBranch:      "merged-pr",
		externalState:       campaigns.ChangesetExternalStateMerged,
		publicationState:    campaigns.ChangesetPublicationStatePublished,
		ownedByCampaign:     campaign.ID,
		campaign:            campaign.ID,
	})

	addChangeset(t, ctx, store, campaign, changeset1.ID)
	addChangeset(t, ctx, store, campaign, changeset2.ID)
	addChangeset(t, ctx, store, campaign, changeset3.ID)

	s, err := graphqlbackend.NewSchema(&Resolver{store: store}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	campaignApiID := string(campaigns.MarshalCampaignID(campaign.ID))

	input := map[string]interface{}{"campaign": campaignApiID}
	var response struct{ Node apitest.Campaign }
	apitest.MustExec(ctx, t, s, input, &response, queryChangesetConnection)

	wantChangesets := apitest.ChangesetConnection{
		Stats: apitest.ChangesetConnectionStats{
			Unpublished: 1,
			Open:        1,
			Merged:      1,
			Closed:      0,
			Total:       3,
		},
		TotalCount: 3,
		Nodes: []apitest.Changeset{
			{
				Typename:   "ExternalChangeset",
				ID:         string(marshalChangesetID(changeset1.ID)),
				Repository: apitest.Repository{Name: repo.Name},
			},
			{
				Typename:   "ExternalChangeset",
				ID:         string(marshalChangesetID(changeset2.ID)),
				Repository: apitest.Repository{Name: repo.Name},
			},
			{
				Typename:   "ExternalChangeset",
				ID:         string(marshalChangesetID(changeset3.ID)),
				Repository: apitest.Repository{Name: repo.Name},
			},
		},
	}

	if diff := cmp.Diff(wantChangesets, response.Node.Changesets); diff != "" {
		t.Fatalf("wrong changesets response (-want +got):\n%s", diff)
	}
}

const queryChangesetConnection = `
query($campaign: ID!){
  node(id: $campaign) {
    ... on Campaign {
      changesets {
        totalCount
        stats { unpublished, open, merged, closed, total }
        nodes {
          __typename

          ... on ExternalChangeset {
            id
			repository { name }
			nextSyncAt
          }
        }
      }
    }
  }
}
`
