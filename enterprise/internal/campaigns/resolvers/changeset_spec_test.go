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
	ct "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/testing"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
)

func TestChangesetSpecResolver(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := backend.WithAuthzBypass(context.Background())
	dbtesting.SetupGlobalTestDB(t)

	userID := insertTestUser(t, dbconn.Global, "changeset-spec-by-id", false)

	reposStore := repos.NewDBStore(dbconn.Global, sql.TxOptions{})
	repo := newGitHubTestRepo("github.com/sourcegraph/sourcegraph", 1)
	if err := reposStore.UpsertRepos(ctx, repo); err != nil {
		t.Fatal(err)
	}

	repoID := graphqlbackend.MarshalRepositoryID(repo.ID)
	spec := &campaigns.ChangesetSpec{
		RawSpec: ct.NewRawChangesetSpec(repoID),
		Spec: campaigns.ChangesetSpecFields{
			RepoID: repoID,
		},
		RepoID: repo.ID,
		UserID: userID,
	}

	store := ee.NewStore(dbconn.Global)
	if err := store.CreateChangesetSpec(ctx, spec); err != nil {
		t.Fatal(err)
	}

	s, err := graphqlbackend.NewSchema(&Resolver{store: store}, nil, nil)
	if err != nil {
		t.Fatal(err)

	}

	apiID := string(marshalChangesetSpecRandID(spec.RandID))

	want := apitest.ChangesetSpec{
		Typename: "ChangesetSpec",
		ID:       apiID,
		ExpiresAt: &graphqlbackend.DateTime{
			Time: spec.CreatedAt.Truncate(time.Second).Add(2 * time.Hour),
		},
	}

	input := map[string]interface{}{"changesetSpec": apiID}
	var response struct{ Node apitest.ChangesetSpec }
	apitest.MustExec(ctx, t, s, input, &response, queryChangesetSpecNode)

	if diff := cmp.Diff(want, response.Node); diff != "" {
		t.Fatalf("unexpected response (-want +got):\n%s", diff)
	}
}

const queryChangesetSpecNode = `
query($changesetSpec: ID!) {
  node(id: $changesetSpec) {
    __typename

    ... on ChangesetSpec {
      id

      expiresAt
    }
  }
}
`
