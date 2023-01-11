package background

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/hexops/autogold"
	"github.com/sourcegraph/log/logtest"

	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func TestCheckBackfillCompleted(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Background()
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	postgres := database.NewDB(logger, dbtest.NewDB(logger, t))
	now := time.Now().Truncate(time.Microsecond).Round(0)

	getBackfillCompleted := func() ([]SeriesBackfillStatus, error) {
		rows, err := insightsDB.QueryContext(context.Background(), fmt.Sprintf("SELECT series_id, backfill_completed_at FROM insight_series WHERE backfill_completed_at IS NOT NULL"))
		if err != nil {
			t.Fatal(err)
		}
		results := make([]SeriesBackfillStatus, 0)
		for rows.Next() {
			var temp SeriesBackfillStatus
			if err := rows.Scan(
				&temp.SeriesId,
				&temp.BackfillCompletedAt,
			); err != nil {
				return nil, err
			}
			results = append(results, temp)
		}
		return results, nil
	}
	resetDatabase := func() {
		_, err := insightsDB.ExecContext(context.Background(), `DELETE FROM insight_series`)
		if err != nil {
			t.Fatal(err)
		}
		_, err = postgres.ExecContext(context.Background(), `DELETE FROM insights_query_runner_jobs`)
		if err != nil {
			t.Fatal(err)
		}
	}

	t.Run("Does nothing if there are no series", func(t *testing.T) {
		err := checkBackfillCompleted(ctx, postgres, insightsDB)
		if err != nil {
			t.Fatal(err)
		}
	})
	t.Run("Does nothing if there are no completed series", func(t *testing.T) {
		_, err := insightsDB.ExecContext(context.Background(), `INSERT INTO insight_series (series_id, query, backfill_completed_at, generation_method)
											VALUES (1, 'query', NULL, 'search'),
												(2, 'query', NULL, 'search');`)
		if err != nil {
			t.Fatal(err)
		}
		_, err = postgres.ExecContext(context.Background(), `INSERT INTO insights_query_runner_jobs (series_id, search_query, state, finished_at)
											VALUES (1, 'query', 'error', NULL),
												(1, 'query', 'completed', $1),
												(2, 'query', 'processing', NULL),
												(2, 'query', 'processing', NULL);`, now)
		if err != nil {
			t.Fatal(err)
		}

		err = checkBackfillCompleted(ctx, postgres, insightsDB)
		if err != nil {
			t.Fatal(err)
		}

		series, err := getBackfillCompleted()
		if err != nil {
			t.Fatal(err)
		}
		autogold.Want("SeriesBackfillCompleted", series).Equal(t, []SeriesBackfillStatus{})
	})
	t.Run("Stamps backfill_completed_at for a series with only completed jobs", func(t *testing.T) {
		resetDatabase()
		_, err := insightsDB.ExecContext(context.Background(), `INSERT INTO insight_series (series_id, query, backfill_completed_at, generation_method)
											VALUES (1, 'query', NULL, 'search'),
												(2, 'query', NULL, 'search');`)
		if err != nil {
			t.Fatal(err)
		}
		_, err = postgres.ExecContext(context.Background(), `INSERT INTO insights_query_runner_jobs (series_id, search_query, state, finished_at)
											VALUES (1, 'query', 'completed', $1),
												(1, 'query', 'completed', $2),
												(2, 'query', 'processing', NULL),
												(2, 'query', 'processing', NULL);`, now, now.Add(-time.Minute*10))
		if err != nil {
			t.Fatal(err)
		}

		err = checkBackfillCompleted(ctx, postgres, insightsDB)
		if err != nil {
			t.Fatal(err)
		}

		series, err := getBackfillCompleted()
		if err != nil {
			t.Fatal(err)
		}
		autogold.Want("SeriesBackfillCompleted", series).Equal(t, []SeriesBackfillStatus{{SeriesId: "1", BackfillCompletedAt: now.UTC()}})
	})
	t.Run("Stamps backfill_completed_at for multiple series", func(t *testing.T) {
		resetDatabase()
		_, err := insightsDB.ExecContext(context.Background(), `INSERT INTO insight_series (series_id, query, backfill_completed_at, generation_method)
											VALUES (1, 'query', NULL, 'search'),
												(2, 'query', NULL, 'search'),
												(3, 'query', NULL, 'search'),
												(4, 'query', NULL, 'search');`)
		if err != nil {
			t.Fatal(err)
		}
		_, err = postgres.ExecContext(context.Background(), `INSERT INTO insights_query_runner_jobs (series_id, search_query, state, finished_at)
											VALUES (1, 'query', 'completed', $1),
												(1, 'query', 'completed', $1),
												(2, 'query', 'completed', $1),
												(2, 'query', 'completed', $1),
												(2, 'query', 'completed', $2),
												(3, 'query', 'completed', $1),
												(3, 'query', 'processing', NULL),
												(3, 'query', 'completed', $1),
												(4, 'query', 'errored', NULL),
												(4, 'query', 'processing', NULL);`, now, now.Add(time.Minute*10))
		if err != nil {
			t.Fatal(err)
		}

		err = checkBackfillCompleted(ctx, postgres, insightsDB)
		if err != nil {
			t.Fatal(err)
		}

		series, err := getBackfillCompleted()
		if err != nil {
			t.Fatal(err)
		}
		sort.SliceStable(series, func(i, j int) bool {
			return strings.Compare(series[i].SeriesId, series[j].SeriesId) < 0
		})
		autogold.Want("SeriesBackfillCompleted", series).Equal(t, []SeriesBackfillStatus{
			{SeriesId: "1", BackfillCompletedAt: now.UTC()},
			{SeriesId: "2", BackfillCompletedAt: now.Add(time.Minute * 10).UTC()},
		})
	})
	t.Run("Does not panic if finished_at is null on a completed job", func(t *testing.T) {
		resetDatabase()
		_, err := insightsDB.ExecContext(context.Background(), `INSERT INTO insight_series (series_id, query, backfill_completed_at, generation_method)
											VALUES (1, 'query', NULL, 'search');`)
		if err != nil {
			t.Fatal(err)
		}
		_, err = postgres.ExecContext(context.Background(), `INSERT INTO insights_query_runner_jobs (series_id, search_query, state, finished_at)
											VALUES (1, 'query', 'completed', NULL);`)
		if err != nil {
			t.Fatal(err)
		}

		err = checkBackfillCompleted(ctx, postgres, insightsDB)
		if err != nil {
			t.Fatal(err)
		}

		series, err := getBackfillCompleted()
		if err != nil {
			t.Fatal(err)
		}
		autogold.Want("SeriesBackfillCompleted", series).Equal(t, []SeriesBackfillStatus{})
	})
}

type SeriesBackfillStatus struct {
	SeriesId            string
	BackfillCompletedAt time.Time
}
