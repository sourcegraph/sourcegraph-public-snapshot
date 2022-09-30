package store

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/database/batch"

	logger "github.com/sourcegraph/log"

	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"

	"github.com/RoaringBitmap/roaring"
	"github.com/keegancsmith/sqlf"
	"github.com/keisku/gorilla"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type SampleStore interface {
	// Transaction
	basestore.ShareableStore
	Transact(ctx context.Context) (SampleStore, error)
	With(other basestore.ShareableStore) SampleStore
	Done(err error) error

	// Sample Operations
	Sample(ctx context.Context, samples []RepoSample) error
	StoreRow(ctx context.Context, row UncompressedRow, seriesId uint32) error
	LoadRows(ctx context.Context, opts CompressedRowsOpts) ([]UncompressedRow, error)
	Append(ctx context.Context, key TimeSeriesKey, samples []RawSample) error
	LoadTimeseries(ctx context.Context, opts CompressedRowsOpts) (timeseries []SeriesPoint, err error)

	// Snapshot Operations
	Snapshot(ctx context.Context, key TimeSeriesKey, snapshot RawSample) error
	ClearSnapshots(ctx context.Context, seriesId uint32) error
}

type SampleWriter interface {
	Append(ctx context.Context, key TimeSeriesKey, samples []RawSample) error
}

type SampleReader interface {
	LoadRows(ctx context.Context, opts CompressedRowsOpts) ([]UncompressedRow, error)
}

type sampleStore struct {
	*basestore.Store
	permStore InsightPermissionStore
	logger    logger.Logger
}

var _ SampleStore = &sampleStore{}

func SampleStoreFromLegacyStore(legacy *Store) SampleStore {
	return &sampleStore{
		Store:     legacy.Store,
		permStore: legacy.permStore,
	}
}

func NewSampleStore(db edb.InsightsDB, permStore InsightPermissionStore) SampleStore {
	return &sampleStore{
		Store:     basestore.NewWithHandle(db.Handle()),
		permStore: permStore,
	}
}

func (s *sampleStore) With(other basestore.ShareableStore) SampleStore {
	return &sampleStore{
		Store:     s.Store.With(other),
		permStore: s.permStore,
	}
}

func (s *sampleStore) Transact(ctx context.Context) (SampleStore, error) {
	return s.transact(ctx)
}

func (s *sampleStore) transact(ctx context.Context) (*sampleStore, error) {
	tx, err := s.Store.Transact(ctx)
	if err != nil {
		return nil, err
	}

	return &sampleStore{
		Store:     tx,
		permStore: s.permStore,
	}, nil
}

type RawSample struct {
	Time  uint32
	Value float64
}

func (r RawSample) String() string {
	return fmt.Sprintf("(%s %f)", time.Unix(int64(r.Time), 0).String(), r.Value)
}

type TimeSeriesKey struct {
	SeriesId uint32
	RepoId   uint32
	Capture  *string
}

type UncompressedRow struct {
	altFormatRowMetadata
	Samples  []RawSample
	Snapshot SnapshotSample
}

type CompressedRow struct {
	altFormatRowMetadata
	Data     []byte
	Snapshot SnapshotSample
}

type SnapshotSample struct {
	Time  *uint32
	Value *float64
}

type altFormatRowMetadata struct {
	Id      uint32
	RepoId  uint32
	Capture *string
}

func (s *sampleStore) StoreRow(ctx context.Context, row UncompressedRow, seriesId uint32) error {
	prepareSamplesForCompression(row.Samples)
	buf, err := compressSamples(row.Samples)
	if err != nil {
		return err
	}

	data := buf.Bytes()

	var q *sqlf.Query
	if row.Id != 0 {
		q = sqlf.Sprintf("update series_points_compressed set data = %s where id = %s", data, row.Id)
	} else {
		q = sqlf.Sprintf("insert into series_points_compressed (series_id, repo_id, capture, data) values (%s, %s, %s, %s)", seriesId, row.RepoId, row.Capture, data)
	}
	return s.Exec(ctx, q)
}

