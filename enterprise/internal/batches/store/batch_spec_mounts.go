package store

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"

	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var batchSpecMountColumns = []*sqlf.Query{
	sqlf.Sprintf("batch_spec_mounts.id"),
	sqlf.Sprintf("batch_spec_mounts.rand_id"),
	sqlf.Sprintf("batch_spec_mounts.batch_spec_id"),
	sqlf.Sprintf("batch_spec_mounts.filename"),
	sqlf.Sprintf("batch_spec_mounts.path"),
	sqlf.Sprintf("batch_spec_mounts.size"),
	sqlf.Sprintf("batch_spec_mounts.modified"),
	sqlf.Sprintf("batch_spec_mounts.created_at"),
	sqlf.Sprintf("batch_spec_mounts.updated_at"),
}

var batchSpecMountInsertColumns = []*sqlf.Query{
	sqlf.Sprintf("rand_id"),
	sqlf.Sprintf("batch_spec_id"),
	sqlf.Sprintf("filename"),
	sqlf.Sprintf("path"),
	sqlf.Sprintf("size"),
	sqlf.Sprintf("modified"),
	sqlf.Sprintf("updated_at"),
}

var batchSpecMountConflictTarget = []*sqlf.Query{
	sqlf.Sprintf("batch_spec_id"),
	sqlf.Sprintf("filename"),
	sqlf.Sprintf("path"),
}

func (s *Store) UpsertBatchSpecMount(ctx context.Context, mount *btypes.BatchSpecMount) (err error) {
	ctx, _, endObservation := s.operations.upsertBatchSpecMount.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	q, err := s.upsertBatchSpecMountQuery(mount)
	if err != nil {
		return err
	}

	return s.query(ctx, q, func(sc dbutil.Scanner) (err error) {
		return scanBatchSpecMount(mount, sc)
	})
}

func (s *Store) upsertBatchSpecMountQuery(m *btypes.BatchSpecMount) (*sqlf.Query, error) {
	m.UpdatedAt = time.Now()

	if m.RandID == "" {
		var err error
		if m.RandID, err = RandomID(); err != nil {
			return nil, errors.Wrap(err, "creating RandID failed")
		}
	}

	return sqlf.Sprintf(
		upsertBatchSpecMountQueryFmtstr,
		sqlf.Join(batchSpecMountInsertColumns, ", "),
		m.RandID,
		m.BatchSpecID,
		m.FileName,
		m.Path,
		m.Size,
		m.Modified,
		m.UpdatedAt,
		sqlf.Join(batchSpecMountConflictTarget, ", "),
		sqlf.Join(batchSpecMountInsertColumns, ", "),
		m.RandID,
		m.BatchSpecID,
		m.FileName,
		m.Path,
		m.Size,
		m.Modified,
		m.UpdatedAt,
		sqlf.Join(batchSpecMountColumns, ", "),
	), nil
}

var upsertBatchSpecMountQueryFmtstr = `
-- source: enterprise/internal/batches/store/batch_spec_mounts.go:UpsertBatchSpecMount
INSERT INTO batch_spec_mounts (%s)
VALUES (%s, %s, %s, %s, %s, %s, %s)
ON CONFLICT (%s) WHERE TRUE
DO UPDATE SET
(%s) = (%s, %s, %s, %s, %s, %s, %s)
RETURNING %s`

func (s *Store) CreateBatchSpecMount(ctx context.Context, mount *btypes.BatchSpecMount) (err error) {
	ctx, _, endObservation := s.operations.createBatchSpecMount.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	q, err := s.createBatchSpecMountQuery(mount)
	if err != nil {
		return err
	}

	return s.query(ctx, q, func(sc dbutil.Scanner) (err error) {
		return scanBatchSpecMount(mount, sc)
	})
}

func (s *Store) createBatchSpecMountQuery(m *btypes.BatchSpecMount) (*sqlf.Query, error) {
	if m.UpdatedAt.IsZero() {
		m.UpdatedAt = time.Now()
	}

	if m.RandID == "" {
		var err error
		if m.RandID, err = RandomID(); err != nil {
			return nil, errors.Wrap(err, "creating RandID failed")
		}
	}

	return sqlf.Sprintf(
		createBatchSpecMountQueryFmtstr,
		sqlf.Join(batchSpecMountInsertColumns, ", "),
		m.BatchSpecID,
		m.FileName,
		m.Path,
		m.Size,
		m.Modified,
		m.CreatedAt,
		m.UpdatedAt,
		sqlf.Join(batchSpecMountColumns, ", "),
	), nil
}

var createBatchSpecMountQueryFmtstr = `
-- source: enterprise/internal/batches/store/batch_spec_mounts.go:CreateBatchSpecMount
INSERT INTO batch_spec_mounts (%s)
VALUES (%s, %s, %s, %s, %s, %s, %s)
RETURNING %s`

