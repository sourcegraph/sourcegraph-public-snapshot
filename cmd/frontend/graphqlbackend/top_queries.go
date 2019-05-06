package graphqlbackend

import (
	"context"

	"github.com/pkg/errors"
)

// TopQueries returns the top most frequent recent queries.
func (s *schemaResolver) TopQueries(ctx context.Context, args *struct{ Limit int32 }) ([]queryCountResolver, error) {
	queriesCounts, err := s.recentSearches.Top(ctx, args.Limit)
	if err != nil {
		return nil, errors.Wrapf(err, "asking table for top %d search queries", args.Limit)
	}
	var qcrs []queryCountResolver
	for q, c := range queriesCounts {
		tqr := queryCountResolver{
			query: q,
			count: c,
		}
		qcrs = append(qcrs, tqr)
	}
	return qcrs, nil
}

type queryCountResolver struct {
	// query is a search query.
	query string

	// count is how many times the search query occurred.
	count int32
}

func (r queryCountResolver) Query() string { return r.query }
func (r queryCountResolver) Count() int32  { return r.count }
