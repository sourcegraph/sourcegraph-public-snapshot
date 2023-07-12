package store

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"go.opentelemetry.io/otel/attribute"

	btypes "github.com/sourcegraph/sourcegraph/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// batchSpecExecutionCacheEntryInsertColumns is the list of
// batch_spec_execution_cache_entry columns that are modified in
// CreateBatchSpecExecutionCacheEntry
var batchSpecExecutionCacheEntryInsertColumns = SQLColumns{
	"user_id",
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
	"batch_spec_execution_cache_entries.user_id",
	"batch_spec_execution_cache_entries.key",
	"batch_spec_execution_cache_entries.value",
	"batch_spec_execution_cache_entries.version",
	"batch_spec_execution_cache_entries.last_used_at",
	"batch_spec_execution_cache_entries.created_at",
}

// CreateBatchSpecExecutionCacheEntry creates the given batch spec workspace jobs.
func (s *Store) CreateBatchSpecExecutionCacheEntry(ctx context.Context, ce *btypes.BatchSpecExecutionCacheEntry) (err error) {
	ctx, _, endObservation := s.operations.createBatchSpecExecutionCacheEntry.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.String("Key", ce.Key),
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
		ce.UserID,
		ce.Key,
		ce.Value,
		ce.Version,
		&dbutil.NullTime{Time: lastUsedAt},
		ce.CreatedAt,
		sqlf.Join(BatchSpecExecutionCacheEntryColums.ToSqlf(), ", "),
	)
}

var createBatchSpecExecutionCacheEntryQueryFmtstr = `
INSERT INTO batch_spec_execution_cache_entries (%s)
VALUES ` + batchSpecExecutionCacheEntryInsertColumns.FmtStr() + `
ON CONFLICT ON CONSTRAINT batch_spec_execution_cache_entries_user_id_key_unique
DO UPDATE SET
	value = EXCLUDED.value,
	version = EXCLUDED.version,
	created_at = EXCLUDED.created_at
RETURNING %s
`

// ListBatchSpecExecutionCacheEntriesOpts captures the query options needed for getting a BatchSpecExecutionCacheEntry
type ListBatchSpecExecutionCacheEntriesOpts struct {
	Keys   []string
	UserID int32
	// If true, explicitly return all entires.
	All bool
}

// ListBatchSpecExecutionCacheEntries gets the BatchSpecExecutionCacheEntries matching the given options.
func (s *Store) ListBatchSpecExecutionCacheEntries(ctx context.Context, opts ListBatchSpecExecutionCacheEntriesOpts) (cs []*btypes.BatchSpecExecutionCacheEntry, err error) {
	ctx, _, endObservation := s.operations.listBatchSpecExecutionCacheEntries.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("Count", len(opts.Keys)),
	}})
	defer endObservation(1, observation.Args{})

	if !opts.All && opts.UserID == 0 {
		return nil, errors.New("cannot query cache entries without specifying UserID")
	}

	if !opts.All && len(opts.Keys) == 0 {
		return nil, errors.New("cannot query cache entries without specifying Keys")
	}

	q := listBatchSpecExecutionCacheEntriesQuery(&opts)

	cs = make([]*btypes.BatchSpecExecutionCacheEntry, 0, len(opts.Keys))
	err = s.query(ctx, q, func(sc dbutil.Scanner) error {
		var c btypes.BatchSpecExecutionCacheEntry
		if err := scanBatchSpecExecutionCacheEntry(&c, sc); err != nil {
			return err
		}
		cs = append(cs, &c)
		return nil
	})

	return cs, err
}

var listBatchSpecExecutionCacheEntriesQueryFmtstr = `
SELECT %s FROM batch_spec_execution_cache_entries
WHERE %s
`

func listBatchSpecExecutionCacheEntriesQuery(opts *ListBatchSpecExecutionCacheEntriesOpts) *sqlf.Query {
	preds := []*sqlf.Query{
		// Only consider records that are in the current cache version.
		sqlf.Sprintf("batch_spec_execution_cache_entries.version = %s", btypes.CurrentCacheVersion),
	}

	if opts.UserID != 0 {
		preds = append(preds, sqlf.Sprintf("batch_spec_execution_cache_entries.user_id = %s", opts.UserID))
	}
	if len(opts.Keys) > 0 {
		preds = append(preds, sqlf.Sprintf("batch_spec_execution_cache_entries.key = ANY (%s)", pq.Array(opts.Keys)))
	}
	return sqlf.Sprintf(
		listBatchSpecExecutionCacheEntriesQueryFmtstr,
		sqlf.Join(BatchSpecExecutionCacheEntryColums.ToSqlf(), ", "),
		sqlf.Join(preds, "\n AND "),
	)
}

const markUsedBatchSpecExecutionCacheEntriesQueryFmtstr = `
UPDATE
	batch_spec_execution_cache_entries
SET last_used_at = %s
WHERE
	batch_spec_execution_cache_entries.id = ANY (%s)
`

// MarkUsedBatchSpecExecutionCacheEntries updates the LastUsedAt of the given cache entries.
func (s *Store) MarkUsedBatchSpecExecutionCacheEntries(ctx context.Context, ids []int64) (err error) {
	ctx, _, endObservation := s.operations.markUsedBatchSpecExecutionCacheEntries.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("count", len(ids)),
	}})
	defer endObservation(1, observation.Args{})

	q := sqlf.Sprintf(
		markUsedBatchSpecExecutionCacheEntriesQueryFmtstr,
		s.now(),
		pq.Array(ids),
	)
	return s.Exec(ctx, q)
}

// cleanBatchSpecExecutionEntriesQueryFmtstr collects cache entries to delete by
// collecting enough so that if we were to delete them we'd be under
// maxCacheSize again. Also, cache entries from older cache versions are always
// deleted.
const cleanBatchSpecExecutionEntriesQueryFmtstr = `
WITH total_size AS (
  SELECT sum(octet_length(value)) AS total FROM batch_spec_execution_cache_entries
),
candidates AS (
  SELECT
    id
  FROM (
    SELECT
      entries.id,
      entries.created_at,
      entries.last_used_at,
      SUM(octet_length(entries.value)) OVER (ORDER BY COALESCE(entries.last_used_at, entries.created_at) ASC, entries.id ASC) AS running_size
    FROM batch_spec_execution_cache_entries entries
  ) t
  WHERE
    ((SELECT total FROM total_size) - t.running_size) >= %s
),
outdated AS (
	SELECT
		id
	FROM batch_spec_execution_cache_entries
	WHERE
		version < %s
),
ids AS (
	SELECT id FROM outdated
	UNION ALL
	SELECT id FROM candidates
)
DELETE FROM batch_spec_execution_cache_entries WHERE id IN (SELECT id FROM ids)
`

func (s *Store) CleanBatchSpecExecutionCacheEntries(ctx context.Context, maxCacheSize int64) (err error) {
	ctx, _, endObservation := s.operations.cleanBatchSpecExecutionCacheEntries.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("MaxTableSize", int(maxCacheSize)),
	}})
	defer endObservation(1, observation.Args{})

	return s.Exec(ctx, sqlf.Sprintf(cleanBatchSpecExecutionEntriesQueryFmtstr, maxCacheSize, btypes.CurrentCacheVersion))
}

func scanBatchSpecExecutionCacheEntry(wj *btypes.BatchSpecExecutionCacheEntry, s dbutil.Scanner) error {
	return s.Scan(
		&wj.ID,
		&wj.UserID,
		&wj.Key,
		&wj.Value,
		&wj.Version,
		&dbutil.NullTime{Time: &wj.LastUsedAt},
		&wj.CreatedAt,
	)
}
