package graphqlbackend

import (
	"context"
	"sort"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
)

// TopQueries returns the top most frequent recent queries.
func (s *schemaResolver) TopQueries(ctx context.Context, args *struct{ Limit int32 }) ([]queryCountResolver, error) {
	searches, err := db.RecentSearches.Get(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "getting recent searches from database")
	}
	histo := make(map[string]int32)
	for _, s := range searches {
		histo[s]++
	}

	var uniques []string
	for k := range histo {
		uniques = append(uniques, k)
	}

	sort.Slice(uniques, func(i, j int) bool {
		hi := histo[uniques[i]]
		hj := histo[uniques[j]]
		switch {
		case hi > hj:
			return true
		case hi < hj:
			return false
		default:
			return uniques[i] < uniques[j]
		}
	})
	if int32(len(uniques)) > args.Limit {
		uniques = uniques[:args.Limit]
	}

	var qcrs []queryCountResolver
	for _, s := range uniques {
		tqr := queryCountResolver{
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

func (r queryCountResolver) Query() string { return r.query }
func (r queryCountResolver) Count() int32  { return r.count }