func (s *Store) UpdateBatchSpecMount(ctx context.Context, mount *btypes.BatchSpecMount) (err error) {
	ctx, _, endObservation := s.operations.updateBatchSpecMount.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("ID", int(mount.ID)),
	}})
	defer endObservation(1, observation.Args{})

	q := s.updateBatchSpecMountQuery(mount)

	return s.query(ctx, q, func(sc dbutil.Scanner) (err error) {
		return scanBatchSpecMount(mount, sc)
	})
}

func (s *Store) updateBatchSpecMountQuery(m *btypes.BatchSpecMount) *sqlf.Query {
	m.UpdatedAt = s.now()

	return sqlf.Sprintf(
		updateBatchSpecMountQueryFmtstr,
		sqlf.Join(batchSpecMountInsertColumns, ", "),
		m.BatchSpecID,
		m.FileName,
		m.Path,
		m.Size,
		m.Modified,
		m.CreatedAt,
		m.UpdatedAt,
		m.ID,
		sqlf.Join(batchChangeColumns, ", "),
	)
}

var updateBatchSpecMountQueryFmtstr = `
-- source: enterprise/internal/batches/store/batch_spec_mounts.go:UpdateBatchSpecMount
UPDATE batch_spec_mounts
SET (%s) = (%s, %s, %s, %s, %s, %s, %s)
WHERE id = %s
RETURNING %s`

type DeleteBatchSpecMountOpts struct {
	ID              int64
	BatchSpecID     int64
	BatchSpecRandID string
}

func (s *Store) DeleteBatchSpecMount(ctx context.Context, opts DeleteBatchSpecMountOpts) (err error) {
	ctx, _, endObservation := s.operations.getBatchChange.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("ID", int(opts.ID)),
	}})
	defer endObservation(1, observation.Args{})

	if opts.ID == 0 && opts.BatchSpecID == 0 {
		return errors.New("cannot delete entries without specifying an option")
	}

	return s.Store.Exec(ctx, deleteBatchSpecMountQuery(opts))
}

func deleteBatchSpecMountQuery(opts DeleteBatchSpecMountOpts) *sqlf.Query {
	var preds []*sqlf.Query
	var joins []*sqlf.Query

	if opts.ID != 0 {
		preds = append(preds, sqlf.Sprintf("batch_spec_mounts.id = %s", opts.ID))
	}

	if opts.BatchSpecID != 0 {
		preds = append(preds, sqlf.Sprintf("batch_spec_mounts.batch_spec_id = %s", opts.BatchSpecID))
	}

	if opts.BatchSpecRandID != "" {
		joins = append(joins, sqlf.Sprintf("INNER JOIN batch_spec ON batch_spec_mounts.batch_spec_id = batch_specs.id"))
		preds = append(preds, sqlf.Sprintf("batch_specs.rand_id = %s", opts.BatchSpecRandID))
	}

	return sqlf.Sprintf(
		deleteBatchSpecMountQueryFmtstr,
		sqlf.Join(joins, "\n"),
		sqlf.Join(preds, "\n AND "),
	)
}

var deleteBatchSpecMountQueryFmtstr = `
-- source: enterprise/internal/batches/store/batch_spec_mount.go:DeleteBatchSpecMount
DELETE FROM batch_spec_mounts
%s
WHERE %s`

type GetBatchSpecMountOpts struct {
	ID     int64
	RandID string
}

func (s *Store) GetBatchSpecMount(ctx context.Context, opts GetBatchSpecMountOpts) (mount *btypes.BatchSpecMount, err error) {
	ctx, _, endObservation := s.operations.getBatchSpecMount.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("ID", int(opts.ID)),
		log.String("RandID", opts.RandID),
	}})
	defer endObservation(1, observation.Args{})

	if opts.ID == 0 && opts.RandID == "" {
		return nil, errors.New("invalid option: require at least one ID to be provided")
	}

	q := getBatchSpecMountQuery(opts)

	var m btypes.BatchSpecMount
	err = s.query(ctx, q, func(sc dbutil.Scanner) (err error) {
		return scanBatchSpecMount(&m, sc)
	})
	if err != nil {
		return nil, err
	}
	if m.ID == 0 {
		return nil, ErrNoResults
	}

	return &m, nil
}

func getBatchSpecMountQuery(opts GetBatchSpecMountOpts) *sqlf.Query {
	var preds []*sqlf.Query

	if opts.ID != 0 {
		preds = append(preds, sqlf.Sprintf("batch_spec_mounts.id = %s", opts.ID))
	}

	if opts.RandID != "" {
		preds = append(preds, sqlf.Sprintf("batch_spec_mounts.rand_id = %s", opts.RandID))
	}

	return sqlf.Sprintf(
		getBatchSpecMountQueryFmtstr,
		sqlf.Join(batchSpecMountColumns, ", "),
		sqlf.Join(preds, "\n AND "),
	)
}

