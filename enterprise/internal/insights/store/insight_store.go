package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/internal/insights"

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
	UserID    []int
	OrgID     []int

	// This field will disable user level authorization checks on the insight views. This should only be used
	// when fetching insights from a container that also has authorization checks, such as a dashboard.
	WithoutAuthorization bool
}

// Get returns all matching viewable insight series.
func (s *InsightStore) Get(ctx context.Context, args InsightQueryArgs) ([]types.InsightViewSeries, error) {
	preds := make([]*sqlf.Query, 0, 4)

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
	preds = append(preds, sqlf.Sprintf("i.deleted_at IS NULL"))
	if !args.WithoutAuthorization {
		preds = append(preds, viewPermissionsQuery(args))
	}

	if len(preds) == 0 {
		preds = append(preds, sqlf.Sprintf("%s", "TRUE"))
	}

	q := sqlf.Sprintf(getInsightByViewSql, sqlf.Join(preds, "\n AND"))
	return scanInsightViewSeries(s.Query(ctx, q))
}

// viewPermissionsQuery generates the SQL query for selecting insight views based on granted permissions.
func viewPermissionsQuery(args InsightQueryArgs) *sqlf.Query {
	permsPreds := make([]*sqlf.Query, 0, 2)
	if len(args.OrgID) > 0 {
		elems := make([]*sqlf.Query, 0, len(args.OrgID))
		for _, id := range args.OrgID {
			elems = append(elems, sqlf.Sprintf("%s", id))
		}
		permsPreds = append(permsPreds, sqlf.Sprintf("ivg.org_id IN (%s)", sqlf.Join(elems, ",")))
	}
	if len(args.UserID) > 0 {
		elems := make([]*sqlf.Query, 0, len(args.UserID))
		for _, id := range args.UserID {
			elems = append(elems, sqlf.Sprintf("%s", id))
		}
		permsPreds = append(permsPreds, sqlf.Sprintf("ivg.user_id IN (%s)", sqlf.Join(elems, ",")))
	}
	permsPreds = append(permsPreds, sqlf.Sprintf("ivg.global is true"))

	return sqlf.Sprintf("(%s)", sqlf.Join(permsPreds, "OR"))
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
			Filters: types.InsightViewFilters{
				IncludeRepoRegex: seriesSet[0].DefaultFilterIncludeRepoRegex,
				ExcludeRepoRegex: seriesSet[0].DefaultFilterExcludeRepoRegex,
			},
		})
	}

	return results, nil
}

func (s *InsightStore) InsertDirtyQuery(ctx context.Context, series *types.InsightSeries, query *types.DirtyQuery) error {
	q := sqlf.Sprintf(insertDirtyQuerySql, series.ID, query.Query, query.Reason, query.ForTime, s.Now())
	return s.Exec(ctx, q)
}

// GetDirtyQueries returns up to 100 dirty queries for a given insight series.
func (s *InsightStore) GetDirtyQueries(ctx context.Context, series *types.InsightSeries) ([]*types.DirtyQuery, error) {
	// We are going to limit this for now to some fixed value, and in the future if necessary add pagination.
	limit := 100
	q := sqlf.Sprintf(getDirtyQueriesSql, series.ID, limit)
	return scanDirtyQueries(s.Query(ctx, q))
}

func scanDirtyQueries(rows *sql.Rows, queryErr error) (_ []*types.DirtyQuery, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	results := make([]*types.DirtyQuery, 0)
	for rows.Next() {
		var temp types.DirtyQuery
		if err := rows.Scan(
			&temp.ID,
			&temp.Query,
			&temp.Reason,
			&temp.ForTime,
			&temp.DirtyAt,
		); err != nil {
			return nil, err
		}
		results = append(results, &temp)
	}
	return results, nil
}

// GetDirtyQueriesAggregated returns aggregated information about dirty queries for a given series.
func (s *InsightStore) GetDirtyQueriesAggregated(ctx context.Context, seriesID string) ([]*types.DirtyQueryAggregate, error) {
	q := sqlf.Sprintf(getDirtyQueriesAggregatedSql, seriesID)
	return scanDirtyQueriesAggregated(s.Query(ctx, q))
}

