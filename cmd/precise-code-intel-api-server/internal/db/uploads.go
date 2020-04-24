package db

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"
)

// StalledUploadMaxAge is the maximum allowable duration between updating the state of an
// upload as "processing" and locking the upload row during processing. An unlocked row that
// is marked as processing likely indicates that the worker that dequeued the upload has died.
// There should be a nearly-zero delay between these states during normal operation.
const StalledUploadMaxAge = time.Second * 5

// Upload is a subset of the lsif_uploads table and stores both processed and unprocessed
// records.
type Upload struct {
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
	Rank              *int       `json:"placeInQueue"`
}

// GetUploadByID returns an upload by its identifier and boolean flag indicating its existence.
func (db *dbImpl) GetUploadByID(ctx context.Context, id int) (Upload, bool, error) {
	query := `
		SELECT
			u.id,
			u.commit,
			u.root,
			u.visible_at_tip,
			u.uploaded_at,
			u.state,
			u.failure_summary,
			u.failure_stacktrace,
			u.started_at,
			u.finished_at,
			u.tracing_context,
			u.repository_id,
			u.indexer,
			s.rank
		FROM lsif_uploads u
		LEFT JOIN (
			SELECT r.id, RANK() OVER (ORDER BY r.uploaded_at) as rank
			FROM lsif_uploads r
			WHERE r.state = 'queued'
		) s
		ON u.id = s.id
		WHERE u.id = %s
	`

	upload, err := scanUpload(db.queryRow(ctx, sqlf.Sprintf(query, id)))
	if err != nil {
		return Upload{}, false, ignoreErrNoRows(err)
	}

	return upload, true, nil
}

// GetUploadsByRepo returns a list of uploads for a particular repo and the total count of records matching the given conditions.
func (db *dbImpl) GetUploadsByRepo(ctx context.Context, repositoryID int, state, term string, visibleAtTip bool, limit, offset int) (_ []Upload, _ int, err error) {
	tw, err := db.beginTx(ctx)
	if err != nil {
		return nil, 0, err
	}
	defer func() {
		err = closeTx(tw.tx, err)
	}()

	var conds []*sqlf.Query
	conds = append(conds, sqlf.Sprintf("u.repository_id = %s", repositoryID))
	if state != "" {
		conds = append(conds, sqlf.Sprintf("u.state = %s", state))
	}
	if term != "" {
		conds = append(conds, makeSearchCondition(term))
	}
	if visibleAtTip {
		conds = append(conds, sqlf.Sprintf("u.visible_at_tip = true"))
	}

	countQuery := `SELECT COUNT(1) FROM lsif_uploads u WHERE %s`
	count, err := scanInt(tw.queryRow(ctx, sqlf.Sprintf(countQuery, sqlf.Join(conds, " AND "))))
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT
			u.id,
			u.commit,
			u.root,
			u.visible_at_tip,
			u.uploaded_at,
			u.state,
			u.failure_summary,
			u.failure_stacktrace,
			u.started_at,
			u.finished_at,
			u.tracing_context,
			u.repository_id,
			u.indexer,
			s.rank
		FROM lsif_uploads u
		LEFT JOIN (
			SELECT r.id, RANK() OVER (ORDER BY r.uploaded_at) as rank
			FROM lsif_uploads r
			WHERE r.state = 'queued'
		) s
		ON u.id = s.id
		WHERE %s ORDER BY uploaded_at DESC LIMIT %d OFFSET %d
	`

	uploads, err := scanUploads(tw.query(ctx, sqlf.Sprintf(query, sqlf.Join(conds, " AND "), limit, offset)))
	if err != nil {
		return nil, 0, err
	}

	return uploads, count, nil
}

// makeSearchCondition returns a disjunction of LIKE clauses against all searchable columns of an upload.
func makeSearchCondition(term string) *sqlf.Query {
	searchableColumns := []string{
		"commit",
		"root",
		"indexer",
		"failure_summary",
		"failure_stacktrace",
	}

	var termConds []*sqlf.Query
	for _, column := range searchableColumns {
		termConds = append(termConds, sqlf.Sprintf("u."+column+" LIKE %s", "%"+term+"%"))
	}

	return sqlf.Sprintf("(%s)", sqlf.Join(termConds, " OR "))
}

// Enqueue inserts a new upload with a "queued" state, returning its identifier and a TxCloser that must be closed to commit the transaction.
func (db *dbImpl) Enqueue(ctx context.Context, commit, root, tracingContext string, repositoryID int, indexerName string) (_ int, _ TxCloser, err error) {
	tw, err := db.beginTx(ctx)
	if err != nil {
		return 0, nil, err
	}
	defer func() {
		if err != nil {
			err = closeTx(tw.tx, err)
		}
	}()

	query := `
		INSERT INTO lsif_uploads (commit, root, tracing_context, repository_id, indexer)
		VALUES (%s, %s, %s, %s, %s)
		RETURNING id
	`

	id, err := scanInt(tw.queryRow(ctx, sqlf.Sprintf(query, commit, root, tracingContext, repositoryID, indexerName)))
	if err != nil {
		return 0, nil, err
	}

	return id, &txCloser{tw.tx}, nil
}

// GetStates returns the states for the uploads with the given identifiers.
func (db *dbImpl) GetStates(ctx context.Context, ids []int) (map[int]string, error) {
	query := `SELECT id, state FROM lsif_uploads WHERE id IN (%s)`
	return scanStates(db.query(ctx, sqlf.Sprintf(query, sqlf.Join(intsToQueries(ids), ", "))))
}

// DeleteUploadByID deletes an upload by its identifier. If the upload was visible at the tip of its repository's default branch,
// the visibility of all uploads for that repository are recalculated. The given function is expected to return the newest commit
// on the default branch when invoked.
func (db *dbImpl) DeleteUploadByID(ctx context.Context, id int, getTipCommit func(repositoryID int) (string, error)) (_ bool, err error) {
	tw, err := db.beginTx(ctx)
	if err != nil {
		return false, err
	}
	defer func() {
		err = closeTx(tw.tx, err)
	}()

	query := `
		DELETE FROM lsif_uploads
		WHERE id = %s
		RETURNING repository_id, visible_at_tip
	`

	repositoryID, visibleAtTip, err := scanVisibility(tw.queryRow(ctx, sqlf.Sprintf(query, id)))
	if err != nil {
		return false, ignoreErrNoRows(err)
	}

	if !visibleAtTip {
		return true, nil
	}

	tipCommit, err := getTipCommit(repositoryID)
	if err != nil {
		return false, err
	}

	if err := db.updateDumpsVisibleFromTip(ctx, tw, repositoryID, tipCommit); err != nil {
		return false, err
	}

	return true, nil
}

// ResetStalled moves all unlocked uploads processing for more than `StalledUploadMaxAge` back to the queued state.
// This method returns a list of updated upload identifiers.
func (db *dbImpl) ResetStalled(ctx context.Context, now time.Time) ([]int, error) {
	query := `
		UPDATE lsif_uploads u SET state = 'queued', started_at = null WHERE id = ANY(
			SELECT id FROM lsif_uploads
			WHERE state = 'processing' AND %s - started_at > (%s * interval '1 second')
			FOR UPDATE SKIP LOCKED
		)
		RETURNING u.id
	`

	ids, err := scanInts(db.query(ctx, sqlf.Sprintf(query, now.UTC(), StalledUploadMaxAge/time.Second)))
	if err != nil {
		return nil, err
	}

	return ids, nil
}
