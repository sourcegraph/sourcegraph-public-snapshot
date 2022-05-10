package graphqlbackend

import (
	"context"
	"fmt"
	"testing"
	"time"

	mockrequire "github.com/derision-test/go-mockgen/testutil/require"
	gqlerrors "github.com/graph-gophers/graphql-go/errors"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/txemail"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestUserEmail_ViewerCanManuallyVerify(t *testing.T) {
	db := database.NewMockDB()
	t.Run("only allowed by authenticated user on Sourcegraph.com", func(t *testing.T) {
		users := database.NewMockUserStore()
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

				ok, _ := (&userEmailResolver{}).ViewerCanManuallyVerify(test.ctx)
				assert.False(t, ok, "ViewerCanManuallyVerify")
			})
		}
	})
}

func TestAddUserEmail(t *testing.T) {
	db := database.NewMockDB()
	t.Run("only allowed by authenticated user on Sourcegraph.com", func(t *testing.T) {
		users := database.NewMockUserStore()
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

				_, err := newSchemaResolver(db).AddUserEmail(
					test.ctx,
					&addUserEmailArgs{
						User: MarshalUserID(1),
					},
				)
				got := fmt.Sprintf("%v", err)
				want := "must be authenticated as user with id 1"
				assert.Equal(t, want, got)
			})
		}
	})
}

func TestRemoveUserEmail(t *testing.T) {
	db := database.NewMockDB()
	t.Run("only allowed by authenticated user on Sourcegraph.com", func(t *testing.T) {
		users := database.NewMockUserStore()
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

				_, err := newSchemaResolver(db).RemoveUserEmail(
					test.ctx,
					&removeUserEmailArgs{
						User: MarshalUserID(1),
					},
				)
				got := fmt.Sprintf("%v", err)
				want := "must be authenticated as user with id 1"
				assert.Equal(t, want, got)
			})
		}
	})
}

func TestSetUserEmailPrimary(t *testing.T) {
	db := database.NewMockDB()
	t.Run("only allowed by authenticated user on Sourcegraph.com", func(t *testing.T) {
		users := database.NewMockUserStore()
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

				_, err := newSchemaResolver(db).SetUserEmailPrimary(
					test.ctx,
					&setUserEmailPrimaryArgs{
						User: MarshalUserID(1),
					},
				)
				got := fmt.Sprintf("%v", err)
				want := "must be authenticated as user with id 1"
				assert.Equal(t, want, got)
			})
		}
	})
}

func TestSetUserEmailVerified(t *testing.T) {
	t.Run("only allowed by authenticated user on Sourcegraph.com", func(t *testing.T) {
		db := database.NewMockDB()
		users := database.NewMockUserStore()
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

				_, err := newSchemaResolver(database.NewMockDB()).SetUserEmailVerified(
					test.ctx,
					&setUserEmailVerifiedArgs{
						User: MarshalUserID(1),
					},
				)
				got := fmt.Sprintf("%v", err)
				want := "manually verify user email is disabled"
				assert.Equal(t, want, got)
			})
		}
	})

	tests := []struct {
		name                                string
		gqlTests                            func(db database.DB) []*Test
		expectCalledGrantPendingPermissions bool
	}{
		{
			name: "set an email to be verified",
			gqlTests: func(db database.DB) []*Test {
				return []*Test{{
					Schema: mustParseGraphQLSchema(t, db),
					Query: `
						mutation {
							setUserEmailVerified(user: "VXNlcjox", email: "alice@example.com", verified: true) {
								alwaysNil
							}
						}
					`,
					ExpectedResult: `
						{
							"setUserEmailVerified": {
								"alwaysNil": null
							}
						}
					`,
				}}
			},
			expectCalledGrantPendingPermissions: true,
		},
		{
			name: "set an email to be unverified",
			gqlTests: func(db database.DB) []*Test {
				return []*Test{{
					Schema: mustParseGraphQLSchema(t, db),
					Query: `
						mutation {
							setUserEmailVerified(user: "VXNlcjox", email: "alice@example.com", verified: false) {
								alwaysNil
							}
						}
					`,
					ExpectedResult: `
						{
							"setUserEmailVerified": {
								"alwaysNil": null
							}
						}
					`,
				}}
			},
			expectCalledGrantPendingPermissions: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			users := database.NewMockUserStore()
			users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)

			userEmails := database.NewMockUserEmailsStore()
			userEmails.SetVerifiedFunc.SetDefaultReturn(nil)

			authz := database.NewMockAuthzStore()
			authz.GrantPendingPermissionsFunc.SetDefaultReturn(nil)

			db := database.NewMockDB()
			db.UsersFunc.SetDefaultReturn(users)
			db.UserEmailsFunc.SetDefaultReturn(userEmails)
			db.AuthzFunc.SetDefaultReturn(authz)

			RunTests(t, test.gqlTests(db))

			if test.expectCalledGrantPendingPermissions {
				mockrequire.Called(t, authz.GrantPendingPermissionsFunc)
			} else {
				mockrequire.NotCalled(t, authz.GrantPendingPermissionsFunc)
			}
		})
	}
}

