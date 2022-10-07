package store

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/RoaringBitmap/roaring"
	"github.com/keegancsmith/sqlf"

	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/batch"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Interface is the interface describing a code insights store. See the Store struct
// for actual API usage.
type Interface interface {
	SeriesPoints(ctx context.Context, opts SeriesPointsOpts) ([]SeriesPoint, error)
	RecordSeriesPoints(ctx context.Context, pts []RecordSeriesPointArgs) error
	CountData(ctx context.Context, opts CountDataOpts) (int, error)
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
	return &Store{Store: s.Store.With(other), now: s.now}
}

var _ Interface = &Store{}

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
	return fmt.Sprintf("SeriesPoint{Time: %q, Value: %v}", s.Time, s.Value)
}

// SeriesPointsOpts describes options for querying insights' series data points.
type SeriesPointsOpts struct {
	// SeriesID is the unique series ID to query, if non-nil.
	SeriesID *string

	// RepoID, if non-nil, indicates to filter results to only points recorded with this repo ID.
	RepoID *api.RepoID

	Excluded []api.RepoID
	Included []api.RepoID

	// TODO(slimsag): Add ability to filter based on repo name, original name.

	IncludeRepoRegex []string
	ExcludeRepoRegex []string

	// Time ranges to query from/to, if non-nil, in UTC.
	From, To *time.Time

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
		return nil
	})
	return points, err
}

func (s *Store) LoadSeriesInMem(ctx context.Context, opts SeriesPointsOpts) (points []SeriesPoint, err error) {
	denylist, err := s.permStore.GetUnauthorizedRepoIDs(ctx)
	if err != nil {
		return []SeriesPoint{}, err
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

	for _, pointTime := range mapping {
		for _, point := range pointTime {
			points = append(points, SeriesPoint{
				SeriesID: *opts.SeriesID,
				Time:     point.Time,
				Value:    point.Value,
				Capture:  point.Capture,
			})
		}
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
-- source: enterprise/internal/insights/store/store.go:Delete
DELETE FROM series_points where series_id = %s;
`

const deleteForSeriesSnapshots = `
-- source: enterprise/internal/insights/store/store.go:Delete
DELETE FROM series_points_snapshots where series_id = %s;
`

// Note: the inner query could return duplicate points on its own if we merely did a SUM(value) over
// all desired repositories. By using the sub-query, we select the per-repository maximum (thus
// eliminating duplicate points that might have been recorded in a given interval for a given repository)
// and then SUM the result for each repository, giving us our final total number.
const fullVectorSeriesAggregation = `
-- source: enterprise/internal/insights/store/store.go:SeriesPoints
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

	if len(opts.Included) > 0 {
		s := fmt.Sprintf("repo_id = any(%v)", values(opts.Included))
		preds = append(preds, sqlf.Sprintf(s))
	}
	if len(opts.Excluded) > 0 {
		preds = append(preds, sqlf.Sprintf("perm.excluded_repo IS NULL"))
	}
	if len(opts.IncludeRepoRegex) > 0 {
		for _, regex := range opts.IncludeRepoRegex {
			if len(regex) == 0 {
				continue
			}
			preds = append(preds, sqlf.Sprintf("rn.name ~ %s", regex))
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
	return nil
}

const deleteSnapshotsSql = `
-- source: enterprise/internal/insights/store/store.go:DeleteSnapshots
delete from %s where series_id = %s;
`

type PersistMode string

const (
	RecordMode     PersistMode = "record"
	SnapshotMode   PersistMode = "snapshot"
	recordingTable string      = "series_points"
	snapshotsTable string      = "series_points_snapshots"
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

// RecordSeriesPoints stores multiple data points atomically.
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

const upsertRepoNameFmtStr = `
-- source: enterprise/internal/insights/store/store.go:RecordSeriesPoint
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
