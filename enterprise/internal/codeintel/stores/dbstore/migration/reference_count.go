package migration

import (
	"context"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
)

type referenceCountMigrator struct {
	store     *dbstore.Store
	batchSize int
}

func NewReferenceCountMigrator(store *dbstore.Store, batchSize int) oobmigration.Migrator {
	return &referenceCountMigrator{
		store:     store,
		batchSize: batchSize,
	}
}

// Progress returns the ratio between the number of upload records that have been
// completely migrated over the total number of upload records. This simply counts
// the number of completed upload records with and without a value for num_references.
func (m *referenceCountMigrator) Progress(ctx context.Context) (float64, error) {
	progress, _, err := basestore.ScanFirstFloat(m.store.Query(ctx, sqlf.Sprintf(referenceCountProgressQuery)))
	return progress, err
}

const referenceCountProgressQuery = `
-- source: enterprise/internal/codeintel/stores/dbstore/migration/reference_count.go:Progress
SELECT CASE c2.count WHEN 0 THEN 1 ELSE CAST(c1.count AS float) / CAST(c2.count AS float) END FROM
(SELECT COUNT(*) as count FROM lsif_uploads WHERE state = 'completed' AND num_references IS NOT NULL) c1,
(SELECT COUNT(*) as count FROM lsif_uploads WHERE state = 'completed') c2
`

// Up runs a batch of the migration. This method TODO
func (m *referenceCountMigrator) Up(ctx context.Context) (err error) {
	tx, err := m.store.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	ids, err := basestore.ScanInts(m.store.Query(ctx, sqlf.Sprintf(referenceCountUpQuery, m.batchSize)))
	if err != nil {
		return err
	}

	return m.store.UpdateNumReferences(ctx, ids)
}

const referenceCountUpQuery = `
-- source: enterprise/internal/codeintel/stores/dbstore/migration/reference_count.go:Up
SELECT u.id
FROM lsif_uploads u
WHERE u.state = 'completed' AND u.num_references IS NULL
ORDER BY u.id
FOR UPDATE SKIP LOCKED
LIMIT %s
`

// Down runs a batch of the migration in reverse. This method simply sets the num_references
// column to null for a number of records matching the configured batch size.
func (m *referenceCountMigrator) Down(ctx context.Context) error {
	return m.store.Exec(ctx, sqlf.Sprintf(referenceCountDownQuery, m.batchSize))
}

const referenceCountDownQuery = `
-- source: enterprise/internal/codeintel/stores/dbstore/migration/reference_count.go:Down
UPDATE lsif_uploads SET num_references = NULL WHERE id IN (SELECT id FROM lsif_uploads WHERE state = 'completed' AND num_references IS NOT NULL LIMIT %s)
`
