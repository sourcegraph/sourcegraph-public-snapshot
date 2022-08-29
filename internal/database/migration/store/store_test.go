package store

import (
	"context"
	"database/sql"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"
	"golang.org/x/sync/errgroup"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/definition"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/shared"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestEnsureSchemaTable(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtest.NewDB(logger, t)
	store := testStore(db)
	ctx := context.Background()

	// Test initially missing table
	if err := store.Exec(ctx, sqlf.Sprintf("SELECT * FROM migration_logs")); err == nil {
		t.Fatalf("expected query to fail due to missing table migration_logs")
	}

	if err := store.EnsureSchemaTable(ctx); err != nil {
		t.Fatalf("unexpected error ensuring schema table exists: %s", err)
	}

	// Test table was created
	if err := store.Exec(ctx, sqlf.Sprintf("SELECT * FROM migration_logs")); err != nil {
		t.Fatalf("unexpected error querying migration_logs: %s", err)
	}

	// Test idempotency
	if err := store.EnsureSchemaTable(ctx); err != nil {
		t.Fatalf("expected method to be idempotent, got error: %s", err)
	}
}

func TestBackfillSchemaVersions(t *testing.T) {
	t.Run("frontend", func(t *testing.T) {
		testViaMigrationLogs(t, "frontend", 1528395834, backfillRange(1528395733, 1528395834)) // squashed root
		testViaGolangMigrate(t, "frontend", 1528395834, backfillRange(1528395733, 1528395834)) // squashed root
		testViaGolangMigrate(t, "frontend", 1528395840, backfillRange(1528395733, 1528395840)) // non-squashed migration
	})

	t.Run("codeintel", func(t *testing.T) {
		testViaMigrationLogs(t, "codeintel", 1000000015, backfillRange(1000000000, 1000000015)) // squashed root
		testViaGolangMigrate(t, "codeintel", 1000000015, backfillRange(1000000000, 1000000015)) // squashed root
		testViaGolangMigrate(t, "codeintel", 1000000020, backfillRange(1000000000, 1000000020)) // non-squashed migration
	})

	t.Run("codeinsights", func(t *testing.T) {
		testViaMigrationLogs(t, "codeinsights", 1000000020, backfillRange(1000000000, 1000000020)) // squashed root
		testViaGolangMigrate(t, "codeinsights", 1000000020, backfillRange(1000000000, 1000000020)) // squashed root
		testViaGolangMigrate(t, "codeinsights", 1000000027, backfillRange(1000000000, 1000000027)) // non-squashed migration
	})
}

// testViaGolangMigrate asserts the given expected versions are backfilled on a new store instance, given
// the .*schema_migrations table has an entry with the given initial version.
func testViaGolangMigrate(t *testing.T, schemaName string, version int, expectedVersions []int) {
	testBackfillSchemaVersion(t, schemaName, expectedVersions, func(ctx context.Context, store *Store) {
		if err := setupGolangMigrateTest(ctx, store, schemaName, version); err != nil {
			t.Fatalf("unexpected error preparing .*schema_migrations tests: %s", err)
		}
	})
}

// setupGolangMigrateTest creates and populates the .*schema_migrations table with the given version.
func setupGolangMigrateTest(ctx context.Context, store *Store, schemaName string, version int) error {
	tableName := quote(schemaName)

	if err := store.Exec(ctx, sqlf.Sprintf(`CREATE TABLE %s (version text, dirty bool)`, tableName)); err != nil {
		return err
	}
	if err := store.Exec(ctx, sqlf.Sprintf(`INSERT INTO %s VALUES (%s, false)`, tableName, strconv.Itoa(version))); err != nil {
		return err
	}

	return nil
}

// testViaMigrationLogs asserts the given expected versions are backfilled on a new store instance, given
// the migration_logs table has an entry with the given initial version.
func testViaMigrationLogs(t *testing.T, schemaName string, initialVersion int, expectedVersions []int) {
	testBackfillSchemaVersion(t, schemaName, expectedVersions, func(ctx context.Context, store *Store) {
		if err := setupMigrationLogsTest(ctx, store, schemaName, initialVersion); err != nil {
			t.Fatalf("unexpected error preparing migration_logs tests: %s", err)
		}
	})
}

// setupMigrationLogsTest populates the migration_logs table with the given version.
func setupMigrationLogsTest(ctx context.Context, store *Store, schemaName string, version int) error {
	return store.Exec(ctx, sqlf.Sprintf(`
		INSERT INTO migration_logs (
			migration_logs_schema_version,
			schema,
			version,
			up,
			started_at,
			finished_at,
			success
		) VALUES (%s, %s, %s, true, NOW(), NOW(), true)
	`,
		currentMigrationLogSchemaVersion,
		schemaName,
		version,
	))
}

// testBackfillSchemaVersion runs the given setup function prior to backfilling a test
// migration store. The versions available post-backfill are checked against the given
// expected versions.
func testBackfillSchemaVersion(
	t *testing.T,
	schemaName string,
	expectedVersions []int,
	setup func(ctx context.Context, store *Store),
) {
	logger := logtest.Scoped(t)
	db := dbtest.NewDB(logger, t)
	store := testStoreWithName(db, schemaName)
	ctx := context.Background()

	if err := store.EnsureSchemaTable(ctx); err != nil {
		t.Fatalf("unexpected error ensuring schema table exists: %s", err)
	}

	setup(ctx, store)

	if err := store.BackfillSchemaVersions(ctx); err != nil {
		t.Fatalf("unexpected error backfilling schema table: %s", err)
	}

	appliedVersions, _, _, err := store.Versions(ctx)
	if err != nil {
		t.Fatalf("unexpected error querying versions: %s", err)
	}
	if diff := cmp.Diff(expectedVersions, appliedVersions); diff != "" {
		t.Errorf("unexpected applied migrations (-want +got):\n%s", diff)
	}
}

// backfillRange creates an integer slice of the shape `[-lo, lo, lo+1, ..., hi-1, hi]`.
// This is used to represent a linear range of migration identifiers from a historic
// squashed migration to a future migration (prior to non-lienar migration identifeirs).
func backfillRange(lo, hi int) []int {
	vs := make([]int, 0, hi-lo+2)
	vs = append(vs, -lo)
	for i := lo; i <= hi; i++ {
		vs = append(vs, i)
	}
	return vs
}

func TestHumanizeSchemaName(t *testing.T) {
	for input, expected := range map[string]string{
		"schema_migrations":              "frontend",
		"codeintel_schema_migrations":    "codeintel",
		"codeinsights_schema_migrations": "codeinsights",
		"test_schema_migrations":         "test",
	} {
		if output := humanizeSchemaName(input); output != expected {
			t.Errorf("unexpected output. want=%q have=%q", expected, output)
		}
	}
}

func TestVersions(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtest.NewDB(logger, t)
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
			defaultTestTableName,
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

func TestTryLock(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtest.NewDB(logger, t)
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
	logger := logtest.Scoped(t)
	db := dbtest.NewDB(logger, t)
	store := testStore(db)
	ctx := context.Background()

	if err := store.EnsureSchemaTable(ctx); err != nil {
		t.Fatalf("unexpected error ensuring schema table exists: %s", err)
	}

	// Seed a few migrations
	for _, id := range []int{13, 14, 15} {
		definition := definition.Definition{
			ID:      id,
			UpQuery: sqlf.Sprintf(`-- No-op`),
		}
		f := func() error {
			return store.Up(ctx, definition)
		}
		if err := store.WithMigrationLog(ctx, definition, true, f); err != nil {
			t.Fatalf("unexpected error running migration: %s", err)
		}
	}

	logs := []migrationLog{
		{
			Schema:  defaultTestTableName,
			Version: 13,
			Up:      true,
			Success: boolPtr(true),
		},
		{
			Schema:  defaultTestTableName,
			Version: 14,
			Up:      true,
			Success: boolPtr(true),
		}, {
			Schema:  defaultTestTableName,
			Version: 15,
			Up:      true,
			Success: boolPtr(true),
		},
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

		logs = append(logs, migrationLog{
			Schema:  defaultTestTableName,
			Version: 16,
			Up:      true,
			Success: boolPtr(true),
		})
		assertLogs(t, ctx, store, logs)
		assertVersions(t, ctx, store, []int{13, 14, 15, 16}, nil, nil)
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
		if err := store.WithMigrationLog(ctx, definition, true, f); err == nil || !strings.Contains(err.Error(), expectedErrorMessage) {
			t.Fatalf("unexpected error want=%q have=%q", expectedErrorMessage, err)
		}

		logs = append(logs, migrationLog{
			Schema:  defaultTestTableName,
			Version: 17,
			Up:      true,
			Success: boolPtr(false),
		})
		assertLogs(t, ctx, store, logs)
		assertVersions(t, ctx, store, []int{13, 14, 15, 16}, nil, []int{17})
	})
}

func TestWrappedDown(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtest.NewDB(logger, t)
	store := testStore(db)
	ctx := context.Background()

	if err := store.EnsureSchemaTable(ctx); err != nil {
		t.Fatalf("unexpected error ensuring schema table exists: %s", err)
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

	// Seed a few migrations
	for _, id := range []int{12, 13, 14} {
		definition := definition.Definition{
			ID:      id,
			UpQuery: sqlf.Sprintf(`-- No-op`),
		}
		f := func() error {
			return store.Up(ctx, definition)
		}
		if err := store.WithMigrationLog(ctx, definition, true, f); err != nil {
			t.Fatalf("unexpected error running migration: %s", err)
		}
	}

	logs := []migrationLog{
		{
			Schema:  defaultTestTableName,
			Version: 12,
			Up:      true,
			Success: boolPtr(true),
		},
		{
			Schema:  defaultTestTableName,
			Version: 13,
			Up:      true,
			Success: boolPtr(true),
		},
		{
			Schema:  defaultTestTableName,
			Version: 14,
			Up:      true,
			Success: boolPtr(true),
		},
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
		if err := store.Exec(ctx, testQuery); err == nil || !strings.Contains(err.Error(), "SQL Error") {
			t.Fatalf("migration query did not succeed; expected missing table. want=%q have=%q", "SQL Error", err)
		}

		logs = append(logs, migrationLog{
			Schema:  defaultTestTableName,
			Version: 14,
			Up:      false,
			Success: boolPtr(true),
		})
		assertLogs(t, ctx, store, logs)
		assertVersions(t, ctx, store, []int{12, 13}, nil, nil)
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
		if err := store.WithMigrationLog(ctx, definition, false, f); err == nil || !strings.Contains(err.Error(), expectedErrorMessage) {
			t.Fatalf("unexpected error want=%q have=%q", expectedErrorMessage, err)
		}

		logs = append(logs, migrationLog{
			Schema:  defaultTestTableName,
			Version: 13,
			Up:      false,
			Success: boolPtr(false),
		})
		assertLogs(t, ctx, store, logs)
		assertVersions(t, ctx, store, []int{12}, nil, []int{13})
	})
}

func TestIndexStatus(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtest.NewDB(logger, t)
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

	// "waiting for old snapshots" will be the phase that is blocked by the concurrent
	// sessions holding advisory locks. We may happen to hit one of the earlier phases
	// if we're quick enough, so we'll keep polling progress until we hit the target.
	blockingPhase := "waiting for old snapshots"
	nonblockingPhasePrefixes := make([]string, 0, len(shared.CreateIndexConcurrentlyPhases))
	for _, prefix := range shared.CreateIndexConcurrentlyPhases {
		if prefix == blockingPhase {
			break
		}

		nonblockingPhasePrefixes = append(nonblockingPhasePrefixes, prefix)
	}
	compareWithPrefix := func(value, prefix string) bool {
		return value == prefix || strings.HasPrefix(value, prefix+":")
	}

	start := time.Now()
	const missingIndexThreshold = time.Second * 10

retryLoop:
	for {
		if status, ok, err := store.IndexStatus(ctx, "tbl", "idx"); err != nil {
			t.Fatalf("unexpected error: %s", err)
		} else if !ok {
			// Give a small amount of time for Session C to begin creating the index. Signaling
			// when Postgres has started to create the index is as difficult and expensive as
			// querying the index the status, so we just poll here for a relatively short time.
			if time.Since(start) >= missingIndexThreshold {
				t.Fatalf("expected index status after %s", missingIndexThreshold)
			}
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

const defaultTestTableName = "test_migrations_table"

func testStore(db *sql.DB) *Store {
	return testStoreWithName(db, defaultTestTableName)
}

func testStoreWithName(db *sql.DB, name string) *Store {
	return NewWithDB(db, name, NewOperations(&observation.TestContext))
}

func strPtr(v string) *string {
	return &v
}

func boolPtr(value bool) *bool {
	return &value
}

func assertLogs(t *testing.T, ctx context.Context, store *Store, expectedLogs []migrationLog) {
	t.Helper()

	sort.Slice(expectedLogs, func(i, j int) bool {
		return expectedLogs[i].Version < expectedLogs[j].Version
	})

	logs, err := scanMigrationLogs(store.Query(ctx, sqlf.Sprintf(`SELECT schema, version, up, success FROM migration_logs ORDER BY version`)))
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
