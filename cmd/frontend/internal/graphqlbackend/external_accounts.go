package graphqlbackend

import (
	"context"
	"errors"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
)

func (r *userResolver) ExternalAccounts(ctx context.Context) ([]*externalAccountResolver, error) {
	// 🚨 SECURITY: Only the user and site admins should be able to see the user's external accounts.
	if err := backend.CheckSiteAdminOrSameUser(ctx, r.user.ID); err != nil {
		return nil, err
	}

	if r.user.ExternalID == nil {
		return nil, nil
	}
	return []*externalAccountResolver{{user: r, serviceID: r.user.ExternalProvider, accountID: *r.user.ExternalID}}, nil
}

func (r *schemaResolver) DeleteExternalAccount(ctx context.Context, args *struct {
	ExternalAccount graphql.ID
}) (*EmptyResponse, error) {
	return nil, errors.New("not yet implemented")
}
