package store

import (
	"context"
	"database/sql"
	"sort"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/timeseries"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type InsightStore struct {
	*basestore.Store
	Now func() time.Time
}

// NewInsightStore returns a new InsightStore backed by the given Postgres db.
func NewInsightStore(db dbutil.DB) *InsightStore {
	return &InsightStore{Store: basestore.NewWithDB(db, sql.TxOptions{}), Now: time.Now}
}

// Handle returns the underlying transactable database handle.
// Needed to implement the ShareableStore interface.
func (s *InsightStore) Handle() *basestore.TransactableHandle { return s.Store.Handle() }

// With creates a new InsightStore with the given basestore.Shareable store as the underlying basestore.Store.
// Needed to implement the basestore.Store interface
func (s *InsightStore) With(other basestore.ShareableStore) *InsightStore {
	return &InsightStore{Store: s.Store.With(other), Now: s.Now}
}

func (s *InsightStore) Transact(ctx context.Context) (*InsightStore, error) {
	txBase, err := s.Store.Transact(ctx)
	return &InsightStore{Store: txBase, Now: s.Now}, err
}

// InsightQueryArgs contains query predicates for fetching viewable insight series. Any provided values will be
// included as query arguments.
type InsightQueryArgs struct {
	UniqueIDs   []string
	UniqueID    string
	UserID      []int
	OrgID       []int
	DashboardID int

	After    string
	Limit    int
	IsFrozen *bool

	// This field will disable user level authorization checks on the insight views. This should only be used
	// when fetching insights from a container that also has authorization checks, such as a dashboard.
	WithoutAuthorization bool
}

// Get returns all matching insight series for insights without any other associations (such as dashboards).
func (s *InsightStore) Get(ctx context.Context, args InsightQueryArgs) ([]types.InsightViewSeries, error) {
	preds := make([]*sqlf.Query, 0, 4)
	var viewConditions []*sqlf.Query

	if len(args.UniqueIDs) > 0 {
		elems := make([]*sqlf.Query, 0, len(args.UniqueIDs))
		for _, id := range args.UniqueIDs {
			elems = append(elems, sqlf.Sprintf("%s", id))
		}
		viewConditions = append(viewConditions, sqlf.Sprintf("unique_id IN (%s)", sqlf.Join(elems, ",")))
	}
	if len(args.UniqueID) > 0 {
		viewConditions = append(viewConditions, sqlf.Sprintf("unique_id = %s", args.UniqueID))
	}
	if args.DashboardID > 0 {
		viewConditions = append(viewConditions, sqlf.Sprintf("id in (select insight_view_id from dashboard_insight_view where dashboard_id = %s)", args.DashboardID))
	}
	preds = append(preds, sqlf.Sprintf("i.deleted_at IS NULL"))
	if !args.WithoutAuthorization {
		viewConditions = append(viewConditions, sqlf.Sprintf("id in (%s)", visibleViewsQuery(args.UserID, args.OrgID)))
	}

	cursor := insightViewPageCursor{
		after: args.After,
		limit: args.Limit,
	}

	q := sqlf.Sprintf(getInsightByViewSql, insightViewQuery(cursor, viewConditions), sqlf.Join(preds, "\n AND"))
	return scanInsightViewSeries(s.Query(ctx, q))
}

// GetAll returns all matching viewable insight series for the provided context, including associated insights (dashboards).
func (s *InsightStore) GetAll(ctx context.Context, args InsightQueryArgs) ([]types.InsightViewSeries, error) {
	preds := make([]*sqlf.Query, 0, 5)

	preds = append(preds, sqlf.Sprintf("i.deleted_at IS NULL"))
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
	if args.DashboardID > 0 {
		preds = append(preds, sqlf.Sprintf("iv.id in (select insight_view_id from dashboard_insight_view where dashboard_id = %s)", args.DashboardID))
	}
	if args.After != "" {
		preds = append(preds, sqlf.Sprintf("iv.unique_id > %s", args.After))
	}
	if args.IsFrozen != nil {
		if *args.IsFrozen {
			preds = append(preds, sqlf.Sprintf("iv.is_frozen = TRUE"))
		} else {
			preds = append(preds, sqlf.Sprintf("iv.is_frozen = FALSE"))
		}
	}

	limit := sqlf.Sprintf("")
	if args.Limit > 0 {
		limit = sqlf.Sprintf("LIMIT %d", args.Limit)
	}

	q := sqlf.Sprintf(getInsightIdsVisibleToUserSql,
		visibleDashboardsQuery(args.UserID, args.OrgID),
		visibleViewsQuery(args.UserID, args.OrgID),
		sqlf.Join(preds, "AND"),
		limit)
	insightIds, err := scanInsightViewIds(s.Query(ctx, q))
	if err != nil {
		return nil, err
	}
	if len(insightIds) == 0 {
		return []types.InsightViewSeries{}, nil
	}

	insightIdElems := make([]*sqlf.Query, 0, len(insightIds))
	for _, id := range insightIds {
		insightIdElems = append(insightIdElems, sqlf.Sprintf("%s", id))
	}

	q = sqlf.Sprintf(getInsightsWithSeriesSql, sqlf.Join(insightIdElems, ","))
	return scanInsightViewSeries(s.Query(ctx, q))
}

