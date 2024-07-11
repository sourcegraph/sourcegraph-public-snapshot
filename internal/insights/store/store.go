package store

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/lib/pq"
	"strconv"
	"strings"
	"time"

	"github.com/RoaringBitmap/roaring"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/api"
	edb "github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/batch"
	"github.com/sourcegraph/sourcegraph/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Interface is the interface describing a code insights store. See the Store struct
// for actual API usage.
type Interface interface {
	WithOther(other basestore.ShareableStore) Interface
	SeriesPoints(ctx context.Context, opts SeriesPointsOpts) ([]SeriesPoint, error)
	CountData(ctx context.Context, opts CountDataOpts) (int, error)
	RecordSeriesPoints(ctx context.Context, pts []RecordSeriesPointArgs) error
	RecordSeriesPointsAndRecordingTimes(ctx context.Context, pts []RecordSeriesPointArgs, recordingTimes types.InsightSeriesRecordingTimes) error
	SetInsightSeriesRecordingTimes(ctx context.Context, recordingTimes []types.InsightSeriesRecordingTimes) error
	GetInsightSeriesRecordingTimes(ctx context.Context, id int, opts SeriesPointsOpts) (types.InsightSeriesRecordingTimes, error)
	LoadIncompleteDatapoints(ctx context.Context, seriesID int) (results []IncompleteDatapoint, err error)
	AddIncompleteDatapoint(ctx context.Context, input AddIncompleteDatapointInput) error
	GetAllDataForInsightViewID(ctx context.Context, opts ExportOpts) ([]SeriesPointForExport, error)
}

var _ Interface = &Store{}

// Store exposes methods to read and write code insights domain models from
// persistent storage.
type Store struct {
	*basestore.Store
	now       func() time.Time
	permStore InsightPermissionStore
}

func (s *Store) Transact(ctx context.Context) (*Store, error) {
	txBase, err := s.Store.Transact(ctx)
	if err != nil {
		return nil, err
	}
	return &Store{
		Store:     txBase,
		now:       s.now,
		permStore: s.permStore,
	}, nil
}

// New returns a new Store backed by the given Postgres db.
func New(db edb.InsightsDB, permStore InsightPermissionStore) *Store {
	return NewWithClock(db, permStore, timeutil.Now)
}

// NewWithClock returns a new Store backed by the given db and
// clock for timestamps.
func NewWithClock(db edb.InsightsDB, permStore InsightPermissionStore, clock func() time.Time) *Store {
	return &Store{Store: basestore.NewWithHandle(db.Handle()), now: clock, permStore: permStore}
}

var _ basestore.ShareableStore = &Store{}

// With creates a new Store with the given basestore.Shareable store as the
// underlying basestore.Store.
// Needed to implement the basestore.Store interface
func (s *Store) With(other basestore.ShareableStore) *Store {
	return &Store{Store: s.Store.With(other), now: s.now, permStore: s.permStore}
}

// WithOther creates a new Store with the given basestore.Shareable store as the
// underlying basestore.Store.
// Needed to implement the basestore.Store interface
func (s *Store) WithOther(other basestore.ShareableStore) Interface {
	return &Store{Store: s.Store.With(other), now: s.now, permStore: s.permStore}
}

// SeriesPoint describes a single insights' series data point.
//
// Some fields that could be queried (series ID, repo ID/names) are omitted as they are primarily
// only useful for filtering the data you get back, and would inflate the data size considerably
// otherwise.
type SeriesPoint struct {
	// Time (always UTC).
	SeriesID string
	Time     time.Time
	Value    float64
	Capture  *string
}

func (s *SeriesPoint) String() string {
	if s.Capture != nil {
		return fmt.Sprintf("SeriesPoint{Time: %q, Capture: %q, Value: %v}", s.Time, *s.Capture, s.Value)
	}
	return fmt.Sprintf("SeriesPoint{Time: %q, Value: %v}", s.Time, s.Value)
}

// SeriesPointsOpts describes options for querying insights' series data points.
type SeriesPointsOpts struct {
	// SeriesID is the unique series ID to query, if non-nil.
	SeriesID *string
	// ID is the unique integer series ID to query, if non-nil.
	ID *int

	// RepoID, if non-nil, indicates to filter results to only points recorded with this repo ID.
	RepoID *api.RepoID

	Excluded []api.RepoID
	Included []api.RepoID

	// TODO(slimsag): Add ability to filter based on repo name, original name.

	IncludeRepoRegex []string
	ExcludeRepoRegex []string

	// Time ranges to query from/to (inclusive) or after (exclusive), if non-nil, in UTC.
	From, To, After *time.Time

	// Whether to augment the series points data with zero values.
	SupportsAugmentation bool

	// Limit is the number of data points to query, if non-zero.
	Limit int
}

