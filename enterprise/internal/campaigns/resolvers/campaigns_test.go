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
)

func TestCampaignResolver(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := backend.WithAuthzBypass(context.Background())
	dbtesting.SetupGlobalTestDB(t)

	userID := insertTestUser(t, dbconn.Global, "campaign-resolver", true)

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

	campaignSpec := &campaigns.CampaignSpec{UserID: userID, NamespaceUserID: userID}
	if err := store.CreateCampaignSpec(ctx, campaignSpec); err != nil {
		t.Fatal(err)
	}

	campaign := &campaigns.Campaign{
		CampaignSpecID:  campaignSpec.ID,
		Name:            "my-unique-name",
		Description:     "The campaign description",
		NamespaceUserID: userID,
		AuthorID:        userID,
	}
	if err := store.CreateCampaign(ctx, campaign); err != nil {
		t.Fatal(err)
	}

	// Unpublished changeset with non-importing changesetSpec
	unpublishedSpec := createChangesetSpec(t, ctx, store, testSpecOpts{
		user:          userID,
		repo:          repo.ID,
		campaignSpec:  campaignSpec.ID,
		headRef:       "refs/heads/my-new-branch",
		published:     false,
		title:         "ChangesetSpec Title",
		body:          "ChangesetSpec Body",
		commitMessage: "The commit message",
		commitDiff:    testDiff,
	})
	unpublishedChangeset := createChangeset(t, ctx, store, testChangesetOpts{
		repo:                repo.ID,
		currentSpec:         unpublishedSpec.ID,
		externalServiceType: "github",
		publicationState:    campaigns.ChangesetPublicationStateUnpublished,
		createdByCampaign:   false,
		ownedByCampaign:     campaign.ID,
		campaign:            campaign.ID,
	})
	addChangeset(t, ctx, store, campaign, unpublishedChangeset.ID)

	// Published changeset with ExternalState Open
	openGitHubPR := buildGithubPR(now, "12345", "OPEN", "Open GitHub PR", "Open GitHub PR Body", "open-pr")
	publishedOpenChangeset := createChangeset(t, ctx, store, testChangesetOpts{
		repo: repo.ID,
		// We don't need a spec, because the resolver should take all the data
		// out of the changeset.
		currentSpec:         0,
		externalServiceType: "github",
		externalID:          "12345",
		externalBranch:      "open-pr",
		externalState:       campaigns.ChangesetExternalStateOpen,
		publicationState:    campaigns.ChangesetPublicationStatePublished,
		createdByCampaign:   false,
		ownedByCampaign:     campaign.ID,
		campaign:            campaign.ID,
		metadata:            openGitHubPR,
	})
	addChangeset(t, ctx, store, campaign, publishedOpenChangeset.ID)

	mergedGitHubPR := buildGithubPR(now, "56789", "MERGED", "Open GitHub PR", "Open GitHub PR Body", "refs/heads/open-pr")
	publishedMergedChangeset := createChangeset(t, ctx, store, testChangesetOpts{
		repo: repo.ID,
		// We don't need a spec, because the resolver should take all the data
		// out of the changeset.
		currentSpec:         0,
		externalServiceType: "github",
		externalID:          "56789",
		externalBranch:      "merged-pr",
		externalState:       campaigns.ChangesetExternalStateMerged,
		publicationState:    campaigns.ChangesetPublicationStatePublished,
		createdByCampaign:   false,
		ownedByCampaign:     campaign.ID,
		campaign:            campaign.ID,
		metadata:            mergedGitHubPR,
	})
	addChangeset(t, ctx, store, campaign, publishedMergedChangeset.ID)

	s, err := graphqlbackend.NewSchema(&Resolver{store: store}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	campaignApiID := string(campaigns.MarshalCampaignID(campaign.ID))

	input := map[string]interface{}{"campaign": campaignApiID}
	var response struct{ Node apitest.Campaign }
	actorCtx := actor.WithActor(ctx, actor.FromUser(userID))
	apitest.MustExec(actorCtx, t, s, input, &response, queryCampaign)

	wantCampaign := apitest.Campaign{
		ID:          campaignApiID,
		Name:        campaign.Name,
		Description: campaign.Description,
		Namespace:   apitest.UserOrg{DatabaseID: userID, SiteAdmin: true},
		Author:      apitest.User{DatabaseID: userID, SiteAdmin: true},
		URL:         "/campaigns/" + campaignApiID,
		Changesets: apitest.ChangesetConnection{
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
					Typename:      "ExternalChangeset",
					ID:            string(marshalChangesetID(unpublishedChangeset.ID)),
					Title:         unpublishedSpec.Spec.Title,
					Body:          unpublishedSpec.Spec.Body,
					ExternalState: "",
				},
				{
					Typename:      "ExternalChangeset",
					ID:            string(marshalChangesetID(publishedOpenChangeset.ID)),
					Title:         openGitHubPR.Title,
					Body:          openGitHubPR.Body,
					ExternalState: "OPEN",
				},
				{
					Typename:      "ExternalChangeset",
					ID:            string(marshalChangesetID(publishedMergedChangeset.ID)),
					Title:         mergedGitHubPR.Title,
					Body:          mergedGitHubPR.Body,
					ExternalState: "MERGED",
				},
			},
		},
	}
	if diff := cmp.Diff(wantCampaign, response.Node); diff != "" {
		t.Fatalf("wrong campaign response (-want +got):\n%s", diff)
	}
}

const queryCampaign = `
fragment u on User { databaseID, siteAdmin }
fragment o on Org  { name }

query($campaign: ID!){
  node(id: $campaign) {
    ... on Campaign {
      id, name, description, branch
      author    { ...u }
      namespace {
        ... on User { ...u }
        ... on Org  { ...o }
      }
      url

      changesets {
        totalCount
		stats { unpublished, open, merged, closed, total }
        nodes {
          __typename

          ... on ExternalChangeset {
            id
            title
            body
            externalState
          }
        }
      }
    }
  }
}
`
