package resolvers

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
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

func TestBulkOperationConnectionResolver(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	ctx := actor.WithInternalActor(context.Background())
	db := database.NewDB(logger, dbtest.NewDB(logger, t))

	userID := bt.CreateTestUser(t, db, true).ID
	now := timeutil.Now()
	clock := func() time.Time { return now }
	bstore := store.NewWithClock(db, &observation.TestContext, nil, clock)

	batchSpec := bt.CreateBatchSpec(t, ctx, bstore, "test", userID, 0)
	batchChange := bt.CreateBatchChange(t, ctx, bstore, "test", userID, batchSpec.ID)
	batchChangeAPIID := marshalBatchChangeID(batchChange.ID)
	repos, _ := bt.CreateTestRepos(t, ctx, db, 3)
	changeset := bt.CreateChangeset(t, ctx, bstore, bt.TestChangesetOpts{
		Repo:             repos[0].ID,
		BatchChange:      batchChange.ID,
		PublicationState: btypes.ChangesetPublicationStatePublished,
	})
	jobs := []*btypes.ChangesetJob{
		{
			BulkGroup:     "group-1",
			UserID:        userID,
			BatchChangeID: batchChange.ID,
			ChangesetID:   changeset.ID,
			JobType:       btypes.ChangesetJobTypeComment,
			Payload:       btypes.ChangesetJobCommentPayload{Message: "test"},
			State:         btypes.ChangesetJobStateQueued,
			StartedAt:     now,
			FinishedAt:    now,
		},
		{
			BulkGroup:     "group-2",
			UserID:        userID,
			BatchChangeID: batchChange.ID,
			ChangesetID:   changeset.ID,
			JobType:       btypes.ChangesetJobTypeComment,
			Payload:       btypes.ChangesetJobCommentPayload{Message: "test"},
			State:         btypes.ChangesetJobStateQueued,
			StartedAt:     now,
			FinishedAt:    now,
		},
		{
			BulkGroup:     "group-3",
			UserID:        userID,
			BatchChangeID: batchChange.ID,
			ChangesetID:   changeset.ID,
			JobType:       btypes.ChangesetJobTypeComment,
			Payload:       btypes.ChangesetJobCommentPayload{Message: "test"},
			State:         btypes.ChangesetJobStateQueued,
			StartedAt:     now,
			FinishedAt:    now,
		},
	}
	if err := bstore.CreateChangesetJob(ctx, jobs...); err != nil {
		t.Fatal(err)
	}

	s, err := graphqlbackend.NewSchema(db, New(bstore), nil, nil, nil, nil, nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	nodes := []apitest.BulkOperation{
		{
			ID:        string(marshalBulkOperationID("group-3")),
			Type:      "COMMENT",
			State:     string(btypes.BulkOperationStateProcessing),
			Errors:    []*apitest.ChangesetJobError{},
			CreatedAt: marshalDateTime(t, now),
		},
		{
			ID:        string(marshalBulkOperationID("group-2")),
			Type:      "COMMENT",
			State:     string(btypes.BulkOperationStateProcessing),
			Errors:    []*apitest.ChangesetJobError{},
			CreatedAt: marshalDateTime(t, now),
		},
		{
			ID:        string(marshalBulkOperationID("group-1")),
			Type:      "COMMENT",
			State:     string(btypes.BulkOperationStateProcessing),
			Errors:    []*apitest.ChangesetJobError{},
			CreatedAt: marshalDateTime(t, now),
		},
	}

	tests := []struct {
		firstParam      int
		wantHasNextPage bool
		wantEndCursor   string
		wantTotalCount  int
		wantNodes       []apitest.BulkOperation
	}{
		{firstParam: 1, wantHasNextPage: true, wantEndCursor: "2", wantTotalCount: 3, wantNodes: nodes[:1]},
		{firstParam: 2, wantHasNextPage: true, wantEndCursor: "1", wantTotalCount: 3, wantNodes: nodes[:2]},
		{firstParam: 3, wantHasNextPage: false, wantTotalCount: 3, wantNodes: nodes[:3]},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("First %d", tc.firstParam), func(t *testing.T) {
			input := map[string]any{"batchChange": batchChangeAPIID, "first": int64(tc.firstParam)}
			var response struct {
				Node apitest.BatchChange
			}
			apitest.MustExec(actor.WithActor(context.Background(), actor.FromUser(userID)), t, s, input, &response, queryBulkOperationConnection)

			var wantEndCursor *string
			if tc.wantEndCursor != "" {
				wantEndCursor = &tc.wantEndCursor
			}

			wantBulkOperations := apitest.BulkOperationConnection{
				TotalCount: tc.wantTotalCount,
				PageInfo: apitest.PageInfo{
					EndCursor:   wantEndCursor,
					HasNextPage: tc.wantHasNextPage,
				},
				Nodes: tc.wantNodes,
			}

			if diff := cmp.Diff(wantBulkOperations, response.Node.BulkOperations); diff != "" {
				t.Fatalf("wrong bulk operations response (-want +got):\n%s", diff)
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

		var response struct {
			Node apitest.BatchChange
		}
		apitest.MustExec(actor.WithActor(context.Background(), actor.FromUser(userID)), t, s, input, &response, queryBulkOperationConnection)

		bulkOperations := response.Node.BulkOperations
		if diff := cmp.Diff(1, len(bulkOperations.Nodes)); diff != "" {
			t.Fatalf("unexpected number of nodes (-want +got):\n%s", diff)
		}

		if diff := cmp.Diff(len(nodes), bulkOperations.TotalCount); diff != "" {
			t.Fatalf("unexpected total count (-want +got):\n%s", diff)
		}

		if diff := cmp.Diff(wantHasNextPage, bulkOperations.PageInfo.HasNextPage); diff != "" {
			t.Fatalf("unexpected hasNextPage (-want +got):\n%s", diff)
		}

		endCursor = bulkOperations.PageInfo.EndCursor
		if want, have := wantHasNextPage, endCursor != nil; have != want {
			t.Fatalf("unexpected endCursor existence. want=%t, have=%t", want, have)
		}
	}
}

const queryBulkOperationConnection = `
query($batchChange: ID!, $first: Int, $after: String){
    node(id: $batchChange) {
        ... on BatchChange {
            bulkOperations(first: $first, after: $after) {
                totalCount
                pageInfo {
                    endCursor
                    hasNextPage
                }
                nodes {
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
    }
}
`