type InsightsOnDashboardQueryArgs struct {
	DashboardID int
	After       string
	Limit       int
}

// GetAllOnDashboard returns a page of insights on a dashboard
func (s *InsightStore) GetAllOnDashboard(ctx context.Context, args InsightsOnDashboardQueryArgs) ([]types.InsightViewSeries, error) {
	where := make([]*sqlf.Query, 0, 2)
	var limit *sqlf.Query

	where = append(where, sqlf.Sprintf("dbiv.dashboard_id = %s", args.DashboardID))
	if args.After != "" {
		where = append(where, sqlf.Sprintf("dbiv.id > %s", args.After))
	}
	if args.Limit > 0 {
		limit = sqlf.Sprintf("LIMIT %s", args.Limit)
	} else {
		limit = sqlf.Sprintf("")
	}

	q := sqlf.Sprintf(getInsightsByDashboardSql, sqlf.Join(where, "AND"), limit)
	return scanInsightViewSeries(s.Query(ctx, q))
}

// visibleViewsQuery generates the SQL query for filtering insight views based on granted permissions.
// This returns a query that will generate a set of insight_view.id that the provided context can see.
func visibleViewsQuery(userIDs, orgIDs []int) *sqlf.Query {
	permsPreds := make([]*sqlf.Query, 0, 2)
	if len(orgIDs) > 0 {
		elems := make([]*sqlf.Query, 0, len(orgIDs))
		for _, id := range orgIDs {
			elems = append(elems, sqlf.Sprintf("%s", id))
		}
		permsPreds = append(permsPreds, sqlf.Sprintf("org_id IN (%s)", sqlf.Join(elems, ",")))
	}
	if len(userIDs) > 0 {
		elems := make([]*sqlf.Query, 0, len(userIDs))
		for _, id := range userIDs {
			elems = append(elems, sqlf.Sprintf("%s", id))
		}
		permsPreds = append(permsPreds, sqlf.Sprintf("user_id IN (%s)", sqlf.Join(elems, ",")))
	}
	permsPreds = append(permsPreds, sqlf.Sprintf("global is true"))

	return sqlf.Sprintf("SELECT insight_view_id FROM insight_view_grants WHERE %s", sqlf.Join(permsPreds, "OR"))
}

func (s *InsightStore) GetMapped(ctx context.Context, args InsightQueryArgs) ([]types.Insight, error) {
	viewSeries, err := s.Get(ctx, args)
	if err != nil {
		return nil, err
	}

	return s.GroupByView(ctx, viewSeries), nil
}

func (s *InsightStore) GetAllMapped(ctx context.Context, args InsightQueryArgs) ([]types.Insight, error) {
	viewSeries, err := s.GetAll(ctx, args)
	if err != nil {
		return nil, err
	}

	return s.GroupByView(ctx, viewSeries), nil
}

