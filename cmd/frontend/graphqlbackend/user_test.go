package graphqlbackend

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/cockroachdb/errors"
	gqlerrors "github.com/graph-gophers/graphql-go/errors"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmock"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestUser(t *testing.T) {
	db := database.NewDB(nil)
	t.Run("by username", func(t *testing.T) {
		checkUserByUsername := func(t *testing.T) {
			t.Helper()
			RunTests(t, []*Test{
				{
					Schema: mustParseGraphQLSchema(t, db),
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
				RunTests(t, []*Test{
					{
						Schema: mustParseGraphQLSchema(t, db),
						Query: `
				{
					user(email: "alice@example.com") {
						username
					}
				}
			`,
						ExpectedResult: `{"user": null}`,
						ExpectedErrors: []*gqlerrors.QueryError{
							{
								Path:          []interface{}{"user"},
								Message:       wantErr,
								ResolverError: errors.New(wantErr),
							},
						},
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
			RunTests(t, []*Test{
				{
					Schema: mustParseGraphQLSchema(t, db),
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

func TestUser_Email(t *testing.T) {
	db := dbmock.NewMockDB()
	t.Run("only allowed by authenticated user on Sourcegraph.com", func(t *testing.T) {
		orig := envvar.SourcegraphDotComMode()
		envvar.MockSourcegraphDotComMode(true)
		defer envvar.MockSourcegraphDotComMode(orig) // reset

		_, err := NewUserResolver(db, &types.User{ID: 1}).Email(context.Background())
		got := fmt.Sprintf("%v", err)
		want := "must be authenticated as user with id 1"
		assert.Equal(t, want, got)
	})
}

func TestUser_LatestSettings(t *testing.T) {
	db := dbmock.NewMockDB()
	t.Run("only allowed by authenticated user on Sourcegraph.com", func(t *testing.T) {
		users := dbmock.NewMockUserStore()
		db.UsersFunc.SetDefaultReturn(users)

		orig := envvar.SourcegraphDotComMode()
		envvar.MockSourcegraphDotComMode(true)
		defer envvar.MockSourcegraphDotComMode(orig) // reset

		tests := []struct {
			name  string
			ctx   context.Context
			setup func()
		}{
			{
				name: "unauthenticated",
				ctx:  context.Background(),
				setup: func() {
					users.GetByIDFunc.SetDefaultReturn(&types.User{ID: 1}, nil)
				},
			},
			{
				name: "another user",
				ctx:  actor.WithActor(context.Background(), &actor.Actor{UID: 2}),
				setup: func() {
					users.GetByIDFunc.SetDefaultHook(func(ctx context.Context, id int32) (*types.User, error) {
						return &types.User{ID: id}, nil
					})
				},
			},
			{
				name: "site admin",
				ctx:  actor.WithActor(context.Background(), &actor.Actor{UID: 2}),
				setup: func() {
					users.GetByIDFunc.SetDefaultHook(func(ctx context.Context, id int32) (*types.User, error) {
						return &types.User{ID: id, SiteAdmin: true}, nil
					})
				},
			},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				test.setup()

				_, err := NewUserResolver(db, &types.User{ID: 1}).LatestSettings(test.ctx)
				got := fmt.Sprintf("%v", err)
				want := "must be authenticated as user with id 1"
				assert.Equal(t, want, got)
			})
		}
	})
}

func TestUser_ViewerCanAdminister(t *testing.T) {
	db := dbmock.NewMockDB()
	t.Run("only allowed by authenticated user on Sourcegraph.com", func(t *testing.T) {
		users := dbmock.NewMockUserStore()
		db.UsersFunc.SetDefaultReturn(users)

		orig := envvar.SourcegraphDotComMode()
		envvar.MockSourcegraphDotComMode(true)
		defer envvar.MockSourcegraphDotComMode(orig) // reset

		tests := []struct {
			name  string
			ctx   context.Context
			setup func()
		}{
			{
				name: "unauthenticated",
				ctx:  context.Background(),
				setup: func() {
					users.GetByIDFunc.SetDefaultReturn(&types.User{ID: 1}, nil)
				},
			},
			{
				name: "another user",
				ctx:  actor.WithActor(context.Background(), &actor.Actor{UID: 2}),
				setup: func() {
					users.GetByIDFunc.SetDefaultHook(func(ctx context.Context, id int32) (*types.User, error) {
						return &types.User{ID: id}, nil
					})
				},
			},
			{
				name: "site admin",
				ctx:  actor.WithActor(context.Background(), &actor.Actor{UID: 2}),
				setup: func() {
					users.GetByIDFunc.SetDefaultHook(func(ctx context.Context, id int32) (*types.User, error) {
						return &types.User{ID: id, SiteAdmin: true}, nil
					})
				},
			},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				test.setup()

				ok, _ := NewUserResolver(db, &types.User{ID: 1}).ViewerCanAdminister(test.ctx)
				assert.False(t, ok, "ViewerCanAdminister")
			})
		}
	})
}

func TestNode_User(t *testing.T) {
	resetMocks()
	database.Mocks.Users.MockGetByID_Return(t, &types.User{ID: 1, Username: "alice"}, nil)
	db := database.NewDB(nil)

	RunTests(t, []*Test{
		{
			Schema: mustParseGraphQLSchema(t, db),
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
	db := database.NewDB(nil)

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

		result, err := (&schemaResolver{db: database.NewDB(db)}).UpdateUser(context.Background(), &updateUserArgs{User: "VXNlcjox"})
		wantErr := "must be authenticated as the authorized user or as an admin (must be site admin)"
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
			return &types.User{ID: 1}, nil
		}
		t.Cleanup(func() {
			envvar.MockSourcegraphDotComMode(oldSourcegraphDotComMode)
			database.Mocks.Users = database.MockUsers{}
		})

		ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
		_, err := newSchemaResolver(db).UpdateUser(ctx,
			&updateUserArgs{
				User:     MarshalUserID(1),
				Username: strptr("about"),
			},
		)
		got := fmt.Sprintf("%v", err)
		want := `rejected suspicious name "about"`
		assert.Equal(t, want, got)
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
		result, err := (&schemaResolver{db: database.NewDB(db)}).UpdateUser(ctx, &updateUserArgs{
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

		RunTests(t, []*Test{
			{
				Context: actor.WithActor(context.Background(), &actor.Actor{UID: 1}),
				Schema:  mustParseGraphQLSchema(t, db),
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

	t.Run("only allowed by authenticated user on Sourcegraph.com", func(t *testing.T) {
		db := dbmock.NewMockDB()
		users := dbmock.NewMockUserStore()
		db.UsersFunc.SetDefaultReturn(users)

		orig := envvar.SourcegraphDotComMode()
		envvar.MockSourcegraphDotComMode(true)
		defer envvar.MockSourcegraphDotComMode(orig)

		tests := []struct {
			name  string
			ctx   context.Context
			setup func()
		}{
			{
				name: "unauthenticated",
				ctx:  context.Background(),
				setup: func() {
					users.GetByIDFunc.SetDefaultReturn(&types.User{ID: 1}, nil)
				},
			},
			{
				name: "another user",
				ctx:  actor.WithActor(context.Background(), &actor.Actor{UID: 2}),
				setup: func() {
					users.GetByIDFunc.SetDefaultHook(func(ctx context.Context, id int32) (*types.User, error) {
						return &types.User{ID: id}, nil
					})
				},
			},
			{
				name: "site admin",
				ctx:  actor.WithActor(context.Background(), &actor.Actor{UID: 2}),
				setup: func() {
					users.GetByIDFunc.SetDefaultHook(func(ctx context.Context, id int32) (*types.User, error) {
						return &types.User{ID: id, SiteAdmin: true}, nil
					})
				},
			},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				test.setup()

				_, err := newSchemaResolver(db).UpdateUser(
					test.ctx,
					&updateUserArgs{
						User: MarshalUserID(1),
					},
				)
				got := fmt.Sprintf("%v", err)
				want := "must be authenticated as user with id 1"
				assert.Equal(t, want, got)
			})
		}
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

		RunTests(t, []*Test{
			{
				Schema: mustParseGraphQLSchema(t, db),
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

func TestUser_Organizations(t *testing.T) {
	users := dbmock.NewMockUserStore()
	users.GetByIDFunc.SetDefaultHook(func(_ context.Context, id int32) (*types.User, error) {
		// Set up a mock set of users, consisting of two regular users and one site admin.
		knownUsers := map[int32]*types.User{
			1: {ID: 1, Username: "alice"},
			2: {ID: 2, Username: "bob"},
			3: {ID: 3, Username: "carol", SiteAdmin: true},
		}

		if user := knownUsers[id]; user != nil {
			return user, nil
		}

		t.Errorf("unknown mock user: got ID %q", id)
		return nil, errors.New("unreachable")
	})
	users.GetByUsernameFunc.SetDefaultHook(func(_ context.Context, username string) (*types.User, error) {
		if want := "alice"; username != want {
			t.Errorf("got %q, want %q", username, want)
		}
		return &types.User{ID: 1, Username: "alice"}, nil
	})
	users.GetByCurrentAuthUserFunc.SetDefaultHook(func(ctx context.Context) (*types.User, error) {
		return users.GetByID(ctx, actor.FromContext(ctx).UID)
	})

	orgs := dbmock.NewMockOrgStore()
	orgs.GetByUserIDFunc.SetDefaultHook(func(_ context.Context, userID int32) ([]*types.Org, error) {
		if want := int32(1); userID != want {
			t.Errorf("got %q, want %q", userID, want)
		}
		return []*types.Org{
			{
				ID:   1,
				Name: "org",
			},
		}, nil
	})

	db := dbmock.NewMockDB()
	db.UsersFunc.SetDefaultReturn(users)
	db.OrgsFunc.SetDefaultReturn(orgs)

	expectOrgFailure := func(t *testing.T, actorUID int32) {
		t.Helper()
		wantErr := "must be authenticated as the authorized user or as an admin (must be site admin)"
		RunTests(t, []*Test{
			{
				Context: actor.WithActor(context.Background(), &actor.Actor{UID: actorUID}),
				Schema:  mustParseGraphQLSchema(t, db),
				Query: `
					{
						user(username: "alice") {
							username
							organizations {
								totalCount
							}
						}
					}
				`,
				ExpectedResult: `{"user": null}`,
				ExpectedErrors: []*gqlerrors.QueryError{
					{
						Path:          []interface{}{"user", "organizations"},
						Message:       wantErr,
						ResolverError: errors.New(wantErr),
					},
				}},
		})
	}

	expectOrgSuccess := func(t *testing.T, actorUID int32) {
		t.Helper()
		RunTests(t, []*Test{
			{
				Context: actor.WithActor(context.Background(), &actor.Actor{UID: actorUID}),
				Schema:  mustParseGraphQLSchema(t, db),
				Query: `
					{
						user(username: "alice") {
							username
							organizations {
								totalCount
							}
						}
					}
				`,
				ExpectedResult: `
					{
						"user": {
							"username": "alice",
							"organizations": {
								"totalCount": 1
							}
						}
					}
				`,
			},
		})
	}

	t.Run("on Sourcegraph.com", func(t *testing.T) {
		orig := envvar.SourcegraphDotComMode()
		envvar.MockSourcegraphDotComMode(true)
		t.Cleanup(func() { envvar.MockSourcegraphDotComMode(orig) })

		t.Run("same user", func(t *testing.T) {
			expectOrgSuccess(t, 1)
		})

		t.Run("different user", func(t *testing.T) {
			expectOrgFailure(t, 2)
		})

		t.Run("site admin", func(t *testing.T) {
			expectOrgSuccess(t, 3)
		})
	})

	t.Run("on non-Sourcegraph.com", func(t *testing.T) {
		t.Run("same user", func(t *testing.T) {
			expectOrgSuccess(t, 1)
		})

		t.Run("different user", func(t *testing.T) {
			expectOrgFailure(t, 2)
		})

		t.Run("site admin", func(t *testing.T) {
			expectOrgSuccess(t, 3)
		})
	})
}
