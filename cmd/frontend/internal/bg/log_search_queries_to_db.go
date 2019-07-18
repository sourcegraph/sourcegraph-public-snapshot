package bg

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"gopkg.in/inconshreveable/log15.v2"
)

var QueryLogChan = make(chan QueryLogItem, 100)

type QueryLogItem struct {
	Query string
	Err   error
}

// LogQueries pulls queries from QueryLogChan and logs them to the recent_searches table in the db.
func LogSearchQueries(ctx context.Context) {
	rs := &db.RecentSearches{}
	for {
		q := <-QueryLogChan
		if err := rs.Log(ctx, q.Query); err != nil {
			log15.Error("adding query to searches table", "error", err)
		}
		if err := rs.Cleanup(ctx, 1e5); err != nil {
			log15.Error("deleting excess rows from searches table", "error", err)
		}
	}
}
