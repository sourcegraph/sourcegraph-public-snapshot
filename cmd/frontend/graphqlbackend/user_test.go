package graphqlbackend

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"

	gqlerrors "github.com/graph-gophers/graphql-go/errors"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestUser(t *testing.T) {
	db := database.NewMockDB()
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

		users := database.NewMockUserStore()
		users.GetByUsernameFunc.SetDefaultHook(func(ctx context.Context, username string) (*types.User, error) {
			assert.Equal(t, "alice", username)
			return &types.User{ID: 1, Username: "alice"}, nil
		})
		db.UsersFunc.SetDefaultReturn(users)

		t.Run("allowed on Sourcegraph.com", func(t *testing.T) {
			orig := envvar.SourcegraphDotComMode()
			envvar.MockSourcegraphDotComMode(true)
			defer envvar.MockSourcegraphDotComMode(orig)

			checkUserByUsername(t)
		})

		t.Run("allowed on non-Sourcegraph.com", func(t *testing.T) {
			checkUserByUsername(t)
		})
	})

	t.Run("by email", func(t *testing.T) {
		users := database.NewMockUserStore()
		users.GetByVerifiedEmailFunc.SetDefaultHook(func(ctx context.Context, email string) (*types.User, error) {
			assert.Equal(t, "alice@example.com", email)
			return &types.User{ID: 1, Username: "alice"}, nil
		})
		db.UsersFunc.SetDefaultReturn(users)

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
								Path:          []any{"user"},
								Message:       wantErr,
								ResolverError: errors.New(wantErr),
							},
						},
					},
				})
			}

			orig := envvar.SourcegraphDotComMode()
			envvar.MockSourcegraphDotComMode(true)
			defer envvar.MockSourcegraphDotComMode(orig)

			t.Run("for anonymous viewer", func(t *testing.T) {
				users.GetByCurrentAuthUserFunc.SetDefaultReturn(nil, database.ErrNoCurrentUser)
				checkUserByEmailError(t, "not authenticated")
			})
			t.Run("for non-site-admin viewer", func(t *testing.T) {
				users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: false}, nil)
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
	db := database.NewMockDB()
	t.Run("only allowed by authenticated user on Sourcegraph.com", func(t *testing.T) {
		orig := envvar.SourcegraphDotComMode()
		envvar.MockSourcegraphDotComMode(true)
		defer envvar.MockSourcegraphDotComMode(orig)

		_, err := NewUserResolver(db, &types.User{ID: 1}).Email(context.Background())
		got := fmt.Sprintf("%v", err)
		want := "must be authenticated as user with id 1"
		assert.Equal(t, want, got)
	})
}

func TestUser_LatestSettings(t *testing.T) {
	db := database.NewMockDB()
	t.Run("only allowed by authenticated user on Sourcegraph.com", func(t *testing.T) {
		users := database.NewMockUserStore()
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

				_, err := NewUserResolver(db, &types.User{ID: 1}).LatestSettings(test.ctx)
				got := fmt.Sprintf("%v", err)
				want := "must be authenticated as user with id 1"
				assert.Equal(t, want, got)
			})
		}
	})
}

func TestUser_ViewerCanAdminister(t *testing.T) {
	db := database.NewMockDB()
	t.Run("only allowed by authenticated user on Sourcegraph.com", func(t *testing.T) {
		users := database.NewMockUserStore()
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

				ok, _ := NewUserResolver(db, &types.User{ID: 1}).ViewerCanAdminister(test.ctx)
				assert.False(t, ok, "ViewerCanAdminister")
			})
		}
	})
}

