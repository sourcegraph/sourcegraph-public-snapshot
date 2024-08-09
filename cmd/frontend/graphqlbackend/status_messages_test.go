package graphqlbackend

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestStatusMessages(t *testing.T) {
	graphqlQuery := `
		query StatusMessages {
			statusMessages {
				__typename

				... on GitUpdatesDisabled {
					message
				}

				... on NoRepositoriesDetected {
					message
				}

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

	db := dbmocks.NewMockDB()
	t.Run("unauthenticated", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(nil, nil)
		db.UsersFunc.SetDefaultReturn(users)

		result, err := newSchemaResolver(db, gitserver.NewTestClient(t), nil).StatusMessages(context.Background())
		if want := auth.ErrNotAuthenticated; err != want {
			t.Errorf("got err %v, want %v", err, want)
		}
		if result != nil {
			t.Errorf("got result %v, want nil", result)
		}
	})

	t.Run("no messages", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 1, SiteAdmin: true}, nil)
		db.UsersFunc.SetDefaultReturn(users)

		repos.MockStatusMessages = func(_ context.Context) ([]repos.StatusMessage, error) {
			return []repos.StatusMessage{}, nil
		}
		t.Cleanup(func() {
			repos.MockStatusMessages = nil
		})

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
		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 1, SiteAdmin: true}, nil)

		externalServices := dbmocks.NewMockExternalServiceStore()
		externalServices.GetByIDFunc.SetDefaultReturn(&types.ExternalService{
			ID:          1,
			DisplayName: "GitHub.com testing",
			Config:      extsvc.NewEmptyConfig(),
		}, nil)

		db.UsersFunc.SetDefaultReturn(users)
		db.ExternalServicesFunc.SetDefaultReturn(externalServices)

		repos.MockStatusMessages = func(_ context.Context) ([]repos.StatusMessage, error) {
			res := []repos.StatusMessage{
				{
					GitUpdatesDisabled: &repos.GitUpdatesDisabled{
						Message: "Repositories will not be cloned or updated.",
					},
				},
				{
					NoRepositoriesDetected: &repos.NoRepositoriesDetected{
						Message: "No repositories have been added to Sourcegraph.",
					},
				},
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

		conf.Mock(&conf.Unified{
			SiteConfiguration: schema.SiteConfiguration{
				DisableAutoGitUpdates: true,
			},
		})

		t.Cleanup(func() {
			repos.MockStatusMessages = nil
			conf.Mock(nil)
		})

		RunTest(t, &Test{
			Schema: mustParseGraphQLSchema(t, db),
			Query:  graphqlQuery,
			ExpectedResult: `
					{
						"statusMessages": [
							{
								"__typename": "GitUpdatesDisabled",
        						"message": "Repositories will not be cloned or updated."
							},
							{
								"__typename": "NoRepositoriesDetected",
        						"message": "No repositories have been added to Sourcegraph."
							},
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
		})
	})
}
