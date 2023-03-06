package iam

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
)

type unifiedPermissionsMigrator struct {
	store           *basestore.Store
	batchSize       int
	intervalSeconds int // time interval in seconds between iterations
}

var _ oobmigration.Migrator = &unifiedPermissionsMigrator{}

// shamelessly copied from scip_migrator.go
func getEnv(name string, defaultValue int) int {
	if value, _ := strconv.Atoi(os.Getenv(name)); value != 0 {
		return value
	}

	return defaultValue
}

func NewUnifiedPermissionsMigrator(store *basestore.Store, batchSize, intervalSeconds int) *unifiedPermissionsMigrator {
	computedBatchSize := getEnv("UNIFIED_PERMISSIONS_MIGRATOR_BATCH_SIZE", batchSize)
	computedInterval := getEnv("UNIFIED_PERMISSIONS_MIGRATOR_INTERVAL_SECONDS", intervalSeconds)

	return &unifiedPermissionsMigrator{
		store:           store,
		batchSize:       computedBatchSize,
		intervalSeconds: computedInterval,
	}
}

func (m *unifiedPermissionsMigrator) ID() int { return 22 }
func (m *unifiedPermissionsMigrator) Interval() time.Duration {
	return time.Duration(m.intervalSeconds) * time.Second
}

func (m *unifiedPermissionsMigrator) Progress(ctx context.Context, _ bool) (float64, error) {
	progress, _, err := basestore.ScanFirstFloat(m.store.Query(ctx, sqlf.Sprintf(unifiedPermissionsMigratorProgressQuery)))
	return progress, err
}

const unifiedPermissionsMigratorProgressQuery = `
SELECT
	CASE c2.count WHEN 0 THEN 1 ELSE
		cast(c1.count as float) / cast(c2.count as float)
	END
FROM
	(SELECT count(*) as count FROM user_permissions WHERE migrated) c1,
	(SELECT count(*) as count FROM user_permissions) c2
`

func (m *unifiedPermissionsMigrator) Up(ctx context.Context) (err error) {
	tx, err := m.store.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	// get the next batch of userIDs to migrate
	userIds, err := basestore.ScanInt32s(tx.Query(ctx, sqlf.Sprintf(unifiedPermissionsNonMigratedUsersQuery, m.batchSize)))
	if err != nil {
		return err
	}

	// write data to user_repo_permissions table
	err = tx.Exec(ctx, sqlf.Sprintf(unifiedPermissionsMigratorQuery, pq.Array(userIds)))
	if err != nil {
		return err
	}

	// mark the userIDs as migrated
	updates := make([]*sqlf.Query, 0, len(userIds))
	for _, userID := range userIds {
		updates = append(updates, sqlf.Sprintf("(%s::integer)", userID))
	}
	if err := tx.Exec(ctx, sqlf.Sprintf(unifiedPermissionsMigratorMarkAsMigratedQuery, sqlf.Join(updates, ", "))); err != nil {
		return err
	}

	return nil
}

// The migration is based on user_permissions table, because current customer instances
// usually have only a few thousand users and potentially tens of thousands of repositories.
// So it should be more efficient to cycle through users instead of repositories.

// Query to get the userIds to migrate
const unifiedPermissionsNonMigratedUsersQuery = `
SELECT u.user_id
	FROM user_permissions as u
	WHERE NOT migrated
	LIMIT %s
	FOR UPDATE SKIP LOCKED
`

// Migration of data to new unified table
const unifiedPermissionsMigratorQuery = `
-- First get a row for each pair of user_id, repo_id by unnesting object_ids_ints array
WITH s AS (
	SELECT u.user_id, unnest(u.object_ids_ints) as repo_id, u.updated_at as created_at, u.updated_at, 'user_sync' as source
	FROM user_permissions as u
	WHERE u.user_id = ANY(%s)
)
-- Insert the data here
INSERT INTO user_repo_permissions(user_id, user_external_account_id, repo_id, created_at, updated_at, source)
SELECT s.user_id, ua.id as user_external_account_id, s.repo_id, s.created_at, s.updated_at, s.source
FROM
  s
-- Need to join with repo table to not transfer permissions for deleted repos
INNER JOIN repo AS r ON
  r.deleted_at IS NULL AND r.id = s.repo_id
-- Need to join with the user_external_accounst table to get the user_external_account_id
INNER JOIN user_external_accounts AS ua ON
  ua.user_id = s.user_id AND ua.deleted_at IS NULL
  AND ua.service_type = r.external_service_type
  AND ua.service_id = r.external_service_id
-- It might be that some of the rows are already there because of repo-centric permission sync
-- In that case do nothing
ON CONFLICT (user_id, user_external_account_id, repo_id) DO NOTHING
RETURNING user_repo_permissions.id
`

// Mark the userIDs as migrated
const unifiedPermissionsMigratorMarkAsMigratedQuery = `
UPDATE user_permissions
SET migrated = TRUE
FROM (VALUES %s) AS updates(user_id)
WHERE user_permissions.user_id = updates.user_id
`

func (m *unifiedPermissionsMigrator) Down(_ context.Context) error {
	// non-destructive
	return nil
}
