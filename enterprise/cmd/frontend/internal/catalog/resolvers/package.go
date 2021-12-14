package resolvers

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/catalog"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

type packageResolver struct {
	pkg catalog.Package
	db  database.DB
}

func (r *packageResolver) TagPackageEntity() {}

func (r *packageResolver) ID() graphql.ID {
	return relay.MarshalID("Package", r.pkg.Name) // TODO(sqs)
}

func (r *packageResolver) Type() gql.CatalogEntityType {
	return "PACKAGE"
}

func (r *packageResolver) Name() string {
	return r.pkg.Name
}

func (r *packageResolver) Description() *string {
	return nil
}

func (r *packageResolver) Owner(ctx context.Context) (*gql.EntityOwnerResolver, error) {
	return nil, nil
}

func (r *packageResolver) URL() string {
	return "/catalog/packages/" + string(r.Name())
}

func (r *packageResolver) Status(context.Context) (gql.CatalogEntityStatusResolver, error) {
	return &catalogEntityStatusResolver{
		contexts: nil,
		entityID: r.ID(),
	}, nil
}

func (r *packageResolver) CodeOwners(context.Context) (*[]gql.CatalogEntityCodeOwnerEdgeResolver, error) {
	return nil, nil
}

func (r *packageResolver) RelatedEntities(context.Context, *gql.CatalogEntityRelatedEntitiesArgs) (gql.CatalogEntityRelatedEntityConnectionResolver, error) {
	return &catalogEntityRelatedEntityConnectionResolver{
		edges: nil,
	}, nil
}

func (r *packageResolver) WhoKnows(context.Context, *gql.WhoKnowsArgs) ([]gql.WhoKnowsEdgeResolver, error) {
	return nil, nil
}
