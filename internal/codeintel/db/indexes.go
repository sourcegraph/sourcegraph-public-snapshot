package db

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"
)

// Index is a subset of the lsif_indexes table and stores both processed and unprocessed
// records.
type Index struct {
	ID                int        `json:"id"`
	Commit            string     `json:"commit"`
	QueuedAt          time.Time  `json:"queuedAt"`
	State             string     `json:"state"`
	FailureSummary    *string    `json:"failureSummary"`
	FailureStacktrace *string    `json:"failureStacktrace"`
	StartedAt         *time.Time `json:"startedAt"`
	FinishedAt        *time.Time `json:"finishedAt"`
	RepositoryID      int        `json:"repositoryId"`
	Rank              *int       `json:"placeInQueue"`
}

// GetIndexByID returns an index by its identifier and boolean flag indicating its existence.
func (db *dbImpl) GetIndexByID(ctx context.Context, id int) (Index, bool, error) {
	return scanFirstIndex(db.query(ctx, sqlf.Sprintf(`
		SELECT
			u.id,
			u.commit,
			u.queued_at,
			u.state,
			u.failure_summary,
			u.failure_stacktrace,
			u.started_at,
			u.finished_at,
			u.repository_id,
			s.rank
		FROM lsif_indexes u
		LEFT JOIN (
			SELECT r.id, RANK() OVER (ORDER BY r.queued_at) as rank
			FROM lsif_indexes r
			WHERE r.state = 'queued'
		) s
		ON u.id = s.id
		WHERE u.id = %s
	`, id)))
}

// IndexQueueSize returns the number of indexes in the queued state.
func (db *dbImpl) IndexQueueSize(ctx context.Context) (int, error) {
	count, _, err := scanFirstInt(db.query(
		ctx,
		sqlf.Sprintf(`SELECT COUNT(*) FROM lsif_indexes WHERE state = 'queued'`),
	))

	return count, err
}

// IsQueued returns true if there is an index or an upload for the repository and commit.
func (db *dbImpl) IsQueued(ctx context.Context, repositoryID int, commit string) (bool, error) {
	count, _, err := scanFirstInt(db.query(ctx, sqlf.Sprintf(`
		SELECT COUNT(*) WHERE EXISTS (
			SELECT id FROM lsif_uploads WHERE repository_id = %s AND commit = %s
			UNION
			SELECT id FROM lsif_indexes WHERE repository_id = %s AND commit = %s
		)
	`, repositoryID, commit, repositoryID, commit)))

	return count > 0, err
}

// InsertIndex inserts a new index and returns its identifier.
func (db *dbImpl) InsertIndex(ctx context.Context, index Index) (int, error) {
	id, _, err := scanFirstInt(db.query(
		ctx,
		sqlf.Sprintf(`
			INSERT INTO lsif_indexes (
				commit,
				repository_id,
				state
			) VALUES (%s, %s, %s)
			RETURNING id
		`, index.Commit, index.RepositoryID, index.State),
	))

	return id, err
}

// MarkIndexComplete updates the state of the index to complete.
func (db *dbImpl) MarkIndexComplete(ctx context.Context, id int) (err error) {
	return db.exec(ctx, sqlf.Sprintf(`
		UPDATE lsif_indexes
		SET state = 'completed', finished_at = clock_timestamp()
		WHERE id = %s
	`, id))
}

// MarkIndexErrored updates the state of the index to errored and updates the failure summary data.
func (db *dbImpl) MarkIndexErrored(ctx context.Context, id int, failureSummary, failureStacktrace string) (err error) {
	return db.exec(ctx, sqlf.Sprintf(`
		UPDATE lsif_indexes
		SET state = 'errored', finished_at = clock_timestamp(), failure_summary = %s, failure_stacktrace = %s
		WHERE id = %s
	`, failureSummary, failureStacktrace, id))
}

var indexColumnsWithNullRank = []*sqlf.Query{
	sqlf.Sprintf("id"),
	sqlf.Sprintf("commit"),
	sqlf.Sprintf("queued_at"),
	sqlf.Sprintf("state"),
	sqlf.Sprintf("failure_summary"),
	sqlf.Sprintf("failure_stacktrace"),
	sqlf.Sprintf("started_at"),
	sqlf.Sprintf("finished_at"),
	sqlf.Sprintf("repository_id"),
	sqlf.Sprintf("NULL"),
}

// DequeueIndex selects the oldest queued index and locks it with a transaction. If there is such an index,
// the index is returned along with a DB instance which wraps the transaction. This transaction must be closed.
// If there is no such unlocked index, a zero-value index and nil DB will be returned along with a false
// valued flag. This method must not be called from within a transaction.
func (db *dbImpl) DequeueIndex(ctx context.Context) (Index, DB, bool, error) {
	index, tx, ok, err := db.dequeueRecord(ctx, "lsif_indexes", indexColumnsWithNullRank, sqlf.Sprintf("queued_at"), scanFirstIndexDequeue)
	if err != nil || !ok {
		return Index{}, tx, ok, err
	}

	return index.(Index), tx, true, nil
}