// SeriesPoints queries data points over time for a specific insights' series.
func (s *Store) SeriesPoints(ctx context.Context, opts SeriesPointsOpts) ([]SeriesPoint, error) {
	points := make([]SeriesPoint, 0, opts.Limit)
	// 🚨 SECURITY: This is a double-negative repo permission enforcement. The list of authorized repos is generally expected to be very large, and nearly the full
	// set of repos installed on Sourcegraph. To make this faster, we query Postgres for a list of repos the current user cannot see, and then exclude those from the
	// time series results. 🚨
	// We think this is faster for a few reasons:
	//
	// 1. Any repos set 'public' show for everyone, and this is the default state without configuring otherwise
	// 2. We have quite a bit of customer feedback that suggests they don't even use repo permissions - they just don't install their private repos onto that Sourcegraph instance.
	// 3. Cloud will likely be one of best case scenarios for this - currently we have indexed 550k+ repos all of which are public. Even if we add 20,000 private repos that's only ~3.5% of the total set that needs to be fetched to do this authorization filter.
	//
	// Since Code Insights is in a different database, we can't trivially join the repo table directly, so this approach is preferred.

	denylist, err := s.permStore.GetUnauthorizedRepoIDs(ctx)
	if err != nil {
		return []SeriesPoint{}, err
	}
	opts.Excluded = append(opts.Excluded, denylist...)

	q := seriesPointsQuery(fullVectorSeriesAggregation, opts)
	pointsMap := make(map[string]*SeriesPoint)
	captureValues := make(map[string]struct{})
	err = s.query(ctx, q, func(sc scanner) error {
		var point SeriesPoint
		err := sc.Scan(
			&point.SeriesID,
			&point.Time,
			&point.Value,
			&point.Capture,
		)
		if err != nil {
			return err
		}
		points = append(points, point)
		capture := ""
		if point.Capture != nil {
			capture = *point.Capture
		}
		captureValues[capture] = struct{}{}
		pointsMap[point.Time.String()+capture] = &point
		return nil
	})
	if err != nil {
		return nil, err
	}

	augmentedPoints, err := s.augmentSeriesPoints(ctx, opts, pointsMap, captureValues)
	if err != nil {
		return nil, errors.Wrap(err, "augmentSeriesPoints")
	}
	if len(augmentedPoints) > 0 {
		points = augmentedPoints
	}

	return points, nil
}

func (s *Store) LoadSeriesInMem(ctx context.Context, opts SeriesPointsOpts) (points []SeriesPoint, err error) {
	denylist, err := s.permStore.GetUnauthorizedRepoIDs(ctx)
	if err != nil {
		return nil, err
	}
	denyBitmap := roaring.New()
	for _, id := range denylist {
		denyBitmap.Add(uint32(id))
	}

	type loadStruct struct {
		Time    time.Time
		Value   float64
		RepoID  int
		Capture *string
	}
	type captureMap map[string]*SeriesPoint
	mapping := make(map[time.Time]captureMap)

	getByKey := func(time time.Time, key *string) *SeriesPoint {
		cm, ok := mapping[time]
		if !ok {
			cm = make(captureMap)
			mapping[time] = cm
		}
		k := ""
		if key != nil {
			k = *key
		}
		v, found := cm[k]
		if !found {
			v = &SeriesPoint{}
			cm[k] = v
		}
		return v
	}

	filter := func(id int) bool {
		return denyBitmap.Contains(uint32(id))
	}

	q := `select date_trunc('seconds', sp.time) AS interval_time, max(value), repo_id, capture FROM (
					select * from series_points
					union all
					select * from series_points_snapshots
					) as sp
			  %s
	          where %s
			  GROUP BY sp.series_id, interval_time, sp.repo_id, capture
	;`
	fullQ := seriesPointsQuery(q, opts)
	err = s.query(ctx, fullQ, func(sc scanner) (err error) {
		var row loadStruct
		err = sc.Scan(
			&row.Time,
			&row.Value,
			&row.RepoID,
			&row.Capture,
		)
		if err != nil {
			return err
		}
		if filter(row.RepoID) {
			return nil
		}

		sp := getByKey(row.Time, row.Capture)
		sp.Capture = row.Capture
		sp.Value += row.Value
		sp.Time = row.Time

		return nil
	})

	if err != nil {
		return nil, err
	}

	pointsMap := make(map[string]*SeriesPoint)
	captureValues := make(map[string]struct{})

	for _, pointTime := range mapping {
		for _, point := range pointTime {
			pt := SeriesPoint{
				SeriesID: *opts.SeriesID,
				Time:     point.Time,
				Value:    point.Value,
				Capture:  point.Capture,
			}
			points = append(points, pt)
			capture := ""
			if point.Capture != nil {
				capture = *point.Capture
			}
			captureValues[capture] = struct{}{}
			pointsMap[point.Time.String()+capture] = &pt
		}
	}

	augmentedPoints, err := s.augmentSeriesPoints(ctx, opts, pointsMap, captureValues)
	if err != nil {
		return nil, errors.Wrap(err, "augmentSeriesPoints")
	}
	if len(augmentedPoints) > 0 {
		points = augmentedPoints
	}

	return points, err
}

