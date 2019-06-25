// +build ignore TODO!(sqs)

package changesets

import (
	"context"
	"reflect"
	"testing"

	"github.com/graph-gophers/graphql-go/gqltesting"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

func TestGraphQL_CreateChangeset(t *testing.T) {
	resetMocks()
	const wantOrgID = 1
	db.Mocks.Orgs.GetByID = func(context.Context, int32) (*types.Org, error) {
		return &types.Org{ID: wantOrgID}, nil
	}
	wantChangeset := &dbChangeset{
		NamespaceOrgID: wantOrgID,
		Name:           "n",
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
