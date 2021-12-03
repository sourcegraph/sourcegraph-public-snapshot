package resolvers

import (
	"context"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/catalog"
)

func (r *catalogComponentResolver) Owner(ctx context.Context) (*gql.EntityOwnerResolver, error) {
	if r.component.OwnedBy == "" {
		return nil, nil
	}

	group := catalog.GroupByName(r.component.OwnedBy)
	if group == nil {
		return nil, nil
	}

	return &gql.EntityOwnerResolver{
		Group: &groupResolver{group: *group, db: r.db},
	}, nil
}
