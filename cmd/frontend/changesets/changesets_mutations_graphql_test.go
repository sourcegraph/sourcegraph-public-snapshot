// +build ignore TODO!(sqs)

package changesets

import (
	"context"
	"reflect"
	"testing"

	"github.com/graph-gophers/graphql-go/gqltesting"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

func TestGraphQL_CreateChangeset(t *testing.T) {
	resetMocks()
	const wantOrgID = 1
	wantChangeset := &dbChangeset{
		NamespaceOrgID: wantOrgID,
		title:          "n",
	}
	mocks.changesets.Create = func(changeset *dbChangeset) (*dbChangeset, error) {
		if !reflect.DeepEqual(changeset, wantChangeset) {
			t.Errorf("got changeset %+v, want %+v", changeset, wantChangeset)
		}
		tmp := *changeset
		tmp.ID = 2
		return &tmp, nil
	}

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Context: backend.WithAuthzBypass(context.Background()),
			Schema:  graphqlbackend.GraphQLSchema,
			Query: `
				mutation {
					changesets {
						createChangeset(input: { namespace: "T3JnOjE=", name: "n" }) {
							id
							name
						}
					}
				}
			`,
			ExpectedResult: `
				{
					"changesets": {
						"createChangeset": {
							"id": "UHJvamVjdDoy",
							"name": "n"
						}
					}
				}
			`,
		},
	})
}
