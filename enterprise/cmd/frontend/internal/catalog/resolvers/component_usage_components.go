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

func (r *componentUsageResolver) Components(ctx context.Context) ([]gql.ComponentUsedByComponentEdgeResolver, error) {
	results, err := r.cachedResults(ctx)
	if err != nil {
		return nil, err
	}

	components := dummyComponents(r.db)
	componentForPath := func(repo api.RepoName, path string) *componentResolver {
		// TODO(sqs): ignores commit SHA - is that ok?
		for _, c := range components {
			if c.component.SourceRepo == repo {
				for _, sourcePath := range c.component.SourcePaths {
					if path == sourcePath || strings.HasPrefix(path, sourcePath+"/") || sourcePath == "." {
						return c
					}
				}
			}
		}
		return nil
	}

	edgesByComponentID := map[graphql.ID]*componentUsedByComponentEdgeResolver{}
	for _, res := range results.Results() {
		if fm, ok := res.ToFileMatch(); ok {
			usedByC := componentForPath(fm.RepoName().Name, fm.Path)
			if usedByC == nil {
				continue
			}

			edge := edgesByComponentID[usedByC.ID()]
			if edge == nil {
				edge = &componentUsedByComponentEdgeResolver{
					db:      r.db,
					usedByC: usedByC,
				}
				edgesByComponentID[usedByC.ID()] = edge
			}

			for _, m := range fm.LineMatches() {
				edge.locations = append(edge.locations, gql.NewLocationResolver(fm.File(), &lsp.Range{
					Start: lsp.Position{Line: int(m.LineNumber()), Character: int(m.OffsetAndLengths()[0][0])},
					End:   lsp.Position{Line: int(m.LineNumber()), Character: int(m.OffsetAndLengths()[0][0] + m.OffsetAndLengths()[0][1])},
				}))
			}
		}
	}

	edges := make([]gql.ComponentUsedByComponentEdgeResolver, 0, len(edgesByComponentID))
	for _, edge := range edgesByComponentID {
		edges = append(edges, edge)
	}
	return edges, nil
}

type componentUsedByComponentEdgeResolver struct {
	usedByC   *componentResolver
	locations []gql.LocationResolver

	db database.DB
}

func (r *componentUsedByComponentEdgeResolver) Node() gql.ComponentResolver {
	return r.usedByC
}

func (r *componentUsedByComponentEdgeResolver) Locations(ctx context.Context) (gql.LocationConnectionResolver, error) {
	return gql.NewStaticLocationConnectionResolver(r.locations, false), nil
}