func scanDirtyQueriesAggregated(rows *sql.Rows, queryErr error) (_ []*types.DirtyQueryAggregate, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	results := make([]*types.DirtyQueryAggregate, 0)
	for rows.Next() {
		var temp types.DirtyQueryAggregate
		if err := rows.Scan(
			&temp.Count,
			&temp.ForTime,
			&temp.Reason,
		); err != nil {
			return nil, err
		}
		results = append(results, &temp)
	}
	return results, nil
}

const insertDirtyQuerySql = `
-- source: enterprise/internal/insights/store/insight_store.go:InsertDirtyQuery
INSERT INTO insight_dirty_queries (insight_series_id, query, reason, for_time, dirty_at)
VALUES (%s, %s, %s, %s, %s);
`

const getDirtyQueriesSql = `
-- source: enterprise/internal/insights/store/insight_store.go:GetDirtyQueries
select id, query, reason, for_time, dirty_at from insight_dirty_queries
where insight_series_id = %s
limit %s;`

const getDirtyQueriesAggregatedSql = `
-- source: enterprise/internal/insights/store/insight_store.go:GetDirtyQueriesAggregated
select count(*) as count, for_time, reason from insight_dirty_queries
where insight_dirty_queries.insight_series_id = (select id from insight_series where series_id = %s)
group by for_time, reason;
`

type GetDataSeriesArgs struct {
	// NextRecordingBefore will filter for results for which the next_recording_after field falls before the specified time.
	NextRecordingBefore time.Time
	NextSnapshotBefore  time.Time
	IncludeDeleted      bool
	BackfillIncomplete  bool
	SeriesID            string
}

