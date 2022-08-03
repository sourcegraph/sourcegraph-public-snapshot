package store

import (
	"context"
	"fmt"

	"github.com/keegancsmith/sqlf"

	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type CountBatchChangeChangesetAssociationsOpts struct {
	BatchChangeID   int64
	ChangesetID     int64
	OnlyArchived    bool
	IncludeArchived bool
}

func (opts CountBatchChangeChangesetAssociationsOpts) preds() []*sqlf.Query {
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

	return preds
}

func (s *Store) CountBatchChangeChangesetAssociations(
	ctx context.Context,
	opts CountBatchChangeChangesetAssociationsOpts,
) (count int64, err error) {
	ctx, _, endObservation := s.operations.countBatchChangeChangesetAssociations.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	q := countBatchChangeChangesetAssociationsQuery(opts)

	err = s.query(ctx, q, func(sc dbutil.Scanner) error {
		return sc.Scan(&count)
	})
	return
}

func countBatchChangeChangesetAssociationsQuery(opts CountBatchChangeChangesetAssociationsOpts) *sqlf.Query {
	return sqlf.Sprintf(
		countBatchChangeChangesetAssociationsFmtstr,
		sqlf.Join(opts.preds(), " AND "),
	)
}

const countBatchChangeChangesetAssociationsFmtstr = `
-- source: batch_change_changesets.go:CountBatchChangeChangesetAssociations
SELECT
  COUNT(*)
FROM
  batch_change_changesets
WHERE
  %s
`

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

func (s *Store) GetAssociatedChangeset(ctx context.Context, cs *btypes.Changeset) (ac *btypes.AssociatedChangeset, err error) {
	// TODO: Not paginated because the previous implementation wasn't, but we
	// should figure out if that's still a reasonable assumption here.
	ctx, _, endObservation := s.operations.getAssociatedChangeset.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	q := sqlf.Sprintf(
		getAssociatedChangesetFmtstr,
		sqlf.Join(batchChangeChangesetAssociationColumns, ","),
		cs.ID,
	)

	ac = btypes.NewAssociatedChangeset(cs)
	if err := s.query(ctx, q, func(sc dbutil.Scanner) error {
		assoc := btypes.BatchChangeChangesetAssociation{}
		if err := scanBatchChangeChangesetAssociation(&assoc, sc); err != nil {
			return err
		}

		ac.AddAssociation(&assoc)
		return nil
	}); err != nil {
		return nil, err
	}

	return ac, nil
}

const getAssociatedChangesetFmtstr = `
-- source: batch_change_changesets.go:GetAssociatedChangeset
SELECT
  %s
FROM
  batch_change_changesets
WHERE
  changeset_id = %s
`

func (s *Store) UpsertAssociatedChangeset(ctx context.Context, ac *btypes.AssociatedChangeset) (err error) {
	// TODO: delegate to either INSERT or UPDATE variant depending on ac.Changeset.ID.
}

// WITH
//   inserted_a AS (
//     INSERT INTO
//       a (name)
//     VALUES
//       ('c')
//     RETURNING
//       id
//   ),
//   data (b, archived) AS (
//     VALUES (1, false), (2, true)
//   )
// INSERT INTO
//   j (a, b, archived)
// SELECT
//   inserted_a.id, data.b, data.archived
// FROM
//   inserted_a CROSS JOIN data;

// WITH
//   a (id) AS (
//     SELECT 1
//   ),
//   data (b, archived) AS (
//     VALUES (1, true), (2, false)
//   ),
//   deleted AS (
//     DELETE FROM
//       j
//     USING
//       a
//     WHERE
//       a = a.id
//     RETURNING
//       0
//   )
// INSERT INTO
//   j (a, b, archived)
// SELECT
//   a.id, data.b, data.archived
// FROM
//   a CROSS JOIN data LEFT OUTER JOIN deleted ON TRUE;

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
	if assoc.BatchChangeID == 0 {
		return nil, ErrNoResults
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
	preds := CountBatchChangeChangesetAssociationsOpts{
		BatchChangeID:   opts.BatchChangeID,
		ChangesetID:     opts.ChangesetID,
		OnlyArchived:    opts.OnlyArchived,
		IncludeArchived: opts.IncludeArchived,
	}.preds()

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