func (s *sampleStore) LoadRows(ctx context.Context, opts CompressedRowsOpts) ([]UncompressedRow, error) {
	// i'd really like to rethink this and load it much earlier - this is getting reloaded for every insight series we fetch
	// maybe a privilged access struct vs unprivd?
	denylist, err := s.permStore.GetUnauthorizedRepoIDs(ctx)
	if err != nil {
		return nil, err
	}
	denyBitmap := roaring.New()
	for _, id := range denylist {
		denyBitmap.Add(uint32(id))
	}

	return s.loadRows(ctx, opts, denyBitmap)
}

func (s *sampleStore) loadRows(ctx context.Context, opts CompressedRowsOpts, denyBitmap *roaring.Bitmap) ([]UncompressedRow, error) {
	var rows []UncompressedRow
	return rows, s.streamRows(ctx, opts, func(ctx context.Context, row *CompressedRow) (err error) {
		if denyBitmap.Contains(row.RepoId) {
			return nil
		}

		dcmp, err := decompressSamples(row.Data)
		if err != nil {
			return err
		}

		rows = append(rows, UncompressedRow{
			altFormatRowMetadata: row.altFormatRowMetadata,
			Samples:              dcmp,
			Snapshot:             row.Snapshot,
		})
		return nil
	})
}

func loadRowsQuery(opts CompressedRowsOpts) *sqlf.Query {
	baseQuery := `select spc.id, spc.repo_id, data, capture, snapshot_time, snapshot_value from series_points_compressed spc`
	return buildQuery(baseQuery, opts)
}

func loadWriteAheadQuery(opts CompressedRowsOpts) *sqlf.Query {
	baseQuery := "select spc.repo_id, spc.capture, spc.time, spc.value from samples_write_ahead spc"
	return buildQuery(baseQuery, opts)
}

func buildQuery(baseQuery string, opts CompressedRowsOpts) *sqlf.Query {
	var preds []*sqlf.Query
	joinCond := " JOIN repo_names rn ON spc.repo_id = rn.repo_id"
	hasJoin := false
	if len(opts.IncludeRepoRegex) > 0 {
		hasJoin = true
		for _, regex := range opts.IncludeRepoRegex {
			if len(regex) == 0 {
				continue
			}
			preds = append(preds, sqlf.Sprintf("rn.name ~ %s", regex))
		}
	}
	if len(opts.ExcludeRepoRegex) > 0 {
		hasJoin = true
		for _, regex := range opts.ExcludeRepoRegex {
			if len(regex) == 0 {
				continue
			}
			preds = append(preds, sqlf.Sprintf("rn.name !~ %s", regex))
		}
	}

	if len(opts.UniversalSeriesID) != 0 {
		preds = append(preds, sqlf.Sprintf("series_id = (select isn.id from insight_series as isn where isn.series_id = %s)", opts.UniversalSeriesID))
	}
	if opts.SeriesID != 0 {
		preds = append(preds, sqlf.Sprintf("series_id = %s", opts.SeriesID))
	}
	if opts.RepoID != 0 {
		preds = append(preds, sqlf.Sprintf("repo_id = %s", opts.RepoID))
	}

	if opts.Key != nil {
		preds = append(preds, sqlf.Sprintf("series_id = %s", opts.Key.SeriesId))
		preds = append(preds, sqlf.Sprintf("repo_id = %s", opts.Key.RepoId))
		if opts.Key.Capture == nil {
			preds = append(preds, sqlf.Sprintf("capture is null"))
		} else {
			preds = append(preds, sqlf.Sprintf("capture = %s", *opts.Key.Capture))
		}
	}
	if hasJoin {
		baseQuery += joinCond
	}
	if len(preds) > 0 {
		baseQuery += " where %s"
	}

	if opts.ShouldLock {
		baseQuery += " order by spc.id FOR UPDATE"
	}
	return sqlf.Sprintf(baseQuery, sqlf.Join(preds, "AND"))
}

