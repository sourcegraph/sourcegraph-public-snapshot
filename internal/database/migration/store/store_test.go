package store

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"
	"golang.org/x/sync/errgroup"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/definition"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/storetypes"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestEnsureSchemaTable(t *testing.T) {
	db := dbtest.NewDB(t)
	store := testStore(db)
	ctx := context.Background()

	if err := store.Exec(ctx, sqlf.Sprintf("SELECT * FROM test_migrations_table")); err == nil {
		t.Fatalf("expected query to fail due to missing schema table")
	}

	if err := store.Exec(ctx, sqlf.Sprintf("SELECT * FROM migration_logs")); err == nil {
		t.Fatalf("expected query to fail due to missing logs table")
	}

	if err := store.EnsureSchemaTable(ctx); err != nil {
		t.Fatalf("unexpected error ensuring schema table exists: %s", err)
	}

	if err := store.Exec(ctx, sqlf.Sprintf("SELECT * FROM test_migrations_table")); err != nil {
		t.Fatalf("unexpected error querying version table: %s", err)
	}

	if err := store.Exec(ctx, sqlf.Sprintf("SELECT * FROM migration_logs")); err != nil {
		t.Fatalf("unexpected error querying logs table: %s", err)
	}

	if err := store.EnsureSchemaTable(ctx); err != nil {
		t.Fatalf("expected method to be idempotent, got error: %s", err)
	}
}

func TestVersion(t *testing.T) {
	db := dbtest.NewDB(t)
	store := testStore(db)
	ctx := context.Background()

	if err := store.EnsureSchemaTable(ctx); err != nil {
		t.Fatalf("unexpected error ensuring schema table exists: %s", err)
	}

	t.Run("empty", func(*testing.T) {
		if _, _, ok, err := store.Version(ctx); err != nil {
			t.Fatalf("unexpected error querying version: %s", err)
		} else if ok {
			t.Fatalf("unexpected version")
		}
	})

	testCases := []struct {
		name    string
		version int
		dirty   bool
	}{
		{"clean", 25, false},
		{"dirty", 32, true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if err := store.Exec(ctx, sqlf.Sprintf(`DELETE FROM test_migrations_table`)); err != nil {
				t.Fatalf("unexpected error clearing data: %s", err)
			}
			if err := store.Exec(ctx, sqlf.Sprintf(`INSERT INTO test_migrations_table VALUES (%s, %s)`, testCase.version, testCase.dirty)); err != nil {
				t.Fatalf("unexpected error inserting data: %s", err)
			}

			if version, dirty, ok, err := store.Version(ctx); err != nil {
				t.Fatalf("unexpected error querying version: %s", err)
			} else if !ok {
				t.Fatalf("expected a version to be found")
			} else if version != testCase.version {
				t.Fatalf("unexpected version. want=%d have=%d", testCase.version, version)
			} else if dirty != testCase.dirty {
				t.Fatalf("unexpected dirty flag. want=%v have=%v", testCase.dirty, dirty)
			}
		})
	}
}

