package iam

import (
	"context"
	"testing"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func TestLicenseKeyFieldsMigrator(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	store := basestore.NewWithHandle(db.Handle())

	// Set up test data
	userID, _, err := basestore.ScanFirstInt(store.Query(ctx, sqlf.Sprintf(`INSERT INTO users(username, display_name, created_at) VALUES(%s, %s, NOW()) RETURNING id`, "alice", "alice")))
	require.NoError(t, err)

	subscriptionID, _, err := basestore.ScanFirstString(store.Query(ctx, sqlf.Sprintf(`INSERT INTO product_subscriptions(id, user_id) VALUES(gen_random_uuid(), %s) RETURNING id`, userID)))
	require.NoError(t, err)
	require.NotEmpty(t, subscriptionID)

	licenseID, _, err := basestore.ScanFirstString(store.Query(ctx, sqlf.Sprintf(`INSERT INTO product_licenses(id, product_subscription_id, license_key) VALUES(gen_random_uuid(), %s, %s) RETURNING id`,
		subscriptionID,
		`eyJzaWciOnsiRm9ybWF0Ijoic3NoLXJzYSIsIkJsb2IiOiIuLi4iLCJSZXN0IjpudWxsfSwiaW5mbyI6ImV5SjJJam94TENKdUlqcGJNVEk0TERrd0xESTBOaXd5TkRRc05qWXNNVFFzTWpVMUxEZ3hYU3dpZENJNld5SmtaWFlpWFN3aWRTSTZPQ3dpWlNJNklqSXdNak10TURZdE1ERlVNVFk2TWpnNk16WmFJbjA9In0`,
	)))
	require.NoError(t, err)

	// Ensure there is no progress before migration
	migrator := NewLicenseKeyFieldsMigrator(store, 500)
	progress, err := migrator.Progress(ctx, false)
	require.NoError(t, err)
	require.Equal(t, 0.0, progress)

	// Perform the migration and recheck the progress
	err = migrator.Up(ctx)
	require.NoError(t, err)

	progress, err = migrator.Progress(ctx, false)
	require.NoError(t, err)
	require.Equal(t, 1.0, progress)

	// Ensure data are at desired states
	var (
		licenseVersion   int
		licenseTags      []string
		licenseUserCount int
		licenseExpiresAt time.Time
	)
	err = store.QueryRow(ctx, sqlf.Sprintf(`SELECT license_version, license_tags, license_user_count, license_expires_at FROM product_licenses WHERE id = %s`, licenseID)).Scan(&licenseVersion, pq.Array(&licenseTags), &licenseUserCount, &licenseExpiresAt)
	require.NoError(t, err)
	assert.Equal(t, 1, licenseVersion)
	assert.Equal(t, []string{"dev"}, licenseTags)
	assert.Equal(t, 8, licenseUserCount)

	wantExpiresAt, err := time.Parse(time.RFC3339, "2023-06-01T16:28:36Z")
	require.NoError(t, err)
	assert.Equal(t, wantExpiresAt, licenseExpiresAt)
}
