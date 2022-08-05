package productsubscription

import (
	"context"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func TestSubscriptionAccountNumberMigrator(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))

	// Set up test data
	alice, err := db.Users().Create(ctx, database.NewUser{Username: "alice"}) // A username without account number shouldn't fail
	require.NoError(t, err)
	bob, err := db.Users().Create(ctx, database.NewUser{Username: "bob-11033746"})
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `INSERT INTO product_subscriptions(id, user_id) VALUES(gen_random_uuid(), $1)`, alice.ID)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO product_subscriptions(id, user_id) VALUES(gen_random_uuid(), $1)`, bob.ID)
	require.NoError(t, err)

	// Ensure there is no progress before migration
	migrator := &subscriptionAccountNumberMigrator{store: basestore.NewWithHandle(db.Handle())}
	progress, err := migrator.Progress(ctx)
	require.NoError(t, err)
	require.Equal(t, 0.0, progress)

	// Perform the migration and recheck the progress
	err = migrator.Up(ctx)
	require.NoError(t, err)

	progress, err = migrator.Progress(ctx)
	require.NoError(t, err)
	require.Equal(t, 1.0, progress)

	// Ensure data are at desired states
	var accountNumber string
	err = db.QueryRowContext(ctx, `SELECT account_number FROM product_subscriptions WHERE user_id = $1`, alice.ID).Scan(&accountNumber)
	require.NoError(t, err)
	assert.Empty(t, accountNumber)

	err = db.QueryRowContext(ctx, `SELECT account_number FROM product_subscriptions WHERE user_id = $1`, bob.ID).Scan(&accountNumber)
	require.NoError(t, err)
	assert.Equal(t, "11033746", accountNumber)
}

func TestLicenseKeyFieldsMigrator(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))

	// Set up test data
	alice, err := db.Users().Create(ctx, database.NewUser{Username: "alice"})
	require.NoError(t, err)

	var subscriptionID string
	err = db.QueryRowContext(ctx, `INSERT INTO product_subscriptions(id, user_id) VALUES(gen_random_uuid(), $1) RETURNING id`, alice.ID).Scan(&subscriptionID)
	require.NoError(t, err)
	require.NotEmpty(t, subscriptionID)

	var licenseID string
	err = db.QueryRowContext(ctx, `
INSERT INTO product_licenses(id, product_subscription_id, license_key)
VALUES(gen_random_uuid(), $1, $2)
RETURNING id
`,
		subscriptionID,
		`eyJzaWciOnsiRm9ybWF0Ijoic3NoLXJzYSIsIkJsb2IiOiIuLi4iLCJSZXN0IjpudWxsfSwiaW5mbyI6ImV5SjJJam94TENKdUlqcGJNVEk0TERrd0xESTBOaXd5TkRRc05qWXNNVFFzTWpVMUxEZ3hYU3dpZENJNld5SmtaWFlpWFN3aWRTSTZPQ3dpWlNJNklqSXdNak10TURZdE1ERlVNVFk2TWpnNk16WmFJbjA9In0`,
	).Scan(&licenseID)
	require.NoError(t, err)

	// Ensure there is no progress before migration
	migrator := &licenseKeyFieldsMigrator{store: basestore.NewWithHandle(db.Handle())}
	progress, err := migrator.Progress(ctx)
	require.NoError(t, err)
	require.Equal(t, 0.0, progress)

	// Perform the migration and recheck the progress
	err = migrator.Up(ctx)
	require.NoError(t, err)

	progress, err = migrator.Progress(ctx)
	require.NoError(t, err)
	require.Equal(t, 1.0, progress)

	// Ensure data are at desired states
	var licenseVersion int
	var licenseTags []string
	var licenseUserCount int
	var licenseExpiresAt time.Time
	err = db.QueryRowContext(ctx, `SELECT license_version, license_tags, license_user_count, license_expires_at FROM product_licenses WHERE id = $1`, licenseID).
		Scan(&licenseVersion, pq.Array(&licenseTags), &licenseUserCount, &licenseExpiresAt)
	require.NoError(t, err)
	assert.Equal(t, 1, licenseVersion)
	assert.Equal(t, []string{"dev"}, licenseTags)
	assert.Equal(t, 8, licenseUserCount)

	wantExpiresAt, err := time.Parse(time.RFC3339, "2023-06-01T16:28:36Z")
	require.NoError(t, err)
	assert.Equal(t, wantExpiresAt, licenseExpiresAt)
}