func TestVersions(t *testing.T) {
	db := dbtest.NewDB(t)
	store := testStore(db)
	ctx := context.Background()
	if err := store.EnsureSchemaTable(ctx); err != nil {
		t.Fatalf("unexpected error ensuring schema table exists: %s", err)
	}

	t.Run("empty", func(*testing.T) {
		if appliedVersions, pendingVersions, failedVersions, err := store.Versions(ctx); err != nil {
			t.Fatalf("unexpected error querying versions: %s", err)
		} else if len(appliedVersions)+len(pendingVersions)+len(failedVersions) > 0 {
			t.Fatalf("unexpected no versions, got applied=%v pending=%v failed=%v", appliedVersions, pendingVersions, failedVersions)
		}
	})

	type testCase struct {
		version      int
		up           bool
		success      *bool
		errorMessage *string
	}
	makeCase := func(version int, up bool, failed *bool) testCase {
		if failed == nil {
			return testCase{version, up, nil, nil}
		}
		if *failed {
			return testCase{version, up, boolPtr(false), strPtr("uh-oh")}
		}
		return testCase{version, up, boolPtr(true), nil}
	}

	for _, migrationLog := range []testCase{
		// Historic attempts
		makeCase(1003, true, boolPtr(true)), makeCase(1003, false, boolPtr(true)), // 1003: successful up, successful down
		makeCase(1004, true, boolPtr(true)),                                       // 1004: successful up
		makeCase(1006, true, boolPtr(false)), makeCase(1006, true, boolPtr(true)), // 1006: failed up, successful up

		// Last attempts
		makeCase(1001, true, boolPtr(false)),  // successful up
		makeCase(1002, false, boolPtr(false)), // successful down
		makeCase(1003, true, nil),             // pending up
		makeCase(1004, false, nil),            // pending down
		makeCase(1005, true, boolPtr(true)),   // failed up
		makeCase(1006, false, boolPtr(true)),  // failed down
	} {
		if err := store.Exec(ctx, sqlf.Sprintf(`INSERT INTO migration_logs (
				migration_logs_schema_version,
				schema,
				version,
				up,
				started_at,
				success,
				finished_at,
				error_message
			) VALUES (%s, %s, %s, %s, NOW(), %s, NOW(), %s)`,
			currentMigrationLogSchemaVersion,
			"test_migrations_table",
			migrationLog.version,
			migrationLog.up,
			migrationLog.success,
			migrationLog.errorMessage,
		)); err != nil {
			t.Fatalf("unexpected error inserting data: %s", err)
		}
	}

	assertVersions(
		t,
		ctx,
		store,
		[]int{1001},       // expectedAppliedVersions
		[]int{1003, 1004}, // expectedPendingVersions
		[]int{1005, 1006}, // expectedFailedVersions
	)
}

func TestLock(t *testing.T) {
	db := dbtest.NewDB(t)
	store := testStore(db)
	ctx := context.Background()

	t.Run("sanity test", func(t *testing.T) {
		acquired, close, err := store.Lock(ctx)
		if err != nil {
			t.Fatalf("unexpected error acquiring lock: %s", err)
		}
		if !acquired {
			t.Fatalf("expected lock to be acquired")
		}

		if err := close(nil); err != nil {
			t.Fatalf("unexpected error releasing lock: %s", err)
		}
	})
}

