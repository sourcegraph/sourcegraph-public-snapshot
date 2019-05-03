package graphqlbackend

import (
	"context"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"sort"
)

// todo: also return the counts
func (s *schemaResolver) TopQueries(ctx context.Context) ([]string, error) {
	searches, err := db.RecentSearches.Get(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "getting recent searches from database")
	}
	histo := make(map[string]int)
	for _, s := range searches {
		histo[s]++
	}
	sort.Slice(searches, func(i, j int) bool { return histo[searches[i]] > histo[searches[j]] })
	wantLen := 1000
	if len(searches) > wantLen {
		searches = searches[:wantLen]
	}
	return searches, nil
}
