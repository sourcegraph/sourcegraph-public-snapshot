package store

import (
	"context"
	"database/sql"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

type InsightStore struct {
	*basestore.Store
}

// NewInsightStore returns a new InsightStore backed by the given Timescale db.
func NewInsightStore(db dbutil.DB) *InsightStore {
	return &InsightStore{Store: basestore.NewWithDB(db, sql.TxOptions{})}
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
	q := sqlf.Sprintf(getInsightByViewSql)
	if len(args.UniqueIDs) > 0 {
		q = sqlf.Join([]*sqlf.Query{q, sqlf.Sprintf("AND iv.unique_id = ANY(%s)", args.UniqueIDs)}, " ")
	}
	log15.Info("querying for insights", "query", q.Query(sqlf.PostgresBindVar), "args", q.Args())
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

const getInsightByViewSql = `
-- source: enterprise/internal/insights/store/insight_store.go:Get
SELECT iv.unique_id, iv.title, iv.description, ivs.label, ivs.stroke,
i.series_id, i.query, i.created_at, i.oldest_historical_at, i.last_recorded_at,
i.next_recording_after, i.recording_interval_days
FROM insight_view iv
         JOIN insight_view_series ivs ON iv.id = ivs.insight_view_id
         JOIN insight_series i ON ivs.insight_series_id = i.id
WHERE i.deleted_at IS NULL
ORDER BY iv.unique_id
`