func TestTryLock(t *testing.T) {
	db := dbtest.NewDB(t)
	store := testStore(db)
	ctx := context.Background()

	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("failed to open new connection: %s", err)
	}
	t.Cleanup(func() { conn.Close() })

	// Acquire lock in distinct session
	if _, err := conn.ExecContext(ctx, `SELECT pg_advisory_lock($1, 0)`, store.lockKey()); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	// TryLock should fail
	if acquired, _, err := store.TryLock(ctx); err != nil {
		t.Fatalf("unexpected error acquiring lock: %s", err)
	} else if acquired {
		t.Fatalf("expected lock to be held by another session")
	}

	// Drop lock
	if _, err := conn.ExecContext(ctx, `SELECT pg_advisory_unlock($1, 0)`, store.lockKey()); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	// TryLock should succeed
	acquired, unlock, err := store.TryLock(ctx)
	if err != nil {
		t.Fatalf("unexpected error acquiring lock: %s", err)
	} else if !acquired {
		t.Fatalf("expected lock to be acquired")
	}

	if err := unlock(nil); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	// Check idempotency
	if err := unlock(nil); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestWrappedUp(t *testing.T) {
	db := dbtest.NewDB(t)
	store := testStore(db)
	ctx := context.Background()

	if err := store.EnsureSchemaTable(ctx); err != nil {
		t.Fatalf("unexpected error ensuring schema table exists: %s", err)
	}
	if err := store.Exec(ctx, sqlf.Sprintf(`INSERT INTO test_migrations_table VALUES (15, false)`)); err != nil {
		t.Fatalf("unexpected error setting initial version: %s", err)
	}

	t.Run("success", func(t *testing.T) {
		definition := definition.Definition{
			ID: 16,
			UpQuery: sqlf.Sprintf(`
				CREATE TABLE test_trees (
					name text,
					leaf_type text,
					seed_type text,
					bark_type text
				);

				INSERT INTO test_trees VALUES
					('oak', 'broad', 'regular', 'strong'),
					('birch', 'narrow', 'regular', 'flaky'),
					('pine', 'needle', 'pine cone', 'soft');
			`),
		}
		f := func() error {
			return store.Up(ctx, definition)
		}
		if err := store.WithMigrationLog(ctx, definition, true, f); err != nil {
			t.Fatalf("unexpected error running migration: %s", err)
		}

		if barkType, _, err := basestore.ScanFirstString(store.Query(ctx, sqlf.Sprintf(`SELECT bark_type FROM test_trees WHERE name = 'birch'`))); err != nil {
			t.Fatalf("migration query did not succeed; unexpected error querying test table: %s", err)
		} else if barkType != "flaky" {
			t.Fatalf("migration query did not succeed; unexpected bark type. want=%s have=%s", "flaky", barkType)
		}

		// Version set to migration ID; not dirty
		if version, dirty, ok, err := store.Version(ctx); err != nil || !ok || dirty || version != 16 {
			t.Fatalf("unexpected version. want=(version=%d, dirty=%v), have=(version=%d, dirty=%v, ok=%v, error=%q)", 16, false, version, dirty, ok, err)
		}

		assertLogs(t, ctx, store, []migrationLog{
			{
				Schema:  "test_migrations_table",
				Version: 16,
				Up:      true,
				Success: boolPtr(true),
			},
		})
		assertVersions(t, ctx, store, []int{16}, nil, nil)
		truncateLogs(t, ctx, store)
	})

	t.Run("unexpected version", func(t *testing.T) {
		expectedErrorMessage := "expected schema to have version 17, but has version 16"

		definition := definition.Definition{
			ID: 18,
			UpQuery: sqlf.Sprintf(`
				-- Does not actually run
			`),
		}
		f := func() error {
			return store.Up(ctx, definition)
		}
		if err := store.WithMigrationLog(ctx, definition, true, f); err == nil || !strings.HasPrefix(err.Error(), expectedErrorMessage) {
			t.Fatalf("unexpected error want=%q have=%q", expectedErrorMessage, err)
		}

		// Version, dirty status unchanged
		if version, dirty, ok, err := store.Version(ctx); err != nil || !ok || dirty || version != 16 {
			t.Fatalf("unexpected version. want=(version=%d, dirty=%v), have=(version=%d, dirty=%v, ok=%v, error=%q)", 16, false, version, dirty, ok, err)
		}

		assertLogs(t, ctx, store, nil)
	})

	t.Run("query failure", func(t *testing.T) {
		expectedErrorMessage := "SQL Error"

		definition := definition.Definition{
			ID: 17,
			UpQuery: sqlf.Sprintf(`
				-- Note: table already exists
				CREATE TABLE test_trees (
					name text,
					leaf_type text,
					seed_type text,
					bark_type text
				);
			`),
		}
		f := func() error {
			return store.Up(ctx, definition)
		}
		if err := store.WithMigrationLog(ctx, definition, true, f); err == nil || !strings.HasPrefix(err.Error(), expectedErrorMessage) {
			t.Fatalf("unexpected error want=%q have=%q", expectedErrorMessage, err)
		}

		// Version set to migration ID; dirty
		if version, dirty, ok, err := store.Version(ctx); err != nil || !ok || !dirty || version != 17 {
			t.Fatalf("unexpected version. want=(version=%d, dirty=%v), have=(version=%d, dirty=%v, ok=%v, error=%q)", 17, true, version, dirty, ok, err)
		}

		assertLogs(t, ctx, store, []migrationLog{
			{
				Schema:  "test_migrations_table",
				Version: 17,
				Up:      true,
				Success: boolPtr(false),
			},
		})
		assertVersions(t, ctx, store, nil, nil, []int{17})
		truncateLogs(t, ctx, store)
	})

	t.Run("dirty", func(t *testing.T) {
		expectedErrorMessage := "dirty database"

		definition := definition.Definition{
			ID: 17,
			UpQuery: sqlf.Sprintf(`
				-- Does not actually run
			`),
		}
		f := func() error {
			return store.Up(ctx, definition)
		}
		if err := store.WithMigrationLog(ctx, definition, true, f); err == nil || !strings.HasPrefix(err.Error(), expectedErrorMessage) {
			t.Fatalf("unexpected error want=%q have=%q", expectedErrorMessage, err)
		}

		// Version, dirty status unchanged
		if version, dirty, ok, err := store.Version(ctx); err != nil || !ok || !dirty || version != 17 {
			t.Fatalf("unexpected version. want=(version=%d, dirty=%v), have=(version=%d, dirty=%v, ok=%v, error=%q)", 17, true, version, dirty, ok, err)
		}

		assertLogs(t, ctx, store, nil)
	})
}