// Delete will delete the time series data for a particular series_id. This will hard (permanently) delete the data.
func (s *Store) Delete(ctx context.Context, seriesId string) (err error) {
	tx, err := s.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	err = tx.Exec(ctx, sqlf.Sprintf(deleteForSeries, seriesId))
	if err != nil {
		return errors.Wrap(err, "DeleteForSeries")
	}
	err = tx.Exec(ctx, sqlf.Sprintf(deleteForSeriesSnapshots, seriesId))
	if err != nil {
		return errors.Wrap(err, "DeleteForSeriesSnapshots")
	}

	return nil
}

const deleteForSeries = `
DELETE FROM series_points where series_id = %s;
`

const deleteForSeriesSnapshots = `
DELETE FROM series_points_snapshots where series_id = %s;
`

// Note: the inner query could return duplicate points on its own if we merely did a SUM(value) over
// all desired repositories. By using the sub-query, we select the per-repository maximum (thus
// eliminating duplicate points that might have been recorded in a given interval for a given repository)
// and then SUM the result for each repository, giving us our final total number.
const fullVectorSeriesAggregation = `
SELECT sub.series_id, sub.interval_time, SUM(sub.value) as value, sub.capture FROM (
	SELECT sp.repo_name_id, sp.series_id, date_trunc('seconds', sp.time) AS interval_time, MAX(value) as value, capture
	FROM (  select * from series_points
			union all
			select * from series_points_snapshots
	) AS sp
	%s
	WHERE %s
	GROUP BY sp.series_id, interval_time, sp.repo_name_id, capture
	ORDER BY sp.series_id, interval_time, sp.repo_name_id
) sub
GROUP BY sub.series_id, sub.interval_time, sub.capture
ORDER BY sub.series_id, sub.interval_time ASC
`

// Note that the series_points table may contain duplicate points, or points recorded at irregular
// intervals. In specific:
//
//  1. Multiple points recorded at the same time T for cardinality C will be considered part of the same vector.
//     For example, series S and repos R1, R2 have a point at time T. The sum over R1,R2 at T will give the
//     aggregated sum for that series at time T.
//  2. Rarely, it may contain duplicate data points due to the at-least once semantics of query execution.
//     This will cause some jitter in the aggregated series, and will skew the results slightly.
//  3. Searches may not complete at the same exact time, so even in a perfect world if the interval
//     should be 12h it may be off by a minute or so.
func seriesPointsQuery(baseQuery string, opts SeriesPointsOpts) *sqlf.Query {
	preds := seriesPointsPredicates(opts)
	limitClause := ""
	if opts.Limit > 0 {
		limitClause = fmt.Sprintf("LIMIT %d", opts.Limit)
	}
	joinClause := " "
	if len(opts.IncludeRepoRegex) > 0 || len(opts.ExcludeRepoRegex) > 0 {
		joinClause = ` JOIN repo_names rn ON sp.repo_name_id = rn.id `
	}
	if len(opts.Excluded) > 0 {
		excludedStrings := []string{}
		for _, id := range opts.Excluded {
			excludedStrings = append(excludedStrings, strconv.Itoa(int(id)))
		}

		excludeReposJoin := ` LEFT JOIN ( select unnest('{%s}'::_int4) as excluded_repo ) perm
			ON sp.repo_id = perm.excluded_repo `

		joinClause = joinClause + fmt.Sprintf(excludeReposJoin, strings.Join(excludedStrings, ","))
	}

	queryWithJoin := fmt.Sprintf(baseQuery, joinClause, `%s`) // this is a little janky
	return sqlf.Sprintf(
		queryWithJoin+limitClause,
		sqlf.Join(preds, "\n AND "),
	)
}

