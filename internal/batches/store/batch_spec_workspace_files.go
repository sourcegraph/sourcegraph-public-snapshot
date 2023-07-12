package store

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"go.opentelemetry.io/otel/attribute"

	btypes "github.com/sourcegraph/sourcegraph/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var batchSpecWorkspaceFileColumns = []*sqlf.Query{
	sqlf.Sprintf("batch_spec_workspace_files.id"),
	sqlf.Sprintf("batch_spec_workspace_files.rand_id"),
	sqlf.Sprintf("batch_spec_workspace_files.batch_spec_id"),
	sqlf.Sprintf("batch_spec_workspace_files.filename"),
	sqlf.Sprintf("batch_spec_workspace_files.path"),
	sqlf.Sprintf("batch_spec_workspace_files.size"),
	sqlf.Sprintf("batch_spec_workspace_files.content"),
	sqlf.Sprintf("batch_spec_workspace_files.modified_at"),
	sqlf.Sprintf("batch_spec_workspace_files.created_at"),
	sqlf.Sprintf("batch_spec_workspace_files.updated_at"),
}

var batchSpecWorkspaceFileInsertColumns = []*sqlf.Query{
	sqlf.Sprintf("rand_id"),
	sqlf.Sprintf("batch_spec_id"),
	sqlf.Sprintf("filename"),
	sqlf.Sprintf("path"),
	sqlf.Sprintf("size"),
	sqlf.Sprintf("content"),
	sqlf.Sprintf("modified_at"),
	sqlf.Sprintf("updated_at"),
}

var batchSpecWorkspaceFileConflictTarget = []*sqlf.Query{
	sqlf.Sprintf("batch_spec_id"),
	sqlf.Sprintf("filename"),
	sqlf.Sprintf("path"),
}

// UpsertBatchSpecWorkspaceFile creates a new BatchSpecWorkspaceFile, if it does not exist already, or updates the existing
// BatchSpecWorkspaceFile.
func (s *Store) UpsertBatchSpecWorkspaceFile(ctx context.Context, file *btypes.BatchSpecWorkspaceFile) (err error) {
	ctx, _, endObservation := s.operations.upsertBatchSpecWorkspaceFile.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	q, err := s.upsertBatchSpecWorkspaceFileQuery(file)
	if err != nil {
		return err
	}

	return s.query(ctx, q, func(sc dbutil.Scanner) (err error) {
		return scanBatchSpecWorkspaceFile(file, sc)
	})
}

func (s *Store) upsertBatchSpecWorkspaceFileQuery(m *btypes.BatchSpecWorkspaceFile) (*sqlf.Query, error) {
	m.UpdatedAt = s.now()

	if m.RandID == "" {
		var err error
		if m.RandID, err = RandomID(); err != nil {
			return nil, errors.Wrap(err, "creating RandID failed")
		}
	}

	return sqlf.Sprintf(
		upsertBatchSpecWorkspaceFileQueryFmtstr,
		sqlf.Join(batchSpecWorkspaceFileInsertColumns, ", "),
		m.RandID,
		m.BatchSpecID,
		m.FileName,
		m.Path,
		m.Size,
		m.Content,
		m.ModifiedAt,
		m.UpdatedAt,
		sqlf.Join(batchSpecWorkspaceFileConflictTarget, ", "),
		sqlf.Join(batchSpecWorkspaceFileInsertColumns, ", "),
		m.RandID,
		m.BatchSpecID,
		m.FileName,
		m.Path,
		m.Size,
		m.Content,
		m.ModifiedAt,
		m.UpdatedAt,
		sqlf.Join(batchSpecWorkspaceFileColumns, ", "),
	), nil
}

var upsertBatchSpecWorkspaceFileQueryFmtstr = `
INSERT INTO batch_spec_workspace_files (%s)
VALUES (%s, %s, %s, %s, %s, %s, %s, %s)
ON CONFLICT (%s) WHERE TRUE
DO UPDATE SET
(%s) = (%s, %s, %s, %s, %s, %s, %s, %s)
RETURNING %s`

// DeleteBatchSpecWorkspaceFileOpts are the options to determine which BatchSpecWorkspaceFiles to delete.
type DeleteBatchSpecWorkspaceFileOpts struct {
	ID          int64
	BatchSpecID int64
}

