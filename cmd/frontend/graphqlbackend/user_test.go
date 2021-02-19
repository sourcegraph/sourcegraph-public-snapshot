package graphqlbackend

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"

	gqlerrors "github.com/graph-gophers/graphql-go/errors"
	"github.com/graph-gophers/graphql-go/gqltesting"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestUser(t *testing.T) {
	t.Run("by username", func(t *testing.T) {
		checkUserByUsername := func(t *testing.T) {
			t.Helper()
			gqltesting.RunTests(t, []*gqltesting.Test{
				{
					Schema: mustParseGraphQLSchema(t),
					Query: `
				{
					user(username: "alice") {
						username
					}
				}
			`,
					ExpectedResult: `
				{
					"user": {
						"username": "alice"
					}
				}
			`,
				},
			})
		}

		resetMocks()
		database.Mocks.Users.GetByUsername = func(_ context.Context, username string) (*types.User, error) {
			if want := "alice"; username != want {
				t.Errorf("got %q, want %q", username, want)
			}
			return &types.User{ID: 1, Username: "alice"}, nil
		}

		t.Run("allowed on Sourcegraph.com", func(t *testing.T) {
			orig := envvar.SourcegraphDotComMode()
			envvar.MockSourcegraphDotComMode(true)
			defer envvar.MockSourcegraphDotComMode(orig) // reset

			checkUserByUsername(t)
		})

		t.Run("allowed on non-Sourcegraph.com", func(t *testing.T) {
			checkUserByUsername(t)
		})
	})

	t.Run("by email", func(t *testing.T) {
		resetMocks()
		database.Mocks.Users.GetByVerifiedEmail = func(_ context.Context, email string) (*types.User, error) {
			if want := "alice@example.com"; email != want {
				t.Errorf("got %q, want %q", email, want)
			}
			return &types.User{ID: 1, Username: "alice"}, nil
		}

		t.Run("disallowed on Sourcegraph.com", func(t *testing.T) {
			checkUserByEmailError := func(t *testing.T, wantErr string) {
				t.Helper()
				gqltesting.RunTests(t, []*gqltesting.Test{
					{
						Schema: mustParseGraphQLSchema(t),
						Query: `
				{
					user(email: "alice@example.com") {
						username
					}
				}
			`,
						ExpectedResult: `{"user": null}`,
						ExpectedErrors: []*gqlerrors.QueryError{{Message: wantErr, Path: []interface{}{"user"}, ResolverError: errors.New(wantErr)}},
					},
				})
			}

			orig := envvar.SourcegraphDotComMode()
			envvar.MockSourcegraphDotComMode(true)
			defer envvar.MockSourcegraphDotComMode(orig) // reset

			t.Run("for anonymous viewer", func(t *testing.T) {
				database.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*types.User, error) {
					return nil, database.ErrNoCurrentUser
				}
				checkUserByEmailError(t, "not authenticated")
			})
			t.Run("for non-site-admin viewer", func(t *testing.T) {
				database.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*types.User, error) {
					return &types.User{SiteAdmin: false}, nil
				}
				checkUserByEmailError(t, "must be site admin")
			})
		})

		t.Run("allowed on non-Sourcegraph.com", func(t *testing.T) {
			gqltesting.RunTests(t, []*gqltesting.Test{
				{
					Schema: mustParseGraphQLSchema(t),
					Query: `
				{
					user(email: "alice@example.com") {
						username
					}
				}
			`,
					ExpectedResult: `
				{
					"user": {
						"username": "alice"
					}
				}
			`,
				},
			})
		})
	})
}

