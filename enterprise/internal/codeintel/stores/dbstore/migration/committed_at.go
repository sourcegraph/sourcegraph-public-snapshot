package migration

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	basegitserver "github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/internal/vcs"
)

type committedAtMigrator struct {
	store           *dbstore.Store
	gitserverClient GitserverClient
	batchSize       int
}

// NewCommittedAtMigrator creates a new Migrator instance that reads records from
// the lsif_uploads table and populates the record's committed_at column based on
// data from gitserver.
func NewCommittedAtMigrator(store *dbstore.Store, gitserverClient GitserverClient, batchSize int) oobmigration.Migrator {
	return &committedAtMigrator{
		store:           store,
		gitserverClient: gitserverClient,
		batchSize:       batchSize,
	}
}

// Progress returns the ratio between the number of upload records that have been
// completely migrated over the total number of upload records. This simply counts
// the number of completed upload records with and without a value for committed_at.
func (m *committedAtMigrator) Progress(ctx context.Context) (float64, error) {
	progress, _, err := basestore.ScanFirstFloat(m.store.Query(ctx, sqlf.Sprintf(committedAtProgressQuery)))
	return progress, err
}

const committedAtProgressQuery = `
-- source: enterprise/internal/codeintel/stores/dbstore/migration/committed_at.go:Progress
SELECT CASE c2.count WHEN 0 THEN 1 ELSE CAST(c1.count AS float) / CAST(c2.count AS float) END FROM
(SELECT COUNT(*) as count FROM lsif_uploads WHERE state = 'completed' AND committed_at IS NOT NULL) c1,
(SELECT COUNT(*) as count FROM lsif_uploads WHERE state = 'completed') c2
`

// Up runs a batch of the migration. This method selects a batch of unique repository
// and commit pairs, then sets the committed_at field for all matching uploads. In
// this sense, the batch size controls the maximum number of gitserver requests, not
// the number of records updated.
func (m *committedAtMigrator) Up(ctx context.Context) (err error) {
	tx, err := m.store.Transact(ctx)
	defer func() {
		err = tx.Done(err)
	}()

	batch, err := m.selectBatch(ctx, tx)
	if err != nil {
		return err
	}

	return m.processBatch(ctx, tx, batch)
}

func (m *committedAtMigrator) selectBatch(ctx context.Context, tx *dbstore.Store) (_ map[int][]string, err error) {
	rows, err := tx.Query(ctx, sqlf.Sprintf(committedAtSelectBatchQuery, m.batchSize))
	if err != nil {
		return nil, err
	}
	defer func() {
		err = basestore.CloseRows(rows, err)
	}()

	batch := map[int][]string{}
	for rows.Next() {
		var repositoryID int
		var commit string
		if err := rows.Scan(&repositoryID, &commit); err != nil {
			return nil, err
		}

		batch[repositoryID] = append(batch[repositoryID], commit)
	}

	return batch, nil
}

const committedAtSelectBatchQuery = `
-- source: enterprise/internal/codeintel/stores/dbstore/migration/committed_at.go:selectBatch
SELECT repository_id, commit
FROM lsif_uploads
WHERE state = 'completed' AND committed_at IS NULL
GROUP BY repository_id, commit
LIMIT %s
`

// processBatch will find the commit date of each repository/commit pair in the given batch
// and issue update statements to the database. Each repository	present in the batch will also
// be marked as dirty so the commit graph can be recalculated (correctly).
//
// Note that we're VERY cautious about the order that we process the batch. We do this so that
// two frontends performing the same migration will not try to update two of the same records
// in the opposite order. If we rely on map iteration order we tend to see a lot of Postgres
// deadlock conditions and very slow migration progress.
func (m *committedAtMigrator) processBatch(ctx context.Context, tx *dbstore.Store, batch map[int][]string) error {
	repositoryIDs := make([]int, 0, len(batch))
	for repositoryID := range batch {
		repositoryIDs = append(repositoryIDs, repositoryID)
	}
	sort.Ints(repositoryIDs)

	for _, repositoryID := range repositoryIDs {
		commits := batch[repositoryID]
		sort.Strings(commits)

		// Note: this is difficult to combine since if we pass in one bad commit it destroys
		// the entire request with a fatal: bad object <unknown sha>. We should at some point
		// come back to this and figure out how to batch these so we're not doing so many
		// gitserver roundtrips on these kind of background tasks for code intelligence.
		for _, commit := range commits {
			var commitDateString string
			if commitDate, err := m.gitserverClient.CommitDate(ctx, repositoryID, commit); err != nil {
				if !isRepositoryNotFound(err) && !isRevisionNotFound(err) {
					return err
				}

				// Set a value here that we'll filter out on the query side so that we don't
				// reprocess the same failing batch infinitely. We could alternative soft
				// delete the record, but it would be better to keep record deletion behavior
				// together in the same place (so we have unified metrics on that event).
				commitDateString = "-infinity"
			} else {
				commitDateString = commitDate.Format(time.RFC3339)
			}

			// Update commit date of all uploads attached to this this repository and commit
			if err := tx.Exec(ctx, sqlf.Sprintf(committedAtProcessBatchUpdateCommitDateQuery, commitDateString, repositoryID, commit)); err != nil {
				return err
			}
		}

		// Mark repository as dirty so the commit graph is recalculated with fresh data
		if err := tx.MarkRepositoryAsDirty(ctx, repositoryID); err != nil {
			return err
		}
	}

	return nil
}

const committedAtProcessBatchUpdateCommitDateQuery = `
-- source: enterprise/internal/codeintel/stores/dbstore/migration/committed_at.go:processBatch
UPDATE lsif_uploads SET committed_at = %s WHERE state = 'completed' AND repository_id = %s AND commit = %s AND committed_at IS NULL
`

// Down runs a batch of the migration in reverse. This method simply sets the committed_at column
// to null for a number of records matching the configured batch size.
func (m *committedAtMigrator) Down(ctx context.Context) error {
	return m.store.Exec(ctx, sqlf.Sprintf(committedAtDownQuery, m.batchSize))
}

const committedAtDownQuery = `
-- source: enterprise/internal/codeintel/stores/dbstore/migration/committed_at.go:Down
UPDATE lsif_uploads SET committed_at = NULL WHERE id IN (SELECT id FROM lsif_uploads WHERE state = 'completed' AND committed_at IS NOT NULL LIMIT %s)
`

func isRepositoryNotFound(err error) bool {
	for ex := err; ex != nil; ex = errors.Unwrap(ex) {
		if vcs.IsRepoNotExist(ex) {
			return true
		}
	}

	return false
}

func isRevisionNotFound(err error) bool {
	for ex := err; ex != nil; ex = errors.Unwrap(ex) {
		if basegitserver.IsRevisionNotFound(ex) {
			return true
		}
	}

	return false
}
