package resolvers

import (
	"context"
	"database/sql"
	"fmt"
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

func TestChangesetEventConnectionResolver(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := backend.WithAuthzBypass(context.Background())
	dbtesting.SetupGlobalTestDB(t)

	userID := insertTestUser(t, dbconn.Global, "changeset-event-connection-resolver", true)

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
		Name:             "my-unique-name",
		NamespaceUserID:  userID,
		InitialApplierID: userID,
	}
	if err := store.CreateCampaign(ctx, campaign); err != nil {
		t.Fatal(err)
	}

	changeset := createChangeset(t, ctx, store, testChangesetOpts{
		repo:                repo.ID,
		externalServiceType: "github",
		publicationState:    campaigns.ChangesetPublicationStateUnpublished,
		externalReviewState: campaigns.ChangesetReviewStatePending,
		ownedByCampaign:     campaign.ID,
		campaign:            campaign.ID,
		metadata: &github.PullRequest{
			TimelineItems: []github.TimelineItem{
				{Type: "PullRequestCommit", Item: &github.PullRequestCommit{
					Commit: github.Commit{
						OID: "d34db33f",
					},
				}},
				{Type: "LabeledEvent", Item: &github.LabelEvent{
					Label: github.Label{
						ID:    "label-event",
						Name:  "cool-label",
						Color: "blue",
					},
				}},
			},
		},
	})

	addChangeset(t, ctx, store, campaign, changeset.ID)

	s, err := graphqlbackend.NewSchema(&Resolver{store: store}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	changesetAPIID := string(marshalChangesetID(changeset.ID))
	nodes := []apitest.ChangesetEvent{
		{
			ID:        string(marshalChangesetEventID(1)),
			Changeset: struct{ ID string }{ID: changesetAPIID},
			CreatedAt: marshalDateTime(t, now),
		},
		{
			ID:        string(marshalChangesetEventID(2)),
			Changeset: struct{ ID string }{ID: changesetAPIID},
			CreatedAt: marshalDateTime(t, now),
		},
	}

	tests := []struct {
		firstParam      int
		wantHasNextPage bool
		wantTotalCount  int
		wantNodes       []apitest.ChangesetEvent
	}{
		{firstParam: 1, wantHasNextPage: true, wantTotalCount: 2, wantNodes: nodes[:1]},
		{firstParam: 2, wantHasNextPage: false, wantTotalCount: 2, wantNodes: nodes},
		{firstParam: 3, wantHasNextPage: false, wantTotalCount: 2, wantNodes: nodes},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("first=%d", tc.firstParam), func(t *testing.T) {
			input := map[string]interface{}{"changeset": changesetAPIID, "first": int64(tc.firstParam)}
			var response struct{ Node apitest.Changeset }
			apitest.MustExec(actor.WithActor(context.Background(), actor.FromUser(userID)), t, s, input, &response, queryChangesetEventConnection)

			wantEvents := apitest.ChangesetEventConnection{
				TotalCount: tc.wantTotalCount,
				PageInfo: apitest.PageInfo{
					HasNextPage: tc.wantHasNextPage,
				},
				Nodes: tc.wantNodes,
			}

			if diff := cmp.Diff(wantEvents, response.Node.Events); diff != "" {
				t.Fatalf("wrong changesets response (-want +got):\n%s", diff)
			}
		})
	}
}

const queryChangesetEventConnection = `
query($changeset: ID!, $first: Int){
  node(id: $changeset) {
    ... on ExternalChangeset {
      events(first: $first) {
        totalCount
        pageInfo {
          hasNextPage
        }
        nodes {
         id
         createdAt
         changeset {
           id
         }
        }
      }
    }
  }
}
`