func seriesPointsPredicates(opts SeriesPointsOpts) []*sqlf.Query {
	preds := []*sqlf.Query{}

	if opts.SeriesID != nil {
		preds = append(preds, sqlf.Sprintf("series_id = %s", *opts.SeriesID))
	}
	if opts.RepoID != nil {
		preds = append(preds, sqlf.Sprintf("repo_id = %d", int32(*opts.RepoID)))
	}
	if opts.From != nil {
		preds = append(preds, sqlf.Sprintf("time >= %s", *opts.From))
	}
	if opts.To != nil {
		preds = append(preds, sqlf.Sprintf("time <= %s", *opts.To))
	}
	if opts.After != nil {
		preds = append(preds, sqlf.Sprintf("time > %s", *opts.After))
	}

	if len(opts.Included) > 0 {
		s := fmt.Sprintf("repo_id = any(%v)", values(opts.Included))
		preds = append(preds, sqlf.Sprintf(s))
	}
	if len(opts.Excluded) > 0 {
		preds = append(preds, sqlf.Sprintf("perm.excluded_repo IS NULL"))
	}
	if len(opts.IncludeRepoRegex) > 0 {
		includePreds := []*sqlf.Query{}
		for _, regex := range opts.IncludeRepoRegex {
			if len(regex) == 0 {
				continue
			}
			includePreds = append(includePreds, sqlf.Sprintf("rn.name ~ %s", regex))
		}
		if len(includePreds) > 0 {
			includes := sqlf.Sprintf("(%s)", sqlf.Join(includePreds, "OR"))
			preds = append(preds, includes)
		}

	}
	if len(opts.ExcludeRepoRegex) > 0 {
		for _, regex := range opts.ExcludeRepoRegex {
			if len(regex) == 0 {
				continue
			}
			preds = append(preds, sqlf.Sprintf("rn.name !~ %s", regex))
		}
	}

	if len(preds) == 0 {
		preds = append(preds, sqlf.Sprintf("TRUE"))
	}
	return preds
}

// values constructs a SQL values statement out of an array of repository ids
func values(ids []api.RepoID) string {
	if len(ids) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("VALUES ")
	for _, repoID := range ids {
		_, err := fmt.Fprintf(&b, "(%v),", repoID)
		if err != nil {
			return ""
		}
	}
	query := b.String()
	query = query[:b.Len()-1] // remove the trailing comma
	return query
}

type CountDataOpts struct {
	// The time range to look for data, if non-nil.
	From, To *time.Time

	// SeriesID, if non-nil, indicates to look for data with this series ID only.
	SeriesID *string

	// RepoID, if non-nil, indicates to look for data with this repo ID only.
	RepoID *api.RepoID
}

// CountData counts the amount of data points in a given time range.
func (s *Store) CountData(ctx context.Context, opts CountDataOpts) (int, error) {
	count, ok, err := basestore.ScanFirstInt(s.Store.Query(ctx, countDataQuery(opts)))
	if err != nil {
		return 0, errors.Wrap(err, "ScanFirstInt")
	}
	if !ok {
		return 0, errors.Wrap(err, "count row not found (this should never happen)")
	}
	return count, nil
}

const countDataFmtstr = `
SELECT COUNT(*) FROM series_points WHERE %s
`

func countDataQuery(opts CountDataOpts) *sqlf.Query {
	preds := []*sqlf.Query{}
	if opts.From != nil {
		preds = append(preds, sqlf.Sprintf("time >= %s", *opts.From))
	}
	if opts.To != nil {
		preds = append(preds, sqlf.Sprintf("time <= %s", *opts.To))
	}
	if opts.SeriesID != nil {
		preds = append(preds, sqlf.Sprintf("series_id = %s", *opts.SeriesID))
	}
	if opts.RepoID != nil {
		preds = append(preds, sqlf.Sprintf("repo_id = %d", int32(*opts.RepoID)))
	}
	if len(preds) == 0 {
		preds = append(preds, sqlf.Sprintf("TRUE"))
	}
	return sqlf.Sprintf(
		countDataFmtstr,
		sqlf.Join(preds, "\n AND "),
	)
}

func (s *Store) DeleteSnapshots(ctx context.Context, series *types.InsightSeries) error {
	if series == nil {
		return errors.New("invalid input for Delete Snapshots")
	}
	err := s.Exec(ctx, sqlf.Sprintf(deleteSnapshotsSql, sqlf.Sprintf(snapshotsTable), series.SeriesID))
	if err != nil {
		return errors.Wrapf(err, "failed to delete insights snapshots for series_id: %s", series.SeriesID)
	}
	err = s.Exec(ctx, sqlf.Sprintf(deleteSnapshotRecordingTimeSql, series.ID))
	if err != nil {
		return errors.Wrapf(err, "failed to delete snapshot recording time for series_id %d", series.ID)
	}
	return nil
}

const deleteSnapshotsSql = `
DELETE FROM %s WHERE series_id = %s;
`

