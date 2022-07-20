package store

import (
	"context"
	"fmt"

	"github.com/keegancsmith/sqlf"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func (s *Store) CreateBatchChangeChangesetAssociation(
	ctx context.Context,
	assoc *btypes.BatchChangeChangesetAssociation,
) (err error) {
	ctx, _, endObservation := s.operations.createBatchChangeChangesetAssociation.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	q := sqlf.Sprintf(
		createBatchChangeChangesetAssociationFmtstr,
		assoc.BatchChangeID,
		assoc.ChangesetID,
		assoc.Detach,
		nullStringColumn(string(assoc.Archived)),
		sqlf.Join(batchChangeChangesetAssociationColumns, ","),
	)

	return s.query(ctx, q, func(sc dbutil.Scanner) error {
		return scanBatchChangeChangesetAssociation(assoc, sc)
	})
}

const createBatchChangeChangesetAssociationFmtstr = `
-- source: batch_change_changesets.go:CreateBatchChangeChangesetAssociation
INSERT INTO
  batch_change_changesets
  (batch_change_id, changeset_id, detach, archived)
VALUES
  (%s, %s, %s, %s)
RETURNING
  %s
`

func (s *Store) DeleteBatchChangeChangesetAssociation(
	ctx context.Context,
	assoc *btypes.BatchChangeChangesetAssociation,
) (err error) {
	ctx, _, endObservation := s.operations.deleteBatchChangeChangesetAssociation.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	q := sqlf.Sprintf(
		deleteBatchChangeChangesetAssociationFmtstr,
		assoc.BatchChangeID, assoc.ChangesetID,
	)
	return s.Store.Exec(ctx, q)
}

const deleteBatchChangeChangesetAssociationFmtstr = `
-- source: batch_change_changesets.go:DeleteBatchChangeChangesetAssociation
DELETE FROM
  batch_change_changesets
WHERE
  batch_change_id = %s AND changeset_id = %s
`

func (s *Store) GetBatchChangeChangesetAssociation(
	ctx context.Context,
	batchChangeID int64,
	changesetID int64,
) (assoc *btypes.BatchChangeChangesetAssociation, err error) {
	ctx, _, endObservation := s.operations.getBatchChangeChangesetAssociation.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	q := sqlf.Sprintf(
		getBatchChangeChangesetAssociationFmtstr,
		sqlf.Join(batchChangeChangesetAssociationColumns, ","),
		batchChangeID, changesetID,
	)

	assoc = &btypes.BatchChangeChangesetAssociation{}
	if err := s.query(ctx, q, func(sc dbutil.Scanner) error {
		return scanBatchChangeChangesetAssociation(assoc, sc)
	}); err != nil {
		return nil, err
	}
	return assoc, nil
}

const getBatchChangeChangesetAssociationFmtstr = `
-- source: batch_change_changesets.go:GetBatchChangeChangesetAssociation
SELECT
  %s
FROM
  batch_change_changesets
WHERE
  batch_change_id = %s AND changeset_id = %s
`

type ListBatchChangeChangesetAssociationsOpts struct {
	LimitOpts
	Cursor          int64
	BatchChangeID   int64
	ChangesetID     int64
	OnlyArchived    bool
	IncludeArchived bool
}

func (s *Store) ListBatchChangeChangesetAssociations(
	ctx context.Context,
	opts ListBatchChangeChangesetAssociationsOpts,
) (assocs []*btypes.BatchChangeChangesetAssociation, next int64, err error) {
	ctx, _, endObservation := s.operations.listBatchChangeChangesetAssociations.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	q := listBatchChangeChangesetAssociationsQuery(opts)

	assocs = make([]*btypes.BatchChangeChangesetAssociation, 0, opts.DBLimit())
	if err = s.query(ctx, q, func(sc dbutil.Scanner) error {
		var assoc btypes.BatchChangeChangesetAssociation
		if err := scanBatchChangeChangesetAssociation(&assoc, sc); err != nil {
			return err
		}
		assocs = append(assocs, &assoc)
		return nil
	}); err != nil {
		return
	}

	if opts.Limit != 0 && len(assocs) == opts.DBLimit() {
		next = int64(opts.Limit) + opts.Cursor
		assocs = assocs[:len(assocs)-1]
	}

	return
}

func listBatchChangeChangesetAssociationsQuery(opts ListBatchChangeChangesetAssociationsOpts) *sqlf.Query {
	preds := []*sqlf.Query{sqlf.Sprintf("TRUE")}

	if opts.BatchChangeID != 0 {
		preds = append(preds, sqlf.Sprintf("batch_change_id = %s", opts.BatchChangeID))
	}
	if opts.ChangesetID != 0 {
		preds = append(preds, sqlf.Sprintf("changeset_id = %s", opts.ChangesetID))
	}
	if opts.OnlyArchived {
		preds = append(preds, sqlf.Sprintf("archived = %s", btypes.BatchChangeChangesetArchived))
	}
	if !opts.IncludeArchived {
		preds = append(preds, sqlf.Sprintf("archived IS NULL"))
	}

	var offset string
	if opts.Cursor != 0 {
		offset = fmt.Sprintf(" OFFSET %d", opts.Cursor)
	}

	return sqlf.Sprintf(
		listBatchChangeChangesetAssociationsFmtstr+opts.ToDB()+offset,
		sqlf.Join(batchChangeChangesetAssociationColumns, ","),
		sqlf.Join(preds, " AND "),
	)
}

const listBatchChangeChangesetAssociationsFmtstr = `
-- source: batch_change_changesets.go:ListBatchChangeChangesetAssociations
SELECT
  %s
FROM
  batch_change_changesets
WHERE
  %s
ORDER BY
  batch_change_id ASC, changeset_id ASC
`

// TODO: this might be combined with Create for an Upsert if that's easier.
func (s *Store) UpdateBatchChangeChangesetAssociation(
	ctx context.Context,
	assoc *btypes.BatchChangeChangesetAssociation,
) (err error) {
	ctx, _, endObservation := s.operations.updateBatchChangeChangesetAssociation.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	q := sqlf.Sprintf(
		updateBatchChangeChangesetAssociationFmtstr,
		assoc.Detach,
		nullStringColumn(string(assoc.Archived)),
		assoc.BatchChangeID,
		assoc.ChangesetID,
		sqlf.Join(batchChangeChangesetAssociationColumns, ","),
	)

	return s.query(ctx, q, func(sc dbutil.Scanner) error {
		return scanBatchChangeChangesetAssociation(assoc, sc)
	})
}

const updateBatchChangeChangesetAssociationFmtstr = `
-- source: batch_change_changesets.go:CreateBatchChangeChangesetAssociation
UPDATE
  batch_change_changesets
SET
  detach = %s,
  archived = %s
WHERE
  batch_change_id = %s AND changeset_id = %s
RETURNING
  %s
`

func scanBatchChangeChangesetAssociation(
	assoc *btypes.BatchChangeChangesetAssociation,
	sc dbutil.Scanner,
) error {
	var archived string

	if err := sc.Scan(
		&assoc.BatchChangeID,
		&assoc.ChangesetID,
		&assoc.Detach,
		&dbutil.NullString{S: &archived},
	); err != nil {
		return err
	}

	assoc.Archived = btypes.BatchChangeChangesetArchivedState(archived)
	return nil
}

var batchChangeChangesetAssociationColumns = []*sqlf.Query{
	sqlf.Sprintf("batch_change_changesets.batch_change_id"),
	sqlf.Sprintf("batch_change_changesets.changeset_id"),
	sqlf.Sprintf("batch_change_changesets.detach"),
	sqlf.Sprintf("batch_change_changesets.archived"),
}