func TestWrappedDown(t *testing.T) {
	db := dbtest.NewDB(t)
	store := testStore(db)
	ctx := context.Background()

	if err := store.EnsureSchemaTable(ctx); err != nil {
		t.Fatalf("unexpected error ensuring schema table exists: %s", err)
	}
	if err := store.Exec(ctx, sqlf.Sprintf(`INSERT INTO test_migrations_table VALUES (14, false)`)); err != nil {
		t.Fatalf("unexpected error setting initial version: %s", err)
	}
	if err := store.Exec(ctx, sqlf.Sprintf(`
		CREATE TABLE test_trees (
			name text,
			leaf_type text,
			seed_type text,
			bark_type text
		);
	`)); err != nil {
		t.Fatalf("unexpected error creating test table: %s", err)
	}

	testQuery := sqlf.Sprintf(`
		INSERT INTO test_trees VALUES
			('oak', 'broad', 'regular', 'strong'),
			('birch', 'narrow', 'regular', 'flaky'),
			('pine', 'needle', 'pine cone', 'soft');
	`)

	// run twice to ensure the error post-migration is not due to an index constraint
	if err := store.Exec(ctx, testQuery); err != nil {
		t.Fatalf("unexpected error inserting into test table: %s", err)
	}
	if err := store.Exec(ctx, testQuery); err != nil {
		t.Fatalf("unexpected error inserting into test table: %s", err)
	}

	t.Run("success", func(t *testing.T) {
		definition := definition.Definition{
			ID: 14,
			DownQuery: sqlf.Sprintf(`
				DROP TABLE test_trees;
			`),
		}
		f := func() error {
			return store.Down(ctx, definition)
		}
		if err := store.WithMigrationLog(ctx, definition, false, f); err != nil {
			t.Fatalf("unexpected error running migration: %s", err)
		}

		// note: this query succeeded twice earlier
		if err := store.Exec(ctx, testQuery); err == nil || !strings.HasPrefix(err.Error(), "SQL Error") {
			t.Fatalf("migration query did not succeed; expected missing table. want=%q have=%q", "SQL Error", err)
		}

		// Version set to migration ID; not dirty
		if version, dirty, ok, err := store.Version(ctx); err != nil || !ok || dirty || version != 13 {
			t.Fatalf("unexpected version. want=(version=%d, dirty=%v), have=(version=%d, dirty=%v, ok=%v, error=%q)", 13, false, version, dirty, ok, err)
		}

		assertLogs(t, ctx, store, []migrationLog{
			{
				Schema:  "test_migrations_table",
				Version: 14,
				Up:      false,
				Success: boolPtr(true),
			},
		})
		assertVersions(t, ctx, store, nil, nil, nil)
		truncateLogs(t, ctx, store)
	})

	t.Run("unexpected version", func(t *testing.T) {
		expectedErrorMessage := "expected schema to have version 12, but has version 13"

		definition := definition.Definition{
			ID: 12,
			DownQuery: sqlf.Sprintf(`
				-- Does not actually run
			`),
		}
		f := func() error {
			return store.Down(ctx, definition)
		}
		if err := store.WithMigrationLog(ctx, definition, false, f); err == nil || !strings.HasPrefix(err.Error(), expectedErrorMessage) {
			t.Fatalf("unexpected error want=%q have=%q", expectedErrorMessage, err)
		}

		// Version, dirty status unchanged
		if version, dirty, ok, err := store.Version(ctx); err != nil || !ok || dirty || version != 13 {
			t.Fatalf("unexpected version. want=(version=%d, dirty=%v), have=(version=%d, dirty=%v, ok=%v, error=%q)", 13, false, version, dirty, ok, err)
		}

		assertLogs(t, ctx, store, nil)
	})

	t.Run("query failure", func(t *testing.T) {
		expectedErrorMessage := "SQL Error"

		definition := definition.Definition{
			ID: 13,
			DownQuery: sqlf.Sprintf(`
				-- Note: table does not exist
				DROP TABLE TABLE test_trees;
			`),
		}
		f := func() error {
			return store.Down(ctx, definition)
		}
		if err := store.WithMigrationLog(ctx, definition, false, f); err == nil || !strings.HasPrefix(err.Error(), expectedErrorMessage) {
			t.Fatalf("unexpected error want=%q have=%q", expectedErrorMessage, err)
		}

		// Version set to migration ID; dirty
		if version, dirty, ok, err := store.Version(ctx); err != nil || !ok || !dirty || version != 12 {
			t.Fatalf("unexpected version. want=(version=%d, dirty=%v), have=(version=%d, dirty=%v, ok=%v, error=%q)", 12, true, version, dirty, ok, err)
		}

		assertLogs(t, ctx, store, []migrationLog{
			{
				Schema:  "test_migrations_table",
				Version: 13,
				Up:      false,
				Success: boolPtr(false),
			},
		})
		assertVersions(t, ctx, store, nil, nil, []int{13})
		truncateLogs(t, ctx, store)
	})

	t.Run("dirty", func(t *testing.T) {
		expectedErrorMessage := "dirty database"

		definition := definition.Definition{
			ID: 12,
			DownQuery: sqlf.Sprintf(`
				-- Does not actually run
			`),
		}
		f := func() error {
			return store.Down(ctx, definition)
		}
		if err := store.WithMigrationLog(ctx, definition, false, f); err == nil || !strings.HasPrefix(err.Error(), expectedErrorMessage) {
			t.Fatalf("unexpected error want=%q have=%q", expectedErrorMessage, err)
		}

		// Version, dirty status unchanged
		if version, dirty, ok, err := store.Version(ctx); err != nil || !ok || !dirty || version != 12 {
			t.Fatalf("unexpected version. want=(version=%d, dirty=%v), have=(version=%d, dirty=%v, ok=%v, error=%q)", 12, true, version, dirty, ok, err)
		}

		assertLogs(t, ctx, store, nil)
	})
}

