package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/keegancsmith/sqlf"

	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Interface is the interface describing a code insights store. See the Store struct
// for actual API usage.
type Interface interface {
	SeriesPoints(ctx context.Context, opts SeriesPointsOpts) ([]SeriesPoint, error)
	RecordSeriesPoint(ctx context.Context, v RecordSeriesPointArgs) error
	RecordSeriesPoints(ctx context.Context, pts []RecordSeriesPointArgs) error
	CountData(ctx context.Context, opts CountDataOpts) (int, error)
}

var _ Interface = &Store{}

// Store exposes methods to read and write code insights domain models from
// persistent storage.
type Store struct {
	*basestore.Store[schemas.CodeInsights]
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

var _ basestore.ShareableStore[schemas.CodeInsights] = &Store{}

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
	Capture  *string
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

	q := seriesPointsQuery(opts)
	err = s.query(ctx, q, func(sc scanner) error {
		var point SeriesPoint
		err := sc.Scan(
			&point.SeriesID,
			&point.Time,
			&point.Value,
			&point.Metadata,
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
SELECT sub.series_id, sub.interval_time, SUM(sub.value) as value, sub.metadata, sub.capture FROM (
	SELECT sp.repo_name_id, sp.series_id, sp.time AS interval_time, MAX(value) as value, null as metadata, capture
	FROM (  select * from series_points
			union
			select * from series_points_snapshots
	) AS sp
	JOIN repo_names rn ON sp.repo_name_id = rn.id
	WHERE %s
	GROUP BY sp.series_id, interval_time, sp.repo_name_id, capture
	ORDER BY sp.series_id, interval_time, sp.repo_name_id
) sub
GROUP BY sub.series_id, sub.interval_time, sub.metadata, sub.capture
ORDER BY sub.series_id, sub.interval_time ASC
`

// Note that the series_points table may contain duplicate points, or points recorded at irregular
// intervals. In specific:
//
// 1. Multiple points recorded at the same time T for cardinality C will be considered part of the same vector.
//    For example, series S and repos R1, R2 have a point at time T. The sum over R1,R2 at T will give the
//    aggregated sum for that series at time T.
// 2. Rarely, it may contain duplicate data points due to the at-least once semantics of query execution.
//    This will cause some jitter in the aggregated series, and will skew the results slightly.
// 3. Searches may not complete at the same exact time, so even in a perfect world if the interval
//    should be 12h it may be off by a minute or so.
func seriesPointsQuery(opts SeriesPointsOpts) *sqlf.Query {
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
	return sqlf.Sprintf(
		fullVectorSeriesAggregation+limitClause,
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

	// Metadata contains arbitrary JSON metadata to associate with the data point, if any.
	//
	// See the DB schema comments for intended use cases. This should generally be small,
	// low-cardinality data to avoid inflating the table.
	Metadata any

	PersistMode PersistMode
}

// RecordSeriesPoint records a data point for the specfied series ID (which is a unique ID for the
// series, not a DB table primary key ID).
func (s *Store) RecordSeriesPoint(ctx context.Context, v RecordSeriesPointArgs) (err error) {
	// Start transaction.
	var txStore *basestore.Store
	txStore, err = s.Store.Transact(ctx)
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

	tableName, err := getTableForPersistMode(v.PersistMode)
	if err != nil {
		return err
	}

	q := sqlf.Sprintf(
		recordSeriesPointFmtstr,
		sqlf.Sprintf(tableName),
		v.SeriesID,         // series_id
		v.Point.Time.UTC(), // time
		v.Point.Value,      // value
		metadataID,         // metadata_id
		v.RepoID,           // repo_id
		repoNameID,         // repo_name_id
		repoNameID,         // original_repo_name_id
		v.Point.Capture,
	)
	// Insert the actual data point.
	return txStore.Exec(ctx, q)
}

func getTableForPersistMode(mode PersistMode) (string, error) {
	switch mode {
	case RecordMode:
		return recordingTable, nil
	case SnapshotMode:
		return snapshotsTable, nil
	default:
		return "", errors.Newf("unsupported insights series point persist mode: %v", mode)
	}
}

// RecordSeriesPoints stores multiple data points atomically.
func (s *Store) RecordSeriesPoints(ctx context.Context, pts []RecordSeriesPointArgs) (err error) {
	tx, err := s.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	for _, pt := range pts {
		// this is a pretty naive implementation, this can be refactored to reduce db calls
		if err := s.RecordSeriesPoint(ctx, pt); err != nil {
			return err
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
INSERT INTO %s (
	series_id,
	time,
	value,
	metadata_id,
	repo_id,
	repo_name_id,
	original_repo_name_id, capture)
VALUES (%s, %s, %s, %s, %s, %s, %s, %s);
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
