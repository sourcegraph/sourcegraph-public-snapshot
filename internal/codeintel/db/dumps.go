package db

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"
)

// Dump is a subset of the lsif_uploads table (queried via the lsif_dumps view) and stores
// only processed records.
type Dump struct {
	ID                int        `json:"id"`
	Commit            string     `json:"commit"`
	Root              string     `json:"root"`
	VisibleAtTip      bool       `json:"visibleAtTip"`
	UploadedAt        time.Time  `json:"uploadedAt"`
	State             string     `json:"state"`
	FailureSummary    *string    `json:"failureSummary"`
	FailureStacktrace *string    `json:"failureStacktrace"`
	StartedAt         *time.Time `json:"startedAt"`
	FinishedAt        *time.Time `json:"finishedAt"`
	RepositoryID      int        `json:"repositoryId"`
	Indexer           string     `json:"indexer"`
}

// GetDumpByID returns a dump by its identifier and boolean flag indicating its existence.
func (db *dbImpl) GetDumpByID(ctx context.Context, id int) (Dump, bool, error) {
	return scanFirstDump(db.query(ctx, sqlf.Sprintf(`
		SELECT
			d.id,
			d.commit,
			d.root,
			d.visible_at_tip,
			d.uploaded_at,
			d.state,
			d.failure_summary,
			d.failure_stacktrace,
			d.started_at,
			d.finished_at,
			d.repository_id,
			d.indexer
		FROM lsif_dumps d WHERE id = %d
	`, id)))
}

// FindClosestDumps returns the set of dumps that can most accurately answer queries for the given repository, commit, and file.
func (db *dbImpl) FindClosestDumps(ctx context.Context, repositoryID int, commit, file string) (_ []Dump, err error) {
	tx, started, err := db.transact(ctx)
	if err != nil {
		return nil, err
	}
	if started {
		defer func() { err = tx.Done(err) }()
	}

	ids, err := scanInts(tx.query(
		ctx,
		withBidirectionalLineage(`
			SELECT d.dump_id FROM lineage_with_dumps d
			WHERE %s LIKE (d.root || '%%%%') AND d.dump_id IN (SELECT * FROM visible_ids)
			ORDER BY d.n
		`, repositoryID, commit, file),
	))
	if err != nil || len(ids) == 0 {
		return nil, err
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
				d.failure_summary,
				d.failure_stacktrace,
				d.started_at,
				d.finished_at,
				d.repository_id,
				d.indexer
			FROM lsif_dumps d WHERE id IN (%s)
		`, sqlf.Join(intsToQueries(ids), ", ")),
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
func (db *dbImpl) DeleteOldestDump(ctx context.Context) (int, bool, error) {
	return scanFirstInt(db.query(ctx, sqlf.Sprintf(`
		DELETE FROM lsif_uploads
		WHERE id IN (
			SELECT id FROM lsif_dumps
			WHERE visible_at_tip = false
			ORDER BY uploaded_at
			LIMIT 1
		) RETURNING id
	`)))
}

// UpdateDumpsVisibleFromTip recalculates the visible_at_tip flag of all dumps of the given repository.
func (db *dbImpl) UpdateDumpsVisibleFromTip(ctx context.Context, repositoryID int, tipCommit string) (err error) {
	return db.exec(ctx, withAncestorLineage(`
		UPDATE lsif_dumps d
		SET visible_at_tip = id IN (SELECT * from visible_ids)
		WHERE d.repository_id = %s AND (d.id IN (SELECT * from visible_ids) OR d.visible_at_tip)
	`, repositoryID, tipCommit, repositoryID))
}

// DeleteOverlapapingDumps deletes all completed uploads for the given repository with the same
// commit, root, and indexer. This is necessary to perform during conversions before changing
// the state of a processing upload to completed as there is a unique index on these four columns.
func (db *dbImpl) DeleteOverlappingDumps(ctx context.Context, repositoryID int, commit, root, indexer string) (err error) {
	return db.exec(ctx, sqlf.Sprintf(`
		DELETE from lsif_uploads
		WHERE repository_id = %d AND commit = %s AND root = %s AND indexer = %s AND state = 'completed'
	`, repositoryID, commit, root, indexer))
}
