package graphqlbackend

import (
	"testing"

	mockrequire "github.com/derision-test/go-mockgen/testutil/require"

	"github.com/sourcegraph/sourcegraph/internal/database/dbmock"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestCreateUser(t *testing.T) {
	users := dbmock.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)
	users.CreateFunc.SetDefaultReturn(&types.User{ID: 1, Username: "alice"}, nil)

	authz := dbmock.NewMockAuthzStore()
	authz.GrantPendingPermissionsFunc.SetDefaultReturn(nil)

	db := dbmock.NewMockDB()
	db.UsersFunc.SetDefaultReturn(users)
	db.AuthzFunc.SetDefaultReturn(authz)

	RunTests(t, []*Test{
		{
			Schema: mustParseGraphQLSchema(t, db),
			Query: `
				mutation {
					createUser(username: "alice") {
						user {
							id
						}
					}
				}
			`,
			ExpectedResult: `
				{
					"createUser": {
						"user": {
							"id": "VXNlcjox"
						}
					}
				}
			`,
		},
	})

	mockrequire.Called(t, authz.GrantPendingPermissionsFunc)
}