func TestIndexStatus(t *testing.T) {
	db := dbtest.NewDB(t)
	store := testStore(db)
	ctx := context.Background()

	if _, err := db.ExecContext(ctx, "CREATE TABLE tbl (id text, name text);"); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	// Index does not (yet) exist
	if _, ok, err := store.IndexStatus(ctx, "tbl", "idx"); err != nil {
		t.Fatalf("unexpected error: %s", err)
	} else if ok {
		t.Fatalf("unexpected index status")
	}

	// Wrap context in a small timeout; we do tight for-loops here to determine
	// when we can continue on to/unblock the next operation, but none of the
	// steps should take any significant time.
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	group, groupCtx := errgroup.WithContext(ctx)
	defer cancel()

	whileEmpty := func(ctx context.Context, conn dbutil.DB, query string) error {
		for {
			rows, err := conn.QueryContext(ctx, query)
			if err != nil {
				return err
			}

			lockVisible := rows.Next()

			if err := basestore.CloseRows(rows, nil); err != nil {
				return err
			}

			if lockVisible {
				return nil
			}
		}
	}

	// Create separate connections to precise control contention of resources
	// so we can examine what this method returns while an index is being created.

	conns := make([]*sql.Conn, 3)
	for i := 0; i < 3; i++ {
		conn, err := db.Conn(ctx)
		if err != nil {
			t.Fatalf("failed to open new connection: %s", err)
		}
		t.Cleanup(func() { conn.Close() })

		conns[i] = conn
	}
	connA, connB, connC := conns[0], conns[1], conns[2]

	lockQuery := `SELECT pg_advisory_lock(10, 10)`
	unlockQuery := `SELECT pg_advisory_unlock(10, 10)`
	createIndexQuery := `CREATE INDEX CONCURRENTLY idx ON tbl(id)`

	// Session A
	// Successfully take and hold advisory lock
	if _, err := connA.ExecContext(ctx, lockQuery); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	// Session B
	// Try to take advisory lock; blocked by Session A
	group.Go(func() error {
		_, err := connB.ExecContext(groupCtx, lockQuery)
		return err
	})

	// Session C
	// try to create index concurrently; blocked by session B waiting on session A
	group.Go(func() error {
		// Wait until we can see Session B's lock before attempting to create index
		if err := whileEmpty(groupCtx, connC, "SELECT 1 FROM pg_locks WHERE locktype = 'advisory' AND NOT granted"); err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		_, err := connC.ExecContext(groupCtx, createIndexQuery)
		return err
	})

	// Wait until we can see Session C's lock before querying index status
	if err := whileEmpty(ctx, db, "SELECT 1 FROM pg_locks WHERE locktype = 'relation' AND mode = 'ShareUpdateExclusiveLock'"); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	// "waiting for old snapshots" will be the phase that is blocked by the concurrent
	// sessions holding advisory locks. We may happen to hit one of the earlier phases
	// if we're quick enough, so we'll keep polling progress until we hit the target.
	blockingPhase := "waiting for old snapshots"
	nonblockingPhasePrefixes := make([]string, 0, len(storetypes.CreateIndexConcurrentlyPhases))
	for _, prefix := range storetypes.CreateIndexConcurrentlyPhases {
		if prefix == blockingPhase {
			break
		}

		nonblockingPhasePrefixes = append(nonblockingPhasePrefixes, prefix)
	}
	compareWithPrefix := func(value, prefix string) bool {
		return value == prefix || strings.HasPrefix(value, prefix+":")
	}

retryLoop:
	for {
		if status, ok, err := store.IndexStatus(ctx, "tbl", "idx"); err != nil {
			t.Fatalf("unexpected error: %s", err)
		} else if !ok {
			t.Fatalf("expected index status")
		} else if status.Phase == nil {
			t.Fatalf("unexpected phase. want=%q have=nil", blockingPhase)
		} else if *status.Phase == blockingPhase {
			break
		} else {
			for _, prefix := range nonblockingPhasePrefixes {
				if compareWithPrefix(*status.Phase, prefix) {
					continue retryLoop
				}
			}

			t.Fatalf("unexpected phase. want=%q have=%q", blockingPhase, *status.Phase)
		}
	}

	// Session A
	// Unlock, unblocking both Session B and Session C
	if _, err := connA.ExecContext(ctx, unlockQuery); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	// Wait for index creation to complete
	if err := group.Wait(); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if status, ok, err := store.IndexStatus(ctx, "tbl", "idx"); err != nil {
		t.Fatalf("unexpected error: %s", err)
	} else if !ok {
		t.Fatalf("expected index status")
	} else {
		if !status.IsValid {
			t.Fatalf("unexpected isvalid. want=%v have=%v", true, status.IsValid)
		}
		if status.Phase != nil {
			t.Fatalf("unexpected phase. want=%v have=%v", nil, status.Phase)
		}
	}
}

