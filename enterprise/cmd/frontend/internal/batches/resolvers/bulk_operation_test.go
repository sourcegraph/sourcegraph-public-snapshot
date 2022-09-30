package resolvers

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/batches/resolvers/apitest"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	bt "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/testing"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
)

func TestBulkOperationResolver(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))

	userID := bt.CreateTestUser(t, db, false).ID

	now := timeutil.Now()
	clock := func() time.Time { return now }
	bstore := store.NewWithClock(db, &observation.TestContext, nil, clock)

	batchSpec := bt.CreateBatchSpec(t, ctx, bstore, "test", userID, 0)
	batchChange := bt.CreateBatchChange(t, ctx, bstore, "test", userID, batchSpec.ID)
	repos, _ := bt.CreateTestRepos(t, ctx, db, 3)
	changeset1 := bt.CreateChangeset(t, ctx, bstore, bt.TestChangesetOpts{
		Repo:             repos[0].ID,
		BatchChange:      batchChange.ID,
		PublicationState: btypes.ChangesetPublicationStatePublished,
	})
	changeset2 := bt.CreateChangeset(t, ctx, bstore, bt.TestChangesetOpts{
		Repo:             repos[1].ID,
		BatchChange:      batchChange.ID,
		PublicationState: btypes.ChangesetPublicationStatePublished,
	})
	changeset3 := bt.CreateChangeset(t, ctx, bstore, bt.TestChangesetOpts{
		Repo:             repos[2].ID,
		BatchChange:      batchChange.ID,
		PublicationState: btypes.ChangesetPublicationStatePublished,
	})
	bt.MockRepoPermissions(t, db, userID, repos[0].ID, repos[1].ID)

	bulkGroupID := "test-group"
	errorMsg := "Very bad error."

	jobs := []*btypes.ChangesetJob{
		// Accessible and failed.
		{
			BulkGroup:      bulkGroupID,
			UserID:         userID,
			BatchChangeID:  batchChange.ID,
			ChangesetID:    changeset1.ID,
			JobType:        btypes.ChangesetJobTypeComment,
			Payload:        btypes.ChangesetJobCommentPayload{Message: "test"},
			State:          btypes.ChangesetJobStateFailed,
			FailureMessage: strPtr(errorMsg),
			StartedAt:      now,
			FinishedAt:     now,
		},
		// Accessible and successful.
		{
			BulkGroup:     bulkGroupID,
			UserID:        userID,
			BatchChangeID: batchChange.ID,
			ChangesetID:   changeset2.ID,
			JobType:       btypes.ChangesetJobTypeComment,
			Payload:       btypes.ChangesetJobCommentPayload{Message: "test"},
			State:         btypes.ChangesetJobStateQueued,
			StartedAt:     now,
		},
		// Not accessible and failed.
		{
			BulkGroup:      bulkGroupID,
			UserID:         userID,
			BatchChangeID:  batchChange.ID,
			ChangesetID:    changeset3.ID,
			JobType:        btypes.ChangesetJobTypeComment,
			Payload:        btypes.ChangesetJobCommentPayload{Message: "test"},
			State:          btypes.ChangesetJobStateFailed,
			FailureMessage: strPtr(errorMsg),
			StartedAt:      now,
			FinishedAt:     now,
		},
	}
	if err := bstore.CreateChangesetJob(ctx, jobs...); err != nil {
		t.Fatal(err)
	}

	s, err := newSchema(db, &Resolver{store: bstore})
	if err != nil {
		t.Fatal(err)
	}

	bulkOperationAPIID := string(marshalBulkOperationID(bulkGroupID))
	wantBatchChange := apitest.BulkOperation{
		ID:       bulkOperationAPIID,
		Type:     "COMMENT",
		State:    string(btypes.BulkOperationStateProcessing),
		Progress: 2.0 / 3.0,
		Errors: []*apitest.ChangesetJobError{
			{
				Changeset: &apitest.Changeset{ID: string(marshalChangesetID(changeset1.ID))},
				Error:     strPtr(errorMsg),
			},
			{
				Changeset: &apitest.Changeset{ID: string(marshalChangesetID(changeset3.ID))},
				// Error should not be exposed.
				Error: nil,
			},
		},
		CreatedAt: marshalDateTime(t, now),
		// Not finished.
		FinishedAt: "",
	}

	input := map[string]any{"bulkOperation": bulkOperationAPIID}
	var response struct{ Node apitest.BulkOperation }
	apitest.MustExec(actor.WithActor(ctx, actor.FromUser(userID)), t, s, input, &response, queryBulkOperation)

	if diff := cmp.Diff(wantBatchChange, response.Node); diff != "" {
		t.Fatalf("wrong bulk operation response (-want +got):\n%s", diff)
	}
}

const queryBulkOperation = `
query($bulkOperation: ID!){
  node(id: $bulkOperation) {
    ... on BulkOperation {
      id
      type
      state
      progress
      errors {
          changeset {
              id
          }
          error
      }
      createdAt
      finishedAt
    }
  }
}
`