func (s *InsightStore) GroupByView(ctx context.Context, viewSeries []types.InsightViewSeries) []types.Insight {
	mapped := make(map[string][]types.InsightViewSeries, len(viewSeries))
	for _, series := range viewSeries {
		mapped[series.UniqueID] = append(mapped[series.UniqueID], series)
	}

	results := make([]types.Insight, 0, len(mapped))
	for _, seriesSet := range mapped {
		var sortOptions *types.SeriesSortOptions
		// TODO what only one of these is set? I think the idea is that they have to be set together, but it's not enforced
		// in the database..
		if seriesSet[0].SeriesSortMode != nil && seriesSet[0].SeriesSortDirection != nil {
			sortOptions = &types.SeriesSortOptions{
				Mode:      *seriesSet[0].SeriesSortMode,
				Direction: *seriesSet[0].SeriesSortDirection,
			}
		}

		results = append(results, types.Insight{
			ViewID:          seriesSet[0].ViewID,
			DashboardViewId: seriesSet[0].DashboardViewID,
			UniqueID:        seriesSet[0].UniqueID,
			Title:           seriesSet[0].Title,
			Description:     seriesSet[0].Description,
			Series:          seriesSet,
			Filters: types.InsightViewFilters{
				IncludeRepoRegex: seriesSet[0].DefaultFilterIncludeRepoRegex,
				ExcludeRepoRegex: seriesSet[0].DefaultFilterExcludeRepoRegex,
				SearchContexts:   seriesSet[0].DefaultFilterSearchContexts,
			},
			OtherThreshold:   seriesSet[0].OtherThreshold,
			PresentationType: seriesSet[0].PresentationType,
			IsFrozen:         seriesSet[0].IsFrozen,
			SeriesOptions: types.SeriesDisplayOptions{
				SortOptions: sortOptions,
				Limit:       seriesSet[0].SeriesLimit,
			},
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].UniqueID < results[j].UniqueID
	})
	return results
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
	GlobalOnly          bool
	ExcludeJustInTime   bool
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
	if args.GlobalOnly {
		preds = append(preds, sqlf.Sprintf("(repositories IS NULL OR CARDINALITY(repositories) = 0)"))
	}
	if args.ExcludeJustInTime {
		preds = append(preds, sqlf.Sprintf("just_in_time = false"))
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
			&temp.SampleIntervalUnit,
			&temp.SampleIntervalValue,
			&temp.GeneratedFromCaptureGroups,
			&temp.JustInTime,
			&temp.GenerationMethod,
			pq.Array(&temp.Repositories),
		); err != nil {
			return []types.InsightSeries{}, err
		}
		results = append(results, temp)
	}
	return results, nil
}

func scanInsightViewIds(rows *sql.Rows, queryErr error) (_ []string, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	results := make([]string, 0)
	for rows.Next() {
		var temp types.InsightViewSeries
		if err := rows.Scan(
			&temp.UniqueID,
		); err != nil {
			return nil, err
		}
		results = append(results, temp.UniqueID)
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
			&temp.ViewID,
			&temp.DashboardViewID,
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
			&temp.OtherThreshold,
			&temp.PresentationType,
			&temp.GeneratedFromCaptureGroups,
			&temp.JustInTime,
			&temp.GenerationMethod,
			&temp.IsFrozen,
			pq.Array(&temp.DefaultFilterSearchContexts),
			&temp.SeriesSortMode,
			&temp.SeriesSortDirection,
			&temp.SeriesLimit,
		); err != nil {
			return []types.InsightViewSeries{}, err
		}
		results = append(results, temp)
	}
	return results, nil
}

type insightViewPageCursor struct {
	after string
	limit int
}

func insightViewQuery(cursor insightViewPageCursor, viewConditions []*sqlf.Query) *sqlf.Query {
	var cond []*sqlf.Query
	if cursor.after != "" {
		cond = append(cond, sqlf.Sprintf("unique_id > %s", cursor.after))
	} else {
		cond = append(cond, sqlf.Sprintf("TRUE"))
	}
	var limit *sqlf.Query
	if cursor.limit > 0 {
		limit = sqlf.Sprintf("LIMIT %s", cursor.limit)
	} else {
		limit = sqlf.Sprintf("")
	}
	cond = append(cond, viewConditions...)

	q := sqlf.Sprintf(insightViewQuerySql, sqlf.Join(cond, "AND"), limit)
	return q
}

const insightViewQuerySql = `
SELECT * FROM insight_view WHERE %s ORDER BY unique_id %s
`

