package graphqlbackend

import (
	"context"
	"testing"

	"github.com/neelance/graphql-go/gqltesting"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	store "sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
)

func TestNode_User(t *testing.T) {
	t.Skip("TODO(sqs): unskip when the migration described in commit 2ac372aa2773080dc3d077beb056e9513e64bf67 is executed")

	resetMocks()
	store.Mocks.Users.MockGetByID_Return(t, &sourcegraph.User{ID: 1, Username: "alice"}, nil)

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Schema: GraphQLSchema,
			Query: `
				{
					node(id: "VXNlcjox") {
						id
						... on User {
							username
						}
					}
				}
			`,
			ExpectedResult: `
				{
					"node": {
						"id": "VXNlcjox",
						"username": "alice"
					}
				}
			`,
		},
	})
}

func TestUsers_Activity(t *testing.T) {
	ctx := context.Background()
	store.Mocks.Users.MockGetByAuth0ID_Return(t, &sourcegraph.User{}, nil)
	u := &userResolver{user: &sourcegraph.User{}, actor: actor.FromContext(ctx)}
	_, err := u.Activity(ctx)
	if err == nil {
		t.Errorf("Non-admin can access endpoint")
	}
}
