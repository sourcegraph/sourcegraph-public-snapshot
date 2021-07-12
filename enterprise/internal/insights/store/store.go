package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
)

// Interface is the interface describing a code insights store. See the Store struct
// for actual API usage.
type Interface interface {
	SeriesPoints(ctx context.Context, opts SeriesPointsOpts) ([]SeriesPoint, error)
	RecordSeriesPoint(ctx context.Context, v RecordSeriesPointArgs) error
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

// New returns a new Store backed by the given Timescale db.
func New(db dbutil.DB, permStore InsightPermissionStore) *Store {
	return NewWithClock(db, permStore, timeutil.Now)
}

// NewWithClock returns a new Store backed by the given db and
// clock for timestamps.
func NewWithClock(db dbutil.DB, permStore InsightPermissionStore, clock func() time.Time) *Store {
	return &Store{Store: basestore.NewWithDB(db, sql.TxOptions{}), now: clock, permStore: permStore}
}

var _ basestore.ShareableStore = &Store{}

// Handle returns the underlying transactable database handle.
// Needed to implement the ShareableStore interface.
func (s *Store) Handle() *basestore.TransactableHandle { return s.Store.Handle() }

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
	Metadata []byte
}

func (s *SeriesPoint) String() string {
	return fmt.Sprintf("SeriesPoint{Time: %q, Value: %v, Metadata: %s}", s.Time, s.Value, s.Metadata)
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
	// TODO(slimsag): Add ability to do limited filtering based on metadata.

	// Time ranges to query from/to, if non-nil, in UTC.
	From, To *time.Time

	// Limit is the number of data points to query, if non-zero.
	Limit int
}

//SeriesPoints queries data points over time for a specific insights' series.
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

	q := seriesPointsQuery(opts)
	err = s.query(ctx, q, func(sc scanner) error {
		var point SeriesPoint
		err := sc.Scan(
			&point.SeriesID,
			&point.Time,
			&point.Value,
			&point.Metadata,
		)
		if err != nil {
			return err
		}
		points = append(points, point)
		return nil
	})
	return points, err
}

// This query is a barebones implementation of per-repo per-series last-observation carried forward. Long term
// this query is too expensive to run in real-time and should be moved to a materialized view.
const lastObservationCarriedPointsSql = `select sub.series_id, sub.interval_time, sum(value) as value, null as metadata from (WITH target_times AS (SELECT *
FROM GENERATE_SERIES(CURRENT_TIMESTAMP::date - INTERVAL '26 weeks', CURRENT_TIMESTAMP::date, '2 weeks') as interval_time)
SELECT sub.series_id, sub.repo_id, sub.value, interval_time
FROM (select distinct repo_id, series_id from series_points) as r
cross join target_times tt
join LATERAL (
    select sp.* from series_points as sp
    where sp.repo_id = r.repo_id and sp.time <= tt.interval_time and sp.series_id = r.series_id
    order by time DESC
    limit 1
    ) sub on sub.repo_id = r.repo_id and r.series_id = sub.series_id
order by interval_time, repo_id) as sub
where %s
group by sub.series_id, sub.interval_time
order by interval_time desc
`

// Note that the series_points table may contain duplicate points, or points recorded at irregular
// intervals. In specific:
//
// 1. It may have multiple points recorded at the same exact point in time, e.g. with different
//    repo_id (datapoint recorded per repository), or only a single point recorded (datapoint
//    recorded globally.)
// 2. Rarely, it may contain duplicate data points. For example, when repo-updater is started the
//    initial jobs for recording insights will be enqueued, and then e.g. 12h later. If repo-updater
//    gets restarted multiple times, there may be many multiple nearly identical data points recorded
//    in a short period of time instead of at the 12h interval.
// 3. Data backfilling may not operate at the same interval, or same # of points per interval, and
//    thus the interval between data points may be irregular.
// 4. Searches may not complete at the same exact time, so even in a perfect world if the interval
//    should be 12h it may be off by a minute or so.
// 5. Intervals that are missing a data point will need to resolve the last observation and carry it forward.
//
// Additionally, it is important to note that there may be data points associated with a repo OR not
// associated with a repo at all (global.)

