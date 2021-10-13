package resolvers

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/batches/resolvers/apitest"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	ct "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/testing"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/batches"
)

func TestBatchSpecWorkspaceResolver(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := actor.WithInternalActor(context.Background())
	db := dbtest.NewDB(t, "")

	bstore := store.New(db, &observation.TestContext, nil)
	repo, _ := ct.CreateTestRepo(t, ctx, db)

	repoID := graphqlbackend.MarshalRepositoryID(repo.ID)

	userID := ct.CreateTestUser(t, db, true).ID

	spec := &btypes.BatchSpec{UserID: userID, NamespaceUserID: userID}
	if err := bstore.CreateBatchSpec(ctx, spec); err != nil {
		t.Fatal(err)
	}
	specID := marshalBatchSpecRandID(spec.RandID)

	workspace := &btypes.BatchSpecWorkspace{
		ID:               0,
		BatchSpecID:      spec.ID,
		ChangesetSpecIDs: []int64{},
		RepoID:           repo.ID,
		Branch:           "refs/heads/main",
		Commit:           "d34db33f",
		Path:             "a/b/c",
		Steps: []batches.Step{
			{
				Run:       "echo 'hello world'",
				Container: "alpine:3",
			},
		},
		FileMatches:        []string{"a/b/c.go"},
		OnlyFetchWorkspace: false,
		Unsupported:        true,
		Ignored:            true,
	}
	if err := bstore.CreateBatchSpecWorkspace(ctx, workspace); err != nil {
		t.Fatal(err)
	}
	apiID := marshalBatchSpecWorkspaceID(workspace.ID)

	s, err := graphqlbackend.NewSchema(db, &Resolver{store: bstore}, nil, nil, nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	want := apitest.BatchSpecWorkspace{
		Typename: "BatchSpecWorkspace",
		ID:       string(apiID),

		Repository: apitest.Repository{
			Name: string(repo.Name),
			ID:   string(repoID),
		},
		BatchSpec: apitest.BatchSpec{
			ID: string(specID),
		},

		State: "PENDING",

		Unsupported: true,
		Ignored:     true,

		Steps: []apitest.BatchSpecWorkspaceStep{
			{
				Run:       workspace.Steps[0].Run,
				Container: workspace.Steps[0].Container,
			},
		},
	}

	input := map[string]interface{}{"batchSpecWorkspace": apiID}

	var response struct{ Node apitest.BatchSpecWorkspace }
	apitest.MustExec(actor.WithActor(context.Background(), actor.FromUser(userID)), t, s, input, &response, queryBatchSpecWorkspaceNode)

	if diff := cmp.Diff(want, response.Node); diff != "" {
		t.Fatalf("unexpected response (-want +got):\n%s", diff)
	}
}

const queryBatchSpecWorkspaceNode = `
query($batchSpecWorkspace: ID!) {
  node(id: $batchSpecWorkspace) {
    __typename

    ... on BatchSpecWorkspace {
      id

      repository {
        id
        name
      }
      batchSpec {
        id
      }

      state
      unsupported
      ignored

      steps {
	    run
	    container
      }
    }
  }
}
`
