package resolvers

import (
	"context"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

func (r *catalogResolver) Graph(ctx context.Context) (gql.CatalogGraphResolver, error) {
	return &catalogGraphResolver{}, nil
}

type catalogGraphResolver struct {
	nodes []*gql.CatalogEntityResolver
	edges []gql.CatalogEntityRelationEdgeResolver
}

func (r *catalogGraphResolver) Nodes() []*gql.CatalogEntityResolver            { return r.nodes }
func (r *catalogGraphResolver) Edges() []gql.CatalogEntityRelationEdgeResolver { return r.edges }

type catalogEntityRelationEdgeResolver struct {
	outNode *gql.CatalogEntityResolver
	outType gql.CatalogEntityRelationType

	inNode *gql.CatalogEntityResolver
	inType gql.CatalogEntityRelationType
}

func (r *catalogEntityRelationEdgeResolver) OutNode() *gql.CatalogEntityResolver    { return r.outNode }
func (r *catalogEntityRelationEdgeResolver) OutType() gql.CatalogEntityRelationType { return r.outType }
func (r *catalogEntityRelationEdgeResolver) InNode() *gql.CatalogEntityResolver     { return r.inNode }
func (r *catalogEntityRelationEdgeResolver) InType() gql.CatalogEntityRelationType  { return r.inType }
