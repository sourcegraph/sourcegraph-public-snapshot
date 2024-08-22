package graphqlbackend

import (
	"testing"

	mockrequire "github.com/derision-test/go-mockgen/v2/testutil/require"

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

	securityLogEvents := dbmocks.NewMockSecurityEventLogsStore()
	securityLogEvents.LogSecurityEventFunc.SetDefaultReturn(nil)
	db.SecurityEventLogsFunc.SetDefaultReturn(securityLogEvents)

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
	mockrequire.Called(t, securityLogEvents.LogSecurityEventFunc)
}
