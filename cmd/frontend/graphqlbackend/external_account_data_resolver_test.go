package graphqlbackend

import (
	"context"
	"testing"

	"github.com/graph-gophers/graphql-go/errors"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth/providers"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestExternalAccountDataResolver_PublicAccountDataFromJSON(t *testing.T) {
	p := providers.MockAuthProvider{
		configID: providers.ConfigID{
			Type: "foo",
			ID:   "mockproviderID",
		},
	}

	providers.Update("foo", []providers.Provider{p})
	defer providers.Update("foo", nil)

	alice := &types.User{ID: 1, Username: "alice", SiteAdmin: false}
	bob := &types.User{ID: 2, Username: "bob", SiteAdmin: true}
	account := extsvc.Account{
		ID:     1,
		UserID: alice.ID,
		AccountSpec: extsvc.AccountSpec{
			ServiceType: "foo",
		},
		AccountData: extsvc.AccountData{
			Data: extsvc.NewUnencryptedData([]byte(`{"username":"alice_2","name":"Alice Smith","id":42}`)),
		},
	}

	db := database.NewMockDB()

	users := database.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(alice, nil)
	users.GetByUsernameFunc.SetDefaultHook(func(ctx context.Context, username string) (*types.User, error) {
		if username == "alice" {
			return alice, nil
		}
		return bob, nil
	})

	externalAccounts := database.NewMockUserExternalAccountsStore()
	externalAccounts.ListFunc.SetDefaultReturn([]*extsvc.Account{&account}, nil)

	db.UsersFunc.SetDefaultReturn(users)
	db.UserExternalAccountsFunc.SetDefaultReturn(externalAccounts)

	query := `
	query UserExternalAccountData($username: String!) {
		user(username: $username) {
			externalAccounts {
				nodes {
					publicAccountData {
						displayName
						login
						url
					}
				}
			}
		}
	}
	`
	ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})

	t.Run("Account not returned if no matching auth provider found", func(t *testing.T) {
		noMatchAccount := account
		noMatchAccount.ServiceType = "no-match"
		externalAccounts.ListFunc.SetDefaultReturn([]*extsvc.Account{&noMatchAccount, &account}, nil)
		defer externalAccounts.ListFunc.SetDefaultReturn([]*extsvc.Account{&account}, nil)

		RunTests(t, []*Test{
			{
				Context:        ctx,
				Schema:         mustParseGraphQLSchema(t, db),
				Query:          query,
				ExpectedResult: `{"user":{"externalAccounts":{"nodes":[{"publicAccountData":null},{"publicAccountData":{"displayName":"Alice Smith","login":"alice_2","url":null}}]}}}`,
				Variables:      map[string]any{"username": "alice"},
			},
		})
	})

	t.Run("Alice cannot see account data for Bob", func(t *testing.T) {
		RunTests(t, []*Test{
			{
				Context:        ctx,
				Schema:         mustParseGraphQLSchema(t, db),
				Query:          query,
				ExpectedResult: `{"user":null}`,
				ExpectedErrors: []*errors.QueryError{
					{
						Message: "must be authenticated as the authorized user or site admin",
						Path:    []any{"user", "externalAccounts"},
					},
				},
				Variables: map[string]any{"username": "bob"},
			},
		})
	})

	t.Run("Works for same user and external auth provider", func(t *testing.T) {
		RunTests(t, []*Test{
			{
				Context:        ctx,
				Schema:         mustParseGraphQLSchema(t, db),
				Query:          query,
				ExpectedResult: `{"user":{"externalAccounts":{"nodes":[{"publicAccountData":{"displayName":"Alice Smith","login":"alice_2","url":null}}]}}}`,
				Variables:      map[string]any{"username": "alice"},
			},
		})
	})

	t.Run("Site admin can see any account data", func(t *testing.T) {
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(bob, nil)
		defer users.GetByCurrentAuthUserFunc.SetDefaultReturn(alice, nil)

		RunTests(t, []*Test{
			{
				Context:        ctx,
				Schema:         mustParseGraphQLSchema(t, db),
				Query:          query,
				ExpectedResult: `{"user":{"externalAccounts":{"nodes":[{"publicAccountData":{"displayName":"Alice Smith","login":"alice_2","url":null}}]}}}`,
				Variables:      map[string]any{"username": "alice"},
			},
		})
	})
}
