package resolvers

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/batches/resolvers/apitest"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	bgql "github.com/sourcegraph/sourcegraph/internal/batches/graphql"
	"github.com/sourcegraph/sourcegraph/internal/batches/store"
	bt "github.com/sourcegraph/sourcegraph/internal/batches/testing"
	btypes "github.com/sourcegraph/sourcegraph/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestChangesetConnectionResolver(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)

	ctx := actor.WithInternalActor(context.Background())
	db := database.NewDB(logger, dbtest.NewDB(t))

	userID := bt.CreateTestUser(t, db, false).ID

	bstore := store.New(db, observation.TestContextTB(t), nil)
	repoStore := database.ReposWith(logger, bstore)
	esStore := database.ExternalServicesWith(logger, bstore)

	repo := newGitHubTestRepo("github.com/sourcegraph/changeset-connection-test", newGitHubExternalService(t, esStore))
	inaccessibleRepo := newGitHubTestRepo("github.com/sourcegraph/private", newGitHubExternalService(t, esStore))
	if err := repoStore.Create(ctx, repo, inaccessibleRepo); err != nil {
		t.Fatal(err)
	}
	bt.MockRepoPermissions(t, db, userID, repo.ID)

	spec := &btypes.BatchSpec{
		NamespaceUserID: userID,
		UserID:          userID,
	}
	if err := bstore.CreateBatchSpec(ctx, spec); err != nil {
		t.Fatal(err)
	}

	batchChange := &btypes.BatchChange{
		Name:            "my-unique-name",
		NamespaceUserID: userID,
		CreatorID:       userID,
		LastApplierID:   userID,
		LastAppliedAt:   time.Now(),
		BatchSpecID:     spec.ID,
	}
	if err := bstore.CreateBatchChange(ctx, batchChange); err != nil {
		t.Fatal(err)
	}

	changeset1 := bt.CreateChangeset(t, ctx, bstore, bt.TestChangesetOpts{
		Repo:                repo.ID,
		ExternalServiceType: "github",
		PublicationState:    btypes.ChangesetPublicationStateUnpublished,
		ExternalReviewState: btypes.ChangesetReviewStatePending,
		OwnedByBatchChange:  batchChange.ID,
		BatchChange:         batchChange.ID,
	})

	changeset2 := bt.CreateChangeset(t, ctx, bstore, bt.TestChangesetOpts{
		Repo:                repo.ID,
		ExternalServiceType: "github",
		ExternalID:          "12345",
		ExternalBranch:      "open-pr",
		ExternalState:       btypes.ChangesetExternalStateOpen,
		PublicationState:    btypes.ChangesetPublicationStatePublished,
		ExternalReviewState: btypes.ChangesetReviewStatePending,
		OwnedByBatchChange:  batchChange.ID,
		BatchChange:         batchChange.ID,
	})

	changeset3 := bt.CreateChangeset(t, ctx, bstore, bt.TestChangesetOpts{
		Repo:                repo.ID,
		ExternalServiceType: "github",
		ExternalID:          "56789",
		ExternalBranch:      "merged-pr",
		ExternalState:       btypes.ChangesetExternalStateMerged,
		PublicationState:    btypes.ChangesetPublicationStatePublished,
		ExternalReviewState: btypes.ChangesetReviewStatePending,
		OwnedByBatchChange:  batchChange.ID,
		BatchChange:         batchChange.ID,
	})
	changeset4 := bt.CreateChangeset(t, ctx, bstore, bt.TestChangesetOpts{
		Repo:                inaccessibleRepo.ID,
		ExternalServiceType: "github",
		ExternalID:          "987651",
		ExternalBranch:      "open-hidden-pr",
		ExternalState:       btypes.ChangesetExternalStateOpen,
		PublicationState:    btypes.ChangesetPublicationStatePublished,
		ExternalReviewState: btypes.ChangesetReviewStatePending,
		OwnedByBatchChange:  batchChange.ID,
		BatchChange:         batchChange.ID,
	})

	addChangeset(t, ctx, bstore, changeset1, batchChange.ID)
	addChangeset(t, ctx, bstore, changeset2, batchChange.ID)
	addChangeset(t, ctx, bstore, changeset3, batchChange.ID)
	addChangeset(t, ctx, bstore, changeset4, batchChange.ID)

	s, err := newSchema(db, &Resolver{store: bstore})
	if err != nil {
		t.Fatal(err)
	}

	batchChangeAPIID := string(bgql.MarshalBatchChangeID(batchChange.ID))
	nodes := []apitest.Changeset{
		{
			Typename:   "ExternalChangeset",
			ID:         string(bgql.MarshalChangesetID(changeset1.ID)),
			Repository: apitest.Repository{Name: string(repo.Name)},
		},
		{
			Typename:   "ExternalChangeset",
			ID:         string(bgql.MarshalChangesetID(changeset2.ID)),
			Repository: apitest.Repository{Name: string(repo.Name)},
		},
		{
			Typename:   "ExternalChangeset",
			ID:         string(bgql.MarshalChangesetID(changeset3.ID)),
			Repository: apitest.Repository{Name: string(repo.Name)},
		},
		{
			Typename: "HiddenExternalChangeset",
			ID:       string(bgql.MarshalChangesetID(changeset4.ID)),
		},
	}

	tests := []struct {
		firstParam      int
		useUnsafeOpts   bool
		wantHasNextPage bool
		wantEndCursor   string
		wantTotalCount  int
		wantOpen        int
		wantNodes       []apitest.Changeset
	}{
		{firstParam: 1, wantHasNextPage: true, wantEndCursor: "2", wantTotalCount: 4, wantOpen: 2, wantNodes: nodes[:1]},
		{firstParam: 2, wantHasNextPage: true, wantEndCursor: "3", wantTotalCount: 4, wantOpen: 2, wantNodes: nodes[:2]},
		{firstParam: 3, wantHasNextPage: true, wantEndCursor: "4", wantTotalCount: 4, wantOpen: 2, wantNodes: nodes[:3]},
		{firstParam: 4, wantHasNextPage: false, wantTotalCount: 4, wantOpen: 2, wantNodes: nodes[:4]},
		// Expect only 3 changesets to be returned when an unsafe filter is applied.
		{firstParam: 1, useUnsafeOpts: true, wantEndCursor: "2", wantHasNextPage: true, wantTotalCount: 3, wantOpen: 1, wantNodes: nodes[:1]},
		{firstParam: 2, useUnsafeOpts: true, wantEndCursor: "3", wantHasNextPage: true, wantTotalCount: 3, wantOpen: 1, wantNodes: nodes[:2]},
		{firstParam: 3, useUnsafeOpts: true, wantHasNextPage: false, wantTotalCount: 3, wantOpen: 1, wantNodes: nodes[:3]},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("Unsafe opts %t, first %d", tc.useUnsafeOpts, tc.firstParam), func(t *testing.T) {
			input := map[string]any{"batchChange": batchChangeAPIID, "first": int64(tc.firstParam)}
			if tc.useUnsafeOpts {
				input["reviewState"] = btypes.ChangesetReviewStatePending
			}
			var response struct{ Node apitest.BatchChange }
			apitest.MustExec(actor.WithActor(context.Background(), actor.FromUser(userID)), t, s, input, &response, queryChangesetConnection)

			var wantEndCursor *string
			if tc.wantEndCursor != "" {
				wantEndCursor = &tc.wantEndCursor
			}

			wantChangesets := apitest.ChangesetConnection{
				TotalCount: tc.wantTotalCount,
				PageInfo: apitest.PageInfo{
					EndCursor:   wantEndCursor,
					HasNextPage: tc.wantHasNextPage,
				},
				Nodes: tc.wantNodes,
			}

			if diff := cmp.Diff(wantChangesets, response.Node.Changesets); diff != "" {
				t.Fatalf("wrong changesets response (-want +got):\n%s", diff)
			}
		})
	}

	var endCursor *string
	for i := range nodes {
		input := map[string]any{"batchChange": batchChangeAPIID, "first": 1}
		if endCursor != nil {
			input["after"] = *endCursor
		}
		wantHasNextPage := i != len(nodes)-1

		var response struct{ Node apitest.BatchChange }
		apitest.MustExec(actor.WithActor(context.Background(), actor.FromUser(userID)), t, s, input, &response, queryChangesetConnection)

		changesets := response.Node.Changesets
		if diff := cmp.Diff(1, len(changesets.Nodes)); diff != "" {
			t.Fatalf("unexpected number of nodes (-want +got):\n%s", diff)
		}

		if diff := cmp.Diff(len(nodes), changesets.TotalCount); diff != "" {
			t.Fatalf("unexpected total count (-want +got):\n%s", diff)
		}

		if diff := cmp.Diff(wantHasNextPage, changesets.PageInfo.HasNextPage); diff != "" {
			t.Fatalf("unexpected hasNextPage (-want +got):\n%s", diff)
		}

		endCursor = changesets.PageInfo.EndCursor
		if want, have := wantHasNextPage, endCursor != nil; have != want {
			t.Fatalf("unexpected endCursor existence. want=%t, have=%t", want, have)
		}
	}
}

const queryChangesetConnection = `
query($batchChange: ID!, $first: Int, $after: String, $reviewState: ChangesetReviewState){
  node(id: $batchChange) {
    ... on BatchChange {
      changesets(first: $first, after: $after, reviewState: $reviewState) {
        totalCount
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
          endCursor
          hasNextPage
        }
      }
    }
  }
}
`
