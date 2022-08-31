package graphqlbackend

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestStatusMessages(t *testing.T) {
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

	db := database.NewMockDB()
	t.Run("unauthenticated", func(t *testing.T) {
		users := database.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(nil, nil)
		db.UsersFunc.SetDefaultReturn(users)

		result, err := newSchemaResolver(db).StatusMessages(context.Background())
		if want := backend.ErrNotAuthenticated; err != want {
			t.Errorf("got err %v, want %v", err, want)
		}
		if result != nil {
			t.Errorf("got result %v, want nil", result)
		}
	})

	t.Run("no messages", func(t *testing.T) {
		users := database.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 1, SiteAdmin: true}, nil)
		db.UsersFunc.SetDefaultReturn(users)

		repos.MockStatusMessages = func(_ context.Context, _ *types.User) ([]repos.StatusMessage, error) {
			return []repos.StatusMessage{}, nil
		}
		defer func() { repos.MockStatusMessages = nil }()

		RunTests(t, []*Test{
			{
				Schema: mustParseGraphQLSchema(t, db),
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
		users := database.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 1, SiteAdmin: true}, nil)

		externalServices := database.NewMockExternalServiceStore()
		externalServices.GetByIDFunc.SetDefaultReturn(&types.ExternalService{
			ID:          1,
			DisplayName: "GitHub.com testing",
			Config:      extsvc.NewEmptyConfig(),
		}, nil)

		db.UsersFunc.SetDefaultReturn(users)
		db.ExternalServicesFunc.SetDefaultReturn(externalServices)

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

		RunTests(t, []*Test{
			{
				Schema: mustParseGraphQLSchema(t, db),
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
