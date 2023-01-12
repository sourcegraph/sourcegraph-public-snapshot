package graphqlbackend

import (
	"context"
	"fmt"
	"testing"

	mockrequire "github.com/derision-test/go-mockgen/testutil/require"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/authz/permssync"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
	"github.com/sourcegraph/sourcegraph/internal/txemail"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func init() {
	txemail.DisableSilently()
}

func TestUserEmail_ViewerCanManuallyVerify(t *testing.T) {
	t.Parallel()

	db := database.NewMockDB()
	t.Run("only allowed by site admin", func(t *testing.T) {
		users := database.NewMockUserStore()
		db.UsersFunc.SetDefaultReturn(users)

		tests := []struct {
			name    string
			ctx     context.Context
			setup   func()
			allowed bool
		}{
			{
				name: "unauthenticated",
				ctx:  context.Background(),
				setup: func() {
					users.GetByCurrentAuthUserFunc.SetDefaultHook(func(ctx context.Context) (*types.User, error) {
						return nil, database.ErrNoCurrentUser
					})
				},
				allowed: false,
			},
			{
				name: "non site admin",
				ctx:  actor.WithActor(context.Background(), &actor.Actor{UID: 2}),
				setup: func() {
					users.GetByCurrentAuthUserFunc.SetDefaultHook(func(ctx context.Context) (*types.User, error) {
						return &types.User{
							ID:        2,
							SiteAdmin: false,
						}, nil
					})
				},
				allowed: false,
			},
			{
				name: "site admin",
				ctx:  actor.WithActor(context.Background(), &actor.Actor{UID: 2}),
				setup: func() {
					users.GetByCurrentAuthUserFunc.SetDefaultHook(func(ctx context.Context) (*types.User, error) {
						return &types.User{
							ID:        2,
							SiteAdmin: true,
						}, nil
					})
				},
				allowed: true,
			},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				test.setup()

				ok, _ := (&userEmailResolver{
					db: db,
				}).ViewerCanManuallyVerify(test.ctx)
				assert.Equal(t, test.allowed, ok, "ViewerCanManuallyVerify")
			})
		}
	})
}

func TestSetUserEmailVerified(t *testing.T) {
	t.Run("only allowed by site admins", func(t *testing.T) {
		t.Parallel()

		db := database.NewMockDB()

		db.TransactFunc.SetDefaultReturn(db, nil)
		db.DoneFunc.SetDefaultHook(func(err error) error {
			return err
		})

		ffs := database.NewMockFeatureFlagStore()
		db.FeatureFlagsFunc.SetDefaultReturn(ffs)

		users := database.NewMockUserStore()
		db.UsersFunc.SetDefaultReturn(users)
		userEmails := database.NewMockUserEmailsStore()
		db.UserEmailsFunc.SetDefaultReturn(userEmails)

		tests := []struct {
			name    string
			ctx     context.Context
			setup   func()
			wantErr string
		}{
			{
				name: "unauthenticated",
				ctx:  context.Background(),
				setup: func() {
					users.GetByCurrentAuthUserFunc.SetDefaultHook(func(ctx context.Context) (*types.User, error) {
						return nil, database.ErrNoCurrentUser
					})
					users.GetByIDFunc.SetDefaultHook(func(ctx context.Context, i int32) (*types.User, error) {
						return nil, nil
					})
				},
				wantErr: "not authenticated",
			},
			{
				name: "another user",
				ctx:  actor.WithActor(context.Background(), &actor.Actor{UID: 2}),
				setup: func() {
					users.GetByCurrentAuthUserFunc.SetDefaultHook(func(ctx context.Context) (*types.User, error) {
						return &types.User{
							ID:        2,
							SiteAdmin: false,
						}, nil
					})
				},
				wantErr: "must be site admin",
			},
			{
				name: "site admin",
				ctx:  actor.WithActor(context.Background(), &actor.Actor{UID: 2}),
				setup: func() {
					users.GetByCurrentAuthUserFunc.SetDefaultHook(func(ctx context.Context) (*types.User, error) {
						return &types.User{
							ID:        2,
							SiteAdmin: true,
						}, nil
					})
					userEmails.SetVerifiedFunc.SetDefaultHook(func(ctx context.Context, i int32, s string, b bool) error {
						// We just care at this point that we passed user authorization
						return errors.Errorf("short circuit")
					})
				},
				wantErr: "short circuit",
			},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				test.setup()

				_, err := newSchemaResolver(db, gitserver.NewClient(db)).SetUserEmailVerified(
					test.ctx,
					&setUserEmailVerifiedArgs{
						User: MarshalUserID(1),
					},
				)
				got := fmt.Sprintf("%v", err)
				assert.Equal(t, test.wantErr, got)
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

			userExternalAccounts := database.NewMockUserExternalAccountsStore()
			userExternalAccounts.DeleteFunc.SetDefaultReturn(nil)

			permssync.MockSchedulePermsSync = func(_ context.Context, _ log.Logger, _ database.DB, _ protocol.PermsSyncRequest) {}
			t.Cleanup(func() { permssync.MockSchedulePermsSync = nil })

			db := database.NewMockDB()
			db.TransactFunc.SetDefaultReturn(db, nil)
			db.DoneFunc.SetDefaultHook(func(err error) error {
				return err
			})

			db.UsersFunc.SetDefaultReturn(users)
			db.UserEmailsFunc.SetDefaultReturn(userEmails)
			db.AuthzFunc.SetDefaultReturn(authz)
			db.UserExternalAccountsFunc.SetDefaultReturn(userExternalAccounts)

			RunTests(t, test.gqlTests(db))

			if test.expectCalledGrantPendingPermissions {
				mockrequire.Called(t, authz.GrantPendingPermissionsFunc)
			} else {
				mockrequire.NotCalled(t, authz.GrantPendingPermissionsFunc)
			}
		})
	}
}
