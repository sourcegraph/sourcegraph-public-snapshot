package db

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/types"
)

// SameRepoPager returns a ReferencePager for dumps that belong to the given repository and commit and reference the package with the
// given scheme, name, and version.
func (db *dbImpl) SameRepoPager(ctx context.Context, repositoryID int, commit, scheme, name, version string, limit int) (_ int, _ ReferencePager, err error) {
	tw, err := db.beginTx(ctx)
	if err != nil {
		return 0, nil, err
	}
	defer func() {
		if err != nil {
			err = closeTx(tw.tx, err)
		}
	}()

	visibleIDsQuery := `SELECT id FROM visible_ids`
	visibleIDs, err := scanInts(tw.query(ctx, withBidirectionalLineage(visibleIDsQuery, repositoryID, commit)))
	if err != nil {
		return 0, nil, err
	}

	if len(visibleIDs) == 0 {
		return 0, newEmptyReferencePager(tw.tx), nil
	}

	conds := []*sqlf.Query{
		sqlf.Sprintf("r.scheme = %s", scheme),
		sqlf.Sprintf("r.name = %s", name),
		sqlf.Sprintf("r.version = %s", version),
		sqlf.Sprintf("r.dump_id IN (%s)", sqlf.Join(intsToQueries(visibleIDs), ", ")),
	}

	countQuery := `SELECT COUNT(1) FROM lsif_references r WHERE %s`
	totalCount, err := scanInt(tw.queryRow(ctx, sqlf.Sprintf(countQuery, sqlf.Join(conds, " AND "))))
	if err != nil {
		return 0, nil, err
	}

	pageFromOffset := func(offset int) ([]types.PackageReference, error) {
		query := `
			SELECT d.id, r.scheme, r.name, r.version, r.filter FROM lsif_references r
			LEFT JOIN lsif_dumps d on r.dump_id = d.id
			WHERE %s ORDER BY d.root LIMIT %d OFFSET %d
		`

		return scanPackageReferences(tw.query(ctx, sqlf.Sprintf(query, sqlf.Join(conds, " AND "), limit, offset)))
	}

	return totalCount, newReferencePager(tw.tx, pageFromOffset), nil
}

// PackageReferencePager returns a ReferencePager for dumps that belong to a remote repository (distinct from the given repository id)
// and reference the package with the given scheme, name, and version. All resulting dumps are visible at the tip of their repository's
// default branch.
func (db *dbImpl) PackageReferencePager(ctx context.Context, scheme, name, version string, repositoryID, limit int) (_ int, _ ReferencePager, err error) {
	tw, err := db.beginTx(ctx)
	if err != nil {
		return 0, nil, err
	}
	defer func() {
		if err != nil {
			err = closeTx(tw.tx, err)
		}
	}()

	conds := []*sqlf.Query{
		sqlf.Sprintf("r.scheme = %s", scheme),
		sqlf.Sprintf("r.name = %s", name),
		sqlf.Sprintf("r.version = %s", version),
		sqlf.Sprintf("d.repository_id != %s", repositoryID),
		sqlf.Sprintf("d.visible_at_tip = true"),
	}

	countQuery := `
		SELECT COUNT(1) FROM lsif_references r
		LEFT JOIN lsif_dumps d ON r.dump_id = d.id
		WHERE %s
	`

	totalCount, err := scanInt(tw.queryRow(ctx, sqlf.Sprintf(countQuery, sqlf.Join(conds, " AND "))))
	if err != nil {
		return 0, nil, err
	}

	pageFromOffset := func(offset int) ([]types.PackageReference, error) {
		query := `
			SELECT d.id, r.scheme, r.name, r.version, r.filter FROM lsif_references r
			LEFT JOIN lsif_dumps d ON r.dump_id = d.id
			WHERE %s ORDER BY d.repository_id, d.root LIMIT %d OFFSET %d
		`

		return scanPackageReferences(tw.query(ctx, sqlf.Sprintf(query, sqlf.Join(conds, " AND "), limit, offset)))
	}

	return totalCount, newReferencePager(tw.tx, pageFromOffset), nil
}
