package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/inconshreveable/log15"

	"github.com/cockroachdb/errors"

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
	return &InsightStore{Store: basestore.NewWithDB(db, sql.TxOptions{}), Now: time.Now}
}

// Handle returns the underlying transactable database handle.
// Needed to implement the ShareableStore interface.
func (s *InsightStore) Handle() *basestore.TransactableHandle { return s.Store.Handle() }

// With creates a new InsightStore with the given basestore.Shareable store as the underlying basestore.Store.
// Needed to implement the basestore.Store interface
func (s *InsightStore) With(other *InsightStore) *InsightStore {
	return &InsightStore{Store: s.Store.With(other.Store), Now: other.Now}
}

func (s *InsightStore) Transact(ctx context.Context) (*InsightStore, error) {
	txBase, err := s.Store.Transact(ctx)
	return &InsightStore{Store: txBase, Now: s.Now}, err
}

// InsightQueryArgs contains query predicates for fetching viewable insight series. Any provided values will be
// included as query arguments.
type InsightQueryArgs struct {
	UniqueIDs []string
	UniqueID  string
}

// Get returns all matching viewable insight series.
func (s *InsightStore) Get(ctx context.Context, args InsightQueryArgs) ([]types.InsightViewSeries, error) {
	preds := make([]*sqlf.Query, 0, 2)

	if len(args.UniqueIDs) > 0 {
		elems := make([]*sqlf.Query, 0, len(args.UniqueIDs))
		for _, id := range args.UniqueIDs {
			elems = append(elems, sqlf.Sprintf("%s", id))
		}
		preds = append(preds, sqlf.Sprintf("iv.unique_id IN (%s)", sqlf.Join(elems, ",")))
	}
	if len(args.UniqueID) > 0 {
		preds = append(preds, sqlf.Sprintf("iv.unique_id = %s", args.UniqueID))
	}

	if len(preds) == 0 {
		preds = append(preds, sqlf.Sprintf("%s", "TRUE"))
	}

	q := sqlf.Sprintf(getInsightByViewSql, sqlf.Join(preds, "\n AND"))
	return scanInsightViewSeries(s.Query(ctx, q))
}

func (s *InsightStore) GetMapped(ctx context.Context, args InsightQueryArgs) ([]types.Insight, error) {
	viewSeries, err := s.Get(ctx, args)
	if err != nil {
		return nil, err
	}

	mapped := make(map[string][]types.InsightViewSeries, len(viewSeries))
	for _, series := range viewSeries {
		mapped[series.UniqueID] = append(mapped[series.UniqueID], series)
	}

	results := make([]types.Insight, 0, len(mapped))
	for _, seriesSet := range mapped {
		results = append(results, types.Insight{
			UniqueID:    seriesSet[0].UniqueID,
			Title:       seriesSet[0].Title,
			Description: seriesSet[0].Description,
			Series:      seriesSet,
		})
	}

	return results, nil
}

type GetDataSeriesArgs struct {
	// NextRecordingBefore will filter for results for which the next_recording_after field falls before the specified time.
	NextRecordingBefore time.Time
	Deleted             bool
}

func (s *InsightStore) GetDataSeries(ctx context.Context, args GetDataSeriesArgs) ([]types.InsightSeries, error) {
	preds := make([]*sqlf.Query, 0, 1)

	if !args.NextRecordingBefore.IsZero() {
		preds = append(preds, sqlf.Sprintf("next_recording_after < %s", args.NextRecordingBefore))
	}
	if args.Deleted {
		preds = append(preds, sqlf.Sprintf("deleted_at IS NOT NULL"))
	} else {
		preds = append(preds, sqlf.Sprintf("deleted_at IS NULL"))
	}
	if len(preds) == 0 {
		preds = append(preds, sqlf.Sprintf("%s", "TRUE"))
	}

	q := sqlf.Sprintf(getInsightDataSeriesSql, sqlf.Join(preds, "\n AND"))
	return scanDataSeries(s.Query(ctx, q))
}

