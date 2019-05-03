package graphqlbackend

import (
	"context"
	"sort"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
)

// TopQueries returns the top 1000 most frequent recent queries.
func (s *schemaResolver) TopQueries(ctx context.Context) ([]*queryCountResolver, error) {
	searches, err := db.RecentSearches.Get(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "getting recent searches from database")
	}
	histo := make(map[string]int32)
	for _, s := range searches {
		histo[s]++
	}
	sort.Slice(searches, func(i, j int) bool { return histo[searches[i]] > histo[searches[j]] })
	wantLen := 1000
	if len(searches) > wantLen {
		searches = searches[:wantLen]
	}

	var qcrs []*queryCountResolver
	for _, s := range searches {
		tqr := &queryCountResolver{
			query: s,
			count: histo[s],
		}
		qcrs = append(qcrs, tqr)
	}
	return qcrs, nil
}

type queryCountResolver struct {
	query string
	count int32
}

func (r *queryCountResolver) Query() string { return r.query }
func (r *queryCountResolver) Count() int32  { return r.count }
