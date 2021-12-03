package resolvers

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/catalog"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

func makeGraphData(db database.DB, filterID graphql.ID) *catalogGraphResolver {
	var graph catalogGraphResolver

	components, _, edges := catalog.Data()
	var entities []gql.CatalogEntity
	for _, c := range components {
		entities = append(entities, &catalogComponentResolver{component: c, db: db})
	}
	graph.nodes = wrapInCatalogEntityInterfaceType(entities)

	findNodeByName := func(name string) *gql.CatalogEntityResolver {
		for _, node := range graph.nodes {
			if node.Name() == name {
				return node
			}
		}
		return nil
	}

	for _, e := range edges {
		outNode := findNodeByName(e.Out)
		inNode := findNodeByName(e.In)
		if outNode == nil || inNode == nil {
			continue
		}
		graph.edges = append(graph.edges, &catalogEntityRelationEdgeResolver{
			type_:   gql.CatalogEntityRelationType(e.Type),
			outNode: outNode,
			inNode:  inNode,
		})
	}

	return &graph
}

func (r *catalogResolver) Graph(ctx context.Context) (gql.CatalogGraphResolver, error) {
	return makeGraphData(r.db, ""), nil
}

type catalogGraphResolver struct {
	nodes []*gql.CatalogEntityResolver
	edges []gql.CatalogEntityRelationEdgeResolver
}

func (r *catalogGraphResolver) Nodes() []*gql.CatalogEntityResolver            { return r.nodes }
func (r *catalogGraphResolver) Edges() []gql.CatalogEntityRelationEdgeResolver { return r.edges }

type catalogEntityRelationEdgeResolver struct {
	type_   gql.CatalogEntityRelationType
	outNode *gql.CatalogEntityResolver
	inNode  *gql.CatalogEntityResolver
}

func (r *catalogEntityRelationEdgeResolver) Type() gql.CatalogEntityRelationType { return r.type_ }
func (r *catalogEntityRelationEdgeResolver) OutNode() *gql.CatalogEntityResolver { return r.outNode }
func (r *catalogEntityRelationEdgeResolver) InNode() *gql.CatalogEntityResolver  { return r.inNode }