func (s *sampleStore) loadWriteAhead(ctx context.Context, opts CompressedRowsOpts, denyBitmap *roaring.Bitmap) (rows []writeAheadRecord, err error) {
	q := loadWriteAheadQuery(opts)
	return rows, s.query(ctx, q, func(sc scanner) (err error) {
		var tmp writeAheadRecord
		if err := sc.Scan(
			&tmp.RepoId,
			&tmp.Capture,
			&tmp.time,
			&tmp.value,
		); err != nil {
			return err
		}
		if denyBitmap.Contains(tmp.RepoId) {
			return nil
		}
		rows = append(rows, tmp)

		return nil
	})
}

func (s *sampleStore) LoadTimeseries(ctx context.Context, opts CompressedRowsOpts) (timeseries []SeriesPoint, err error) {
	// i'd really like to rethink this and load it much earlier - this is getting reloaded for every insight series we fetch
	// maybe a privilged access struct vs unprivd?
	denylist, err := s.permStore.GetUnauthorizedRepoIDs(ctx)
	if err != nil {
		return nil, err
	}
	denyBitmap := roaring.New()
	for _, id := range denylist {
		denyBitmap.Add(uint32(id))
	}

	rows, err := s.loadRows(ctx, opts, denyBitmap)
	if err != nil {
		return nil, errors.Wrap(err, "LoadRows")
	}
	writeAhead, err := s.loadWriteAhead(ctx, opts, denyBitmap)
	if err != nil {
		return nil, errors.Wrap(err, "loadWriteAhead")
	}

	return toTimeseries(rows, opts.UniversalSeriesID, writeAhead), nil
}

func (s *sampleStore) Append(ctx context.Context, key TimeSeriesKey, samples []RawSample) (err error) {
	tx, err := s.transact(ctx)
	if err != nil {
		return err
	}
	defer func() {
		txErr := tx.Done(err)
		if txErr != nil {
			err = errors.Wrap(err, "txErr")
		}
	}()

	lgr := logger.Scoped("Append", "append")

	// err = tx.Exec(ctx, sqlf.Sprintf("LOCK TABLE series_points_compressed IN ROW EXCLUSIVE MODE;"))
	// if err != nil {
	// 	return errors.Wrap(err, "Append.Lock")
	// }

	var keyMatch UncompressedRow
	// var count int

	got, err := tx.consistentReadRow(ctx, key)
	if err != nil {
		return errors.Wrap(err, "consistentReadRow")
	} else if got.Id == 0 {
		lgr.Info("trying to fetch")
		// lets try loading it instead
		err = tx.streamRows(ctx, CompressedRowsOpts{Key: &key, ShouldLock: true}, func(ctx context.Context, row *CompressedRow) error {
			// decompressed, err := decompressSamples(row.Data)
			// if err != nil {
			// 	return errors.Wrap(err, "failed to decompress sample data")
			// }
			// keyMatch = &UncompressedRow{
			// 	altFormatRowMetadata: row.altFormatRowMetadata,
			// 	Samples:              decompressed,
			// }
			got = *row
			return nil
		})
		if err != nil {
			return errors.Wrap(err, "inner stream rows")
		}
	}
	if got.Id == 0 {

		return errors.New("unable to load row")
	}
	keyMatch = UncompressedRow{
		altFormatRowMetadata: got.altFormatRowMetadata,
		Samples:              nil,
		Snapshot:             got.Snapshot,
	}
	if len(got.Data) > 0 {
		keyMatch.Samples, err = decompressSamples(got.Data)
		if err != nil {
			return errors.Wrap(err, "failed to decompress sample data")
		}
	}
	lgr.Info("consistent load",
		logger.String("key", fmt.Sprintf("%v", key)),
		logger.String("compressed", fmt.Sprintf("%v", got)),
		logger.String("uncompressed", fmt.Sprintf("%v", keyMatch)))
	//
	// if err != nil {
	// 	return errors.Wrapf(err, "failed to read row for key: %v", key)
	// }
	// log15.Info("num_rows", "rows", count, "key", key, "capture", *key.Capture)

	// Sample: Sample.Append: failed to read row for key: {124 11 0xc001641f70}: ERROR: deadlock detected (SQLSTATE 40P01)
	// if keyMatch == nil {
	// 	// no data exists already for this key, so we will write a new row with the provided samples
	// 	keyMatch = &UncompressedRow{
	// 		altFormatRowMetadata: altFormatRowMetadata{
	// 			RepoId:  key.RepoId,
	// 			Capture: key.Capture,
	// 		},
	// 		Samples: nil, // not setting this because this is the "original" samples, we will append the new ones later
	// 	}
	// }
	lgr.Info("before store")
	keyMatch.Samples = append(keyMatch.Samples, samples...)
	err = tx.StoreRow(ctx, keyMatch, key.SeriesId)
	if err != nil {
		return errors.Wrap(err, "StoreRow")
	}
	lgr.Info("after store")
	return nil
}

