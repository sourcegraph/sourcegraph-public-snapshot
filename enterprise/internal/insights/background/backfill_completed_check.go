package background

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"

	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Periodically check for series that have been backfilled since the last check and set timestamps for them.
func NewBackfillCompletedCheckJob(ctx context.Context, postgres database.DB, insightsdb edb.InsightsDB) goroutine.BackgroundRoutine {
	interval := time.Hour * 2

	return goroutine.NewPeriodicGoroutine(
		ctx,
		"insights.backill_completed_check", "sets timestamps for when a series' backfilled operation has completed",
		interval,
		goroutine.HandlerFunc(func(ctx context.Context) (err error) {
			return checkBackfillCompleted(ctx, postgres, insightsdb)
		}),
	)
}

func checkBackfillCompleted(ctx context.Context, postgres database.DB, insightsdb edb.InsightsDB) (err error) {
	insightStore := store.NewInsightStore(insightsdb)
	insightTx, err := insightStore.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = insightTx.Done(err) }()

	// First select all series ids for which backfill has not yet been marked as complete.
	series, err := insightStore.GetDataSeries(ctx, store.GetDataSeriesArgs{BackfillNotComplete: true})
	if err != nil {
		return errors.Wrap(err, "GetSeriesIdsBackfillNotComplete")
	}
	if len(series) == 0 {
		return nil
	}

	// Query the jobs queue to find out the status of the series.
	statusRows, err := getStatusRows(ctx, series, postgres)
	if err != nil {
		return errors.Wrap(err, "getStatusRows")
	}

	lastCompletedJob := make(map[string]time.Time)
	inProgressSeries := map[string]struct{}{}
	exists := struct{}{}
	for _, r := range statusRows {
		if r.StatusType == "completed" && r.FinishedAt != nil {
			lastCompletedJob[r.SeriesId] = *r.FinishedAt
		} else {
			// If this is any status other than "completed", the series is still backfilling
			inProgressSeries[r.SeriesId] = exists
		}
	}

	// For each series that has completed jobs and is not stil in progress, stamp it.
	for seriesId, timestamp := range lastCompletedJob {
		if _, ok := inProgressSeries[seriesId]; ok {
			continue
		}
		err = insightTx.SetSeriesBackfillComplete(ctx, seriesId, timestamp)
		if err != nil {
			return errors.Wrap(err, "SetSeriesBackfillComplete")
		}
	}

	return nil
}

func getStatusRows(ctx context.Context, series []types.InsightSeries, postgres database.DB) ([]JobStatus, error) {
	queueStore := basestore.NewWithHandle(postgres.Handle())

	elems := make([]*sqlf.Query, 0, len(series))
	for _, s := range series {
		elems = append(elems, sqlf.Sprintf("%s", s.SeriesID))
	}
	q := sqlf.Sprintf(getStateForSeriesJobs, sqlf.Join(elems, ","))
	rows, err := queueStore.Query(ctx, q)
	if err != nil {
		return nil, err
	}

	results := make([]JobStatus, 0)
	for rows.Next() {
		var temp JobStatus
		if err := rows.Scan(
			&temp.SeriesId,
			&temp.StatusType,
			&temp.StatusCount,
			&temp.FinishedAt,
		); err != nil {
			return nil, err
		}
		results = append(results, temp)
	}

	return results, nil
}

// This will return one result for each series/status grouping.
const getStateForSeriesJobs = `
SELECT series_id, state, COUNT(state), MAX(finished_at)
FROM insights_query_runner_jobs
WHERE series_id IN (%s) GROUP BY state, series_id
`

type JobStatus struct {
	SeriesId    string
	StatusType  string
	StatusCount int
	FinishedAt  *time.Time
}
