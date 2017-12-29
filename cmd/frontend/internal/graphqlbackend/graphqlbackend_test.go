package graphqlbackend

import (
	"testing"

	"github.com/neelance/graphql-go/gqltesting"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/db"
)

func TestRepository(t *testing.T) {
	resetMocks()
	db.Mocks.Repos.MockGetByURI(t, "github.com/gorilla/mux", 2)
	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Schema: GraphQLSchema,
			Query: `
				{
					repository(uri: "github.com/gorilla/mux") {
						uri
					}
				}
			`,
			ExpectedResult: `
				{
					"repository": {
						"uri": "github.com/gorilla/mux"
					}
				}
			`,
		},
	})
}