var getBatchSpecMountQueryFmtstr = `
-- source: enterprise/internal/batches/store/batch_spec_mounts.go:GetBatchSpecMount
SELECT %s FROM batch_spec_mounts
WHERE %s
LIMIT 1`

type ListBatchSpecMountsOpts struct {
	LimitOpts
	Cursor int64

	BatchSpecID     int64
	BatchSpecRandID string
}

func (s *Store) CountBatchSpecMounts(ctx context.Context, opts ListBatchSpecMountsOpts) (count int64, err error) {
	ctx, _, endObservation := s.operations.countBatchSpecMounts.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	q := countBatchSpecMountsQuery(opts)

	count, _, err = basestore.ScanFirstInt64(s.Query(ctx, q))
	return count, err
}

func countBatchSpecMountsQuery(opts ListBatchSpecMountsOpts) *sqlf.Query {
	var preds []*sqlf.Query
	var joins []*sqlf.Query

	if opts.BatchSpecID != 0 {
		preds = append(preds, sqlf.Sprintf("batch_spec_mounts.batch_spec_id = %s", opts.BatchSpecID))
	}

	if opts.BatchSpecRandID != "" {
		joins = append(joins, sqlf.Sprintf("INNER JOIN batch_specs ON batch_spec_mounts.batch_spec_id = batch_specs.id"))
		preds = append(preds, sqlf.Sprintf("batch_specs.rand_id = %s", opts.BatchSpecRandID))
	}

	if len(preds) == 0 {
		preds = append(preds, sqlf.Sprintf("TRUE"))
	}

	return sqlf.Sprintf(
		countBatchSpecMountQueryFmtstr,
		sqlf.Join(joins, "\n"),
		sqlf.Join(preds, "\n AND "),
	)
}

var countBatchSpecMountQueryFmtstr = `
-- source: enterprise/internal/batches/store/batch_spec_mounts.go:ListBatchSpecMounts
SELECT COUNT(1) FROM batch_spec_mounts
%s
WHERE %s`

func (s *Store) ListBatchSpecMounts(ctx context.Context, opts ListBatchSpecMountsOpts) (mounts []*btypes.BatchSpecMount, next int64, err error) {
	ctx, _, endObservation := s.operations.listBatchSpecMounts.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	q := listBatchSpecMountsQuery(opts)

	mounts = make([]*btypes.BatchSpecMount, 0, opts.DBLimit())
	err = s.query(ctx, q, func(sc dbutil.Scanner) (err error) {
		var m btypes.BatchSpecMount
		if err := scanBatchSpecMount(&m, sc); err != nil {
			return err
		}
		mounts = append(mounts, &m)
		return nil
	})

	if opts.Limit != 0 && len(mounts) == opts.DBLimit() {
		next = mounts[len(mounts)-1].ID
		mounts = mounts[:len(mounts)-1]
	}

	return mounts, next, err
}

func listBatchSpecMountsQuery(opts ListBatchSpecMountsOpts) *sqlf.Query {
	var preds []*sqlf.Query
	var joins []*sqlf.Query

	if opts.Cursor != 0 {
		preds = append(preds, sqlf.Sprintf("batch_spec_mounts.id <= %s", opts.Cursor))
	}

	if opts.BatchSpecID != 0 {
		preds = append(preds, sqlf.Sprintf("batch_spec_mounts.batch_spec_id = %s", opts.BatchSpecID))
	}

	if opts.BatchSpecRandID != "" {
		joins = append(joins, sqlf.Sprintf("INNER JOIN batch_specs ON batch_spec_mounts.batch_spec_id = batch_specs.id"))
		preds = append(preds, sqlf.Sprintf("batch_specs.rand_id = %s", opts.BatchSpecRandID))
	}

	if len(preds) == 0 {
		preds = append(preds, sqlf.Sprintf("TRUE"))
	}

	return sqlf.Sprintf(
		listBatchSpecMountQueryFmtstr,
		sqlf.Join(batchSpecMountColumns, ", "),
		sqlf.Join(joins, "\n"),
		sqlf.Join(preds, "\n AND "),
	)
}

var listBatchSpecMountQueryFmtstr = `
-- source: enterprise/internal/batches/store/batch_spec_mounts.go:ListBatchSpecMounts
SELECT %s FROM batch_spec_mounts
%s
WHERE %s`

func scanBatchSpecMount(m *btypes.BatchSpecMount, s dbutil.Scanner) error {
	return s.Scan(
		&m.ID,
		&m.RandID,
		&m.BatchSpecID,
		&m.FileName,
		&m.Path,
		&m.Size,
		&m.Modified,
		&m.CreatedAt,
		&m.UpdatedAt,
	)
}