func TestResendUserEmailVerification(t *testing.T) {
	t.Run("only allowed by authenticated user on Sourcegraph.com", func(t *testing.T) {
		db := database.NewMockDB()
		users := database.NewMockUserStore()
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

				_, err := newSchemaResolver(database.NewMockDB()).ResendVerificationEmail(
					test.ctx,
					&resendVerificationEmailArgs{
						User: MarshalUserID(1),
					},
				)
				got := fmt.Sprintf("%v", err)
				want := "must be authenticated as user with id 1"
				assert.Equal(t, want, got)
			})
		}
	})

	users := database.NewMockUserStore()
	users.GetByIDFunc.SetDefaultHook(func(ctx context.Context, id int32) (*types.User, error) {
		return &types.User{ID: id, SiteAdmin: true}, nil
	})
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 1, SiteAdmin: true}, nil)

	userEmails := database.NewMockUserEmailsStore()
	userEmails.SetLastVerificationFunc.SetDefaultReturn(nil)

	db := database.NewMockDB()
	db.UsersFunc.SetDefaultReturn(users)
	db.UserEmailsFunc.SetDefaultReturn(userEmails)

	knownTime := time.Time{}.Add(1337 * time.Hour)
	timeNow = func() time.Time {
		return knownTime
	}

	tests := []struct {
		name            string
		gqlTests        []*Test
		email           *database.UserEmail
		expectEmailSent bool
	}{
		{
			name: "resend a verification email",
			gqlTests: []*Test{
				{
					Schema: mustParseGraphQLSchema(t, db),
					Query: `
				mutation {
					resendVerificationEmail(user: "VXNlcjox", email: "alice@example.com") {
						alwaysNil
					}
				}
			`,
					ExpectedResult: `
				{
					"resendVerificationEmail": {
						"alwaysNil": null
    				}
				}
			`,
				},
			},
			email: &database.UserEmail{
				Email:  "alice@example.com",
				UserID: 1,
			},
			expectEmailSent: true,
		},
		{
			name: "email already verified",
			gqlTests: []*Test{
				{
					Schema: mustParseGraphQLSchema(t, db),
					Query: `
				mutation {
					resendVerificationEmail(user: "VXNlcjox", email: "alice@example.com") {
						alwaysNil
					}
				}
			`,
					ExpectedResult: `
				{
					"resendVerificationEmail": {
						"alwaysNil": null
    				}
				}
			`,
				},
			},
			email: &database.UserEmail{
				Email:      "alice@example.com",
				UserID:     1,
				VerifiedAt: &knownTime,
			},
			expectEmailSent: false,
		},
		{
			name: "invalid email",
			gqlTests: []*Test{
				{
					Schema: mustParseGraphQLSchema(t, db),
					Query: `
				mutation {
					resendVerificationEmail(user: "VXNlcjox", email: "alan@example.com") {
						alwaysNil
					}
				}
			`,
					ExpectedResult: "null",
					ExpectedErrors: []*gqlerrors.QueryError{
						{
							Path:          []any{"resendVerificationEmail"},
							Message:       "oh no!",
							ResolverError: errors.New("oh no!"),
						},
					},
				},
			},
			email: &database.UserEmail{
				Email:      "alice@example.com",
				UserID:     1,
				VerifiedAt: &knownTime,
			},
			expectEmailSent: false,
		},
		{
			name: "resend a verification email, too soon",
			gqlTests: []*Test{
				{
					Schema: mustParseGraphQLSchema(t, db),
					Query: `
				mutation {
					resendVerificationEmail(user: "VXNlcjox", email: "alice@example.com") {
						alwaysNil
					}
				}
			`,
					ExpectedResult: "null",
					ExpectedErrors: []*gqlerrors.QueryError{
						{
							Message:       "Last verification email sent too recently",
							Path:          []any{"resendVerificationEmail"},
							ResolverError: errors.New("Last verification email sent too recently"),
						},
					},
				},
			},
			email: &database.UserEmail{
				Email:  "alice@example.com",
				UserID: 1,
				LastVerificationSentAt: func() *time.Time {
					t := knownTime.Add(-30 * time.Second)
					return &t
				}(),
			},
			expectEmailSent: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var emailSent bool
			txemail.MockSend = func(ctx context.Context, msg txemail.Message) error {
				emailSent = true
				return nil
			}

			userEmails.GetFunc.SetDefaultHook(func(ctx context.Context, id int32, email string) (string, bool, error) {
				if email != test.email.Email {
					return "", false, errors.New("oh no!")
				}
				return test.email.Email, test.email.VerifiedAt != nil, nil
			})
			userEmails.GetLatestVerificationSentEmailFunc.SetDefaultReturn(test.email, nil)

			RunTests(t, test.gqlTests)

			if emailSent != test.expectEmailSent {
				t.Errorf("Expected emailSent == %t, got %t", test.expectEmailSent, emailSent)
			}
		})
	}
}