const deleteSnapshotRecordingTimeSql = `
DELETE FROM insight_series_recording_times WHERE insight_series_id = %s and snapshot = true;
`

type PersistMode string

const (
	RecordMode          PersistMode = "record"
	SnapshotMode        PersistMode = "snapshot"
	recordingTable      string      = "series_points"
	snapshotsTable      string      = "series_points_snapshots"
	recordingTimesTable string      = "insight_series_recording_times"

	recordingTableArchive      string = "archived_series_points"
	recordingTimesTableArchive string = "archived_insight_series_recording_times"
)

// RecordSeriesPointArgs describes arguments for the RecordSeriesPoint method.
type RecordSeriesPointArgs struct {
	// SeriesID is the unique series ID to query. It should describe the series of data uniquely,
	// but is not a DB table primary key ID.
	SeriesID string

	// Point is the actual data point recorded and at what time.
	Point SeriesPoint

	// Repository name and DB ID to associate with this data point, if any.
	//
	// Both must be specified if one is specified.
	RepoName *string
	RepoID   *api.RepoID

	PersistMode PersistMode
}

// RecordSeriesPoints stores multiple data points atomically. Use this in favour of RecordSeriesPointsAndRecordingTimes
// if recording times are not known.
func (s *Store) RecordSeriesPoints(ctx context.Context, pts []RecordSeriesPointArgs) (err error) {
	tx, err := s.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	tableColumns := []string{"series_id", "time", "value", "repo_id", "repo_name_id", "original_repo_name_id", "capture"}

	// In our current use cases we should only ever use one of these for one function call, but this could change.
	inserters := map[PersistMode]*batch.Inserter{
		RecordMode:   batch.NewInserter(ctx, tx.Handle(), recordingTable, batch.MaxNumPostgresParameters, tableColumns...),
		SnapshotMode: batch.NewInserter(ctx, tx.Handle(), snapshotsTable, batch.MaxNumPostgresParameters, tableColumns...),
	}

	for _, pt := range pts {
		inserter, ok := inserters[pt.PersistMode]
		if !ok {
			return errors.Newf("unsupported insights series point persist mode: %v", pt.PersistMode)
		}

		if (pt.RepoName != nil && pt.RepoID == nil) || (pt.RepoID != nil && pt.RepoName == nil) {
			return errors.New("RepoName and RepoID must be mutually specified")
		}

		// Upsert the repository name into a separate table, so we get a small ID we can reference
		// many times from the series_points table without storing the repo name multiple times.
		var repoNameID *int
		if pt.RepoName != nil {
			repoNameIDValue, ok, err := basestore.ScanFirstInt(tx.Query(ctx, sqlf.Sprintf(upsertRepoNameFmtStr, *pt.RepoName, *pt.RepoName)))
			if err != nil {
				return errors.Wrap(err, "upserting repo name ID")
			}
			if !ok {
				return errors.Wrap(err, "repo name ID not found (this should never happen)")
			}
			repoNameID = &repoNameIDValue
		}

		if err := inserter.Insert(
			ctx,
			pt.SeriesID,         // series_id
			pt.Point.Time.UTC(), // time
			pt.Point.Value,      // value
			pt.RepoID,           // repo_id
			repoNameID,          // repo_name_id
			repoNameID,          // original_repo_name_id
			pt.Point.Capture,    // capture
		); err != nil {
			return errors.Wrap(err, "Insert")
		}
	}

	for _, inserter := range inserters {
		if err := inserter.Flush(ctx); err != nil {
			return errors.Wrap(err, "Flush")
		}
	}
	return nil
}

func (s *Store) SetInsightSeriesRecordingTimes(ctx context.Context, seriesRecordingTimes []types.InsightSeriesRecordingTimes) (err error) {
	if len(seriesRecordingTimes) == 0 {
		return nil
	}
	inserter := batch.NewInserterWithConflict(ctx, s.Handle(), "insight_series_recording_times", batch.MaxNumPostgresParameters, "ON CONFLICT DO NOTHING", "insight_series_id", "recording_time", "snapshot")

	for _, series := range seriesRecordingTimes {
		id := series.InsightSeriesID
		for _, record := range series.RecordingTimes {
			if err := inserter.Insert(
				ctx,
				id,                     // insight_series_id
				record.Timestamp.UTC(), // recording_time
				record.Snapshot,        // snapshot

			); err != nil {
				return errors.Wrap(err, "Insert")
			}
		}
	}

	if err := inserter.Flush(ctx); err != nil {
		return errors.Wrap(err, "Flush")
	}
	return nil
}