func TestNode_User(t *testing.T) {
	users := database.NewMockUserStore()
	users.GetByIDFunc.SetDefaultReturn(&types.User{ID: 1, Username: "alice"}, nil)

	db := database.NewMockDB()
	db.UsersFunc.SetDefaultReturn(users)

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
	db := database.NewMockDB()

	t.Run("not site admin nor the same user", func(t *testing.T) {
		users := database.NewMockUserStore()
		users.GetByIDFunc.SetDefaultHook(func(ctx context.Context, id int32) (*types.User, error) {
			return &types.User{
				ID:       id,
				Username: strconv.Itoa(int(id)),
			}, nil
		})
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 2, Username: "2"}, nil)
		db.UsersFunc.SetDefaultReturn(users)

		result, err := newSchemaResolver(db).UpdateUser(context.Background(),
			&updateUserArgs{
				User: "VXNlcjox",
			},
		)
		got := fmt.Sprintf("%v", err)
		want := "must be authenticated as the authorized user or as an admin (must be site admin)"
		assert.Equal(t, want, got)
		assert.Nil(t, result)
	})

	t.Run("disallow suspicious names", func(t *testing.T) {
		orig := envvar.SourcegraphDotComMode()
		envvar.MockSourcegraphDotComMode(true)
		defer envvar.MockSourcegraphDotComMode(orig)

		users := database.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 1}, nil)
		db.UsersFunc.SetDefaultReturn(users)

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
		defer conf.Mock(nil)

		users := database.NewMockUserStore()
		users.GetByIDFunc.SetDefaultHook(func(ctx context.Context, id int32) (*types.User, error) {
			return &types.User{ID: id}, nil
		})
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 1}, nil)
		db.UsersFunc.SetDefaultReturn(users)

		ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
		result, err := newSchemaResolver(db).UpdateUser(ctx,
			&updateUserArgs{
				User:     "VXNlcjox",
				Username: strptr("alice"),
			},
		)
		got := fmt.Sprintf("%v", err)
		want := "unable to change username because auth.enableUsernameChanges is false in site configuration"
		assert.Equal(t, want, got)
		assert.Nil(t, result)
	})

	t.Run("non site admin can change non-username fields", func(t *testing.T) {
		conf.Mock(&conf.Unified{
			SiteConfiguration: schema.SiteConfiguration{
				AuthEnableUsernameChanges: false,
			},
		})
		defer conf.Mock(nil)

		mockUser := &types.User{
			ID:          1,
			Username:    "alice",
			DisplayName: "alice-updated",
			AvatarURL:   "http://www.example.com/alice-updated",
		}
		users := database.NewMockUserStore()
		users.GetByIDFunc.SetDefaultReturn(mockUser, nil)
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(mockUser, nil)
		users.UpdateFunc.SetDefaultReturn(nil)
		db.UsersFunc.SetDefaultReturn(users)

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
		users := database.NewMockUserStore()
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

	t.Run("bad avatarURL", func(t *testing.T) {
		users := database.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)
		db.UsersFunc.SetDefaultReturn(users)

		tests := []struct {
			name      string
			avatarURL string
			wantErr   string
		}{
			{
				name:      "exceeded 3000 characters",
				avatarURL: strings.Repeat("bad", 1001),
				wantErr:   "avatar URL exceeded 3000 characters",
			},
			{
				name:      "not HTTP nor HTTPS",
				avatarURL: "ftp://avatars3.githubusercontent.com/u/404",
				wantErr:   "avatar URL must be an HTTP or HTTPS URL",
			},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				_, err := newSchemaResolver(db).UpdateUser(
					context.Background(),
					&updateUserArgs{
						User:      MarshalUserID(1),
						AvatarURL: &test.avatarURL,
					},
				)
				got := fmt.Sprintf("%v", err)
				assert.Equal(t, test.wantErr, got)
			})
		}
	})

	t.Run("success", func(t *testing.T) {
		users := database.NewMockUserStore()
		users.GetByIDFunc.SetDefaultHook(func(ctx context.Context, id int32) (*types.User, error) {
			return &types.User{
				ID:       id,
				Username: strconv.Itoa(int(id)),
			}, nil
		})
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)
		users.UpdateFunc.SetDefaultReturn(nil)
		db.UsersFunc.SetDefaultReturn(users)

		RunTests(t, []*Test{
			{
				Schema: mustParseGraphQLSchema(t, db),
				Query: `
			mutation {
				updateUser(
					user: "VXNlcjox",
					username: "alice.bob-chris-",
					avatarURL: "https://avatars3.githubusercontent.com/u/404"
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
	users := database.NewMockUserStore()
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

	orgs := database.NewMockOrgStore()
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

	db := database.NewMockDB()
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
						Path:          []any{"user", "organizations"},
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
