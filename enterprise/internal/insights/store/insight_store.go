package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

type InsightStore struct {
	*basestore.Store
	Now func() time.Time
}

// NewInsightStore returns a new InsightStore backed by the given Timescale db.
func NewInsightStore(db dbutil.DB) *InsightStore {
	return &InsightStore{Store: basestore.NewWithDB(db, sql.TxOptions{}), Now: func() time.Time {
		return time.Now()
	}}
}

// Handle returns the underlying transactable database handle.
// Needed to implement the ShareableStore interface.
func (s *InsightStore) Handle() *basestore.TransactableHandle { return s.Store.Handle() }

// With creates a new InsightStore with the given basestore.Shareable store as the underlying basestore.Store.
// Needed to implement the basestore.Store interface
func (s *InsightStore) With(other basestore.ShareableStore) *Store {
	return &Store{Store: s.Store.With(other)}
}

type InsightQueryArgs struct {
	UniqueIDs []string
}

func (s *InsightStore) Get(ctx context.Context, args InsightQueryArgs) ([]types.InsightViewSeries, error) {
	preds := make([]*sqlf.Query, 0)

	if len(args.UniqueIDs) > 0 {
		elems := make([]*sqlf.Query, 0)
		for _, id := range args.UniqueIDs {
			elems = append(elems, sqlf.Sprintf("%s", id))
		}
		preds = append(preds, sqlf.Sprintf("iv.unique_id IN (%s)", sqlf.Join(elems, ",")))
	}
	if len(preds) == 0 {
		preds = append(preds, sqlf.Sprintf("%s", "TRUE"))
	}

	q := sqlf.Sprintf(getInsightByViewSql, sqlf.Join(preds, "\n AND"))
	rows, err := s.Query(ctx, q)
	if err != nil {
		return []types.InsightViewSeries{}, err
	}

	results := make([]types.InsightViewSeries, 0)
	for rows.Next() {
		var temp types.InsightViewSeries
		if err := rows.Scan(
			&temp.UniqueID,
			&temp.Title,
			&temp.Description,
			&temp.Label,
			&temp.Stroke,
			&temp.SeriesID,
			&temp.Query,
			&temp.CreatedAt,
			&temp.OldestHistoricalAt,
			&temp.LastRecordedAt,
			&temp.NextRecordingAfter,
			&temp.RecordingIntervalDays); err != nil {
			return []types.InsightViewSeries{}, err
		}
		results = append(results, temp)
	}
	return results, nil
}

func (s *InsightStore) CreateView(ctx context.Context, view types.InsightView) (types.InsightView, error) {
	row := s.QueryRow(ctx, sqlf.Sprintf(createInsightViewSql,
		view.Title,
		view.Description,
		view.UniqueID,
	))
	if row.Err() != nil {
		return types.InsightView{}, row.Err()
	}
	var id int
	err := row.Scan(&id)
	if err != nil {
		return types.InsightView{}, err
	}
	view.ID = id
	return view, nil
}

func (s *InsightStore) CreateSeries(ctx context.Context, series types.InsightSeries) (types.InsightSeries, error) {
	if series.CreatedAt.IsZero() {
		series.CreatedAt = s.Now()
	}
	row := s.QueryRow(ctx, sqlf.Sprintf(createInsightSeriesSql,
		series.SeriesID,
		series.Query,
		series.CreatedAt,
		series.OldestHistoricalAt,
		series.LastRecordedAt,
		series.NextRecordingAfter,
		series.RecordingIntervalDays,
	))
	if row.Err() != nil {
		return types.InsightSeries{}, row.Err()
	}
	var id int
	err := row.Scan(&id)
	if err != nil {
		return types.InsightSeries{}, err
	}
	series.ID = id
	return series, nil
}

const createInsightViewSql = `
INSERT INTO insight_view (title, description, unique_id)
VALUES (%s, %s, %s)
returning id;`

const createInsightSeriesSql = `
-- source: enterprise/internal/insights/store/insight_store.go:CreateSeries
INSERT INTO insight_series (series_id, query, created_at, oldest_historical_at, last_recorded_at,
                            next_recording_after, recording_interval_days)
VALUES (%s, %s, %s, %s, %s, %s, %s)
RETURNING id;`

const getInsightByViewSql = `
-- source: enterprise/internal/insights/store/insight_store.go:Get
SELECT iv.unique_id, iv.title, iv.description, ivs.label, ivs.stroke,
i.series_id, i.query, i.created_at, i.oldest_historical_at, i.last_recorded_at,
i.next_recording_after, i.recording_interval_days
FROM insight_view iv
         JOIN insight_view_series ivs ON iv.id = ivs.insight_view_id
         JOIN insight_series i ON ivs.insight_series_id = i.id
WHERE %s
ORDER BY iv.unique_id, i.series_id
`
