package graphqlbackend

import (
	"context"
	"testing"

	"github.com/graph-gophers/graphql-go/gqltesting"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestStatusMessages(t *testing.T) {
	db := new(dbtesting.MockDB)

	graphqlQuery := `
		query StatusMessages {
			statusMessages {
				__typename

				... on CloningProgress {
					message
				}

				... on SyncError {
					message
				}

				... on ExternalServiceSyncError {
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
		result, err := (&schemaResolver{db: db}).StatusMessages(context.Background())
		if want := backend.ErrNotAuthenticated; err != want {
			t.Errorf("got err %v, want %v", err, want)
		}
		if result != nil {
			t.Errorf("got result %v, want nil", result)
		}
	})

	t.Run("no messages", func(t *testing.T) {
		database.Mocks.Users.GetByCurrentAuthUser = func(ctx context.Context) (*types.User, error) {
			return &types.User{ID: 1, SiteAdmin: true}, nil
		}
		defer func() { database.Mocks.Users.GetByCurrentAuthUser = nil }()

		repos.MockStatusMessages = func(_ context.Context, _ *types.User) ([]repos.StatusMessage, error) {
			return []repos.StatusMessage{}, nil
		}
		defer func() { repos.MockStatusMessages = nil }()

		gqltesting.RunTests(t, []*gqltesting.Test{
			{
				Schema: mustParseGraphQLSchema(t),
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
		database.Mocks.Users.GetByCurrentAuthUser = func(ctx context.Context) (*types.User, error) {
			return &types.User{ID: 1, SiteAdmin: true}, nil
		}
		defer func() { database.Mocks.Users.GetByCurrentAuthUser = nil }()

		database.Mocks.ExternalServices.GetByID = func(id int64) (*types.ExternalService, error) {
			return &types.ExternalService{ID: 1, DisplayName: "GitHub.com testing"}, nil
		}
		defer func() { database.Mocks.ExternalServices.GetByID = nil }()

		repos.MockStatusMessages = func(_ context.Context, _ *types.User) ([]repos.StatusMessage, error) {
			res := []repos.StatusMessage{
				{
					Cloning: &repos.CloningProgress{
						Message: "Currently cloning 5 repositories in parallel...",
					},
				},
				{
					ExternalServiceSyncError: &repos.ExternalServiceSyncError{
						Message:           "Authentication failed. Please check credentials.",
						ExternalServiceId: 1,
					},
				},
				{
					SyncError: &repos.SyncError{
						Message: "Could not save to database",
					},
				},
			}
			return res, nil
		}
		defer func() { repos.MockStatusMessages = nil }()

		gqltesting.RunTests(t, []*gqltesting.Test{
			{
				Schema: mustParseGraphQLSchema(t),
				Query:  graphqlQuery,
				ExpectedResult: `
					{
						"statusMessages": [
							{
								"__typename": "CloningProgress",
								"message": "Currently cloning 5 repositories in parallel..."
							},
							{
								"__typename": "ExternalServiceSyncError",
								"externalService": {
									"displayName": "GitHub.com testing",
									"id": "RXh0ZXJuYWxTZXJ2aWNlOjE="
								},
								"message": "Authentication failed. Please check credentials."
							},
							{
								"__typename": "SyncError",
								"message": "Could not save to database"
							}
						]
					}
				`,
			},
		})
	})
}
