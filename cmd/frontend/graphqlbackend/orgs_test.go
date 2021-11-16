package graphqlbackend

import (
	"context"
	"testing"

	"github.com/graph-gophers/graphql-go/errors"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmock"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestOrgs(t *testing.T) {
	users := dbmock.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)

	orgs := dbmock.NewMockOrgStore()
	orgs.ListFunc.SetDefaultReturn([]*types.Org{{Name: "org1"}, {Name: "org2"}}, nil)
	orgs.CountFunc.SetDefaultReturn(2, nil)

	db := dbmock.NewMockDB()
	db.UsersFunc.SetDefaultReturn(users)
	db.OrgsFunc.SetDefaultReturn(orgs)

	RunTests(t, []*Test{
		{
			Schema: mustParseGraphQLSchema(t, db),
			Query: `
				{
					organizations {
						nodes { name }
						totalCount
					}
				}
			`,
			ExpectedResult: `
				{
					"organizations": {
						"nodes": [
							{
								"name": "org1"
							},
							{
								"name": "org2"
							}
						],
						"totalCount": 2
					}
				}
			`,
		},
	})
}

func TestListOrgsForCloud(t *testing.T) {
	db := database.NewDB(nil)
	resetMocks()
	envvar.MockSourcegraphDotComMode(true)
	database.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*types.User, error) {
		return &types.User{SiteAdmin: true}, nil
	}

	t.Cleanup(func() {
		resetMocks()
		envvar.MockSourcegraphDotComMode(false)
	})
	RunTests(t, []*Test{
		{
			Schema: mustParseGraphQLSchema(t, db),
			Query: `
				{
					organizations {
						nodes { name }
						totalCount
					}
				}
			`,
			ExpectedResult: "null",
			ExpectedErrors: []*errors.QueryError{
				{
					Message: "listing organizations is not allowed",
					Path:    []interface{}{string("organizations"), string("nodes")},
				},
				{
					Message: "counting organizations is not allowed",
					Path:    []interface{}{string("organizations"), string("totalCount")},
				},
			},
		},
	})
}
