package resolvers

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/batches/resolvers/apitest"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	ct "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/testing"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestBatchSpecExecutionConnectionResolver(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := actor.WithInternalActor(context.Background())
	db := dbtest.NewDB(t, "")

	userID := ct.CreateTestUser(t, db, true).ID

	cstore := store.New(db, &observation.TestContext, nil)
	repoStore := database.ReposWith(cstore)
	esStore := database.ExternalServicesWith(cstore)

	repo := newGitHubTestRepo("github.com/sourcegraph/batch-spec-execution-connection-test", newGitHubExternalService(t, esStore))
	if err := repoStore.Create(ctx, repo); err != nil {
		t.Fatal(err)
	}

	exec1 := &btypes.BatchSpecExecution{
		NamespaceUserID: userID,
		UserID:          userID,
		State:           btypes.BatchSpecExecutionStateProcessing,
		BatchSpec:       `name: testing`,
	}
	if err := cstore.CreateBatchSpecExecution(ctx, exec1); err != nil {
		t.Fatal(err)
	}
	exec2 := &btypes.BatchSpecExecution{
		NamespaceUserID: userID,
		UserID:          userID,
		State:           btypes.BatchSpecExecutionStateQueued,
		BatchSpec:       `name: testing-2`,
	}
	if err := cstore.CreateBatchSpecExecution(ctx, exec2); err != nil {
		t.Fatal(err)
	}

	s, err := graphqlbackend.NewSchema(db, &Resolver{store: cstore}, nil, nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Batch spec executions are returned in reverse order.
	nodes := []apitest.BatchSpecExecution{
		{
			ID: string(marshalBatchSpecExecutionRandID(exec2.RandID)),
		},
		{
			ID: string(marshalBatchSpecExecutionRandID(exec1.RandID)),
		},
	}

	tests := []struct {
		firstParam      int
		wantHasNextPage bool
		wantTotalCount  int
		wantNodes       []apitest.BatchSpecExecution
	}{
		{firstParam: 1, wantHasNextPage: true, wantTotalCount: 2, wantNodes: nodes[:1]},
		{firstParam: 2, wantHasNextPage: false, wantTotalCount: 2, wantNodes: nodes},
		{firstParam: 3, wantHasNextPage: false, wantTotalCount: 2, wantNodes: nodes},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("first=%d", tc.firstParam), func(t *testing.T) {
			input := map[string]interface{}{"first": int64(tc.firstParam)}
			var response struct {
				BatchSpecExecutions apitest.BatchSpecExecutionConnection
			}
			apitest.MustExec(actor.WithActor(context.Background(), actor.FromUser(userID)), t, s, input, &response, queryBatchSpecExecutionsConnection)

			wantConnection := apitest.BatchSpecExecutionConnection{
				TotalCount: tc.wantTotalCount,
				PageInfo: apitest.PageInfo{
					HasNextPage: tc.wantHasNextPage,
					// We don't test on the cursor here.
					EndCursor: response.BatchSpecExecutions.PageInfo.EndCursor,
				},
				Nodes: tc.wantNodes,
			}

			if diff := cmp.Diff(wantConnection, response.BatchSpecExecutions); diff != "" {
				t.Fatalf("wrong batchSpecExecutions response (-want +got):\n%s", diff)
			}
		})
	}

	t.Run("Cursor based pagination", func(t *testing.T) {
		var endCursor *string
		for i := range nodes {
			input := map[string]interface{}{"first": 1}
			if endCursor != nil {
				input["after"] = *endCursor
			}
			wantHasNextPage := i != len(nodes)-1

			var response struct {
				BatchSpecExecutions apitest.BatchSpecExecutionConnection
			}
			apitest.MustExec(ctx, t, s, input, &response, queryBatchSpecExecutionsConnection)

			if diff := cmp.Diff(1, len(response.BatchSpecExecutions.Nodes)); diff != "" {
				t.Fatalf("unexpected number of nodes (-want +got):\n%s", diff)
			}

			if diff := cmp.Diff(len(nodes), response.BatchSpecExecutions.TotalCount); diff != "" {
				t.Fatalf("unexpected total count (-want +got):\n%s", diff)
			}

			if diff := cmp.Diff(wantHasNextPage, response.BatchSpecExecutions.PageInfo.HasNextPage); diff != "" {
				t.Fatalf("unexpected hasNextPage (-want +got):\n%s", diff)
			}

			endCursor = response.BatchSpecExecutions.PageInfo.EndCursor
			if want, have := wantHasNextPage, endCursor != nil; have != want {
				t.Fatalf("unexpected endCursor existence. want=%t, have=%t", want, have)
			}
		}
	})
}

const queryBatchSpecExecutionsConnection = `
query($first: Int, $after: String) {
  batchSpecExecutions(first: $first, after: $after) {
    totalCount
    pageInfo {
	  hasNextPage
	  endCursor
    }
    nodes {
      id
    }
  }
}
`
