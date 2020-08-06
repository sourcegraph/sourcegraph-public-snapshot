package resolvers

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	ee "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/resolvers/apitest"
	ct "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/testing"
	"github.com/sourcegraph/sourcegraph/internal/actor"
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
	inaccessibleRepo := newGitHubTestRepo("github.com/sourcegraph/private", 2)
	if err := rstore.UpsertRepos(ctx, repo, inaccessibleRepo); err != nil {
		t.Fatal(err)
	}
	ct.AuthzFilterRepos(t, inaccessibleRepo.ID)

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
		externalReviewState: campaigns.ChangesetReviewStatePending,
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
		externalReviewState: campaigns.ChangesetReviewStatePending,
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
		externalReviewState: campaigns.ChangesetReviewStatePending,
		ownedByCampaign:     campaign.ID,
		campaign:            campaign.ID,
	})
	changeset4 := createChangeset(t, ctx, store, testChangesetOpts{
		repo:                inaccessibleRepo.ID,
		externalServiceType: "github",
		externalID:          "987651",
		externalBranch:      "open-hidden-pr",
		externalState:       campaigns.ChangesetExternalStateOpen,
		publicationState:    campaigns.ChangesetPublicationStatePublished,
		externalReviewState: campaigns.ChangesetReviewStatePending,
		ownedByCampaign:     campaign.ID,
		campaign:            campaign.ID,
	})

	addChangeset(t, ctx, store, campaign, changeset1.ID)
	addChangeset(t, ctx, store, campaign, changeset2.ID)
	addChangeset(t, ctx, store, campaign, changeset3.ID)
	addChangeset(t, ctx, store, campaign, changeset4.ID)

	s, err := graphqlbackend.NewSchema(&Resolver{store: store}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	campaignApiID := string(campaigns.MarshalCampaignID(campaign.ID))
	nodes := []apitest.Changeset{
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
		{
			Typename: "HiddenExternalChangeset",
			ID:       string(marshalChangesetID(changeset4.ID)),
		},
	}

	tests := []struct {
		First           int
		unsafeOpts      bool
		wantHasNextPage bool
		wantTotalCount  int
		wantOpen        int
		wantNodes       []apitest.Changeset
	}{
		{
			First:           1,
			wantHasNextPage: true,
			wantTotalCount:  4,
			wantOpen:        2,
			wantNodes:       nodes[:1],
		},
		{
			First:           2,
			wantHasNextPage: true,
			wantTotalCount:  4,
			wantOpen:        2,
			wantNodes:       nodes[:2],
		},
		{
			First:           3,
			wantHasNextPage: true,
			wantTotalCount:  4,
			wantOpen:        2,
			wantNodes:       nodes[:3],
		},
		{
			First:           4,
			wantHasNextPage: false,
			wantTotalCount:  4,
			wantOpen:        2,
			wantNodes:       nodes[:4],
		},
		{
			First:           1,
			unsafeOpts:      true,
			wantHasNextPage: true,
			// Expect only 3 changesets to be returned when an unsafe filter is applied.
			wantTotalCount: 3,
			wantOpen:       1,
			wantNodes:      nodes[:1],
		},
		{
			First:           2,
			unsafeOpts:      true,
			wantHasNextPage: true,
			wantTotalCount:  3,
			wantOpen:        1,
			wantNodes:       nodes[:2],
		},
		{
			First:           3,
			unsafeOpts:      true,
			wantHasNextPage: false,
			wantTotalCount:  3,
			wantOpen:        1,
			wantNodes:       nodes[:3],
		},
	}

	reviewStatePending := campaigns.ChangesetReviewStatePending

	for _, tc := range tests {
		t.Run(fmt.Sprintf("Unsafe opts %t, first %d", tc.unsafeOpts, tc.First), func(t *testing.T) {
			var reviewState *campaigns.ChangesetReviewState
			if tc.unsafeOpts {
				reviewState = &reviewStatePending
			}
			input := map[string]interface{}{"campaign": campaignApiID, "first": int64(tc.First), "reviewState": reviewState}
			var response struct{ Node apitest.Campaign }
			apitest.MustExec(actor.WithActor(context.Background(), actor.FromUser(userID)), t, s, input, &response, queryChangesetConnection)

			wantChangesets := apitest.ChangesetConnection{
				Stats: apitest.ChangesetConnectionStats{
					Unpublished: 1,
					Open:        tc.wantOpen,
					Merged:      1,
					Closed:      0,
					Total:       tc.wantTotalCount,
				},
				TotalCount: tc.wantTotalCount,
				PageInfo: apitest.PageInfo{
					HasNextPage: tc.wantHasNextPage,
				},
				Nodes: tc.wantNodes,
			}

			if diff := cmp.Diff(wantChangesets, response.Node.Changesets); diff != "" {
				t.Fatalf("wrong changesets response (-want +got):\n%s", diff)
			}
		})
	}
}

const queryChangesetConnection = `
query($campaign: ID!, $first: Int, $reviewState: ChangesetReviewState){
  node(id: $campaign) {
    ... on Campaign {
      changesets(first: $first, reviewState: $reviewState) {
        totalCount
        stats { unpublished, open, merged, closed, total }
        nodes {
          __typename

          ... on ExternalChangeset {
            id
			repository { name }
			nextSyncAt
          }
          ... on HiddenExternalChangeset {
            id
			nextSyncAt
          }
		}
		pageInfo {
		  hasNextPage
		}
      }
    }
  }
}
`