func scanDataSeries(rows *sql.Rows, queryErr error) (_ []types.InsightSeries, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	results := make([]types.InsightSeries, 0)
	for rows.Next() {
		var temp types.InsightSeries
		if err := rows.Scan(
			&temp.ID,
			&temp.SeriesID,
			&temp.Query,
			&temp.CreatedAt,
			&temp.OldestHistoricalAt,
			&temp.LastRecordedAt,
			&temp.NextRecordingAfter,
			&temp.RecordingIntervalDays,
		); err != nil {
			return []types.InsightSeries{}, err
		}
		log15.Info("temp", "temp", temp)
		results = append(results, temp)
	}
	return results, nil
}

func scanInsightViewSeries(rows *sql.Rows, queryErr error) (_ []types.InsightViewSeries, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

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
			&temp.RecordingIntervalDays,
		); err != nil {
			return []types.InsightViewSeries{}, err
		}
		results = append(results, temp)
	}
	return results, nil
}

// AttachSeriesToView will associate a given insight data series with a given insight view.
func (s *InsightStore) AttachSeriesToView(ctx context.Context,
	series types.InsightSeries,
	view types.InsightView,
	metadata types.InsightViewSeriesMetadata) error {
	if series.ID == 0 || view.ID == 0 {
		return errors.New("input series or view not found")
	}
	return s.Exec(ctx, sqlf.Sprintf(attachSeriesToViewSql, series.ID, view.ID, metadata.Label, metadata.Stroke))
}

// CreateView will create a new insight view with no associated data series. This view must have a unique identifier.
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

// CreateSeries will create a new insight data series. This series must be uniquely identified by the series ID.
func (s *InsightStore) CreateSeries(ctx context.Context, series types.InsightSeries) (types.InsightSeries, error) {
	if series.CreatedAt.IsZero() {
		series.CreatedAt = s.Now()
	}
	if series.NextRecordingAfter.IsZero() {
		series.NextRecordingAfter = s.Now()
	}
	if series.OldestHistoricalAt.IsZero() {
		// TODO(insights): this value should probably somewhere more discoverable / obvious than here
		series.OldestHistoricalAt = s.Now().Add(-time.Hour * 24 * 7 * 26)
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
	var id int
	err := row.Scan(&id)
	if err != nil {
		return types.InsightSeries{}, err
	}
	series.ID = id
	return series, nil
}

type DataSeriesStore interface {
	GetDataSeries(ctx context.Context, args GetDataSeriesArgs) ([]types.InsightSeries, error)
	StampRecording(ctx context.Context, series types.InsightSeries) (types.InsightSeries, error)
}

type InsightMetadataStore interface {
	GetMapped(ctx context.Context, args InsightQueryArgs) ([]types.Insight, error)
}

// StampRecording will update the recording metadata for this series and return the InsightSeries struct with updated values.
func (s *InsightStore) StampRecording(ctx context.Context, series types.InsightSeries) (types.InsightSeries, error) {
	current := s.Now()
	next := current.Add(time.Hour * 24 * time.Duration(series.RecordingIntervalDays))
	if err := s.Exec(ctx, sqlf.Sprintf(stampRecordingSql, current, next, series.ID)); err != nil {
		return types.InsightSeries{}, err
	}
	series.LastRecordedAt = current
	series.NextRecordingAfter = next
	return series, nil
}

const stampRecordingSql = `
-- source: enterprise/internal/insights/store/insight_store.go:StampRecording
UPDATE insight_series
SET last_recorded_at = %s,
    next_recording_after = %s
WHERE id = %s;
`

const attachSeriesToViewSql = `
-- source: enterprise/internal/insights/store/insight_store.go:AttachSeriesToView
INSERT INTO insight_view_series (insight_series_id, insight_view_id, label, stroke)
VALUES (%s, %s, %s, %s);
`

const createInsightViewSql = `
-- source: enterprise/internal/insights/store/insight_store.go:CreateView
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

const getInsightDataSeriesSql = `
-- source: enterprise/internal/insights/store/insight_store.go:GetDataSeries
select id, series_id, query, created_at, oldest_historical_at, last_recorded_at, next_recording_after, recording_interval_days from insight_series
WHERE %s
`