func prepareSamplesForCompression(samples []RawSample) {
	uniques := make(map[uint32]RawSample)
	for i := range samples {
		if val, ok := uniques[samples[i].Time]; ok {
			// deduplicate by time - keep the highest value for any duplicate times
			// eventually we may want to be less granular about what we consider a duplicate time,
			// for example round down to the near hour
			uniques[samples[i].Time] = RawSample{
				Time:  samples[i].Time,
				Value: math.Max(val.Value, samples[i].Value),
			}
		} else {
			uniques[samples[i].Time] = samples[i]
		}
	}
	samples = samples[:0]

	for _, v := range uniques {
		samples = append(samples, v)
	}

	sort.Slice(samples, func(i, j int) bool {
		return samples[i].Time < samples[j].Time
	})
}

type rowStreamFunc func(ctx context.Context, row *CompressedRow) error

// CompressedRowsOpts describes options for querying insights' series data points.
type CompressedRowsOpts struct {
	// SeriesID is the unique series ID to query, if non-nil.
	UniversalSeriesID string
	SeriesID          uint32
	RepoID            uint32

	IncludeRepoRegex []string
	ExcludeRepoRegex []string

	Key *TimeSeriesKey

	// Limit is the number of data points to query, if non-zero.
	Limit int

	ShouldLock bool
}

func (s *sampleStore) streamRows(ctx context.Context, opts CompressedRowsOpts, callback rowStreamFunc) error {
	return s.query(ctx, loadRowsQuery(opts), func(sc scanner) (err error) {
		var tmp CompressedRow
		if err := sc.Scan(
			&tmp.Id,
			&tmp.RepoId,
			&tmp.Data,
			&tmp.Capture,
			&tmp.Snapshot.Time,
			&tmp.Snapshot.Value,
		); err != nil {
			return err
		}

		return callback(ctx, &tmp)
	})
}

func (s *sampleStore) consistentReadRow(ctx context.Context, key TimeSeriesKey) (row CompressedRow, err error) {
	q := "insert into series_points_compressed (series_id, repo_id, capture) values (%s, %s, %s) ON CONFLICT DO NOTHING RETURNING id, repo_id, data, capture, snapshot_time, snapshot_value"

	err = s.query(ctx, sqlf.Sprintf(q, key.SeriesId, key.RepoId, key.Capture), func(sc scanner) (err error) {
		if err := sc.Scan(
			&row.Id,
			&row.RepoId,
			&row.Data,
			&row.Capture,
			&row.Snapshot.Time,
			&row.Snapshot.Value,
		); err != nil {
			return err
		}
		return nil
	})
	return row, err
}

