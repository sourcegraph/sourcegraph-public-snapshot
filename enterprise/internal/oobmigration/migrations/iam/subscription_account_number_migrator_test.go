package iam

import (
	"context"
	"testing"

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
	migrator := NewSubscriptionAccountNumberMigrator(basestore.NewWithHandle(db.Handle()))
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