func (s *InsightStore) GetDataSeries(ctx context.Context, args GetDataSeriesArgs) ([]types.InsightSeries, error) {
	preds := make([]*sqlf.Query, 0, 1)

	if !args.NextRecordingBefore.IsZero() {
		preds = append(preds, sqlf.Sprintf("next_recording_after < %s", args.NextRecordingBefore))
	}
	if !args.NextSnapshotBefore.IsZero() {
		preds = append(preds, sqlf.Sprintf("next_snapshot_after < %s", args.NextSnapshotBefore))
	}
	if !args.IncludeDeleted {
		preds = append(preds, sqlf.Sprintf("deleted_at IS NULL"))
	}
	if len(preds) == 0 {
		preds = append(preds, sqlf.Sprintf("%s", "TRUE"))
	}
	if args.BackfillIncomplete {
		preds = append(preds, sqlf.Sprintf("backfill_queued_at IS NULL"))
	}
	if len(args.SeriesID) > 0 {
		preds = append(preds, sqlf.Sprintf("series_id = %s", args.SeriesID))
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
			&temp.LastSnapshotAt,
			&temp.NextSnapshotAfter,
			&temp.Enabled,
		); err != nil {
			return []types.InsightSeries{}, err
		}
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
			&temp.LineColor,
			&temp.SeriesID,
			&temp.Query,
			&temp.CreatedAt,
			&temp.OldestHistoricalAt,
			&temp.LastRecordedAt,
			&temp.NextRecordingAfter,
			&temp.BackfillQueuedAt,
			&temp.LastSnapshotAt,
			&temp.NextSnapshotAfter,
			pq.Array(&temp.Repositories),
			&temp.SampleIntervalUnit,
			&temp.SampleIntervalValue,
			&temp.DefaultFilterIncludeRepoRegex,
			&temp.DefaultFilterExcludeRepoRegex,
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
func (s *InsightStore) CreateView(ctx context.Context, view types.InsightView, grants []InsightViewGrant) (_ types.InsightView, err error) {
	tx, err := s.Transact(ctx)
	if err != nil {
		return types.InsightView{}, err
	}
	defer func() { err = tx.Done(err) }()

	row := tx.QueryRow(ctx, sqlf.Sprintf(createInsightViewSql,
		view.Title,
		view.Description,
		view.UniqueID,
	))
	if row.Err() != nil {
		return types.InsightView{}, row.Err()
	}
	var id int
	err = row.Scan(&id)
	if err != nil {
		return types.InsightView{}, errors.Wrap(err, "failed to insert insight view")
	}
	view.ID = id
	err = tx.AddViewGrants(ctx, view, grants)
	if err != nil {
		return types.InsightView{}, errors.Wrap(err, "failed to attach view grants")
	}
	return view, nil
}

func (s *InsightStore) AddViewGrants(ctx context.Context, view types.InsightView, grants []InsightViewGrant) error {
	if view.ID == 0 {
		return errors.New("unable to grant view permissions invalid insight view id")
	} else if len(grants) == 0 {
		return nil
	}

	values := make([]*sqlf.Query, 0, len(grants))
	for _, grant := range grants {
		values = append(values, grant.toQuery(view.ID))
	}
	q := sqlf.Sprintf(addViewGrantsSql, sqlf.Join(values, ",\n"))
	err := s.Exec(ctx, q)
	if err != nil {
		return err
	}
	return nil
}

const addViewGrantsSql = `
-- source: enterprise/internal/insights/store/insight_store.go:AddViewGrants
INSERT INTO insight_view_grants (insight_view_id, org_id, user_id, global)
VALUES %s;
`

// DeleteViewByUniqueID deletes an insight view (cascading to dependent child tables) given a unique ID. This operation
// is idempotent and can be executed many times with only one effect or error.
func (s *InsightStore) DeleteViewByUniqueID(ctx context.Context, uniqueID string) error {
	if len(uniqueID) == 0 {
		return errors.New("unable to delete view invalid view ID")
	}
	conds := sqlf.Sprintf("unique_id = %s", uniqueID)
	q := sqlf.Sprintf(deleteViewSql, conds)
	err := s.Exec(ctx, q)
	if err != nil {
		return err
	}
	return nil
}

const deleteViewSql = `
-- source: enterprise/internal/insights/store/insight_store.go:DeleteView
delete from insight_view where %s;
`

// CreateSeries will create a new insight data series. This series must be uniquely identified by the series ID.
func (s *InsightStore) CreateSeries(ctx context.Context, series types.InsightSeries) (types.InsightSeries, error) {
	if series.CreatedAt.IsZero() {
		series.CreatedAt = s.Now()
	}
	if series.NextRecordingAfter.IsZero() {
		series.NextRecordingAfter = s.Now()
	}
	if series.NextSnapshotAfter.IsZero() {
		series.NextSnapshotAfter = s.Now()
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
		series.LastSnapshotAt,
		series.NextSnapshotAfter,
	))
	var id int
	err := row.Scan(&id)
	if err != nil {
		return types.InsightSeries{}, err
	}
	series.ID = id
	series.Enabled = true
	return series, nil
}

type DataSeriesStore interface {
	GetDataSeries(ctx context.Context, args GetDataSeriesArgs) ([]types.InsightSeries, error)
	StampRecording(ctx context.Context, series types.InsightSeries) (types.InsightSeries, error)
	StampSnapshot(ctx context.Context, series types.InsightSeries) (types.InsightSeries, error)
	StampBackfill(ctx context.Context, series types.InsightSeries) (types.InsightSeries, error)
	SetSeriesEnabled(ctx context.Context, seriesId string, enabled bool) error
}

type InsightMetadataStore interface {
	GetMapped(ctx context.Context, args InsightQueryArgs) ([]types.Insight, error)
	GetDirtyQueries(ctx context.Context, series *types.InsightSeries) ([]*types.DirtyQuery, error)
	GetDirtyQueriesAggregated(ctx context.Context, seriesID string) ([]*types.DirtyQueryAggregate, error)
}

