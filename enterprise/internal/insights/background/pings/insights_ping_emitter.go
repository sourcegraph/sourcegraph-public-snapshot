package pings

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/usagestats"

	"github.com/inconshreveable/log15"

	"github.com/cockroachdb/errors"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"

	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// NewInsightsPingEmitterJob will emit pings from Code Insights that involve enterprise features such as querying
// directly against the code insights database.
func NewInsightsPingEmitterJob(ctx context.Context, base dbutil.DB, insights dbutil.DB) goroutine.BackgroundRoutine {
	interval := time.Minute * 60
	e := InsightsPingEmitter{
		postgresDb: base,
		insightsDb: insights,
	}

	return goroutine.NewPeriodicGoroutine(ctx, interval,
		goroutine.NewHandlerWithErrorMessage("insights_pings_emitter", e.emit))
}

type InsightsPingEmitter struct {
	postgresDb dbutil.DB
	insightsDb dbutil.DB
}

func (e *InsightsPingEmitter) emit(ctx context.Context) error {
	log15.Info("Emitting Code Insights Pings")

	var counts types.InsightTotalCounts
	byViewType, err := e.GetTotalCountByViewType(ctx)
	if err != nil {
		return errors.Wrap(err, "GetTotalCountByViewType")
	}
	counts.ViewCounts = byViewType

	bySeriesType, err := e.GetTotalCountBySeriesType(ctx)
	if err != nil {
		return errors.Wrap(err, "GetTotalCountBySeriesType")
	}
	counts.SeriesCounts = bySeriesType

	byViewSeriesType, err := e.GetTotalCountByViewSeriesType(ctx)
	if err != nil {
		return errors.Wrap(err, "GetTotalCountByViewSeriesType")
	}
	counts.ViewSeriesCounts = byViewSeriesType

	marshal, err := json.Marshal(counts)
	if err != nil {
		return errors.Wrap(err, "Marshal")
	}

	err = e.SaveEvent(ctx, usagestats.InsightsTotalCountPingName, marshal)
	if err != nil {
		return errors.Wrap(err, "InsightsPingEmit")
	}
	return nil
}

func (e *InsightsPingEmitter) SaveEvent(ctx context.Context, name string, argument json.RawMessage) error {
	store := database.EventLogs(e.postgresDb)

	err := store.Insert(ctx, &database.Event{
		Name:            name,
		UserID:          0,
		AnonymousUserID: "backend",
		Argument:        argument,
		Timestamp:       time.Now(),
		Source:          "BACKEND",
	})
	if err != nil {
		return err
	}
	return nil
}

func (e *InsightsPingEmitter) GetTotalCountByViewType(ctx context.Context) (_ []types.InsightViewsCountPing, err error) {
	rows, err := e.insightsDb.QueryContext(ctx, insightViewTotalCountQuery)
	if err != nil {
		return nil, err
	}

	if err != nil {
		return []types.InsightViewsCountPing{}, err
	}
	defer func() { err = rows.Close() }()

	results := make([]types.InsightViewsCountPing, 0)
	for rows.Next() {
		stats := types.InsightViewsCountPing{}
		if err := rows.Scan(&stats.ViewType, &stats.TotalCount); err != nil {
			return []types.InsightViewsCountPing{}, err
		}
		results = append(results, stats)
	}

	return results, nil
}

func (e *InsightsPingEmitter) GetTotalCountByViewSeriesType(ctx context.Context) (_ []types.InsightViewSeriesCountPing, err error) {
	q := fmt.Sprintf(insightViewSeriesTotalCountQuery, generationMethodCaseStr)
	rows, err := e.insightsDb.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}

	if err != nil {
		return []types.InsightViewSeriesCountPing{}, err
	}
	defer func() { err = rows.Close() }()

	results := make([]types.InsightViewSeriesCountPing, 0)
	for rows.Next() {
		stats := types.InsightViewSeriesCountPing{}
		if err := rows.Scan(&stats.ViewType, &stats.GenerationType, &stats.TotalCount); err != nil {
			return []types.InsightViewSeriesCountPing{}, err
		}
		results = append(results, stats)
	}

	return results, nil
}

func (e *InsightsPingEmitter) GetTotalCountBySeriesType(ctx context.Context) (_ []types.InsightSeriesCountPing, err error) {
	q := fmt.Sprintf(insightSeriesTotalCountQuery, generationMethodCaseStr)
	rows, err := e.insightsDb.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}

	if err != nil {
		return []types.InsightSeriesCountPing{}, err
	}
	defer func() { err = rows.Close() }()

	results := make([]types.InsightSeriesCountPing, 0)
	for rows.Next() {
		stats := types.InsightSeriesCountPing{}
		if err := rows.Scan(&stats.GenerationType, &stats.TotalCount); err != nil {
			return []types.InsightSeriesCountPing{}, err
		}
		results = append(results, stats)
	}

	return results, nil
}

const generationMethodCaseStr = `
CASE
   WHEN (sample_interval_unit = 'MONTH' AND sample_interval_value = 0) THEN 'language-stats'
   WHEN (CARDINALITY(repositories) = 0 OR repositories IS NULL) THEN 'search-global'
   ELSE 'search'
END AS generation_method
`

const insightViewSeriesTotalCountQuery = `
SELECT presentation_type,
       %s,
       COUNT(*)
FROM insight_series
         JOIN insight_view_series ivs ON insight_series.id = ivs.insight_series_id
         JOIN insight_view iv ON ivs.insight_view_id = iv.id
WHERE deleted_at IS NULL
GROUP BY presentation_type, generation_method;
`

const insightSeriesTotalCountQuery = `
SELECT %s,
       COUNT(*)
FROM insight_series
WHERE deleted_at IS NULL
GROUP BY generation_method;
`

const insightViewTotalCountQuery = `
SELECT presentation_type, COUNT(*)
FROM insight_view
GROUP BY presentation_type;
`