func (s *Store) GetInsightSeriesRecordingTimes(ctx context.Context, id int, opts SeriesPointsOpts) (series types.InsightSeriesRecordingTimes, err error) {
	series.InsightSeriesID = id

	preds := []*sqlf.Query{
		sqlf.Sprintf("insight_series_id = %s", id),
	}
	if opts.From != nil {
		preds = append(preds, sqlf.Sprintf("recording_time >= %s", opts.From.UTC()))
	}
	if opts.To != nil {
		preds = append(preds, sqlf.Sprintf("recording_time <= %s", opts.To.UTC()))
	}
	if opts.After != nil {
		preds = append(preds, sqlf.Sprintf("recording_time > %s", opts.After.UTC()))
	}
	timesQuery := sqlf.Sprintf(getInsightSeriesRecordingTimesStr, sqlf.Join(preds, "\n AND"))

	recordingTimes := []types.RecordingTime{}
	err = s.query(ctx, timesQuery, func(sc scanner) (err error) {
		var recordingTime time.Time
		err = sc.Scan(
			&recordingTime,
		)
		if err != nil {
			return err
		}

		recordingTimes = append(recordingTimes, types.RecordingTime{Timestamp: recordingTime})
		return nil
	})
	if err != nil {
		return series, err
	}
	series.RecordingTimes = recordingTimes

	return series, nil
}

func (s *Store) GetOffsetNRecordingTime(ctx context.Context, seriesId, n int, excludeSnapshot bool) (time.Time, error) {
	preds := []*sqlf.Query{sqlf.Sprintf("insight_series_id = %s", seriesId)}
	if excludeSnapshot {
		preds = append(preds, sqlf.Sprintf("snapshot is false"))
	}

	var tempTime time.Time
	oldestTime, got, err := basestore.ScanFirstTime(s.Query(ctx, sqlf.Sprintf(getOffsetNRecordingTimeSql, sqlf.Join(preds, "and"), n)))
	if err != nil {
		return tempTime, err
	}
	if !got {
		return tempTime, nil
	}
	return oldestTime, nil
}

const getOffsetNRecordingTimeSql = `
select recording_time from insight_series_recording_times where %s order by recording_time desc offset %s limit 1
`

