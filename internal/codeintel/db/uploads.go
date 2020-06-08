package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

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
	RepositoryID      int        `json:"repositoryId"`
	Indexer           string     `json:"indexer"`
	NumParts          int        `json:"numParts"`
	UploadedParts     []int      `json:"uploadedParts"`
	Rank              *int       `json:"placeInQueue"`
}

// scanUploads scans a slice of uploads from the return value of `*dbImpl.query`.
func scanUploads(rows *sql.Rows, queryErr error) (_ []Upload, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = closeRows(rows, err) }()

	var uploads []Upload
	for rows.Next() {
		var upload Upload
		var rawUploadedParts []sql.NullInt32
		if err := rows.Scan(
			&upload.ID,
			&upload.Commit,
			&upload.Root,
			&upload.VisibleAtTip,
			&upload.UploadedAt,
			&upload.State,
			&upload.FailureSummary,
			&upload.FailureStacktrace,
			&upload.StartedAt,
			&upload.FinishedAt,
			&upload.RepositoryID,
			&upload.Indexer,
			&upload.NumParts,
			pq.Array(&rawUploadedParts),
			&upload.Rank,
		); err != nil {
			return nil, err
		}

		var uploadedParts = []int{}
		for _, uploadedPart := range rawUploadedParts {
			uploadedParts = append(uploadedParts, int(uploadedPart.Int32))
		}
		upload.UploadedParts = uploadedParts

		uploads = append(uploads, upload)
	}

	return uploads, nil
}

// scanFirstUpload scans a slice of uploads from the return value of `*dbImpl.query` and returns the first.
func scanFirstUpload(rows *sql.Rows, err error) (Upload, bool, error) {
	uploads, err := scanUploads(rows, err)
	if err != nil || len(uploads) == 0 {
		return Upload{}, false, err
	}
	return uploads[0], true, nil
}

// scanFirstUploadInterface scans a slice of uploads from the return value of `*dbImpl.query` and returns the first.
func scanFirstUploadInterface(rows *sql.Rows, err error) (interface{}, bool, error) {
	return scanFirstUpload(rows, err)
}

// scanStates scans pairs of id/states from the return value of `*dbImpl.query`.
func scanStates(rows *sql.Rows, queryErr error) (_ map[int]string, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = closeRows(rows, err) }()

	states := map[int]string{}
	for rows.Next() {
		var id int
		var state string
		if err := rows.Scan(&id, &state); err != nil {
			return nil, err
		}

		states[id] = state
	}

	return states, nil
}

// scanVisibility scans pairs of id/visibleAtTip from the return value of `*dbImpl.query`.
func scanVisibilities(rows *sql.Rows, queryErr error) (_ map[int]bool, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = closeRows(rows, err) }()

	visibilities := map[int]bool{}
	for rows.Next() {
		var id int
		var visibleAtTip bool
		if err := rows.Scan(&id, &visibleAtTip); err != nil {
			return nil, err
		}

		visibilities[id] = visibleAtTip
	}

	return visibilities, nil
}

// GetUploadByID returns an upload by its identifier and boolean flag indicating its existence.
func (db *dbImpl) GetUploadByID(ctx context.Context, id int) (Upload, bool, error) {
	return scanFirstUpload(db.query(ctx, sqlf.Sprintf(`
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
			u.repository_id,
			u.indexer,
			u.num_parts,
			u.uploaded_parts,
			s.rank
		FROM lsif_uploads u
		LEFT JOIN (
			SELECT r.id, RANK() OVER (ORDER BY r.uploaded_at) as rank
			FROM lsif_uploads r
			WHERE r.state = 'queued'
		) s
		ON u.id = s.id
		WHERE u.id = %s
	`, id)))
}

type GetUploadsOptions struct {
	RepositoryID int
	State        string
	Term         string
	VisibleAtTip bool
	Limit        int
	Offset       int
}