// DeleteBatchSpecWorkspaceFile deletes BatchSpecWorkspaceFiles that match the specified DeleteBatchSpecWorkspaceFileOpts.
func (s *Store) DeleteBatchSpecWorkspaceFile(ctx context.Context, opts DeleteBatchSpecWorkspaceFileOpts) (err error) {
	ctx, _, endObservation := s.operations.deleteBatchSpecWorkspaceFile.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("ID", int(opts.ID)),
	}})
	defer endObservation(1, observation.Args{})

	if opts.ID == 0 && opts.BatchSpecID == 0 {
		return errors.New("cannot delete entries without specifying an option")
	}

	q := deleteBatchSpecWorkspaceFileQuery(opts)
	return s.Store.Exec(ctx, q)
}

func deleteBatchSpecWorkspaceFileQuery(opts DeleteBatchSpecWorkspaceFileOpts) *sqlf.Query {
	var preds []*sqlf.Query

	if opts.ID != 0 {
		preds = append(preds, sqlf.Sprintf("batch_spec_workspace_files.id = %s", opts.ID))
	}

	if opts.BatchSpecID != 0 {
		preds = append(preds, sqlf.Sprintf("batch_spec_workspace_files.batch_spec_id = %s", opts.BatchSpecID))
	}

	return sqlf.Sprintf(
		deleteBatchSpecWorkspaceFileQueryFmtstr,
		sqlf.Join(preds, "\n AND "),
	)
}

var deleteBatchSpecWorkspaceFileQueryFmtstr = `
DELETE FROM batch_spec_workspace_files
WHERE %s`

// GetBatchSpecWorkspaceFileOpts are the options to determine which BatchSpecWorkspaceFile to retrieve.
type GetBatchSpecWorkspaceFileOpts struct {
	ID     int64
	RandID string
}

// GetBatchSpecWorkspaceFile retrieves the matching BatchSpecWorkspaceFile based on the provided GetBatchSpecWorkspaceFileOpts.
func (s *Store) GetBatchSpecWorkspaceFile(ctx context.Context, opts GetBatchSpecWorkspaceFileOpts) (file *btypes.BatchSpecWorkspaceFile, err error) {
	ctx, _, endObservation := s.operations.getBatchSpecWorkspaceFile.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("ID", int(opts.ID)),
		attribute.String("RandID", opts.RandID),
	}})
	defer endObservation(1, observation.Args{})

	if opts.ID == 0 && opts.RandID == "" {
		return nil, errors.New("invalid option: require at least one ID to be provided")
	}

	q := getBatchSpecWorkspaceFileQuery(opts)

	var m btypes.BatchSpecWorkspaceFile
	err = s.query(ctx, q, func(sc dbutil.Scanner) (err error) {
		return scanBatchSpecWorkspaceFile(&m, sc)
	})
	if err != nil {
		return nil, err
	}
	if m.ID == 0 {
		return nil, ErrNoResults
	}

	return &m, nil
}

func getBatchSpecWorkspaceFileQuery(opts GetBatchSpecWorkspaceFileOpts) *sqlf.Query {
	var preds []*sqlf.Query

	if opts.ID != 0 {
		preds = append(preds, sqlf.Sprintf("batch_spec_workspace_files.id = %s", opts.ID))
	}

	if opts.RandID != "" {
		preds = append(preds, sqlf.Sprintf("batch_spec_workspace_files.rand_id = %s", opts.RandID))
	}

	return sqlf.Sprintf(
		getBatchSpecWorkspaceFileQueryFmtstr,
		sqlf.Join(batchSpecWorkspaceFileColumns, ", "),
		sqlf.Join(preds, "\n AND "),
	)
}

var getBatchSpecWorkspaceFileQueryFmtstr = `
SELECT %s FROM batch_spec_workspace_files
WHERE %s
LIMIT 1`

// ListBatchSpecWorkspaceFileOpts are the options to determine which BatchSpecWorkspaceFiles to list.
type ListBatchSpecWorkspaceFileOpts struct {
	LimitOpts
	Cursor int64

	ID              int64
	RandID          string
	BatchSpecID     int64
	BatchSpecRandID string
}

// CountBatchSpecWorkspaceFiles counts the number of BatchSpecWorkspaceFiles based on the provided ListBatchSpecWorkspaceFileOpts.
func (s *Store) CountBatchSpecWorkspaceFiles(ctx context.Context, opts ListBatchSpecWorkspaceFileOpts) (count int, err error) {
	ctx, _, endObservation := s.operations.countBatchSpecWorkspaceFiles.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.queryCount(ctx, countBatchSpecWorkspaceFilesQuery(opts))
}

