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
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
)

func TestChangesetResolver(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := backend.WithAuthzBypass(context.Background())
	dbtesting.SetupGlobalTestDB(t)

	userID := insertTestUser(t, dbconn.Global, "campaign-resolver", true)

	now := time.Now().UTC().Truncate(time.Microsecond)
	store := ee.NewStore(dbconn.Global)
	rstore := repos.NewDBStore(dbconn.Global, sql.TxOptions{})

	repo := newGitHubTestRepo("github.com/sourcegraph/sourcegraph", 1)
	if err := rstore.UpsertRepos(ctx, repo); err != nil {
		t.Fatal(err)
	}

	unpublishedSpec := createChangesetSpec(t, ctx, store, testSpecOpts{
		user:          userID,
		repo:          repo.ID,
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
	})

	syncedGitHubChangeset := createChangeset(t, ctx, store, testChangesetOpts{
		repo: repo.ID,
		// We don't need a spec, because the resolver should take all the data
		// out of the changeset.
		currentSpec:         0,
		externalServiceType: "github",
		externalID:          "12345",
		externalBranch:      "open-pr",
		externalState:       campaigns.ChangesetExternalStateOpen,
		externalCheckState:  campaigns.ChangesetCheckStatePending,
		externalReviewState: campaigns.ChangesetReviewStateChangesRequested,
		publicationState:    campaigns.ChangesetPublicationStatePublished,
		createdByCampaign:   false,
		metadata: &github.PullRequest{
			ID:          "12345",
			Title:       "GitHub PR Title",
			Body:        "GitHub PR Body",
			Number:      12345,
			State:       "OPEN",
			URL:         "https://github.com/sourcegraph/sourcegraph/pull/12345",
			HeadRefName: "open-pr",
			HeadRefOid:  "d34db33f",
			BaseRefOid:  "f00b4r",
			BaseRefName: "master",
			TimelineItems: []github.TimelineItem{
				{Type: "PullRequestCommit", Item: &github.PullRequestCommit{
					Commit: github.Commit{
						OID:           "d34db33f",
						PushedDate:    now,
						CommittedDate: now,
					},
				}},
				{Type: "LabeledEvent", Item: &github.LabelEvent{
					CreatedAt: now.Add(5 * time.Second),
					Label: github.Label{
						Name:        "cool-label",
						Color:       "blue",
						Description: "the best label in town",
					},
				}},
			},
			CreatedAt: now,
			UpdatedAt: now,
		},
	})
	events := syncedGitHubChangeset.Events()
	if err := store.UpsertChangesetEvents(ctx, events...); err != nil {
		t.Fatal(err)
	}

	s, err := graphqlbackend.NewSchema(&Resolver{store: store}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		changeset *campaigns.Changeset
		want      apitest.Changeset
	}{
		{
			changeset: unpublishedChangeset,
			want: apitest.Changeset{
				Typename:   "ExternalChangeset",
				Title:      unpublishedSpec.Spec.Title,
				Body:       unpublishedSpec.Spec.Body,
				Repository: apitest.Repository{Name: repo.Name},
				Labels:     []apitest.Label{},
			},
		},
		{
			changeset: syncedGitHubChangeset,
			want: apitest.Changeset{
				Typename:      "ExternalChangeset",
				Title:         "GitHub PR Title",
				Body:          "GitHub PR Body",
				ExternalState: "OPEN",
				ExternalID:    "12345",
				CheckState:    "PENDING",
				ReviewState:   "CHANGES_REQUESTED",
				Repository:    apitest.Repository{Name: repo.Name},
				ExternalURL: apitest.ExternalURL{
					URL:         "https://github.com/sourcegraph/sourcegraph/pull/12345",
					ServiceType: "github",
				},
				Events: apitest.ChangesetEventConnection{
					TotalCount: 2,
				},
				Labels: []apitest.Label{
					{Text: "cool-label", Color: "blue", Description: "the best label in town"},
				},
			},
		},
	}

	for _, tc := range tests {
		apiID := marshalChangesetID(tc.changeset.ID)
		input := map[string]interface{}{"changeset": apiID}

		var response struct{ Node apitest.Changeset }
		apitest.MustExec(ctx, t, s, input, &response, queryChangeset)

		tc.want.ID = string(apiID)
		if diff := cmp.Diff(tc.want, response.Node); diff != "" {
			t.Fatalf("wrong campaign response (-want +got):\n%s", diff)
		}
	}
}

const queryChangeset = `
query($changeset: ID!) {
  node(id: $changeset) {
    __typename

    ... on ExternalChangeset {
      id

      title
      body

      externalID
      externalState
      reviewState
      checkState
      externalURL { url, serviceType }
      nextSyncAt

      repository { name }

      events(first: 100) { totalCount }
      labels { text, color, description }
    }
  }
}
`
