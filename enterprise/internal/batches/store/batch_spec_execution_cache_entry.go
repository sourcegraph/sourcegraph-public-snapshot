package store

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"

	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// batchSpecExecutionCacheEntryInsertColumns is the list of
// batch_spec_execution_cache_entry columns that are modified in
// CreateBatchSpecExecutionCacheEntry
var batchSpecExecutionCacheEntryInsertColumns = SQLColumns{
	"key",
	"value",
	"version",
	"last_used_at",
	"created_at",
}

// BatchSpecExecutionCacheEntryColums are used by the changeset job related Store methods to query
// and create changeset jobs.
var BatchSpecExecutionCacheEntryColums = SQLColumns{
	"batch_spec_execution_cache_entries.id",
	"batch_spec_execution_cache_entries.key",
	"batch_spec_execution_cache_entries.value",
	"batch_spec_execution_cache_entries.version",
	"batch_spec_execution_cache_entries.last_used_at",
	"batch_spec_execution_cache_entries.created_at",
}

// CreateBatchSpecExecutionCacheEntry creates the given batch spec workspace jobs.
func (s *Store) CreateBatchSpecExecutionCacheEntry(ctx context.Context, ce *btypes.BatchSpecExecutionCacheEntry) (err error) {
	ctx, endObservation := s.operations.createBatchSpecExecutionCacheEntry.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("Key", ce.Key),
	}})
	defer endObservation(1, observation.Args{})

	q := s.createBatchSpecExecutionCacheEntryQuery(ce)

	err = s.query(ctx, q, func(sc dbutil.Scanner) (err error) {
		return scanBatchSpecExecutionCacheEntry(ce, sc)
	})

	return err
}

func (s *Store) createBatchSpecExecutionCacheEntryQuery(ce *btypes.BatchSpecExecutionCacheEntry) *sqlf.Query {
	if ce.CreatedAt.IsZero() {
		ce.CreatedAt = s.now()
	}

	if ce.Version == 0 {
		ce.Version = btypes.CurrentCacheVersion
	}

	lastUsedAt := &ce.LastUsedAt
	if ce.LastUsedAt.IsZero() {
		lastUsedAt = nil
	}

	return sqlf.Sprintf(
		createBatchSpecExecutionCacheEntryQueryFmtstr,
		sqlf.Join(batchSpecExecutionCacheEntryInsertColumns.ToSqlf(), ", "),
		ce.Key,
		ce.Value,
		ce.Version,
		&dbutil.NullTime{Time: lastUsedAt},
		ce.CreatedAt,
		sqlf.Join(BatchSpecExecutionCacheEntryColums.ToSqlf(), ", "),
	)
}

var createBatchSpecExecutionCacheEntryQueryFmtstr = `
-- source: enterprise/internal/batches/store/batch_spec_execution_cache_entry.go:CreateBatchSpecExecutionCacheEntry
INSERT INTO batch_spec_execution_cache_entries (%s)
VALUES ` + batchSpecExecutionCacheEntryInsertColumns.FmtStr() + `
RETURNING %s
`

// GetBatchSpecExecutionCacheEntryOpts captures the query options needed for getting a BatchSpecExecutionCacheEntry
type GetBatchSpecExecutionCacheEntryOpts struct {
	Key string
}

// GetBatchSpecExecutionCacheEntry gets a BatchSpecExecutionCacheEntry matching the given options.
func (s *Store) GetBatchSpecExecutionCacheEntry(ctx context.Context, opts GetBatchSpecExecutionCacheEntryOpts) (job *btypes.BatchSpecExecutionCacheEntry, err error) {
	ctx, endObservation := s.operations.getBatchSpecExecutionCacheEntry.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("Key", opts.Key),
	}})
	defer endObservation(1, observation.Args{})

	q := getBatchSpecExecutionCacheEntryQuery(&opts)
	var c btypes.BatchSpecExecutionCacheEntry
	err = s.query(ctx, q, func(sc dbutil.Scanner) (err error) {
		return scanBatchSpecExecutionCacheEntry(&c, sc)
	})
	if err != nil {
		return nil, err
	}

	if c.ID == 0 {
		return nil, ErrNoResults
	}

	return &c, nil
}

var getBatchSpecExecutionCacheEntrysQueryFmtstr = `
-- source: enterprise/internal/batches/store/batch_spec_execution_cache_entry.go:GetBatchSpecExecutionCacheEntry
SELECT %s FROM batch_spec_execution_cache_entries
WHERE %s
LIMIT 1
`

func getBatchSpecExecutionCacheEntryQuery(opts *GetBatchSpecExecutionCacheEntryOpts) *sqlf.Query {
	preds := []*sqlf.Query{
		sqlf.Sprintf("batch_spec_execution_cache_entries.key = %s", opts.Key),
	}

	return sqlf.Sprintf(
		getBatchSpecExecutionCacheEntrysQueryFmtstr,
		sqlf.Join(BatchSpecExecutionCacheEntryColums.ToSqlf(), ", "),
		sqlf.Join(preds, "\n AND "),
	)
}

const markUsedBatchSpecExecutionCacheEntryQueryFmtstr = `
-- source: enterprise/internal/batches/store/batch_spec_execution_cache_entry.go:MarkUsedBatchSpecExecutionCacheEntry
UPDATE
	batch_spec_execution_cache_entries
SET last_used_at = %s
WHERE
	batch_spec_execution_cache_entries.id = %s
`

// MarkUsedBatchSpecExecutionCacheEntry updates the LastUsedAt of the given cache entry.
func (s *Store) MarkUsedBatchSpecExecutionCacheEntry(ctx context.Context, id int64) (err error) {
	ctx, endObservation := s.operations.markUsedBatchSpecExecutionCacheEntry.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("ID", int(id)),
	}})
	defer endObservation(1, observation.Args{})

	q := sqlf.Sprintf(
		markUsedBatchSpecExecutionCacheEntryQueryFmtstr,
		s.now(),
		id,
	)
	return s.Exec(ctx, q)
}

func scanBatchSpecExecutionCacheEntry(wj *btypes.BatchSpecExecutionCacheEntry, s dbutil.Scanner) error {
	return s.Scan(
		&wj.ID,
		&wj.Key,
		&wj.Value,
		&wj.Version,
		&dbutil.NullTime{Time: &wj.LastUsedAt},
		&wj.CreatedAt,
	)
}
