package iam

import (
	"context"
	"testing"

	"github.com/keegancsmith/sqlf"
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
	db := database.NewDB(logger, dbtest.NewDB(t))
	store := basestore.NewWithHandle(db.Handle())

	// Set up test data
	aliceID, _, err := basestore.ScanFirstInt(store.Query(ctx, sqlf.Sprintf(`INSERT INTO users(username, display_name, created_at) VALUES(%s, %s, NOW()) RETURNING id`, "alice", "alice")))
	require.NoError(t, err)
	bobID, _, err := basestore.ScanFirstInt(store.Query(ctx, sqlf.Sprintf(`INSERT INTO users(username, display_name, created_at) VALUES(%s, %s, NOW()) RETURNING id`, "bob-11033746", "bob")))
	require.NoError(t, err)

	err = store.Exec(ctx, sqlf.Sprintf(`INSERT INTO product_subscriptions(id, user_id) VALUES(gen_random_uuid(), %s)`, aliceID))
	require.NoError(t, err)
	err = store.Exec(ctx, sqlf.Sprintf(`INSERT INTO product_subscriptions(id, user_id) VALUES(gen_random_uuid(), %s)`, bobID))
	require.NoError(t, err)

	// Ensure there is no progress before migration
	migrator := NewSubscriptionAccountNumberMigrator(store, 500)
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
	var accountNumber string
	accountNumber, _, err = basestore.ScanFirstString(store.Query(ctx, sqlf.Sprintf(`SELECT account_number FROM product_subscriptions WHERE user_id = %s`, aliceID)))
	require.NoError(t, err)
	assert.Empty(t, accountNumber)

	accountNumber, _, err = basestore.ScanFirstString(store.Query(ctx, sqlf.Sprintf(`SELECT account_number FROM product_subscriptions WHERE user_id = %s`, bobID)))
	require.NoError(t, err)
	assert.Equal(t, "11033746", accountNumber)
}
