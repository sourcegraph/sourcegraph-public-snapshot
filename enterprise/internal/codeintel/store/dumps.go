package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/keegancsmith/sqlf"
)

// Dump is a subset of the lsif_uploads table (queried via the lsif_dumps_with_repository_name view)
// and stores only processed records.
type Dump struct {
	ID             int        `json:"id"`
	Commit         string     `json:"commit"`
	Root           string     `json:"root"`
	VisibleAtTip   bool       `json:"visibleAtTip"`
	UploadedAt     time.Time  `json:"uploadedAt"`
	State          string     `json:"state"`
	FailureMessage *string    `json:"failureMessage"`
	StartedAt      *time.Time `json:"startedAt"`
	FinishedAt     *time.Time `json:"finishedAt"`
	ProcessAfter   *time.Time `json:"processAfter"`
	NumResets      int        `json:"numResets"`
	RepositoryID   int        `json:"repositoryId"`
	RepositoryName string     `json:"repositoryName"`
	Indexer        string     `json:"indexer"`
}

// scanDumps scans a slice of dumps from the return value of `*store.query`.
func scanDumps(rows *sql.Rows, queryErr error) (_ []Dump, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = closeRows(rows, err) }()

	var dumps []Dump
	for rows.Next() {
		var dump Dump
		if err := rows.Scan(
			&dump.ID,
			&dump.Commit,
			&dump.Root,
			&dump.VisibleAtTip,
			&dump.UploadedAt,
			&dump.State,
			&dump.FailureMessage,
			&dump.StartedAt,
			&dump.FinishedAt,
			&dump.ProcessAfter,
			&dump.NumResets,
			&dump.RepositoryID,
			&dump.RepositoryName,
			&dump.Indexer,
		); err != nil {
			return nil, err
		}

		dumps = append(dumps, dump)
	}

	return dumps, nil
}

// scanFirstDump scans a slice of dumps from the return value of `*store.query` and returns the first.
func scanFirstDump(rows *sql.Rows, err error) (Dump, bool, error) {
	dumps, err := scanDumps(rows, err)
	if err != nil || len(dumps) == 0 {
		return Dump{}, false, err
	}
	return dumps[0], true, nil
}

// GetDumpByID returns a dump by its identifier and boolean flag indicating its existence.
func (s *store) GetDumpByID(ctx context.Context, id int) (Dump, bool, error) {
	return scanFirstDump(s.query(ctx, sqlf.Sprintf(`
		SELECT
			d.id,
			d.commit,
			d.root,
			d.visible_at_tip,
			d.uploaded_at,
			d.state,
			d.failure_message,
			d.started_at,
			d.finished_at,
			d.process_after,
			d.num_resets,
			d.repository_id,
			d.repository_name,
			d.indexer
		FROM lsif_dumps_with_repository_name d WHERE d.id = %s
	`, id)))
}

// FindClosestDumps returns the set of dumps that can most accurately answer queries for the given repository, commit, path, and
// optional indexer. If rootMustEnclosePath is true, then only dumps with a root which is a prefix of path are returned. Otherwise,
// any dump with a root intersecting the given path is returned.
func (s *store) FindClosestDumps(ctx context.Context, repositoryID int, commit, path string, rootMustEnclosePath bool, indexer string) (_ []Dump, err error) {
	tx, err := s.transact(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()

	var cond *sqlf.Query
	if rootMustEnclosePath {
		// Ensure that the root is a prefix of the path
		cond = sqlf.Sprintf(`%s LIKE (d.root || '%%%%')`, path)
	} else {
		// Ensure that the root is a prefix of the path or vice versa
		cond = sqlf.Sprintf(`%s LIKE (d.root || '%%%%') OR d.root LIKE (%s || '%%%%')`, path, path)
	}

	ids, err := scanInts(tx.query(
		ctx,
		withBidirectionalLineage(`
			SELECT d.dump_id FROM lineage_with_dumps d
			WHERE %s AND d.dump_id IN (SELECT * FROM visible_ids)
			ORDER BY d.n
		`, repositoryID, commit, cond),
	))
	if err != nil || len(ids) == 0 {
		return nil, err
	}

	var conds []*sqlf.Query
	conds = append(conds, sqlf.Sprintf("d.id IN (%s)", sqlf.Join(intsToQueries(ids), ", ")))
	if indexer != "" {
		conds = append(conds, sqlf.Sprintf("indexer = %s", indexer))
	}

	dumps, err := scanDumps(tx.query(
		ctx,
		sqlf.Sprintf(`
			SELECT
				d.id,
				d.commit,
				d.root,
				d.visible_at_tip,
				d.uploaded_at,
				d.state,
				d.failure_message,
				d.started_at,
				d.finished_at,
				d.process_after,
				d.num_resets,
				d.repository_id,
				d.repository_name,
				d.indexer
			FROM lsif_dumps_with_repository_name d WHERE %s
		`, sqlf.Join(conds, " AND ")),
	))
	if err != nil {
		return nil, err
	}

	return deduplicateDumps(dumps), nil
}

// deduplicateDumps returns a copy of the given slice of dumps with duplicate identifiers removed.
// The first dump with a unique identifier is retained.
func deduplicateDumps(allDumps []Dump) (dumps []Dump) {
	dumpIDs := map[int]struct{}{}
	for _, dump := range allDumps {
		if _, ok := dumpIDs[dump.ID]; ok {
			continue
		}

		dumpIDs[dump.ID] = struct{}{}
		dumps = append(dumps, dump)
	}

	return dumps
}

// DeleteOldestDump deletes the oldest dump that is not currently visible at the tip of its repository's default branch.
// This method returns the deleted dump's identifier and a flag indicating its (previous) existence.
func (s *store) DeleteOldestDump(ctx context.Context) (int, bool, error) {
	return scanFirstInt(s.query(ctx, sqlf.Sprintf(`
		DELETE FROM lsif_uploads
		WHERE id IN (
			SELECT id FROM lsif_dumps_with_repository_name
			WHERE visible_at_tip = false
			ORDER BY uploaded_at
			LIMIT 1
		) RETURNING id
	`)))
}

// UpdateDumpsVisibleFromTip recalculates the visible_at_tip flag of all dumps of the given repository.
func (s *store) UpdateDumpsVisibleFromTip(ctx context.Context, repositoryID int, tipCommit string) (err error) {
	return s.queryForEffect(ctx, withAncestorLineage(`
		UPDATE lsif_uploads d
		SET visible_at_tip = id IN (SELECT * from visible_ids)
		WHERE d.repository_id = %s AND (d.id IN (SELECT * from visible_ids) OR d.visible_at_tip)
	`, repositoryID, tipCommit, repositoryID))
}

// DeleteOverlapapingDumps deletes all completed uploads for the given repository with the same
// commit, root, and indexer. This is necessary to perform during conversions before changing
// the state of a processing upload to completed as there is a unique index on these four columns.
func (s *store) DeleteOverlappingDumps(ctx context.Context, repositoryID int, commit, root, indexer string) (err error) {
	return s.queryForEffect(ctx, sqlf.Sprintf(`
		DELETE from lsif_uploads
		WHERE repository_id = %s AND commit = %s AND root = %s AND indexer = %s AND state = 'completed'
	`, repositoryID, commit, root, indexer))
}