// StampRecording will update the recording metadata for this series and return the InsightSeries struct with updated values.
func (s *InsightStore) StampRecording(ctx context.Context, series types.InsightSeries) (types.InsightSeries, error) {
	current := s.Now()
	next := insights.NextRecording(current)
	if err := s.Exec(ctx, sqlf.Sprintf(stampRecordingSql, current, next, series.ID)); err != nil {
		return types.InsightSeries{}, err
	}
	series.LastRecordedAt = current
	series.NextRecordingAfter = next
	return series, nil
}

// StampSnapshot will update the recording metadata for this series and return the InsightSeries struct with updated values.
func (s *InsightStore) StampSnapshot(ctx context.Context, series types.InsightSeries) (types.InsightSeries, error) {
	current := s.Now()
	next := insights.NextSnapshot(current)
	if err := s.Exec(ctx, sqlf.Sprintf(stampSnapshotSql, current, next, series.ID)); err != nil {
		return types.InsightSeries{}, err
	}
	series.LastRecordedAt = current
	series.NextRecordingAfter = next
	return series, nil
}

// StampBackfill will update the backfill queued time for this series and return the InsightSeries struct with updated values.
func (s *InsightStore) StampBackfill(ctx context.Context, series types.InsightSeries) (types.InsightSeries, error) {
	current := s.Now()
	if err := s.Exec(ctx, sqlf.Sprintf(stampBackfillSql, current, series.ID)); err != nil {
		return types.InsightSeries{}, err
	}
	series.BackfillQueuedAt = current
	return series, nil
}

func (s *InsightStore) SetSeriesEnabled(ctx context.Context, seriesId string, enabled bool) error {
	var arg *sqlf.Query
	if enabled {
		arg = sqlf.Sprintf("null")
	} else {
		arg = sqlf.Sprintf("%s", s.Now())
	}
	return s.Exec(ctx, sqlf.Sprintf(setSeriesStatusSql, arg, seriesId))
}

const setSeriesStatusSql = `
-- source: enterprise/internal/insights/store/insight_store.go:SetSeriesStatus
UPDATE insight_series
SET deleted_at = %s
WHERE series_id = %s;
`

const stampBackfillSql = `
-- source: enterprise/internal/insights/store/insight_store.go:StampRecording
UPDATE insight_series
SET backfill_queued_at = %s
WHERE id = %s;
`

const stampRecordingSql = `
-- source: enterprise/internal/insights/store/insight_store.go:StampRecording
UPDATE insight_series
SET last_recorded_at = %s,
    next_recording_after = %s
WHERE id = %s;
`

const stampSnapshotSql = `
-- source: enterprise/internal/insights/store/insight_store.go:StampSnapshot
UPDATE insight_series
SET last_snapshot_at = %s,
    next_snapshot_after = %s
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
                            next_recording_after, last_snapshot_at, next_snapshot_after)
VALUES (%s, %s, %s, %s, %s, %s, %s, %s)
RETURNING id;`

const getInsightByViewSql = `
-- source: enterprise/internal/insights/store/insight_store.go:Get
SELECT iv.unique_id, iv.title, iv.description, ivs.label, ivs.stroke,
i.series_id, i.query, i.created_at, i.oldest_historical_at, i.last_recorded_at,
i.next_recording_after, i.backfill_queued_at, i.last_snapshot_at, i.next_snapshot_after, i.repositories,
i.sample_interval_unit, i.sample_interval_value, iv.default_filter_include_repo_regex, iv.default_filter_exclude_repo_regex
FROM insight_view iv
         JOIN insight_view_series ivs ON iv.id = ivs.insight_view_id
         JOIN insight_series i ON ivs.insight_series_id = i.id
         JOIN insight_view_grants ivg ON iv.id = ivg.insight_view_id
WHERE %s
ORDER BY iv.unique_id, i.series_id
`

const getInsightDataSeriesSql = `
-- source: enterprise/internal/insights/store/insight_store.go:GetDataSeries
select id, series_id, query, created_at, oldest_historical_at, last_recorded_at, next_recording_after, last_snapshot_at, next_snapshot_after, (CASE WHEN deleted_at IS NULL THEN TRUE ELSE FALSE END) AS enabled from insight_series
WHERE %s
`
