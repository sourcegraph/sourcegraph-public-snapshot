package resolvers

import (
	"context"

	"github.com/sourcegraph/go-langserver/pkg/lsp"
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

func (r *componentUsageResolver) Locations(ctx context.Context) (gql.LocationConnectionResolver, error) {
	results, err := r.cachedResults(ctx)
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