// GetUploads returns a list of uploads and the total count of records matching the given conditions.
func (db *dbImpl) GetUploads(ctx context.Context, opts GetUploadsOptions) (_ []Upload, _ int, err error) {
	tx, started, err := db.transact(ctx)
	if err != nil {
		return nil, 0, err
	}
	if started {
		defer func() { err = tx.Done(err) }()
	}

	var conds []*sqlf.Query

	if opts.RepositoryID != 0 {
		conds = append(conds, sqlf.Sprintf("u.repository_id = %s", opts.RepositoryID))
	}
	if opts.Term != "" {
		conds = append(conds, makeSearchCondition(opts.Term))
	}
	if opts.State != "" {
		conds = append(conds, sqlf.Sprintf("u.state = %s", opts.State))
	}
	if opts.VisibleAtTip {
		conds = append(conds, sqlf.Sprintf("u.visible_at_tip = true"))
	}

	if len(conds) == 0 {
		conds = append(conds, sqlf.Sprintf("TRUE"))
	}

	count, _, err := scanFirstInt(tx.query(
		ctx,
		sqlf.Sprintf(`SELECT COUNT(*) FROM lsif_uploads u WHERE %s`, sqlf.Join(conds, " AND ")),
	))
	if err != nil {
		return nil, 0, err
	}

	uploads, err := scanUploads(tx.query(
		ctx,
		sqlf.Sprintf(`
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
				u.repository_id,
				u.indexer,
				u.num_parts,
				u.uploaded_parts,
				s.rank
			FROM lsif_uploads u
			LEFT JOIN (
				SELECT r.id, RANK() OVER (ORDER BY r.uploaded_at) as rank
				FROM lsif_uploads r
				WHERE r.state = 'queued'
			) s
			ON u.id = s.id
			WHERE %s ORDER BY uploaded_at DESC LIMIT %d OFFSET %d
		`, sqlf.Join(conds, " AND "), opts.Limit, opts.Offset),
	))
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

// QueueSize returns the number of uploads in the queued state.
func (db *dbImpl) QueueSize(ctx context.Context) (int, error) {
	count, _, err := scanFirstInt(db.query(ctx, sqlf.Sprintf(`SELECT COUNT(*) FROM lsif_uploads WHERE state = 'queued'`)))
	return count, err
}

// InsertUpload inserts a new upload and returns its identifier.
func (db *dbImpl) InsertUpload(ctx context.Context, upload Upload) (int, error) {
	if upload.UploadedParts == nil {
		upload.UploadedParts = []int{}
	}

	id, _, err := scanFirstInt(db.query(
		ctx,
		sqlf.Sprintf(`
			INSERT INTO lsif_uploads (
				commit,
				root,
				repository_id,
				indexer,
				state,
				num_parts,
				uploaded_parts
			) VALUES (%s, %s, %s, %s, %s, %s, %s)
			RETURNING id
		`,
			upload.Commit,
			upload.Root,
			upload.RepositoryID,
			upload.Indexer,
			upload.State,
			upload.NumParts,
			pq.Array(upload.UploadedParts),
		),
	))

	return id, err
}

// AddUploadPart adds the part index to the given upload's uploaded parts array. This method is idempotent
// (the resulting array is deduplicated on update).
func (db *dbImpl) AddUploadPart(ctx context.Context, uploadID, partIndex int) error {
	return db.queryForEffect(ctx, sqlf.Sprintf(`
		UPDATE lsif_uploads
		SET uploaded_parts = array(SELECT DISTINCT * FROM unnest(array_append(uploaded_parts, %s)))
		WHERE id = %s
	`, partIndex, uploadID))
}

// MarkQueued updates the state of the upload to queued.
func (db *dbImpl) MarkQueued(ctx context.Context, uploadID int) error {
	return db.queryForEffect(ctx, sqlf.Sprintf(`UPDATE lsif_uploads SET state = 'queued' WHERE id = %s`, uploadID))
}

// MarkComplete updates the state of the upload to complete.
func (db *dbImpl) MarkComplete(ctx context.Context, id int) (err error) {
	return db.queryForEffect(ctx, sqlf.Sprintf(`
		UPDATE lsif_uploads
		SET state = 'completed', finished_at = clock_timestamp()
		WHERE id = %s
	`, id))
}

// MarkErrored updates the state of the upload to errored and updates the failure summary data.
func (db *dbImpl) MarkErrored(ctx context.Context, id int, failureSummary, failureStacktrace string) (err error) {
	return db.queryForEffect(ctx, sqlf.Sprintf(`
		UPDATE lsif_uploads
		SET state = 'errored', finished_at = clock_timestamp(), failure_summary = %s, failure_stacktrace = %s
		WHERE id = %s
	`, failureSummary, failureStacktrace, id))
}

var uploadColumnsWithNullRank = []*sqlf.Query{
	sqlf.Sprintf("id"),
	sqlf.Sprintf("commit"),
	sqlf.Sprintf("root"),
	sqlf.Sprintf("visible_at_tip"),
	sqlf.Sprintf("uploaded_at"),
	sqlf.Sprintf("state"),
	sqlf.Sprintf("failure_summary"),
	sqlf.Sprintf("failure_stacktrace"),
	sqlf.Sprintf("started_at"),
	sqlf.Sprintf("finished_at"),
	sqlf.Sprintf("repository_id"),
	sqlf.Sprintf("indexer"),
	sqlf.Sprintf("num_parts"),
	sqlf.Sprintf("uploaded_parts"),
	sqlf.Sprintf("NULL"),
}

// Dequeue selects the oldest queued upload and locks it with a transaction. If there is such an upload, the
// upload is returned along with a DB instance which wraps the transaction. This transaction must be closed.
// If there is no such unlocked upload, a zero-value upload and nil DB will be returned along with a false
// valued flag. This method must not be called from within a transaction.
func (db *dbImpl) Dequeue(ctx context.Context) (Upload, DB, bool, error) {
	upload, tx, ok, err := db.dequeueRecord(ctx, "lsif_uploads", uploadColumnsWithNullRank, sqlf.Sprintf("uploaded_at"), scanFirstUploadInterface)
	if err != nil || !ok {
		return Upload{}, tx, ok, err
	}

	return upload.(Upload), tx, true, nil
}

// GetStates returns the states for the uploads with the given identifiers.
func (db *dbImpl) GetStates(ctx context.Context, ids []int) (map[int]string, error) {
	return scanStates(db.query(ctx, sqlf.Sprintf(`
		SELECT id, state FROM lsif_uploads
		WHERE id IN (%s)
	`, sqlf.Join(intsToQueries(ids), ", "))))
}

// DeleteUploadByID deletes an upload by its identifier. If the upload was visible at the tip of its repository's default branch,
// the visibility of all uploads for that repository are recalculated. The getTipCommit function is expected to return the newest
// commit on the default branch when invoked.
func (db *dbImpl) DeleteUploadByID(ctx context.Context, id int, getTipCommit GetTipCommitFn) (_ bool, err error) {
	tx, started, err := db.transact(ctx)
	if err != nil {
		return false, err
	}
	if started {
		defer func() { err = tx.Done(err) }()
	}

	visibilities, err := scanVisibilities(tx.query(
		ctx,
		sqlf.Sprintf(`
			DELETE FROM lsif_uploads
			WHERE id = %s
			RETURNING repository_id, visible_at_tip
		`, id),
	))
	if err != nil {
		return false, err
	}

	for repositoryID, visibleAtTip := range visibilities {
		if visibleAtTip {
			tipCommit, err := getTipCommit(repositoryID)
			if err != nil {
				return false, err
			}

			if err := tx.UpdateDumpsVisibleFromTip(ctx, repositoryID, tipCommit); err != nil {
				return false, errors.Wrap(err, "db.UpdateDumpsVisibleFromTip")
			}
		}

		return true, nil
	}

	return false, nil
}

// StalledUploadMaxAge is the maximum allowable duration between updating the state of an
// upload as "processing" and locking the upload row during processing. An unlocked row that
// is marked as processing likely indicates that the worker that dequeued the upload has died.
// There should be a nearly-zero delay between these states during normal operation.
const StalledUploadMaxAge = time.Second * 5

// ResetStalled moves all unlocked uploads processing for more than `StalledUploadMaxAge` back to the queued state.
// This method returns a list of updated upload identifiers.
func (db *dbImpl) ResetStalled(ctx context.Context, now time.Time) ([]int, error) {
	ids, err := scanInts(db.query(
		ctx,
		sqlf.Sprintf(`
			UPDATE lsif_uploads u SET state = 'queued', started_at = null WHERE id = ANY(
				SELECT id FROM lsif_uploads
				WHERE state = 'processing' AND %s - started_at > (%s * interval '1 second')
				FOR UPDATE SKIP LOCKED
			)
			RETURNING u.id
		`, now.UTC(), StalledUploadMaxAge/time.Second),
	))
	if err != nil {
		return nil, err
	}

	return ids, nil
}
