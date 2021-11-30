package resolvers

import (
	"context"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

func (r *catalogComponentResolver) RelatedEntities(ctx context.Context) (gql.CatalogEntityRelatedEntityConnectionResolver, error) {
	return &catalogEntityRelatedEntityConnectionResolver{}, nil
}
