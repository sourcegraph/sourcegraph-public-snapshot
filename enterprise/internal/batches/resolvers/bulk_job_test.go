package resolvers

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/resolvers/apitest"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	ct "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/testing"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
)

func TestBulkJobResolver(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()
	db := dbtesting.GetDB(t)

	userID := ct.CreateTestUser(t, db, false).ID

	now := timeutil.Now()
	clock := func() time.Time { return now }
	cstore := store.NewWithClock(db, clock)

	batchSpec := ct.CreateBatchSpec(t, ctx, cstore, "test", userID)
	batchChange := ct.CreateBatchChange(t, ctx, cstore, "test", userID, batchSpec.ID)
	repos, _ := ct.CreateTestRepos(t, ctx, db, 3)
	changeset1 := ct.CreateChangeset(t, ctx, cstore, ct.TestChangesetOpts{
		Repo:             repos[0].ID,
		BatchChange:      batchChange.ID,
		PublicationState: btypes.ChangesetPublicationStatePublished,
	})
	changeset2 := ct.CreateChangeset(t, ctx, cstore, ct.TestChangesetOpts{
		Repo:             repos[1].ID,
		BatchChange:      batchChange.ID,
		PublicationState: btypes.ChangesetPublicationStatePublished,
	})
	changeset3 := ct.CreateChangeset(t, ctx, cstore, ct.TestChangesetOpts{
		Repo:             repos[2].ID,
		BatchChange:      batchChange.ID,
		PublicationState: btypes.ChangesetPublicationStatePublished,
	})
	ct.MockRepoPermissions(t, db, userID, repos[0].ID, repos[1].ID)

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
	if err := cstore.CreateChangesetJob(ctx, jobs...); err != nil {
		t.Fatal(err)
	}

	s, err := graphqlbackend.NewSchema(db, New(cstore), nil, nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	bulkJobAPIID := string(marshalBulkJobID(bulkGroupID))
	wantBatchChange := apitest.BulkJob{
		ID:       bulkJobAPIID,
		Type:     "COMMENT",
		State:    string(btypes.BulkJobStateProcessing),
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

	input := map[string]interface{}{"bulkJob": bulkJobAPIID}
	var response struct{ Node apitest.BulkJob }
	apitest.MustExec(actor.WithActor(ctx, actor.FromUser(userID)), t, s, input, &response, queryBulkJob)

	if diff := cmp.Diff(wantBatchChange, response.Node); diff != "" {
		t.Fatalf("wrong bulk job response (-want +got):\n%s", diff)
	}
}

const queryBulkJob = `
query($bulkJob: ID!){
  node(id: $bulkJob) {
    ... on BulkJob {
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
