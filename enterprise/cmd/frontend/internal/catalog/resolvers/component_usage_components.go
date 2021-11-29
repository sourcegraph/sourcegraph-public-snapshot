package resolvers

import (
	"context"
	"strings"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

func (r *catalogComponentUsageResolver) Components(ctx context.Context) ([]gql.CatalogComponentUsageComponentEdgeResolver, error) {
	results, err := r.cachedResults(ctx)
	if err != nil {
		return nil, err
	}

	components := dummyData(r.db)
	componentForPath := func(repo api.RepoName, path string) *catalogComponentResolver {
		// TODO(sqs): ignores commit SHA - is that ok?
		for _, c := range components {
			if c.sourceRepo == repo {
				for _, sourcePath := range c.sourcePaths {
					if path == sourcePath || strings.HasPrefix(path, sourcePath+"/") || sourcePath == "." {
						return c
					}
				}
			}
		}
		return nil
	}

	edgesByComponentID := map[graphql.ID]*catalogComponentUsageComponentEdgeResolver{}
	for _, res := range results.Results() {
		if fm, ok := res.ToFileMatch(); ok {
			inC := componentForPath(fm.RepoName().Name, fm.Path)
			if inC == nil {
				continue
			}

			edge := edgesByComponentID[inC.ID()]
			if edge == nil {
				edge = &catalogComponentUsageComponentEdgeResolver{
					db:   r.db,
					outC: r.component,
					inC:  inC,
				}
				edgesByComponentID[inC.ID()] = edge
			}

			for _, m := range fm.LineMatches() {
				edge.locations = append(edge.locations, gql.NewLocationResolver(fm.File(), &lsp.Range{
					Start: lsp.Position{Line: int(m.LineNumber()), Character: int(m.OffsetAndLengths()[0][0])},
					End:   lsp.Position{Line: int(m.LineNumber()), Character: int(m.OffsetAndLengths()[0][0] + m.OffsetAndLengths()[0][1])},
				}))
			}
		}
	}

	edges := make([]gql.CatalogComponentUsageComponentEdgeResolver, 0, len(edgesByComponentID))
	for _, edge := range edges {
		edges = append(edges, edge)
	}
	return edges, nil
}

type catalogComponentUsageComponentEdgeResolver struct {
	db        database.DB
	outC      *catalogComponentResolver
	inC       *catalogComponentResolver
	locations []gql.LocationResolver
}

func (r *catalogComponentUsageComponentEdgeResolver) OutComponent() gql.CatalogComponentResolver {
	return r.outC
}

func (r *catalogComponentUsageComponentEdgeResolver) InComponent() gql.CatalogComponentResolver {
	return r.inC
}

func (r *catalogComponentUsageComponentEdgeResolver) Locations(ctx context.Context) (gql.LocationConnectionResolver, error) {
	return gql.NewStaticLocationConnectionResolver(r.locations, false), nil
}
