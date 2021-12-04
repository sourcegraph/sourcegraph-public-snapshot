package resolvers

import (
	"context"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/catalog"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

func makeGraphData(db database.DB, match func(*catalogComponentResolver) bool) *catalogGraphResolver {
	if match == nil {
		match = func(*catalogComponentResolver) bool { return true }
	}

	var graph catalogGraphResolver

	components, _, edges := catalog.Data()
	var entities []gql.CatalogEntity
	for _, c := range components {
		cr := &catalogComponentResolver{component: c, db: db}
		entities = append(entities, cr)
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

	edgeMatches := map[*gql.CatalogEntityResolver]struct{}{}
	for _, e := range edges {
		outNode := findNodeByName(e.Out)
		inNode := findNodeByName(e.In)
		if outNode == nil || inNode == nil {
			continue
		}
		if match(outNode.CatalogEntity.(*catalogComponentResolver)) || match(inNode.CatalogEntity.(*catalogComponentResolver)) {
			graph.edges = append(graph.edges, &catalogEntityRelationEdgeResolver{
				type_:   gql.CatalogEntityRelationType(e.Type),
				outNode: outNode,
				inNode:  inNode,
			})
			edgeMatches[inNode] = struct{}{}
			edgeMatches[outNode] = struct{}{}
		}
	}

	keepNodes := graph.nodes[:0]
	for _, node := range graph.nodes {
		if _, edgeMatches := edgeMatches[node]; edgeMatches {
			keepNodes = append(keepNodes, node)
		}
	}
	graph.nodes = keepNodes

	return &graph
}

func (r *catalogResolver) Graph(ctx context.Context, args *gql.CatalogGraphArgs) (gql.CatalogGraphResolver, error) {
	// TODO(sqs): support literal query search
	var query string
	if args.Query != nil {
		query = *args.Query
	}

	return makeGraphData(r.db, getQueryMatcher(query)), nil
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
