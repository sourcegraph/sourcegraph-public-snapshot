package graphqlbackend

import (
	"context"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
)

func (r *userResolver) ExternalAccounts(ctx context.Context) ([]*externalAccountResolver, error) {
	// 🚨 SECURITY: Only the user and site admins should be able to see the user's external accounts.
	if err := backend.CheckSiteAdminOrSameUser(ctx, r.user.ID); err != nil {
		return nil, err
	}

	accounts, err := db.ExternalAccounts.List(ctx, db.ExternalAccountsListOptions{UserID: r.user.ID})
	if err != nil {
		return nil, err
	}
	resolvers := make([]*externalAccountResolver, len(accounts))
	for i, account := range accounts {
		resolvers[i] = &externalAccountResolver{user: r, account: account}
	}
	return resolvers, nil
}

func (r *schemaResolver) DeleteExternalAccount(ctx context.Context, args *struct {
	ExternalAccount graphql.ID
}) (*EmptyResponse, error) {
	id, err := unmarshalExternalAccountID(args.ExternalAccount)
	if err != nil {
		return nil, err
	}
	account, err := db.ExternalAccounts.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Only the user and site admins should be able to see a user's external accounts.
	if err := backend.CheckSiteAdminOrSameUser(ctx, account.UserID); err != nil {
		return nil, err
	}

	if err := db.ExternalAccounts.Delete(ctx, account.ID); err != nil {
		return nil, err
	}

	return &EmptyResponse{}, nil
}
