package background

import (
	"context"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

var _ dbworker.Handler = &queryRunner{}

// queryRunner implements the dbworker.Handler interface by executing search queries and inserting
// insights about them to the insights database.
type queryRunner struct {
	workerStore   *store.Store // TODO(slimsag): should not create in TimescaleDB
	insightsStore *store.Store
}

func (r *queryRunner) Handle(ctx context.Context, workerStore dbworkerstore.Store, record workerutil.Record) (err error) {
	defer func() {
		if err != nil {
			log15.Error("insights.queryRunner.Handle", "error", err)
		}
	}()

	s := r.workerStore.With(workerStore)

	// TODO(slimsag): get query from work queue similar to below:
	var q = struct {
		ID int
	}{}
	newQuery := "errorf"
	/*
		var q *cm.MonitorQuery
		q, err = s.GetQueryByRecordID(ctx, record.RecordID())
		if err != nil {
			return err
		}
	*/

	// Search.
	var results *gqlSearchResponse
	results, err = search(ctx, newQuery)
	if err != nil {
		return err
	}
	var matchCount int
	if results != nil {
		matchCount = results.Data.Search.Results.MatchCount
	}
	// TODO(slimsag): record result count to insights DB

	// TODO(slimsag): implement equivilent?
	_ = s
	_ = matchCount
	_ = q
	/*
		// Log next_run and latest_result to table cm_queries.
		newLatestResult := latestResultTime(q.LatestResult, results, err)
		err = s.SetTriggerQueryNextRun(ctx, q.Id, s.Clock()().Add(5*time.Minute), newLatestResult.UTC())
		if err != nil {
			return err
		}
		// Log the actual query we ran and whether we got any new results.
		err = s.LogSearch(ctx, newQuery, numResults, record.RecordID())
		if err != nil {
			return fmt.Errorf("LogSearch: %w", err)
		}
	*/
	return nil
}