func countBatchSpecWorkspaceFilesQuery(opts ListBatchSpecWorkspaceFileOpts) *sqlf.Query {
	var preds []*sqlf.Query
	var joins []*sqlf.Query

	if opts.ID != 0 {
		preds = append(preds, sqlf.Sprintf("batch_spec_workspace_files.id = %s", opts.ID))
	}

	if opts.RandID != "" {
		preds = append(preds, sqlf.Sprintf("batch_spec_workspace_files.rand_id = %s", opts.RandID))
	}

	if opts.BatchSpecID != 0 {
		preds = append(preds, sqlf.Sprintf("batch_spec_workspace_files.batch_spec_id = %s", opts.BatchSpecID))
	}

	if opts.BatchSpecRandID != "" {
		joins = append(joins, sqlf.Sprintf("INNER JOIN batch_specs ON batch_spec_workspace_files.batch_spec_id = batch_specs.id"))
		preds = append(preds, sqlf.Sprintf("batch_specs.rand_id = %s", opts.BatchSpecRandID))
	}

	if len(preds) == 0 {
		preds = append(preds, sqlf.Sprintf("TRUE"))
	}

	return sqlf.Sprintf(
		countBatchSpecWorkspaceFileQueryFmtstr,
		sqlf.Join(joins, "\n"),
		sqlf.Join(preds, "\n AND "),
	)
}

var countBatchSpecWorkspaceFileQueryFmtstr = `
SELECT COUNT(1) FROM batch_spec_workspace_files
%s
WHERE %s`

// ListBatchSpecWorkspaceFiles retrieves the matching BatchSpecWorkspaceFiles that match the provided ListBatchSpecWorkspaceFileOpts.
func (s *Store) ListBatchSpecWorkspaceFiles(ctx context.Context, opts ListBatchSpecWorkspaceFileOpts) (files []*btypes.BatchSpecWorkspaceFile, next int64, err error) {
	ctx, _, endObservation := s.operations.listBatchSpecWorkspaceFiles.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	q := listBatchSpecWorkspaceFilesQuery(opts)

	files = make([]*btypes.BatchSpecWorkspaceFile, 0, opts.DBLimit())
	err = s.query(ctx, q, func(sc dbutil.Scanner) (err error) {
		var m btypes.BatchSpecWorkspaceFile
		if err := scanBatchSpecWorkspaceFile(&m, sc); err != nil {
			return err
		}
		files = append(files, &m)
		return nil
	})

	if opts.Limit != 0 && len(files) == opts.DBLimit() {
		next = files[len(files)-1].ID
		files = files[:len(files)-1]
	}

	return files, next, err
}

func listBatchSpecWorkspaceFilesQuery(opts ListBatchSpecWorkspaceFileOpts) *sqlf.Query {
	var preds []*sqlf.Query
	var joins []*sqlf.Query

	if opts.Cursor != 0 {
		preds = append(preds, sqlf.Sprintf("batch_spec_workspace_files.id <= %s", opts.Cursor))
	}

	if opts.BatchSpecID != 0 {
		preds = append(preds, sqlf.Sprintf("batch_spec_workspace_files.batch_spec_id = %s", opts.BatchSpecID))
	}

	if opts.BatchSpecRandID != "" {
		joins = append(joins, sqlf.Sprintf("INNER JOIN batch_specs ON batch_spec_workspace_files.batch_spec_id = batch_specs.id"))
		preds = append(preds, sqlf.Sprintf("batch_specs.rand_id = %s", opts.BatchSpecRandID))
	}

	if len(preds) == 0 {
		preds = append(preds, sqlf.Sprintf("TRUE"))
	}

	return sqlf.Sprintf(
		listBatchSpecWorkspaceFileQueryFmtstr+opts.LimitOpts.ToDB(),
		sqlf.Join(batchSpecWorkspaceFileColumns, ", "),
		sqlf.Join(joins, "\n"),
		sqlf.Join(preds, "\n AND "),
	)
}

var listBatchSpecWorkspaceFileQueryFmtstr = `
SELECT %s FROM batch_spec_workspace_files
%s
WHERE %s
ORDER BY id DESC
`

func scanBatchSpecWorkspaceFile(m *btypes.BatchSpecWorkspaceFile, s dbutil.Scanner) error {
	return s.Scan(
		&m.ID,
		&m.RandID,
		&m.BatchSpecID,
		&m.FileName,
		&m.Path,
		&m.Size,
		&m.Content,
		&m.ModifiedAt,
		&m.CreatedAt,
		&m.UpdatedAt,
	)
}
