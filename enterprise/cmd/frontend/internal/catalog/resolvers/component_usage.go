package resolvers

import (
	"context"
	"strings"
	"sync"

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
		search:    search,
		component: r,
		db:        r.db,
	}, nil
}

type catalogComponentUsageResolver struct {
	search    gql.SearchImplementer
	component *catalogComponentResolver
	db        database.DB

	resultsOnce sync.Once
	results     *gql.SearchResultsResolver
	resultsErr  error
}

func (r *catalogComponentUsageResolver) cachedResults(ctx context.Context) (*gql.SearchResultsResolver, error) {
	r.resultsOnce.Do(func() {
		r.results, r.resultsErr = r.search.Results(ctx)
	})
	return r.results, r.resultsErr
}
