package graphqlbackend

import (
	"context"
	"testing"

	"github.com/graph-gophers/graphql-go/errors"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/auth/providers"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

type mockAuthnProvider struct {
	configID providers.ConfigID
	// serviceID string
}

func (m mockAuthnProvider) ConfigID() providers.ConfigID {
	return m.configID
}

func (m mockAuthnProvider) Config() schema.AuthProviders {
	return schema.AuthProviders{
		Github: &schema.GitHubAuthProvider{
			Type: m.configID.Type,
		},
	}
}

func (m mockAuthnProvider) CachedInfo() *providers.Info {
	panic("should not be called")

	// return &providers.Info{ServiceID: m.serviceID}
}

func (m mockAuthnProvider) Refresh(ctx context.Context) error {
	panic("should not be called")
}

type mockAuthnProviderUser struct {
	Username string `json:"username,omitempty"`
	ID       int32  `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
}

func (m mockAuthnProvider) ExternalAccountInfo(ctx context.Context, account extsvc.Account) (*extsvc.PublicAccountData, error) {
	data, err := encryption.DecryptJSON[mockAuthnProviderUser](ctx, account.AccountData.Data)
	if err != nil {
		return nil, err
	}

	return &extsvc.PublicAccountData{
		Login:       data.Username,
		DisplayName: data.Name,
	}, nil
}

func TestExternalAccountDataResolver_PublicAccountDataFromJSON(t *testing.T) {
	p := mockAuthnProvider{
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

	db := dbmocks.NewMockDB()

	users := dbmocks.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(alice, nil)
	users.GetByUsernameFunc.SetDefaultHook(func(ctx context.Context, username string) (*types.User, error) {
		if username == "alice" {
			return alice, nil
		}
		return bob, nil
	})

	externalAccounts := dbmocks.NewMockUserExternalAccountsStore()
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
