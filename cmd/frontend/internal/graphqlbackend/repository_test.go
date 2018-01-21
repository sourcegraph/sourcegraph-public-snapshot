package graphqlbackend

import (
	"context"
	"testing"

	"github.com/neelance/graphql-go/gqltesting"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

const exampleCommitSHA1 = "1234567890123456789012345678901234567890"

func TestRepository_Commit(t *testing.T) {
	resetMocks()
	db.Mocks.Repos.MockGetByURI(t, "github.com/gorilla/mux", 2)
	backend.Mocks.Repos.ResolveRev = func(ctx context.Context, repo int32, rev string) (vcs.CommitID, error) {
		if repo != 2 || rev != "abc" {
			t.Error("wrong arguments to ResolveRev")
		}
		return exampleCommitSHA1, nil
	}
	backend.Mocks.Repos.MockGetCommit_Return_NoCheck(t, &vcs.Commit{ID: exampleCommitSHA1})

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Schema: GraphQLSchema,
			Query: `
				{
					repository(uri: "github.com/gorilla/mux") {
						commit(rev: "abc") {
							oid
						}
					}
				}
			`,
			ExpectedResult: `
				{
					"repository": {
						"commit": {
							"oid": "` + exampleCommitSHA1 + `"
						}
					}
				}
			`,
		},
	})
}