// AttachSeriesToView will associate a given insight data series with a given insight view.
func (s *InsightStore) AttachSeriesToView(ctx context.Context,
	series types.InsightSeries,
	view types.InsightView,
	metadata types.InsightViewSeriesMetadata) error {
	if series.ID == 0 || view.ID == 0 {
		return errors.New("input series or view not found")
	}
	err := s.Exec(ctx, sqlf.Sprintf(attachSeriesToViewSql, series.ID, view.ID, metadata.Label, metadata.Stroke))
	if err != nil {
		return err
	}
	// Enable the series in case it had previously been soft-deleted.
	err = s.SetSeriesEnabled(ctx, series.SeriesID, true)
	if err != nil {
		return err
	}
	return nil
}

func (s *InsightStore) RemoveSeriesFromView(ctx context.Context, seriesId string, viewId int) error {
	err := s.Exec(ctx, sqlf.Sprintf(removeSeriesFromViewSql, seriesId, viewId))
	if err != nil {
		return err
	}
	// Delete the series if there are no longer any references to it.
	count, _, err := basestore.ScanFirstInt(s.Query(ctx, sqlf.Sprintf(countSeriesReferencesSql, seriesId)))
	if err != nil {
		return err
	}
	if count != 0 {
		return nil
	}
	err = s.SetSeriesEnabled(ctx, seriesId, false)
	if err != nil {
		return err
	}
	return nil
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
		view.Filters.IncludeRepoRegex,
		view.Filters.ExcludeRepoRegex,
		pq.Array(view.Filters.SearchContexts),
		view.OtherThreshold,
		view.PresentationType,
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

func (s *InsightStore) UpdateView(ctx context.Context, view types.InsightView) (_ types.InsightView, err error) {
	tx, err := s.Transact(ctx)
	if err != nil {
		return types.InsightView{}, err
	}
	defer func() { err = tx.Done(err) }()

	row := tx.QueryRow(ctx, sqlf.Sprintf(updateInsightViewSql,
		view.Title,
		view.Description,
		view.Filters.IncludeRepoRegex,
		view.Filters.ExcludeRepoRegex,
		pq.Array(view.Filters.SearchContexts),
		view.OtherThreshold,
		view.PresentationType,
		view.SeriesSortMode,
		view.SeriesSortDirection,
		view.SeriesLimit,
		view.UniqueID,
	))
	var id int
	err = row.Scan(&id)
	if err != nil {
		return types.InsightView{}, errors.Wrap(err, "failed to update insight view")
	}
	view.ID = id
	return view, nil
}

func (s *InsightStore) UpdateViewSeries(ctx context.Context, seriesId string, viewId int, metadata types.InsightViewSeriesMetadata) error {
	return s.Exec(ctx, sqlf.Sprintf(updateInsightViewSeries, metadata.Label, metadata.Stroke, seriesId, viewId))
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
	interval := timeseries.TimeInterval{
		Unit:  types.IntervalUnit(series.SampleIntervalUnit),
		Value: series.SampleIntervalValue,
	}
	if !interval.IsValid() {
		interval = timeseries.DefaultInterval
	}

	if series.NextRecordingAfter.IsZero() {
		series.NextRecordingAfter = interval.StepForwards(s.Now())
	}
	if series.NextSnapshotAfter.IsZero() {
		series.NextSnapshotAfter = NextSnapshot(s.Now())
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
		pq.Array(series.Repositories),
		series.SampleIntervalUnit,
		series.SampleIntervalValue,
		series.GeneratedFromCaptureGroups,
		series.JustInTime,
		series.GenerationMethod,
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
	next := timeseries.TimeInterval{
		Unit:  types.IntervalUnit(series.SampleIntervalUnit),
		Value: series.SampleIntervalValue,
	}.StepForwards(current)
	if err := s.Exec(ctx, sqlf.Sprintf(stampRecordingSql, current, next, series.ID)); err != nil {
		return types.InsightSeries{}, err
	}
	series.LastRecordedAt = current
	series.NextRecordingAfter = next
	return series, nil
}

func NextSnapshot(current time.Time) time.Time {
	year, month, day := current.In(time.UTC).Date()
	return time.Date(year, month, day+1, 0, 0, 0, 0, time.UTC)
}

// StampSnapshot will update the recording metadata for this series and return the InsightSeries struct with updated values.
func (s *InsightStore) StampSnapshot(ctx context.Context, series types.InsightSeries) (types.InsightSeries, error) {
	current := s.Now()
	next := NextSnapshot(current)
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

type MatchSeriesArgs struct {
	Query                     string
	StepIntervalUnit          string
	StepIntervalValue         int
	GenerateFromCaptureGroups bool
}

func (s *InsightStore) FindMatchingSeries(ctx context.Context, args MatchSeriesArgs) (_ types.InsightSeries, found bool, _ error) {
	where := sqlf.Sprintf(
		"(repositories = '{}' OR repositories is NULL) AND query = %s AND sample_interval_unit = %s AND sample_interval_value = %s AND generated_from_capture_groups = %s",
		args.Query, args.StepIntervalUnit, args.StepIntervalValue, args.GenerateFromCaptureGroups,
	)

	q := sqlf.Sprintf(getInsightDataSeriesSql, where)
	rows, err := scanDataSeries(s.Query(ctx, q))
	if err != nil {
		return types.InsightSeries{}, false, err
	}
	if len(rows) == 0 {
		return types.InsightSeries{}, false, nil
	}
	return rows[0], true, nil
}

type UpdateFrontendSeriesArgs struct {
	SeriesID          string
	Query             string
	Repositories      []string
	StepIntervalUnit  string
	StepIntervalValue int
}

func (s *InsightStore) UpdateFrontendSeries(ctx context.Context, args UpdateFrontendSeriesArgs) error {
	return s.Exec(ctx, sqlf.Sprintf(updateFrontendSeriesSql, args.Query, pq.Array(args.Repositories), args.StepIntervalUnit, args.StepIntervalValue, args.SeriesID))
}

func (s *InsightStore) GetReferenceCount(ctx context.Context, id int) (int, error) {
	count, _, err := basestore.ScanFirstInt(s.Query(ctx, sqlf.Sprintf(getReferenceCountSql, id)))
	if err != nil {
		return 0, nil
	}
	return count, nil
}

func (s *InsightStore) GetSoftDeletedSeries(ctx context.Context, deletedBefore time.Time) ([]string, error) {
	return basestore.ScanStrings(s.Query(ctx, sqlf.Sprintf(getSoftDeletedSeries, deletedBefore)))
}

func (s *InsightStore) HardDeleteSeries(ctx context.Context, seriesId string) error {
	return s.Exec(ctx, sqlf.Sprintf(hardDeleteSeries, seriesId))
}

func (s *InsightStore) GetUnfrozenInsightCount(ctx context.Context) (globalCount int, totalCount int, err error) {
	rows := s.QueryRow(ctx, sqlf.Sprintf(getUnfrozenInsightCountSql))
	err = rows.Scan(
		&globalCount,
		&totalCount,
	)
	return
}

func (s *InsightStore) GetUnfrozenInsightUniqueIds(ctx context.Context) ([]string, error) {
	return basestore.ScanStrings(s.Query(ctx, sqlf.Sprintf(getUnfrozenInsightUniqueIdsSql)))
}

func (s *InsightStore) FreezeAllInsights(ctx context.Context) error {
	return s.Exec(ctx, sqlf.Sprintf(freezeAllInsightsSql))
}

func (s *InsightStore) UnfreezeAllInsights(ctx context.Context) error {
	return s.Exec(ctx, sqlf.Sprintf(unfreezeAllInsightsSql))
}

func (s *InsightStore) UnfreezeGlobalInsights(ctx context.Context, count int) error {
	return s.Exec(ctx, sqlf.Sprintf(unfreezeGlobalInsightsSql, count))
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

const removeSeriesFromViewSql = `
-- source: enterprise/internal/insights/store/insight_store.go:RemoveSeriesFromView
DELETE FROM insight_view_series vs
USING insight_series s
WHERE s.series_id = %s AND vs.insight_series_id = s.id AND vs.insight_view_id = %s;
`

const updateInsightViewSeries = `
-- source: enterprise/internal/insights/store/insight_store.go:UpdateViewSeries
UPDATE insight_view_series vs
SET label = %s, stroke = %s
FROM insight_series s
WHERE s.series_id = %s AND vs.insight_series_id = s.id AND vs.insight_view_id = %s
`

const createInsightViewSql = `
-- source: enterprise/internal/insights/store/insight_store.go:CreateView
INSERT INTO insight_view (title, description, unique_id, default_filter_include_repo_regex, default_filter_exclude_repo_regex,
default_filter_search_contexts, other_threshold, presentation_type)
VALUES (%s, %s, %s, %s, %s, %s, %s, %s)
returning id;`

const updateInsightViewSql = `
-- source: enterprise/internal/insights/store/insight_store.go:UpdateView
UPDATE insight_view SET title = %s, description = %s, default_filter_include_repo_regex = %s, default_filter_exclude_repo_regex = %s,
default_filter_search_contexts = %s, other_threshold = %s, presentation_type = %s, series_sort_mode = %s, series_sort_direction = %s,
series_limit = %s
WHERE unique_id = %s
RETURNING id;`

const createInsightSeriesSql = `
-- source: enterprise/internal/insights/store/insight_store.go:CreateSeries
INSERT INTO insight_series (series_id, query, created_at, oldest_historical_at, last_recorded_at,
                            next_recording_after, last_snapshot_at, next_snapshot_after, repositories,
							sample_interval_unit, sample_interval_value, generated_from_capture_groups,
							just_in_time, generation_method)
VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
RETURNING id;`

const getInsightByViewSql = `
-- source: enterprise/internal/insights/store/insight_store.go:Get
SELECT iv.id, 0 as dashboard_insight_id, iv.unique_id, iv.title, iv.description, ivs.label, ivs.stroke,
i.series_id, i.query, i.created_at, i.oldest_historical_at, i.last_recorded_at,
i.next_recording_after, i.backfill_queued_at, i.last_snapshot_at, i.next_snapshot_after, i.repositories,
i.sample_interval_unit, i.sample_interval_value, iv.default_filter_include_repo_regex, iv.default_filter_exclude_repo_regex,
iv.other_threshold, iv.presentation_type, i.generated_from_capture_groups, i.just_in_time, i.generation_method, iv.is_frozen,
default_filter_search_contexts, iv.series_sort_mode, iv.series_sort_direction, iv.series_limit
FROM (%s) iv
         JOIN insight_view_series ivs ON iv.id = ivs.insight_view_id
         JOIN insight_series i ON ivs.insight_series_id = i.id
WHERE %s
ORDER BY iv.id, i.series_id
`

const getInsightsByDashboardSql = `
SELECT iv.id, dbiv.id as dashboard_insight_id, iv.unique_id, iv.title, iv.description, ivs.label, ivs.stroke,
i.series_id, i.query, i.created_at, i.oldest_historical_at, i.last_recorded_at,
i.next_recording_after, i.backfill_queued_at, i.last_snapshot_at, i.next_snapshot_after, i.repositories,
i.sample_interval_unit, i.sample_interval_value, iv.default_filter_include_repo_regex, iv.default_filter_exclude_repo_regex,
iv.other_threshold, iv.presentation_type, i.generated_from_capture_groups, i.just_in_time, i.generation_method, iv.is_frozen,
default_filter_search_contexts, iv.series_sort_mode, iv.series_sort_direction, iv.series_limit
FROM dashboard_insight_view as dbiv
		 JOIN insight_view iv ON iv.id = dbiv.insight_view_id
         JOIN insight_view_series ivs ON iv.id = ivs.insight_view_id
         JOIN insight_series i ON ivs.insight_series_id = i.id
WHERE %s
ORDER BY dbiv.id
%s;
`

const getInsightDataSeriesSql = `
-- source: enterprise/internal/insights/store/insight_store.go:GetDataSeries
select id, series_id, query, created_at, oldest_historical_at, last_recorded_at, next_recording_after,
last_snapshot_at, next_snapshot_after, (CASE WHEN deleted_at IS NULL THEN TRUE ELSE FALSE END) AS enabled,
sample_interval_unit, sample_interval_value, generated_from_capture_groups,
just_in_time, generation_method, repositories
from insight_series
WHERE %s
`

const getInsightIdsVisibleToUserSql = `
SELECT DISTINCT iv.unique_id
FROM insight_view iv
JOIN insight_view_series ivs ON iv.id = ivs.insight_view_id
JOIN insight_series i ON ivs.insight_series_id = i.id
WHERE (iv.id IN (SELECT insight_view_id
			 FROM dashboard db
			 JOIN dashboard_insight_view div ON db.id = div.dashboard_id
				 WHERE deleted_at IS NULL AND db.id IN (%s))
   OR iv.id IN (%s))
AND %s
ORDER BY iv.unique_id
%s
`

const getInsightsWithSeriesSql = `
-- source: enterprise/internal/insights/store/insight_store.go:GetAllInsights
SELECT iv.id, 0 as dashboard_insight_id, iv.unique_id, iv.title, iv.description, ivs.label, ivs.stroke,
       i.series_id, i.query, i.created_at, i.oldest_historical_at, i.last_recorded_at,
       i.next_recording_after, i.backfill_queued_at, i.last_snapshot_at, i.next_snapshot_after, i.repositories,
       i.sample_interval_unit, i.sample_interval_value, iv.default_filter_include_repo_regex, iv.default_filter_exclude_repo_regex,
	   iv.other_threshold, iv.presentation_type, i.generated_from_capture_groups, i.just_in_time, i.generation_method, iv.is_frozen,
default_filter_search_contexts, iv.series_sort_mode, iv.series_sort_direction, iv.series_limit
FROM insight_view iv
JOIN insight_view_series ivs ON iv.id = ivs.insight_view_id
JOIN insight_series i ON ivs.insight_series_id = i.id
WHERE iv.unique_id IN (%s)
ORDER BY iv.unique_id
`

const countSeriesReferencesSql = `
-- source: enterprise/internal/insights/store/insight_store.go:CountSeriesReferences
SELECT COUNT(*) FROM insight_view_series viewSeries
	INNER JOIN insight_series series ON viewSeries.insight_series_id = series.id
WHERE series.series_id = %s
`

const updateFrontendSeriesSql = `
-- source: enterprise/internal/insights/store/insight_store.go:UpdateFrontendSeries
UPDATE insight_series
SET query = %s, repositories = %s, sample_interval_unit = %s, sample_interval_value = %s
WHERE series_id = %s
`

const getReferenceCountSql = `
-- source: enterprise/internal/insights/store/insight_store.go:GetReferenceCount
SELECT COUNT(*) from dashboard_insight_view
WHERE insight_view_id = %s
`

const getSoftDeletedSeries = `
-- source: enterprise/internal/insights/store/insight_store.go:GetSoftDeletedSeries
SELECT series_id
FROM insight_series i
LEFT JOIN insight_view_series ivs ON i.id = ivs.insight_series_id
WHERE i.deleted_at IS NOT NULL
  AND i.deleted_at < %s
  AND ivs.insight_series_id IS NULL;
`

const hardDeleteSeries = `
-- source: enterprise/internal/insights/store/insight_store.go:HardDeleteSeries
DELETE FROM insight_series WHERE series_id = %s;
`

const freezeAllInsightsSql = `
-- source: enterprise/internal/insights/store/insight_store.go:FreezeAllInsights
UPDATE insight_view SET is_frozen = TRUE
`

const unfreezeAllInsightsSql = `
-- source: enterprise/internal/insights/store/insight_store.go:UnfreezeAllInsights
UPDATE insight_view SET is_frozen = FALSE
`

const getUnfrozenInsightCountSql = `
-- source: enterprise/internal/insights/store/insight_store.go:GetFrozenInsightCounts
SELECT unfrozenGlobal.total as unfrozenGlobal, unfrozenTotal.total as unfrozenTotal FROM (
	SELECT COUNT(DISTINCT(iv.id)) as total from insight_view as iv
	JOIN dashboard_insight_view as d on iv.id = d.insight_view_id
	JOIN dashboard_grants as dg on d.dashboard_id = dg.dashboard_id
	WHERE iv.is_frozen = FALSE AND dg.global = TRUE
) as unfrozenGlobal
CROSS JOIN
(
	SELECT COUNT(DISTINCT(iv.id)) as total from insight_view as iv
	WHERE iv.is_frozen = FALSE
) as unfrozenTotal;
`

const unfreezeGlobalInsightsSql = `
-- source: enterprise/internal/insights/store/insight_store.go:UnfreezeGlobalInsights
UPDATE insight_view SET is_frozen = FALSE
WHERE id IN (
	SELECT DISTINCT(iv.id) from insight_view as iv
	JOIN dashboard_insight_view as d on iv.id = d.insight_view_id
	JOIN dashboard_grants as dg on d.dashboard_id = dg.dashboard_id
	WHERE dg.global = TRUE
	ORDER BY iv.id ASC
	LIMIT %s
)
`

const getUnfrozenInsightUniqueIdsSql = `
-- source: enterprise/internal/insights/store/insight_store.go:UnfreezeGlobalInsights
SELECT unique_id FROM insight_view WHERE is_frozen = FALSE;
`
