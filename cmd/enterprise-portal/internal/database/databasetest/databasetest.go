package databasetest

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

// NewTestDB creates a new test database and initializes the given list of
// tables for the suite. The test database is dropped after testing is completed
// unless failed.
func NewTestDB(t testing.TB, system, suite string, tables ...schema.Tabler) *pgxpool.Pool {
	if testing.Short() {
		t.Skip("skipping DB test since -short specified")
	}

	dsn, err := dbtest.GetDSN()
	require.NoError(t, err)

	// Open a connection to control the test database lifecycle.
	sqlDB, err := sql.Open("pgx", dsn.String())
	require.NoError(t, err)

	// Set up test suite database.
	dbName := fmt.Sprintf("sourcegraph-test-%s-%s-%d", system, suite, time.Now().Unix())
	_, err = sqlDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %q", dbName))
	require.NoError(t, err)

	_, err = sqlDB.Exec(fmt.Sprintf("CREATE DATABASE %q WITH ENCODING=UTF8", dbName))
	require.NoError(t, err)

	// Swap out the database name to be the test suite database in the DSN.
	dsn.Path = "/" + dbName

	now := time.Now().UTC().Truncate(time.Second)
	db, err := gorm.Open(
		postgres.Open(dsn.String()),
		&gorm.Config{
			SkipDefaultTransaction: true,
			NowFunc: func() time.Time {
				return now
			},
		},
	)
	require.NoError(t, err)
	for _, table := range tables {
		err = db.AutoMigrate(table)
		require.NoError(t, err)
	}

	// Close the connection used to auto-migrate the database.
	migrateDB, err := db.DB()
	require.NoError(t, err)
	err = migrateDB.Close()
	require.NoError(t, err)

	// Open a new connection to the test suite database.
	testDB, err := pgxpool.New(context.Background(), dsn.String())
	require.NoError(t, err)

	t.Cleanup(func() {
		if t.Failed() {
			t.Logf("Database %q left intact for inspection", dbName)
			return
		}

		testDB.Close()

		_, err = sqlDB.Exec(fmt.Sprintf(`DROP DATABASE %q`, dbName))
		if err != nil {
			t.Errorf("Failed to drop test suite database %q: %v", dbName, err)
		}
		err = sqlDB.Close()
		if err != nil {
			t.Errorf("Failed to close test database connection %q: %v", dbName, err)
		}
	})

	return testDB
}

// ClearTablesAfterTest removes all rows from the list of tables in the original
// order as a t.Cleanup hook. It uses soft-deletion when available and skips
// deletion when the test suite failed.
func ClearTablesAfterTest(t *testing.T, db *pgxpool.Pool, tables ...schema.Tabler) {
	t.Cleanup(func() {
		if t.Failed() {
			t.Log("Leaving table data intact after test failure")
			return
		}

		tableNames := make([]string, 0, len(tables))
		for _, t := range tables {
			tableNames = append(tableNames, t.TableName())
		}
		_, err := db.Exec(context.Background(), "TRUNCATE TABLE "+strings.Join(tableNames, ", ")+" RESTART IDENTITY")
		if err != nil {
			t.Errorf("Failed to clear table data: %v", err)
		}
	})
}
