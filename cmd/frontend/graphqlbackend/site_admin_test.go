package graphqlbackend

import (
	"context"
	"testing"

	"github.com/cockroachdb/errors"
	"github.com/google/go-cmp/cmp"
	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestDeleteUser(t *testing.T) {
	t.Run("authenticated as non-admin", func(t *testing.T) {
		users := database.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{}, nil)

		db := database.NewMockDB()
		db.UsersFunc.SetDefaultReturn(users)

		ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
		result, err := (&schemaResolver{db: db}).DeleteUser(ctx, &struct {
			User graphql.ID
			Hard *bool
		}{
			User: MarshalUserID(1),
		})
		if want := backend.ErrMustBeSiteAdmin; err != want {
			t.Errorf("err: want %q but got %v", want, err)
		}
		if result != nil {
			t.Errorf("result: want nil but got %v", result)
		}
	})

	t.Run("delete current user", func(t *testing.T) {
		users := database.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 1, SiteAdmin: true}, nil)

		db := database.NewMockDB()
		db.UsersFunc.SetDefaultReturn(users)

		ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
		_, err := (&schemaResolver{db: db}).DeleteUser(ctx, &struct {
			User graphql.ID
			Hard *bool
		}{
			User: MarshalUserID(1),
		})
		want := "unable to delete current user"
		if err == nil || err.Error() != want {
			t.Fatalf("err: want %q but got %v", want, err)
		}
	})

	// Mocking all database interactions here, but they are all thoroughly tested in the lower layer in "database" package.
	users := database.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)
	users.DeleteFunc.SetDefaultReturn(nil)
	users.HardDeleteFunc.SetDefaultReturn(nil)
	users.GetByIDFunc.SetDefaultHook(func(_ context.Context, id int32) (*types.User, error) {
		return &types.User{ID: id, Username: "alice"}, nil
	})

	userEmails := database.NewMockUserEmailsStore()
	userEmails.ListByUserFunc.SetDefaultReturn([]*database.UserEmail{{Email: "alice@example.com"}}, nil)

	externalAccounts := database.NewMockUserExternalAccountsStore()
	externalAccounts.ListFunc.SetDefaultReturn(
		[]*extsvc.Account{{
			AccountSpec: extsvc.AccountSpec{
				ServiceType: extsvc.TypeGitLab,
				ServiceID:   "https://gitlab.com/",
				AccountID:   "alice_gitlab",
			},
		}},
		nil,
	)

	authzStore := database.NewMockAuthzStore()
	authzStore.RevokeUserPermissionsFunc.SetDefaultHook(func(_ context.Context, args *database.RevokeUserPermissionsArgs) error {
		if args.UserID != 6 {
			return errors.Errorf("args.UserID: want 6 but got %v", args.UserID)
		}

		expAccounts := []*extsvc.Accounts{
			{
				ServiceType: extsvc.TypeGitLab,
				ServiceID:   "https://gitlab.com/",
				AccountIDs:  []string{"alice_gitlab"},
			},
			{
				ServiceType: authz.SourcegraphServiceType,
				ServiceID:   authz.SourcegraphServiceID,
				AccountIDs:  []string{"alice@example.com", "alice"},
			},
		}
		if diff := cmp.Diff(expAccounts, args.Accounts); diff != "" {
			t.Fatalf("args.Accounts: %v", diff)
		}
		return nil
	})

	db := database.NewMockDB()
	db.UsersFunc.SetDefaultReturn(users)
	db.UserEmailsFunc.SetDefaultReturn(userEmails)
	db.UserExternalAccountsFunc.SetDefaultReturn(externalAccounts)
	db.AuthzFunc.SetDefaultReturn(authzStore)

	tests := []struct {
		name     string
		gqlTests []*Test
	}{
		{
			name: "soft delete a user",
			gqlTests: []*Test{
				{
					Schema: mustParseGraphQLSchema(t, db),
					Query: `
				mutation {
					deleteUser(user: "VXNlcjo2") {
						alwaysNil
					}
				}
			`,
					ExpectedResult: `
				{
					"deleteUser": {
						"alwaysNil": null
					}
				}
			`,
				},
			},
		},
		{
			name: "hard delete a user",
			gqlTests: []*Test{
				{
					Schema: mustParseGraphQLSchema(t, db),
					Query: `
				mutation {
					deleteUser(user: "VXNlcjo2", hard: true) {
						alwaysNil
					}
				}
			`,
					ExpectedResult: `
				{
					"deleteUser": {
						"alwaysNil": null
					}
				}
			`,
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			RunTests(t, test.gqlTests)
		})
	}
}
