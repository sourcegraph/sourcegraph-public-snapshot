package graphqlbackend

import (
	"context"
	"testing"

	"github.com/graph-gophers/graphql-go/gqltesting"
	"sourcegraph.com/cmd/frontend/backend"
	"sourcegraph.com/cmd/frontend/db"
	"sourcegraph.com/cmd/frontend/types"
	"sourcegraph.com/pkg/repoupdater"
	"sourcegraph.com/pkg/repoupdater/protocol"
)

func TestStatusMessages(t *testing.T) {
	resetMocks()
	t.Run("unauthenticated", func(t *testing.T) {
		result, err := (&schemaResolver{}).StatusMessages(context.Background())
		if want := backend.ErrNotAuthenticated; err != want {
			t.Errorf("got err %v, want %v", err, want)
		}
		if result != nil {
			t.Errorf("got result %v, want nil", result)
		}
	})

	t.Run("authenticated as non-site-admin", func(t *testing.T) {
		db.Mocks.Users.GetByCurrentAuthUser = func(ctx context.Context) (*types.User, error) {
			return &types.User{ID: 1, SiteAdmin: false}, nil
		}
		defer func() { db.Mocks.Users.GetByCurrentAuthUser = nil }()

		result, err := (&schemaResolver{}).StatusMessages(context.Background())
		if want := backend.ErrMustBeSiteAdmin; err != want {
			t.Errorf("got err %v, want %v", err, want)
		}
		if result != nil {
			t.Errorf("got result %v, want nil", result)
		}
	})

	t.Run("no messages", func(t *testing.T) {
		db.Mocks.Users.GetByCurrentAuthUser = func(ctx context.Context) (*types.User, error) {
			return &types.User{ID: 1, SiteAdmin: true}, nil
		}
		defer func() { db.Mocks.Users.GetByCurrentAuthUser = nil }()

		repoupdater.MockStatusMessages = func(_ context.Context) (*protocol.StatusMessagesResponse, error) {
			res := &protocol.StatusMessagesResponse{Messages: []protocol.StatusMessage{}}
			return res, nil
		}
		defer func() { repoupdater.MockStatusMessages = nil }()

		gqltesting.RunTests(t, []*gqltesting.Test{
			{
				Schema: GraphQLSchema,
				Query: `
				query {
					statusMessages {
					    type
						message
					}
				}
			`,
				ExpectedResult: `
				{
					"statusMessages": []
				}
			`,
			},
		})
	})

	t.Run("messages", func(t *testing.T) {
		db.Mocks.Users.GetByCurrentAuthUser = func(ctx context.Context) (*types.User, error) {
			return &types.User{ID: 1, SiteAdmin: true}, nil
		}
		defer func() { db.Mocks.Users.GetByCurrentAuthUser = nil }()

		repoupdater.MockStatusMessages = func(_ context.Context) (*protocol.StatusMessagesResponse, error) {
			res := &protocol.StatusMessagesResponse{Messages: []protocol.StatusMessage{
				{
					Type:    protocol.CloningStatusMessage,
					Message: "Currently cloning 5 repositories in parallel...",
				},
			}}
			return res, nil
		}
		defer func() { repoupdater.MockStatusMessages = nil }()

		gqltesting.RunTests(t, []*gqltesting.Test{
			{
				Schema: GraphQLSchema,
				Query: `
				query {
					statusMessages {
					    type
						message
					}
				}
			`,
				ExpectedResult: `
				{
					"statusMessages": [
					{
						"type": "CLONING",
						"message": "Currently cloning 5 repositories in parallel..."
					}
					]
				}
			`,
			},
		})
	})
}