// RecordSeriesPointsAndRecordingTimes is a wrapper around the RecordSeriesPoints and SetInsightSeriesRecordingTimes
// functions. It makes the assumption that this is called per-series, so all the points will share the same SeriesID.
// Use this in favour of RecordSeriesPoints if recording times are known.
func (s *Store) RecordSeriesPointsAndRecordingTimes(ctx context.Context, pts []RecordSeriesPointArgs, recordingTimes types.InsightSeriesRecordingTimes) error {
	tx, err := s.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	if len(pts) > 0 {
		if err := tx.RecordSeriesPoints(ctx, pts); err != nil {
			return err
		}
	}
	if len(recordingTimes.RecordingTimes) > 0 {
		if err := tx.SetInsightSeriesRecordingTimes(ctx, []types.InsightSeriesRecordingTimes{recordingTimes}); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) augmentSeriesPoints(ctx context.Context, opts SeriesPointsOpts, pointsMap map[string]*SeriesPoint, captureValues map[string]struct{}) ([]SeriesPoint, error) {
	if opts.ID == nil || opts.SeriesID == nil || !opts.SupportsAugmentation {
		return []SeriesPoint{}, nil
	}
	recordingsData, err := s.GetInsightSeriesRecordingTimes(ctx, *opts.ID, opts)
	if err != nil {
		return nil, errors.Wrap(err, "GetInsightSeriesRecordingTimes")
	}
	var augmentedPoints []SeriesPoint
	if len(recordingsData.RecordingTimes) > 0 {
		augmentedPoints = coalesceZeroValues(*opts.SeriesID, pointsMap, captureValues, recordingsData.RecordingTimes)
	}
	return augmentedPoints, nil
}

func coalesceZeroValues(seriesID string, pointsMap map[string]*SeriesPoint, captureValues map[string]struct{}, recordingTimes []types.RecordingTime) []SeriesPoint {
	augmentedPoints := []SeriesPoint{}
	for _, recordingTime := range recordingTimes {
		timestamp := recordingTime.Timestamp
		// We have to pivot on potential capture values as well. This is because for capture group data we need to know
		// which capture group values to attach zero data to. Take points [{oct 20, "a"}, {oct 24 "a"}, {oct 24 "b"}]
		// and recording times [oct 20, oct 24]. Without the capture value data we would not be able to know we have a
		// missing {oct 20, "b"} entry.
		for captureValue := range captureValues {
			captureValue := captureValue
			if point, ok := pointsMap[timestamp.String()+captureValue]; ok {
				augmentedPoints = append(augmentedPoints, *point)
			} else {
				var capture *string
				if captureValue != "" {
					capture = &captureValue
				}
				augmentedPoints = append(augmentedPoints, SeriesPoint{
					SeriesID: seriesID,
					Time:     timestamp,
					Value:    0,
					Capture:  capture,
				})
			}
		}
	}
	return augmentedPoints
}

const upsertRepoNameFmtStr = `
WITH e AS(
	INSERT INTO repo_names(name)
	VALUES (%s)
	ON CONFLICT DO NOTHING
	RETURNING id
)
SELECT * FROM e
UNION
	SELECT id FROM repo_names WHERE name = %s;
`

const getInsightSeriesRecordingTimesStr = `
SELECT date_trunc('seconds', recording_time) FROM insight_series_recording_times
WHERE %s
ORDER BY recording_time ASC;
`

func (s *Store) query(ctx context.Context, q *sqlf.Query, sc scanFunc) error {
	rows, err := s.Store.Query(ctx, q)
	if err != nil {
		return err
	}
	return scanAll(rows, sc)
}

// scanner captures the Scan method of sql.Rows and sql.Row
type scanner interface {
	Scan(dst ...any) error
}

// a scanFunc scans one or more rows from a scanner, returning
// the last id column scanned and the count of scanned rows.
type scanFunc func(scanner) (err error)

func scanAll(rows *sql.Rows, scan scanFunc) (err error) {
	defer func() { err = basestore.CloseRows(rows, err) }()
	for rows.Next() {
		if err = scan(rows); err != nil {
			return err
		}
	}
	return rows.Err()
}

var quote = sqlf.Sprintf

// LoadIncompleteDatapoints returns incomplete datapoints for a given series aggregated for each reason and time. This will effectively
// remove any repository granularity information from the result, but the repoIds are retained as a list.
func (s *Store) LoadIncompleteDatapoints(ctx context.Context, seriesID int) (results []IncompleteDatapoint, err error) {
	if seriesID == 0 {
		return nil, errors.New("invalid seriesID")
	}

	q := "select reason, time, ARRAY_AGG(repo_id) from insight_series_incomplete_points where series_id = %s group by reason, time;"
	rows, err := s.Query(ctx, sqlf.Sprintf(q, seriesID))
	if err != nil {
		return nil, err
	}
	return results, scanAll(rows, func(s scanner) (err error) {
		var tmp IncompleteDatapoint
		var repoIds []sql.NullInt64
		if err = rows.Scan(
			&tmp.Reason,
			&tmp.Time,
			pq.Array(&repoIds)); err != nil {
			return err
		}
		mappedRepoIds := make([]int, len(repoIds))
		for i, repoId := range repoIds {
			if repoId.Valid {
				mappedRepoIds[i] = int(repoId.Int64)
			}
		}
		results = append(results, IncompleteDatapoint{
			Reason:  tmp.Reason,
			Time:    tmp.Time,
			RepoIds: mappedRepoIds,
		})
		return nil
	})
}

type AddIncompleteDatapointInput struct {
	SeriesID int
	RepoID   *int
	Reason   IncompleteReason
	Time     time.Time
}

func (s *Store) AddIncompleteDatapoint(ctx context.Context, input AddIncompleteDatapointInput) error {
	q := "insert into insight_series_incomplete_points (series_id, repo_id, reason, time) values (%s, %s, %s, %s) on conflict do nothing;"
	return s.Exec(ctx, sqlf.Sprintf(q, input.SeriesID, input.RepoID, input.Reason, input.Time))
}

type IncompleteDatapoint struct {
	Reason  IncompleteReason
	RepoId  *int
	Time    time.Time
	RepoIds []int
}

type IncompleteReason string

const (
	ReasonTimeout           IncompleteReason = "timeout"
	ReasonGeneric           IncompleteReason = "generic"
	ReasonExceedsErrorLimit IncompleteReason = "exceeds-error-limit"
)

// SeriesPointForExport contains series points data that has additional metadata, like insight view title.
// It should only be used for code insight data exporting.
type SeriesPointForExport struct {
	InsightViewTitle string
	SeriesLabel      string
	SeriesQuery      string
	RecordingTime    time.Time
	RepoName         *string
	RepoId           *api.RepoID
	Value            int
	Capture          *string
}

type ExportOpts struct {
	InsightViewUniqueID string
	IncludeRepoRegex    []string
	ExcludeRepoRegex    []string
}

func (s *Store) GetAllDataForInsightViewID(ctx context.Context, opts ExportOpts) (_ []SeriesPointForExport, err error) {
	// 🚨 SECURITY: this function will only be called if the insight with the given insightViewId is visible given
	// this user context. This is similar to how `SeriesPoints` works.
	// We enforce repo permissions here as we store repository data at this level.
	denylist, err := s.permStore.GetUnauthorizedRepoIDs(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "GetUnauthorizedRepoIDs")
	}

	var preds []*sqlf.Query
	if len(opts.IncludeRepoRegex) > 0 {
		includePreds := []*sqlf.Query{}
		for _, regex := range opts.IncludeRepoRegex {
			if len(regex) == 0 {
				continue
			}
			includePreds = append(includePreds, sqlf.Sprintf("rn.name ~ %s", regex))
		}
		if len(includePreds) > 0 {
			includes := sqlf.Sprintf("(%s)", sqlf.Join(includePreds, "OR"))
			preds = append(preds, includes)
		}
	}
	if len(opts.ExcludeRepoRegex) > 0 {
		for _, regex := range opts.ExcludeRepoRegex {
			if len(regex) == 0 {
				continue
			}
			preds = append(preds, sqlf.Sprintf("rn.name !~ %s", regex))
		}
	}
	if len(preds) == 0 {
		preds = append(preds, sqlf.Sprintf("true"))
	}

	tx, err := s.Transact(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()

	var results []SeriesPointForExport
	exportScanner := func(sc scanner) error {
		var tmp SeriesPointForExport
		if err = sc.Scan(
			&tmp.InsightViewTitle,
			&tmp.SeriesLabel,
			&tmp.SeriesQuery,
			&tmp.RecordingTime,
			&tmp.RepoName,
			&tmp.RepoId,
			&tmp.Value,
			&tmp.Capture,
		); err != nil {
			return err
		}
		// if this is a capture group insight the label will be the capture
		if tmp.Capture != nil {
			tmp.SeriesLabel = *tmp.Capture
		}
		results = append(results, tmp)
		return nil
	}

	formattedPreds := sqlf.Join(preds, "AND")
	// start with the oldest archived points and add them to the results
	if err := tx.query(ctx, sqlf.Sprintf(exportCodeInsightsDataSql, quote(recordingTimesTableArchive), quote(recordingTableArchive), opts.InsightViewUniqueID, formattedPreds), exportScanner); err != nil {
		return nil, errors.Wrap(err, "fetching archived code insights data")
	}
	// then add live points
	// we join both series points tables
	if err := tx.query(ctx, sqlf.Sprintf(exportCodeInsightsDataSql, quote(recordingTimesTable), quote("(select * from series_points union all select * from series_points_snapshots)"), opts.InsightViewUniqueID, formattedPreds), exportScanner); err != nil {
		return nil, errors.Wrap(err, "fetching code insights data")
	}

	// 🚨 SECURITY: The function below filters out repositories that a user should not have access to.
	// This operation was previously done via the SQL predicates, but to enable customers with more than
	// 65535 repositories we need to run this filter in the application layer.
	results = FilterSeriesPoints(denylist, results)

	return results, nil
}

func FilterSeriesPoints(denyList []api.RepoID, points []SeriesPointForExport) []SeriesPointForExport {
	// If there is nothing to filter, then return early. This skips the assignment in the later for-loop.
	if len(denyList) == 0 {
		return points
	}
	// We turn the denyList into a map to ensure O(n) time complexity
	denyMap := make(map[api.RepoID]struct{})
	for _, repoID := range denyList {
		denyMap[repoID] = struct{}{}
	}
	// Based on https://stackoverflow.com/a/59051095
	n := 0
	for _, x := range points {
		// 🚨 SECURITY: Points that have no RepoId must only be visible to users who have no excluded repositories.
		// See https://github.com/sourcegraph/sourcegraph/pull/61580#discussion_r1552271816 for more details.
		if x.RepoId != nil && !isDenied(*x.RepoId, denyMap) {
			points[n] = x
			n++
		}
	}
	return points[:n]
}

func isDenied(id api.RepoID, denyMap map[api.RepoID]struct{}) bool {
	_, ok := denyMap[id]
	return ok
}

const exportCodeInsightsDataSql = `
select iv.title, ivs.label, i.query, isrt.recording_time, rn.name, sp.repo_id, coalesce(sp.value, 0) as value, sp.capture
from %s isrt
    join insight_series i on i.id = isrt.insight_series_id
    join insight_view_series ivs ON i.id = ivs.insight_series_id
    join insight_view iv ON ivs.insight_view_id = iv.id
    left outer join %s sp on sp.series_id = i.series_id and sp.time = isrt.recording_time
    left outer join repo_names rn on sp.repo_name_id = rn.id
	where iv.unique_id = %s and %s
    order by iv.title, isrt.recording_time, ivs.label, sp.capture;
`
