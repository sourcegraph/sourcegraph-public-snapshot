package graphqlbackend

import (
	"testing"

	"github.com/graph-gophers/graphql-go/gqltesting"
	"sourcegraph.com/cmd/frontend/db"
	"sourcegraph.com/cmd/frontend/internal/pkg/usagestats"
	"sourcegraph.com/cmd/frontend/types"
)

func TestUser_UsageStatistics(t *testing.T) {
	resetMocks()
	db.Mocks.Users.MockGetByID_Return(t, &types.User{ID: 1, Username: "alice"}, nil)
	usagestats.MockGetByUserID = func(userID int32) (*types.UserUsageStatistics, error) {
		return &types.UserUsageStatistics{
			SearchQueries: 2,
		}, nil
	}
	defer func() { usagestats.MockGetByUserID = nil }()
	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Schema: GraphQLSchema,
			Query: `
				{
					node(id: "VXNlcjox") {
						id
						... on User {
							usageStatistics {
								searchQueries
							}
						}
					}
				}
			`,
			ExpectedResult: `
				{
					"node": {
						"id": "VXNlcjox",
						"usageStatistics": {
							"searchQueries": 2
						}
					}
				}
			`,
		},
	})
}
