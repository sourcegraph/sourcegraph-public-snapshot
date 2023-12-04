package graphqlbackend

import (
	"testing"

	"github.com/graph-gophers/graphql-go/errors"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestOrgs(t *testing.T) {
	users := dbmocks.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)

	orgs := dbmocks.NewMockOrgStore()
	orgs.ListFunc.SetDefaultReturn([]*types.Org{{Name: "org1"}, {Name: "org2"}}, nil)
	orgs.CountFunc.SetDefaultReturn(2, nil)

	db := dbmocks.NewMockDB()
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
	orig := envvar.SourcegraphDotComMode()
	envvar.MockSourcegraphDotComMode(true)
	defer envvar.MockSourcegraphDotComMode(orig)

	users := dbmocks.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)

	orgs := dbmocks.NewMockOrgStore()
	orgs.CountFunc.SetDefaultReturn(42, nil)

	db := dbmocks.NewMockDB()
	db.UsersFunc.SetDefaultReturn(users)
	db.OrgsFunc.SetDefaultReturn(orgs)

	RunTests(t, []*Test{
		{
			Schema: mustParseGraphQLSchema(t, db),
			Query: `
				{
					organizations {
						nodes { name },
						totalCount
					}
				}
			`,
			ExpectedResult: "null",
			ExpectedErrors: []*errors.QueryError{
				{
					Message: "listing organizations is not allowed",
					Path:    []any{"organizations", "nodes"},
				},
			},
		},
		{
			Schema: mustParseGraphQLSchema(t, db),
			Query: `
				{
					organizations {
						totalCount
					}
				}
			`,
			ExpectedResult: `
				{
					"organizations": {
						"totalCount": 42
					}
				}
			`,
		},
	})
}