func decompressSamples(data []byte) (samples []RawSample, err error) {
	decompressor, _, err := gorilla.NewDecompressor(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	iter := decompressor.Iterator()
	for iter.Next() {
		t, v := iter.At()
		if len(samples) > 0 {
			// this is a little optimization because Code Insights operates with minimum hourly intervals.
			// we still store the first value as the full timestamp in seconds, and all the delta-of-delta as hours
			t = t * 3600
		}
		samples = append(samples, RawSample{
			Time:  t,
			Value: v,
		})
	}

	return samples, err
}

func compressSamples(samples []RawSample) (buf *bytes.Buffer, err error) {
	if len(samples) == 0 {
		return nil, errors.New("no samples provided to compress")
	}
	buf = new(bytes.Buffer)
	header := samples[0]
	c, finish, err := gorilla.NewCompressor(buf, header.Time)
	if err != nil {
		return nil, err
	}

	if err = c.Compress(header.Time, header.Value); err != nil {
		return nil, err
	}

	for i := 1; i < len(samples); i++ {
		smpl := samples[i]
		if err = c.Compress(smpl.Time/3600, smpl.Value); err != nil {
			// we convert the time to hours as a little optimization because Code Insights operates with minimum hourly intervals.
			// we still store the first value as the full timestamp in seconds, and all the delta-of-delta as hours
			return nil, err
		}
	}

	return buf, finish()
}

func (s *sampleStore) query(ctx context.Context, q *sqlf.Query, sc scanFunc) error {
	rows, err := s.Store.Query(ctx, q)
	if err != nil {
		return err
	}
	return scanAll(rows, sc)
}

func toTimeseries(data []UncompressedRow, seriesId string, extra []writeAheadRecord) (results []SeriesPoint) {
	getKey := func(s *string) string {
		if s == nil {
			return ""
		}
		return *s
	}

	lgr := logger.Scoped("toTimeseries", "asdf")
	lgr.Info("totimeseries", logger.Int("rows", len(data)), logger.Int("extra", len(extra)))

	byCapture := make(map[string][]UncompressedRow)
	for _, datum := range data {
		byCapture[getKey(datum.Capture)] = append(byCapture[getKey(datum.Capture)], datum)
	}
	extraByCapture := make(map[string][]writeAheadRecord)
	for _, extraRecord := range extra {
		extraByCapture[getKey(extraRecord.Capture)] = append(extraByCapture[getKey(extraRecord.Capture)], extraRecord)
	}

	coalesce := func(val *float64, coal float64) float64 {
		if val == nil {
			return coal
		}
		return *val
	}

	toPtr := func(s string) *string {
		if s == "" {
			return nil
		}
		return &s
	}

	for key, vals := range byCapture {
		mapped := make(map[uint32]float64)
		snapshots := make(map[uint32]float64)

		extraPts := extraByCapture[key]

		for _, val := range vals {
			for _, sample := range val.Samples {
				mapped[sample.Time] += sample.Value
			}
			if val.Snapshot.Time != nil {
				snapshots[*val.Snapshot.Time] += coalesce(val.Snapshot.Value, 0)
			}
			for _, record := range extraPts {
				mapped[record.time] += record.value
				// if there are any extra points for this capture add them here (ie aggregate across repo_id)
			}
		}

		for utime, agg := range mapped {
			results = append(results, SeriesPoint{
				SeriesID: seriesId,
				Time:     time.Unix(int64(utime), 0),
				Value:    agg,
				Capture:  toPtr(key),
			})
		}

		for utime, agg := range snapshots {
			results = append(results, SeriesPoint{
				SeriesID: seriesId,
				Time:     time.Unix(int64(utime), 0),
				Value:    agg,
				Capture:  toPtr(key),
			})

			delete(extraByCapture, key)
		}
	}

	for key, records := range extraByCapture {
		// this is so ugly :(
		mapped := make(map[uint32]float64)

		for _, val := range records {
			mapped[val.time] += val.value
		}

		for mapkey, mapval := range mapped {
			results = append(results, SeriesPoint{
				SeriesID: seriesId,
				Time:     time.Unix(int64(mapkey), 0),
				Value:    mapval,
				Capture:  toPtr(key),
			})
		}
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].Time.Before(results[j].Time)
	})
	return results
}