func testStore(db dbutil.DB) *Store {
	return NewWithDB(db, "test_migrations_table", NewOperations(&observation.TestContext))
}

func strPtr(v string) *string {
	return &v
}

func boolPtr(value bool) *bool {
	return &value
}

func truncateLogs(t *testing.T, ctx context.Context, store *Store) {
	t.Helper()

	if err := store.Exec(ctx, sqlf.Sprintf(`TRUNCATE migration_logs`)); err != nil {
		t.Fatalf("unexpected error truncating logs: %s", err)
	}
}

func assertLogs(t *testing.T, ctx context.Context, store *Store, expectedLogs []migrationLog) {
	t.Helper()

	logs, err := scanMigrationLogs(store.Query(ctx, sqlf.Sprintf(`SELECT schema, version, up, success FROM migration_logs ORDER BY started_at`)))
	if err != nil {
		t.Fatalf("unexpected error scanning logs: %s", err)
	}

	if diff := cmp.Diff(expectedLogs, logs); diff != "" {
		t.Errorf("unexpected migration logs (-want +got):\n%s", diff)
	}
}

func assertVersions(t *testing.T, ctx context.Context, store *Store, expectedAppliedVersions, expectedPendingVersions, expectedFailedVersions []int) {
	t.Helper()

	appliedVersions, pendingVersions, failedVersions, err := store.Versions(ctx)
	if err != nil {
		t.Fatalf("unexpected error querying version: %s", err)
	}

	if diff := cmp.Diff(expectedAppliedVersions, appliedVersions); diff != "" {
		t.Errorf("unexpected applied migration logs (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(expectedPendingVersions, pendingVersions); diff != "" {
		t.Errorf("unexpected pending migration logs (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(expectedFailedVersions, failedVersions); diff != "" {
		t.Errorf("unexpected failed migration logs (-want +got):\n%s", diff)
	}
}
