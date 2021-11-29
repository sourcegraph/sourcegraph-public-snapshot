package resolvers

import (
	"context"
	"strings"

	"github.com/sourcegraph/go-langserver/pkg/lsp"
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

func (r *catalogComponentResolver) Usage(ctx context.Context, args *gql.CatalogComponentUsageArgs) (gql.CatalogComponentUsageResolver, error) {
	var queries []string
	for _, p := range r.usagePatterns {
		queries = append(queries, p.query)
	}

	search, err := gql.NewSearchImplementer(ctx, r.db, &gql.SearchArgs{
		Version: "V2",
		Query:   "((" + strings.Join(queries, ") OR (") + "))",
	})
	if err != nil {
		return nil, err
	}
	return &catalogComponentUsageResolver{
		search: search,
		db:     r.db,
	}, nil
}

type catalogComponentUsageResolver struct {
	search gql.SearchImplementer
	db     database.DB
}

func (r *catalogComponentUsageResolver) Locations(ctx context.Context) (gql.LocationConnectionResolver, error) {
	results, err := r.search.Results(ctx)
	if err != nil {
		return nil, err
	}

	var locs []gql.LocationResolver
	for _, res := range results.Results() {
		if fm, ok := res.ToFileMatch(); ok {
			for _, m := range fm.LineMatches() {
				locs = append(locs, gql.NewLocationResolver(fm.File(), &lsp.Range{
					Start: lsp.Position{Line: int(m.LineNumber()), Character: int(m.OffsetAndLengths()[0][0])},
					End:   lsp.Position{Line: int(m.LineNumber()), Character: int(m.OffsetAndLengths()[0][0] + m.OffsetAndLengths()[0][1])},
				}))
			}
		}
	}

	return gql.NewStaticLocationConnectionResolver(locs, false), nil
}
