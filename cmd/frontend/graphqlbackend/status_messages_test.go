package graphqlbackend

import (
	"context"
	"testing"

	"github.com/graph-gophers/graphql-go/gqltesting"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/repoupdater"
	"github.com/sourcegraph/sourcegraph/pkg/repoupdater/protocol"
)

func TestStatusMessages(t *testing.T) {
	graphqlQuery := `
		query StatusMessages {
			statusMessages {
				__typename

				... on CloningStatusMessage {
					message
				}

				... on SyncErrorStatusMessage {
					message
					externalService {
						id
						displayName
					}
				}
			}
		}
	`

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
				Query:  graphqlQuery,
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

		db.Mocks.ExternalServices.GetByID = func(id int64) (*types.ExternalService, error) {
			return &types.ExternalService{ID: 1, DisplayName: "GitHub.com testing"}, nil
		}
		defer func() { db.Mocks.ExternalServices.GetByID = nil }()

		repoupdater.MockStatusMessages = func(_ context.Context) (*protocol.StatusMessagesResponse, error) {
			res := &protocol.StatusMessagesResponse{Messages: []protocol.StatusMessage{
				{
					Cloning: &protocol.CloningStatusMessage{
						Message: "Currently cloning 5 repositories in parallel...",
					},
				},
				{
					SyncError: &protocol.SyncErrorStatusMessage{
						Message:           "Authentication failed. Please check credentials.",
						ExternalServiceId: 1,
					},
				},
			}}
			return res, nil
		}
		defer func() { repoupdater.MockStatusMessages = nil }()

		gqltesting.RunTests(t, []*gqltesting.Test{
			{
				Schema: GraphQLSchema,
				Query:  graphqlQuery,
				ExpectedResult: `
					{
						"statusMessages": [
							{
								"__typename": "CloningStatusMessage",
								"message": "Currently cloning 5 repositories in parallel..."
							},
							{
								"__typename": "SyncErrorStatusMessage",
								"externalService": {
									"displayName": "GitHub.com testing",
									"id": "RXh0ZXJuYWxTZXJ2aWNlOjE="
								},
								"message": "Authentication failed. Please check credentials."
							}
						]
					}
				`,
			},
		})
	})
}
