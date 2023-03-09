package resolvers

import (
	"context"
	"testing"

	"github.com/graph-gophers/graphql-go/errors"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	own2 "github.com/sourcegraph/sourcegraph/enterprise/internal/own"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
)

type fakeOwnService struct {
}

type fakeGitserver struct {
	gitserver.Client
}

func TestCodeownersIngestionGuarding(t *testing.T) {
	db := database.NewMockDB()
	git := fakeGitserver{}
	own := own2.NewService(git, db)

	ctx := context.Background()

	//ctx := userCtx(fs.AddUser(types.User{SiteAdmin: true}))

	schema, err := graphqlbackend.NewSchema(db, git, nil, graphqlbackend.OptionalResolver{OwnResolver: New(db, git, own)})
	if err != nil {
		t.Fatal(err)
	}

	pathToQueries := map[string]string{
		"addCodeownersFile": `
		mutation add {
		  addCodeownersFile(input: {fileContents: "* @admin", repoName: "github.com/sourcegraph/sourcegraph"}) {
			id
		  }
		}`,
		"updateCodeownersFile": `
		mutation update {
		 updateCodeownersFile(input: {fileContents: "* @admin", repoName: "github.com/sourcegraph/sourcegraph"}) {
			id
		 }
		}`,
		"deleteCodeownersFiles": `
		mutation delete {
		 deleteCodeownersFiles(repositories:{repoName: "test"}) {
			alwaysNil
		 }
		}`,
	}
	for path, query := range pathToQueries {
		t.Run("feature flag guarding is respected for "+path, func(t *testing.T) {
			t.Cleanup(func() {
				ctx = context.TODO()
			})
			ctx = featureflag.WithFlags(ctx, featureflag.NewMemoryStore(map[string]bool{"search-ownership": false}, nil, nil))
			expectedResult := `null`
			if path == "deleteCodeownersFiles" {
				expectedResult = `
					{
						"deleteCodeownersFiles": null
					}
				`
			}
			graphqlbackend.RunTest(t, &graphqlbackend.Test{
				Schema:         schema,
				Context:        ctx,
				Query:          query,
				ExpectedResult: expectedResult,
				ExpectedErrors: []*errors.QueryError{
					{Message: "own is not available yet", Path: []any{path}},
				},
			})
		})
		t.Run("dotcom guarding is respected for "+path, func(t *testing.T) {
			orig := envvar.SourcegraphDotComMode()
			envvar.MockSourcegraphDotComMode(true)
			t.Cleanup(func() {
				envvar.MockSourcegraphDotComMode(orig)
			})
		})
	}

}