func seriesPointsQuery(opts SeriesPointsOpts) *sqlf.Query {
	preds := []*sqlf.Query{}

	if opts.SeriesID != nil {
		preds = append(preds, sqlf.Sprintf("series_id = %s", *opts.SeriesID))
	}
	if opts.RepoID != nil {
		preds = append(preds, sqlf.Sprintf("repo_id = %d", int32(*opts.RepoID)))
	}
	if opts.From != nil {
		preds = append(preds, sqlf.Sprintf("interval_time >= %s", *opts.From))
	}
	if opts.To != nil {
		preds = append(preds, sqlf.Sprintf("interval_time <= %s", *opts.To))
	}
	limitClause := ""
	if opts.Limit > 0 {
		limitClause = fmt.Sprintf("LIMIT %d", opts.Limit)
	}
	if len(opts.Included) > 0 {
		s := fmt.Sprintf("repo_id = any(%v)", values(opts.Included))
		preds = append(preds, sqlf.Sprintf(s))
	}
	if len(opts.Excluded) > 0 {
		s := fmt.Sprintf("repo_id != all(%v)", values(opts.Excluded))
		preds = append(preds, sqlf.Sprintf(s))
	}

	if len(preds) == 0 {
		preds = append(preds, sqlf.Sprintf("TRUE"))
	}
	return sqlf.Sprintf(
		lastObservationCarriedPointsSql+limitClause,
		sqlf.Join(preds, "\n AND "),
	)
}

//values constructs a SQL values statement out of an array of repository ids
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

	// Metadata contains arbitrary JSON metadata to associate with the data point, if any.
	//
	// See the DB schema comments for intended use cases. This should generally be small,
	// low-cardinality data to avoid inflating the table.
	Metadata interface{}
}

// RecordSeriesPoint records a data point for the specfied series ID (which is a unique ID for the
// series, not a DB table primary key ID).
func (s *Store) RecordSeriesPoint(ctx context.Context, v RecordSeriesPointArgs) (err error) {
	// Start transaction.
	var txStore *basestore.Store
	txStore, err = s.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = txStore.Done(err) }()

	if (v.RepoName != nil && v.RepoID == nil) || (v.RepoID != nil && v.RepoName == nil) {
		return errors.New("RepoName and RepoID must be mutually specified")
	}

	// Upsert the repository name into a separate table, so we get a small ID we can reference
	// many times from the series_points table without storing the repo name multiple times.
	var repoNameID *int
	if v.RepoName != nil {
		repoNameIDValue, ok, err := basestore.ScanFirstInt(txStore.Query(ctx, sqlf.Sprintf(upsertRepoNameFmtStr, *v.RepoName, *v.RepoName)))
		if err != nil {
			return errors.Wrap(err, "upserting repo name ID")
		}
		if !ok {
			return errors.Wrap(err, "repo name ID not found (this should never happen)")
		}
		repoNameID = &repoNameIDValue
	}

	// Upsert the metadata into a separate table, so we get a small ID we can reference many times
	// from the series_points table without storing the metadata multiple times.
	var metadataID *int
	if v.Metadata != nil {
		jsonMetadata, err := json.Marshal(v.Metadata)
		if err != nil {
			return errors.Wrap(err, "upserting: encoding metadata")
		}
		metadataIDValue, ok, err := basestore.ScanFirstInt(txStore.Query(ctx, sqlf.Sprintf(upsertMetadataFmtStr, jsonMetadata, jsonMetadata)))
		if err != nil {
			return errors.Wrap(err, "upserting metadata ID")
		}
		if !ok {
			return errors.Wrap(err, "metadata ID not found (this should never happen)")
		}
		metadataID = &metadataIDValue
	}

	// Insert the actual data point.
	return txStore.Exec(ctx, sqlf.Sprintf(
		recordSeriesPointFmtstr,
		v.SeriesID,         // series_id
		v.Point.Time.UTC(), // time
		v.Point.Value,      // value
		metadataID,         // metadata_id
		v.RepoID,           // repo_id
		repoNameID,         // repo_name_id
		repoNameID,         // original_repo_name_id
	))
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

const upsertMetadataFmtStr = `
-- source: enterprise/internal/insights/store/store.go:RecordSeriesPoint
WITH e AS(
    INSERT INTO metadata(metadata)
    VALUES (%s)
    ON CONFLICT DO NOTHING
    RETURNING id
)
SELECT * FROM e
UNION
	SELECT id FROM metadata WHERE metadata = %s;
`

const recordSeriesPointFmtstr = `
-- source: enterprise/internal/insights/store/store.go:RecordSeriesPoint
INSERT INTO series_points(
	series_id,
	time,
	value,
	metadata_id,
	repo_id,
	repo_name_id,
	original_repo_name_id)
VALUES (%s, %s, %s, %s, %s, %s, %s);
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
	Scan(dst ...interface{}) error
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
