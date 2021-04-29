package resolvers

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/resolvers/apitest"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	ct "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/testing"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
)

func TestBulkJobConnectionResolver(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := backend.WithAuthzBypass(context.Background())
	db := dbtesting.GetDB(t)

	userID := ct.CreateTestUser(t, db, true).ID
	now := timeutil.Now()
	clock := func() time.Time { return now }
	cstore := store.NewWithClock(db, clock)

	batchSpec := ct.CreateBatchSpec(t, ctx, cstore, "test", userID)
	batchChange := ct.CreateBatchChange(t, ctx, cstore, "test", userID, batchSpec.ID)
	batchChangeAPIID := marshalBatchChangeID(batchChange.ID)
	repos, _ := ct.CreateTestRepos(t, ctx, db, 3)
	changeset := ct.CreateChangeset(t, ctx, cstore, ct.TestChangesetOpts{
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
	if err := cstore.CreateChangesetJob(ctx, jobs...); err != nil {
		t.Fatal(err)
	}

	s, err := graphqlbackend.NewSchema(db, New(cstore), nil, nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	nodes := []apitest.BulkJob{
		{
			ID:        string(marshalBulkJobID("group-1")),
			Type:      "COMMENT",
			State:     string(btypes.BulkJobStateProcessing),
			Errors:    []*apitest.ChangesetJobError{},
			CreatedAt: marshalDateTime(t, now),
		},
		{
			ID:        string(marshalBulkJobID("group-2")),
			Type:      "COMMENT",
			State:     string(btypes.BulkJobStateProcessing),
			Errors:    []*apitest.ChangesetJobError{},
			CreatedAt: marshalDateTime(t, now),
		},
		{
			ID:        string(marshalBulkJobID("group-3")),
			Type:      "COMMENT",
			State:     string(btypes.BulkJobStateProcessing),
			Errors:    []*apitest.ChangesetJobError{},
			CreatedAt: marshalDateTime(t, now),
		},
	}

	tests := []struct {
		firstParam      int
		wantHasNextPage bool
		wantEndCursor   string
		wantTotalCount  int
		wantNodes       []apitest.BulkJob
	}{
		{firstParam: 1, wantHasNextPage: true, wantEndCursor: "2", wantTotalCount: 3, wantNodes: nodes[:1]},
		{firstParam: 2, wantHasNextPage: true, wantEndCursor: "3", wantTotalCount: 3, wantNodes: nodes[:2]},
		{firstParam: 3, wantHasNextPage: false, wantTotalCount: 3, wantNodes: nodes[:3]},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("First %d", tc.firstParam), func(t *testing.T) {
			input := map[string]interface{}{"batchChange": batchChangeAPIID, "first": int64(tc.firstParam)}
			var response struct {
				Node apitest.BatchChange
			}
			apitest.MustExec(actor.WithActor(context.Background(), actor.FromUser(userID)), t, s, input, &response, queryBulkJobConnection)

			var wantEndCursor *string
			if tc.wantEndCursor != "" {
				wantEndCursor = &tc.wantEndCursor
			}

			wantBulkJobs := apitest.BulkJobConnection{
				TotalCount: tc.wantTotalCount,
				PageInfo: apitest.PageInfo{
					EndCursor:   wantEndCursor,
					HasNextPage: tc.wantHasNextPage,
				},
				Nodes: tc.wantNodes,
			}

			if diff := cmp.Diff(wantBulkJobs, response.Node.BulkJobs); diff != "" {
				t.Fatalf("wrong bulk jobs response (-want +got):\n%s", diff)
			}
		})
	}

	var endCursor *string
	for i := range nodes {
		input := map[string]interface{}{"batchChange": batchChangeAPIID, "first": 1}
		if endCursor != nil {
			input["after"] = *endCursor
		}
		wantHasNextPage := i != len(nodes)-1

		var response struct {
			Node apitest.BatchChange
		}
		apitest.MustExec(actor.WithActor(context.Background(), actor.FromUser(userID)), t, s, input, &response, queryBulkJobConnection)

		bulkJobs := response.Node.BulkJobs
		if diff := cmp.Diff(1, len(bulkJobs.Nodes)); diff != "" {
			t.Fatalf("unexpected number of nodes (-want +got):\n%s", diff)
		}

		if diff := cmp.Diff(len(nodes), bulkJobs.TotalCount); diff != "" {
			t.Fatalf("unexpected total count (-want +got):\n%s", diff)
		}

		if diff := cmp.Diff(wantHasNextPage, bulkJobs.PageInfo.HasNextPage); diff != "" {
			t.Fatalf("unexpected hasNextPage (-want +got):\n%s", diff)
		}

		endCursor = bulkJobs.PageInfo.EndCursor
		if want, have := wantHasNextPage, endCursor != nil; have != want {
			t.Fatalf("unexpected endCursor existence. want=%t, have=%t", want, have)
		}
	}
}

const queryBulkJobConnection = `
query($batchChange: ID!, $first: Int, $after: String){
    node(id: $batchChange) {
        ... on BatchChange {
            bulkJobs(first: $first, after: $after) {
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
