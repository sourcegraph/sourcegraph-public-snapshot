package graphqlbackend

import (
	"testing"

	"github.com/neelance/graphql-go/gqltesting"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
)

func TestRepository(t *testing.T) {
	resetMocks()
	localstore.Mocks.Repos.MockGetByURI(t, "github.com/gorilla/mux", 2)
	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Schema: GraphQLSchema,
			Query: `
				{
					root {
						repository(uri: "github.com/gorilla/mux") {
							uri
						}
					}
				}
			`,
			ExpectedResult: `
				{
					"root": {
						"repository": {
							"uri": "github.com/gorilla/mux"
						}
					}
				}
			`,
		},
	})
}
