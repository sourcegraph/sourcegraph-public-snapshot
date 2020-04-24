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
	TracingContext    string     `json:"tracingContext"`
	RepositoryID      int        `json:"repositoryId"`
	Indexer           string     `json:"indexer"`
}

// GetDumpByID returns a dump by its identifier and boolean flag indicating its existence.
func (db *dbImpl) GetDumpByID(ctx context.Context, id int) (Dump, bool, error) {
	query := `
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
			d.tracing_context,
			d.repository_id,
			d.indexer
		FROM lsif_dumps d WHERE id = %d
	`

	dump, err := scanDump(db.queryRow(ctx, sqlf.Sprintf(query, id)))
	if err != nil {
		return Dump{}, false, ignoreErrNoRows(err)
	}

	return dump, true, nil
}

// FindClosestDumps returns the set of dumps that can most accurately answer queries for the given repository, commit, and file.
func (db *dbImpl) FindClosestDumps(ctx context.Context, repositoryID int, commit, file string) ([]Dump, error) {
	tw, err := db.beginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = closeTx(tw.tx, err)
	}()

	visibleIDsQuery := `
		SELECT d.dump_id FROM lineage_with_dumps d
		WHERE %s LIKE (d.root || '%%%%') AND d.dump_id IN (SELECT * FROM visible_ids)
		ORDER BY d.n
	`

	ids, err := scanInts(tw.query(ctx, withBidirectionalLineage(visibleIDsQuery, repositoryID, commit, file)))
	if err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		return nil, nil
	}

	query := `
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
			d.tracing_context,
			d.repository_id,
			d.indexer
		FROM lsif_dumps d WHERE id IN (%s)
	`

	dumps, err := scanDumps(tw.query(ctx, sqlf.Sprintf(query, sqlf.Join(intsToQueries(ids), ", "))))
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
	query := `
		DELETE FROM lsif_uploads
		WHERE id IN (
			SELECT id FROM lsif_dumps
			WHERE visible_at_tip = false
			ORDER BY uploaded_at
			LIMIT 1
		) RETURNING id
	`

	id, err := scanInt(db.queryRow(ctx, sqlf.Sprintf(query)))
	if err != nil {
		return 0, false, ignoreErrNoRows(err)
	}

	return id, true, nil
}

// updateDumpsVisibleFromTip recalculates the visible_at_tip flag of all dumps of the given repository.
func (db *dbImpl) updateDumpsVisibleFromTip(ctx context.Context, tw *transactionWrapper, repositoryID int, tipCommit string) (err error) {
	if tw == nil {
		tw, err = db.beginTx(ctx)
		if err != nil {
			return err
		}
		defer func() {
			err = closeTx(tw.tx, err)
		}()
	}

	// Update dump records by:
	//   (1) unsetting the visibility flag of all previously visible dumps, and
	//   (2) setting the visibility flag of all currently visible dumps
	query := `
		UPDATE lsif_dumps d
		SET visible_at_tip = id IN (SELECT * from visible_ids)
		WHERE d.repository_id = %s AND (d.id IN (SELECT * from visible_ids) OR d.visible_at_tip)
	`

	_, err = tw.exec(ctx, withAncestorLineage(query, repositoryID, tipCommit, repositoryID))
	return err
}
