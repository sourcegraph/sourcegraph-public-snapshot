package resolvers

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/batches/resolvers/apitest"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/batches/store"
	bt "github.com/sourcegraph/sourcegraph/internal/batches/testing"
	btypes "github.com/sourcegraph/sourcegraph/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/batches"
)

func TestBatchSpecWorkspaceResolver(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)

	ctx := actor.WithInternalActor(context.Background())
	db := database.NewDB(logger, dbtest.NewDB(t))

	bstore := store.New(db, observation.TestContextTB(t), nil)
	repo, _ := bt.CreateTestRepo(t, ctx, db)

	repoID := graphqlbackend.MarshalRepositoryID(repo.ID)

	userID := bt.CreateTestUser(t, db, true).ID
	adminCtx := actor.WithActor(context.Background(), actor.FromUser(userID))

	spec := &btypes.BatchSpec{
		UserID:          userID,
		NamespaceUserID: userID,
		Spec: &batches.BatchSpec{
			Steps: []batches.Step{
				{
					Run:       "echo 'hello world'",
					Container: "alpine:3",
				},
			},
		},
	}
	if err := bstore.CreateBatchSpec(ctx, spec); err != nil {
		t.Fatal(err)
	}
	specID := marshalBatchSpecRandID(spec.RandID)

	testRev := api.CommitID("b69072d5f687b31b9f6ae3ceafdc24c259c4b9ec")
	mockBackendCommits(t, testRev)

	workspace := &btypes.BatchSpecWorkspace{
		ID:                 0,
		BatchSpecID:        spec.ID,
		ChangesetSpecIDs:   []int64{},
		RepoID:             repo.ID,
		Branch:             "refs/heads/main",
		Commit:             string(testRev),
		Path:               "a/b/c",
		FileMatches:        []string{"a/b/c.go"},
		OnlyFetchWorkspace: false,
		Unsupported:        true,
		Ignored:            true,
	}

	if err := bstore.CreateBatchSpecWorkspace(ctx, workspace); err != nil {
		t.Fatal(err)
	}
	apiID := string(marshalBatchSpecWorkspaceID(workspace.ID))

	s, err := newSchema(db, &Resolver{store: bstore})
	if err != nil {
		t.Fatal(err)
	}

	wantTmpl := apitest.BatchSpecWorkspace{
		Typename: "VisibleBatchSpecWorkspace",
		ID:       apiID,

		Repository: apitest.Repository{
			Name: string(repo.Name),
			ID:   string(repoID),
		},
		BatchSpec: apitest.BatchSpec{
			ID: string(specID),
		},

		SearchResultPaths: []string{
			"a/b/c.go",
		},
		Branch: apitest.GitRef{
			DisplayName: "main",
			Target:      apitest.GitTarget{OID: string(testRev)},
		},
		Path: "a/b/c",

		OnlyFetchWorkspace: false,
		Unsupported:        true,
		Ignored:            true,

		Steps: []apitest.BatchSpecWorkspaceStep{
			{
				Run:       spec.Spec.Steps[0].Run,
				Container: spec.Spec.Steps[0].Container,
			},
		},
	}

	t.Run("Pending", func(t *testing.T) {
		want := wantTmpl

		want.State = "PENDING"

		queryAndAssertBatchSpecWorkspace(t, adminCtx, s, apiID, want)
	})
	t.Run("Queued", func(t *testing.T) {
		job := &btypes.BatchSpecWorkspaceExecutionJob{
			BatchSpecWorkspaceID: workspace.ID,
			UserID:               userID,
		}
		if err := bt.CreateBatchSpecWorkspaceExecutionJob(ctx, bstore, store.ScanBatchSpecWorkspaceExecutionJob, job); err != nil {
			t.Fatal(err)
		}

		want := wantTmpl
		want.State = "QUEUED"
		want.PlaceInQueue = 1

		queryAndAssertBatchSpecWorkspace(t, adminCtx, s, apiID, want)
	})
}

func queryAndAssertBatchSpecWorkspace(t *testing.T, ctx context.Context, s *graphql.Schema, id string, want apitest.BatchSpecWorkspace) {
	t.Helper()

	input := map[string]any{"batchSpecWorkspace": id}

	var response struct{ Node apitest.BatchSpecWorkspace }

	apitest.MustExec(ctx, t, s, input, &response, queryBatchSpecWorkspaceNode)

	if diff := cmp.Diff(want, response.Node); diff != "" {
		t.Fatalf("unexpected batch spec workspace (-want +got):\n%s", diff)
	}
}

const queryBatchSpecWorkspaceNode = `
query($batchSpecWorkspace: ID!) {
  node(id: $batchSpecWorkspace) {
    __typename

    ... on BatchSpecWorkspace {
      id

      batchSpec {
        id
      }

      onlyFetchWorkspace
      unsupported
      ignored

      state
      placeInQueue
    }
    ... on VisibleBatchSpecWorkspace {
      repository {
        id
        name
      }

      searchResultPaths
      branch {
        displayName
        target {
          oid
        }
      }

      path

      steps {
        run
        container
      }
    }
  }
}
`
