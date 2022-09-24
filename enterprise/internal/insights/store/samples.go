package store

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"sort"
	"time"

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
	StoreRow(ctx context.Context, row UncompressedRow, seriesId uint32) error
	LoadRows(ctx context.Context, opts SeriesPointsOpts) ([]UncompressedRow, error)
	Append(ctx context.Context, key TimeSeriesKey, samples []RawSample) (err error)
}

type sampleStore struct {
	*basestore.Store
	permStore InsightPermissionStore
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
	Samples []RawSample
}

type CompressedRow struct {
	altFormatRowMetadata
	Data []byte
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

func (s *sampleStore) LoadRows(ctx context.Context, opts SeriesPointsOpts) ([]UncompressedRow, error) {
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
		})
		return nil
	})
}

func loadRowsQuery(opts SeriesPointsOpts) *sqlf.Query {
	baseQuery := `select sp.id, repo_id, data, capture from series_points_compressed sp`
	var preds []*sqlf.Query

	// todo(insights): add repo filtering
	if opts.SeriesID != nil {
		preds = append(preds, sqlf.Sprintf("series_id = (select isn.id from insight_series as isn where isn.series_id = %s)", *opts.SeriesID))
	}
	if opts.RepoID != nil {
		preds = append(preds, sqlf.Sprintf("repo_id = %s", *opts.RepoID))
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
	if len(preds) > 0 {
		baseQuery += " where %s"
	}

	final := sqlf.Sprintf(baseQuery, sqlf.Join(preds, "AND"))
	return final
}

func (s *sampleStore) Append(ctx context.Context, key TimeSeriesKey, samples []RawSample) (err error) {
	tx, err := s.transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	var keyMatch *UncompressedRow
	err = tx.streamRows(ctx, SeriesPointsOpts{Key: &key}, func(ctx context.Context, row *CompressedRow) error {
		decompressed, err := decompressSamples(row.Data)
		if err != nil {
			return errors.Wrap(err, "failed to decompress sample data")
		}
		keyMatch = &UncompressedRow{
			altFormatRowMetadata: row.altFormatRowMetadata,
			Samples:              decompressed,
		}
		return nil
	})
	if err != nil {
		return errors.Wrap(err, "failed to decompress row")
	}

	if keyMatch == nil {
		// no data exists already for this key, so we will write a new row with the provided samples
		keyMatch = &UncompressedRow{
			altFormatRowMetadata: altFormatRowMetadata{
				RepoId:  key.RepoId,
				Capture: key.Capture,
			},
			Samples: nil, // not setting this because this is the "original" samples, we will append the new ones later
		}
	}
	keyMatch.Samples = append(keyMatch.Samples, samples...)
	err = tx.StoreRow(ctx, *keyMatch, key.SeriesId)
	if err != nil {
		return errors.Wrap(err, "StoreAlternateFormat")
	}
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

func (s *sampleStore) streamRows(ctx context.Context, opts SeriesPointsOpts, callback rowStreamFunc) error {
	return s.query(ctx, loadRowsQuery(opts), func(s scanner) (err error) {
		var tmp CompressedRow
		if err := s.Scan(
			&tmp.Id,
			&tmp.RepoId,
			&tmp.Data,
			&tmp.Capture,
		); err != nil {
			return err
		}

		return callback(ctx, &tmp)
	})
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

func ToTimeseries(data []UncompressedRow, seriesId string) (results []SeriesPoint) {
	getKey := func(s *string) string {
		if s == nil {
			return ""
		}
		return *s
	}

	byCapture := make(map[string][]UncompressedRow)
	for _, datum := range data {
		byCapture[getKey(datum.Capture)] = append(byCapture[getKey(datum.Capture)], datum)
	}

	for key, vals := range byCapture {
		mapped := make(map[uint32]float64)

		for _, val := range vals {
			for _, sample := range val.Samples {
				mapped[sample.Time] += sample.Value
			}
		}

		toPtr := func(s string) *string {
			if s == "" {
				return nil
			}
			return &s
		}

		for utime, agg := range mapped {
			results = append(results, SeriesPoint{
				SeriesID: seriesId,
				Time:     time.Unix(int64(utime), 0),
				Value:    agg,
				Capture:  toPtr(key),
			})
		}
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].Time.Before(results[j].Time)
	})
	return results
}