func (s *sampleStore) Snapshot(ctx context.Context, key TimeSeriesKey, snapshot RawSample) (err error) {
	var preds []*sqlf.Query
	q := "update series_points_compressed set snapshot_time = %s, snapshot_value = %s where %s"

	preds = append(preds, sqlf.Sprintf("series_id = %s", key.SeriesId))
	preds = append(preds, sqlf.Sprintf("repo_id = %s", key.RepoId))

	if key.Capture == nil {
		preds = append(preds, sqlf.Sprintf("capture is null"))
	} else {
		preds = append(preds, sqlf.Sprintf("capture = %s", key.Capture))
	}

	return s.Exec(ctx, sqlf.Sprintf(q, snapshot.Time, snapshot.Value, sqlf.Join(preds, "AND")))
}

func (s *sampleStore) ClearSnapshots(ctx context.Context, seriesId uint32) error {
	q := "update series_points_compressed set snapshot_time = null, snapshot_value = null where series_id = %s"

	return s.Exec(ctx, sqlf.Sprintf(q, seriesId))
}

type RepoSample struct {
	RepoName string
	TimeSeriesKey
	RawSample
}

func (s *sampleStore) Sample(ctx context.Context, samples []RepoSample) (err error) {
	tx, err := s.transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	if err = tx.AppendWriteAhead(ctx, samples); err != nil {
		return errors.Wrap(err, "Sample.AppendWriteAhead")
	}
	for _, sample := range samples {
		if err = tx.sampleRepoName(ctx, sample.RepoId, sample.RepoName); err != nil {
			return errors.Wrap(err, "Sample.sampleRepoName")
		}
	}
	return err
}

func (s *sampleStore) sampleRepoName(ctx context.Context, repoId uint32, repoName string) error {
	q := "insert into sampled_repo_names (repo_id, name) values (%s, %s) on conflict do nothing;"
	return s.Exec(ctx, sqlf.Sprintf(q, repoId, repoName))
}

type writeAheadRecord struct {
	Id       uint32
	SeriesId string
	RepoId   uint32
	Capture  *string
	time     uint32
	value    float64
}

var writeAheadColumns = []string{"series_id", "repo_id", "capture", "time", "value"}
var writeAheadTable = "samples_write_ahead"

func (s *sampleStore) AppendWriteAhead(ctx context.Context, samples []RepoSample) (err error) {
	tx, err := s.transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	inserter := batch.NewInserter(ctx, tx.Handle(), writeAheadTable, batch.MaxNumPostgresParameters, writeAheadColumns...)
	for _, sample := range samples {
		if err = inserter.Insert(ctx, sample.SeriesId, sample.RepoId, sample.Capture, sample.Time, sample.Value); err != nil {
			return errors.Wrap(err, "inserter.Insert")
		}
	}
	return inserter.Flush(ctx)
}

// Compact will truncate the write ahead records and compress the values for a given time series key.
func (s *sampleStore) Compact(ctx context.Context, key TimeSeriesKey) error {
	// consider locking the entire table?

	tx, err := s.transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	type whKey struct {
		seriesId uint32
		repoId   uint32
		capture  string
		time     uint32
	}

	uniques := make(map[whKey]float64)

	var preds []*sqlf.Query
	q := "delete from samples_write_ahead where %s returning repo_id, capture, time, value;"
	preds = append(preds, sqlf.Sprintf("series_id = %s", key.SeriesId))
	preds = append(preds, sqlf.Sprintf("repo_id = %s", key.RepoId))
	if key.Capture == nil {
		preds = append(preds, sqlf.Sprintf("capture is null"))
	} else {
		preds = append(preds, sqlf.Sprintf("capture = %s", *key.Capture))
	}

	if err = tx.query(ctx, sqlf.Sprintf(q, sqlf.Join(preds, "AND")), func(sc scanner) (err error) {
		var tmp whKey
		var val float64
		if err := sc.Scan(
			&tmp.seriesId,
			&tmp.repoId,
			&tmp.capture,
			&tmp.time,
			&val,
		); err != nil {
			return err
		}
		if _, ok := uniques[tmp]; ok {
			// duplicate, skip the row
			return nil
		}
		uniques[tmp] = val
		return nil
	}); err != nil {
		return err
	}

	// select some rows
	// group
	// for each group select compressed row
	// what do we do if there are failures writing? :sad:

	return nil
}