func TestNode_User(t *testing.T) {
	resetMocks()
	database.Mocks.Users.MockGetByID_Return(t, &types.User{ID: 1, Username: "alice"}, nil)

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Schema: mustParseGraphQLSchema(t),
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

func TestUpdateUser(t *testing.T) {
	db := new(dbtesting.MockDB)

	t.Run("not site admin nor the same user", func(t *testing.T) {
		database.Mocks.Users.GetByID = func(ctx context.Context, id int32) (*types.User, error) {
			return &types.User{ID: id, Username: strconv.Itoa(int(id))}, nil
		}
		database.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*types.User, error) {
			return &types.User{ID: 2, Username: "2", SiteAdmin: false}, nil
		}
		t.Cleanup(func() {
			database.Mocks.Users = database.MockUsers{}
		})

		result, err := (&schemaResolver{db: db}).UpdateUser(context.Background(), &updateUserArgs{User: "VXNlcjox"})
		wantErr := "must be authenticated as 1 or as an admin (must be site admin)"
		gotErr := fmt.Sprintf("%v", err)
		if wantErr != gotErr {
			t.Fatalf("err: want %q but got %q", wantErr, gotErr)
		}
		if result != nil {
			t.Fatalf("result: want nil but got %v", result)
		}
	})

	t.Run("disallow suspicious names", func(t *testing.T) {
		oldSourcegraphDotComMode := envvar.SourcegraphDotComMode()
		envvar.MockSourcegraphDotComMode(true)
		database.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*types.User, error) {
			return &types.User{SiteAdmin: true}, nil
		}
		t.Cleanup(func() {
			envvar.MockSourcegraphDotComMode(oldSourcegraphDotComMode)
			database.Mocks.Users = database.MockUsers{}
		})

		result, err := (&schemaResolver{db: db}).UpdateUser(context.Background(), &updateUserArgs{
			User:     "VXNlcjox",
			Username: strptr("about"),
		})
		wantErr := `rejected suspicious name "about"`
		gotErr := fmt.Sprintf("%v", err)
		if wantErr != gotErr {
			t.Fatalf("err: want %q but got %q", wantErr, gotErr)
		}
		if result != nil {
			t.Fatalf("result: want nil but got %v", result)
		}
	})

	t.Run("non site admin cannot change username when not enabled", func(t *testing.T) {
		conf.Mock(&conf.Unified{
			SiteConfiguration: schema.SiteConfiguration{
				AuthEnableUsernameChanges: false,
			},
		})
		database.Mocks.Users.GetByID = func(ctx context.Context, id int32) (*types.User, error) {
			return &types.User{ID: id}, nil
		}
		database.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*types.User, error) {
			return &types.User{ID: 1, SiteAdmin: false}, nil
		}
		t.Cleanup(func() {
			conf.Mock(nil)
			database.Mocks.Users = database.MockUsers{}
		})

		ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
		result, err := (&schemaResolver{db: db}).UpdateUser(ctx, &updateUserArgs{
			User:     "VXNlcjox",
			Username: strptr("alice"),
		})
		wantErr := "unable to change username because auth.enableUsernameChanges is false in site configuration"
		gotErr := fmt.Sprintf("%v", err)
		if wantErr != gotErr {
			t.Fatalf("err: want %q but got %q", wantErr, gotErr)
		}
		if result != nil {
			t.Fatalf("result: want nil but got %v", result)
		}
	})

	t.Run("non site admin can change non-username fields", func(t *testing.T) {
		conf.Mock(&conf.Unified{
			SiteConfiguration: schema.SiteConfiguration{
				AuthEnableUsernameChanges: false,
			},
		})
		database.Mocks.Users.GetByID = func(ctx context.Context, id int32) (*types.User, error) {
			return &types.User{ID: 1, Username: "alice", DisplayName: "alice-updated", AvatarURL: "http://www.example.com/alice-updated", SiteAdmin: false}, nil
		}
		database.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*types.User, error) {
			return &types.User{ID: 1, Username: "alice", DisplayName: "alice-updated", AvatarURL: "http://www.example.com/alice-updated", SiteAdmin: false}, nil
		}
		database.Mocks.Users.Update = func(userID int32, update database.UserUpdate) error {
			return nil
		}
		t.Cleanup(func() {
			database.Mocks.Users = database.MockUsers{}
		})

		gqltesting.RunTests(t, []*gqltesting.Test{
			{
				Context: actor.WithActor(context.Background(), &actor.Actor{UID: 1}),
				Schema:  mustParseGraphQLSchema(t),
				Query: `
			mutation {
				updateUser(
					user: "VXNlcjox",
					displayName: "alice-updated"
					avatarURL: "http://www.example.com/alice-updated"
				) {
					displayName,
					avatarURL
				}
			}
		`,
				ExpectedResult: `
			{
				"updateUser": {
					"displayName": "alice-updated",
					"avatarURL": "http://www.example.com/alice-updated"
				}
			}
		`,
			},
		})
	})

	t.Run("success", func(t *testing.T) {
		database.Mocks.Users.GetByID = func(ctx context.Context, id int32) (*types.User, error) {
			return &types.User{ID: id, Username: strconv.Itoa(int(id))}, nil
		}
		database.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*types.User, error) {
			return &types.User{SiteAdmin: true}, nil
		}
		database.Mocks.Users.Update = func(userID int32, update database.UserUpdate) error {
			return nil
		}
		t.Cleanup(func() {
			database.Mocks.Users = database.MockUsers{}
		})

		gqltesting.RunTests(t, []*gqltesting.Test{
			{
				Schema: mustParseGraphQLSchema(t),
				Query: `
			mutation {
				updateUser(
					user: "VXNlcjox",
					username: "alice.bob-chris-"
				) {
					username
				}
			}
		`,
				ExpectedResult: `
			{
				"updateUser": {
					"username": "1"
				}
			}
		`,
			},
		})
	})
}
